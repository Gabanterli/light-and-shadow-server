package rules

import (
	"errors"
	"testing"
)

func TestIsLevelEligibleForOpenPvP(t *testing.T) {
	testCases := []struct {
		name     string
		level    uint32
		expected bool
	}{
		{"level 0 is not eligible", 0, false},
		{"level 1 is not eligible", 1, false},
		{"level 9 is not eligible", 9, false},
		{"level 10 is eligible", 10, true},
		{"level 11 is eligible", 11, true},
		{"level 100 is eligible", 100, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsLevelEligibleForOpenPvP(tc.level); got != tc.expected {
				t.Errorf("IsLevelEligibleForOpenPvP(%d) = %v; want %v", tc.level, got, tc.expected)
			}
		})
	}
}

func TestCanEngageOpenPvP(t *testing.T) {
	t.Run("allows engagement when both are at minimum level", func(t *testing.T) {
		req := PvPLevelGateRequest{AttackerLevel: 10, TargetLevel: 10}
		err := CanEngageOpenPvP(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("allows engagement when both are above minimum level", func(t *testing.T) {
		req := PvPLevelGateRequest{AttackerLevel: 100, TargetLevel: 50}
		err := CanEngageOpenPvP(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("rejects if attacker level is too low", func(t *testing.T) {
		req := PvPLevelGateRequest{AttackerLevel: 9, TargetLevel: 10}
		err := CanEngageOpenPvP(req)
		if !errors.Is(err, ErrPvPAttackerLevelTooLow) {
			t.Errorf("expected ErrPvPAttackerLevelTooLow, got %v", err)
		}
	})

	t.Run("rejects if target level is too low", func(t *testing.T) {
		req := PvPLevelGateRequest{AttackerLevel: 10, TargetLevel: 9}
		err := CanEngageOpenPvP(req)
		if !errors.Is(err, ErrPvPTargetLevelTooLow) {
			t.Errorf("expected ErrPvPTargetLevelTooLow, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		// Setup: Both attacker and target are below the minimum level.
		// Expected: Should return ErrPvPAttackerLevelTooLow first.
		req := PvPLevelGateRequest{AttackerLevel: 1, TargetLevel: 1}
		err := CanEngageOpenPvP(req)
		if !errors.Is(err, ErrPvPAttackerLevelTooLow) {
			t.Errorf("expected ErrPvPAttackerLevelTooLow due to deterministic order, but got %v", err)
		}
	})
}

func TestMustReachLevelForOpenPvP(t *testing.T) {
	t.Run("returns correct level", func(t *testing.T) {
		level := MustReachLevelForOpenPvP()
		if level != 10 {
			t.Errorf("expected level 10, got %d", level)
		}
	})

	t.Run("returns the constant value", func(t *testing.T) {
		level := MustReachLevelForOpenPvP()
		if level != OpenPvPMinimumLevel {
			t.Errorf("expected function to return OpenPvPMinimumLevel constant")
		}
	})
}
