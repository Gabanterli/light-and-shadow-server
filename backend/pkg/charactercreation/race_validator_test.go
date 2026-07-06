package charactercreation

import (
	"testing"

	"github.com/light-and-shadow/backend/pkg/gamedata/rules"
)

func TestRuleRegistryRaceValidator_IsPlayableRace(t *testing.T) {
	registry, err := rules.NewDefaultRegistry()
	if err != nil {
		t.Fatalf("Failed to create default rule registry for testing: %v", err)
	}

	validator := NewRuleRegistryRaceValidator(registry)

	testCases := []struct {
		name     string
		raceID   string
		expected bool
	}{
		// Valid races
		{"Official race human", "human", true},
		{"Official race forest_elf", "forest_elf", true},
		{"Official race dwarf", "dwarf", true},
		{"Official race ice_elf", "ice_elf", true},
		{"Official race green_orc", "green_orc", true},
		{"Official race with surrounding spaces", "  human  ", true},

		// Invalid inputs
		{"Empty raceID", "", false},
		{"Whitespace raceID", "   ", false},
		{"Non-playable race ogre", "ogre", false},
		{"Unknown raceID", "unknown_race", false},

		// Wrong category
		{"Rule with wrong category (class)", "knight", false},
		{"Rule with wrong category (element)", "fire", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.IsPlayableRace(tc.raceID)
			if result != tc.expected {
				t.Errorf("IsPlayableRace(%q) = %v; want %v", tc.raceID, result, tc.expected)
			}
		})
	}

	t.Run("Nil receiver returns false", func(t *testing.T) {
		var nilValidator *RuleRegistryRaceValidator
		if nilValidator.IsPlayableRace("human") {
			t.Error("expected a nil receiver to return false, but got true")
		}
	})

	t.Run("Nil registry returns false", func(t *testing.T) {
		validatorWithNilRegistry := NewRuleRegistryRaceValidator(nil)
		if validatorWithNilRegistry.IsPlayableRace("human") {
			t.Error("expected a validator with a nil registry to return false, but got true")
		}
	})
}
