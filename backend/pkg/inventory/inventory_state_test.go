package inventory

import (
	"reflect"
	"sync"
	"testing"
)

func canonicalStateTestCatalog(t *testing.T) *ItemCatalog {
	t.Helper()
	catalog, err := NewItemCatalog([]ItemDefinition{
		{
			ID:        "sword_test",
			Name:      "Test Sword",
			Type:      ItemTypeWeapon,
			Category:  "Sword",
			Tier:      1,
			Stackable: false,
			MaxStack:  1,
		},
		{
			ID:        "material_test",
			Name:      "Test Material",
			Type:      ItemTypeMaterial,
			Tier:      0,
			Stackable: true,
			MaxStack:  20,
		},
	})
	if err != nil {
		t.Fatalf("NewItemCatalog() error = %v", err)
	}
	return catalog
}

func validCanonicalInventorySnapshot() InventoryStateSnapshot {
	return InventoryStateSnapshot{
		OwnerID:  "Gabriela",
		Revision: 7,
		Containers: []InventoryContainerSnapshot{
			{
				ID:       "equipment",
				Kind:     ContainerKindEquipment,
				Capacity: 8,
				Placements: []InventorySlotPlacement{
					{Slot: 2, InstanceID: "instance_sword"},
				},
			},
			{
				ID:       "backpack_main",
				Kind:     ContainerKindBackpack,
				Capacity: 30,
				Placements: []InventorySlotPlacement{
					{Slot: 5, InstanceID: "stack_ore"},
				},
			},
		},
		Items: []InventoryItemStack{
			{InstanceID: "stack_ore", DefinitionID: "material_test", Quantity: 10},
			{InstanceID: "instance_sword", DefinitionID: "sword_test", Quantity: 1},
		},
	}
}

func TestCanonicalInventoryStateValidAndDeterministic(t *testing.T) {
	state, err := NewInventoryState(canonicalStateTestCatalog(t), validCanonicalInventorySnapshot())
	if err != nil {
		t.Fatalf("NewInventoryState() error = %v", err)
	}

	if got := state.OwnerID(); got != "Gabriela" {
		t.Fatalf("OwnerID() = %q", got)
	}
	if got := state.Revision(); got != 7 {
		t.Fatalf("Revision() = %d", got)
	}
	if got := state.ContainerCount(); got != 2 {
		t.Fatalf("ContainerCount() = %d", got)
	}
	if got := state.ItemCount(); got != 2 {
		t.Fatalf("ItemCount() = %d", got)
	}

	wantContainers := []ContainerID{"backpack_main", "equipment"}
	if got := state.ContainerIDs(); !reflect.DeepEqual(got, wantContainers) {
		t.Fatalf("ContainerIDs() = %v, want %v", got, wantContainers)
	}
	wantItems := []ItemInstanceID{"instance_sword", "stack_ore"}
	if got := state.ItemInstanceIDs(); !reflect.DeepEqual(got, wantItems) {
		t.Fatalf("ItemInstanceIDs() = %v, want %v", got, wantItems)
	}

	item, ok := state.Item("instance_sword")
	if !ok || item.DefinitionID != "sword_test" || item.Quantity != 1 {
		t.Fatalf("Item(instance_sword) = %+v, %v", item, ok)
	}
	if _, ok := state.Item("missing"); ok {
		t.Fatal("unexpected missing item")
	}

	location, ok := state.LocationOf("stack_ore")
	if !ok || location.ContainerID != "backpack_main" || location.Slot != 5 {
		t.Fatalf("LocationOf(stack_ore) = %+v, %v", location, ok)
	}
	if _, ok := state.LocationOf("missing"); ok {
		t.Fatal("unexpected missing location")
	}

	container, ok := state.Container("equipment")
	if !ok || len(container.Placements) != 1 || container.Placements[0].Slot != 2 {
		t.Fatalf("Container(equipment) = %+v, %v", container, ok)
	}
	if _, ok := state.Container("missing"); ok {
		t.Fatal("unexpected missing container")
	}
}

func TestCanonicalInventoryStateDefensiveCopies(t *testing.T) {
	input := validCanonicalInventorySnapshot()
	state, err := NewInventoryState(canonicalStateTestCatalog(t), input)
	if err != nil {
		t.Fatal(err)
	}

	input.OwnerID = "Mutated"
	input.Containers[0].ID = "mutated"
	input.Containers[0].Placements[0].InstanceID = "mutated"
	input.Items[0].DefinitionID = "mutated"

	if state.OwnerID() != "Gabriela" {
		t.Fatal("state mutated through constructor input")
	}
	if _, ok := state.Container("equipment"); !ok {
		t.Fatal("container mutated through constructor input")
	}
	if item, ok := state.Item("stack_ore"); !ok || item.DefinitionID != "material_test" {
		t.Fatal("item mutated through constructor input")
	}

	ids := state.ContainerIDs()
	ids[0] = "mutated"
	if state.ContainerIDs()[0] == "mutated" {
		t.Fatal("ContainerIDs exposed internal storage")
	}

	container, _ := state.Container("equipment")
	container.Placements[0].InstanceID = "mutated"
	again, _ := state.Container("equipment")
	if again.Placements[0].InstanceID == "mutated" {
		t.Fatal("Container exposed internal placements")
	}

	snapshot := state.Snapshot()
	snapshot.OwnerID = "Mutated"
	snapshot.Containers[0].Placements = nil
	snapshot.Items[0].DefinitionID = "mutated"
	fresh := state.Snapshot()
	if fresh.OwnerID != "Gabriela" || len(fresh.Containers[0].Placements) == 0 || fresh.Items[0].DefinitionID == "mutated" {
		t.Fatal("Snapshot exposed internal state")
	}
}

func TestCanonicalInventoryStateAllowsEmptyInventory(t *testing.T) {
	snapshot := validCanonicalInventorySnapshot()
	snapshot.Items = nil
	for index := range snapshot.Containers {
		snapshot.Containers[index].Placements = nil
	}
	state, err := NewInventoryState(canonicalStateTestCatalog(t), snapshot)
	if err != nil {
		t.Fatalf("empty inventory rejected: %v", err)
	}
	if state.ItemCount() != 0 {
		t.Fatalf("ItemCount() = %d", state.ItemCount())
	}
}

func TestCanonicalInventoryStateValidation(t *testing.T) {
	catalog := canonicalStateTestCatalog(t)
	tests := []struct {
		name    string
		catalog *ItemCatalog
		mutate  func(*InventoryStateSnapshot)
	}{
		{"nil catalog", nil, func(*InventoryStateSnapshot) {}},
		{"empty owner", catalog, func(s *InventoryStateSnapshot) { s.OwnerID = "" }},
		{"spaced owner", catalog, func(s *InventoryStateSnapshot) { s.OwnerID = " Gabriela" }},
		{"zero revision", catalog, func(s *InventoryStateSnapshot) { s.Revision = 0 }},
		{"no containers", catalog, func(s *InventoryStateSnapshot) { s.Containers = nil }},
		{"missing backpack", catalog, func(s *InventoryStateSnapshot) { s.Containers = s.Containers[:1]; s.Items = s.Items[1:] }},
		{"missing equipment", catalog, func(s *InventoryStateSnapshot) { s.Containers = s.Containers[1:]; s.Items = s.Items[:1] }},
		{"duplicate container", catalog, func(s *InventoryStateSnapshot) { s.Containers[1].ID = s.Containers[0].ID }},
		{"invalid container ID", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].ID = "Equipment" }},
		{"invalid container kind", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Kind = "bank" }},
		{"zero capacity", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Capacity = 0 }},
		{"capacity above maximum", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Capacity = MaxCanonicalContainerSlots + 1 }},
		{"slot outside capacity", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Placements[0].Slot = SlotID(s.Containers[0].Capacity) }},
		{"duplicate slot", catalog, func(s *InventoryStateSnapshot) {
			s.Containers[1].Placements = append(s.Containers[1].Placements, InventorySlotPlacement{Slot: 5, InstanceID: "instance_sword"})
			s.Containers[0].Placements = nil
		}},
		{"invalid placement instance ID", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Placements[0].InstanceID = "Instance Sword" }},
		{"duplicate item instance", catalog, func(s *InventoryStateSnapshot) { s.Items[1].InstanceID = s.Items[0].InstanceID }},
		{"invalid item instance ID", catalog, func(s *InventoryStateSnapshot) { s.Items[0].InstanceID = "Stack Ore" }},
		{"unknown definition", catalog, func(s *InventoryStateSnapshot) { s.Items[0].DefinitionID = "missing" }},
		{"zero quantity", catalog, func(s *InventoryStateSnapshot) { s.Items[0].Quantity = 0 }},
		{"negative quantity", catalog, func(s *InventoryStateSnapshot) { s.Items[0].Quantity = -1 }},
		{"non-stackable quantity", catalog, func(s *InventoryStateSnapshot) { s.Items[1].Quantity = 2 }},
		{"stack exceeds max", catalog, func(s *InventoryStateSnapshot) { s.Items[0].Quantity = 21 }},
		{"placement references unknown item", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Placements[0].InstanceID = "missing_instance" }},
		{"item placed twice", catalog, func(s *InventoryStateSnapshot) {
			s.Containers[1].Placements = append(s.Containers[1].Placements, InventorySlotPlacement{Slot: 6, InstanceID: "instance_sword"})
		}},
		{"unplaced item", catalog, func(s *InventoryStateSnapshot) { s.Containers[0].Placements = nil }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := validCanonicalInventorySnapshot()
			test.mutate(&snapshot)
			if _, err := NewInventoryState(test.catalog, snapshot); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCanonicalInventoryStateConcurrentReads(t *testing.T) {
	state, err := NewInventoryState(canonicalStateTestCatalog(t), validCanonicalInventorySnapshot())
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for worker := 0; worker < 64; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < 1000; iteration++ {
				if _, ok := state.Item("instance_sword"); !ok {
					t.Errorf("missing item")
					return
				}
				if _, ok := state.Container("backpack_main"); !ok {
					t.Errorf("missing container")
					return
				}
				if _, ok := state.LocationOf("stack_ore"); !ok {
					t.Errorf("missing location")
					return
				}
				if got := state.Snapshot(); len(got.Containers) != 2 || len(got.Items) != 2 {
					t.Errorf("invalid snapshot")
					return
				}
			}
		}()
	}
	wg.Wait()
}
