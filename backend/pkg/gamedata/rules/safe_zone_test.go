package rules

import (
	"errors"
	"testing"
)

func TestOfficialZoneTypes(t *testing.T) {
	t.Run("returns exactly 3 zones", func(t *testing.T) {
		types := OfficialZoneTypes()
		if len(types) != 3 {
			t.Errorf("expected 3 official zone types, but got %d", len(types))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		types1 := OfficialZoneTypes()
		if len(types1) == 0 {
			t.Fatal("expected non-empty slice")
		}
		originalFirstType := types1[0]
		types1[0] = "mutated"

		types2 := OfficialZoneTypes()
		if types2[0] == "mutated" {
			t.Error("mutation of returned slice affected the internal source")
		}
		if types2[0] != originalFirstType {
			t.Errorf("expected first type to be '%s', but got '%s'", originalFirstType, types2[0])
		}
	})
}

func TestIsOfficialZoneType(t *testing.T) {
	testCases := []struct {
		name     string
		zoneType ZoneType
		expected bool
	}{
		{"safe is official", ZoneTypeSafe, true},
		{"combat is official", ZoneTypeCombat, true},
		{"neutral is official", ZoneTypeNeutral, true},
		{"empty string is not official", "", false},
		{"city is not official", "city", false},
		{"pvp is not official", "pvp", false},
		{"water is not official", "water", false},
		{"air is not official", "air", false},
		{"unknown is not official", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialZoneType(tc.zoneType); got != tc.expected {
				t.Errorf("IsOfficialZoneType('%s') = %v; want %v", tc.zoneType, got, tc.expected)
			}
		})
	}
}

func TestIsSafeZone(t *testing.T) {
	testCases := []struct {
		name     string
		zoneType ZoneType
		expected bool
	}{
		{"safe is a safe zone", ZoneTypeSafe, true},
		{"combat is not a safe zone", ZoneTypeCombat, false},
		{"neutral is not a safe zone", ZoneTypeNeutral, false},
		{"empty string is not a safe zone", "", false},
		{"city is not a safe zone", "city", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSafeZone(tc.zoneType); got != tc.expected {
				t.Errorf("IsSafeZone('%s') = %v; want %v", tc.zoneType, got, tc.expected)
			}
		})
	}
}

func TestCanCombatOccurInZone(t *testing.T) {
	t.Run("rejects invalid zone types", func(t *testing.T) {
		invalidTypes := []ZoneType{"", "city", "unknown"}
		for _, zt := range invalidTypes {
			req := ZoneCombatRuleRequest{ZoneType: zt}
			err := CanCombatOccurInZone(req)
			if !errors.Is(err, ErrInvalidZoneType) {
				t.Errorf("for type '%s', expected ErrInvalidZoneType, got %v", zt, err)
			}
		}
	})

	t.Run("rejects combat in safe zone", func(t *testing.T) {
		req := ZoneCombatRuleRequest{ZoneType: ZoneTypeSafe}
		err := CanCombatOccurInZone(req)
		if !errors.Is(err, ErrCombatBlockedInSafeZone) {
			t.Errorf("expected ErrCombatBlockedInSafeZone, got %v", err)
		}
	})

	t.Run("allows combat in combat zone", func(t *testing.T) {
		req := ZoneCombatRuleRequest{ZoneType: ZoneTypeCombat}
		err := CanCombatOccurInZone(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("allows combat in neutral zone", func(t *testing.T) {
		req := ZoneCombatRuleRequest{ZoneType: ZoneTypeNeutral}
		err := CanCombatOccurInZone(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		// Setup: An invalid zone type.
		// Expected: Should return ErrInvalidZoneType first.
		req := ZoneCombatRuleRequest{ZoneType: "city"}
		err := CanCombatOccurInZone(req)
		if !errors.Is(err, ErrInvalidZoneType) {
			t.Errorf("expected ErrInvalidZoneType due to deterministic order, but got %v", err)
		}
	})
}

func TestMustBlockCombatInSafeZone(t *testing.T) {
	if !MustBlockCombatInSafeZone() {
		t.Error("expected MustBlockCombatInSafeZone to return true")
	}
}
