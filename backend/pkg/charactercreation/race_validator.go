package charactercreation

import (
	"strings"

	"github.com/light-and-shadow/backend/pkg/gamedata/rules"
)

// RuleRegistryRaceValidator is an adapter that implements the RaceValidator
// interface by using the central Rule Registry. This decouples the creation
// service from the concrete implementation of game rules.
type RuleRegistryRaceValidator struct {
	registry *rules.Registry
}

// NewRuleRegistryRaceValidator creates a new validator that uses the provided Rule Registry.
func NewRuleRegistryRaceValidator(registry *rules.Registry) *RuleRegistryRaceValidator {
	return &RuleRegistryRaceValidator{
		registry: registry,
	}
}

// IsPlayableRace checks if the given raceID is a valid, playable race
// according to the configured Rule Registry. It returns true only if the rule
// exists and its category is 'race'.
func (v *RuleRegistryRaceValidator) IsPlayableRace(raceID string) bool {
	if v == nil || v.registry == nil {
		return false
	}

	normalizedRaceID := strings.TrimSpace(raceID)
	if normalizedRaceID == "" {
		return false
	}

	def, found := v.registry.Get(rules.RuleID(normalizedRaceID))
	if !found {
		return false
	}

	return def.Category == rules.CategoryRace
}
