package dungeon

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/protocol"
)

// Configurações de Dungeons
type DungeonTemplate struct {
	ID            string
	Name          string
	RequiredLevel int
	Duration      time.Duration
	Bosses        []string
}

// Checkpoint representa um marco salvo de progressão na masmorra
type Checkpoint struct {
	ID   string
	Name string
	X, Y float64
}

// BossState representa o estado interno do boss de IA avançada
type BossState struct {
	ID                 string
	TemplateID         string
	Name               string
	X, Y               float64
	HomeX, HomeY       float64
	Stats              *combat.EntityStats
	Phase              uint32
	IsEnraged          bool
	InCombatSince      time.Time
	LastPulseTime      time.Time
	AntiResetTimer     *time.Time // PATCH 2: prevenção contra exploit de reset imediato
	SpawnedAdds        []string
	TargetID           string
	TelegraphActive    bool
	TelegraphX, TelegraphY float64
	TelegraphTime      time.Time
}

// DungeonInstance representa uma masmorra instanciada ativa
type DungeonInstance struct {
	mu             sync.RWMutex
	ID             string
	TemplateID     string
	Mode           string // solo, party, raid
	StartTime      time.Time
	EndTime        time.Time
	Players        map[string]bool
	CheckpointID   string
	PositionOffset float64 // PATCH 1: Isolamento de coordenada espacial
	Bosses         map[string]*BossState
	IsDestroyed    bool
}

// WorldBossState gerencia o boss mundial persistente e compartilhado
type WorldBossState struct {
	mu             sync.RWMutex
	ID             string
	Name           string
	X, Y           float64
	Active         bool
	Stats          *combat.EntityStats
	InCombatSince      time.Time
	LastPulseTime      time.Time
	NextSpawnTime  time.Time
	ThreatTable    map[string]float64
	HealTable      map[string]float64
}

// LootItem representação de loot associado à masmorra
type LootItem struct {
	ItemID   string
	Quantity int
	Claimed  bool
}

// LootReservation armazena itens temporariamente reservados para jogadores elegíveis
type LootReservation struct {
	InstanceID string
	PlayerID   string
	Items      []LootItem
	ExpiresAt  time.Time
}

// DungeonManager orquestra todo o sistema de PvE end-game
type DungeonManager struct {
	mu                sync.RWMutex
	db                *sql.DB
	combatManager     *combat.CombatManager
	spatialIndex      *movement.SpatialIndex
	aoiManager        *movement.AOIManager
	inventories       map[string]*inventory.PlayerInventory
	
	templates         map[string]DungeonTemplate
	instances         map[string]*DungeonInstance
	playerInstanceMap map[string]string // playerID -> instanceID
	instanceCounter   int
	
	lootReservations  map[string]*LootReservation // playerID -> loot
	worldBoss         *WorldBossState
	
	stopChan          chan struct{}
	onGetPartyMembers func(playerID string, x, y float64) []string
}

func NewDungeonManager(db *sql.DB, cm *combat.CombatManager, si *movement.SpatialIndex, aoi *movement.AOIManager, invs map[string]*inventory.PlayerInventory) *DungeonManager {
	dm := &DungeonManager{
		db:                db,
		combatManager:     cm,
		spatialIndex:      si,
		aoiManager:        aoi,
		inventories:       invs,
		templates:         make(map[string]DungeonTemplate),
		instances:         make(map[string]*DungeonInstance),
		playerInstanceMap: make(map[string]string),
		lootReservations:  make(map[string]*LootReservation),
		stopChan:          make(chan struct{}),
	}
	dm.loadTemplates()
	dm.initializeWorldBoss()
	return dm
}

func (dm *DungeonManager) RegisterGetPartyMembersCallback(cb func(playerID string, x, y float64) []string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.onGetPartyMembers = cb
}

func (dm *DungeonManager) loadTemplates() {
	dm.templates["crypt_of_shadows"] = DungeonTemplate{
		ID:            "crypt_of_shadows",
		Name:          "Crypt of Shadows",
		RequiredLevel: 45,
		Duration:      30 * time.Minute,
		Bosses:        []string{"shadow_lord", "abyss_terror"},
	}
	dm.templates["dragon_lair"] = DungeonTemplate{
		ID:            "dragon_lair",
		Name:          "Dragon's Lair",
		RequiredLevel: 45,
		Duration:      60 * time.Minute,
		Bosses:        []string{"dragon_king"},
	}
}

func (dm *DungeonManager) initializeWorldBoss() {
	dm.worldBoss = &WorldBossState{
		ID:            "world_boss_behemoth",
		Name:          "World Boss Behemoth",
		X:             500.0,
		Y:             500.0,
		Active:        false,
		NextSpawnTime: time.Now().Add(5 * time.Second), // Primeiro spawn rápido para fins de teste
		ThreatTable:   make(map[string]float64),
		HealTable:     make(map[string]float64),
	}
}

// Start inicia o ciclo lógico em segundo plano para tick de IA de bosses, encerramento de masmorras e re-spawns do World Boss
func (dm *DungeonManager) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				dm.TickDungeons()
				dm.TickWorldBoss()
			case <-dm.stopChan:
				return
			}
		}
	}()
}

func (dm *DungeonManager) Stop() {
	close(dm.stopChan)
}

// EnterDungeon lida com a entrada autoritativa de jogadores em masmorras instanciadas, respeitando o PATCH 4 (Disconnect Recovery)
func (dm *DungeonManager) EnterDungeon(playerID string, dungeonID string, mode string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	tmpl, exists := dm.templates[dungeonID]
	if !exists {
		return fmt.Errorf("masmorra não encontrada")
	}

	// Valida nível do jogador
	playerInv, existsInv := dm.inventories[playerID]
	if !existsInv || playerInv.BaseStats.Level < tmpl.RequiredLevel {
		return fmt.Errorf("nível necessário: %d", tmpl.RequiredLevel)
	}

	// PATCH 4: Disconnect Recovery. Verifica se o jogador já possui uma instância ativa correspondente
	if activeInstID, ok := dm.playerInstanceMap[playerID]; ok {
		if inst, found := dm.instances[activeInstID]; found && !inst.IsDestroyed {
			slog.Info("Disconnect recovery triggered. Re-routing player to existing dungeon instance", "player", playerID, "instance", activeInstID)
			dm.teleportPlayerToInstanceLocked(playerID, inst)
			return nil
		}
	}

	// Verifica se jogador é lider de grupo para aplicar modo correto de entrada do grupo
	var groupPlayers []string
	if mode != "solo" && dm.onGetPartyMembers != nil {
		groupPlayers = dm.onGetPartyMembers(playerID, 0, 0)
	}
	if len(groupPlayers) == 0 {
		groupPlayers = []string{playerID}
	}

	// Cria nova instância
	dm.instanceCounter++
	offset := float64(dm.instanceCounter * 2000) // PATCH 1: Isolamento de coordenadas
	
	instID := fmt.Sprintf("inst_%s_%d", dungeonID, dm.instanceCounter)
	inst := &DungeonInstance{
		ID:             instID,
		TemplateID:     dungeonID,
		Mode:           mode,
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(tmpl.Duration),
		Players:        make(map[string]bool),
		CheckpointID:   "Entrance",
		PositionOffset: offset,
		Bosses:         make(map[string]*BossState),
	}

	// Carrega Lockout
	for _, pID := range groupPlayers {
		locked, err := dm.checkRaidLockout(pID, tmpl.Bosses[0])
		if err == nil && locked {
			return fmt.Errorf("o jogador %s está sob lockout ativo nesta masmorra", pID)
		}
	}

	// Registra jogadores
	for _, pID := range groupPlayers {
		inst.Players[pID] = true
		dm.playerInstanceMap[pID] = instID
	}

	// Cria Bosses para essa instância com offset
	for _, bossType := range tmpl.Bosses {
		bossID := fmt.Sprintf("boss_%s_%s", instID, bossType)
		bossName := "Lorde das Sombras"
		maxHP := 5000.0
		baseAtt := 50.0
		x, y := 150.0, 150.0 // Posição local padrão

		if bossType == "abyss_terror" {
			bossName = "Terror do Abismo"
			maxHP = 8000.0
			baseAtt = 65.0
			x, y = 200.0, 200.0
		} else if bossType == "dragon_king" {
			bossName = "Rei Dragão Vermelho"
			maxHP = 15000.0
			baseAtt = 90.0
			x, y = 300.0, 300.0
		}

		// Ajuste de vida com base no modo
		if mode == "party" {
			maxHP *= 3.0
			baseAtt *= 1.2
		} else if mode == "raid" {
			maxHP *= 8.0
			baseAtt *= 1.5
		}

		stats := &combat.EntityStats{
			ID:               bossID,
			Name:             bossName,
			IsPlayer:         false,
			Faction:          "Monsters",
			Level:            50,
			BaseAttack:       baseAtt,
			WeaponDamage:     20.0,
			Defense:          50.0,
			Resistance:       30.0,
			Accuracy:         99.0,
			Evasion:          10.0,
			CritChance:       0.15,
			CritMultiplier:   1.8,
			ArmorPenetration: 0.20,
			Element:          "Dark",
			Health:           maxHP,
			MaxHealth:        maxHP,
			Mana:             500.0,
			MaxMana:          500.0,
			LastCombatTime:   time.Now(),
		}

		// Registra no combate com coordenadas de offset isoladas (PATCH 1)
		dm.combatManager.RegisterEntity(stats, x+offset, y+offset)

		// Insere NPC no SpatialIndex com offset
		dm.spatialIndex.RegisterEntity(&movement.Entity{
			ID:   bossID,
			Name: bossName,
			X:    x + offset,
			Y:    y + offset,
			Z:    0,
			Type: "npc",
		})

		inst.Bosses[bossType] = &BossState{
			ID:         bossID,
			TemplateID: bossType,
			Name:       bossName,
			X:          x + offset,
			Y:          y + offset,
			HomeX:      x + offset,
			HomeY:      y + offset,
			Stats:      stats,
			Phase:      1,
		}
	}

	dm.instances[instID] = inst
	slog.Info("Created isolated dungeon instance", "id", instID, "mode", mode, "offset", offset)

	// Teleporta os jogadores do grupo para o ponto inicial da instância
	for _, pID := range groupPlayers {
		dm.teleportPlayerToInstanceLocked(pID, inst)
	}

	return nil
}

func (dm *DungeonManager) teleportPlayerToInstanceLocked(playerID string, inst *DungeonInstance) {
	// Teleporta para a coordenada com offset correspondente ao checkpoint atual
	baseX, baseY := 100.0, 100.0 // Entrada padrão
	checkpoint, err := dm.loadCheckpoint(playerID, inst.TemplateID)
	if err == nil && checkpoint != "" {
		inst.CheckpointID = checkpoint
		if checkpoint == "Antechamber" {
			baseX, baseY = 140.0, 140.0
		} else if checkpoint == "Throne Room" {
			baseX, baseY = 190.0, 190.0
		}
	}

	targetX := baseX + inst.PositionOffset
	targetY := baseY + inst.PositionOffset

	// Atualiza Spatial Index
	dm.spatialIndex.UpdateEntityPosition(playerID, targetX, targetY, 0)
	dm.combatManager.UpdateEntityPosition(playerID, targetX, targetY)

	// Envia pacote de sincronização com estado da instância
	if _, exists := dm.inventories[playerID]; exists {
		// Envia mensagem informativa
		msg := fmt.Sprintf("Entrou em: %s (%s). Tempo limite: %.0f min.", inst.TemplateID, inst.Mode, inst.EndTime.Sub(time.Now()).Minutes())
		packet := &protocol.Packet{
			Opcode:  protocol.SC_CHAT_MESSAGE,
			Payload: protocol.EncodeChatMessage(0, "System", msg),
		}
		dm.sendPacketToPlayer(playerID, packet)

		// Envia estado da masmorra
		statePack := &protocol.Packet{
			Opcode:  protocol.SC_DUNGEON_STATE,
			Payload: protocol.EncodeDungeonState(inst.ID, inst.TemplateID, inst.CheckpointID, inst.EndTime.Sub(time.Now()).Seconds(), 1),
		}
		dm.sendPacketToPlayer(playerID, statePack)
	}
}

// TickDungeons varre todas as instâncias ativas e executa lógica de tempo e IA dos bosses
func (dm *DungeonManager) TickDungeons() {
	dm.mu.Lock()
	activeInsts := make([]*DungeonInstance, 0, len(dm.instances))
	for _, inst := range dm.instances {
		if !inst.IsDestroyed {
			activeInsts = append(activeInsts, inst)
		}
	}
	dm.mu.Unlock()

	now := time.Now()

	for _, inst := range activeInsts {
		inst.mu.Lock()

		// 1. Verifica tempo expirado
		if now.After(inst.EndTime) {
			slog.Info("Dungeon instance expired. Disbanding", "id", inst.ID)
			inst.IsDestroyed = true
			inst.mu.Unlock()
			dm.destroyInstance(inst.ID)
			continue
		}

		// 2. Processa IA de cada Boss vivo na masmorra
		for _, boss := range inst.Bosses {
			dm.processBossAI(inst, boss)
		}

		inst.mu.Unlock()
	}
}

func (dm *DungeonManager) processBossAI(inst *DungeonInstance, boss *BossState) {
	if boss.Stats.Health <= 0 {
		// Boss derrotado de forma autoritativa
		if boss.Phase != 99 {
			boss.Phase = 99 // Estado morto definitivo
			go dm.handleBossDefeat(inst, boss)
		}
		return
	}

	now := time.Now()

	// Wipe detection & Anti-Reset Exploit (PATCH 2)
	playersAlive := false
	var nearestPlayerID string
	minDist := 100.0

	for pID := range inst.Players {
		pStats, exists := dm.combatManager.GetEntityStats(pID)
		if exists && pStats.Health > 0 {
			px, py, _ := dm.combatManager.GetEntityPosition(pID)
			dist := math.Hypot(px-boss.HomeX, py-boss.HomeY)
			if dist <= 45.0 { // Jogador vivo dentro da arena de combate do boss
				playersAlive = true
				if dist < minDist {
					minDist = dist
					nearestPlayerID = pID
				}
			}
		}
	}

	if !playersAlive {
		// PATCH 2: prevenção contra reset imediato por instabilidade de rede ou mortes rápidas.
		// Espera 10 segundos antes de resetar completamente o Boss
		if boss.AntiResetTimer == nil {
			t := now.Add(10 * time.Second)
			boss.AntiResetTimer = &t
			slog.Info("All players wiped/left boss room. Starting anti-reset timer", "boss", boss.ID, "duration", "10s")
		} else if now.After(*boss.AntiResetTimer) {
			slog.Info("Anti-reset timer expired. Resetting boss stats completely", "boss", boss.ID)
			boss.Stats.Health = boss.Stats.MaxHealth
			boss.Phase = 1
			boss.IsEnraged = false
			boss.InCombatSince = time.Time{}
			boss.AntiResetTimer = nil
			boss.TelegraphActive = false
			// Remove Adds remanescentes
			for _, addID := range boss.SpawnedAdds {
				dm.combatManager.DeregisterEntity(addID)
				dm.spatialIndex.RemoveEntity(addID)
			}
			boss.SpawnedAdds = nil
		}
		return
	}

	// Jogadores ativos lutando com o boss
	if boss.AntiResetTimer != nil {
		boss.AntiResetTimer = nil // Cancela reset pendente
		slog.Info("Players re-engaged boss. Anti-reset timer cancelled", "boss", boss.ID)
	}

	if boss.InCombatSince.IsZero() {
		boss.InCombatSince = now
	}

	// 1. Enrage Timer (PATCH 5: anti-stall)
	if !boss.IsEnraged && now.Sub(boss.InCombatSince) > 5*time.Minute {
		boss.IsEnraged = true
		boss.Stats.BaseAttack *= 5.0 // Dano massivo contra stall
		slog.Warn("BOSS ENRAGED! Damage increased by 500% to prevent stalls", "boss", boss.ID)
		dm.broadcastInstanceMessage(inst, fmt.Sprintf("%s enfureceu! O dano dele aumentou em 500%%!", boss.Stats.Name))
	}

	// Ataque anti-stalling periódico do enrage
	if boss.IsEnraged && now.Sub(boss.LastPulseTime) > 3*time.Second {
		boss.LastPulseTime = now
		// Pulso de dano inevitável para todos na instância
		for pID := range inst.Players {
			pStats, exists := dm.combatManager.GetEntityStats(pID)
			if exists && pStats.Health > 0 {
				pStats.Health -= 150.0
				if pStats.Health < 0 {
					pStats.Health = 0
				}
				// Envia notificação visual via efeito/dano
				dm.sendCombatDamageSync(pID, boss.ID, pID, 150.0, false, "Enrage Shadow Pulse")
			}
		}
	}

	// 2. Phase transitions (Phase 2 em 70% e Phase 3 em 35% HP)
	pctHP := boss.Stats.Health / boss.Stats.MaxHealth
	if boss.Phase == 1 && pctHP <= 0.70 {
		boss.Phase = 2
		dm.triggerBossPhaseTransition(inst, boss)
	} else if boss.Phase == 2 && pctHP <= 0.35 {
		boss.Phase = 3
		dm.triggerBossPhaseTransition(inst, boss)
	}

	// 3. Processa ataque de telegrafo ativo
	if boss.TelegraphActive && now.After(boss.TelegraphTime) {
		boss.TelegraphActive = false
		// Aplica dano em área telegrafado
		radius := 8.0
		for pID := range inst.Players {
			pStats, exists := dm.combatManager.GetEntityStats(pID)
			if exists && pStats.Health > 0 {
				px, py, _ := dm.combatManager.GetEntityPosition(pID)
				if math.Hypot(px-boss.TelegraphX, py-boss.TelegraphY) <= radius {
					dmg := 200.0
					pStats.Health -= dmg
					if pStats.Health < 0 {
						pStats.Health = 0
					}
					dm.sendCombatDamageSync(pID, boss.ID, pID, dmg, true, "Telegraphed Dark Crash")
				}
			}
		}
		slog.Info("Telegraphed attack hit the ground", "boss", boss.ID, "x", boss.TelegraphX, "y", boss.TelegraphY)
	}

	// 4. Inteligência de Ataque e Skill aleatória
	if nearestPlayerID != "" {
		boss.TargetID = nearestPlayerID
		if rand.Float64() < 0.25 && !boss.TelegraphActive {
			// Ativa telegrafo de ataque AoE
			boss.TelegraphActive = true
			boss.TelegraphX, boss.TelegraphY, _ = dm.combatManager.GetEntityPosition(nearestPlayerID)
			boss.TelegraphTime = now.Add(2500 * time.Millisecond) // 2.5 segundos para fugir!

			// Envia aviso via protocolo oficial para a interface (SC_BOSS_AI_TELEGRAPH)
			telegraphPack := &protocol.Packet{
				Opcode:  protocol.SC_BOSS_AI_TELEGRAPH,
				Payload: protocol.EncodeBossAITelegraph(boss.ID, boss.TelegraphX, boss.TelegraphY, 8.0, 2.5),
			}
			dm.broadcastInstancePacket(inst, telegraphPack)
		} else if rand.Float64() < 0.40 {
			// Ataque básico contra alvo mais próximo
			dm.combatManager.ProcessAttackRequest(boss.ID, boss.TargetID, "sword")
		}
	}
}

func (dm *DungeonManager) triggerBossPhaseTransition(inst *DungeonInstance, boss *BossState) {
	slog.Info("Boss transitioned phase!", "boss", boss.ID, "newPhase", boss.Phase)
	
	// Sincroniza via rede
	phasePack := &protocol.Packet{
		Opcode:  protocol.SC_BOSS_AI_PHASE,
		Payload: protocol.EncodeBossAIPhase(boss.ID, boss.Phase),
	}
	dm.broadcastInstancePacket(inst, phasePack)
	dm.broadcastInstanceMessage(inst, fmt.Sprintf("%s entrou na Fase %d!", boss.Stats.Name, boss.Phase))

	// Invoca 2 adds (mobs ajudantes)
	for i := 0; i < 2; i++ {
		addID := fmt.Sprintf("boss_add_%s_%d_%d", boss.ID, boss.Phase, i)
		x := boss.HomeX + (rand.Float64()*10 - 5)
		y := boss.HomeY + (rand.Float64()*10 - 5)

		stats := &combat.EntityStats{
			ID:               addID,
			Name:             fmt.Sprintf("Servo de IA de %s", boss.Stats.Name),
			IsPlayer:         false,
			Faction:          "Monsters",
			Level:            45,
			BaseAttack:       25.0,
			WeaponDamage:     10.0,
			Defense:          20.0,
			Resistance:       10.0,
			Accuracy:         95.0,
			Evasion:          5.0,
			CritChance:       0.05,
			CritMultiplier:   1.5,
			ArmorPenetration: 0.10,
			Element:          "Dark",
			Health:           500.0,
			MaxHealth:        500.0,
			Mana:             100.0,
			MaxMana:          100.0,
			LastCombatTime:   time.Now(),
		}

		dm.combatManager.RegisterEntity(stats, x, y)
		dm.spatialIndex.RegisterEntity(&movement.Entity{
			ID:   addID,
			Name: stats.Name,
			X:    x,
			Y:    y,
			Z:    0,
			Type: "npc",
		})
		boss.SpawnedAdds = append(boss.SpawnedAdds, addID)
	}
}

func (dm *DungeonManager) handleBossDefeat(inst *DungeonInstance, boss *BossState) {
	slog.Info("Dungeon Boss defeated!", "boss", boss.ID, "instance", inst.ID)
	dm.broadcastInstanceMessage(inst, fmt.Sprintf("Parabéns! O temível %s foi derrotado!", boss.Stats.Name))

	// Remove adds que sobraram
	for _, addID := range boss.SpawnedAdds {
		dm.combatManager.DeregisterEntity(addID)
		dm.spatialIndex.RemoveEntity(addID)
	}

	// Atualiza checkpoints da masmorra
	nextCheckpoint := "Entrance"
	if boss.TemplateID == "shadow_lord" {
		nextCheckpoint = "Antechamber"
	} else if boss.TemplateID == "abyss_terror" {
		nextCheckpoint = "Throne Room"
	} else {
		nextCheckpoint = "Completed"
	}
	inst.CheckpointID = nextCheckpoint

	// Persiste progresso no PostgreSQL
	for pID := range inst.Players {
		dm.saveCheckpoint(pID, inst.TemplateID, nextCheckpoint)
		dm.saveRaidLockout(pID, boss.TemplateID)
		dm.saveBossKillState(pID, boss.TemplateID)
	}

	// Calcula as contribuições e distribui recompensas (PATCH 6)
	dm.distributeDungeonLoot(inst, boss)
}

func (dm *DungeonManager) distributeDungeonLoot(inst *DungeonInstance, boss *BossState) {
	// Pega tabela de ódio do boss do combate para determinar a contribuição
	threatTable, exists := dm.combatManager.GetAggroTable(boss.ID)
	if !exists {
		return
	}

	topThreatID, hasTop := threatTable.GetTopTarget()
	if !hasTop {
		return
	}

	// Distribui loot baseado em contribuição (PATCH 6)
	totalThreat := 0.0
	for pID := range inst.Players {
		threat := threatTable.GetThreat(pID)
		totalThreat += threat
	}

	for pID := range inst.Players {
		threat := threatTable.GetThreat(pID)
		contribPct := 0.0
		if totalThreat > 0 {
			contribPct = (threat / totalThreat) * 100
		}

		// Determina o loot
		itemID := "epic_shadow_shard"
		qty := 1
		if pID == topThreatID {
			itemID = "legendary_void_essence"
			qty = 2
		}

		// PATCH 6: Auditoria persistente
		dm.saveContributionAuditLog(pID, boss.TemplateID, threat, 0.0, contribPct, itemID)

		// PATCH 3: Reserva de Loot Idempotente para evitar duplicações
		dm.mu.Lock()
		reservation, existsRes := dm.lootReservations[pID]
		if !existsRes {
			reservation = &LootReservation{
				InstanceID: inst.ID,
				PlayerID:   pID,
				Items:      []LootItem{},
				ExpiresAt:  time.Now().Add(30 * time.Second), // 30 segundos para resgatar
			}
			dm.lootReservations[pID] = reservation
		}
		reservation.Items = append(reservation.Items, LootItem{
			ItemID:   itemID,
			Quantity: qty,
			Claimed:  false,
		})
		dm.mu.Unlock()

		// Notifica o jogador da recompensa e da possibilidade de resgate (SC_LOOT_NOTIFICATION)
		notifPack := &protocol.Packet{
			Opcode:  protocol.SC_LOOT_NOTIFICATION,
			Payload: protocol.EncodeLootNotification(itemID, uint32(qty), 1, 1),
		}
		dm.sendPacketToPlayer(pID, notifPack)
		dm.sendSystemMessage(pID, fmt.Sprintf("Você obteve o direito de reivindicar %dx %s por contribuir com %.1f%% do combate!", qty, itemID, contribPct))
	}
}

// ClaimDungeonLoot resgata o loot garantindo o PATCH 3 (Loot Idempotency)
func (dm *DungeonManager) ClaimDungeonLoot(playerID string, itemID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	res, exists := dm.lootReservations[playerID]
	if !exists {
		return fmt.Errorf("nenhum loot reservado para resgatar")
	}

	if time.Now().After(res.ExpiresAt) {
		return fmt.Errorf("o tempo limite de resgate de 30 segundos expirou")
	}

	playerInv, existsInv := dm.inventories[playerID]
	if !existsInv {
		return fmt.Errorf("inventário do jogador não carregado")
	}

	foundIdx := -1
	for i, item := range res.Items {
		if item.ItemID == itemID && !item.Claimed {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		return fmt.Errorf("recompensa já resgatada ou não encontrada")
	}

	// Adiciona ao inventário
	success := playerInv.AddItem(itemID, res.Items[foundIdx].Quantity)
	if !success {
		return fmt.Errorf("inventário cheio! Limpe espaço antes de resgatar")
	}

	res.Items[foundIdx].Claimed = true
	slog.Info("Player successfully claimed reserved loot", "playerID", playerID, "item", itemID, "qty", res.Items[foundIdx].Quantity)
	
	// Notifica
	dm.sendSystemMessage(playerID, fmt.Sprintf("Você resgatou com sucesso: %dx %s", res.Items[foundIdx].Quantity, itemID))
	
	// Sync inventário
	stats, _ := dm.combatManager.GetEntityStats(playerID)
	syncPayload := protocol.EncodeInventorySync(&protocol.InventorySyncEvent{
	MaxHealth: stats.MaxHealth,
	MaxMana:   stats.MaxMana,
    })
	dm.sendPacketToPlayer(playerID, &protocol.Packet{
		Opcode:  protocol.SC_INVENTORY_SYNC,
		Payload: syncPayload,
	})

	return nil
}

// TickWorldBoss lida com re-spawns agendados de World Bosses, ticks de combate compartilhado e recompensas ranqueadas
func (dm *DungeonManager) TickWorldBoss() {
	dm.worldBoss.mu.Lock()
	defer dm.worldBoss.mu.Unlock()

	now := time.Now()

	// 1. Lógica de Respawn agendada
	if !dm.worldBoss.Active && now.After(dm.worldBoss.NextSpawnTime) {
		dm.worldBoss.Active = true
		dm.worldBoss.ThreatTable = make(map[string]float64)
		dm.worldBoss.HealTable = make(map[string]float64)
		dm.worldBoss.InCombatSince = time.Time{}

		bossHP := 50000.0
		dm.worldBoss.Stats = &combat.EntityStats{
			ID:               dm.worldBoss.ID,
			Name:             dm.worldBoss.Name,
			IsPlayer:         false,
			Faction:          "Monsters",
			Level:            60,
			BaseAttack:       120.0,
			WeaponDamage:     40.0,
			Defense:          100.0,
			Resistance:       80.0,
			Accuracy:         100.0,
			Evasion:          15.0,
			CritChance:       0.20,
			CritMultiplier:   2.0,
			ArmorPenetration: 0.30,
			Element:          "Fire",
			Health:           bossHP,
			MaxHealth:        bossHP,
			Mana:             2000.0,
			MaxMana:          2000.0,
			LastCombatTime:   now,
		}

		dm.combatManager.RegisterEntity(dm.worldBoss.Stats, dm.worldBoss.X, dm.worldBoss.Y)
		dm.spatialIndex.RegisterEntity(&movement.Entity{
			ID:   dm.worldBoss.ID,
			Name: dm.worldBoss.Name,
			X:    dm.worldBoss.X,
			Y:    dm.worldBoss.Y,
			Z:    0,
			Type: "npc",
		})

		slog.Info("WORLD BOSS SPAWNED!", "id", dm.worldBoss.ID)
		dm.broadcastGlobalSystemMessage("CUIDADO! O World Boss Behemoth surgiu nas coordenadas (500, 500)!")
		return
	}

	if !dm.worldBoss.Active {
		return
	}

	// 2. Verifica se foi derrotado
	if dm.worldBoss.Stats.Health <= 0 {
		dm.worldBoss.Active = false
		dm.worldBoss.NextSpawnTime = now.Add(15 * time.Minute) // Agendamento de re-spawn a cada 15 minutos

		dm.combatManager.DeregisterEntity(dm.worldBoss.ID)
		dm.spatialIndex.RemoveEntity(dm.worldBoss.ID)

		slog.Info("World Boss defeated!")
		dm.broadcastGlobalSystemMessage("URGENTE! O World Boss Behemoth foi derrotado pelos heróis do servidor!")

		// Distribui recompensas ranqueadas baseadas em contribuição direta
		go dm.distributeWorldBossRankedRewards()
		return
	}

	if dm.worldBoss.InCombatSince.IsZero() {
		// Detecta se entrou em combate buscando contribuição ativa
		threatTable, exists := dm.combatManager.GetAggroTable(dm.worldBoss.ID)
		if exists {
			_, hasTop := threatTable.GetTopTarget()
			if hasTop {
				dm.worldBoss.InCombatSince = now
			}
		}
	}

	// 3. IA básica do World Boss
	if !dm.worldBoss.InCombatSince.IsZero() {
		threatTable, exists := dm.combatManager.GetAggroTable(dm.worldBoss.ID)
		if exists {
			topTarget, hasTop := threatTable.GetTopTarget()
			if hasTop {
				// Ataca alvo prioritário
				dm.combatManager.ProcessAttackRequest(dm.worldBoss.ID, topTarget, "sword")

				// Periodicamente desfere ataque telegrafado em área contra o top agressor
				if rand.Float64() < 0.15 {
					tx, ty, _ := dm.combatManager.GetEntityPosition(topTarget)
					// Causa dano imediato de terremoto
					for pID, pInv := range dm.inventories {
						pStats, ok := dm.combatManager.GetEntityStats(pID)
						if ok && pStats.Health > 0 {
							px, py, _ := dm.combatManager.GetEntityPosition(pID)
							if math.Hypot(px-tx, py-ty) <= 12.0 {
								dmg := 350.0
								pStats.Health -= dmg
								if pStats.Health < 0 {
									pStats.Health = 0
								}
								dm.sendCombatDamageSync(pID, dm.worldBoss.ID, pID, dmg, true, "Behemoth Earthshaker")
								pInv.SetDirty(true)
							}
						}
					}
				}
			}
		}
	}
}

func (dm *DungeonManager) distributeWorldBossRankedRewards() {
	// Acessa o aggro table para medir o dano total de cada herói do servidor
	threatTable, exists := dm.combatManager.GetAggroTable(dm.worldBoss.ID)
	if !exists {
		return
	}

	type Contributor struct {
		PlayerID string
		Score    float64
	}

	var contribs []Contributor
	totalScore := 0.0

	dm.mu.RLock()
	for pID := range dm.inventories {
		threat := threatTable.GetThreat(pID)
		if threat > 0 {
			contribs = append(contribs, Contributor{PlayerID: pID, Score: threat})
			totalScore += threat
		}
	}
	dm.mu.RUnlock()

	if len(contribs) == 0 {
		return
	}

	// Ordena por maior pontuação de contribuição (Rank)
	for i := 0; i < len(contribs)-1; i++ {
		for j := i + 1; j < len(contribs); j++ {
			if contribs[j].Score > contribs[i].Score {
				contribs[i], contribs[j] = contribs[j], contribs[i]
			}
		}
	}

	// Entrega loot baseado na classificação de ranqueamento
	for rank, c := range contribs {
		contribPct := (c.Score / totalScore) * 100
		
		rewardItemID := "rare_behemoth_scale"
		if rank == 0 {
			rewardItemID = "legendary_behemoth_eye" // Primeiro colocado ganha loot épico
		} else if rank < 3 {
			rewardItemID = "epic_behemoth_shard" // Top 3
		}

		qty := 1
		dm.saveContributionAuditLog(c.PlayerID, "world_boss", c.Score, 0.0, contribPct, rewardItemID)

		// Insere diretamente na reserva do jogador
		dm.mu.Lock()
		res, ok := dm.lootReservations[c.PlayerID]
		if !ok {
			res = &LootReservation{
				InstanceID: "world_boss_loot",
				PlayerID:   c.PlayerID,
				Items:      []LootItem{},
				ExpiresAt:  time.Now().Add(60 * time.Second),
			}
			dm.lootReservations[c.PlayerID] = res
		}
		res.Items = append(res.Items, LootItem{
			ItemID:   rewardItemID,
			Quantity: qty,
			Claimed:  false,
		})
		dm.mu.Unlock()

		dm.sendSystemMessage(c.PlayerID, fmt.Sprintf("World Boss derrotado! Você classificou-se no Rank %d com %.1f%% de dano! Resgate o seu item %s!", rank+1, contribPct, rewardItemID))
	}
}

// ForceWorldBossSpawn permite forçar manualmente o spawn para fins de testes rápidos (CS_WORLD_BOSS_SPAWN_REQ)
func (dm *DungeonManager) ForceWorldBossSpawn() {
	dm.worldBoss.mu.Lock()
	dm.worldBoss.NextSpawnTime = time.Now()
	dm.worldBoss.mu.Unlock()
	dm.TickWorldBoss()
}

// LeaveDungeon remove o jogador da masmorra limpando seu mapa ativo de instâncias
func (dm *DungeonManager) LeaveDungeon(playerID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	instID, ok := dm.playerInstanceMap[playerID]
	if !ok {
		return
	}

	delete(dm.playerInstanceMap, playerID)
	slog.Info("Player left dungeon instance", "player", playerID, "instance", instID)

	inst, ok := dm.instances[instID]
	if !ok {
		return
	}

	delete(inst.Players, playerID)

	// Teleporta de volta para fora
	dm.spatialIndex.UpdateEntityPosition(playerID, 100.0, 100.0, 0)
	dm.combatManager.UpdateEntityPosition(playerID, 100.0, 100.0)

	dm.sendSystemMessage(playerID, "Você saiu da masmorra instanciada.")

	// Se não houver mais jogadores, destrói a masmorra instanciada
	if len(inst.Players) == 0 {
		inst.IsDestroyed = true
		delete(dm.instances, instID)
		slog.Info("Dungeon instance destroyed due to empty team roster", "instance", instID)
		
		// Remove bosses ativos da masmorra instanciada
		for _, boss := range inst.Bosses {
			dm.combatManager.DeregisterEntity(boss.ID)
			dm.spatialIndex.RemoveEntity(boss.ID)
		}
	}
}

func (dm *DungeonManager) destroyInstance(instID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	inst, ok := dm.instances[instID]
	if !ok {
		return
	}

	inst.IsDestroyed = true
	delete(dm.instances, instID)

	for pID := range inst.Players {
		delete(dm.playerInstanceMap, pID)
		// Teleporta de volta para a cidade
		dm.spatialIndex.UpdateEntityPosition(pID, 100.0, 100.0, 0)
		dm.combatManager.UpdateEntityPosition(pID, 100.0, 100.0)
		dm.sendSystemMessage(pID, "O tempo limite esgotou e você foi retirado da masmorra.")
	}

	// Remove NPCs
	for _, boss := range inst.Bosses {
		dm.combatManager.DeregisterEntity(boss.ID)
		dm.spatialIndex.RemoveEntity(boss.ID)
	}
}

// --- Métodos de Persistência PostgreSQL e Fallback seguro ---

func (dm *DungeonManager) saveRaidLockout(playerID string, bossID string) {
	if dm.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// LOCKOUT por 1 dia
	lockedUntil := time.Now().Add(24 * time.Hour)

	_, err := dm.db.ExecContext(ctx, `
		INSERT INTO raid_lockouts (character_id, boss_id, locked_until)
		VALUES ((SELECT id FROM characters WHERE name = $1), $2, $3)
		ON CONFLICT (character_id, boss_id) DO UPDATE SET locked_until = $3
	`, playerID, bossID, lockedUntil)
	if err != nil {
		slog.Error("Failed to save raid lockout to database", "player", playerID, "boss", bossID, "error", err)
	}
}

func (dm *DungeonManager) checkRaidLockout(playerID string, bossID string) (bool, error) {
	if dm.db == nil {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var lockedUntil time.Time
	err := dm.db.QueryRowContext(ctx, `
		SELECT locked_until FROM raid_lockouts 
		WHERE character_id = (SELECT id FROM characters WHERE name = $1) AND boss_id = $2
	`, playerID, bossID).Scan(&lockedUntil)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return time.Now().Before(lockedUntil), nil
}

func (dm *DungeonManager) saveCheckpoint(playerID string, dungeonID string, checkpointID string) {
	if dm.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := dm.db.ExecContext(ctx, `
		INSERT INTO dungeon_checkpoints (character_id, dungeon_id, checkpoint_id)
		VALUES ((SELECT id FROM characters WHERE name = $1), $2, $3)
		ON CONFLICT (character_id, dungeon_id) DO UPDATE SET checkpoint_id = $3
	`, playerID, dungeonID, checkpointID)
	if err != nil {
		slog.Error("Failed to save dungeon checkpoint", "player", playerID, "dungeon", dungeonID, "error", err)
	}
}

func (dm *DungeonManager) loadCheckpoint(playerID string, dungeonID string) (string, error) {
	if dm.db == nil {
		return "", nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cp string
	err := dm.db.QueryRowContext(ctx, `
		SELECT checkpoint_id FROM dungeon_checkpoints
		WHERE character_id = (SELECT id FROM characters WHERE name = $1) AND dungeon_id = $2
	`, playerID, dungeonID).Scan(&cp)

	if err == sql.ErrNoRows {
		return "", nil
	}
	return cp, err
}

func (dm *DungeonManager) saveBossKillState(playerID string, bossID string) {
	if dm.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := dm.db.ExecContext(ctx, `
		INSERT INTO boss_kill_states (character_id, boss_id)
		VALUES ((SELECT id FROM characters WHERE name = $1), $2)
		ON CONFLICT DO NOTHING
	`, playerID, bossID)
	if err != nil {
		slog.Error("Failed to save boss kill state", "player", playerID, "boss", bossID, "error", err)
	}
}

func (dm *DungeonManager) saveContributionAuditLog(playerID string, bossID string, dmg float64, heal float64, score float64, rewardID string) {
	if dm.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := dm.db.ExecContext(ctx, `
		INSERT INTO contribution_audit_logs (character_id, boss_id, damage_dealt, healing_done, contribution_score, reward_item_id)
		VALUES ((SELECT id FROM characters WHERE name = $1), $2, $3, $4, $5, $6)
	`, playerID, bossID, dmg, heal, score, rewardID)
	if err != nil {
		slog.Error("Failed to save contribution audit log to database", "player", playerID, "boss", bossID, "error", err)
	}
}

// --- Funções Auxiliares de Comunicação e Transmissão ---

func (dm *DungeonManager) sendPacketToPlayer(playerID string, packet *protocol.Packet) {
	if conn, exists := dm.aoiManager.GetPlayerConn(playerID); exists {
		conn.Write(packet.Serialize())
	}
}

func (dm *DungeonManager) sendSystemMessage(playerID string, text string) {
	packet := &protocol.Packet{
		Opcode:  protocol.SC_CHAT_MESSAGE,
		Payload: protocol.EncodeChatMessage(0, "System", text),
	}
	dm.sendPacketToPlayer(playerID, packet)
}

func (dm *DungeonManager) broadcastGlobalSystemMessage(text string) {
	packet := &protocol.Packet{
		Opcode:  protocol.SC_CHAT_MESSAGE,
		Payload: protocol.EncodeChatMessage(0, "System", text),
	}
	dm.aoiManager.BroadcastMovement("system", protocol.SC_CHAT_MESSAGE, packet.Serialize())
}

func (dm *DungeonManager) broadcastInstancePacket(inst *DungeonInstance, packet *protocol.Packet) {
	serialized := packet.Serialize()
	for pID := range inst.Players {
		if conn, exists := dm.aoiManager.GetPlayerConn(pID); exists {
			conn.Write(serialized)
		}
	}
}

func (dm *DungeonManager) broadcastInstanceMessage(inst *DungeonInstance, text string) {
	packet := &protocol.Packet{
		Opcode:  protocol.SC_CHAT_MESSAGE,
		Payload: protocol.EncodeChatMessage(0, "System", text),
	}
	dm.broadcastInstancePacket(inst, packet)
}

func (dm *DungeonManager) sendCombatDamageSync(playerID string, attackerID string, targetID string, damage float64, crit bool, source string) {
	payload := protocol.EncodeDamageEvent(attackerID, targetID, damage, crit, true, true, source)
	packet := &protocol.Packet{
		Opcode:  protocol.SC_DAMAGE_EVENT,
		Payload: payload,
	}
	dm.sendPacketToPlayer(playerID, packet)
}
