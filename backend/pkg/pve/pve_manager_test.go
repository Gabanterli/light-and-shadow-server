package pve

import (
	"testing"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/movement"
)

func TestPveManagerInitialization(t *testing.T) {
	si := movement.NewSpatialIndex()
	chunkMgr := movement.NewChunkManager()
	aoi := movement.NewAOIManager(si)
	cm := combat.NewCombatManager(chunkMgr)
	invs := make(map[string]*inventory.PlayerInventory)

	pm := NewPveManager(si, aoi, cm, invs)
	if pm == nil {
		t.Fatal("Expected NewPveManager to return a valid instance, got nil")
	}

	// Verifica se os mapas internos foram instanciados corretamente
	if pm.monsters == nil || pm.lootTables == nil || pm.activeMonsters == nil || pm.spawnCounts == nil {
		t.Error("Internal maps were not initialized correctly")
	}
}

func TestXpSplitAndLevelUpProgression(t *testing.T) {
	si := movement.NewSpatialIndex()
	chunkMgr := movement.NewChunkManager()
	aoi := movement.NewAOIManager(si)
	cm := combat.NewCombatManager(chunkMgr)
	invs := make(map[string]*inventory.PlayerInventory)

	pm := NewPveManager(si, aoi, cm, invs)

	// Cria jogador
	playerID := "TestPlayer"
	pStats := &combat.EntityStats{
		ID:        playerID,
		Name:      "Hero",
		IsPlayer:  true,
		Level:     1,
		Health:    100,
		MaxHealth: 100,
		Mana:      50,
		MaxMana:   50,
	}
	cm.RegisterEntity(pStats, 100, 100)
	si.RegisterEntity(&movement.Entity{
		ID:   playerID,
		Type: "player",
		X:    100,
		Y:    100,
	})

	playerInv := inventory.NewPlayerInventory(playerID)
	invs[playerID] = playerInv

	// Inicializa XP
	SetPlayerXp(playerID, 0)

	// Concede XP suficiente para subir do Level 1 para o Level 2 (Requer 1*1*100 = 100 XP)
	pm.awardXp(playerID, 120)

	// Verifica se subiu de nível e recalculou os stats base de nível de forma autoritativa
	if pStats.Level != 2 {
		t.Errorf("Expected player level to be 2, got %d", pStats.Level)
	}

	if pStats.MaxHealth != 120.0 {
		t.Errorf("Expected max health to be 120.0 after level up, got %f", pStats.MaxHealth)
	}

	if GetPlayerXp(playerID) != 20 {
		t.Errorf("Expected remaining XP to be 20, got %d", GetPlayerXp(playerID))
	}
}

func TestMonsterStateMachineChaseToLeash(t *testing.T) {
	si := movement.NewSpatialIndex()
	chunkMgr := movement.NewChunkManager()
	aoi := movement.NewAOIManager(si)
	cm := combat.NewCombatManager(chunkMgr)
	invs := make(map[string]*inventory.PlayerInventory)

	pm := NewPveManager(si, aoi, cm, invs)

	// Template de monstro para o teste
	mTemplate := MonsterTemplate{
		ID:            "goblin_test",
		Name:          "Goblin de Teste",
		Level:         1,
		MaxHealth:     100,
		AggroRadius:   5.0,
		LeashDistance: 10.0,
		ChaseSpeed:    2.0,
	}

	mStats := &combat.EntityStats{
		ID:        "mob_1",
		Name:      mTemplate.Name,
		IsPlayer:  false,
		Level:     1,
		Health:    100,
		MaxHealth: 100,
	}

	cm.RegisterEntity(mStats, 10, 10)
	si.RegisterEntity(&movement.Entity{
		ID:   "mob_1",
		Type: "npc",
		X:    10,
		Y:    10,
	})

	instance := &MonsterInstance{
		ID:             "mob_1",
		Template:       mTemplate,
		SpawnPointID:   "spawn_1",
		State:          "Idle",
		HomeX:          10,
		HomeY:          10,
		X:              10,
		Y:              10,
		Stats:          mStats,
		ThreatTable:    combat.NewAggroTable(),
		LastActionTime: time.Now(),
	}

	pm.activeMonsters["mob_1"] = instance

	// Cria jogador a 3.0 tiles de distância (dentro do AggroRadius de 5.0)
	pStats := &combat.EntityStats{
		ID:       "player_1",
		Name:     "Player 1",
		IsPlayer: true,
		Level:    1,
		Health:   100,
	}
	cm.RegisterEntity(pStats, 12.0, 10.0)
	si.RegisterEntity(&movement.Entity{
		ID:   "player_1",
		Type: "player",
		X:    12.0,
		Y:    10.0,
	})

	// 1. Processa IA: Deve agrar o jogador por proximidade
	pm.processMonsterAI(instance)

	if instance.State != "Aggro" {
		t.Errorf("Expected state to transition to Aggro, got %s", instance.State)
	}

	// Avança o relógio interno artificialmente
	instance.LastActionTime = time.Now().Add(-1 * time.Second)

	// 2. Processa IA: Deve transicionar para Chase
	pm.processMonsterAI(instance)

	if instance.State != "Chase" {
		t.Errorf("Expected state to transition to Chase, got %s", instance.State)
	}

	// Afasta o jogador para além da LeashDistance ( Home = 10,10. Player = 30,30. Dist = 28.28 > Leash = 10.0)
	cm.UpdateEntityPosition("player_1", 30.0, 30.0)
	si.UpdateEntityPosition("player_1", 30.0, 30.0, 0)

	// 3. Processa IA: Deve detectar o distanciamento da Home e retornar para o estado ReturnHome
	pm.processMonsterAI(instance)

	if instance.State != "ReturnHome" {
		t.Errorf("Expected state to transition to ReturnHome, got %s", instance.State)
	}
}
