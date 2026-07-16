package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/persistence"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

const maxAuthoritativeLoginSpawnClaimAttempts = 3

type authoritativeLoginSpawnActivation struct {
	Placement                authoritativeSpawnPlacement
	Version                  int
	ClaimAttempts            int
	OverlappedExistingPlayer bool
}

type authoritativeLoginSpawnUnavailableError struct {
	PlayerID string
	Attempts int
	Err      error
}

func (e *authoritativeLoginSpawnUnavailableError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf(
			"authoritative login spawn unavailable for player %q after %d attempts",
			e.PlayerID,
			e.Attempts,
		)
	}

	return fmt.Sprintf(
		"authoritative login spawn unavailable for player %q after %d attempts: %v",
		e.PlayerID,
		e.Attempts,
		e.Err,
	)
}

func (e *authoritativeLoginSpawnUnavailableError) Unwrap() error {
	return e.Err
}

type authoritativeLoginSpawnResolver interface {
	Resolve(
		request authoritativeSpawnPlacementRequest,
	) (authoritativeSpawnPlacement, error)
}

type authoritativeLoginSpawnClaimer interface {
	TryClaimPlayerLoginEntity(
		entity *movement.Entity,
	) error

	RemoveEntity(id string)
}

// authoritativeLoginSpawnClaimerWithResult is implemented by
// claimers that report facts observed under the same lock that
// publishes the player in the authoritative SpatialIndex.
type authoritativeLoginSpawnClaimerWithResult interface {
	TryClaimPlayerLoginEntityWithResult(
		entity *movement.Entity,
	) (movement.PlayerLoginClaimResult, error)
}

func claimAuthoritativeLoginSpawnEntity(
	claimer authoritativeLoginSpawnClaimer,
	entity *movement.Entity,
) (movement.PlayerLoginClaimResult, error) {
	claimerWithResult, supportsResult :=
		claimer.(authoritativeLoginSpawnClaimerWithResult)

	if supportsResult {
		return claimerWithResult.TryClaimPlayerLoginEntityWithResult(
			entity,
		)
	}

	err := claimer.TryClaimPlayerLoginEntity(entity)

	return movement.PlayerLoginClaimResult{}, err
}

type authoritativeLoginPositionRelocator interface {
	RelocateCharacterPosition(
		ctx context.Context,
		request persistence.CharacterPositionRelocationRequest,
	) (persistence.CharacterPositionRelocationResult, error)
}

type spatialIndexPlayerLoginSpawnDynamicOccupancy struct {
	index *movement.SpatialIndex
}

func newSpatialIndexPlayerLoginSpawnDynamicOccupancy(
	index *movement.SpatialIndex,
) authoritativeSpawnDynamicOccupancy {
	if index == nil {
		return nil
	}

	return &spatialIndexPlayerLoginSpawnDynamicOccupancy{
		index: index,
	}
}

func (a *spatialIndexPlayerLoginSpawnDynamicOccupancy) IsSpawnTileOccupied(
	excludeEntityID string,
	_ worldmap.WorldSpaceID,
	x, y float64,
	z int,
) bool {
	return a.index.IsPlayerLoginTileBlocked(
		excludeEntityID,
		x,
		y,
		z,
	)
}

func (s *GatewayServer) prepareAuthoritativeLoginSpawn(
	request authoritativeSpawnPlacementRequest,
	expectedVersion int,
) (authoritativeLoginSpawnActivation, error) {
	if s == nil {
		return authoritativeLoginSpawnActivation{},
			errors.New(
				"authoritative login spawn requires GatewayServer",
			)
	}

	return resolveClaimAndPersistAuthoritativeLoginSpawn(
		context.Background(),
		s.authoritativeSpawnResolver,
		s.spatialIndex,
		s.persistenceMgr,
		request,
		expectedVersion,
	)
}

func resolveClaimAndPersistAuthoritativeLoginSpawn(
	ctx context.Context,
	resolver authoritativeLoginSpawnResolver,
	claimer authoritativeLoginSpawnClaimer,
	relocator authoritativeLoginPositionRelocator,
	request authoritativeSpawnPlacementRequest,
	expectedVersion int,
) (authoritativeLoginSpawnActivation, error) {
	playerID := strings.TrimSpace(request.PlayerID)
	request.PlayerID = playerID

	if playerID == "" {
		return authoritativeLoginSpawnActivation{},
			errors.New(
				"authoritative login spawn player ID cannot be empty",
			)
	}

	if resolver == nil {
		return authoritativeLoginSpawnActivation{},
			errors.New(
				"authoritative login spawn resolver is unavailable",
			)
	}

	if claimer == nil {
		return authoritativeLoginSpawnActivation{},
			errors.New(
				"authoritative login spawn claimer is unavailable",
			)
	}

	if relocator == nil {
		return authoritativeLoginSpawnActivation{},
			errors.New(
				"authoritative login position relocator is unavailable",
			)
	}

	if expectedVersion < 1 {
		return authoritativeLoginSpawnActivation{},
			fmt.Errorf(
				"authoritative login spawn expected version must be positive, got %d",
				expectedVersion,
			)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var lastClaimError error

	for attempt := 1; attempt <= maxAuthoritativeLoginSpawnClaimAttempts; attempt++ {
		placement, err := resolver.Resolve(request)
		if err != nil {
			return authoritativeLoginSpawnActivation{},
				fmt.Errorf(
					"resolve authoritative login spawn for player %q: %w",
					playerID,
					err,
				)
		}

		position := placement.Position

		claimResult, claimError :=
			claimAuthoritativeLoginSpawnEntity(
				claimer,
				&movement.Entity{
					ID:             playerID,
					Name:           "Player_" + playerID,
					X:              position.X,
					Y:              position.Y,
					Z:              position.Z,
					Type:           "player",
					BlocksMovement: true,
				},
			)
		if claimError != nil {
			var blocked *movement.SpatialPlayerLoginClaimBlockedError

			if errors.As(claimError, &blocked) {
				lastClaimError = claimError
				continue
			}

			return authoritativeLoginSpawnActivation{},
				fmt.Errorf(
					"claim authoritative login spawn for player %q: %w",
					playerID,
					claimError,
				)
		}

		resultVersion := expectedVersion

		if placement.Relocated {
			relocationResult, relocationError :=
				relocator.RelocateCharacterPosition(
					ctx,
					persistence.CharacterPositionRelocationRequest{
						PlayerID:        playerID,
						X:               position.X,
						Y:               position.Y,
						Z:               float64(position.Z),
						ExpectedVersion: expectedVersion,
					},
				)
			if relocationError != nil {
				claimer.RemoveEntity(playerID)

				return authoritativeLoginSpawnActivation{},
					fmt.Errorf(
						"persist authoritative login relocation for player %q: %w",
						playerID,
						relocationError,
					)
			}

			expectedNewVersion := expectedVersion + 1

			if relocationResult.NewVersion != expectedNewVersion {
				claimer.RemoveEntity(playerID)

				return authoritativeLoginSpawnActivation{},
					fmt.Errorf(
						"unexpected authoritative login relocation version for player %q: got %d, want %d",
						playerID,
						relocationResult.NewVersion,
						expectedNewVersion,
					)
			}

			resultVersion = relocationResult.NewVersion
		}

		return authoritativeLoginSpawnActivation{
			Placement:                placement,
			Version:                  resultVersion,
			ClaimAttempts:            attempt,
			OverlappedExistingPlayer: claimResult.OverlappedExistingPlayer,
		}, nil
	}

	return authoritativeLoginSpawnActivation{},
		&authoritativeLoginSpawnUnavailableError{
			PlayerID: playerID,
			Attempts: maxAuthoritativeLoginSpawnClaimAttempts,
			Err:      lastClaimError,
		}
}
