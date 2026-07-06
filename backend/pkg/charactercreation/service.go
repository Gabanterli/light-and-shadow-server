package charactercreation

import (
	"context"
	"fmt"
	"strings"

	"github.com/light-and-shadow/backend/pkg/persistence"
)

// CreateRequest holds the client's intent for creating a new character.
type CreateRequest struct {
	DesiredName string
	RaceID      string
}

// CharacterCreator is an interface that represents the persistence layer's
// capability to create a character. This allows for dependency injection and testing.
type CharacterCreator interface {
	CreateCharacterForAccount(ctx context.Context, accountID int, desiredName string, raceID string) (*persistence.CharacterSummary, string, error)
}

// RaceValidator is an interface for validating if a race is playable.
// This decouples the service from the hardcoded list of races, which will
// eventually be managed by the Rule Registry.
type RaceValidator interface {
	IsPlayableRace(raceID string) bool
}

// Service is the central point for handling character creation logic.
// It validates business rules before calling the persistence layer.
type Service struct {
	creator       CharacterCreator
	raceValidator RaceValidator
}

// NewService creates a new instance of the CharacterCreationService.
func NewService(creator CharacterCreator, raceValidator RaceValidator) *Service {
	return &Service{
		creator:       creator,
		raceValidator: raceValidator,
	}
}

// CreateCharacter processes a character creation request, validates it,
// and calls the persistence layer. It returns a summary of the created
// character or a client-safe error code.
func (s *Service) CreateCharacter(ctx context.Context, accountID int, req CreateRequest) (*persistence.CharacterSummary, string, error) {
	if accountID <= 0 {
		return nil, "not_authenticated", fmt.Errorf("invalid accountID: %d", accountID)
	}

	normalizedName := strings.TrimSpace(req.DesiredName)
	if normalizedName == "" {
		return nil, "invalid_name", fmt.Errorf("desired name cannot be empty")
	}

	normalizedRaceID := strings.TrimSpace(req.RaceID)
	if normalizedRaceID == "" {
		return nil, "invalid_race", fmt.Errorf("raceID cannot be empty")
	}

	if s.creator == nil {
		return nil, "internal_error", fmt.Errorf("character creator dependency is not configured")
	}
	if s.raceValidator == nil {
		return nil, "internal_error", fmt.Errorf("race validator dependency is not configured")
	}

	if !s.raceValidator.IsPlayableRace(normalizedRaceID) {
		return nil, "invalid_race", fmt.Errorf("raceID %q is not a playable race", normalizedRaceID)
	}

	// Call the persistence layer to perform the transactional creation.
	summary, errorCode, err := s.creator.CreateCharacterForAccount(ctx, accountID, normalizedName, normalizedRaceID)
	if err != nil {
		// If the persistence layer provided a specific error code, propagate it.
		if errorCode != "" {
			return nil, errorCode, err
		}
		// Otherwise, return a generic internal error to avoid leaking details.
		return nil, "internal_error", err
	}

	return summary, "", nil
}
