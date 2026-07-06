package rules

import (
	"errors"
	"testing"
)

func TestOfficialBaseClassIDs(t *testing.T) {
	t.Run("returns exactly 5 classes", func(t *testing.T) {
		ids := OfficialBaseClassIDs()
		if len(ids) != 5 {
			t.Errorf("expected 5 official base classes, but got %d", len(ids))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		ids1 := OfficialBaseClassIDs()
		if len(ids1) == 0 {
			t.Fatal("expected non-empty slice")
		}
		originalFirstID := ids1[0]
		ids1[0] = "mutated"

		ids2 := OfficialBaseClassIDs()
		if ids2[0] == "mutated" {
			t.Error("mutation of returned slice affected the internal source")
		}
		if ids2[0] != originalFirstID {
			t.Errorf("expected first ID to be '%s', but got '%s'", originalFirstID, ids2[0])
		}
	})
}

func TestIsOfficialBaseClass(t *testing.T) {
	testCases := []struct {
		name     string
		id       RuleID
		expected bool
	}{
		{"knight is official", ClassKnight, true},
		{"mage is official", ClassMage, true},
		{"archer is official", ClassArcher, true},
		{"assassin is official", ClassAssassin, true},
		{"cleric is official", ClassCleric, true},
		{"novice is not official base class", StartingClassNovice, false},
		{"ogre is not official base class", "ogre", false},
		{"air is not official base class", "air", false},
		{"water is not official base class", "water", false},
		{"paladin is not official base class", "paladin", false},
		{"empty string is not official", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialBaseClass(tc.id); got != tc.expected {
				t.Errorf("IsOfficialBaseClass('%s') = %v; want %v", tc.id, got, tc.expected)
			}
		})
	}
}

func TestCanSelectBaseClass(t *testing.T) {
	t.Run("allows selection at level 10", func(t *testing.T) {
		err := CanSelectBaseClass(10, StartingClassNovice, ClassKnight)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("allows selection after level 10", func(t *testing.T) {
		err := CanSelectBaseClass(11, StartingClassNovice, ClassMage)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("rejects invalid target class", func(t *testing.T) {
		invalidTargets := []RuleID{"ogre", "air", "water", StartingClassNovice, ""}
		for _, target := range invalidTargets {
			err := CanSelectBaseClass(10, StartingClassNovice, target)
			if !errors.Is(err, ErrInvalidBaseClass) {
				t.Errorf("for target '%s', expected ErrInvalidBaseClass, got %v", target, err)
			}
		}
	})

	t.Run("rejects selection below minimum level", func(t *testing.T) {
		err := CanSelectBaseClass(9, StartingClassNovice, ClassKnight)
		if !errors.Is(err, ErrClassSelectionLevelTooLow) {
			t.Errorf("expected ErrClassSelectionLevelTooLow, got %v", err)
		}
	})

	t.Run("rejects selection if class already chosen", func(t *testing.T) {
		err := CanSelectBaseClass(10, ClassMage, ClassKnight)
		if !errors.Is(err, ErrClassSelectionAlreadyChosen) {
			t.Errorf("expected ErrClassSelectionAlreadyChosen, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		// Setup: Level is too low, class is already chosen, AND target is invalid.
		// Expected: Should return ErrInvalidBaseClass first.
		err := CanSelectBaseClass(1, ClassMage, "ogre")
		if !errors.Is(err, ErrInvalidBaseClass) {
			t.Errorf("expected ErrInvalidBaseClass due to deterministic order, but got %v", err)
		}

		// Setup: Level is too low, class is already chosen. Target is valid.
		// Expected: Should return ErrClassSelectionLevelTooLow.
		err = CanSelectBaseClass(1, ClassMage, ClassKnight)
		if !errors.Is(err, ErrClassSelectionLevelTooLow) {
			t.Errorf("expected ErrClassSelectionLevelTooLow due to deterministic order, but got %v", err)
		}

		// Setup: Level is ok, class is already chosen. Target is valid.
		// Expected: Should return ErrClassSelectionAlreadyChosen.
		err = CanSelectBaseClass(10, ClassMage, ClassKnight)
		if !errors.Is(err, ErrClassSelectionAlreadyChosen) {
			t.Errorf("expected ErrClassSelectionAlreadyChosen due to deterministic order, but got %v", err)
		}
	})
}

func TestMustSelectClassAtOrAfterLevel(t *testing.T) {
	level := MustSelectClassAtOrAfterLevel()
	if level != 10 {
		t.Errorf("expected level 10, got %d", level)
	}
	if level != ClassSelectionMinimumLevel {
		t.Errorf("expected function to return ClassSelectionMinimumLevel constant")
	}
}
