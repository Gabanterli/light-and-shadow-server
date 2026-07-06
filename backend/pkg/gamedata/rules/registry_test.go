package rules

import (
	"strings"
	"testing"
)

func TestNewRegistryValid(t *testing.T) {
	defs := []RuleDefinition{
		{ID: "test_rule_1", Category: CategorySystem, DisplayName: "Test Rule 1"},
		{ID: "another_rule", Category: CategoryClass, DisplayName: "Another Rule"},
	}

	registry, err := NewRegistry(defs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if registry.Count() != 2 {
		t.Fatalf("expected count 2, got %d", registry.Count())
	}
}

func TestRegistryGetExistingRule(t *testing.T) {
	registry, err := NewRegistry([]RuleDefinition{
		{ID: "find_me", Category: CategoryRace, DisplayName: "Find Me"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	def, ok := registry.Get("find_me")
	if !ok {
		t.Fatal("expected to find rule")
	}

	if def.ID != "find_me" {
		t.Fatalf("expected ID find_me, got %s", def.ID)
	}

	if def.DisplayName != "Find Me" {
		t.Fatalf("expected display name Find Me, got %s", def.DisplayName)
	}
}

func TestRegistryGetMissingRule(t *testing.T) {
	registry, err := NewRegistry([]RuleDefinition{
		{ID: "existing_rule", Category: CategorySystem, DisplayName: "Existing Rule"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, ok := registry.Get("missing_rule")
	if ok {
		t.Fatal("expected missing rule to return false")
	}
}

func TestNewRegistryRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name          string
		definitions   []RuleDefinition
		expectedError string
	}{
		{
			name: "empty ID",
			definitions: []RuleDefinition{
				{ID: "", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "rule ID cannot be empty",
		},
		{
			name: "uppercase ID",
			definitions: []RuleDefinition{
				{ID: "Invalid_ID", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "invalid format",
		},
		{
			name: "hyphen ID",
			definitions: []RuleDefinition{
				{ID: "invalid-id", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "invalid format",
		},
		{
			name: "starts with number",
			definitions: []RuleDefinition{
				{ID: "1invalid", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "invalid format",
		},
		{
			name: "starts with underscore",
			definitions: []RuleDefinition{
				{ID: "_invalid", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "invalid format",
		},
		{
			name: "ends with underscore",
			definitions: []RuleDefinition{
				{ID: "invalid_", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "invalid format",
		},
		{
			name: "double underscore",
			definitions: []RuleDefinition{
				{ID: "invalid__id", Category: CategorySystem, DisplayName: "Invalid"},
			},
			expectedError: "cannot contain consecutive underscores",
		},
		{
			name: "empty category",
			definitions: []RuleDefinition{
				{ID: "valid_id", Category: "", DisplayName: "Invalid"},
			},
			expectedError: "empty category",
		},
		{
			name: "empty display name",
			definitions: []RuleDefinition{
				{ID: "valid_id", Category: CategorySystem, DisplayName: ""},
			},
			expectedError: "empty display name",
		},
		{
			name: "duplicate ID",
			definitions: []RuleDefinition{
				{ID: "duplicate_id", Category: CategorySystem, DisplayName: "First"},
				{ID: "duplicate_id", Category: CategoryClass, DisplayName: "Second"},
			},
			expectedError: "duplicate rule ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRegistry(tt.definitions)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestRegistryListDeterministicOrder(t *testing.T) {
	registry, err := NewRegistry([]RuleDefinition{
		{ID: "z_system_rule", Category: CategorySystem, DisplayName: "Z System"},
		{ID: "a_system_rule", Category: CategorySystem, DisplayName: "A System"},
		{ID: "b_class_rule", Category: CategoryClass, DisplayName: "B Class"},
		{ID: "a_class_rule", Category: CategoryClass, DisplayName: "A Class"},
		{ID: "c_race_rule", Category: CategoryRace, DisplayName: "C Race"},
		{ID: "a_element_rule", Category: CategoryElement, DisplayName: "A Element"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	list := registry.List()

	expected := []struct {
		category RuleCategory
		id       RuleID
	}{
		{CategoryClass, "a_class_rule"},
		{CategoryClass, "b_class_rule"},
		{CategoryElement, "a_element_rule"},
		{CategoryRace, "c_race_rule"},
		{CategorySystem, "a_system_rule"},
		{CategorySystem, "z_system_rule"},
	}

	if len(list) != len(expected) {
		t.Fatalf("expected list length %d, got %d", len(expected), len(list))
	}

	for i, item := range expected {
		if list[i].Category != item.category || list[i].ID != item.id {
			t.Fatalf(
				"unexpected item at index %d: expected category=%s id=%s, got category=%s id=%s",
				i,
				item.category,
				item.id,
				list[i].Category,
				list[i].ID,
			)
		}
	}
}

func TestRegistryListReturnsCopy(t *testing.T) {
	registry, err := NewRegistry([]RuleDefinition{
		{ID: "copy_rule", Category: CategorySystem, DisplayName: "Copy Rule"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	list := registry.List()
	if len(list) != 1 {
		t.Fatalf("expected one rule, got %d", len(list))
	}

	list[0].DisplayName = "Mutated Rule"

	original, ok := registry.Get("copy_rule")
	if !ok {
		t.Fatal("expected original rule to exist")
	}

	if original.DisplayName != "Copy Rule" {
		t.Fatalf("expected internal registry to remain unchanged, got %s", original.DisplayName)
	}
}
