package inventory

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	MaxCanonicalInventoryContainers = 256
	MaxCanonicalInventoryItems      = 65535
	MaxCanonicalContainerSlots      = 65535
)

type ItemInstanceID string
type ContainerID string
type SlotID uint32
type InventoryRevision uint64

type ContainerKind string

const (
	ContainerKindBackpack  ContainerKind = "backpack"
	ContainerKindEquipment ContainerKind = "equipment"
)

var canonicalInventoryIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9:_-]*$`)

type InventoryItemStack struct {
	InstanceID   ItemInstanceID
	DefinitionID string
	Quantity     int
}

type InventorySlotPlacement struct {
	Slot       SlotID
	InstanceID ItemInstanceID
}

type InventoryContainerSnapshot struct {
	ID         ContainerID
	Kind       ContainerKind
	Capacity   uint32
	Placements []InventorySlotPlacement
}

type InventoryStateSnapshot struct {
	OwnerID    string
	Revision   InventoryRevision
	Containers []InventoryContainerSnapshot
	Items      []InventoryItemStack
}

type InventoryItemLocation struct {
	ContainerID ContainerID
	Slot        SlotID
}

type InventoryState struct {
	ownerID             string
	revision            InventoryRevision
	containers          map[ContainerID]InventoryContainerSnapshot
	items               map[ItemInstanceID]InventoryItemStack
	locations           map[ItemInstanceID]InventoryItemLocation
	orderedContainerIDs []ContainerID
	orderedItemIDs      []ItemInstanceID
}

func NewInventoryState(catalog *ItemCatalog, snapshot InventoryStateSnapshot) (*InventoryState, error) {
	if catalog == nil {
		return nil, errors.New("item catalog is required")
	}
	if strings.TrimSpace(snapshot.OwnerID) == "" {
		return nil, errors.New("inventory owner ID is required")
	}
	if strings.TrimSpace(snapshot.OwnerID) != snapshot.OwnerID {
		return nil, errors.New("inventory owner ID must not contain surrounding whitespace")
	}
	if snapshot.Revision == 0 {
		return nil, errors.New("inventory revision must be greater than zero")
	}
	if len(snapshot.Containers) == 0 {
		return nil, errors.New("inventory state must contain at least one container")
	}
	if len(snapshot.Containers) > MaxCanonicalInventoryContainers {
		return nil, fmt.Errorf("inventory state exceeds %d containers", MaxCanonicalInventoryContainers)
	}
	if len(snapshot.Items) > MaxCanonicalInventoryItems {
		return nil, fmt.Errorf("inventory state exceeds %d item instances", MaxCanonicalInventoryItems)
	}

	state := &InventoryState{
		ownerID:             snapshot.OwnerID,
		revision:            snapshot.Revision,
		containers:          make(map[ContainerID]InventoryContainerSnapshot, len(snapshot.Containers)),
		items:               make(map[ItemInstanceID]InventoryItemStack, len(snapshot.Items)),
		locations:           make(map[ItemInstanceID]InventoryItemLocation, len(snapshot.Items)),
		orderedContainerIDs: make([]ContainerID, 0, len(snapshot.Containers)),
		orderedItemIDs:      make([]ItemInstanceID, 0, len(snapshot.Items)),
	}

	for index, item := range snapshot.Items {
		if err := validateCanonicalInventoryID("item instance ID", string(item.InstanceID)); err != nil {
			return nil, fmt.Errorf("invalid item at index %d: %w", index, err)
		}
		if _, duplicate := state.items[item.InstanceID]; duplicate {
			return nil, fmt.Errorf("duplicate item instance ID %q", item.InstanceID)
		}
		definition, exists := catalog.Get(item.DefinitionID)
		if !exists {
			return nil, fmt.Errorf("item instance %q references unknown definition %q", item.InstanceID, item.DefinitionID)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("item instance %q quantity must be greater than zero", item.InstanceID)
		}
		if definition.Stackable {
			if item.Quantity > definition.MaxStack {
				return nil, fmt.Errorf("item instance %q quantity %d exceeds MaxStack %d", item.InstanceID, item.Quantity, definition.MaxStack)
			}
		} else if item.Quantity != 1 {
			return nil, fmt.Errorf("non-stackable item instance %q must have quantity 1", item.InstanceID)
		}
		state.items[item.InstanceID] = item
		state.orderedItemIDs = append(state.orderedItemIDs, item.InstanceID)
	}

	hasBackpack := false
	hasEquipment := false
	for index, container := range snapshot.Containers {
		if err := validateCanonicalInventoryID("container ID", string(container.ID)); err != nil {
			return nil, fmt.Errorf("invalid container at index %d: %w", index, err)
		}
		if _, duplicate := state.containers[container.ID]; duplicate {
			return nil, fmt.Errorf("duplicate container ID %q", container.ID)
		}
		switch container.Kind {
		case ContainerKindBackpack:
			hasBackpack = true
		case ContainerKindEquipment:
			hasEquipment = true
		default:
			return nil, fmt.Errorf("container %q has unsupported kind %q", container.ID, container.Kind)
		}
		if container.Capacity == 0 {
			return nil, fmt.Errorf("container %q capacity must be greater than zero", container.ID)
		}
		if container.Capacity > MaxCanonicalContainerSlots {
			return nil, fmt.Errorf("container %q capacity exceeds %d", container.ID, MaxCanonicalContainerSlots)
		}
		if len(container.Placements) > int(container.Capacity) {
			return nil, fmt.Errorf("container %q has more placements than capacity", container.ID)
		}

		copied := cloneInventoryContainerSnapshot(container)
		seenSlots := make(map[SlotID]struct{}, len(copied.Placements))
		for placementIndex, placement := range copied.Placements {
			if uint32(placement.Slot) >= copied.Capacity {
				return nil, fmt.Errorf("container %q placement %d uses slot %d outside capacity %d", copied.ID, placementIndex, placement.Slot, copied.Capacity)
			}
			if _, duplicate := seenSlots[placement.Slot]; duplicate {
				return nil, fmt.Errorf("container %q contains duplicate slot %d", copied.ID, placement.Slot)
			}
			seenSlots[placement.Slot] = struct{}{}
			if err := validateCanonicalInventoryID("placed item instance ID", string(placement.InstanceID)); err != nil {
				return nil, fmt.Errorf("container %q placement %d: %w", copied.ID, placementIndex, err)
			}
			if _, exists := state.items[placement.InstanceID]; !exists {
				return nil, fmt.Errorf("container %q slot %d references unknown item instance %q", copied.ID, placement.Slot, placement.InstanceID)
			}
			if previous, duplicate := state.locations[placement.InstanceID]; duplicate {
				return nil, fmt.Errorf("item instance %q is placed more than once (%s:%d and %s:%d)", placement.InstanceID, previous.ContainerID, previous.Slot, copied.ID, placement.Slot)
			}
			state.locations[placement.InstanceID] = InventoryItemLocation{ContainerID: copied.ID, Slot: placement.Slot}
		}
		sort.Slice(copied.Placements, func(i, j int) bool { return copied.Placements[i].Slot < copied.Placements[j].Slot })
		state.containers[copied.ID] = copied
		state.orderedContainerIDs = append(state.orderedContainerIDs, copied.ID)
	}

	if !hasBackpack {
		return nil, errors.New("inventory state requires at least one backpack container")
	}
	if !hasEquipment {
		return nil, errors.New("inventory state requires at least one equipment container")
	}
	for _, instanceID := range state.orderedItemIDs {
		if _, placed := state.locations[instanceID]; !placed {
			return nil, fmt.Errorf("item instance %q is not placed in any container", instanceID)
		}
	}

	sort.Slice(state.orderedContainerIDs, func(i, j int) bool { return state.orderedContainerIDs[i] < state.orderedContainerIDs[j] })
	sort.Slice(state.orderedItemIDs, func(i, j int) bool { return state.orderedItemIDs[i] < state.orderedItemIDs[j] })
	return state, nil
}

func (s *InventoryState) OwnerID() string {
	if s == nil {
		return ""
	}
	return s.ownerID
}
func (s *InventoryState) Revision() InventoryRevision {
	if s == nil {
		return 0
	}
	return s.revision
}
func (s *InventoryState) ContainerCount() int {
	if s == nil {
		return 0
	}
	return len(s.containers)
}
func (s *InventoryState) ItemCount() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}
func (s *InventoryState) ContainerIDs() []ContainerID {
	if s == nil {
		return nil
	}
	return append([]ContainerID(nil), s.orderedContainerIDs...)
}
func (s *InventoryState) ItemInstanceIDs() []ItemInstanceID {
	if s == nil {
		return nil
	}
	return append([]ItemInstanceID(nil), s.orderedItemIDs...)
}
func (s *InventoryState) Container(id ContainerID) (InventoryContainerSnapshot, bool) {
	if s == nil {
		return InventoryContainerSnapshot{}, false
	}
	c, ok := s.containers[id]
	if !ok {
		return InventoryContainerSnapshot{}, false
	}
	return cloneInventoryContainerSnapshot(c), true
}
func (s *InventoryState) Item(id ItemInstanceID) (InventoryItemStack, bool) {
	if s == nil {
		return InventoryItemStack{}, false
	}
	item, ok := s.items[id]
	return item, ok
}
func (s *InventoryState) LocationOf(id ItemInstanceID) (InventoryItemLocation, bool) {
	if s == nil {
		return InventoryItemLocation{}, false
	}
	loc, ok := s.locations[id]
	return loc, ok
}

func (s *InventoryState) Snapshot() InventoryStateSnapshot {
	if s == nil {
		return InventoryStateSnapshot{}
	}
	snapshot := InventoryStateSnapshot{OwnerID: s.ownerID, Revision: s.revision, Containers: make([]InventoryContainerSnapshot, 0, len(s.orderedContainerIDs)), Items: make([]InventoryItemStack, 0, len(s.orderedItemIDs))}
	for _, id := range s.orderedContainerIDs {
		snapshot.Containers = append(snapshot.Containers, cloneInventoryContainerSnapshot(s.containers[id]))
	}
	for _, id := range s.orderedItemIDs {
		snapshot.Items = append(snapshot.Items, s.items[id])
	}
	return snapshot
}

func validateCanonicalInventoryID(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s %q must not contain surrounding whitespace", field, value)
	}
	if !canonicalInventoryIDPattern.MatchString(value) {
		return fmt.Errorf("%s %q must match %s", field, value, canonicalInventoryIDPattern.String())
	}
	return nil
}

func cloneInventoryContainerSnapshot(container InventoryContainerSnapshot) InventoryContainerSnapshot {
	copyContainer := container
	copyContainer.Placements = append([]InventorySlotPlacement(nil), container.Placements...)
	return copyContainer
}
