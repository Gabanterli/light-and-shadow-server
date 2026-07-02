package pvp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/blessing"
	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/housing"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/movement"
)

// Skull constants
const (
	SkullNone  = "none"
	SkullWhite = "white"
	SkullRed   = "red"
	SkullBlack = "black"
)

// Default bounty rewards
const (
	BountyWhite = 10000  // 10k gold
	BountyRed   = 50000  // 50k gold
	BountyBlack = 250000 // 250k gold
)

// SkullState holds the current PvP status of a character
type SkullState struct {
	CharacterID     int       `json:"character_id"`
	CharacterName   string    `json:"character_name"`
	SkullTier       string    `json:"skull_tier"`
	UnjustKills     int       `json:"unjust_kills"`
	BountyReward    int64     `json:"bounty_reward"`
	LastKnownRegion string    `json:"last_known_region"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BountyState represents a wanted poster on the Wanted Board
type BountyState struct {
	PlayerName      string    `json:"player_name"`
	SkullTier       string    `json:"skull_tier"`
	UnjustKillCount int       `json:"unjust_kill_count"`
	BountyReward    int64     `json:"bounty_reward"`
	LastKnownRegion string    `json:"last_known_region"`
	CreatedAt       time.Time `json:"created_at"`
}

// HospitalConfig matches the structure loaded from JSON
type HospitalConfig struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContinentName string  `json:"continent_name"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             int     `json:"z"`
}

// AltarConfig matches the structure loaded from JSON
type AltarConfig struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContinentName string  `json:"continent_name"`
	BlessingID    string  `json:"blessing_id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             int     `json:"z"`
}

type ConfigWrapper struct {
	Altars []AltarConfig `json:"altars"`
}

type HospitalsWrapper struct {
	Hospitals []HospitalConfig `json:"hospitals"`
}

// PvPManager coordinates PvP permissions, safe zones, skulls, and bounties
type PvPManager struct {
	mu                sync.RWMutex
	db                *sql.DB
	housingMgr        *housing.HousingManager
	blessingMgr       *blessing.BlessingManager
	combatMgr         *combat.CombatManager
	inventories       map[string]*inventory.PlayerInventory

	// Config maps for Safe-Zone validations
	hospitals         []HospitalConfig
	altars            []AltarConfig
	regions           []movement.WorldRegion

	// In-memory states and slide tracking
	unjustKillsCache  map[string][]time.Time // playerID -> list of kill timestamps
	skullsCache       map[string]*SkullState  // playerID -> skull
	bounties          map[string]*BountyState // playerID -> bounty

	// Exemption status registers
	arenaPlayers      map[string]bool         // playerID -> true
	activeDuels       map[string]string       // player1 -> player2 (and vice versa)
	guildWars         map[string]map[string]bool // guildID -> targetGuildIDs

	// Hardening records
	killTradeTracker   map[string]time.Time    // "attacker:victim" -> last kill time
	combatLogTimeout   map[string]time.Time    // playerID -> vulnerability expiration
	combatLockTimeout  map[string]time.Time    // playerID -> combat lock expiration (PATCH 6)
	bountyClaimHistory map[string][]time.Time  // "killer:victim" -> claim timestamps (PATCH 8)
}

// NewPvPManager instantiates a clean PvPManager with canonical configurations and fallbacks
func NewPvPManager(
	db *sql.DB,
	hm *housing.HousingManager,
	bm *blessing.BlessingManager,
	cm *combat.CombatManager,
	invs map[string]*inventory.PlayerInventory,
) *PvPManager {
	pm := &PvPManager{
		db:                 db,
		housingMgr:         hm,
		blessingMgr:        bm,
		combatMgr:          cm,
		inventories:        invs,
		unjustKillsCache:   make(map[string][]time.Time),
		skullsCache:        make(map[string]*SkullState),
		bounties:           make(map[string]*BountyState),
		arenaPlayers:       make(map[string]bool),
		activeDuels:        make(map[string]string),
		guildWars:          make(map[string]map[string]bool),
		killTradeTracker:   make(map[string]time.Time),
		combatLogTimeout:   make(map[string]time.Time),
		combatLockTimeout:  make(map[string]time.Time),
		bountyClaimHistory: make(map[string][]time.Time),
	}

	if cm != nil {
		cm.OnPvPDealt = func(attackerID, targetID string) {
			pm.TriggerCombatLock(attackerID)
			pm.TriggerCombatLock(targetID)
		}
		cm.OnPvPSpellCast = func(attackerID string) {
			pm.TriggerCombatLock(attackerID)
		}
	}

	pm.loadStaticConfigurations()
	return pm
}

// loadStaticConfigurations populates hospital, altar, and region settings safely
func (pm *PvPManager) loadStaticConfigurations() {
	paths := []string{"backend/config/", "config/", "../config/", "../../config/"}

	// 1. Load Hospitals
	var hospData []byte
	var err error
	for _, p := range paths {
		if data, errRead := os.ReadFile(p + "hospitals_config.json"); errRead == nil {
			hospData = data
			break
		}
	}
	if len(hospData) > 0 {
		var wrapper HospitalsWrapper
		if err = json.Unmarshal(hospData, &wrapper); err == nil {
			pm.hospitals = wrapper.Hospitals
		}
	}
	// Fallback hospitals if file not found
	if len(pm.hospitals) == 0 {
		pm.hospitals = []HospitalConfig{
			{"beginner_infirmary", "Beginner Camp Military Infirmary", "Main Continent", 120.0, 120.0, 0},
			{"fire_temple", "Hearth Healing Temple", "Fire Continent", 2110.0, 2110.0, 0},
			{"ice_hospital", "Frostbite Infirmary", "Ice Continent", 2310.0, 2310.0, 0},
			{"holy_cathedral", "Sanctum Cathedral Hospital", "Holy Continent", 2510.0, 2510.0, 0},
			{"shadow_infirmary", "Eclipse Military Infirmary", "Shadow Continent", 2710.0, 2710.0, 0},
			{"nature_shrine", "Wildwood Healing Shrine", "Nature Continent", 2910.0, 2910.0, 0},
			{"abyssia_temple", "Last Bastion Recovery Clinic", "Abyssia", 3410.0, 3110.0, 0},
		}
	}

	// 2. Load Blessing Altars
	var altarData []byte
	for _, p := range paths {
		if data, errRead := os.ReadFile(p + "blessings_config.json"); errRead == nil {
			altarData = data
			break
		}
	}
	if len(altarData) > 0 {
		var wrapper ConfigWrapper
		if err = json.Unmarshal(altarData, &wrapper); err == nil {
			pm.altars = wrapper.Altars
		}
	}
	// Fallback altars if file not found
	if len(pm.altars) == 0 {
		pm.altars = []AltarConfig{
			{"altar_fire", "Hearth Altar", "Fire Continent", "blessing_fire", 2100.0, 2100.0, 0},
			{"altar_ice", "Frost Altar", "Ice Continent", "blessing_ice", 2300.0, 2300.0, 0},
			{"altar_holy", "Sanctum Altar", "Holy Continent", "blessing_holy", 2500.0, 2500.0, 0},
			{"altar_shadow", "Eclipse Altar", "Shadow Continent", "blessing_shadow", 2700.0, 2700.0, 0},
			{"altar_nature", "Wild Altar", "Nature Continent", "blessing_nature", 2900.0, 2900.0, 0},
			{"altar_abyss", "Void Altar", "Abyssia", "blessing_abyss", 3405.0, 3105.0, 0},
			{"altar_light", "Aurora Altar", "Main Continent", "blessing_light", 100.0, 100.0, 0},
		}
	}

	// 3. Load Regions
	var regionData []byte
	for _, p := range paths {
		if data, errRead := os.ReadFile(p + "world_regions.json"); errRead == nil {
			regionData = data
			break
		}
	}
	if len(regionData) > 0 {
		var list []movement.WorldRegion
		if err = json.Unmarshal(regionData, &list); err == nil {
			pm.regions = list
		}
	}
	// Fallback regions
	if len(pm.regions) == 0 {
		pm.regions = []movement.WorldRegion{
			{
				RegionID:      "main_continent",
				ContinentName: "Main Continent",
				MinLevel:      1,
				SafeRespawn:   movement.SafeRespawn{X: 100, Y: 100, Z: 0},
			},
			{
				RegionID:      "fire_continent",
				ContinentName: "Fire Continent",
				MinLevel:      50,
				SafeRespawn:   movement.SafeRespawn{X: 2100, Y: 2100, Z: 0},
			},
		}
	}
}

// IsInSafeZone checks if coordinates fall inside any designated PvP Safe Zone
func (pm *PvPManager) IsInSafeZone(x, y float64, z int) bool {
	// 1. Check proximity to Hospitals (within 20 meters)
	for _, h := range pm.hospitals {
		if z == h.Z && math.Hypot(x-h.X, y-h.Y) <= 20.0 {
			return true
		}
	}

	// 2. Check proximity to Blessing Altars (within 20 meters)
	for _, a := range pm.altars {
		if z == a.Z && math.Hypot(x-a.X, y-a.Y) <= 20.0 {
			return true
		}
	}

	// 3. Check proximity to House interiors (within 10 meters of any configured house coordinates)
	if pm.housingMgr != nil {
		for _, house := range pm.housingMgr.GetAllHouses() {
			if z == house.Z && math.Hypot(x-house.X, y-house.Y) <= 10.0 {
				return true
			}
		}
	}

	// 4. Check proximity to Capitals/Starting Barracks (defined as 50 meters surrounding Region/Continent Safe Respawns)
	for _, r := range pm.regions {
		if z == r.SafeRespawn.Z && math.Hypot(x-r.SafeRespawn.X, y-r.SafeRespawn.Y) <= 50.0 {
			return true
		}
	}

	return false
}

// ValidatePvPPermission enforces level 50+ limits and safe zone validation
func (pm *PvPManager) ValidatePvPPermission(attackerID, targetID string) error {
	attackerStats, existsAtt := pm.combatMgr.GetEntityStats(attackerID)
	targetStats, existsTar := pm.combatMgr.GetEntityStats(targetID)

	if !existsAtt || !existsTar {
		return errors.New("attacker or target does not exist in the combat world")
	}

	// PvP only applies if both entities are players
	if !attackerStats.IsPlayer || !targetStats.IsPlayer {
		return nil
	}

	// 1. Open PvP Level Protection (Level 1-49 fully protected)
	if attackerStats.Level < 50 {
		return fmt.Errorf("PvP combat is disabled for players under level 50 (%s is level %d)", attackerID, attackerStats.Level)
	}
	if targetStats.Level < 50 {
		return fmt.Errorf("target player %s is under level 50 and fully protected from PvP", targetID)
	}

	// 2. Safe Zone Protection check for both Attacker and Target
	ax, ay, hasAPos := pm.combatMgr.GetEntityPosition(attackerID)
	tx, ty, hasTPos := pm.combatMgr.GetEntityPosition(targetID)

	if hasAPos {
		if pm.IsInSafeZone(ax, ay, 0) && !pm.IsCombatLocked(attackerID) { // assuming standard ground floor Z=0
			return errors.New("cannot initiate PvP combat inside a safe zone (capitals, barracks, hospitals, temples, house interiors)")
		}
	}
	if hasTPos {
		if pm.IsInSafeZone(tx, ty, 0) && !pm.IsCombatLocked(targetID) {
			return errors.New("target is sheltered inside a safe zone and cannot be attacked")
		}
	}

	return nil
}

// PruneSlidingWindow removes unjust kills older than 24 hours for a player and returns the remaining count
func (pm *PvPManager) PruneSlidingWindow(playerID string) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	kills, exists := pm.unjustKillsCache[playerID]
	if !exists {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	var validKills []time.Time
	for _, t := range kills {
		if t.After(cutoff) {
			validKills = append(validKills, t)
		}
	}

	pm.unjustKillsCache[playerID] = validKills
	return len(validKills)
}

// GetSkullState returns the current skull tier and details for a player
func (pm *PvPManager) GetSkullState(playerID string) (*SkullState, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	state, ok := pm.skullsCache[playerID]
	return state, ok
}

// SetArenaStatus allows manual test/arena systems to toggle PvP exemption
func (pm *PvPManager) SetArenaStatus(playerID string, inside bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if inside {
		pm.arenaPlayers[playerID] = true
	} else {
		delete(pm.arenaPlayers, playerID)
	}
}

// RegisterConsensualDuel registers a duel mapping to exempt both players from unjust kill tracking
func (pm *PvPManager) RegisterConsensualDuel(player1, player2 string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.activeDuels[player1] = player2
	pm.activeDuels[player2] = player1
}

// EndConsensualDuel clears the registered duel between players
func (pm *PvPManager) EndConsensualDuel(playerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	opponent, exists := pm.activeDuels[playerID]
	if exists {
		delete(pm.activeDuels, playerID)
		delete(pm.activeDuels, opponent)
	}
}

// SetGuildWarState configures friendly/hostile war flags to skip unjust kills
func (pm *PvPManager) SetGuildWarState(guildA, guildB string, active bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if active {
		if pm.guildWars[guildA] == nil {
			pm.guildWars[guildA] = make(map[string]bool)
		}
		pm.guildWars[guildA][guildB] = true

		if pm.guildWars[guildB] == nil {
			pm.guildWars[guildB] = make(map[string]bool)
		}
		pm.guildWars[guildB][guildA] = true
	} else {
		if pm.guildWars[guildA] != nil {
			delete(pm.guildWars[guildA], guildB)
		}
		if pm.guildWars[guildB] != nil {
			delete(pm.guildWars[guildB], guildA)
		}
	}
}

// checkExemptions returns true if the kill was exempt from the unjust skull system
func (pm *PvPManager) checkExemptions(killer, victim string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 1. Consensual duels
	if pm.activeDuels[killer] == victim {
		return true
	}

	// 2. Arena players
	if pm.arenaPlayers[killer] || pm.arenaPlayers[victim] {
		return true
	}

	// 3. Killing black skull targets (free-for-all globalwanted)
	if vicSkull, exists := pm.skullsCache[victim]; exists && vicSkull.SkullTier == SkullBlack {
		return true
	}

	// 4. Killing hostile flagged targets (any active skull is hostile)
	if vicSkull, exists := pm.skullsCache[victim]; exists && vicSkull.SkullTier != SkullNone {
		return true
	}

	// 5. Guild Wars (look up guilds)
	// We check combat factions or guilds if registered
	attackerStats, ok1 := pm.combatMgr.GetEntityStats(killer)
	targetStats, ok2 := pm.combatMgr.GetEntityStats(victim)
	if ok1 && ok2 && attackerStats.Faction != "" && targetStats.Faction != "" {
		if pm.guildWars[attackerStats.Faction] != nil && pm.guildWars[attackerStats.Faction][targetStats.Faction] {
			return true
		}
	}

	return false
}

// HandleKillRecord processes a player-on-player kill, updates skull tiers, and registers bounties
func (pm *PvPManager) HandleKillRecord(killer, victim string) (*SkullState, error) {
	if killer == victim {
		return nil, errors.New("cannot register self-kill")
	}

	// Anti-exploit: Multi-account / Same account check (requires DB characters query)
	ctx := context.Background()
	var killerCharID, victimCharID int
	var killerAccID, victimAccID int

	if pm.db != nil {
		errK := pm.db.QueryRowContext(ctx, "SELECT id, account_id FROM characters WHERE name = $1", killer).Scan(&killerCharID, &killerAccID)
		errV := pm.db.QueryRowContext(ctx, "SELECT id, account_id FROM characters WHERE name = $1", victim).Scan(&victimCharID, &victimAccID)
		if errK == nil && errV == nil {
			if killerAccID == victimAccID {
				slog.Warn("[EXPLOIT BLOCK] Multi-account PvP abuse detected: killer and victim belong to the same account", "killer", killer, "victim", victim)
				return nil, fmt.Errorf("multi-account kill trading exploit blocked")
			}
		}
	}

	// Anti-exploit: Kill trading rapid repetition cooldown (1 hour cooldown per player pair)
	key := killer + ":" + victim
	pm.mu.Lock()
	lastKill, existsTrade := pm.killTradeTracker[key]
	if existsTrade && time.Since(lastKill) < 1*time.Hour {
		pm.mu.Unlock()
		slog.Warn("[EXPLOIT BLOCK] Excessive kill trading detected on the same target", "killer", killer, "victim", victim)
		return nil, fmt.Errorf("excessive kill trading detected: 1-hour cooldown required for skull credit")
	}
	pm.killTradeTracker[key] = time.Now()
	pm.mu.Unlock()

	// Check if this kill is exempt (e.g. Arena, Duel, Faction War)
	if pm.checkExemptions(killer, victim) {
		slog.Info("Kill is exempt from unjust penalty (consensual, guild war, arena, or hostile target)", "killer", killer, "victim", victim)
		return nil, nil
	}

	// Increment unjust kills list
	pm.mu.Lock()
	pm.unjustKillsCache[killer] = append(pm.unjustKillsCache[killer], time.Now())
	pm.mu.Unlock()

	// Recalculate total valid unjust kills in the sliding 24h window
	unjustCount := pm.PruneSlidingWindow(killer)

	// Persist unjust kill event to DB if available
	if pm.db != nil && killerCharID > 0 && victimCharID > 0 {
		go func() {
			_, errExec := pm.db.ExecContext(context.Background(), `
				INSERT INTO pvp_unjust_kills (killer_id, victim_id) VALUES ($1, $2)
			`, killerCharID, victimCharID)
			if errExec != nil {
				slog.Error("Failed to persist unjust kill to DB", "error", errExec)
			}
		}()
	}

	// Determine new skull tier
	newTier := SkullNone
	var baseReward int64 = 0
	if unjustCount >= 10 {
		newTier = SkullBlack
		baseReward = BountyBlack
	} else if unjustCount >= 5 {
		newTier = SkullRed
		baseReward = BountyRed
	} else if unjustCount >= 3 {
		newTier = SkullWhite
		baseReward = BountyWhite
	}

	// Calculate dynamic reward (PATCH 7)
	var reward int64 = 0
	if newTier != SkullNone {
		reward = pm.CalculateDynamicBounty(killer, baseReward)
	}

	// Get last known region of the wanted player
	regionName := "Unknown Territory"
	if vx, vy, okPos := pm.combatMgr.GetEntityPosition(killer); okPos {
		for _, r := range pm.regions {
			// Find region name match if possible
			regionName = r.ContinentName
		}
		if pm.housingMgr != nil {
			// Check if inside a house
			for _, house := range pm.housingMgr.GetAllHouses() {
				if math.Hypot(vx-house.X, vy-house.Y) <= 10.0 {
					regionName = "House: " + house.Name
					break
				}
			}
		}
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	state, exists := pm.skullsCache[killer]
	if !exists {
		state = &SkullState{
			CharacterID: killerCharID,
			CharacterName: killer,
		}
		pm.skullsCache[killer] = state
	}

	state.SkullTier = newTier
	state.UnjustKills = unjustCount
	state.BountyReward = reward
	state.LastKnownRegion = regionName
	state.UpdatedAt = time.Now()

	// Update wanted board list
	if newTier != SkullNone {
		pm.bounties[killer] = &BountyState{
			PlayerName:      killer,
			SkullTier:       newTier,
			UnjustKillCount: unjustCount,
			BountyReward:    reward,
			LastKnownRegion: regionName,
			CreatedAt:       time.Now(),
		}
	} else {
		delete(pm.bounties, killer)
	}

	// Persist skull updates to DB
	if pm.db != nil && killerCharID > 0 {
		go func(charID int, tier string, kills int, bounty int64, reg string) {
			_, errExec := pm.db.ExecContext(context.Background(), `
				INSERT INTO pvp_skulls (character_id, skull_tier, unjust_kill_count, bounty_reward, last_known_region, updated_at)
				VALUES ($1, $2, $3, $4, $5, NOW())
				ON CONFLICT (character_id) DO UPDATE SET
					skull_tier = EXCLUDED.skull_tier,
					unjust_kill_count = EXCLUDED.unjust_kill_count,
					bounty_reward = EXCLUDED.bounty_reward,
					last_known_region = EXCLUDED.last_known_region,
					updated_at = NOW()
			`, charID, tier, kills, bounty, reg)
			if errExec != nil {
				slog.Error("Failed to persist pvp_skull state to DB", "error", errExec)
			}
		}(killerCharID, newTier, unjustCount, reward, regionName)
	}

	slog.Info("Unjust kill processed, skull state updated", "player", killer, "kills_24h", unjustCount, "skull", newTier, "bounty", reward)
	return state, nil
}

// HandleBountyClaim rewards a killer automatically upon taking down a flagged wanted skull player
func (pm *PvPManager) HandleBountyClaim(killerName, victimName string) (int64, error) {
	pm.mu.Lock()
	bounty, hasBounty := pm.bounties[victimName]
	if !hasBounty {
		pm.mu.Unlock()
		return 0, nil // No bounty on the victim
	}

	// Apply anti-kill farming reward decay (PATCH 8)
	decayFactor := pm.GetBountyDecayFactor(killerName, victimName)
	rewardAmount := int64(math.Round(float64(bounty.BountyReward) * decayFactor))

	// Double Reward / Duplication Prevention Lock: Delete immediately to avoid re-entry
	delete(pm.bounties, victimName)

	// Clean skull status for the victim upon death
	if vicSkull, exists := pm.skullsCache[victimName]; exists {
		vicSkull.SkullTier = SkullNone
		vicSkull.UnjustKills = 0
		vicSkull.BountyReward = 0
		vicSkull.UpdatedAt = time.Now()
	}
	pm.unjustKillsCache[victimName] = []time.Time{}
	pm.mu.Unlock()

	// Persist cleanup to DB
	if pm.db != nil {
		go func(vic string) {
			_, _ = pm.db.ExecContext(context.Background(), `
				UPDATE pvp_skulls SET skull_tier = 'none', unjust_kill_count = 0, bounty_reward = 0, updated_at = NOW()
				WHERE character_id = (SELECT id FROM characters WHERE name = $1)
			`, vic)
			_, _ = pm.db.ExecContext(context.Background(), `
				DELETE FROM pvp_unjust_kills WHERE killer_id = (SELECT id FROM characters WHERE name = $1)
			`, vic)
		}(victimName)
	}

	// Reward Killer: Add gold (bronze) coins to inventory safely with thread safety
	killerInv, hasInv := pm.inventories[killerName]
	if hasInv {
		killerInv.AddGold(rewardAmount)
		slog.Info("Bounty claimed and rewarded", "killer", killerName, "victim", victimName, "reward", rewardAmount)

		if pm.db != nil {
			go func(kil string, amt int64) {
				_, errExec := pm.db.ExecContext(context.Background(), `
					UPDATE characters SET gold = gold + $1 WHERE name = $2
				`, amt, kil)
				if errExec != nil {
					slog.Error("Failed to persist gold payout for bounty claim", "error", errExec)
				}
			}(killerName, rewardAmount)
		}
	}

	return rewardAmount, nil
}

// GetWantedBoard returns all current active bounty posters
func (pm *PvPManager) GetWantedBoard() []BountyState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	list := make([]BountyState, 0, len(pm.bounties))
	for _, b := range pm.bounties {
		list = append(list, *b)
	}
	return list
}

// HandlePlayerDisconnect protects against combat-logging by adding vulnerability delay
func (pm *PvPManager) HandlePlayerDisconnect(playerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// Mark player as vulnerable for the next 30 seconds to prevent combat logging
	pm.combatLogTimeout[playerID] = time.Now().Add(30 * time.Second)
	slog.Info("Player disconnected in combat zone; combat log anti-cheat activated", "player", playerID)
}

// IsVulnerable checks if disconnected player is still in vulnerability penalty window
func (pm *PvPManager) IsVulnerable(playerID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	exp, exists := pm.combatLogTimeout[playerID]
	if !exists {
		return false
	}
	return time.Now().Before(exp)
}

// TriggerCombatLock sets or refreshes the combat lock for 5 minutes (PATCH 6)
func (pm *PvPManager) TriggerCombatLock(playerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.combatLockTimeout[playerID] = time.Now().Add(5 * time.Minute)
	slog.Info("Player combat-locked", "player", playerID, "until", pm.combatLockTimeout[playerID])
}

// IsCombatLocked checks if a player is currently combat-locked (PATCH 6)
func (pm *PvPManager) IsCombatLocked(playerID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	exp, exists := pm.combatLockTimeout[playerID]
	if !exists {
		return false
	}
	return time.Now().Before(exp)
}

// CanTeleport checks if a player is allowed to teleport (PATCH 6)
func (pm *PvPManager) CanTeleport(playerID string) bool {
	return !pm.IsCombatLocked(playerID)
}

// CanInstantLogout checks if a player is allowed to instantly log out (PATCH 6)
func (pm *PvPManager) CanInstantLogout(playerID string) bool {
	return !pm.IsCombatLocked(playerID)
}

// CanUseHouseRespawn checks if a player is allowed to use their house as respawn location (PATCH 6)
func (pm *PvPManager) CanUseHouseRespawn(playerID string) bool {
	return !pm.IsCombatLocked(playerID)
}

// CalculateDynamicBounty calculates the bounty reward based on target attributes (PATCH 7)
func (pm *PvPManager) CalculateDynamicBounty(victimName string, baseReward int64) int64 {
	level := 50
	unjustKills := 0
	
	// Get level from combat manager
	if stats, ok := pm.combatMgr.GetEntityStats(victimName); ok {
		level = stats.Level
	}
	
	// Get unjust kills from cache
	if skull, ok := pm.skullsCache[victimName]; ok {
		unjustKills = skull.UnjustKills
	} else {
		// Or count from window
		unjustKills = len(pm.unjustKillsCache[victimName])
	}
	
	// base_reward = 500 gold
	// level_multiplier = 1 + ((player_level - 10) * 0.01)
	levelMultiplier := 1.0 + (float64(level - 10) * 0.01)
	if levelMultiplier < 0.1 {
		levelMultiplier = 0.1
	}

	// gear_multiplier:
	// Evaluate equipped slots 0-3. Each high-tier item adds +0.15. Cap multiplier at 2.0x.
	gearMultiplier := 1.0
	if inv, exists := pm.inventories[victimName]; exists && inv != nil {
		highTierCount := 0
		for slot := 0; slot <= 3; slot++ {
			if item, existsItem := inv.Items[slot]; existsItem && item != nil {
				if def, ok := inventory.GetItemDef(item.ItemID); ok {
					if def.Tier >= 2 {
						highTierCount++
					}
				}
			}
		}
		gearMultiplier = 1.0 + float64(highTierCount)*0.15
		if gearMultiplier > 2.0 {
			gearMultiplier = 2.0
		}
	}

	// notoriety_multiplier:
	// Each unjust kill in rolling 24h adds +0.10. Cap multiplier at 3.0x.
	notorietyMultiplier := 1.0 + float64(unjustKills)*0.10
	if notorietyMultiplier > 3.0 {
		notorietyMultiplier = 3.0
	}

	finalReward := float64(500) * levelMultiplier * gearMultiplier * notorietyMultiplier
	return int64(math.Round(finalReward))
}

// PlayerPvPState represents the comprehensive PvP status of a player as per canonical specifications
type PlayerPvPState struct {
	PvpEnabled        bool      `json:"pvp_enabled"`
	CombatLocked      bool      `json:"combat_locked"`
	SkullType         string    `json:"skull_type"`
	UnjustKills24h    int       `json:"unjust_kills_24h"`
	BountyValue       int64     `json:"bounty_value"`
	BlessingsActive   []string  `json:"blessings_active"`
	AolEquipped       bool      `json:"aol_equipped"`
	LastHostileAction time.Time `json:"last_hostile_action"`
}

// GetPlayerPvPState returns the PlayerPvPState for a player
func (pm *PvPManager) GetPlayerPvPState(playerID string) *PlayerPvPState {
	// 1. PvpEnabled: true if level >= 10
	level := 0
	if stats, ok := pm.combatMgr.GetEntityStats(playerID); ok {
		level = stats.Level
	}
	pvpEnabled := level >= 10

	// 2. CombatLocked
	combatLocked := pm.IsCombatLocked(playerID)

	// 3. SkullType & UnjustKills
	skullType := "none"
	unjustCount := 0
	pm.mu.Lock()
	if skull, ok := pm.skullsCache[playerID]; ok {
		skullType = skull.SkullTier
		unjustCount = skull.UnjustKills
	} else {
		unjustCount = len(pm.unjustKillsCache[playerID])
	}
	pm.mu.Unlock()

	// 4. BountyValue
	bountyValue := pm.CalculateDynamicBounty(playerID, 500)

	// 5. BlessingsActive (retrieved from database or default)
	var blessings []string
	if pm.db != nil {
		rows, err := pm.db.QueryContext(context.Background(), `
			SELECT blessing_id FROM character_blessings cb 
			JOIN characters c ON cb.character_id = c.id 
			WHERE c.name = $1 AND cb.acquired = true
		`, playerID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var bName string
				if errScan := rows.Scan(&bName); errScan == nil {
					blessings = append(blessings, bName)
				}
			}
		}
	}

	// 6. AolEquipped
	aolEquipped := false
	if inv, exists := pm.inventories[playerID]; exists && inv != nil {
		aolEquipped = inv.IsAoLEquipped()
	}

	// 7. LastHostileAction
	var lastHostile time.Time
	pm.mu.RLock()
	pm.mu.RUnlock()

	return &PlayerPvPState{
		PvpEnabled:        pvpEnabled,
		CombatLocked:      combatLocked,
		SkullType:         skullType,
		UnjustKills24h:    unjustCount,
		BountyValue:       bountyValue,
		BlessingsActive:   blessings,
		AolEquipped:       aolEquipped,
		LastHostileAction: lastHostile,
	}
}

// GetBountyDecayFactor tracks kills between a specific killer and victim in the last 24h (PATCH 8)
func (pm *PvPManager) GetBountyDecayFactor(killer, victim string) float64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := killer + ":" + victim
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	// Get and prune history
	history := pm.bountyClaimHistory[key]
	var validHistory []time.Time
	for _, t := range history {
		if t.After(cutoff) {
			validHistory = append(validHistory, t)
		}
	}

	// This is the N-th claim in the last 24 hours
	count := len(validHistory)
	
	// Append current claim timestamp
	validHistory = append(validHistory, now)
	pm.bountyClaimHistory[key] = validHistory

	if count == 0 {
		return 1.0  // 100% reward
	} else if count == 1 {
		return 0.25 // 25% reward
	}
	return 0.0      // 0% reward
}
