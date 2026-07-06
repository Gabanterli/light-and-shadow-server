package rules

import (
	"errors"
	"testing"
)

func TestOfficialSkillProgressionSources(t *testing.T) {
	t.Run("returns exactly 3 sources", func(t *testing.T) {
		sources := OfficialSkillProgressionSources()
		if len(sources) != 3 {
			t.Errorf("expected 3 official sources, but got %d", len(sources))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		sources1 := OfficialSkillProgressionSources()
		if len(sources1) == 0 {
			t.Fatal("expected non-empty slice")
		}
		originalFirstSource := sources1[0]
		sources1[0] = "mutated"

		sources2 := OfficialSkillProgressionSources()
		if sources2[0] == "mutated" {
			t.Error("mutation of returned slice affected the internal source")
		}
		if sources2[0] != originalFirstSource {
			t.Errorf("expected first source to be '%s', but got '%s'", originalFirstSource, sources2[0])
		}
	})
}

func TestIsOfficialSkillProgressionSource(t *testing.T) {
	testCases := []struct {
		name     string
		source   SkillProgressionSource
		expected bool
	}{
		{"combat_use is official", SkillProgressionSourceCombatUse, true},
		{"training_use is official", SkillProgressionSourceTrainingUse, true},
		{"profession_use is official", SkillProgressionSourceProfessionUse, true},
		{"empty string is not official", "", false},
		{"client_grant is not official", "client_grant", false},
		{"admin is not official", "admin", false},
		{"quest_reward is not official", "quest_reward", false},
		{"offline_macro is not official", "offline_macro", false},
		{"water is not official", "water", false},
		{"air is not official", "air", false},
		{"unknown is not official", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialSkillProgressionSource(tc.source); got != tc.expected {
				t.Errorf("IsOfficialSkillProgressionSource('%s') = %v; want %v", tc.source, got, tc.expected)
			}
		})
	}
}

func TestCanGrantSkillProgression(t *testing.T) {
	validRequest := SkillProgressionEligibilityRequest{
		Source:                    SkillProgressionSourceCombatUse,
		ActionValidatedByBackend:  true,
		ContextValidatedByBackend: true,
		MillisecondsSinceLastGain: 1000,
		DiminishingReturnsBlocked: false,
	}

	t.Run("allows progression with exact requirements", func(t *testing.T) {
		err := CanGrantSkillProgression(validRequest)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("allows progression with training source and long interval", func(t *testing.T) {
		req := validRequest
		req.Source = SkillProgressionSourceTrainingUse
		req.MillisecondsSinceLastGain = 5000
		err := CanGrantSkillProgression(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("rejects invalid source", func(t *testing.T) {
		req := validRequest
		invalidSources := []SkillProgressionSource{"", "client_grant", "admin", "offline_macro"}
		for _, source := range invalidSources {
			req.Source = source
			err := CanGrantSkillProgression(req)
			if !errors.Is(err, ErrInvalidSkillProgressionSource) {
				t.Errorf("for source '%s', expected ErrInvalidSkillProgressionSource, got %v", source, err)
			}
		}
	})

	t.Run("rejects if action not validated", func(t *testing.T) {
		req := validRequest
		req.ActionValidatedByBackend = false
		err := CanGrantSkillProgression(req)
		if !errors.Is(err, ErrSkillProgressionActionNotValidated) {
			t.Errorf("expected ErrSkillProgressionActionNotValidated, got %v", err)
		}
	})

	t.Run("rejects if context not validated", func(t *testing.T) {
		req := validRequest
		req.ContextValidatedByBackend = false
		err := CanGrantSkillProgression(req)
		if !errors.Is(err, ErrSkillProgressionContextNotValidated) {
			t.Errorf("expected ErrSkillProgressionContextNotValidated, got %v", err)
		}
	})

	t.Run("rejects if interval too short", func(t *testing.T) {
		req := validRequest
		req.MillisecondsSinceLastGain = 999
		err := CanGrantSkillProgression(req)
		if !errors.Is(err, ErrSkillProgressionIntervalTooShort) {
			t.Errorf("expected ErrSkillProgressionIntervalTooShort, got %v", err)
		}
	})

	t.Run("rejects if diminishing returns blocked", func(t *testing.T) {
		req := validRequest
		req.DiminishingReturnsBlocked = true
		err := CanGrantSkillProgression(req)
		if !errors.Is(err, ErrSkillProgressionDiminishingReturnsBlocked) {
			t.Errorf("expected ErrSkillProgressionDiminishingReturnsBlocked, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		baseReq := SkillProgressionEligibilityRequest{
			MillisecondsSinceLastGain: 0,
			DiminishingReturnsBlocked: true,
			ActionValidatedByBackend:  false,
			ContextValidatedByBackend: false,
		}

		t.Run("invalid source is first error", func(t *testing.T) {
			req := baseReq
			req.Source = "client_grant"
			err := CanGrantSkillProgression(req)
			if !errors.Is(err, ErrInvalidSkillProgressionSource) {
				t.Errorf("expected ErrInvalidSkillProgressionSource, got %v", err)
			}
		})

		t.Run("action not validated is second error", func(t *testing.T) {
			req := baseReq
			req.Source = SkillProgressionSourceCombatUse
			err := CanGrantSkillProgression(req)
			if !errors.Is(err, ErrSkillProgressionActionNotValidated) {
				t.Errorf("expected ErrSkillProgressionActionNotValidated, got %v", err)
			}
		})

		t.Run("context not validated is third error", func(t *testing.T) {
			req := baseReq
			req.Source = SkillProgressionSourceCombatUse
			req.ActionValidatedByBackend = true
			err := CanGrantSkillProgression(req)
			if !errors.Is(err, ErrSkillProgressionContextNotValidated) {
				t.Errorf("expected ErrSkillProgressionContextNotValidated, got %v", err)
			}
		})
	})
}

func TestMustWaitMillisecondsBetweenSkillProgressionGains(t *testing.T) {
	t.Run("returns correct milliseconds", func(t *testing.T) {
		ms := MustWaitMillisecondsBetweenSkillProgressionGains()
		if ms != 1000 {
			t.Errorf("expected 1000, got %d", ms)
		}
	})

	t.Run("returns the constant value", func(t *testing.T) {
		ms := MustWaitMillisecondsBetweenSkillProgressionGains()
		if ms != SkillProgressionMinimumIntervalMilliseconds {
			t.Errorf("expected function to return SkillProgressionMinimumIntervalMilliseconds constant")
		}
	})
}

func TestMustValidateSkillProgressionOnBackend(t *testing.T) {
	if !MustValidateSkillProgressionOnBackend() {
		t.Error("expected MustValidateSkillProgressionOnBackend to return true")
	}
}
