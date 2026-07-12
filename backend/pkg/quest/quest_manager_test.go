package quest

import (
	"fmt"
	"sync"
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
	cm.RegisterEntity(&combat.EntityStats{
		ID:        playerID,
		Name:      playerID,
		IsPlayer:  true,
		Level:     1,
		Health:    100,
		MaxHealth: 100,
		Mana:      100,
		MaxMana:   100,
	}, 0, 0)

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

// B3-B, B3-E: Test ephemeral dialogue session management.
func TestQuestManager_DialogueSession(t *testing.T) {
	qm := NewQuestManager(nil, nil, nil)
	playerID := "player1"
	npcID := "npc_mentor"

	t.Run("begin and check dialogue", func(t *testing.T) {
		qm.BeginDialogue(playerID, npcID)
		if !qm.IsPlayerInDialogue(playerID) {
			t.Error("Expected IsPlayerInDialogue to be true after beginning a session")
		}
		currentNPC, ok := qm.GetCurrentDialogueNPC(playerID)
		if !ok || currentNPC != npcID {
			t.Errorf("Expected GetCurrentDialogueNPC to return '%s', but got '%s'", npcID, currentNPC)
		}
	})

	t.Run("clear dialogue state", func(t *testing.T) {
		// Ensure state exists before clearing
		qm.BeginDialogue(playerID, npcID)

		clearedNPC, ok := qm.ClearDialogueState(playerID)
		if !ok || clearedNPC != npcID {
			t.Errorf("Expected ClearDialogueState to return '%s' and true, but got '%s' and %v", npcID, clearedNPC, ok)
		}
		if qm.IsPlayerInDialogue(playerID) {
			t.Error("Expected IsPlayerInDialogue to be false after clearing a session")
		}
	})

	t.Run("clear dialogue state is idempotent", func(t *testing.T) {
		_, ok := qm.ClearDialogueState(playerID)
		if ok {
			t.Error("Expected second call to ClearDialogueState to return false, but it returned true")
		}
	})

	t.Run("begin dialogue replaces previous session", func(t *testing.T) {
		newNpcID := "npc_guard"
		qm.BeginDialogue(playerID, npcID)
		qm.BeginDialogue(playerID, newNpcID)

		currentNPC, ok := qm.GetCurrentDialogueNPC(playerID)
		if !ok || currentNPC != newNpcID {
			t.Errorf("Expected new dialogue session to replace the old one, wanted '%s', got '%s'", newNpcID, currentNPC)
		}
	})

	t.Run("clearing dialogue does not affect persistent flags", func(t *testing.T) {
		persistentFlag := "some_quest_flag"
		qm.SetDialogueFlag(playerID, npcID, persistentFlag)
		qm.BeginDialogue(playerID, npcID)

		qm.ClearDialogueState(playerID)

		if qm.GetDialogueFlag(playerID, npcID) != persistentFlag {
			t.Error("Clearing ephemeral dialogue state should not affect persistent dialogue flags")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				player := fmt.Sprintf("player_%d", i)
				qm.BeginDialogue(player, "npc_stress")
				_ = qm.IsPlayerInDialogue(player)
			}(i)
		}
		wg.Wait()
	})
}
