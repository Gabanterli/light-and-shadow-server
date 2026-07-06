package rules

import (
	"testing"
)

func TestDefaultDefinitions_Count(t *testing.T) {
	defs := DefaultDefinitions()
	expectedCount := 15 // 5 raças + 5 classes + 5 elementos
	if len(defs) != expectedCount {
		t.Errorf("expected %d default definitions, but got %d", expectedCount, len(defs))
	}
}

func TestNewDefaultRegistry_Creation(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() returned an unexpected error: %v", err)
	}

	expectedCount := 15
	if r.Count() != expectedCount {
		t.Errorf("expected registry count to be %d, but got %d", expectedCount, r.Count())
	}
}

func TestDefaultRegistry_ContainsOfficialRules(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("Failed to create default registry: %v", err)
	}

	t.Run("Contains all playable races", func(t *testing.T) {
		races := []RuleID{RaceHuman, RaceForestElf, RaceDwarf, RaceIceElf, RaceGreenOrc}
		for _, id := range races {
			def, ok := r.Get(id)
			if !ok {
				t.Errorf("expected to find race '%s', but it was not in the registry", id)
			}
			if def.Category != CategoryRace {
				t.Errorf("rule '%s' expected to have category '%s', but got '%s'", id, CategoryRace, def.Category)
			}
		}
	})

	t.Run("Contains all base classes", func(t *testing.T) {
		classes := []RuleID{ClassKnight, ClassMage, ClassArcher, ClassAssassin, ClassCleric}
		for _, id := range classes {
			def, ok := r.Get(id)
			if !ok {
				t.Errorf("expected to find class '%s', but it was not in the registry", id)
			}
			if def.Category != CategoryClass {
				t.Errorf("rule '%s' expected to have category '%s', but got '%s'", id, CategoryClass, def.Category)
			}
		}
	})

	t.Run("Contains all elements", func(t *testing.T) {
		elements := []RuleID{ElementFire, ElementEarth, ElementIce, ElementShadow, ElementSacred}
		for _, id := range elements {
			def, ok := r.Get(id)
			if !ok {
				t.Errorf("expected to find element '%s', but it was not in the registry", id)
			}
			if def.Category != CategoryElement {
				t.Errorf("rule '%s' expected to have category '%s', but got '%s'", id, CategoryElement, def.Category)
			}
		}
	})
}

func TestDefaultRegistry_DoesNotContainDisallowedRules(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("Failed to create default registry: %v", err)
	}

	disallowedIDs := []RuleID{"ogre", "air", "water"}
	for _, id := range disallowedIDs {
		if _, ok := r.Get(id); ok {
			t.Errorf("found disallowed rule '%s' in the registry", id)
		}
	}
}

func TestDefaultDefinitions_ReturnsCopy(t *testing.T) {
	// Chamar uma vez e modificar o resultado
	defs1 := DefaultDefinitions()
	if len(defs1) == 0 {
		t.Fatal("DefaultDefinitions returned an empty slice")
	}
	originalName := defs1[0].DisplayName
	defs1[0].DisplayName = "MUTATED"

	// Chamar de novo e verificar se não foi alterado
	defs2 := DefaultDefinitions()
	if len(defs2) == 0 {
		t.Fatal("Second call to DefaultDefinitions returned an empty slice")
	}

	if defs2[0].DisplayName == "MUTATED" {
		t.Error("modifying a returned slice from DefaultDefinitions mutated the internal source")
	}
	if defs2[0].DisplayName != originalName {
		t.Errorf("expected display name to be '%s', but got '%s'", originalName, defs2[0].DisplayName)
	}
}

func TestDefaultDefinitions_AllFieldsArePopulated(t *testing.T) {
	defs := DefaultDefinitions()

	if len(defs) == 0 {
		t.Fatal("DefaultDefinitions should not be empty")
	}

	for _, def := range defs {
		if def.ID == "" {
			t.Errorf("found definition with empty ID: %+v", def)
		}
		if def.Category == "" {
			t.Errorf("found definition with empty Category for ID '%s'", def.ID)
		}
		if def.DisplayName == "" {
			t.Errorf("found definition with empty DisplayName for ID '%s'", def.ID)
		}
		if def.Description == "" {
			t.Errorf("found definition with empty Description for ID '%s'", def.ID)
		}
	}
}
