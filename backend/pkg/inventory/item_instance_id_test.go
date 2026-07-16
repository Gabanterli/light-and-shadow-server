package inventory

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"sync"
	"testing"
)

type itemInstanceIDErrorReader struct {
	err error
}

func (reader itemInstanceIDErrorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func itemInstanceIDBytes(seed byte) []byte {
	data := make([]byte, ItemInstanceIDEntropyBytes)
	for index := range data {
		data[index] = seed + byte(index)
	}
	return data
}

func itemInstanceIDFromSeed(t *testing.T, seed byte) ItemInstanceID {
	t.Helper()
	id, err := GenerateItemInstanceID(bytes.NewReader(itemInstanceIDBytes(seed)))
	if err != nil {
		t.Fatalf("GenerateItemInstanceID() error = %v", err)
	}
	return id
}

func TestAuthoritativeItemInstanceIDGenerateParseAndValidate(t *testing.T) {
	id, err := GenerateItemInstanceID(bytes.NewReader(itemInstanceIDBytes(0)))
	if err != nil {
		t.Fatalf("GenerateItemInstanceID() error = %v", err)
	}
	const expected ItemInstanceID = "item:000102030405060708090a0b0c0d0e0f"
	if id != expected {
		t.Fatalf("id = %q, want %q", id, expected)
	}
	if err := ValidateItemInstanceID(id); err != nil {
		t.Fatalf("ValidateItemInstanceID() error = %v", err)
	}
	parsed, err := ParseItemInstanceID(string(id))
	if err != nil {
		t.Fatalf("ParseItemInstanceID() error = %v", err)
	}
	if parsed != id {
		t.Fatalf("parsed = %q, want %q", parsed, id)
	}

	invalid := []string{
		"",
		"item:",
		"item:000102030405060708090a0b0c0d0e0",
		"item:000102030405060708090a0b0c0d0e0f00",
		"item:000102030405060708090A0B0C0D0E0F",
		"item:000102030405060708090a0b0c0d0e0g",
		"item_000102030405060708090a0b0c0d0e0f",
		" item:000102030405060708090a0b0c0d0e0f",
	}
	for _, value := range invalid {
		t.Run(fmt.Sprintf("invalid_%q", value), func(t *testing.T) {
			if _, err := ParseItemInstanceID(value); !errors.Is(err, ErrItemInstanceIDInvalid) {
				t.Fatalf("ParseItemInstanceID(%q) error = %v, want ErrItemInstanceIDInvalid", value, err)
			}
		})
	}
}

func TestAuthoritativeItemInstanceIDEntropyFailures(t *testing.T) {
	if _, err := GenerateItemInstanceID(nil); !errors.Is(err, ErrItemInstanceIDEntropy) {
		t.Fatalf("nil entropy error = %v", err)
	}

	sentinel := errors.New("entropy unavailable")
	if _, err := GenerateItemInstanceID(itemInstanceIDErrorReader{err: sentinel}); !errors.Is(err, ErrItemInstanceIDEntropy) {
		t.Fatalf("reader error = %v", err)
	}

	if _, err := GenerateItemInstanceID(bytes.NewReader([]byte{1, 2, 3})); !errors.Is(err, ErrItemInstanceIDEntropy) {
		t.Fatalf("short reader error = %v", err)
	}
}

func TestAuthoritativeItemInstanceIDCollisionRetry(t *testing.T) {
	firstEntropy := itemInstanceIDBytes(1)
	secondEntropy := itemInstanceIDBytes(2)
	authority, err := NewItemInstanceIDAuthorityWithEntropy(
		bytes.NewReader(append(append([]byte(nil), firstEntropy...), secondEntropy...)),
		2,
	)
	if err != nil {
		t.Fatal(err)
	}
	firstID := itemInstanceIDFromSeed(t, 1)
	secondID := itemInstanceIDFromSeed(t, 2)
	if err := authority.ReserveExisting(firstID); err != nil {
		t.Fatal(err)
	}

	got, err := authority.ReserveNew()
	if err != nil {
		t.Fatalf("ReserveNew() error = %v", err)
	}
	if got != secondID {
		t.Fatalf("ReserveNew() = %q, want %q", got, secondID)
	}
	if authority.Count() != 2 {
		t.Fatalf("Count() = %d, want 2", authority.Count())
	}
}

func TestAuthoritativeItemInstanceIDCollisionExhaustionIsAtomic(t *testing.T) {
	entropy := itemInstanceIDBytes(3)
	authority, err := NewItemInstanceIDAuthorityWithEntropy(
		bytes.NewReader(bytes.Repeat(entropy, 3)),
		3,
	)
	if err != nil {
		t.Fatal(err)
	}
	id := itemInstanceIDFromSeed(t, 3)
	if err := authority.ReserveExisting(id); err != nil {
		t.Fatal(err)
	}

	if _, err := authority.ReserveNew(); !errors.Is(err, ErrItemInstanceIDCollision) {
		t.Fatalf("ReserveNew() error = %v, want collision", err)
	}
	if authority.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", authority.Count())
	}
	if got := authority.ReservedIDs(); !reflect.DeepEqual(got, []ItemInstanceID{id}) {
		t.Fatalf("ReservedIDs() = %v", got)
	}
}

func TestAuthoritativeItemInstanceIDReserveReleaseAndDefensiveSnapshot(t *testing.T) {
	authority := NewItemInstanceIDAuthority()
	first := itemInstanceIDFromSeed(t, 9)
	second := itemInstanceIDFromSeed(t, 8)
	if err := authority.ReserveExisting(first); err != nil {
		t.Fatal(err)
	}
	if err := authority.ReserveExisting(second); err != nil {
		t.Fatal(err)
	}
	if err := authority.ReserveExisting(first); !errors.Is(err, ErrItemInstanceIDAlreadyReserved) {
		t.Fatalf("duplicate reserve error = %v", err)
	}

	ids := authority.ReservedIDs()
	if !sort.SliceIsSorted(ids, func(i, j int) bool { return ids[i] < ids[j] }) {
		t.Fatalf("ReservedIDs() not sorted: %v", ids)
	}
	ids[0] = itemInstanceIDFromSeed(t, 7)
	if authority.Count() != 2 {
		t.Fatalf("snapshot mutation changed authority count")
	}

	reserved, err := authority.IsReserved(first)
	if err != nil || !reserved {
		t.Fatalf("IsReserved(first) = %v, %v", reserved, err)
	}
	if err := authority.Release(first); err != nil {
		t.Fatalf("Release(first) error = %v", err)
	}
	if err := authority.Release(first); !errors.Is(err, ErrItemInstanceIDNotReserved) {
		t.Fatalf("double release error = %v", err)
	}
	reserved, err = authority.IsReserved(first)
	if err != nil || reserved {
		t.Fatalf("IsReserved(first) after release = %v, %v", reserved, err)
	}
}

func TestAuthoritativeItemInstanceIDConcurrentUniqueness(t *testing.T) {
	authority := NewItemInstanceIDAuthority()
	const workers = 64
	const perWorker = 16

	ids := make(chan ItemInstanceID, workers*perWorker)
	errorsChannel := make(chan error, workers)
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < perWorker; iteration++ {
				id, err := authority.ReserveNew()
				if err != nil {
					errorsChannel <- err
					return
				}
				ids <- id
			}
		}()
	}
	wait.Wait()
	close(ids)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("ReserveNew() concurrent error = %v", err)
	}
	seen := make(map[ItemInstanceID]struct{}, workers*perWorker)
	for id := range ids {
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate generated ID %q", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != workers*perWorker {
		t.Fatalf("unique count = %d, want %d", len(seen), workers*perWorker)
	}
	if authority.Count() != workers*perWorker {
		t.Fatalf("authority Count() = %d", authority.Count())
	}
}

func TestAuthoritativeItemInstanceIDConcurrentReserveExisting(t *testing.T) {
	authority := NewItemInstanceIDAuthority()
	id := itemInstanceIDFromSeed(t, 20)
	const workers = 64

	var wait sync.WaitGroup
	results := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- authority.ReserveExisting(id)
		}()
	}
	wait.Wait()
	close(results)

	successes := 0
	duplicates := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrItemInstanceIDAlreadyReserved):
			duplicates++
		default:
			t.Fatalf("unexpected error = %v", err)
		}
	}
	if successes != 1 || duplicates != workers-1 {
		t.Fatalf("successes=%d duplicates=%d", successes, duplicates)
	}
}

func TestAuthoritativeItemInstanceIDCanonicalStateCompatibility(t *testing.T) {
	catalog, err := NewItemCatalog([]ItemDefinition{{
		ID:        "material_test",
		Name:      "Material Test",
		Type:      ItemTypeMaterial,
		Tier:      0,
		Stackable: true,
		MaxStack:  99,
	}})
	if err != nil {
		t.Fatal(err)
	}
	authority := NewItemInstanceIDAuthority()
	id, err := authority.ReserveNew()
	if err != nil {
		t.Fatal(err)
	}

	state, err := NewInventoryState(catalog, InventoryStateSnapshot{
		OwnerID:  "player:test",
		Revision: 1,
		Containers: []InventoryContainerSnapshot{
			{
				ID:       "backpack:main",
				Kind:     ContainerKindBackpack,
				Capacity: 10,
				Placements: []InventorySlotPlacement{{
					Slot:       0,
					InstanceID: id,
				}},
			},
			{
				ID:       "equipment:main",
				Kind:     ContainerKindEquipment,
				Capacity: 3,
			},
		},
		Items: []InventoryItemStack{{
			InstanceID:   id,
			DefinitionID: "material_test",
			Quantity:     1,
		}},
	})
	if err != nil {
		t.Fatalf("NewInventoryState() error = %v", err)
	}
	if _, ok := state.Item(id); !ok {
		t.Fatalf("generated identity missing from canonical state")
	}
}

func TestAuthoritativeItemInstanceIDValidation(t *testing.T) {
	if _, err := NewItemInstanceIDAuthorityWithEntropy(nil, 1); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("nil entropy constructor error = %v", err)
	}
	if _, err := NewItemInstanceIDAuthorityWithEntropy(bytes.NewReader(nil), 0); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("zero attempts constructor error = %v", err)
	}
	if _, err := NewItemInstanceIDAuthorityWithEntropy(bytes.NewReader(nil), MaxItemInstanceIDCollisionAttempts+1); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("too many attempts constructor error = %v", err)
	}

	var nilAuthority *ItemInstanceIDAuthority
	if _, err := nilAuthority.ReserveNew(); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("nil authority ReserveNew error = %v", err)
	}
	if err := nilAuthority.ReserveExisting(itemInstanceIDFromSeed(t, 30)); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("nil authority ReserveExisting error = %v", err)
	}
	if err := nilAuthority.Release(itemInstanceIDFromSeed(t, 31)); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("nil authority Release error = %v", err)
	}
	if _, err := nilAuthority.IsReserved(itemInstanceIDFromSeed(t, 32)); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("nil authority IsReserved error = %v", err)
	}
	if nilAuthority.Count() != 0 || nilAuthority.ReservedIDs() != nil {
		t.Fatal("nil authority read methods returned non-zero data")
	}

	authority := &ItemInstanceIDAuthority{}
	if _, err := authority.ReserveNew(); !errors.Is(err, ErrItemInstanceIDAuthorityInvalid) {
		t.Fatalf("zero authority error = %v", err)
	}
}

func TestAuthoritativeItemInstanceIDEntropyReaderIsSerialized(t *testing.T) {
	reader := &serializedEntropyReader{}
	authority, err := NewItemInstanceIDAuthorityWithEntropy(reader, 1)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 64
	var wait sync.WaitGroup
	errorsChannel := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := authority.ReserveNew()
			errorsChannel <- err
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("ReserveNew() error = %v", err)
		}
	}
	if reader.concurrentReadDetected {
		t.Fatal("entropy reader was used concurrently")
	}
}

type serializedEntropyReader struct {
	mu                     sync.Mutex
	active                 bool
	counter                byte
	concurrentReadDetected bool
}

func (reader *serializedEntropyReader) Read(destination []byte) (int, error) {
	reader.mu.Lock()
	if reader.active {
		reader.concurrentReadDetected = true
	}
	reader.active = true
	seed := reader.counter
	reader.counter++
	reader.mu.Unlock()

	for index := range destination {
		destination[index] = seed + byte(index)
	}

	reader.mu.Lock()
	reader.active = false
	reader.mu.Unlock()
	return len(destination), nil
}

var _ io.Reader = itemInstanceIDErrorReader{}
var _ io.Reader = (*serializedEntropyReader)(nil)
