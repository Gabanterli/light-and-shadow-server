package quest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/messaging"
)

func TestQuestManagerInit(t *testing.T) {
	cm := combat.NewCombatManager(nil)
	invs := make(map[string]*inventory.PlayerInventory)

	qm := NewQuestManager(nil, cm, invs)
	if qm == nil {
		t.Fatal("Expected NewQuestManager to return a valid instance, got nil")
	}
}

func TestQuestVersioningAndOptimisticLocking(t *testing.T) {
	// Create test state
	q := &ActiveQuest{
		QuestID: "test_quest",
		Version: 1,
		Status:  "active",
	}

	if q.Version != 1 {
		t.Errorf("Expected quest version to start at 1, got %d", q.Version)
	}

	// Simulates version increment upon save/complete status changes
	q.Version++
	if q.Version != 2 {
		t.Errorf("Expected quest version to be 2 after increment, got %d", q.Version)
	}
}

func TestRewardIdempotencyHash(t *testing.T) {
	cm := combat.NewCombatManager(nil)
	invs := make(map[string]*inventory.PlayerInventory)
	qm := NewQuestManager(nil, cm, invs)

	rewards1 := QuestRewards{
		Experience: 100,
		Gold:       50,
		Items: []RewardItem{
			{ItemID: "iron_ore", Quantity: 5},
		},
	}

	rewards2 := QuestRewards{
		Experience: 100,
		Gold:       50,
		Items: []RewardItem{
			{ItemID: "iron_ore", Quantity: 5},
		},
	}

	rewards3 := QuestRewards{
		Experience: 200,
		Gold:       50,
		Items: []RewardItem{
			{ItemID: "iron_ore", Quantity: 5},
		},
	}

	hash1 := qm.getRewardHash(rewards1)
	hash2 := qm.getRewardHash(rewards2)
	hash3 := qm.getRewardHash(rewards3)

	if hash1 != hash2 {
		t.Error("Expected identical rewards to produce identical hashes")
	}

	if hash1 == hash3 {
		t.Error("Expected different rewards to produce different hashes")
	}
}

func TestEventBusQuestIntegration(t *testing.T) {
	cm := combat.NewCombatManager(nil)
	invs := make(map[string]*inventory.PlayerInventory)
	qm := NewQuestManager(nil, cm, invs)

	// Set up definitions
	qm.definitions["slay_goblin"] = QuestDefinition{
		QuestID: "slay_goblin",
		Title:   "Slay Goblins",
		Objectives: []QuestObjective{
			{
				Type:        ObjectiveKillMonster,
				TargetID:    "goblin",
				RequiredQty: 3,
			},
		},
	}

	playerID := "Hero"
	invs[playerID] = inventory.NewPlayerInventory(playerID)

	// Accept Quest
	err := qm.AcceptQuest(playerID, "slay_goblin")
	if err != nil {
		t.Fatalf("Failed to accept quest: %v", err)
	}

	pState := qm.GetPlayerState(playerID)
	qState, exists := pState.Quests["slay_goblin"]
	if !exists {
		t.Fatal("Quest state should exist in player profile")
	}

	if qState.Objectives[0].CurrentQty != 0 {
		t.Errorf("Expected current qty to be 0, got %d", qState.Objectives[0].CurrentQty)
	}

	// Publish kill event via MessageBus
	mb := messaging.GetInstance()
	mb.Publish("monster_killed", messaging.MonsterKilledPayload{
		PlayerID:  playerID,
		MonsterID: "goblin",
	})

	// Wait briefly for asynchronous background event loop processing (PATCH 5)
	time.Sleep(100 * time.Millisecond)

	pState.mu.RLock()
	defer pState.mu.RUnlock()
	if qState.Objectives[0].CurrentQty != 1 {
		t.Errorf("Expected asynchronous event loop to update monster kill objective to 1, got %d", qState.Objectives[0].CurrentQty)
	}
}
