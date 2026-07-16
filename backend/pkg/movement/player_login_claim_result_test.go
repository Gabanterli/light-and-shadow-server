package movement

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func playerLoginClaimResultTestEntity(
	id string,
	x, y float64,
	z int,
) *Entity {
	return &Entity{
		ID:             id,
		Name:           "Player_" + id,
		X:              x,
		Y:              y,
		Z:              z,
		Type:           "player",
		BlocksMovement: true,
	}
}

func TestPlayerLoginClaimResultSequentialEntrants(
	t *testing.T,
) {
	index := NewSpatialIndex()

	first, err :=
		index.TryClaimPlayerLoginEntityWithResult(
			playerLoginClaimResultTestEntity(
				"player-first",
				100,
				100,
				0,
			),
		)
	if err != nil {
		t.Fatalf("first claim failed: %v", err)
	}

	if first.OverlappedExistingPlayer {
		t.Fatal(
			"first player reported an existing player overlap",
		)
	}

	second, err :=
		index.TryClaimPlayerLoginEntityWithResult(
			playerLoginClaimResultTestEntity(
				"player-second",
				100,
				100,
				0,
			),
		)
	if err != nil {
		t.Fatalf("second claim failed: %v", err)
	}

	if !second.OverlappedExistingPlayer {
		t.Fatal(
			"second player did not report the existing player",
		)
	}

	differentFloor, err :=
		index.TryClaimPlayerLoginEntityWithResult(
			playerLoginClaimResultTestEntity(
				"player-other-floor",
				100,
				100,
				1,
			),
		)
	if err != nil {
		t.Fatalf("different-floor claim failed: %v", err)
	}

	if differentFloor.OverlappedExistingPlayer {
		t.Fatal(
			"different-floor player reported an overlap",
		)
	}

	differentTile, err :=
		index.TryClaimPlayerLoginEntityWithResult(
			playerLoginClaimResultTestEntity(
				"player-other-tile",
				101,
				100,
				0,
			),
		)
	if err != nil {
		t.Fatalf("different-tile claim failed: %v", err)
	}

	if differentTile.OverlappedExistingPlayer {
		t.Fatal(
			"different-tile player reported an overlap",
		)
	}
}

func TestPlayerLoginClaimResultConcurrentSameTile(
	t *testing.T,
) {
	index := NewSpatialIndex()

	const playerCount = 64

	start := make(chan struct{})
	errorsChannel := make(chan error, playerCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(playerCount)

	var firstEntrants int32
	var laterEntrants int32

	for playerNumber := 0; playerNumber < playerCount; playerNumber++ {
		playerNumber := playerNumber

		go func() {
			defer waitGroup.Done()

			<-start

			playerID := fmt.Sprintf(
				"concurrent-player-%02d",
				playerNumber,
			)

			result, err :=
				index.TryClaimPlayerLoginEntityWithResult(
					playerLoginClaimResultTestEntity(
						playerID,
						200,
						200,
						0,
					),
				)
			if err != nil {
				errorsChannel <- fmt.Errorf(
					"claim %s: %w",
					playerID,
					err,
				)
				return
			}

			if result.OverlappedExistingPlayer {
				atomic.AddInt32(&laterEntrants, 1)
				return
			}

			atomic.AddInt32(&firstEntrants, 1)
		}()
	}

	close(start)
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		t.Error(err)
	}

	if got := atomic.LoadInt32(&firstEntrants); got != 1 {
		t.Fatalf(
			"first entrant results = %d, want 1",
			got,
		)
	}

	if got := atomic.LoadInt32(&laterEntrants); got != playerCount-1 {
		t.Fatalf(
			"later entrant results = %d, want %d",
			got,
			playerCount-1,
		)
	}

	for playerNumber := 0; playerNumber < playerCount; playerNumber++ {
		playerID := fmt.Sprintf(
			"concurrent-player-%02d",
			playerNumber,
		)

		if _, exists := index.GetEntity(playerID); !exists {
			t.Fatalf(
				"claimed player %q is missing",
				playerID,
			)
		}
	}
}

func TestPlayerLoginClaimResultConcurrentDifferentTiles(
	t *testing.T,
) {
	index := NewSpatialIndex()

	const playerCount = 32

	start := make(chan struct{})
	errorsChannel := make(chan error, playerCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(playerCount)

	var unexpectedOverlaps int32

	for playerNumber := 0; playerNumber < playerCount; playerNumber++ {
		playerNumber := playerNumber

		go func() {
			defer waitGroup.Done()

			<-start

			playerID := fmt.Sprintf(
				"different-tile-player-%02d",
				playerNumber,
			)

			result, err :=
				index.TryClaimPlayerLoginEntityWithResult(
					playerLoginClaimResultTestEntity(
						playerID,
						float64(300+playerNumber),
						300,
						0,
					),
				)
			if err != nil {
				errorsChannel <- fmt.Errorf(
					"claim %s: %w",
					playerID,
					err,
				)
				return
			}

			if result.OverlappedExistingPlayer {
				atomic.AddInt32(
					&unexpectedOverlaps,
					1,
				)
			}
		}()
	}

	close(start)
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		t.Error(err)
	}

	if got := atomic.LoadInt32(&unexpectedOverlaps); got != 0 {
		t.Fatalf(
			"unexpected overlap results = %d, want 0",
			got,
		)
	}
}

func TestPlayerLoginClaimOriginalMethodCompatibility(
	t *testing.T,
) {
	index := NewSpatialIndex()

	err := index.TryClaimPlayerLoginEntity(
		playerLoginClaimResultTestEntity(
			"legacy-wrapper-player",
			400,
			400,
			0,
		),
	)
	if err != nil {
		t.Fatalf(
			"original claim method failed: %v",
			err,
		)
	}

	entity, exists :=
		index.GetEntity("legacy-wrapper-player")
	if !exists || entity == nil {
		t.Fatal(
			"original claim method did not register the player",
		)
	}
}
