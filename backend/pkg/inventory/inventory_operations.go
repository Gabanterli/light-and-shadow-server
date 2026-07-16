package inventory

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

var (
	ErrInventoryRevisionConflict   = errors.New("inventory revision conflict")
	ErrInventoryRevisionOverflow   = errors.New("inventory revision overflow")
	ErrInventoryOperationInvalid   = errors.New("invalid inventory operation")
	ErrInventoryOperationNoOp      = errors.New("inventory operation would not change state")
	ErrInventoryItemNotFound       = errors.New("inventory item instance not found")
	ErrInventoryContainerNotFound  = errors.New("inventory container not found")
	ErrInventorySlotOutOfRange     = errors.New("inventory slot is outside container capacity")
	ErrInventorySlotOccupied       = errors.New("inventory slot is occupied")
	ErrInventoryInstanceExists     = errors.New("inventory item instance already exists")
	ErrInventoryDefinitionMismatch = errors.New("inventory item definitions do not match")
	ErrInventoryStackRule          = errors.New("inventory stack rule violation")
)

type InventoryOperationKind string

const (
	InventoryOperationMove   InventoryOperationKind = "move"
	InventoryOperationSwap   InventoryOperationKind = "swap"
	InventoryOperationMerge  InventoryOperationKind = "merge"
	InventoryOperationSplit  InventoryOperationKind = "split"
	InventoryOperationAdd    InventoryOperationKind = "add"
	InventoryOperationRemove InventoryOperationKind = "remove"
)

type InventoryOperation interface {
	inventoryOperationKind() InventoryOperationKind
	applyInventoryOperation(*inventoryOperationWorkingState, *ItemCatalog, *InventoryOperationResult) error
}

type MoveInventoryItemOperation struct {
	InstanceID ItemInstanceID
	To         InventoryItemLocation
}

func (MoveInventoryItemOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationMove
}

func (operation MoveInventoryItemOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	_ *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.InstanceID); err != nil {
		return err
	}
	if err := validateOperationLocationID(operation.To); err != nil {
		return err
	}

	currentLocation, exists := working.locations[operation.InstanceID]
	if !exists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.InstanceID)
	}
	if currentLocation == operation.To {
		return fmt.Errorf("%w: item %q is already at %s:%d", ErrInventoryOperationNoOp, operation.InstanceID, operation.To.ContainerID, operation.To.Slot)
	}

	targetContainer, err := working.requireAvailableSlot(operation.To)
	if err != nil {
		return err
	}
	sourceContainer := working.containers[currentLocation.ContainerID]
	delete(sourceContainer.placements, currentLocation.Slot)
	targetContainer.placements[operation.To.Slot] = operation.InstanceID
	working.locations[operation.InstanceID] = operation.To
	result.ChangedInstanceIDs = append(result.ChangedInstanceIDs, operation.InstanceID)
	return nil
}

type SwapInventoryItemsOperation struct {
	FirstInstanceID  ItemInstanceID
	SecondInstanceID ItemInstanceID
}

func (SwapInventoryItemsOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationSwap
}

func (operation SwapInventoryItemsOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	_ *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.FirstInstanceID); err != nil {
		return err
	}
	if err := validateOperationItemInstanceID(operation.SecondInstanceID); err != nil {
		return err
	}
	if operation.FirstInstanceID == operation.SecondInstanceID {
		return fmt.Errorf("%w: swap requires two distinct item instances", ErrInventoryOperationNoOp)
	}

	firstLocation, firstExists := working.locations[operation.FirstInstanceID]
	if !firstExists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.FirstInstanceID)
	}
	secondLocation, secondExists := working.locations[operation.SecondInstanceID]
	if !secondExists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.SecondInstanceID)
	}

	working.containers[firstLocation.ContainerID].placements[firstLocation.Slot] = operation.SecondInstanceID
	working.containers[secondLocation.ContainerID].placements[secondLocation.Slot] = operation.FirstInstanceID
	working.locations[operation.FirstInstanceID] = secondLocation
	working.locations[operation.SecondInstanceID] = firstLocation
	result.ChangedInstanceIDs = append(
		result.ChangedInstanceIDs,
		operation.FirstInstanceID,
		operation.SecondInstanceID,
	)
	return nil
}

type MergeInventoryStacksOperation struct {
	SourceInstanceID      ItemInstanceID
	DestinationInstanceID ItemInstanceID
	Quantity              int
}

func (MergeInventoryStacksOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationMerge
}

func (operation MergeInventoryStacksOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	catalog *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.SourceInstanceID); err != nil {
		return err
	}
	if err := validateOperationItemInstanceID(operation.DestinationInstanceID); err != nil {
		return err
	}
	if operation.SourceInstanceID == operation.DestinationInstanceID {
		return fmt.Errorf("%w: merge requires distinct source and destination instances", ErrInventoryOperationNoOp)
	}
	if operation.Quantity <= 0 {
		return fmt.Errorf("%w: merge quantity must be greater than zero", ErrInventoryStackRule)
	}

	source, sourceExists := working.items[operation.SourceInstanceID]
	if !sourceExists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.SourceInstanceID)
	}
	destination, destinationExists := working.items[operation.DestinationInstanceID]
	if !destinationExists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.DestinationInstanceID)
	}
	if source.DefinitionID != destination.DefinitionID {
		return fmt.Errorf(
			"%w: source %q uses %q and destination %q uses %q",
			ErrInventoryDefinitionMismatch,
			operation.SourceInstanceID,
			source.DefinitionID,
			operation.DestinationInstanceID,
			destination.DefinitionID,
		)
	}

	definition, exists := catalog.Get(source.DefinitionID)
	if !exists {
		return fmt.Errorf("%w: unknown definition %q", ErrInventoryOperationInvalid, source.DefinitionID)
	}
	if !definition.Stackable {
		return fmt.Errorf("%w: definition %q is not stackable", ErrInventoryStackRule, source.DefinitionID)
	}
	if operation.Quantity > source.Quantity {
		return fmt.Errorf("%w: merge quantity %d exceeds source quantity %d", ErrInventoryStackRule, operation.Quantity, source.Quantity)
	}
	if destination.Quantity+operation.Quantity > definition.MaxStack {
		return fmt.Errorf(
			"%w: destination quantity would exceed MaxStack %d",
			ErrInventoryStackRule,
			definition.MaxStack,
		)
	}

	source.Quantity -= operation.Quantity
	destination.Quantity += operation.Quantity
	working.items[operation.DestinationInstanceID] = destination
	result.ChangedInstanceIDs = append(result.ChangedInstanceIDs, operation.DestinationInstanceID)

	if source.Quantity == 0 {
		working.removeInstance(operation.SourceInstanceID)
		result.RemovedInstanceIDs = append(result.RemovedInstanceIDs, operation.SourceInstanceID)
	} else {
		working.items[operation.SourceInstanceID] = source
		result.ChangedInstanceIDs = append(result.ChangedInstanceIDs, operation.SourceInstanceID)
	}
	return nil
}

type SplitInventoryStackOperation struct {
	SourceInstanceID ItemInstanceID
	NewInstanceID    ItemInstanceID
	Quantity         int
	To               InventoryItemLocation
}

func (SplitInventoryStackOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationSplit
}

func (operation SplitInventoryStackOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	catalog *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.SourceInstanceID); err != nil {
		return err
	}
	if err := validateOperationItemInstanceID(operation.NewInstanceID); err != nil {
		return err
	}
	if err := validateOperationLocationID(operation.To); err != nil {
		return err
	}
	if operation.SourceInstanceID == operation.NewInstanceID {
		return fmt.Errorf("%w: split destination instance must be new", ErrInventoryInstanceExists)
	}
	if _, exists := working.items[operation.NewInstanceID]; exists {
		return fmt.Errorf("%w: %q", ErrInventoryInstanceExists, operation.NewInstanceID)
	}
	if operation.Quantity <= 0 {
		return fmt.Errorf("%w: split quantity must be greater than zero", ErrInventoryStackRule)
	}

	source, exists := working.items[operation.SourceInstanceID]
	if !exists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.SourceInstanceID)
	}
	definition, exists := catalog.Get(source.DefinitionID)
	if !exists {
		return fmt.Errorf("%w: unknown definition %q", ErrInventoryOperationInvalid, source.DefinitionID)
	}
	if !definition.Stackable {
		return fmt.Errorf("%w: definition %q is not stackable", ErrInventoryStackRule, source.DefinitionID)
	}
	if operation.Quantity >= source.Quantity {
		return fmt.Errorf(
			"%w: split quantity %d must be less than source quantity %d",
			ErrInventoryStackRule,
			operation.Quantity,
			source.Quantity,
		)
	}

	targetContainer, err := working.requireAvailableSlot(operation.To)
	if err != nil {
		return err
	}

	source.Quantity -= operation.Quantity
	working.items[operation.SourceInstanceID] = source
	working.items[operation.NewInstanceID] = InventoryItemStack{
		InstanceID:   operation.NewInstanceID,
		DefinitionID: source.DefinitionID,
		Quantity:     operation.Quantity,
	}
	targetContainer.placements[operation.To.Slot] = operation.NewInstanceID
	working.locations[operation.NewInstanceID] = operation.To
	result.CreatedInstanceIDs = append(result.CreatedInstanceIDs, operation.NewInstanceID)
	result.ChangedInstanceIDs = append(result.ChangedInstanceIDs, operation.SourceInstanceID)
	return nil
}

type AddInventoryItemOperation struct {
	InstanceID   ItemInstanceID
	DefinitionID string
	Quantity     int
	To           InventoryItemLocation
}

func (AddInventoryItemOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationAdd
}

func (operation AddInventoryItemOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	catalog *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.InstanceID); err != nil {
		return err
	}
	if err := validateOperationLocationID(operation.To); err != nil {
		return err
	}
	if _, exists := working.items[operation.InstanceID]; exists {
		return fmt.Errorf("%w: %q", ErrInventoryInstanceExists, operation.InstanceID)
	}

	definition, exists := catalog.Get(operation.DefinitionID)
	if !exists {
		return fmt.Errorf("%w: unknown definition %q", ErrInventoryOperationInvalid, operation.DefinitionID)
	}
	if err := validateOperationItemQuantity(definition, operation.Quantity); err != nil {
		return err
	}
	targetContainer, err := working.requireAvailableSlot(operation.To)
	if err != nil {
		return err
	}

	working.items[operation.InstanceID] = InventoryItemStack{
		InstanceID:   operation.InstanceID,
		DefinitionID: operation.DefinitionID,
		Quantity:     operation.Quantity,
	}
	targetContainer.placements[operation.To.Slot] = operation.InstanceID
	working.locations[operation.InstanceID] = operation.To
	result.CreatedInstanceIDs = append(result.CreatedInstanceIDs, operation.InstanceID)
	return nil
}

type RemoveInventoryItemOperation struct {
	InstanceID ItemInstanceID
	Quantity   int
}

func (RemoveInventoryItemOperation) inventoryOperationKind() InventoryOperationKind {
	return InventoryOperationRemove
}

func (operation RemoveInventoryItemOperation) applyInventoryOperation(
	working *inventoryOperationWorkingState,
	_ *ItemCatalog,
	result *InventoryOperationResult,
) error {
	if err := validateOperationItemInstanceID(operation.InstanceID); err != nil {
		return err
	}
	if operation.Quantity <= 0 {
		return fmt.Errorf("%w: remove quantity must be greater than zero", ErrInventoryStackRule)
	}

	item, exists := working.items[operation.InstanceID]
	if !exists {
		return fmt.Errorf("%w: %q", ErrInventoryItemNotFound, operation.InstanceID)
	}
	if operation.Quantity > item.Quantity {
		return fmt.Errorf(
			"%w: remove quantity %d exceeds instance quantity %d",
			ErrInventoryStackRule,
			operation.Quantity,
			item.Quantity,
		)
	}

	item.Quantity -= operation.Quantity
	if item.Quantity == 0 {
		working.removeInstance(operation.InstanceID)
		result.RemovedInstanceIDs = append(result.RemovedInstanceIDs, operation.InstanceID)
	} else {
		working.items[operation.InstanceID] = item
		result.ChangedInstanceIDs = append(result.ChangedInstanceIDs, operation.InstanceID)
	}
	return nil
}

type InventoryOperationResult struct {
	Kind               InventoryOperationKind
	PreviousRevision   InventoryRevision
	NewRevision        InventoryRevision
	CreatedInstanceIDs []ItemInstanceID
	RemovedInstanceIDs []ItemInstanceID
	ChangedInstanceIDs []ItemInstanceID
}

func ApplyInventoryOperation(
	catalog *ItemCatalog,
	current *InventoryState,
	expectedRevision InventoryRevision,
	operation InventoryOperation,
) (*InventoryState, InventoryOperationResult, error) {
	if catalog == nil {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: item catalog is required", ErrInventoryOperationInvalid)
	}
	if current == nil {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: current inventory state is required", ErrInventoryOperationInvalid)
	}
	if operation == nil || inventoryOperationIsNil(operation) {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: operation is required", ErrInventoryOperationInvalid)
	}

	validatedCurrent, err := NewInventoryState(catalog, current.Snapshot())
	if err != nil {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: current state failed catalog validation: %v", ErrInventoryOperationInvalid, err)
	}
	if expectedRevision != validatedCurrent.Revision() {
		return nil, InventoryOperationResult{}, fmt.Errorf(
			"%w: expected %d, current %d",
			ErrInventoryRevisionConflict,
			expectedRevision,
			validatedCurrent.Revision(),
		)
	}
	if validatedCurrent.Revision() == InventoryRevision(^uint64(0)) {
		return nil, InventoryOperationResult{}, ErrInventoryRevisionOverflow
	}

	working := newInventoryOperationWorkingState(validatedCurrent)
	result := InventoryOperationResult{
		Kind:             operation.inventoryOperationKind(),
		PreviousRevision: validatedCurrent.Revision(),
		NewRevision:      validatedCurrent.Revision() + 1,
	}
	if err := operation.applyInventoryOperation(working, catalog, &result); err != nil {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: %s: %w", ErrInventoryOperationInvalid, operation.inventoryOperationKind(), err)
	}

	working.revision = result.NewRevision
	next, err := NewInventoryState(catalog, working.snapshot())
	if err != nil {
		return nil, InventoryOperationResult{}, fmt.Errorf("%w: %s produced invalid state: %v", ErrInventoryOperationInvalid, operation.inventoryOperationKind(), err)
	}

	normalizeInventoryOperationResult(&result)
	return next, result, nil
}

type inventoryOperationContainer struct {
	id         ContainerID
	kind       ContainerKind
	capacity   uint32
	placements map[SlotID]ItemInstanceID
}

type inventoryOperationWorkingState struct {
	ownerID    string
	revision   InventoryRevision
	containers map[ContainerID]*inventoryOperationContainer
	items      map[ItemInstanceID]InventoryItemStack
	locations  map[ItemInstanceID]InventoryItemLocation
}

func newInventoryOperationWorkingState(state *InventoryState) *inventoryOperationWorkingState {
	working := &inventoryOperationWorkingState{
		ownerID:    state.ownerID,
		revision:   state.revision,
		containers: make(map[ContainerID]*inventoryOperationContainer, len(state.containers)),
		items:      make(map[ItemInstanceID]InventoryItemStack, len(state.items)),
		locations:  make(map[ItemInstanceID]InventoryItemLocation, len(state.locations)),
	}
	for id, container := range state.containers {
		copyContainer := &inventoryOperationContainer{
			id:         container.ID,
			kind:       container.Kind,
			capacity:   container.Capacity,
			placements: make(map[SlotID]ItemInstanceID, len(container.Placements)),
		}
		for _, placement := range container.Placements {
			copyContainer.placements[placement.Slot] = placement.InstanceID
		}
		working.containers[id] = copyContainer
	}
	for id, item := range state.items {
		working.items[id] = item
	}
	for id, location := range state.locations {
		working.locations[id] = location
	}
	return working
}

func (working *inventoryOperationWorkingState) requireAvailableSlot(
	location InventoryItemLocation,
) (*inventoryOperationContainer, error) {
	container, exists := working.containers[location.ContainerID]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrInventoryContainerNotFound, location.ContainerID)
	}
	if uint32(location.Slot) >= container.capacity {
		return nil, fmt.Errorf(
			"%w: %s:%d has capacity %d",
			ErrInventorySlotOutOfRange,
			location.ContainerID,
			location.Slot,
			container.capacity,
		)
	}
	if occupant, occupied := container.placements[location.Slot]; occupied {
		return nil, fmt.Errorf(
			"%w: %s:%d contains %q",
			ErrInventorySlotOccupied,
			location.ContainerID,
			location.Slot,
			occupant,
		)
	}
	return container, nil
}

func (working *inventoryOperationWorkingState) removeInstance(instanceID ItemInstanceID) {
	location := working.locations[instanceID]
	delete(working.containers[location.ContainerID].placements, location.Slot)
	delete(working.locations, instanceID)
	delete(working.items, instanceID)
}

func (working *inventoryOperationWorkingState) snapshot() InventoryStateSnapshot {
	containerIDs := make([]ContainerID, 0, len(working.containers))
	for id := range working.containers {
		containerIDs = append(containerIDs, id)
	}
	sort.Slice(containerIDs, func(i, j int) bool { return containerIDs[i] < containerIDs[j] })

	containers := make([]InventoryContainerSnapshot, 0, len(containerIDs))
	for _, id := range containerIDs {
		container := working.containers[id]
		slots := make([]SlotID, 0, len(container.placements))
		for slot := range container.placements {
			slots = append(slots, slot)
		}
		sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })

		placements := make([]InventorySlotPlacement, 0, len(slots))
		for _, slot := range slots {
			placements = append(placements, InventorySlotPlacement{
				Slot:       slot,
				InstanceID: container.placements[slot],
			})
		}
		containers = append(containers, InventoryContainerSnapshot{
			ID:         container.id,
			Kind:       container.kind,
			Capacity:   container.capacity,
			Placements: placements,
		})
	}

	itemIDs := make([]ItemInstanceID, 0, len(working.items))
	for id := range working.items {
		itemIDs = append(itemIDs, id)
	}
	sort.Slice(itemIDs, func(i, j int) bool { return itemIDs[i] < itemIDs[j] })
	items := make([]InventoryItemStack, 0, len(itemIDs))
	for _, id := range itemIDs {
		items = append(items, working.items[id])
	}

	return InventoryStateSnapshot{
		OwnerID:    working.ownerID,
		Revision:   working.revision,
		Containers: containers,
		Items:      items,
	}
}

func validateOperationItemInstanceID(instanceID ItemInstanceID) error {
	if err := validateCanonicalInventoryID("item instance ID", string(instanceID)); err != nil {
		return fmt.Errorf("%w: %v", ErrInventoryOperationInvalid, err)
	}
	return nil
}

func validateOperationLocationID(location InventoryItemLocation) error {
	if err := validateCanonicalInventoryID("container ID", string(location.ContainerID)); err != nil {
		return fmt.Errorf("%w: %v", ErrInventoryOperationInvalid, err)
	}
	return nil
}

func validateOperationItemQuantity(definition ItemDefinition, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", ErrInventoryStackRule)
	}
	if definition.Stackable {
		if quantity > definition.MaxStack {
			return fmt.Errorf(
				"%w: quantity %d exceeds MaxStack %d for %q",
				ErrInventoryStackRule,
				quantity,
				definition.MaxStack,
				definition.ID,
			)
		}
		return nil
	}
	if quantity != 1 {
		return fmt.Errorf("%w: non-stackable definition %q requires quantity 1", ErrInventoryStackRule, definition.ID)
	}
	return nil
}

func inventoryOperationIsNil(operation InventoryOperation) bool {
	value := reflect.ValueOf(operation)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func normalizeInventoryOperationResult(result *InventoryOperationResult) {
	result.CreatedInstanceIDs = sortedUniqueItemInstanceIDs(result.CreatedInstanceIDs)
	result.RemovedInstanceIDs = sortedUniqueItemInstanceIDs(result.RemovedInstanceIDs)
	result.ChangedInstanceIDs = sortedUniqueItemInstanceIDs(result.ChangedInstanceIDs)
}

func sortedUniqueItemInstanceIDs(ids []ItemInstanceID) []ItemInstanceID {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[ItemInstanceID]struct{}, len(ids))
	result := make([]ItemInstanceID, 0, len(ids))
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
