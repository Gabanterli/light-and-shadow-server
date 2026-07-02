package pvp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/light-and-shadow/backend/pkg/blessing"
	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/housing"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/pve"
)

func TestPvPOpenRulesAndSafeZones(t *testing.T) {
	// Initialize subsystems
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)
	cm.PvPValidator = pm.ValidatePvPPermission

	// Create test entities
	lowLevelPlayer := &combat.EntityStats{
		ID:       "Lowbie",
		IsPlayer: true,
		Level:    10,
		Health:   100,
	}
	highLevelPlayer1 := &combat.EntityStats{
		ID:       "Hero1",
		IsPlayer: true,
		Level:    50,
		Health:   500,
	}
	highLevelPlayer2 := &combat.EntityStats{
		ID:       "Hero2",
		IsPlayer: true,
		Level:    55,
		Health:   600,
	}

	cm.RegisterEntity(lowLevelPlayer, 100, 100)
	cm.RegisterEntity(highLevelPlayer1, 101, 101)
	cm.RegisterEntity(highLevelPlayer2, 102, 102)

	// 1. Level 1-49 protection check
	err := pm.ValidatePvPPermission("Hero1", "Lowbie")
	if err == nil {
		t.Error("Expected error when high level attacks protected lowbie, got nil")
	}

	err = pm.ValidatePvPPermission("Lowbie", "Hero1")
	if err == nil {
		t.Error("Expected error when protected lowbie tries to attack high level, got nil")
	}

	// Level 50+ vs Level 50+ (Open World)
	err = pm.ValidatePvPPermission("Hero1", "Hero2")
	if err != nil {
		t.Errorf("Expected level 50+ open PvP to be allowed, got error: %v", err)
	}

	// 2. Safe zone - Altar proximity (X: 2100, Y: 2100 is Hearth Altar)
	cm.UpdateEntityPosition("Hero1", 2100, 2100)
	err = pm.ValidatePvPPermission("Hero1", "Hero2")
	if err == nil {
		t.Error("Expected error attacking from inside Altar safe-zone, got nil")
	}

	// Move out of Altar
	cm.UpdateEntityPosition("Hero1", 101, 101)

	// Safe zone - Target in hospital (X: 120.0, Y: 120.0)
	cm.UpdateEntityPosition("Hero2", 120, 120)
	err = pm.ValidatePvPPermission("Hero1", "Hero2")
	if err == nil {
		t.Error("Expected error attacking target sheltered inside Hospital safe-zone, got nil")
	}
}

func TestSkullSlidingWindowAndExemptions(t *testing.T) {
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)

	p1 := &combat.EntityStats{ID: "Killer", IsPlayer: true, Level: 50, Health: 100}
	p2 := &combat.EntityStats{ID: "Victim1", IsPlayer: true, Level: 50, Health: 100}
	p3 := &combat.EntityStats{ID: "Victim2", IsPlayer: true, Level: 50, Health: 100}

	cm.RegisterEntity(p1, 500, 500)
	cm.RegisterEntity(p2, 501, 501)
	cm.RegisterEntity(p3, 502, 502)

	// Initialize inventory
	inventories["Killer"] = inventory.NewPlayerInventory("Killer")
	inventories["Victim1"] = inventory.NewPlayerInventory("Victim1")

	// 1. Test Unjust kill count escalating skulls
	for i := 1; i <= 3; i++ {
		vicName := fmt.Sprintf("Target%d", i)
		cm.RegisterEntity(&combat.EntityStats{ID: vicName, IsPlayer: true, Level: 50, Health: 100}, 505, 505)
		inventories[vicName] = inventory.NewPlayerInventory(vicName)
		_, err := pm.HandleKillRecord("Killer", vicName)
		if err != nil {
			t.Fatalf("Failed to handle unjust kill: %v", err)
		}
		// Clear kill trade tracker manually to simulate distinct targets
		pm.mu.Lock()
		pm.killTradeTracker = make(map[string]time.Time)
		pm.mu.Unlock()
	}

	state, ok := pm.GetSkullState("Killer")
	if !ok || state.SkullTier != SkullWhite {
		t.Errorf("Expected White Skull (3 kills), got tier=%s (ok=%t)", state.SkullTier, ok)
	}

	// Escalating to Red Skull (5 kills)
	for i := 4; i <= 5; i++ {
		vicName := fmt.Sprintf("Target%d", i)
		cm.RegisterEntity(&combat.EntityStats{ID: vicName, IsPlayer: true, Level: 50, Health: 100}, 505, 505)
		inventories[vicName] = inventory.NewPlayerInventory(vicName)
		_, _ = pm.HandleKillRecord("Killer", vicName)
		pm.mu.Lock()
		pm.killTradeTracker = make(map[string]time.Time)
		pm.mu.Unlock()
	}

	state, ok = pm.GetSkullState("Killer")
	if !ok || state.SkullTier != SkullRed {
		t.Errorf("Expected Red Skull (5 kills), got tier=%s", state.SkullTier)
	}

	// 2. Consensual duel exemption check
	pm.RegisterConsensualDuel("Killer", "Victim1")
	isExempt := pm.checkExemptions("Killer", "Victim1")
	if !isExempt {
		t.Error("Expected consensual duel kill to be exempt, got non-exempt")
	}

	pm.EndConsensualDuel("Killer")
	isExempt = pm.checkExemptions("Killer", "Victim1")
	if isExempt {
		t.Error("Expected consensual duel end to remove exemption status")
	}
}

func TestBountyAndDeathPenaltyIntegration(t *testing.T) {
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)

	// Setup managers
	dpm := lifecycle.NewDeathPenaltyManager(nil, bm, nil, lifecycle.NewRespawnManager(), hm, inventories, cm)
	dpm.SetPvPManager(pm)

	killer := "BountyHunter"
	victim := "WantedPlayer"

	cm.RegisterEntity(&combat.EntityStats{ID: killer, IsPlayer: true, Level: 50, Health: 100}, 500, 500)
	cm.RegisterEntity(&combat.EntityStats{ID: victim, IsPlayer: true, Level: 50, Health: 100}, 501, 501)

	inventories[killer] = inventory.NewPlayerInventory(killer)
	inventories[victim] = inventory.NewPlayerInventory(victim)

	// Make victim wanted (White Skull)
	for i := 1; i <= 3; i++ {
		vicName := fmt.Sprintf("VTarget%d", i)
		cm.RegisterEntity(&combat.EntityStats{ID: vicName, IsPlayer: true, Level: 50, Health: 100}, 505, 505)
		inventories[vicName] = inventory.NewPlayerInventory(vicName)
		_, _ = pm.HandleKillRecord(victim, vicName)
		pm.mu.Lock()
		pm.killTradeTracker = make(map[string]time.Time)
		pm.mu.Unlock()
	}

	// Verify wanted board has the victim
	wanted := pm.GetWantedBoard()
	if len(wanted) != 1 || wanted[0].PlayerName != victim {
		t.Errorf("Expected victim to be on wanted board, got: %v", wanted)
	}

	// Record killer dealing damage to victim
	cm.RecordAttacker(victim, killer)

	// Apply death penalties (victim dies unblessed)
	_, _, _, _, _, _, _, _, _, err := dpm.ApplyDeathPenalties(victim, "Fire Continent")
	if err != nil {
		t.Fatalf("Failed to apply death penalties: %v", err)
	}

	// Killer should receive gold bounty payment
	goldReward := inventories[killer].GetGold()
	if goldReward != BountyWhite {
		t.Errorf("Expected killer to receive %d gold, got %d", BountyWhite, goldReward)
	}

	// Victim's skull state should be cleaned
	state, ok := pm.GetSkullState(victim)
	if ok && state.SkullTier != SkullNone {
		t.Errorf("Expected victim skull to be reset after death/claim, got: %s", state.SkullTier)
	}
}

func TestBlackSkullBlessingInteraction(t *testing.T) {
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)
	dpm := lifecycle.NewDeathPenaltyManager(nil, bm, nil, lifecycle.NewRespawnManager(), hm, inventories, cm)
	dpm.SetPvPManager(pm)

	player := "BlackSkullPlayer"
	cm.RegisterEntity(&combat.EntityStats{ID: player, IsPlayer: true, Level: 50, Health: 100}, 500, 500)
	playerInv := inventory.NewPlayerInventory(player)
	playerInv.Items[0] = &inventory.InventoryItem{ItemID: "SuperSword", Quantity: 1, Durability: 100}
	playerInv.Items[4] = &inventory.InventoryItem{ItemID: "HealingPotion", Quantity: 5, Durability: 100}
	inventories[player] = playerInv

	// 1. Force Black Skull Status (10 unjust kills)
	for i := 1; i <= 10; i++ {
		vicName := fmt.Sprintf("BSTarget%d", i)
		cm.RegisterEntity(&combat.EntityStats{ID: vicName, IsPlayer: true, Level: 50, Health: 100}, 505, 505)
		inventories[vicName] = inventory.NewPlayerInventory(vicName)
		_, _ = pm.HandleKillRecord(player, vicName)
		pm.mu.Lock()
		pm.killTradeTracker = make(map[string]time.Time)
		pm.mu.Unlock()
	}

	// 2. Scenario A: Black Skull dies fully blessed
	// Give all 7 blessings
	bm.GrantBlessing(player, "blessing_fire")
	bm.GrantBlessing(player, "blessing_ice")
	bm.GrantBlessing(player, "blessing_holy")
	bm.GrantBlessing(player, "blessing_shadow")
	bm.GrantBlessing(player, "blessing_nature")
	bm.GrantBlessing(player, "blessing_abyss")
	bm.GrantBlessing(player, "blessing_light")

	if !bm.IsFullyBlessed(player) {
		t.Fatal("Player should be fully blessed")
	}

	// Execute Death
	_, _, _, drops, _, _, _, _, wasProtected, err := dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("Failed to apply death penalties: %v", err)
	}

	if !wasProtected {
		t.Error("Expected fully-blessed Black Skull player to be protected from penalties")
	}
	if len(drops) > 0 {
		t.Errorf("Expected no items to drop for blessed player, got drops: %v", drops)
	}
	if bm.IsFullyBlessed(player) {
		t.Error("Expected all blessings to be consumed on death protection")
	}

	// Restore items
	playerInv.Items[0] = &inventory.InventoryItem{ItemID: "SuperSword", Quantity: 1, Durability: 100}
	playerInv.Items[4] = &inventory.InventoryItem{ItemID: "HealingPotion", Quantity: 5, Durability: 100}

	// 3. Scenario B: Black Skull dies unblessed (Probabilistic item drop penalty - PATCH 9)
	_, _, _, drops, _, _, _, _, wasProtected, err = dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("Failed to apply unblessed death penalties: %v", err)
	}

	if wasProtected {
		t.Error("Expected unblessed Black Skull player to NOT be protected")
	}
	// Probabilistic drops (can be anything from 0 to 2 items depending on rand)
	if len(drops) > 2 {
		t.Errorf("Expected at most 2 item drops under Black Skull penalty, got: %v", drops)
	}
	// Verify that any dropped item is removed from inventory and any kept item remains
	for _, dItem := range drops {
		for slot, invItem := range playerInv.Items {
			if invItem != nil && invItem.ItemID == dItem.ItemID {
				t.Errorf("Item %s was dropped but still exists in player inventory slot %d", dItem.ItemID, slot)
			}
		}
	}
}

func TestNewPvPPatches(t *testing.T) {
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)
	cm.PvPValidator = pm.ValidatePvPPermission

	// Register combat players
	p1 := &combat.EntityStats{ID: "Hero1", IsPlayer: true, Level: 50, Health: 100}
	p2 := &combat.EntityStats{ID: "Hero2", IsPlayer: true, Level: 60, Health: 100}
	cm.RegisterEntity(p1, 100, 100)
	cm.RegisterEntity(p2, 101, 101)

	inventories["Hero1"] = inventory.NewPlayerInventory("Hero1")
	inventories["Hero2"] = inventory.NewPlayerInventory("Hero2")

	// --- Test PATCH 6: Combat Lock ---
	if pm.IsCombatLocked("Hero1") {
		t.Error("Expected player 1 to not be combat-locked initially")
	}

	// Trigger combat lock
	pm.TriggerCombatLock("Hero1")
	if !pm.IsCombatLocked("Hero1") {
		t.Error("Expected player 1 to be combat-locked")
	}
	if pm.CanTeleport("Hero1") {
		t.Error("Expected combat-locked player to be unable to teleport")
	}
	if pm.CanInstantLogout("Hero1") {
		t.Error("Expected combat-locked player to be unable to instantly log out")
	}
	if pm.CanUseHouseRespawn("Hero1") {
		t.Error("Expected combat-locked player to be unable to use house respawn")
	}

	// --- Test PATCH 7: Dynamic Bounty Scaling ---
	// Level 60 player, with 0 equipped items, 0 unjust kills
	// Set item definitions in ItemDictionary for test
	inventory.ItemDictionary["Armor"] = inventory.ItemDef{ItemID: "Armor", Tier: 2}
	inventory.ItemDictionary["Shield"] = inventory.ItemDef{ItemID: "Shield", Tier: 2}

	bounty60 := pm.CalculateDynamicBounty("Hero2", 500)
	// Base level multiplier for level 60 should be: 1.0 + (60-10)*0.01 = 1.5
	// Gear score multiplier = 1.0
	// Notoriety multiplier = 1.0
	// Expected: 500 * 1.5 * 1.0 * 1.0 = 750
	if bounty60 != 750 {
		t.Errorf("Expected dynamic bounty of 750 for level 60 player, got: %d", bounty60)
	}

	// Equip 2 items in slots 0 and 1 (both high-tier Tier: 2)
	inventories["Hero2"].Items[0] = &inventory.InventoryItem{ItemID: "Armor", Quantity: 1}
	inventories["Hero2"].Items[1] = &inventory.InventoryItem{ItemID: "Shield", Quantity: 1}
	// Gear multiplier = 1.0 + 2 * 0.15 = 1.3
	// Expected: 500 * 1.5 * 1.3 * 1.0 = 975
	bounty60WithGear := pm.CalculateDynamicBounty("Hero2", 500)
	if bounty60WithGear != 975 {
		t.Errorf("Expected dynamic bounty of 975 for level 60 with 2 equipped high-tier items, got: %d", bounty60WithGear)
	}

	// --- Test PATCH 8: Anti Kill Farming Reward Decay ---
	pm.mu.Lock()
	pm.bounties["Hero2"] = &BountyState{
		PlayerName:   "Hero2",
		SkullTier:    SkullWhite,
		BountyReward: 10000,
	}
	pm.mu.Unlock()

	// Claim 1: Should give 100% of the bounty (10000)
	claim1, err := pm.HandleBountyClaim("Hero1", "Hero2")
	if err != nil || claim1 != 10000 {
		t.Errorf("Expected 100%% reward (10000) on first claim, got %d (err: %v)", claim1, err)
	}

	// Restore bounty for next claim test
	pm.mu.Lock()
	pm.bounties["Hero2"] = &BountyState{
		PlayerName:   "Hero2",
		SkullTier:    SkullWhite,
		BountyReward: 10000,
	}
	pm.mu.Unlock()

	// Claim 2: Should give 25% of the bounty (2500)
	claim2, _ := pm.HandleBountyClaim("Hero1", "Hero2")
	if claim2 != 2500 {
		t.Errorf("Expected 25%% reward (2500) on second claim within 24h, got %d", claim2)
	}

	// Restore bounty for next claim test
	pm.mu.Lock()
	pm.bounties["Hero2"] = &BountyState{
		PlayerName:   "Hero2",
		SkullTier:    SkullWhite,
		BountyReward: 10000,
	}
	pm.mu.Unlock()

	// Claim 3: Should give 0% of the bounty (0)
	claim3, _ := pm.HandleBountyClaim("Hero1", "Hero2")
	if claim3 != 0 {
		t.Errorf("Expected 0%% reward on third claim within 24h, got %d", claim3)
	}
}

func TestAoLBlessingsInteraction(t *testing.T) {
	chunkMgr := movement.NewChunkManager()
	cm := combat.NewCombatManager(chunkMgr)
	bm := blessing.NewBlessingManager(nil)
	hm := housing.NewHousingManager(nil)

	inventories := make(map[string]*inventory.PlayerInventory)
	pm := NewPvPManager(nil, hm, bm, cm, inventories)
	dpm := lifecycle.NewDeathPenaltyManager(nil, bm, nil, lifecycle.NewRespawnManager(), hm, inventories, cm)
	dpm.SetPvPManager(pm)

	// Register player (level 50, so next level XP requirement is 50 * 50 * 100 = 250000)
	player := "TestPlayer"
	cm.RegisterEntity(&combat.EntityStats{ID: player, IsPlayer: true, Level: 50, Health: 100}, 100, 100)
	
	// Prepare items in inventory
	playerInv := inventory.NewPlayerInventory(player)
	playerInv.Items[0] = &inventory.InventoryItem{ItemID: "SuperSword", Quantity: 1, Durability: 100}
	playerInv.Items[4] = &inventory.InventoryItem{ItemID: "HealingPotion", Quantity: 5, Durability: 100}
	inventories[player] = playerInv

	// Helper to reset player state for tests
	resetPlayerState := func() {
		pve.SetPlayerXp(player, 300000) // generous current XP
		playerInv.Items = make(map[int]*inventory.InventoryItem)
		playerInv.Items[0] = &inventory.InventoryItem{ItemID: "SuperSword", Quantity: 1, Durability: 100}
		playerInv.Items[4] = &inventory.InventoryItem{ItemID: "HealingPotion", Quantity: 5, Durability: 100}
		// Clear all blessings
		bm.ConsumeAllBlessings(player)
	}

	// 1. Scenario: No Protection (no blessings, no AoL)
	resetPlayerState()
	// Set an attacker to ensure isPvPDeath is true
	cm.RecordAttacker(player, "Attacker")
	xpLost, _, _, drops, _, _, _, _, wasProtected, err := dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("ApplyDeathPenalties failed: %v", err)
	}
	// No protection: XP loss = 10% of level up requirement = 25000 XP, item drops enabled
	if xpLost != 25000 {
		t.Errorf("Expected 25000 XP lost (10%% of 250000), got %d", xpLost)
	}
	if wasProtected {
		t.Error("Expected wasProtected to be false on no protection")
	}

	// 2. Scenario: AoL only (no blessings, AoL equipped)
	resetPlayerState()
	// Equip AoL in slot 2 (accessory slot or any slot)
	playerInv.Items[2] = &inventory.InventoryItem{ItemID: "amulet_of_loss", Quantity: 1}
	cm.RecordAttacker(player, "Attacker")
	xpLost, _, _, drops, _, _, _, _, wasProtected, err = dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("ApplyDeathPenalties failed: %v", err)
	}
	// AoL only: XP loss reduced to 5% = 12500 XP, items protected, AoL consumed
	if xpLost != 12500 {
		t.Errorf("Expected 12500 XP lost (5%% of 250000), got %d", xpLost)
	}
	if !wasProtected {
		t.Error("Expected wasProtected to be true on AoL only")
	}
	if len(drops) > 0 {
		t.Errorf("Expected no items to drop with AoL equipped, got drops: %v", drops)
	}
	// Verify AoL is consumed
	if playerInv.IsAoLEquipped() {
		t.Error("Expected Amulet of Loss to be consumed from inventory on death")
	}

	// 3. Scenario: Blessings Active (3 blessings, no AoL)
	resetPlayerState()
	bm.GrantBlessing(player, "blessing_fire")
	bm.GrantBlessing(player, "blessing_ice")
	bm.GrantBlessing(player, "blessing_holy")
	cm.RecordAttacker(player, "Attacker")
	xpLost, _, _, drops, _, _, _, _, wasProtected, err = dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("ApplyDeathPenalties failed: %v", err)
	}
	// Blessings active: XP loss reduced by 12% per blessing.
	// 3 blessings -> 36% reduction. 10% * (1.0 - 0.36) = 6.4% XP loss = 16000 XP
	if xpLost != 16000 {
		t.Errorf("Expected 16000 XP lost (6.4%% of 250000), got %d", xpLost)
	}
	if !wasProtected {
		t.Error("Expected wasProtected to be true with active blessings")
	}
	if len(drops) > 0 {
		t.Errorf("Expected no items to drop with active blessings, got drops: %v", drops)
	}

	// 4. Scenario: AoL + Blessings (3 blessings, AoL equipped)
	resetPlayerState()
	playerInv.Items[2] = &inventory.InventoryItem{ItemID: "amulet_of_loss", Quantity: 1}
	bm.GrantBlessing(player, "blessing_fire")
	bm.GrantBlessing(player, "blessing_ice")
	bm.GrantBlessing(player, "blessing_holy")
	cm.RecordAttacker(player, "Attacker")
	xpLost, _, _, drops, _, _, _, _, wasProtected, err = dpm.ApplyDeathPenalties(player, "Fire Continent")
	if err != nil {
		t.Fatalf("ApplyDeathPenalties failed: %v", err)
	}
	// AoL + Blessings: AoL still protects items if equipped and is consumed.
	// But AoL does NOT further reduce XP loss. Determined exclusively by blessings (6.4% = 16000 XP)
	if xpLost != 16000 {
		t.Errorf("Expected 16000 XP lost (blessings-only reduction), got %d (AoL should not reduce further when blessings are active)", xpLost)
	}
	if !wasProtected {
		t.Error("Expected wasProtected to be true on AoL + Blessings")
	}
	if len(drops) > 0 {
		t.Errorf("Expected no items to drop, got drops: %v", drops)
	}
	// Verify AoL is consumed
	if playerInv.IsAoLEquipped() {
		t.Error("Expected Amulet of Loss to be consumed on death even when blessings are active")
	}
}


