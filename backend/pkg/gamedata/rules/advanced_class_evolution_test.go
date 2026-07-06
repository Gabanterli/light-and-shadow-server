package rules

import (
	"errors"
	"testing"
)

func TestOfficialElementIDs(t *testing.T) {
	t.Run("returns exactly 5 elements", func(t *testing.T) {
		ids := OfficialElementIDs()
		if len(ids) != 5 {
			t.Errorf("expected 5 official elements, but got %d", len(ids))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		ids1 := OfficialElementIDs()
		if len(ids1) == 0 {
			t.Fatal("expected non-empty slice")
		}
		originalFirstID := ids1[0]
		ids1[0] = "mutated"

		ids2 := OfficialElementIDs()
		if ids2[0] == "mutated" {
			t.Error("mutation of returned slice affected the internal source")
		}
		if ids2[0] != originalFirstID {
			t.Errorf("expected first ID to be '%s', but got '%s'", originalFirstID, ids2[0])
		}
	})
}

func TestIsOfficialElement(t *testing.T) {
	testCases := []struct {
		name     string
		id       RuleID
		expected bool
	}{
		{"fire is official", ElementFire, true},
		{"earth is official", ElementEarth, true},
		{"ice is official", ElementIce, true},
		{"shadow is official", ElementShadow, true},
		{"sacred is official", ElementSacred, true},
		{"air is not official", "air", false},
		{"water is not official", "water", false},
		{"ogre is not official", "ogre", false},
		{"novice is not an element", StartingClassNovice, false},
		{"knight is not an element", ClassKnight, false},
		{"empty string is not official", "", false},
		{"lightning is not official", "lightning", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialElement(tc.id); got != tc.expected {
				t.Errorf("IsOfficialElement('%s') = %v; want %v", tc.id, got, tc.expected)
			}
		})
	}
}

func TestCanEvolveAdvancedClass(t *testing.T) {
	validRequest := AdvancedClassEvolutionRequest{
		CharacterLevel: 100,
		AffinityLevel:  100,
		QuestCompleted: true,
		BaseClass:      ClassKnight,
		Element:        ElementFire,
	}

	t.Run("allows evolution with exact requirements", func(t *testing.T) {
		err := CanEvolveAdvancedClass(validRequest)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("allows evolution above requirements", func(t *testing.T) {
		req := AdvancedClassEvolutionRequest{
			CharacterLevel: 150,
			AffinityLevel:  101,
			QuestCompleted: true,
			BaseClass:      ClassAssassin,
			Element:        ElementShadow,
		}
		err := CanEvolveAdvancedClass(req)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("rejects invalid base class", func(t *testing.T) {
		req := validRequest
		invalidClasses := []RuleID{StartingClassNovice, "ogre", "paladin", "", ElementFire}
		for _, class := range invalidClasses {
			req.BaseClass = class
			err := CanEvolveAdvancedClass(req)
			if !errors.Is(err, ErrAdvancedClassInvalidBaseClass) {
				t.Errorf("for class '%s', expected ErrAdvancedClassInvalidBaseClass, got %v", class, err)
			}
		}
	})

	t.Run("rejects invalid element", func(t *testing.T) {
		req := validRequest
		invalidElements := []RuleID{"air", "water", "ogre", StartingClassNovice, ClassKnight, ""}
		for _, el := range invalidElements {
			req.Element = el
			err := CanEvolveAdvancedClass(req)
			if !errors.Is(err, ErrAdvancedClassInvalidElement) {
				t.Errorf("for element '%s', expected ErrAdvancedClassInvalidElement, got %v", el, err)
			}
		}
	})

	t.Run("rejects if character level too low", func(t *testing.T) {
		req := validRequest
		req.CharacterLevel = 99
		err := CanEvolveAdvancedClass(req)
		if !errors.Is(err, ErrAdvancedClassCharacterLevelTooLow) {
			t.Errorf("expected ErrAdvancedClassCharacterLevelTooLow, got %v", err)
		}
	})

	t.Run("rejects if affinity level too low", func(t *testing.T) {
		req := validRequest
		req.AffinityLevel = 99
		err := CanEvolveAdvancedClass(req)
		if !errors.Is(err, ErrAdvancedClassAffinityLevelTooLow) {
			t.Errorf("expected ErrAdvancedClassAffinityLevelTooLow, got %v", err)
		}
	})

	t.Run("rejects if quest not completed", func(t *testing.T) {
		req := validRequest
		req.QuestCompleted = false
		err := CanEvolveAdvancedClass(req)
		if !errors.Is(err, ErrAdvancedClassQuestNotCompleted) {
			t.Errorf("expected ErrAdvancedClassQuestNotCompleted, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		baseReq := AdvancedClassEvolutionRequest{
			CharacterLevel: 1,
			AffinityLevel:  1,
			QuestCompleted: false,
		}

		t.Run("invalid base class is first error", func(t *testing.T) {
			req := baseReq
			req.BaseClass = "ogre"
			req.Element = "air"
			err := CanEvolveAdvancedClass(req)
			if !errors.Is(err, ErrAdvancedClassInvalidBaseClass) {
				t.Errorf("expected ErrAdvancedClassInvalidBaseClass, got %v", err)
			}
		})

		t.Run("invalid element is second error", func(t *testing.T) {
			req := baseReq
			req.BaseClass = ClassKnight
			req.Element = "air"
			err := CanEvolveAdvancedClass(req)
			if !errors.Is(err, ErrAdvancedClassInvalidElement) {
				t.Errorf("expected ErrAdvancedClassInvalidElement, got %v", err)
			}
		})

		t.Run("character level is third error", func(t *testing.T) {
			req := baseReq
			req.BaseClass = ClassKnight
			req.Element = ElementFire
			err := CanEvolveAdvancedClass(req)
			if !errors.Is(err, ErrAdvancedClassCharacterLevelTooLow) {
				t.Errorf("expected ErrAdvancedClassCharacterLevelTooLow, got %v", err)
			}
		})
	})
}

func TestMustReachCharacterLevelForAdvancedClass(t *testing.T) {
	level := MustReachCharacterLevelForAdvancedClass()
	if level != 100 {
		t.Errorf("expected level 100, got %d", level)
	}
	if level != AdvancedClassMinimumCharacterLevel {
		t.Errorf("expected function to return AdvancedClassMinimumCharacterLevel constant")
	}
}

func TestMustReachAffinityLevelForAdvancedClass(t *testing.T) {
	level := MustReachAffinityLevelForAdvancedClass()
	if level != 100 {
		t.Errorf("expected level 100, got %d", level)
	}
	if level != AdvancedClassMinimumAffinityLevel {
		t.Errorf("expected function to return AdvancedClassMinimumAffinityLevel constant")
	}
}
