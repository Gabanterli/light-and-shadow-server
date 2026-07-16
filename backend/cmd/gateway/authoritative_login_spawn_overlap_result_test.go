package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/persistence"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

type overlapResultFixedLoginSpawnResolver struct {
	placement authoritativeSpawnPlacement
}

func (r overlapResultFixedLoginSpawnResolver) Resolve(
	_ authoritativeSpawnPlacementRequest,
) (authoritativeSpawnPlacement, error) {
	return r.placement, nil
}

type overlapResultUnusedPositionRelocator struct{}

func (overlapResultUnusedPositionRelocator) RelocateCharacterPosition(
	_ context.Context,
	_ persistence.CharacterPositionRelocationRequest,
) (persistence.CharacterPositionRelocationResult, error) {
	return persistence.CharacterPositionRelocationResult{}, nil
}

func activateOverlapResultPlayer(
	index *movement.SpatialIndex,
	playerID string,
	start <-chan struct{},
) (authoritativeLoginSpawnActivation, error) {
	if start != nil {
		<-start
	}

	return resolveClaimAndPersistAuthoritativeLoginSpawn(
		context.Background(),
		overlapResultFixedLoginSpawnResolver{
			placement: authoritativeLoginSpawnTestPlacement(
				500,
				500,
				0,
				false,
			),
		},
		index,
		overlapResultUnusedPositionRelocator{},
		authoritativeSpawnPlacementRequest{
			PlayerID:     playerID,
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			PersistedX:   500,
			PersistedY:   500,
			PersistedZ:   0,
		},
		1,
	)
}

func TestAuthoritativeLoginSpawnPropagatesEmptyTileOverlapResult(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	activation, err := activateOverlapResultPlayer(
		index,
		"single-player",
		nil,
	)
	if err != nil {
		t.Fatalf("activate single player: %v", err)
	}

	if activation.OverlappedExistingPlayer {
		t.Fatal(
			"empty-tile activation reported an existing player",
		)
	}
}

func TestAuthoritativeLoginSpawnPropagatesSequentialEntrantResult(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	first, err := activateOverlapResultPlayer(
		index,
		"first-player",
		nil,
	)
	if err != nil {
		t.Fatalf("activate first player: %v", err)
	}

	second, err := activateOverlapResultPlayer(
		index,
		"second-player",
		nil,
	)
	if err != nil {
		t.Fatalf("activate second player: %v", err)
	}

	if first.OverlappedExistingPlayer {
		t.Fatal(
			"first activation reported an existing player",
		)
	}

	if !second.OverlappedExistingPlayer {
		t.Fatal(
			"second activation did not report the first player",
		)
	}
}

func TestAuthoritativeLoginSpawnPropagatesConcurrentTwoPlayerAsymmetry(
	t *testing.T,
) {
	testConcurrentAuthoritativeLoginSpawnOverlapResults(
		t,
		2,
	)
}

func TestAuthoritativeLoginSpawnPropagatesConcurrentThreePlayerAsymmetry(
	t *testing.T,
) {
	testConcurrentAuthoritativeLoginSpawnOverlapResults(
		t,
		3,
	)
}

func testConcurrentAuthoritativeLoginSpawnOverlapResults(
	t *testing.T,
	playerCount int,
) {
	t.Helper()

	index := movement.NewSpatialIndex()
	start := make(chan struct{})

	type activationResult struct {
		activation authoritativeLoginSpawnActivation
		err        error
	}

	results := make(
		chan activationResult,
		playerCount,
	)

	var waitGroup sync.WaitGroup
	waitGroup.Add(playerCount)

	for playerNumber := 0; playerNumber < playerCount; playerNumber++ {
		playerID := fmt.Sprintf(
			"concurrent-player-%d",
			playerNumber,
		)

		go func(id string) {
			defer waitGroup.Done()

			activation, err := activateOverlapResultPlayer(
				index,
				id,
				start,
			)

			results <- activationResult{
				activation: activation,
				err:        err,
			}
		}(playerID)
	}

	close(start)
	waitGroup.Wait()
	close(results)

	firstEntrants := 0
	laterEntrants := 0

	for result := range results {
		if result.err != nil {
			t.Fatalf(
				"concurrent activation failed: %v",
				result.err,
			)
		}

		if result.activation.OverlappedExistingPlayer {
			laterEntrants++
			continue
		}

		firstEntrants++
	}

	if firstEntrants != 1 {
		t.Fatalf(
			"first entrant activations = %d, want 1",
			firstEntrants,
		)
	}

	if laterEntrants != playerCount-1 {
		t.Fatalf(
			"later entrant activations = %d, want %d",
			laterEntrants,
			playerCount-1,
		)
	}
}
