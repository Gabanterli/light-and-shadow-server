package inventory

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
)

type itemCreationEntropyReader struct {
	mu    sync.Mutex
	next  byte
	reads int
	err   error
}

func (reader *itemCreationEntropyReader) Read(destination []byte) (int, error) {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	reader.reads++
	if reader.err != nil {
		return 0, reader.err
	}
	for index := range destination {
		destination[index] = reader.next + byte(index)
	}
	reader.next++
	return len(destination), nil
}

func (reader *itemCreationEntropyReader) ReadCount() int {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	return reader.reads
}

func newItemCreationCatalog(t *testing.T) *ItemCatalog {
	t.Helper()
	catalog, err := NewItemCatalog([]ItemDefinition{
		{
			ID:        "potion_heal",
			Name:      "Potion",
			Type:      ItemTypeConsumable,
			Stackable: true,
			MaxStack:  100,
		},
		{
			ID:        "sword_basic",
			Name:      "Sword",
			Type:      ItemTypeWeapon,
			Category:  "sword",
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

func newItemCreationState(
	t *testing.T,
	catalog *ItemCatalog,
	ownerID string,
	capacity uint32,
	items []InventoryItemStack,
	placements []InventorySlotPlacement,
) *InventoryState {
	t.Helper()
	state, err := NewInventoryState(catalog, InventoryStateSnapshot{
		OwnerID:  ownerID,
		Revision: 1,
		Containers: []InventoryContainerSnapshot{
			{
				ID:         ContainerID(ownerID + ":backpack"),
				Kind:       ContainerKindBackpack,
				Capacity:   capacity,
				Placements: placements,
			},
			{
				ID:       ContainerID(ownerID + ":equipment"),
				Kind:     ContainerKindEquipment,
				Capacity: 3,
			},
		},
		Items: items,
	})
	if err != nil {
		t.Fatalf("NewInventoryState() error = %v", err)
	}
	return state
}

func newItemCreationService(
	t *testing.T,
	catalog *ItemCatalog,
	reader io.Reader,
	states ...*InventoryState,
) (*AuthoritativeItemCreationService, *InventoryStore, *ItemInstanceIDAuthority) {
	t.Helper()
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatalf("NewInventoryStore() error = %v", err)
	}
	for _, state := range states {
		if err := store.Register(state); err != nil {
			t.Fatalf("Register(%q) error = %v", state.OwnerID(), err)
		}
	}
	authority, err := NewItemInstanceIDAuthorityWithEntropy(reader, 8)
	if err != nil {
		t.Fatalf("NewItemInstanceIDAuthorityWithEntropy() error = %v", err)
	}
	service, err := NewAuthoritativeItemCreationService(store, authority)
	if err != nil {
		t.Fatalf("NewAuthoritativeItemCreationService() error = %v", err)
	}
	return service, store, authority
}

func TestAuthoritativeItemCreationServiceCreateItem(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:create"
	state := newItemCreationState(t, catalog, ownerID, 8, nil, nil)
	entropy := bytes.NewReader([]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	})
	service, store, authority := newItemCreationService(t, catalog, entropy, state)

	result, err := service.CreateItem(CreateInventoryItemRequest{
		OwnerID:          ownerID,
		ExpectedRevision: 1,
		DefinitionID:     "potion_heal",
		Quantity:         5,
		To: InventoryItemLocation{
			ContainerID: ContainerID(ownerID + ":backpack"),
			Slot:        0,
		},
	})
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}

	expectedID := ItemInstanceID("item:000102030405060708090a0b0c0d0e0f")
	if result.InstanceID != expectedID {
		t.Fatalf("InstanceID = %q, want %q", result.InstanceID, expectedID)
	}
	if result.Operation.Kind != InventoryOperationAdd || result.Operation.NewRevision != 2 {
		t.Fatalf("operation = %+v", result.Operation)
	}
	if !itemCreationResultContainsID(result.Operation.CreatedInstanceIDs, expectedID) {
		t.Fatalf("created IDs = %v, want %q", result.Operation.CreatedInstanceIDs, expectedID)
	}
	if result.Snapshot.Revision != 2 || len(result.Snapshot.Items) != 1 {
		t.Fatalf("snapshot = %+v", result.Snapshot)
	}
	item := result.Snapshot.Items[0]
	if item.InstanceID != expectedID || item.DefinitionID != "potion_heal" || item.Quantity != 5 {
		t.Fatalf("item = %+v", item)
	}
	reserved, err := authority.IsReserved(expectedID)
	if err != nil || !reserved || authority.Count() != 1 {
		t.Fatalf("reservation: reserved=%v count=%d err=%v", reserved, authority.Count(), err)
	}
	latest, err := store.Snapshot(ownerID)
	if err != nil || latest.Revision != 2 || len(latest.Items) != 1 {
		t.Fatalf("store snapshot = %+v, err=%v", latest, err)
	}
}

func TestAuthoritativeItemCreationServiceSplitStack(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:split"
	sourceID := ItemInstanceID("item:11111111111111111111111111111111")
	state := newItemCreationState(
		t,
		catalog,
		ownerID,
		8,
		[]InventoryItemStack{{InstanceID: sourceID, DefinitionID: "potion_heal", Quantity: 10}},
		[]InventorySlotPlacement{{Slot: 0, InstanceID: sourceID}},
	)
	entropy := bytes.NewReader(bytes.Repeat([]byte{0x22}, ItemInstanceIDEntropyBytes))
	service, _, authority := newItemCreationService(t, catalog, entropy, state)

	result, err := service.SplitStack(SplitInventoryStackRequest{
		OwnerID:          ownerID,
		ExpectedRevision: 1,
		SourceInstanceID: sourceID,
		Quantity:         3,
		To: InventoryItemLocation{
			ContainerID: ContainerID(ownerID + ":backpack"),
			Slot:        1,
		},
	})
	if err != nil {
		t.Fatalf("SplitStack() error = %v", err)
	}

	newID := ItemInstanceID("item:22222222222222222222222222222222")
	if result.InstanceID != newID || result.Operation.Kind != InventoryOperationSplit {
		t.Fatalf("result = %+v", result)
	}
	quantities := map[ItemInstanceID]int{}
	for _, item := range result.Snapshot.Items {
		quantities[item.InstanceID] = item.Quantity
	}
	if quantities[sourceID] != 7 || quantities[newID] != 3 {
		t.Fatalf("quantities = %v", quantities)
	}
	if authority.Count() != 1 {
		t.Fatalf("authority.Count() = %d, want 1", authority.Count())
	}
}

func TestAuthoritativeItemCreationServiceReleasesReservationOnApplyFailure(t *testing.T) {
	t.Run("create occupied target", func(t *testing.T) {
		catalog := newItemCreationCatalog(t)
		ownerID := "player:occupied"
		existingID := ItemInstanceID("item:33333333333333333333333333333333")
		state := newItemCreationState(
			t,
			catalog,
			ownerID,
			4,
			[]InventoryItemStack{{InstanceID: existingID, DefinitionID: "sword_basic", Quantity: 1}},
			[]InventorySlotPlacement{{Slot: 0, InstanceID: existingID}},
		)
		reader := &itemCreationEntropyReader{next: 0x40}
		service, store, authority := newItemCreationService(t, catalog, reader, state)

		_, err := service.CreateItem(CreateInventoryItemRequest{
			OwnerID:          ownerID,
			ExpectedRevision: 1,
			DefinitionID:     "potion_heal",
			Quantity:         1,
			To: InventoryItemLocation{
				ContainerID: ContainerID(ownerID + ":backpack"),
				Slot:        0,
			},
		})
		if !errors.Is(err, ErrInventorySlotOccupied) {
			t.Fatalf("CreateItem() error = %v, want ErrInventorySlotOccupied", err)
		}
		if authority.Count() != 0 {
			t.Fatalf("authority.Count() = %d, want 0", authority.Count())
		}
		snapshot, snapshotErr := store.Snapshot(ownerID)
		if snapshotErr != nil || snapshot.Revision != 1 || len(snapshot.Items) != 1 {
			t.Fatalf("snapshot = %+v, err=%v", snapshot, snapshotErr)
		}
	})

	t.Run("invalid split", func(t *testing.T) {
		catalog := newItemCreationCatalog(t)
		ownerID := "player:split-fail"
		sourceID := ItemInstanceID("item:44444444444444444444444444444444")
		state := newItemCreationState(
			t,
			catalog,
			ownerID,
			4,
			[]InventoryItemStack{{InstanceID: sourceID, DefinitionID: "potion_heal", Quantity: 5}},
			[]InventorySlotPlacement{{Slot: 0, InstanceID: sourceID}},
		)
		reader := &itemCreationEntropyReader{next: 0x50}
		service, store, authority := newItemCreationService(t, catalog, reader, state)

		_, err := service.SplitStack(SplitInventoryStackRequest{
			OwnerID:          ownerID,
			ExpectedRevision: 1,
			SourceInstanceID: sourceID,
			Quantity:         5,
			To: InventoryItemLocation{
				ContainerID: ContainerID(ownerID + ":backpack"),
				Slot:        1,
			},
		})
		if !errors.Is(err, ErrInventoryStackRule) {
			t.Fatalf("SplitStack() error = %v, want ErrInventoryStackRule", err)
		}
		if authority.Count() != 0 {
			t.Fatalf("authority.Count() = %d, want 0", authority.Count())
		}
		snapshot, snapshotErr := store.Snapshot(ownerID)
		if snapshotErr != nil || snapshot.Revision != 1 || len(snapshot.Items) != 1 {
			t.Fatalf("snapshot = %+v, err=%v", snapshot, snapshotErr)
		}
	})
}

func TestAuthoritativeItemCreationServiceStaleRevisionDoesNotReserve(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:stale"
	state := newItemCreationState(t, catalog, ownerID, 4, nil, nil)
	reader := &itemCreationEntropyReader{next: 0x60}
	service, _, authority := newItemCreationService(t, catalog, reader, state)

	_, err := service.CreateItem(CreateInventoryItemRequest{
		OwnerID:          ownerID,
		ExpectedRevision: 2,
		DefinitionID:     "potion_heal",
		Quantity:         1,
		To: InventoryItemLocation{
			ContainerID: ContainerID(ownerID + ":backpack"),
			Slot:        0,
		},
	})
	if !errors.Is(err, ErrInventoryRevisionConflict) {
		t.Fatalf("CreateItem() error = %v, want ErrInventoryRevisionConflict", err)
	}
	if reader.ReadCount() != 0 || authority.Count() != 0 {
		t.Fatalf("reads=%d reservations=%d, want zero", reader.ReadCount(), authority.Count())
	}
}

func TestAuthoritativeItemCreationServiceEntropyFailureIsAtomic(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:entropy"
	state := newItemCreationState(t, catalog, ownerID, 4, nil, nil)
	reader := &itemCreationEntropyReader{err: fmt.Errorf("entropy unavailable")}
	service, store, authority := newItemCreationService(t, catalog, reader, state)

	_, err := service.CreateItem(CreateInventoryItemRequest{
		OwnerID:          ownerID,
		ExpectedRevision: 1,
		DefinitionID:     "potion_heal",
		Quantity:         1,
		To: InventoryItemLocation{
			ContainerID: ContainerID(ownerID + ":backpack"),
			Slot:        0,
		},
	})
	if !errors.Is(err, ErrItemInstanceIDEntropy) {
		t.Fatalf("CreateItem() error = %v, want ErrItemInstanceIDEntropy", err)
	}
	if authority.Count() != 0 {
		t.Fatalf("authority.Count() = %d, want 0", authority.Count())
	}
	snapshot, snapshotErr := store.Snapshot(ownerID)
	if snapshotErr != nil || snapshot.Revision != 1 || len(snapshot.Items) != 0 {
		t.Fatalf("snapshot = %+v, err=%v", snapshot, snapshotErr)
	}
}

func TestAuthoritativeItemCreationServiceConcurrentSameOwnerHasNoReservationLeaks(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:concurrent"
	state := newItemCreationState(t, catalog, ownerID, 64, nil, nil)
	reader := &itemCreationEntropyReader{next: 0x70}
	service, store, authority := newItemCreationService(t, catalog, reader, state)

	const workers = 32
	start := make(chan struct{})
	errorsByWorker := make(chan error, workers)
	results := make(chan AuthoritativeItemCreationResult, workers)
	var wait sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		worker := worker
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			result, err := service.CreateItem(CreateInventoryItemRequest{
				OwnerID:          ownerID,
				ExpectedRevision: 1,
				DefinitionID:     "potion_heal",
				Quantity:         1,
				To: InventoryItemLocation{
					ContainerID: ContainerID(ownerID + ":backpack"),
					Slot:        SlotID(worker),
				},
			})
			if err != nil {
				errorsByWorker <- err
				return
			}
			results <- result
		}()
	}
	close(start)
	wait.Wait()
	close(errorsByWorker)
	close(results)

	successes := 0
	for range results {
		successes++
	}
	if successes != 1 {
		t.Fatalf("successes = %d, want 1", successes)
	}
	for err := range errorsByWorker {
		if !errors.Is(err, ErrInventoryRevisionConflict) {
			t.Fatalf("concurrent error = %v, want ErrInventoryRevisionConflict", err)
		}
	}
	if authority.Count() != 1 {
		t.Fatalf("authority.Count() = %d, want 1", authority.Count())
	}
	snapshot, err := store.Snapshot(ownerID)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snapshot.Revision != 2 || len(snapshot.Items) != 1 {
		t.Fatalf("snapshot revision=%d items=%d, want 2 and 1", snapshot.Revision, len(snapshot.Items))
	}
}

func TestAuthoritativeItemCreationServiceConcurrentOwnersAreIsolated(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	stateA := newItemCreationState(t, catalog, "player:a", 4, nil, nil)
	stateB := newItemCreationState(t, catalog, "player:b", 4, nil, nil)
	reader := &itemCreationEntropyReader{next: 0x90}
	service, store, authority := newItemCreationService(t, catalog, reader, stateA, stateB)

	owners := []string{"player:a", "player:b"}
	results := make(chan AuthoritativeItemCreationResult, len(owners))
	errorsChannel := make(chan error, len(owners))
	var wait sync.WaitGroup
	for _, ownerID := range owners {
		ownerID := ownerID
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := service.CreateItem(CreateInventoryItemRequest{
				OwnerID:          ownerID,
				ExpectedRevision: 1,
				DefinitionID:     "sword_basic",
				Quantity:         1,
				To: InventoryItemLocation{
					ContainerID: ContainerID(ownerID + ":backpack"),
					Slot:        0,
				},
			})
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- result
		}()
	}
	wait.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("CreateItem() concurrent owner error = %v", err)
	}
	seen := map[ItemInstanceID]struct{}{}
	for result := range results {
		seen[result.InstanceID] = struct{}{}
	}
	if len(seen) != 2 || authority.Count() != 2 {
		t.Fatalf("unique IDs=%d reservations=%d, want 2 and 2", len(seen), authority.Count())
	}
	for _, ownerID := range owners {
		snapshot, err := store.Snapshot(ownerID)
		if err != nil || snapshot.Revision != 2 || len(snapshot.Items) != 1 {
			t.Fatalf("owner %q snapshot=%+v err=%v", ownerID, snapshot, err)
		}
	}
}

func TestAuthoritativeItemCreationServiceRetryWithOldRevisionDoesNotDuplicate(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	ownerID := "player:retry"
	state := newItemCreationState(t, catalog, ownerID, 4, nil, nil)
	reader := &itemCreationEntropyReader{next: 0xa0}
	service, store, authority := newItemCreationService(t, catalog, reader, state)
	request := CreateInventoryItemRequest{
		OwnerID:          ownerID,
		ExpectedRevision: 1,
		DefinitionID:     "potion_heal",
		Quantity:         1,
		To: InventoryItemLocation{
			ContainerID: ContainerID(ownerID + ":backpack"),
			Slot:        0,
		},
	}

	if _, err := service.CreateItem(request); err != nil {
		t.Fatalf("first CreateItem() error = %v", err)
	}
	readsAfterSuccess := reader.ReadCount()
	_, err := service.CreateItem(request)
	if !errors.Is(err, ErrInventoryRevisionConflict) {
		t.Fatalf("retry error = %v, want ErrInventoryRevisionConflict", err)
	}
	if reader.ReadCount() != readsAfterSuccess {
		t.Fatalf("retry consumed entropy: before=%d after=%d", readsAfterSuccess, reader.ReadCount())
	}
	if authority.Count() != 1 {
		t.Fatalf("authority.Count() = %d, want 1", authority.Count())
	}
	snapshot, snapshotErr := store.Snapshot(ownerID)
	if snapshotErr != nil || snapshot.Revision != 2 || len(snapshot.Items) != 1 {
		t.Fatalf("snapshot = %+v, err=%v", snapshot, snapshotErr)
	}
}

func TestAuthoritativeItemCreationServiceValidation(t *testing.T) {
	catalog := newItemCreationCatalog(t)
	store, err := NewInventoryStore(catalog)
	if err != nil {
		t.Fatalf("NewInventoryStore() error = %v", err)
	}
	authority := NewItemInstanceIDAuthority()

	if _, err := NewAuthoritativeItemCreationService(nil, authority); !errors.Is(err, ErrItemCreationServiceInvalid) {
		t.Fatalf("nil store error = %v", err)
	}
	if _, err := NewAuthoritativeItemCreationService(store, nil); !errors.Is(err, ErrItemCreationServiceInvalid) {
		t.Fatalf("nil authority error = %v", err)
	}

	var nilService *AuthoritativeItemCreationService
	if _, err := nilService.CreateItem(CreateInventoryItemRequest{}); !errors.Is(err, ErrItemCreationServiceInvalid) {
		t.Fatalf("nil service error = %v", err)
	}

	ownerID := "player:validation"
	state := newItemCreationState(t, catalog, ownerID, 4, nil, nil)
	reader := &itemCreationEntropyReader{next: 0xb0}
	service, _, reservations := newItemCreationService(t, catalog, reader, state)

	tests := []struct {
		name    string
		request CreateInventoryItemRequest
		want    error
	}{
		{
			name: "invalid owner",
			request: CreateInventoryItemRequest{
				OwnerID:          " player:validation",
				ExpectedRevision: 1,
				DefinitionID:     "potion_heal",
				Quantity:         1,
				To:               InventoryItemLocation{ContainerID: ContainerID(ownerID + ":backpack")},
			},
			want: ErrInventoryStoreInvalid,
		},
		{
			name: "zero revision",
			request: CreateInventoryItemRequest{
				OwnerID:      ownerID,
				DefinitionID: "potion_heal",
				Quantity:     1,
				To:           InventoryItemLocation{ContainerID: ContainerID(ownerID + ":backpack")},
			},
			want: ErrItemCreationServiceInvalid,
		},
		{
			name: "unknown definition",
			request: CreateInventoryItemRequest{
				OwnerID:          ownerID,
				ExpectedRevision: 1,
				DefinitionID:     "unknown",
				Quantity:         1,
				To:               InventoryItemLocation{ContainerID: ContainerID(ownerID + ":backpack")},
			},
			want: ErrInventoryOperationInvalid,
		},
		{
			name: "invalid quantity",
			request: CreateInventoryItemRequest{
				OwnerID:          ownerID,
				ExpectedRevision: 1,
				DefinitionID:     "potion_heal",
				Quantity:         0,
				To:               InventoryItemLocation{ContainerID: ContainerID(ownerID + ":backpack")},
			},
			want: ErrInventoryStackRule,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.CreateItem(test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("CreateItem() error = %v, want %v", err, test.want)
			}
		})
	}
	if reader.ReadCount() != 0 || reservations.Count() != 0 {
		t.Fatalf("validation consumed entropy or reserved IDs: reads=%d reservations=%d", reader.ReadCount(), reservations.Count())
	}
}
