package inventory

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	ErrInventoryStoreInvalid              = errors.New("invalid inventory store request")
	ErrInventoryStoreOwnerNotFound        = errors.New("inventory store owner not found")
	ErrInventoryStoreOwnerAlreadyExists   = errors.New("inventory store owner already registered")
	ErrInventoryStoreOwnerUnavailable     = errors.New("inventory store owner is unavailable")
	ErrInventoryStoreOperationUnsupported = errors.New("inventory store operation is unsupported")
)

type inventoryStoreEntryLifecycle uint32

const (
	inventoryStoreEntryActive inventoryStoreEntryLifecycle = iota + 1
	inventoryStoreEntryRemoving
	inventoryStoreEntryRemoved
)

type inventoryStoreEntry struct {
	applyMu   sync.Mutex
	state     atomic.Pointer[InventoryState]
	lifecycle atomic.Uint32
}

// InventoryStore is an in-memory publication boundary for canonical inventory
// states. Registry access is brief and global, while mutations are serialized
// independently per owner.
type InventoryStore struct {
	catalog *ItemCatalog

	mu      sync.RWMutex
	entries map[string]*inventoryStoreEntry
}

func NewInventoryStore(catalog *ItemCatalog) (*InventoryStore, error) {
	if catalog == nil {
		return nil, fmt.Errorf("%w: item catalog is required", ErrInventoryStoreInvalid)
	}
	return &InventoryStore{
		catalog: catalog,
		entries: make(map[string]*inventoryStoreEntry),
	}, nil
}

// Register defensively validates and publishes an inventory state. The caller's
// state is never retained directly.
func (store *InventoryStore) Register(state *InventoryState) error {
	if err := store.validate(); err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("%w: inventory state is required", ErrInventoryStoreInvalid)
	}

	validated, err := NewInventoryState(store.catalog, state.Snapshot())
	if err != nil {
		return fmt.Errorf("%w: inventory state validation failed: %v", ErrInventoryStoreInvalid, err)
	}
	ownerID := validated.OwnerID()

	entry := &inventoryStoreEntry{}
	entry.state.Store(validated)
	entry.lifecycle.Store(uint32(inventoryStoreEntryActive))

	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.entries[ownerID]; exists {
		return fmt.Errorf("%w: %q", ErrInventoryStoreOwnerAlreadyExists, ownerID)
	}
	store.entries[ownerID] = entry
	return nil
}

// Unregister prevents new operations, waits for an already-running operation,
// and returns the last published defensive snapshot.
func (store *InventoryStore) Unregister(ownerID string) (InventoryStateSnapshot, error) {
	if err := store.validate(); err != nil {
		return InventoryStateSnapshot{}, err
	}
	if err := validateInventoryStoreOwnerID(ownerID); err != nil {
		return InventoryStateSnapshot{}, err
	}

	store.mu.RLock()
	entry, exists := store.entries[ownerID]
	store.mu.RUnlock()
	if !exists {
		return InventoryStateSnapshot{}, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerNotFound, ownerID)
	}
	if !entry.lifecycle.CompareAndSwap(
		uint32(inventoryStoreEntryActive),
		uint32(inventoryStoreEntryRemoving),
	) {
		return InventoryStateSnapshot{}, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerUnavailable, ownerID)
	}

	entry.applyMu.Lock()
	defer entry.applyMu.Unlock()

	current := entry.state.Load()
	if current == nil {
		entry.lifecycle.Store(uint32(inventoryStoreEntryRemoved))
		store.deleteEntryIfSame(ownerID, entry)
		return InventoryStateSnapshot{}, fmt.Errorf("%w: owner %q has no published state", ErrInventoryStoreInvalid, ownerID)
	}
	snapshot := current.Snapshot()

	store.mu.Lock()
	registered, stillRegistered := store.entries[ownerID]
	if !stillRegistered || registered != entry {
		store.mu.Unlock()
		entry.lifecycle.Store(uint32(inventoryStoreEntryRemoved))
		return InventoryStateSnapshot{}, fmt.Errorf("%w: owner %q registry changed during unregister", ErrInventoryStoreInvalid, ownerID)
	}
	delete(store.entries, ownerID)
	store.mu.Unlock()

	entry.lifecycle.Store(uint32(inventoryStoreEntryRemoved))
	return snapshot, nil
}

// Snapshot returns a defensive copy of the latest published state.
func (store *InventoryStore) Snapshot(ownerID string) (InventoryStateSnapshot, error) {
	entry, err := store.lookupActiveEntry(ownerID)
	if err != nil {
		return InventoryStateSnapshot{}, err
	}

	current := entry.state.Load()
	if current == nil {
		return InventoryStateSnapshot{}, fmt.Errorf("%w: owner %q has no published state", ErrInventoryStoreInvalid, ownerID)
	}
	if inventoryStoreEntryLifecycle(entry.lifecycle.Load()) != inventoryStoreEntryActive {
		return InventoryStateSnapshot{}, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerUnavailable, ownerID)
	}
	return current.Snapshot(), nil
}

// Apply serializes mutations for one owner, applies the canonical pure
// operation, and publishes only a fully validated successful result.
func (store *InventoryStore) Apply(
	ownerID string,
	expectedRevision InventoryRevision,
	operation InventoryOperation,
) (InventoryStateSnapshot, InventoryOperationResult, error) {
	if err := store.validate(); err != nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, err
	}
	if err := validateInventoryStoreOwnerID(ownerID); err != nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, err
	}

	canonicalOperation, err := cloneCanonicalInventoryStoreOperation(operation)
	if err != nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, err
	}
	entry, err := store.lookupActiveEntry(ownerID)
	if err != nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, err
	}

	entry.applyMu.Lock()
	defer entry.applyMu.Unlock()
	if inventoryStoreEntryLifecycle(entry.lifecycle.Load()) != inventoryStoreEntryActive {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerUnavailable, ownerID)
	}

	current := entry.state.Load()
	if current == nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, fmt.Errorf("%w: owner %q has no published state", ErrInventoryStoreInvalid, ownerID)
	}
	if current.OwnerID() != ownerID {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, fmt.Errorf(
			"%w: registry owner %q contains state for %q",
			ErrInventoryStoreInvalid,
			ownerID,
			current.OwnerID(),
		)
	}

	next, result, err := ApplyInventoryOperation(
		store.catalog,
		current,
		expectedRevision,
		canonicalOperation,
	)
	if err != nil {
		return InventoryStateSnapshot{}, InventoryOperationResult{}, err
	}

	entry.state.Store(next)
	return next.Snapshot(), result, nil
}

func (store *InventoryStore) Count() int {
	if store == nil {
		return 0
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return len(store.entries)
}

func (store *InventoryStore) OwnerIDs() []string {
	if store == nil {
		return nil
	}
	store.mu.RLock()
	owners := make([]string, 0, len(store.entries))
	for ownerID := range store.entries {
		owners = append(owners, ownerID)
	}
	store.mu.RUnlock()
	sort.Strings(owners)
	return owners
}

func (store *InventoryStore) validate() error {
	if store == nil {
		return fmt.Errorf("%w: inventory store is required", ErrInventoryStoreInvalid)
	}
	if store.catalog == nil {
		return fmt.Errorf("%w: item catalog is required", ErrInventoryStoreInvalid)
	}
	if store.entries == nil {
		return fmt.Errorf("%w: inventory store registry is not initialized", ErrInventoryStoreInvalid)
	}
	return nil
}

func (store *InventoryStore) lookupActiveEntry(ownerID string) (*inventoryStoreEntry, error) {
	if err := store.validate(); err != nil {
		return nil, err
	}
	if err := validateInventoryStoreOwnerID(ownerID); err != nil {
		return nil, err
	}

	store.mu.RLock()
	entry, exists := store.entries[ownerID]
	store.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerNotFound, ownerID)
	}
	if inventoryStoreEntryLifecycle(entry.lifecycle.Load()) != inventoryStoreEntryActive {
		return nil, fmt.Errorf("%w: %q", ErrInventoryStoreOwnerUnavailable, ownerID)
	}
	return entry, nil
}

func (store *InventoryStore) deleteEntryIfSame(ownerID string, entry *inventoryStoreEntry) {
	store.mu.Lock()
	if current, exists := store.entries[ownerID]; exists && current == entry {
		delete(store.entries, ownerID)
	}
	store.mu.Unlock()
}

func validateInventoryStoreOwnerID(ownerID string) error {
	if ownerID == "" {
		return fmt.Errorf("%w: owner ID is required", ErrInventoryStoreInvalid)
	}
	if strings.TrimSpace(ownerID) != ownerID {
		return fmt.Errorf("%w: owner ID %q must not contain surrounding whitespace", ErrInventoryStoreInvalid, ownerID)
	}
	return nil
}

// cloneCanonicalInventoryStoreOperation both seals the store boundary to the
// six canonical operations and snapshots pointer arguments before locking.
func cloneCanonicalInventoryStoreOperation(operation InventoryOperation) (InventoryOperation, error) {
	if operation == nil || inventoryOperationIsNil(operation) {
		return nil, fmt.Errorf("%w: operation is required", ErrInventoryStoreInvalid)
	}

	switch typed := operation.(type) {
	case MoveInventoryItemOperation:
		return typed, nil
	case *MoveInventoryItemOperation:
		return *typed, nil
	case SwapInventoryItemsOperation:
		return typed, nil
	case *SwapInventoryItemsOperation:
		return *typed, nil
	case MergeInventoryStacksOperation:
		return typed, nil
	case *MergeInventoryStacksOperation:
		return *typed, nil
	case SplitInventoryStackOperation:
		return typed, nil
	case *SplitInventoryStackOperation:
		return *typed, nil
	case AddInventoryItemOperation:
		return typed, nil
	case *AddInventoryItemOperation:
		return *typed, nil
	case RemoveInventoryItemOperation:
		return typed, nil
	case *RemoveInventoryItemOperation:
		return *typed, nil
	default:
		return nil, fmt.Errorf(
			"%w: %T",
			ErrInventoryStoreOperationUnsupported,
			operation,
		)
	}
}
