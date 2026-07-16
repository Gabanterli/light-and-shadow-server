package inventory

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func canonicalInventoryStoreTestCatalog(t *testing.T) *ItemCatalog {
	t.Helper()
	catalog, err := NewItemCatalog([]ItemDefinition{
		{
			ID:        "store_material",
			Name:      "Store Material",
			Type:      ItemTypeMaterial,
			Tier:      0,
			Stackable: true,
			MaxStack:  100,
		},
		{
			ID:        "store_sword",
			Name:      "Store Sword",
			Type:      ItemTypeWeapon,
			Category:  "Sword",
			Tier:      1,
			Stackable: false,
			MaxStack:  1,
		},
	})
	if err != nil {
		t.Fatalf("NewItemCatalog() error = %v", err)
	}
	return catalog
}

func canonicalInventoryStoreTestState(
	t *testing.T,
	catalog *ItemCatalog,
	ownerID string,
	revision InventoryRevision,
) *InventoryState {
	t.Helper()
	state, err := NewInventoryState(catalog, InventoryStateSnapshot{
		OwnerID:  ownerID,
		Revision: revision,
		Containers: []InventoryContainerSnapshot{
			{
				ID:       "backpack_main",
				Kind:     ContainerKindBackpack,
				Capacity: 128,
				Placements: []InventorySlotPlacement{
					{Slot: 0, InstanceID: "stack_material"},
				},
			},
			{
				ID:       "equipment",
				Kind:     ContainerKindEquipment,
				Capacity: 8,
			},
		},
		Items: []InventoryItemStack{
			{
				InstanceID:   "stack_material",
				DefinitionID: "store_material",
				Quantity:     10,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewInventoryState() error = %v", err)
	}
	return state
}

func canonicalInventoryStoreMove(slot SlotID) MoveInventoryItemOperation {
	return MoveInventoryItemOperation{
		InstanceID: "stack_material",
		To: InventoryItemLocation{
			ContainerID: "backpack_main",
			Slot:        slot,
		},
	}
}

func TestCanonicalInventoryStoreRegisterSnapshotAndDefensiveCopies(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if store.Count() != 0 || len(store.OwnerIDs()) != 0 {
		t.Fatal("new store is not empty")
	}

	state := canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)
	if err := store.Register(state); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := store.Register(state); !errors.Is(err, ErrInventoryStoreOwnerAlreadyExists) {
		t.Fatalf("duplicate Register() error = %v", err)
	}
	if store.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", store.Count())
	}
	if got := store.OwnerIDs(); !reflect.DeepEqual(got, []string{"Gabriela"}) {
		t.Fatalf("OwnerIDs() = %v", got)
	}

	snapshot, err := store.Snapshot("Gabriela")
	if err != nil {
		t.Fatal(err)
	}
	snapshot.OwnerID = "Mutated"
	snapshot.Items[0].Quantity = 99
	snapshot.Containers[0].Placements = nil

	fresh, err := store.Snapshot("Gabriela")
	if err != nil {
		t.Fatal(err)
	}
	if fresh.OwnerID != "Gabriela" || fresh.Items[0].Quantity != 10 {
		t.Fatalf("store mutated through snapshot: %+v", fresh)
	}
	location := canonicalInventoryStoreLocation(t, fresh, "stack_material")
	if location.ContainerID != "backpack_main" || location.Slot != 0 {
		t.Fatalf("location = %+v", location)
	}
}

func TestCanonicalInventoryStoreApplyPublishesOnlySuccess(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	original := canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)
	if err := store.Register(original); err != nil {
		t.Fatal(err)
	}

	next, result, err := store.Apply("Gabriela", 1, canonicalInventoryStoreMove(1))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if next.Revision != 2 || result.PreviousRevision != 1 || result.NewRevision != 2 {
		t.Fatalf("next/result revision = %d / %+v", next.Revision, result)
	}
	if got := canonicalInventoryStoreLocation(t, next, "stack_material"); got.Slot != 1 {
		t.Fatalf("published location = %+v", got)
	}
	if got, _ := original.LocationOf("stack_material"); got.Slot != 0 {
		t.Fatalf("registered input was mutated: %+v", got)
	}

	if _, _, err := store.Apply("Gabriela", 1, canonicalInventoryStoreMove(2)); !errors.Is(err, ErrInventoryRevisionConflict) {
		t.Fatalf("stale Apply() error = %v", err)
	}
	if _, _, err := store.Apply("Gabriela", 2, canonicalInventoryStoreMove(1)); !errors.Is(err, ErrInventoryOperationNoOp) {
		t.Fatalf("no-op Apply() error = %v", err)
	}

	fresh, err := store.Snapshot("Gabriela")
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Revision != 2 || canonicalInventoryStoreLocation(t, fresh, "stack_material").Slot != 1 {
		t.Fatalf("failed operation changed published state: %+v", fresh)
	}
}

func TestCanonicalInventoryStoreSerializesSameOwnerRevision(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)); err != nil {
		t.Fatal(err)
	}

	const workers = 64
	start := make(chan struct{})
	results := make(chan SlotID, workers)
	errorsChannel := make(chan error, workers)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		slot := SlotID(worker + 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _, applyErr := store.Apply("Gabriela", 1, canonicalInventoryStoreMove(slot))
			if applyErr == nil {
				results <- slot
				return
			}
			errorsChannel <- applyErr
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errorsChannel)

	var successfulSlots []SlotID
	for slot := range results {
		successfulSlots = append(successfulSlots, slot)
	}
	if len(successfulSlots) != 1 {
		t.Fatalf("successful operations = %d, want 1", len(successfulSlots))
	}
	for applyErr := range errorsChannel {
		if !errors.Is(applyErr, ErrInventoryRevisionConflict) {
			t.Fatalf("concurrent Apply() error = %v", applyErr)
		}
	}

	final, err := store.Snapshot("Gabriela")
	if err != nil {
		t.Fatal(err)
	}
	if final.Revision != 2 {
		t.Fatalf("final revision = %d, want 2", final.Revision)
	}
	if got := canonicalInventoryStoreLocation(t, final, "stack_material").Slot; got != successfulSlots[0] {
		t.Fatalf("final slot = %d, successful slot = %d", got, successfulSlots[0])
	}
}

func TestCanonicalInventoryStoreIsolatesOwners(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, ownerID := range []string{"Alice", "Bob"} {
		if err := store.Register(canonicalInventoryStoreTestState(t, catalog, ownerID, 1)); err != nil {
			t.Fatal(err)
		}
	}

	if _, _, err := store.Apply("Alice", 1, canonicalInventoryStoreMove(5)); err != nil {
		t.Fatal(err)
	}
	alice, _ := store.Snapshot("Alice")
	bob, _ := store.Snapshot("Bob")
	if alice.Revision != 2 || canonicalInventoryStoreLocation(t, alice, "stack_material").Slot != 5 {
		t.Fatalf("Alice state = %+v", alice)
	}
	if bob.Revision != 1 || canonicalInventoryStoreLocation(t, bob, "stack_material").Slot != 0 {
		t.Fatalf("Bob state changed with Alice: %+v", bob)
	}
	if got := store.OwnerIDs(); !reflect.DeepEqual(got, []string{"Alice", "Bob"}) {
		t.Fatalf("OwnerIDs() = %v", got)
	}
}

func TestCanonicalInventoryStoreConcurrentReadersAndWriter(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var failed atomic.Bool
	for reader := 0; reader < 8; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < 100; iteration++ {
				snapshot, snapshotErr := store.Snapshot("Gabriela")
				if snapshotErr != nil {
					t.Errorf("Snapshot() error = %v", snapshotErr)
					failed.Store(true)
					return
				}
				if _, validationErr := NewInventoryState(catalog, snapshot); validationErr != nil {
					t.Errorf("published snapshot invalid: %v", validationErr)
					failed.Store(true)
					return
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		revision := InventoryRevision(1)
		currentSlot := SlotID(0)
		for iteration := 0; iteration < 50; iteration++ {
			nextSlot := SlotID(1)
			if currentSlot == 1 {
				nextSlot = 0
			}
			_, _, applyErr := store.Apply("Gabriela", revision, canonicalInventoryStoreMove(nextSlot))
			if applyErr != nil {
				t.Errorf("Apply() error = %v", applyErr)
				failed.Store(true)
				return
			}
			revision++
			currentSlot = nextSlot
		}
	}()

	wg.Wait()
	if failed.Load() {
		t.Fatal("concurrent store access failed")
	}
	final, err := store.Snapshot("Gabriela")
	if err != nil {
		t.Fatal(err)
	}
	if final.Revision != 51 {
		t.Fatalf("final revision = %d, want 51", final.Revision)
	}
}

func TestCanonicalInventoryStoreUnregisterAndReregister(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 3)); err != nil {
		t.Fatal(err)
	}

	removed, err := store.Unregister("Gabriela")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}
	if removed.Revision != 3 || removed.OwnerID != "Gabriela" {
		t.Fatalf("removed snapshot = %+v", removed)
	}
	if store.Count() != 0 {
		t.Fatalf("Count() = %d, want 0", store.Count())
	}
	if _, err := store.Snapshot("Gabriela"); !errors.Is(err, ErrInventoryStoreOwnerNotFound) {
		t.Fatalf("Snapshot() after unregister error = %v", err)
	}
	if _, _, err := store.Apply("Gabriela", 3, canonicalInventoryStoreMove(1)); !errors.Is(err, ErrInventoryStoreOwnerNotFound) {
		t.Fatalf("Apply() after unregister error = %v", err)
	}

	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 9)); err != nil {
		t.Fatalf("re-register error = %v", err)
	}
	fresh, err := store.Snapshot("Gabriela")
	if err != nil || fresh.Revision != 9 {
		t.Fatalf("re-registered snapshot = %+v, %v", fresh, err)
	}
}

func TestCanonicalInventoryStoreConcurrentApplyAndUnregister(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)); err != nil {
		t.Fatal(err)
	}

	const workers = 32
	start := make(chan struct{})
	var wg sync.WaitGroup
	var successes atomic.Int32
	successSlot := make(chan SlotID, 1)
	applyErrors := make(chan error, workers)
	unregisterResult := make(chan InventoryStateSnapshot, 1)
	unregisterError := make(chan error, 1)

	for worker := 0; worker < workers; worker++ {
		slot := SlotID(worker + 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _, applyErr := store.Apply("Gabriela", 1, canonicalInventoryStoreMove(slot))
			if applyErr == nil {
				successes.Add(1)
				successSlot <- slot
				return
			}
			applyErrors <- applyErr
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		snapshot, unregisterErr := store.Unregister("Gabriela")
		if unregisterErr != nil {
			unregisterError <- unregisterErr
			return
		}
		unregisterResult <- snapshot
	}()

	close(start)
	wg.Wait()
	close(applyErrors)
	close(unregisterError)
	close(unregisterResult)

	for unregisterErr := range unregisterError {
		t.Fatalf("Unregister() error = %v", unregisterErr)
	}
	removed, ok := <-unregisterResult
	if !ok {
		t.Fatal("missing unregister snapshot")
	}
	if got := successes.Load(); got > 1 {
		t.Fatalf("successful applies = %d, want at most 1", got)
	}
	for applyErr := range applyErrors {
		if !errors.Is(applyErr, ErrInventoryRevisionConflict) &&
			!errors.Is(applyErr, ErrInventoryStoreOwnerUnavailable) &&
			!errors.Is(applyErr, ErrInventoryStoreOwnerNotFound) {
			t.Fatalf("concurrent Apply() error = %v", applyErr)
		}
	}

	if successes.Load() == 0 {
		if removed.Revision != 1 || canonicalInventoryStoreLocation(t, removed, "stack_material").Slot != 0 {
			t.Fatalf("removed snapshot without apply = %+v", removed)
		}
	} else {
		if removed.Revision != 2 {
			t.Fatalf("removed revision = %d, want 2", removed.Revision)
		}
		winner := <-successSlot
		if got := canonicalInventoryStoreLocation(t, removed, "stack_material").Slot; got != winner {
			t.Fatalf("removed slot = %d, winner = %d", got, winner)
		}
	}
	if store.Count() != 0 {
		t.Fatalf("Count() = %d after unregister", store.Count())
	}
}

type unsupportedInventoryStoreOperation struct{}

func (unsupportedInventoryStoreOperation) inventoryOperationKind() InventoryOperationKind {
	return "unsupported_store_test"
}

func (unsupportedInventoryStoreOperation) applyInventoryOperation(
	*inventoryOperationWorkingState,
	*ItemCatalog,
	*InventoryOperationResult,
) error {
	panic("unsupported operation must never execute")
}

func TestCanonicalInventoryStoreRejectsUnsupportedOperationWithoutDeadlock(t *testing.T) {
	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, _, applyErr := store.Apply("Gabriela", 1, unsupportedInventoryStoreOperation{})
		done <- applyErr
	}()

	select {
	case applyErr := <-done:
		if !errors.Is(applyErr, ErrInventoryStoreOperationUnsupported) {
			t.Fatalf("Apply() error = %v", applyErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("unsupported operation deadlocked the store")
	}

	fresh, err := store.Snapshot("Gabriela")
	if err != nil || fresh.Revision != 1 {
		t.Fatalf("snapshot after unsupported operation = %+v, %v", fresh, err)
	}
}

func TestCanonicalInventoryStoreValidation(t *testing.T) {
	if _, err := NewInventoryStore(nil); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("NewInventoryStore(nil) error = %v", err)
	}

	catalog := canonicalInventoryStoreTestCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatal(err)
	}
	var nilStore *InventoryStore
	if err := nilStore.Register(canonicalInventoryStoreTestState(t, catalog, "Gabriela", 1)); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("nil Register() error = %v", err)
	}
	if _, err := nilStore.Snapshot("Gabriela"); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("nil Snapshot() error = %v", err)
	}
	if _, _, err := nilStore.Apply("Gabriela", 1, canonicalInventoryStoreMove(1)); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("nil Apply() error = %v", err)
	}
	if _, err := nilStore.Unregister("Gabriela"); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("nil Unregister() error = %v", err)
	}
	if nilStore.Count() != 0 || nilStore.OwnerIDs() != nil {
		t.Fatal("nil store read helpers returned data")
	}

	if err := store.Register(nil); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("Register(nil) error = %v", err)
	}
	if _, err := store.Snapshot(""); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("Snapshot(empty) error = %v", err)
	}
	if _, err := store.Unregister(" Gabriela"); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("Unregister(spaced) error = %v", err)
	}
	if _, _, err := store.Apply("Gabriela", 1, nil); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("Apply(nil operation) error = %v", err)
	}
	var typedNil *MoveInventoryItemOperation
	if _, _, err := store.Apply("Gabriela", 1, typedNil); !errors.Is(err, ErrInventoryStoreInvalid) {
		t.Fatalf("Apply(typed nil) error = %v", err)
	}
	if _, err := store.Unregister("missing"); !errors.Is(err, ErrInventoryStoreOwnerNotFound) {
		t.Fatalf("Unregister(missing) error = %v", err)
	}
}

func canonicalInventoryStoreLocation(
	t *testing.T,
	snapshot InventoryStateSnapshot,
	instanceID ItemInstanceID,
) InventoryItemLocation {
	t.Helper()
	for _, container := range snapshot.Containers {
		for _, placement := range container.Placements {
			if placement.InstanceID == instanceID {
				return InventoryItemLocation{
					ContainerID: container.ID,
					Slot:        placement.Slot,
				}
			}
		}
	}
	t.Fatalf("instance %q has no location", instanceID)
	return InventoryItemLocation{}
}
