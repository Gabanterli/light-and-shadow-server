package inventory

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func canonicalOperationTestCatalog(t *testing.T) *ItemCatalog {
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
		{
			ID:        "material_other",
			Name:      "Other Material",
			Type:      ItemTypeMaterial,
			Tier:      0,
			Stackable: true,
			MaxStack:  10,
		},
	})
	if err != nil {
		t.Fatalf("NewItemCatalog() error = %v", err)
	}
	return catalog
}

func canonicalOperationTestState(t *testing.T, catalog *ItemCatalog) *InventoryState {
	t.Helper()
	state, err := NewInventoryState(catalog, InventoryStateSnapshot{
		OwnerID:  "Gabriela",
		Revision: 11,
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
					{Slot: 5, InstanceID: "stack_ore_a"},
					{Slot: 6, InstanceID: "stack_ore_b"},
					{Slot: 7, InstanceID: "stack_herb"},
				},
			},
		},
		Items: []InventoryItemStack{
			{InstanceID: "instance_sword", DefinitionID: "sword_test", Quantity: 1},
			{InstanceID: "stack_ore_a", DefinitionID: "material_test", Quantity: 10},
			{InstanceID: "stack_ore_b", DefinitionID: "material_test", Quantity: 5},
			{InstanceID: "stack_herb", DefinitionID: "material_other", Quantity: 4},
		},
	})
	if err != nil {
		t.Fatalf("NewInventoryState() error = %v", err)
	}
	return state
}

func TestCanonicalInventoryOperationMove(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)
	original := current.Snapshot()

	next, result, err := ApplyInventoryOperation(
		catalog,
		current,
		11,
		MoveInventoryItemOperation{
			InstanceID: "instance_sword",
			To: InventoryItemLocation{
				ContainerID: "backpack_main",
				Slot:        0,
			},
		},
	)
	if err != nil {
		t.Fatalf("ApplyInventoryOperation() error = %v", err)
	}
	if got := next.Revision(); got != 12 {
		t.Fatalf("Revision() = %d, want 12", got)
	}
	if got, ok := next.LocationOf("instance_sword"); !ok || got.ContainerID != "backpack_main" || got.Slot != 0 {
		t.Fatalf("LocationOf(instance_sword) = %+v, %v", got, ok)
	}
	if result.Kind != InventoryOperationMove || result.PreviousRevision != 11 || result.NewRevision != 12 {
		t.Fatalf("result = %+v", result)
	}
	if !reflect.DeepEqual(result.ChangedInstanceIDs, []ItemInstanceID{"instance_sword"}) {
		t.Fatalf("ChangedInstanceIDs = %v", result.ChangedInstanceIDs)
	}
	assertCanonicalOperationOriginalUnchanged(t, current, original)
	assertCanonicalOperationQuantityTotalsEqual(t, current, next)
}

func TestCanonicalInventoryOperationSwap(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)

	next, result, err := ApplyInventoryOperation(
		catalog,
		current,
		11,
		SwapInventoryItemsOperation{
			FirstInstanceID:  "instance_sword",
			SecondInstanceID: "stack_ore_a",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := next.LocationOf("instance_sword"); got.ContainerID != "backpack_main" || got.Slot != 5 {
		t.Fatalf("sword location = %+v", got)
	}
	if got, _ := next.LocationOf("stack_ore_a"); got.ContainerID != "equipment" || got.Slot != 2 {
		t.Fatalf("ore location = %+v", got)
	}
	wantChanged := []ItemInstanceID{"instance_sword", "stack_ore_a"}
	if !reflect.DeepEqual(result.ChangedInstanceIDs, wantChanged) {
		t.Fatalf("ChangedInstanceIDs = %v, want %v", result.ChangedInstanceIDs, wantChanged)
	}
	assertCanonicalOperationQuantityTotalsEqual(t, current, next)
}

func TestCanonicalInventoryOperationMergePartialAndFull(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)

	t.Run("partial", func(t *testing.T) {
		current := canonicalOperationTestState(t, catalog)
		next, result, err := ApplyInventoryOperation(
			catalog,
			current,
			11,
			MergeInventoryStacksOperation{
				SourceInstanceID:      "stack_ore_b",
				DestinationInstanceID: "stack_ore_a",
				Quantity:              2,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if source, _ := next.Item("stack_ore_b"); source.Quantity != 3 {
			t.Fatalf("source quantity = %d", source.Quantity)
		}
		if destination, _ := next.Item("stack_ore_a"); destination.Quantity != 12 {
			t.Fatalf("destination quantity = %d", destination.Quantity)
		}
		wantChanged := []ItemInstanceID{"stack_ore_a", "stack_ore_b"}
		if !reflect.DeepEqual(result.ChangedInstanceIDs, wantChanged) {
			t.Fatalf("ChangedInstanceIDs = %v, want %v", result.ChangedInstanceIDs, wantChanged)
		}
		if len(result.RemovedInstanceIDs) != 0 {
			t.Fatalf("RemovedInstanceIDs = %v", result.RemovedInstanceIDs)
		}
		assertCanonicalOperationQuantityTotalsEqual(t, current, next)
	})

	t.Run("full", func(t *testing.T) {
		current := canonicalOperationTestState(t, catalog)
		next, result, err := ApplyInventoryOperation(
			catalog,
			current,
			11,
			MergeInventoryStacksOperation{
				SourceInstanceID:      "stack_ore_b",
				DestinationInstanceID: "stack_ore_a",
				Quantity:              5,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := next.Item("stack_ore_b"); ok {
			t.Fatal("source stack still exists")
		}
		if _, ok := next.LocationOf("stack_ore_b"); ok {
			t.Fatal("source location still exists")
		}
		if destination, _ := next.Item("stack_ore_a"); destination.Quantity != 15 {
			t.Fatalf("destination quantity = %d", destination.Quantity)
		}
		if !reflect.DeepEqual(result.RemovedInstanceIDs, []ItemInstanceID{"stack_ore_b"}) {
			t.Fatalf("RemovedInstanceIDs = %v", result.RemovedInstanceIDs)
		}
		assertCanonicalOperationQuantityTotalsEqual(t, current, next)
	})
}

func TestCanonicalInventoryOperationSplit(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)

	next, result, err := ApplyInventoryOperation(
		catalog,
		current,
		11,
		SplitInventoryStackOperation{
			SourceInstanceID: "stack_ore_a",
			NewInstanceID:    "stack_ore_split",
			Quantity:         4,
			To: InventoryItemLocation{
				ContainerID: "backpack_main",
				Slot:        8,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if source, _ := next.Item("stack_ore_a"); source.Quantity != 6 {
		t.Fatalf("source quantity = %d", source.Quantity)
	}
	if split, ok := next.Item("stack_ore_split"); !ok || split.DefinitionID != "material_test" || split.Quantity != 4 {
		t.Fatalf("split item = %+v, %v", split, ok)
	}
	if location, _ := next.LocationOf("stack_ore_split"); location.ContainerID != "backpack_main" || location.Slot != 8 {
		t.Fatalf("split location = %+v", location)
	}
	if !reflect.DeepEqual(result.CreatedInstanceIDs, []ItemInstanceID{"stack_ore_split"}) {
		t.Fatalf("CreatedInstanceIDs = %v", result.CreatedInstanceIDs)
	}
	if !reflect.DeepEqual(result.ChangedInstanceIDs, []ItemInstanceID{"stack_ore_a"}) {
		t.Fatalf("ChangedInstanceIDs = %v", result.ChangedInstanceIDs)
	}
	assertCanonicalOperationQuantityTotalsEqual(t, current, next)
}

func TestCanonicalInventoryOperationAddAndRemove(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)

	t.Run("add", func(t *testing.T) {
		current := canonicalOperationTestState(t, catalog)
		next, result, err := ApplyInventoryOperation(
			catalog,
			current,
			11,
			AddInventoryItemOperation{
				InstanceID:   "stack_added",
				DefinitionID: "material_test",
				Quantity:     7,
				To: InventoryItemLocation{
					ContainerID: "backpack_main",
					Slot:        9,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if item, ok := next.Item("stack_added"); !ok || item.Quantity != 7 {
			t.Fatalf("added item = %+v, %v", item, ok)
		}
		if !reflect.DeepEqual(result.CreatedInstanceIDs, []ItemInstanceID{"stack_added"}) {
			t.Fatalf("CreatedInstanceIDs = %v", result.CreatedInstanceIDs)
		}
		want := canonicalOperationDefinitionTotals(current)
		want["material_test"] += 7
		if got := canonicalOperationDefinitionTotals(next); !reflect.DeepEqual(got, want) {
			t.Fatalf("totals = %v, want %v", got, want)
		}
	})

	t.Run("remove partial", func(t *testing.T) {
		current := canonicalOperationTestState(t, catalog)
		next, result, err := ApplyInventoryOperation(
			catalog,
			current,
			11,
			RemoveInventoryItemOperation{InstanceID: "stack_ore_a", Quantity: 3},
		)
		if err != nil {
			t.Fatal(err)
		}
		if item, _ := next.Item("stack_ore_a"); item.Quantity != 7 {
			t.Fatalf("remaining quantity = %d", item.Quantity)
		}
		if !reflect.DeepEqual(result.ChangedInstanceIDs, []ItemInstanceID{"stack_ore_a"}) {
			t.Fatalf("ChangedInstanceIDs = %v", result.ChangedInstanceIDs)
		}
		want := canonicalOperationDefinitionTotals(current)
		want["material_test"] -= 3
		if got := canonicalOperationDefinitionTotals(next); !reflect.DeepEqual(got, want) {
			t.Fatalf("totals = %v, want %v", got, want)
		}
	})

	t.Run("remove full", func(t *testing.T) {
		current := canonicalOperationTestState(t, catalog)
		next, result, err := ApplyInventoryOperation(
			catalog,
			current,
			11,
			RemoveInventoryItemOperation{InstanceID: "instance_sword", Quantity: 1},
		)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := next.Item("instance_sword"); ok {
			t.Fatal("removed sword still exists")
		}
		if !reflect.DeepEqual(result.RemovedInstanceIDs, []ItemInstanceID{"instance_sword"}) {
			t.Fatalf("RemovedInstanceIDs = %v", result.RemovedInstanceIDs)
		}
	})
}

func TestCanonicalInventoryOperationRevisionAndAtomicFailure(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)
	original := current.Snapshot()

	tests := []struct {
		name      string
		expected  InventoryRevision
		operation InventoryOperation
		wantError error
	}{
		{
			name:      "stale revision",
			expected:  10,
			operation: MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 0}},
			wantError: ErrInventoryRevisionConflict,
		},
		{
			name:      "occupied target",
			expected:  11,
			operation: MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 5}},
			wantError: ErrInventorySlotOccupied,
		},
		{
			name:      "move no-op",
			expected:  11,
			operation: MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "equipment", Slot: 2}},
			wantError: ErrInventoryOperationNoOp,
		},
		{
			name:      "merge overflow",
			expected:  11,
			operation: MergeInventoryStacksOperation{SourceInstanceID: "stack_ore_b", DestinationInstanceID: "stack_ore_a", Quantity: 11},
			wantError: ErrInventoryStackRule,
		},
		{
			name:      "merge definition mismatch",
			expected:  11,
			operation: MergeInventoryStacksOperation{SourceInstanceID: "stack_herb", DestinationInstanceID: "stack_ore_a", Quantity: 1},
			wantError: ErrInventoryDefinitionMismatch,
		},
		{
			name:      "split duplicate ID",
			expected:  11,
			operation: SplitInventoryStackOperation{SourceInstanceID: "stack_ore_a", NewInstanceID: "stack_ore_b", Quantity: 1, To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 8}},
			wantError: ErrInventoryInstanceExists,
		},
		{
			name:      "split entire source",
			expected:  11,
			operation: SplitInventoryStackOperation{SourceInstanceID: "stack_ore_a", NewInstanceID: "stack_new", Quantity: 10, To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 8}},
			wantError: ErrInventoryStackRule,
		},
		{
			name:      "add unknown definition",
			expected:  11,
			operation: AddInventoryItemOperation{InstanceID: "new_item", DefinitionID: "missing", Quantity: 1, To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 8}},
			wantError: ErrInventoryOperationInvalid,
		},
		{
			name:      "add non-stackable quantity",
			expected:  11,
			operation: AddInventoryItemOperation{InstanceID: "new_sword", DefinitionID: "sword_test", Quantity: 2, To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 8}},
			wantError: ErrInventoryStackRule,
		},
		{
			name:      "remove too much",
			expected:  11,
			operation: RemoveInventoryItemOperation{InstanceID: "stack_ore_b", Quantity: 6},
			wantError: ErrInventoryStackRule,
		},
		{
			name:      "missing container",
			expected:  11,
			operation: MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "missing", Slot: 0}},
			wantError: ErrInventoryContainerNotFound,
		},
		{
			name:      "slot outside capacity",
			expected:  11,
			operation: MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "equipment", Slot: 8}},
			wantError: ErrInventorySlotOutOfRange,
		},
		{
			name:      "missing item",
			expected:  11,
			operation: RemoveInventoryItemOperation{InstanceID: "missing_item", Quantity: 1},
			wantError: ErrInventoryItemNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next, result, err := ApplyInventoryOperation(catalog, current, test.expected, test.operation)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, test.wantError)
			}
			if next != nil {
				t.Fatalf("next state = %+v, want nil", next)
			}
			if !reflect.DeepEqual(result, InventoryOperationResult{}) {
				t.Fatalf("result = %+v, want zero", result)
			}
			assertCanonicalOperationOriginalUnchanged(t, current, original)
		})
	}
}

func TestCanonicalInventoryOperationRejectsInvalidInputs(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)

	var nilMove *MoveInventoryItemOperation
	tests := []struct {
		name      string
		catalog   *ItemCatalog
		current   *InventoryState
		operation InventoryOperation
		wantError error
	}{
		{"nil catalog", nil, current, MoveInventoryItemOperation{}, ErrInventoryOperationInvalid},
		{"nil state", catalog, nil, MoveInventoryItemOperation{}, ErrInventoryOperationInvalid},
		{"nil operation", catalog, current, nil, ErrInventoryOperationInvalid},
		{"typed nil operation", catalog, current, nilMove, ErrInventoryOperationInvalid},
		{"invalid instance ID", catalog, current, MoveInventoryItemOperation{InstanceID: "Bad ID", To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 0}}, ErrInventoryOperationInvalid},
		{"invalid container ID", catalog, current, MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "Bad ID", Slot: 0}}, ErrInventoryOperationInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := ApplyInventoryOperation(test.catalog, test.current, 11, test.operation)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, test.wantError)
			}
		})
	}
}

func TestCanonicalInventoryOperationRevisionOverflow(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	snapshot := canonicalOperationTestState(t, catalog).Snapshot()
	snapshot.Revision = InventoryRevision(^uint64(0))
	state, err := NewInventoryState(catalog, snapshot)
	if err != nil {
		t.Fatal(err)
	}

	next, result, err := ApplyInventoryOperation(
		catalog,
		state,
		state.Revision(),
		MoveInventoryItemOperation{InstanceID: "instance_sword", To: InventoryItemLocation{ContainerID: "backpack_main", Slot: 0}},
	)
	if !errors.Is(err, ErrInventoryRevisionOverflow) {
		t.Fatalf("error = %v", err)
	}
	if next != nil || !reflect.DeepEqual(result, InventoryOperationResult{}) {
		t.Fatalf("next = %+v, result = %+v", next, result)
	}
}

func TestCanonicalInventoryOperationConcurrentPureApplications(t *testing.T) {
	catalog := canonicalOperationTestCatalog(t)
	current := canonicalOperationTestState(t, catalog)
	original := current.Snapshot()

	var wg sync.WaitGroup
	for worker := 0; worker < 64; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < 200; iteration++ {
				next, result, err := ApplyInventoryOperation(
					catalog,
					current,
					11,
					MoveInventoryItemOperation{
						InstanceID: "instance_sword",
						To:         InventoryItemLocation{ContainerID: "backpack_main", Slot: 0},
					},
				)
				if err != nil {
					t.Errorf("ApplyInventoryOperation() error = %v", err)
					return
				}
				if next.Revision() != 12 || result.NewRevision != 12 {
					t.Errorf("revision mismatch: state=%d result=%d", next.Revision(), result.NewRevision)
					return
				}
			}
		}()
	}
	wg.Wait()
	assertCanonicalOperationOriginalUnchanged(t, current, original)
}

func assertCanonicalOperationOriginalUnchanged(
	t *testing.T,
	state *InventoryState,
	want InventoryStateSnapshot,
) {
	t.Helper()
	if got := state.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("original state mutated:\ngot  %+v\nwant %+v", got, want)
	}
}

func assertCanonicalOperationQuantityTotalsEqual(t *testing.T, before, after *InventoryState) {
	t.Helper()
	if beforeTotals, afterTotals := canonicalOperationDefinitionTotals(before), canonicalOperationDefinitionTotals(after); !reflect.DeepEqual(beforeTotals, afterTotals) {
		t.Fatalf("quantity totals changed: before=%v after=%v", beforeTotals, afterTotals)
	}
}

func canonicalOperationDefinitionTotals(state *InventoryState) map[string]int {
	totals := make(map[string]int)
	for _, item := range state.Snapshot().Items {
		totals[item.DefinitionID] += item.Quantity
	}
	return totals
}
