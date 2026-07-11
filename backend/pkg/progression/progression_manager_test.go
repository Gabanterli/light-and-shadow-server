package progression

import (
	"testing"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/gamedata/rules"
	"github.com/light-and-shadow/backend/pkg/inventory"
)

func setupProgressionTest() (*ProgressionManager, *combat.CombatManager, map[string]*inventory.PlayerInventory) {
	cm := combat.NewCombatManager(nil, nil)
	invs := make(map[string]*inventory.PlayerInventory)
	pm := NewProgressionManager(nil, cm, invs)
	return pm, cm, invs
}

func TestChooseVocation_Success(t *testing.T) {
	pm, cm, invs := setupProgressionTest()
	playerID := "player1"
	cm.RegisterEntity(&combat.EntityStats{ID: playerID, Class: "novice", Level: 10}, 0, 0)
	invs[playerID] = inventory.NewPlayerInventory(playerID)

	err := pm.ChooseVocation(playerID, "knight")
	if err != nil {
		t.Fatalf("Expected vocation choice to succeed, but got error: %v", err)
	}

	stats, _ := cm.GetEntityStats(playerID)
	if stats.Class != "knight" {
		t.Errorf("Expected player class to be 'knight', but got '%s'", stats.Class)
	}
}

func TestChooseVocation_AlreadyChosen(t *testing.T) {
	pm, cm, invs := setupProgressionTest()
	playerID := "player1"
	cm.RegisterEntity(&combat.EntityStats{ID: playerID, Class: "knight", Level: 10}, 0, 0)
	invs[playerID] = inventory.NewPlayerInventory(playerID)

	err := pm.ChooseVocation(playerID, "mage")
	if err == nil {
		t.Fatal("Expected vocation choice to fail because class was already chosen, but it succeeded")
	}
	if err != rules.ErrClassSelectionAlreadyChosen {
		t.Errorf("Expected error '%v', but got '%v'", rules.ErrClassSelectionAlreadyChosen, err)
	}

	stats, _ := cm.GetEntityStats(playerID)
	if stats.Class != "knight" {
		t.Errorf("Expected player class to remain 'knight', but got '%s'", stats.Class)
	}
}

func TestChooseVocation_GMLevelBypass(t *testing.T) {
	pm, cm, invs := setupProgressionTest()
	playerID := "gm_novice"

	cm.RegisterEntity(&combat.EntityStats{
		ID:    playerID,
		Class: "novice",
		Level: 1,
	}, 0, 0)
	invs[playerID] = inventory.NewPlayerInventory(playerID)

	err := pm.ChooseVocationWithOptions(
		playerID,
		"mage",
		ChooseVocationOptions{
			AllowDevBypass: true,
			DevReason:      "unit test",
		},
	)
	if err != nil {
		t.Fatalf("Expected GM level bypass to allow novice selection: %v", err)
	}
}

func TestChooseVocation_GMBypassCannotReselect(t *testing.T) {
	pm, cm, invs := setupProgressionTest()
	playerID := "gm_already_chosen"

	cm.RegisterEntity(&combat.EntityStats{
		ID:    playerID,
		Class: "knight",
		Level: 1,
	}, 0, 0)
	invs[playerID] = inventory.NewPlayerInventory(playerID)

	err := pm.ChooseVocationWithOptions(
		playerID,
		"mage",
		ChooseVocationOptions{
			AllowDevBypass: true,
			DevReason:      "unit test",
		},
	)
	if err != rules.ErrClassSelectionAlreadyChosen {
		t.Fatalf(
			"Expected ErrClassSelectionAlreadyChosen, got: %v",
			err,
		)
	}
}
