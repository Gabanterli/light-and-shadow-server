package inventory

import (
	"errors"
	"fmt"
)

var (
	ErrItemCreationServiceInvalid = errors.New("invalid authoritative item creation service")
	ErrItemCreationResultInvalid  = errors.New("authoritative item creation result is inconsistent")
	ErrItemCreationRollback       = errors.New("authoritative item creation reservation rollback failed")
)

// AuthoritativeItemCreationService coordinates item identity reservation with
// canonical inventory publication. Callers provide business data and target
// locations, but never provide the new ItemInstanceID.
type AuthoritativeItemCreationService struct {
	store     *InventoryStore
	authority *ItemInstanceIDAuthority
}

type CreateInventoryItemRequest struct {
	OwnerID          string
	ExpectedRevision InventoryRevision
	DefinitionID     string
	Quantity         int
	To               InventoryItemLocation
}

type SplitInventoryStackRequest struct {
	OwnerID          string
	ExpectedRevision InventoryRevision
	SourceInstanceID ItemInstanceID
	Quantity         int
	To               InventoryItemLocation
}

type AuthoritativeItemCreationResult struct {
	InstanceID ItemInstanceID
	Snapshot   InventoryStateSnapshot
	Operation  InventoryOperationResult
}

func NewAuthoritativeItemCreationService(
	store *InventoryStore,
	authority *ItemInstanceIDAuthority,
) (*AuthoritativeItemCreationService, error) {
	service := &AuthoritativeItemCreationService{
		store:     store,
		authority: authority,
	}
	if err := service.validate(); err != nil {
		return nil, err
	}
	return service, nil
}

// CreateItem reserves a new identity, applies an Add operation, and retains the
// reservation only when the canonical store publishes the item successfully.
func (service *AuthoritativeItemCreationService) CreateItem(
	request CreateInventoryItemRequest,
) (AuthoritativeItemCreationResult, error) {
	if err := service.validate(); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := validateItemCreationOwnerRevision(request.OwnerID, request.ExpectedRevision); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := validateOperationLocationID(request.To); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}

	definition, exists := service.store.catalog.Get(request.DefinitionID)
	if !exists {
		return AuthoritativeItemCreationResult{}, fmt.Errorf(
			"%w: unknown definition %q",
			ErrInventoryOperationInvalid,
			request.DefinitionID,
		)
	}
	if err := validateOperationItemQuantity(definition, request.Quantity); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := service.requireCurrentRevision(request.OwnerID, request.ExpectedRevision); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}

	instanceID, err := service.authority.ReserveNew()
	if err != nil {
		return AuthoritativeItemCreationResult{}, err
	}

	return service.applyReservedCreation(
		request.OwnerID,
		request.ExpectedRevision,
		instanceID,
		AddInventoryItemOperation{
			InstanceID:   instanceID,
			DefinitionID: request.DefinitionID,
			Quantity:     request.Quantity,
			To:           request.To,
		},
	)
}

// SplitStack reserves the identity for the newly created stack and releases it
// automatically when the split cannot be published.
func (service *AuthoritativeItemCreationService) SplitStack(
	request SplitInventoryStackRequest,
) (AuthoritativeItemCreationResult, error) {
	if err := service.validate(); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := validateItemCreationOwnerRevision(request.OwnerID, request.ExpectedRevision); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := validateOperationItemInstanceID(request.SourceInstanceID); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if err := validateOperationLocationID(request.To); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}
	if request.Quantity <= 0 {
		return AuthoritativeItemCreationResult{}, fmt.Errorf(
			"%w: split quantity must be greater than zero",
			ErrInventoryStackRule,
		)
	}
	if err := service.requireCurrentRevision(request.OwnerID, request.ExpectedRevision); err != nil {
		return AuthoritativeItemCreationResult{}, err
	}

	instanceID, err := service.authority.ReserveNew()
	if err != nil {
		return AuthoritativeItemCreationResult{}, err
	}

	return service.applyReservedCreation(
		request.OwnerID,
		request.ExpectedRevision,
		instanceID,
		SplitInventoryStackOperation{
			SourceInstanceID: request.SourceInstanceID,
			NewInstanceID:    instanceID,
			Quantity:         request.Quantity,
			To:               request.To,
		},
	)
}

func (service *AuthoritativeItemCreationService) applyReservedCreation(
	ownerID string,
	expectedRevision InventoryRevision,
	instanceID ItemInstanceID,
	operation InventoryOperation,
) (AuthoritativeItemCreationResult, error) {
	snapshot, operationResult, err := service.store.Apply(
		ownerID,
		expectedRevision,
		operation,
	)
	if err != nil {
		if releaseErr := service.authority.Release(instanceID); releaseErr != nil {
			return AuthoritativeItemCreationResult{}, errors.Join(
				err,
				fmt.Errorf("%w for %q: %v", ErrItemCreationRollback, instanceID, releaseErr),
			)
		}
		return AuthoritativeItemCreationResult{}, err
	}

	// Store.Apply has already published the state at this point. A consistency
	// failure must retain the reservation because releasing it would make a
	// published item identity available for reuse.
	if !itemCreationResultContainsID(operationResult.CreatedInstanceIDs, instanceID) ||
		!itemCreationSnapshotContainsID(snapshot, instanceID) {
		return AuthoritativeItemCreationResult{}, fmt.Errorf(
			"%w: published state does not report created identity %q",
			ErrItemCreationResultInvalid,
			instanceID,
		)
	}

	return AuthoritativeItemCreationResult{
		InstanceID: instanceID,
		Snapshot:   snapshot,
		Operation:  operationResult,
	}, nil
}

func (service *AuthoritativeItemCreationService) requireCurrentRevision(
	ownerID string,
	expectedRevision InventoryRevision,
) error {
	snapshot, err := service.store.Snapshot(ownerID)
	if err != nil {
		return err
	}
	if snapshot.Revision != expectedRevision {
		return fmt.Errorf(
			"%w: expected %d, current %d",
			ErrInventoryRevisionConflict,
			expectedRevision,
			snapshot.Revision,
		)
	}
	return nil
}

func (service *AuthoritativeItemCreationService) validate() error {
	if service == nil {
		return fmt.Errorf("%w: service is required", ErrItemCreationServiceInvalid)
	}
	if service.store == nil {
		return fmt.Errorf("%w: inventory store is required", ErrItemCreationServiceInvalid)
	}
	if err := service.store.validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrItemCreationServiceInvalid, err)
	}
	if service.authority == nil {
		return fmt.Errorf("%w: item instance ID authority is required", ErrItemCreationServiceInvalid)
	}
	if err := service.authority.validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrItemCreationServiceInvalid, err)
	}
	return nil
}

func validateItemCreationOwnerRevision(ownerID string, revision InventoryRevision) error {
	if err := validateInventoryStoreOwnerID(ownerID); err != nil {
		return err
	}
	if revision == 0 {
		return fmt.Errorf("%w: expected revision must be greater than zero", ErrItemCreationServiceInvalid)
	}
	return nil
}

func itemCreationResultContainsID(ids []ItemInstanceID, expected ItemInstanceID) bool {
	for _, id := range ids {
		if id == expected {
			return true
		}
	}
	return false
}

func itemCreationSnapshotContainsID(snapshot InventoryStateSnapshot, expected ItemInstanceID) bool {
	for _, item := range snapshot.Items {
		if item.InstanceID == expected {
			return true
		}
	}
	return false
}
