package movement

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func playerLoginOverlapTestEntity(
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

func TestPlayerLoginClaimAllowsPlayerOverlap(
	t *testing.T,
) {
	index := NewSpatialIndex()

	first := playerLoginOverlapTestEntity(
		"player-a",
		100,
		100,
		0,
	)
	second := playerLoginOverlapTestEntity(
		"player-b",
		100,
		100,
		0,
	)

	if err := index.TryClaimPlayerLoginEntity(first); err != nil {
		t.Fatalf(
			"claim first player: %v",
			err,
		)
	}

	if err := index.TryClaimPlayerLoginEntity(second); err != nil {
		t.Fatalf(
			"claim overlapping player: %v",
			err,
		)
	}

	if !index.HasOverlappingPlayer("player-a") {
		t.Fatal(
			"first player did not detect overlap",
		)
	}

	if !index.HasOverlappingPlayer("player-b") {
		t.Fatal(
			"second player did not detect overlap",
		)
	}
}

func TestPlayerLoginClaimRejectsBlockingNPC(
	t *testing.T,
) {
	index := NewSpatialIndex()

	index.RegisterEntity(&Entity{
		ID:             "npc",
		Name:           "Blocking NPC",
		X:              100,
		Y:              100,
		Z:              0,
		Type:           "npc",
		BlocksMovement: true,
	})

	err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player",
			100,
			100,
			0,
		),
	)
	if err == nil {
		t.Fatal(
			"expected blocking NPC rejection",
		)
	}

	var blocked *SpatialPlayerLoginClaimBlockedError

	if !errors.As(err, &blocked) {
		t.Fatalf(
			"error type = %T, want SpatialPlayerLoginClaimBlockedError",
			err,
		)
	}

	if blocked.BlockerID != "npc" ||
		blocked.BlockerType != "npc" {
		t.Fatalf(
			"unexpected blocker: %+v",
			blocked,
		)
	}

	if _, exists := index.GetEntity("player"); exists {
		t.Fatal(
			"rejected player was registered",
		)
	}
}

func TestPlayerLoginClaimRejectsBlockingCreature(
	t *testing.T,
) {
	index := NewSpatialIndex()

	index.RegisterEntity(&Entity{
		ID:             "creature",
		Name:           "Blocking Creature",
		X:              100,
		Y:              100,
		Z:              0,
		Type:           "creature",
		BlocksMovement: true,
	})

	err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player",
			100,
			100,
			0,
		),
	)
	if err == nil {
		t.Fatal(
			"expected blocking creature rejection",
		)
	}

	var blocked *SpatialPlayerLoginClaimBlockedError

	if !errors.As(err, &blocked) {
		t.Fatalf(
			"error type = %T, want SpatialPlayerLoginClaimBlockedError",
			err,
		)
	}
}

func TestPlayerLoginClaimAllowsNonBlockingEntity(
	t *testing.T,
) {
	index := NewSpatialIndex()

	index.RegisterEntity(&Entity{
		ID:             "effect",
		Name:           "Visual Effect",
		X:              100,
		Y:              100,
		Z:              0,
		Type:           "vfx",
		BlocksMovement: false,
	})

	err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player",
			100,
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf(
			"claim over non-blocking entity: %v",
			err,
		)
	}
}

func TestPlayerLoginClaimConcurrentOverlap(
	t *testing.T,
) {
	index := NewSpatialIndex()

	const playerCount = 64

	var successes int32
	var failures int32

	var waitGroup sync.WaitGroup
	waitGroup.Add(playerCount)

	for playerNumber := 0; playerNumber < playerCount; playerNumber++ {
		playerNumber := playerNumber

		go func() {
			defer waitGroup.Done()

			err := index.TryClaimPlayerLoginEntity(
				playerLoginOverlapTestEntity(
					fmt.Sprintf(
						"player-%d",
						playerNumber,
					),
					100,
					100,
					0,
				),
			)
			if err != nil {
				atomic.AddInt32(
					&failures,
					1,
				)
				return
			}

			atomic.AddInt32(
				&successes,
				1,
			)
		}()
	}

	waitGroup.Wait()

	if successes != playerCount ||
		failures != 0 {
		t.Fatalf(
			"successes=%d failures=%d, want %d/0",
			successes,
			failures,
			playerCount,
		)
	}

	if !index.HasOverlappingPlayer(
		"player-0",
	) {
		t.Fatal(
			"concurrently claimed player did not detect overlap",
		)
	}
}

func TestPlayerOverlapClearsAfterRemoval(
	t *testing.T,
) {
	index := NewSpatialIndex()

	if err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player-a",
			100,
			100,
			0,
		),
	); err != nil {
		t.Fatalf(
			"claim player-a: %v",
			err,
		)
	}

	if err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player-b",
			100,
			100,
			0,
		),
	); err != nil {
		t.Fatalf(
			"claim player-b: %v",
			err,
		)
	}

	index.RemoveEntity("player-b")

	if index.HasOverlappingPlayer(
		"player-a",
	) {
		t.Fatal(
			"overlap remained after other player removal",
		)
	}
}

func TestPlayerLoginTileBlockedIgnoresPlayers(
	t *testing.T,
) {
	index := NewSpatialIndex()

	if err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"existing-player",
			100,
			100,
			0,
		),
	); err != nil {
		t.Fatalf(
			"claim existing player: %v",
			err,
		)
	}

	if index.IsPlayerLoginTileBlocked(
		"new-player",
		100,
		100,
		0,
	) {
		t.Fatal(
			"existing player incorrectly blocked login tile",
		)
	}

	index.RegisterEntity(&Entity{
		ID:             "npc",
		X:              100,
		Y:              100,
		Z:              0,
		Type:           "npc",
		BlocksMovement: true,
	})

	if !index.IsPlayerLoginTileBlocked(
		"new-player",
		100,
		100,
		0,
	) {
		t.Fatal(
			"blocking NPC did not block login tile",
		)
	}
}

func TestPlayerLoginClaimRejectsNonPlayerType(
	t *testing.T,
) {
	index := NewSpatialIndex()

	entity := playerLoginOverlapTestEntity(
		"creature",
		100,
		100,
		0,
	)
	entity.Type = "creature"

	err := index.TryClaimPlayerLoginEntity(entity)
	if err == nil {
		t.Fatal(
			"expected non-player validation error",
		)
	}

	var validation *SpatialSpawnClaimValidationError

	if !errors.As(err, &validation) {
		t.Fatalf(
			"error type = %T, want SpatialSpawnClaimValidationError",
			err,
		)
	}

	if validation.Field != "entity_type" {
		t.Fatalf(
			"validation field = %q, want entity_type",
			validation.Field,
		)
	}
}

func TestExclusiveClaimStillRejectsLoginPlayerTile(
	t *testing.T,
) {
	index := NewSpatialIndex()

	if err := index.TryClaimPlayerLoginEntity(
		playerLoginOverlapTestEntity(
			"player",
			100,
			100,
			0,
		),
	); err != nil {
		t.Fatalf(
			"claim login player: %v",
			err,
		)
	}

	err := index.TryClaimBlockingEntity(&Entity{
		ID:             "creature",
		X:              100,
		Y:              100,
		Z:              0,
		Type:           "creature",
		BlocksMovement: true,
	})
	if err == nil {
		t.Fatal(
			"exclusive claim unexpectedly accepted an occupied player tile",
		)
	}

	var occupied *SpatialSpawnClaimOccupiedError

	if !errors.As(err, &occupied) {
		t.Fatalf(
			"error type = %T, want SpatialSpawnClaimOccupiedError",
			err,
		)
	}
}
