package pve

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/protocol"
)

// MonsterTemplate representa as propriedades estáticas de um monstro vindas do JSON
type MonsterTemplate struct {
	ID               string  `json:"ID"`
	Name             string  `json:"Name"`
	Level            int     `json:"Level"`
	MaxHealth        float64 `json:"MaxHealth"`
	MaxMana          float64 `json:"MaxMana"`
	BaseAttack       float64 `json:"BaseAttack"`
	WeaponDamage     float64 `json:"WeaponDamage"`
	Defense          float64 `json:"Defense"`
	Resistance       float64 `json:"Resistance"`
	Accuracy         float64 `json:"Accuracy"`
	Evasion          float64 `json:"Evasion"`
	CritChance       float64 `json:"CritChance"`
	CritMultiplier   float64 `json:"CritMultiplier"`
	ArmorPenetration float64 `json:"ArmorPenetration"`
	Element          string  `json:"Element"`
	Faction          string  `json:"Faction"`
	AggroRadius      float64 `json:"AggroRadius"`
	LeashDistance    float64 `json:"LeashDistance"`
	ChaseSpeed       float64 `json:"ChaseSpeed"`
	XpReward         int     `json:"XpReward"`
	GoldMin          int     `json:"GoldMin"`
	GoldMax          int     `json:"GoldMax"`
	LootTableID      string  `json:"LootTableID"`
}

// SpawnPoint representa um ponto de desova configurado no JSON
type SpawnPoint struct {
	ID          string  `json:"ID"`
	MonsterID   string  `json:"MonsterID"`
	X           float64 `json:"X"`
	Y           float64 `json:"Y"`
	MaxAlive    int     `json:"MaxAlive"`
	SpawnRadius float64 `json:"SpawnRadius"`
	RespawnTime float64 `json:"RespawnTime"` // Segundos
}

// LootItem representa a chance de queda de um item específico
type LootItem struct {
	ItemID string  `json:"ItemID"`
	Chance float64 `json:"Chance"` // 0.0 a 1.0 (ex: 0.10 para 10%)
	MinQty int     `json:"MinQty"`
	MaxQty int     `json:"MaxQty"`
}

// LootTable representa uma tabela de drops de monstros
type LootTable struct {
	ID    string     `json:"ID"`
	Items []LootItem `json:"Items"`
}

// MonsterInstance representa a instância viva e autoritativa de um monstro no mundo
type MonsterInstance struct {
	mu             sync.RWMutex
	ID             string
	Template       MonsterTemplate
	SpawnPointID   string
	State          string // Idle, Wander, Patrol, Aggro, Chase, Attack, CastSkill, ReturnHome, Dead, Respawn
	HomeX, HomeY   float64
	X, Y           float64
	Z              int
	Stats          *combat.EntityStats
	ThreatTable    *combat.AggroTable
	CurrentTarget  string
	LastActionTime time.Time
	RespawnAt      time.Time
	WanderTargetX  float64
	WanderTargetY  float64
}

// PveManager gerencia todo o ecossistema PvE do servidor de forma centralizada e thread-safe
type PveManager struct {
	mu            sync.RWMutex
	spatialIndex  *movement.SpatialIndex
	aoiManager    *movement.AOIManager
	combatManager *combat.CombatManager
	inventories   map[string]*inventory.PlayerInventory // Ponteiro compartilhado de inventários ativos

	monsters   map[string]MonsterTemplate
	spawns     []SpawnPoint
	lootTables map[string]LootTable

	activeMonsters map[string]*MonsterInstance
	spawnCounts    map[string]int // spawnPointID -> número atual de mobs vivos
	stopChan       chan struct{}

	onPlayerLevelUp   func(playerID string, level int, stats *combat.EntityStats) // Callback para sincronização e broadcast de level up
	onMonsterKilled   func(playerID string, monsterTemplateID string)
	onItemLooted      func(playerID string, itemID string, qty int)
	onGetPartyMembers func(playerID string, x, y float64) []string
}

func (pm *PveManager) RegisterGetPartyMembersCallback(cb func(playerID string, x, y float64) []string) {
	pm.onGetPartyMembers = cb
}

// NewPveManager instancia o gerenciador PvE
func NewPveManager(si *movement.SpatialIndex, aoi *movement.AOIManager, cm *combat.CombatManager, invs map[string]*inventory.PlayerInventory) *PveManager {
	pm := &PveManager{
		spatialIndex:   si,
		aoiManager:     aoi,
		combatManager:  cm,
		inventories:    invs,
		monsters:       make(map[string]MonsterTemplate),
		lootTables:     make(map[string]LootTable),
		activeMonsters: make(map[string]*MonsterInstance),
		spawnCounts:    make(map[string]int),
		stopChan:       make(chan struct{}),
	}

	pm.loadConfigs()
	return pm
}

// RegisterLevelUpCallback registra o callback para quando um jogador passa de nível
func (pm *PveManager) RegisterLevelUpCallback(cb func(string, int, *combat.EntityStats)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onPlayerLevelUp = cb
}

// RegisterMonsterKilledCallback registra o callback para quando um monstro é derrotado
func (pm *PveManager) RegisterMonsterKilledCallback(cb func(string, string)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onMonsterKilled = cb
}

// RegisterItemLootedCallback registra o callback para quando um item é saqueado
func (pm *PveManager) RegisterItemLootedCallback(cb func(string, string, int)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onItemLooted = cb
}

// loadConfigs busca e carrega as definições de arquivos JSON de forma resiliente
func (pm *PveManager) loadConfigs() {
	paths := []string{"backend/config/", "config/", "../config/"}

	// 1. Monstros
	for _, p := range paths {
		filePath := p + "monsters.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []MonsterTemplate
				if err := json.Unmarshal(data, &list); err == nil {
					for _, m := range list {
						pm.monsters[m.ID] = m
					}
					slog.Info("Successfully loaded monsters.json", "count", len(pm.monsters), "path", filePath)
					break
				}
			}
		}
	}

	// 2. Spawn Points
	for _, p := range paths {
		filePath := p + "spawns.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []SpawnPoint
				if err := json.Unmarshal(data, &list); err == nil {
					pm.spawns = list
					slog.Info("Successfully loaded spawns.json", "count", len(pm.spawns), "path", filePath)
					break
				}
			}
		}
	}

	// 3. Loot Tables
	for _, p := range paths {
		filePath := p + "loot_tables.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []LootTable
				if err := json.Unmarshal(data, &list); err == nil {
					for _, l := range list {
						pm.lootTables[l.ID] = l
					}
					slog.Info("Successfully loaded loot_tables.json", "count", len(pm.lootTables), "path", filePath)
					break
				}
			}
		}
	}
}

// Start inicializa o ciclo contínuo da máquina de estados do PvE e re-spawns
func (pm *PveManager) Start() {
	pm.InitializeSpawns()

	// AI Tick loop a cada 500ms
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pm.TickAI()
			case <-pm.stopChan:
				return
			}
		}
	}()
}

// Stop finaliza as rotinas de segundo plano
func (pm *PveManager) Stop() {
	close(pm.stopChan)
}

// InitializeSpawns pré-popula o mapa de acordo com spawns.json
func (pm *PveManager) InitializeSpawns() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, s := range pm.spawns {
		for i := 0; i < s.MaxAlive; i++ {
			pm.spawnMonsterLocked(s, i)
		}
	}
}

// spawnMonsterLocked instancia e insere o monstro no mundo espacial e no CombatManager
func (pm *PveManager) spawnMonsterLocked(s SpawnPoint, index int) {
	template, ok := pm.monsters[s.MonsterID]
	if !ok {
		slog.Error("Cannot spawn monster: template not found", "monsterID", s.MonsterID)
		return
	}

	monsterID := fmt.Sprintf("mob_%s_%s_%d", s.ID, s.MonsterID, index)

	// Determina a posição inicial com base no SpawnRadius
	var x, y float64
	if s.SpawnRadius > 0 {
		angle := rand.Float64() * 2 * math.Pi
		r := rand.Float64() * s.SpawnRadius
		x = s.X + r*math.Cos(angle)
		y = s.Y + r*math.Sin(angle)
	} else {
		x = s.X
		y = s.Y
	}
	z := 0 // Ground floor default

	stats := &combat.EntityStats{
		ID:                 monsterID,
		Name:               template.Name,
		IsPlayer:           false,
		Faction:            template.Faction,
		Level:              template.Level,
		BaseAttack:         template.BaseAttack,
		WeaponDamage:       template.WeaponDamage,
		Defense:            template.Defense,
		Resistance:         template.Resistance,
		Accuracy:           template.Accuracy,
		Evasion:            template.Evasion,
		CritChance:         template.CritChance,
		CritMultiplier:     template.CritMultiplier,
		ArmorPenetration:   template.ArmorPenetration,
		Element:            template.Element,
		ElementAttackBonus: 0.0,
		ElementDefBonus:    0.0,
		Health:             template.MaxHealth,
		MaxHealth:          template.MaxHealth,
		Mana:               template.MaxMana,
		MaxMana:            template.MaxMana,
		LastCombatTime:     time.Now(),
	}

	// Registra o NPC no CombatManager
	pm.combatManager.RegisterEntity(stats, x, y)

	// Registra o NPC no SpatialIndex
	pm.spatialIndex.RegisterEntity(&movement.Entity{
		ID:   monsterID,
		Name: template.Name,
		X:    x,
		Y:    y,
		Z:    z,
		Type: "npc",
	})

	instance := &MonsterInstance{
		ID:             monsterID,
		Template:       template,
		SpawnPointID:   s.ID,
		State:          "Idle",
		HomeX:          s.X,
		HomeY:          s.Y,
		X:              x,
		Y:              y,
		Z:              z,
		Stats:          stats,
		ThreatTable:    combat.NewAggroTable(),
		LastActionTime: time.Now(),
	}

	pm.activeMonsters[monsterID] = instance
	slog.Info("Monster registered successfully in PvE Ecosystem", "id", monsterID, "name", template.Name, "x", x, "y", y)
}

// TickAI processa o ciclo lógico individual para todas as instâncias vivas de monstros
func (pm *PveManager) TickAI() {
	pm.mu.Lock()
	mobs := make([]*MonsterInstance, 0, len(pm.activeMonsters))
	for _, m := range pm.activeMonsters {
		mobs = append(mobs, m)
	}
	pm.mu.Unlock()

	for _, m := range mobs {
		pm.processMonsterAI(m)
	}
}

// processMonsterAI executa a máquina de estados para um monstro individual
func (pm *PveManager) processMonsterAI(m *MonsterInstance) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 1. Tratamento se estiver morto
	if m.Stats.Health <= 0 {
		if m.State != "Dead" {
			m.State = "Dead"
			m.RespawnAt = now.Add(60 * time.Second)

			// Executa recompensas de XP e Loot (Tagged loot ownership, Party split)
			go pm.handleMonsterDeath(m)

			// Remove o monstro temporariamente do mapa espacial para que suma dos AOIs
			pm.spatialIndex.RemoveEntity(m.ID)

			// Envia Despawn para vizinhos próximos na AOI
			packet := &protocol.Packet{
				Opcode:  protocol.SC_DESPAWN_ENTITY,
				Payload: []byte(m.ID),
			}
			pm.aoiManager.BroadcastMovement(m.ID, protocol.SC_DESPAWN_ENTITY, packet.Payload)
			slog.Info("Monster died, scheduling respawn", "id", m.ID, "seconds", 60)
		}
		return
	}

	// 2. Se o monstro estiver morto mas voltou à vida de alguma forma sem mudar estado
	if m.State == "Dead" {
		m.State = "Idle"
	}

	// 3. Atualizar posição no CombatManager
	pm.combatManager.UpdateEntityPosition(m.ID, m.X, m.Y)

	// 4. Fluxo da Máquina de Estados
	switch m.State {
	case "Idle":
		// Verifica se há jogadores para agrar dentro de AggroRadius
		targetPlayerID := pm.findNearestPlayer(m.X, m.Y, m.Template.AggroRadius, m.Z)
		if targetPlayerID != "" {
			m.State = "Aggro"
			m.CurrentTarget = targetPlayerID
			m.ThreatTable.AddThreat(targetPlayerID, 50.0) // Ameaça inicial de aggro
			m.LastActionTime = now
			slog.Info("Monster aggroed target player", "id", m.ID, "target", targetPlayerID)
			break
		}

		// Chance de vagar aleatoriamente (Wander) a cada 4 segundos
		if now.Sub(m.LastActionTime) > 4*time.Second {
			m.LastActionTime = now
			if rand.Float64() < 0.40 {
				m.State = "Wander"
				angle := rand.Float64() * 2 * math.Pi
				r := rand.Float64() * 3.0 // Vaga em raio de até 3 tiles
				m.WanderTargetX = m.HomeX + r*math.Cos(angle)
				m.WanderTargetY = m.HomeY + r*math.Sin(angle)
				slog.Debug("Monster started wandering", "id", m.ID, "toX", m.WanderTargetX, "toY", m.WanderTargetY)
			}
		}

	case "Wander":
		// Verifica se há jogadores para agrar durante o Wander
		targetPlayerID := pm.findNearestPlayer(m.X, m.Y, m.Template.AggroRadius, m.Z)
		if targetPlayerID != "" {
			m.State = "Aggro"
			m.CurrentTarget = targetPlayerID
			m.ThreatTable.AddThreat(targetPlayerID, 50.0)
			m.LastActionTime = now
			break
		}

		// Move em direção ao ponto de Wander
		step := m.Template.ChaseSpeed * 0.5 // Passo de 500ms
		dist := math.Hypot(m.WanderTargetX-m.X, m.WanderTargetY-m.Y)
		if dist <= step {
			m.X = m.WanderTargetX
			m.Y = m.WanderTargetY
			m.State = "Idle"
			m.LastActionTime = now
		} else {
			m.X += ((m.WanderTargetX - m.X) / dist) * step
			m.Y += ((m.WanderTargetY - m.Y) / dist) * step
			pm.broadcastNpcMovement(m.ID, m.X, m.Y, m.Z)
		}

	case "Aggro":
		// Roar ou transição rápida para Chase
		if now.Sub(m.LastActionTime) >= 500*time.Millisecond {
			m.State = "Chase"
			m.LastActionTime = now
		}

	case "Chase":
		// Valida se o alvo principal ainda é o maior ameaça na ThreatTable (Threat-based target switching)
		topTarget, hasTarget := m.ThreatTable.GetTopTarget()
		if !hasTarget {
			m.State = "ReturnHome"
			m.LastActionTime = now
			break
		}
		m.CurrentTarget = topTarget

		// Leashing check: Se ultrapassar LeashDistance da Home, reseta aggro
		distToHome := math.Hypot(m.X-m.HomeX, m.Y-m.HomeY)
		if distToHome > m.Template.LeashDistance {
			m.State = "ReturnHome"
			m.ThreatTable.ClearThreat()
			m.CurrentTarget = ""
			m.LastActionTime = now
			slog.Info("Monster leashed! Returning home", "id", m.ID, "distToHome", distToHome)
			break
		}

		// Obtém posição do jogador alvo
		px, py, exists := pm.combatManager.GetEntityPosition(m.CurrentTarget)
		targetStats, targetExists := pm.combatManager.GetEntityStats(m.CurrentTarget)
		if !exists || !targetExists || targetStats.Health <= 0 {
			m.ThreatTable.RemovePlayer(m.CurrentTarget)
			break
		}

		// Range check para ataque Melee ou Skill
		distToPlayer := math.Hypot(m.X-px, m.Y-py)

		// Decisão se ataca ou se continua a perseguir
		if distToPlayer <= 1.5 {
			m.State = "Attack"
			m.LastActionTime = now
			break
		} else if m.Template.ID == "dragon_boss" && distToPlayer <= 6.0 && rand.Float64() < 0.30 {
			m.State = "CastSkill"
			m.LastActionTime = now
			break
		}

		// Move em direção ao jogador alvo
		step := m.Template.ChaseSpeed * 0.5
		m.X += ((px - m.X) / distToPlayer) * step
		m.Y += ((py - m.Y) / distToPlayer) * step
		pm.broadcastNpcMovement(m.ID, m.X, m.Y, m.Z)

	case "Attack":
		// Melee attack
		topTarget, hasTarget := m.ThreatTable.GetTopTarget()
		if !hasTarget {
			m.State = "ReturnHome"
			break
		}
		m.CurrentTarget = topTarget

		px, py, exists := pm.combatManager.GetEntityPosition(m.CurrentTarget)
		targetStats, targetExists := pm.combatManager.GetEntityStats(m.CurrentTarget)
		if !exists || !targetExists || targetStats.Health <= 0 {
			m.ThreatTable.RemovePlayer(m.CurrentTarget)
			m.State = "Chase"
			break
		}

		distToPlayer := math.Hypot(m.X-px, m.Y-py)
		if distToPlayer > 1.8 {
			m.State = "Chase"
			break
		}

		// Executa ataque autoritativo
		_, _, _, err := pm.combatManager.ProcessAttackRequest(m.ID, m.CurrentTarget, "sword")
		if err == nil {
			slog.Debug("Monster auto-attacked player", "id", m.ID, "target", m.CurrentTarget)
		}

		m.State = "Chase"
		m.LastActionTime = now

	case "CastSkill":
		// Conjura skill de monstro (ex: Lorde Dragão lança Fireball)
		topTarget, hasTarget := m.ThreatTable.GetTopTarget()
		if !hasTarget {
			m.State = "ReturnHome"
			break
		}
		m.CurrentTarget = topTarget

		px, py, exists := pm.combatManager.GetEntityPosition(m.CurrentTarget)
		targetStats, targetExists := pm.combatManager.GetEntityStats(m.CurrentTarget)
		if !exists || !targetExists || targetStats.Health <= 0 {
			m.ThreatTable.RemovePlayer(m.CurrentTarget)
			m.State = "Chase"
			break
		}

		// Tenta castar Fireball (ID 2) ou Slash (ID 1)
		skillID := uint32(1) // Slash
		if m.Template.ID == "dragon_boss" {
			skillID = 2 // Fireball
		}

		_, err := pm.combatManager.ProcessCastSkillRequest(m.ID, skillID, m.CurrentTarget, px, py)
		if err == nil {
			slog.Info("Monster cast skill successfully", "id", m.ID, "skill", skillID, "target", m.CurrentTarget)
		}

		m.State = "Chase"
		m.LastActionTime = now

	case "ReturnHome":
		// Cura completamente o monstro enquanto retorna para ficar invulnerável
		m.Stats.Health = m.Template.MaxHealth

		distToHome := math.Hypot(m.X-m.HomeX, m.Y-m.HomeY)
		step := m.Template.ChaseSpeed * 0.5
		if distToHome <= step {
			m.X = m.HomeX
			m.Y = m.HomeY
			m.State = "Idle"
			m.ThreatTable.ClearThreat()
			m.CurrentTarget = ""
			m.LastActionTime = now
			slog.Info("Monster safely returned home", "id", m.ID)
		} else {
			m.X += ((m.HomeX - m.X) / distToHome) * step
			m.Y += ((m.HomeY - m.Y) / distToHome) * step
			pm.broadcastNpcMovement(m.ID, m.X, m.Y, m.Z)
		}
	}
}

// TickRespawn processa o respawn de monstros mortos
func (pm *PveManager) TickRespawn() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for id, m := range pm.activeMonsters {
		m.mu.Lock()
		if m.State == "Dead" && now.After(m.RespawnAt) {
			// Revive o monstro
			m.Stats.Health = m.Template.MaxHealth
			m.Stats.Mana = m.Template.MaxMana
			m.X = m.HomeX
			m.Y = m.HomeY
			m.State = "Idle"
			m.ThreatTable.ClearThreat()
			m.CurrentTarget = ""
			m.LastActionTime = now

			// Registra no SpatialIndex novamente para re-spawn na tela dos jogadores
			pm.spatialIndex.RegisterEntity(&movement.Entity{
				ID:   m.ID,
				Name: m.Template.Name,
				X:    m.X,
				Y:    m.Y,
				Z:    m.Z,
				Type: "npc",
			})

			slog.Info("Monster respawned and returned to world active status", "id", id, "name", m.Template.Name)
		}
		m.mu.Unlock()
	}
}

// findNearestPlayer busca o jogador mais próximo na área visível
func (pm *PveManager) findNearestPlayer(x, y float64, radius float64, z int) string {
	entities := pm.spatialIndex.GetEntitiesInRegion(x, y, radius, z)

	var nearestID string
	minDist := radius + 1.0

	for _, ent := range entities {
		if ent.Type == "player" {
			dist := math.Hypot(ent.X-x, ent.Y-y)
			if dist < minDist {
				minDist = dist
				nearestID = ent.ID
			}
		}
	}

	return nearestID
}

// broadcastNpcMovement envia a nova coordenada para os jogadores que observam o NPC
func (pm *PveManager) broadcastNpcMovement(npcID string, x, y float64, z int) {
	pm.spatialIndex.UpdateEntityPosition(npcID, x, y, z)

	payload, _ := json.Marshal(struct {
		PlayerID string  `json:"id"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		Dir      uint8   `json:"dir"`
	}{
		PlayerID: npcID,
		X:        x,
		Y:        y,
		Dir:      1,
	})

	pm.aoiManager.BroadcastMove(npcID, x, y, z, payload)
}

// handleMonsterDeath calcula XP, redistribuição em grupo, drops e itens adicionados no inventário
func (pm *PveManager) handleMonsterDeath(m *MonsterInstance) {
	// 1. Identificar jogadores que participaram da luta próximos
	topThreatPlayerID, hasTarget := m.ThreatTable.GetTopTarget()
	if !hasTarget {
		return
	}

	// Coleta jogador elegível para XP/loot.
	// Build stabilization: usa o top threat como participante válido sem acessar internals do SpatialIndex.
	eligiblePlayers := []string{topThreatPlayerID}

	if len(eligiblePlayers) == 0 {
		eligiblePlayers = append(eligiblePlayers, topThreatPlayerID)
	}

	// Expand to party members within 30-tile radius if any eligible player is in a party
	if pm.onGetPartyMembers != nil {
		partySet := make(map[string]bool)
		for _, pID := range eligiblePlayers {
			partySet[pID] = true
			members := pm.onGetPartyMembers(pID, m.X, m.Y)
			for _, member := range members {
				partySet[member] = true
			}
		}
		expandedPlayers := []string{}
		for pID := range partySet {
			expandedPlayers = append(expandedPlayers, pID)
		}
		eligiblePlayers = expandedPlayers
	}

	// 2. XP Split Support
	rawXp := m.Template.XpReward
	xpPerPlayer := int64(rawXp)
	if len(eligiblePlayers) > 1 {
		// Multiplicador leve para incentivar grupo (+10% de XP por membro adicional)
		bonusCoeff := 1.0 + 0.10*float64(len(eligiblePlayers)-1)
		xpPerPlayer = int64(math.Round((float64(rawXp) * bonusCoeff) / float64(len(eligiblePlayers))))
		if xpPerPlayer < 1 {
			xpPerPlayer = 1
		}
	}

	// Distribui XP e verifica level up
	for _, pID := range eligiblePlayers {
		pm.awardXp(pID, xpPerPlayer)
	}

	// 3. Gold & Items Drops (Tagged loot ownership ao Top Threat Player)
	playerInv, exists := pm.inventories[topThreatPlayerID]
	if !exists {
		slog.Warn("Top threat player inventory not loaded. Drops discarded", "player", topThreatPlayerID)
		return
	}

	// Sorteia Gold
	droppedGold := 0
	if m.Template.GoldMax > m.Template.GoldMin {
		droppedGold = m.Template.GoldMin + rand.Intn(m.Template.GoldMax-m.Template.GoldMin+1)
	} else {
		droppedGold = m.Template.GoldMin
	}

	// Envia notificação de Gold obtido (representado via log ou mensagem do sistema)
	slog.Info("Player looted gold from monster", "player", topThreatPlayerID, "monster", m.Template.Name, "gold", droppedGold)

	// Sorteia itens da Loot Table (Weighted drop tables)
	lootTable, hasLoot := pm.lootTables[m.Template.LootTableID]
	if hasLoot {
		for _, drop := range lootTable.Items {
			roll := rand.Float64()
			if roll <= drop.Chance {
				// Drop bem sucedido! Determina quantidade
				qty := drop.MinQty
				if drop.MaxQty > drop.MinQty {
					qty = drop.MinQty + rand.Intn(drop.MaxQty-drop.MinQty+1)
				}

				// Adiciona ao inventário de forma segura
				success := playerInv.AddItem(drop.ItemID, qty)
				if success {
					slog.Info("Rare/Weighted drop added to player inventory", "player", topThreatPlayerID, "item", drop.ItemID, "qty", qty)

					pm.mu.RLock()
					lootedCb := pm.onItemLooted
					pm.mu.RUnlock()
					if lootedCb != nil {
						lootedCb(topThreatPlayerID, drop.ItemID, qty)
					}
				} else {
					slog.Warn("Player backpack full. Item drop lost", "player", topThreatPlayerID, "item", drop.ItemID)
				}
			}
		}
	}

	// Aciona callback para quest hooks (PATCH 5)
	pm.mu.RLock()
	killedCb := pm.onMonsterKilled
	pm.mu.RUnlock()
	if killedCb != nil {
		for _, pID := range eligiblePlayers {
			killedCb(pID, m.Template.ID)
		}
	}
}

// awardXp concede experiência a um jogador de forma autoritativa e calcula level-ups acumulativos
func (pm *PveManager) awardXp(playerID string, xp int64) {
	playerInv, existsInv := pm.inventories[playerID]
	pStats, existsStats := pm.combatManager.GetEntityStats(playerID)
	if !existsInv || !existsStats {
		return
	}

	// Usamos a BaseStats do inventário para controlar o XP atual de forma persistente
	// Como a tabela character possui 'experience', podemos armazenar lá temporariamente
	// Vamos usar o campo LastCombatTime ou uma propriedade em memória estática, ou carregar dinamicamente.
	// Vamos guardar o XP acumulado de forma local em memória e fazer o level up

	// Vamos definir a fórmula de XP necessário: Level * Level * 100
	xpNeeded := int64(pStats.Level * pStats.Level * 100)

	// Carrega XP atual em memória
	// Como o banco characters já grava a coluna 'experience', vamos adicionar o XP
	// e salvar. Para fazer isso, guardamos no BaseStats (podemos ler ou embutir um contador temporário).
	// Vamos criar um mapeamento temporário de XP em memória para os jogadores logados!
	pveXpMu.Lock()
	currentXp := playerXpRegistry[playerID]
	currentXp += xp

	leveledUp := false
	for currentXp >= xpNeeded {
		currentXp -= xpNeeded
		pStats.Level++
		playerInv.BaseStats.Level++

		// Incremento estrito de Atributos de Combate no Level Up
		pStats.MaxHealth += 20.0
		pStats.Health = pStats.MaxHealth
		pStats.MaxMana += 5.0
		pStats.Mana = pStats.MaxMana
		pStats.BaseAttack += 2.0

		playerInv.BaseStats.MaxHealth = pStats.MaxHealth
		playerInv.BaseStats.Health = pStats.Health
		playerInv.BaseStats.MaxMana = pStats.MaxMana
		playerInv.BaseStats.Mana = pStats.Mana
		playerInv.BaseStats.BaseAttack = pStats.BaseAttack

		leveledUp = true
		slog.Info("PLAYER LEVEL UP!", "player", playerID, "newLevel", pStats.Level)
		xpNeeded = int64(pStats.Level * pStats.Level * 100)
	}
	playerXpRegistry[playerID] = currentXp
	pveXpMu.Unlock()

	// Força persistência marcando como dirty
	playerInv.SetDirty(true)

	if leveledUp && pm.onPlayerLevelUp != nil {
		pm.onPlayerLevelUp(playerID, pStats.Level, pStats)
	}
}

// Variáveis estáticas de controle de XP em memória para evitar quebras de retrocompatibilidade com EntityStats
var (
	pveXpMu          sync.Mutex
	playerXpRegistry = make(map[string]int64)
)

// GetPlayerXp retorna o XP em memória do jogador
func GetPlayerXp(playerID string) int64 {
	pveXpMu.Lock()
	defer pveXpMu.Unlock()
	return playerXpRegistry[playerID]
}

// SetPlayerXp inicializa o XP do jogador vindo da persistência
func SetPlayerXp(playerID string, xp int64) {
	pveXpMu.Lock()
	defer pveXpMu.Unlock()
	playerXpRegistry[playerID] = xp
}

// LootRollResult describes one backend-authoritative roll against a loaded loot table.
// It is intentionally separated from inventory mutation so callers can apply capacity,
// persistence, audit logging and duplicate-loot guards in their own authoritative flow.
type LootRollResult struct {
	ItemID   string
	Quantity int
	Chance   float64
	Roll     float64
	Dropped  bool
	Reason   string
}

// RollLootTable rolls a loaded loot table without mutating inventory state.
// Returns false when the table is missing, allowing callers to log and fall back safely.
func (pm *PveManager) RollLootTable(tableID string) ([]LootRollResult, bool) {
	if pm == nil || tableID == "" {
		return nil, false
	}

	lootTable, exists := pm.lootTables[tableID]
	if !exists {
		return nil, false
	}

	results := make([]LootRollResult, 0, len(lootTable.Items))
	for _, drop := range lootTable.Items {
		result := LootRollResult{
			ItemID:  drop.ItemID,
			Chance:  drop.Chance,
			Dropped: false,
			Reason:  "roll_missed",
		}

		if drop.ItemID == "" {
			result.Reason = "invalid_item_id"
			results = append(results, result)
			continue
		}

		if drop.Chance <= 0 {
			result.Reason = "zero_or_negative_chance"
			results = append(results, result)
			continue
		}

		minQty := drop.MinQty
		if minQty <= 0 {
			minQty = 1
		}

		maxQty := drop.MaxQty
		if maxQty < minQty {
			maxQty = minQty
		}

		result.Roll = rand.Float64()
		if result.Roll <= drop.Chance {
			qty := minQty
			if maxQty > minQty {
				qty = rand.Intn(maxQty-minQty+1) + minQty
			}

			result.Quantity = qty
			result.Dropped = true
			result.Reason = "dropped"
		}

		results = append(results, result)
	}

	return results, true
}
