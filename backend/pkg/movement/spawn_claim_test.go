package movement

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
)

func spatialSpawnClaimTestEntity(
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

func TestSpatialSpawnClaimSuccess(
	t *testing.T,
) {
	index := NewSpatialIndex()

	source := spatialSpawnClaimTestEntity(
		"player",
		100,
		101,
		2,
	)

	err := index.TryClaimBlockingEntity(source)
	if err != nil {
		t.Fatalf(
			"claim blocking entity: %v",
			err,
		)
	}

	source.X = 999
	source.Y = 999

	claimed, exists := index.GetEntity("player")
	if !exists {
		t.Fatal(
			"claimed entity was not registered",
		)
	}

	if claimed == source {
		t.Fatal(
			"SpatialIndex retained caller-owned pointer",
		)
	}

	if claimed.X != 100 ||
		claimed.Y != 101 ||
		claimed.Z != 2 ||
		!claimed.BlocksMovement {
		t.Fatalf(
			"claimed entity = %+v",
			claimed,
		)
	}

	if !index.IsTileOccupied(
		"other",
		100,
		101,
		2,
	) {
		t.Fatal(
			"claimed tile is not reported as occupied",
		)
	}
}

func TestSpatialSpawnClaimRejectsOccupiedTile(
	t *testing.T,
) {
	index := NewSpatialIndex()

	index.RegisterEntity(
		&Entity{
			ID:             "npc",
			Name:           "Blocking NPC",
			X:              100,
			Y:              100,
			Z:              0,
			Type:           "npc",
			BlocksMovement: true,
		},
	)

	err := index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
			"player",
			100,
			100,
			0,
		),
	)
	if err == nil {
		t.Fatal(
			"expected occupied tile error",
		)
	}

	var occupied *SpatialSpawnClaimOccupiedError

	if !errors.As(err, &occupied) {
		t.Fatalf(
			"error type = %T, want SpatialSpawnClaimOccupiedError",
			err,
		)
	}

	if _, exists := index.GetEntity("player"); exists {
		t.Fatal(
			"rejected player was registered",
		)
	}
}

func TestSpatialSpawnClaimAllowsNonBlockingEntityOnTile(
	t *testing.T,
) {
	index := NewSpatialIndex()

	index.RegisterEntity(
		&Entity{
			ID:             "effect",
			Name:           "Visual effect",
			X:              100,
			Y:              100,
			Z:              0,
			Type:           "vfx",
			BlocksMovement: false,
		},
	)

	err := index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
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

	if !index.IsTileOccupied(
		"other",
		100,
		100,
		0,
	) {
		t.Fatal(
			"blocking player claim is not visible",
		)
	}
}

func TestSpatialSpawnClaimRejectsDuplicateEntityID(
	t *testing.T,
) {
	index := NewSpatialIndex()

	err := index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
			"player",
			100,
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf(
			"first claim: %v",
			err,
		)
	}

	err = index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
			"player",
			101,
			100,
			0,
		),
	)
	if err == nil {
		t.Fatal(
			"expected duplicate entity ID error",
		)
	}

	var duplicate *SpatialSpawnClaimEntityExistsError

	if !errors.As(err, &duplicate) {
		t.Fatalf(
			"error type = %T, want SpatialSpawnClaimEntityExistsError",
			err,
		)
	}

	claimed, exists := index.GetEntity("player")
	if !exists ||
		claimed.X != 100 ||
		claimed.Y != 100 {
		t.Fatalf(
			"original entity changed after duplicate claim: %+v",
			claimed,
		)
	}
}

func TestSpatialSpawnClaimValidation(
	t *testing.T,
) {
	testCases := []struct {
		name   string
		entity *Entity
		field  string
	}{
		{
			name:   "nil entity",
			entity: nil,
			field:  "entity",
		},
		{
			name: "empty ID",
			entity: spatialSpawnClaimTestEntity(
				" ",
				100,
				100,
				0,
			),
			field: "entity_id",
		},
		{
			name: "non-blocking entity",
			entity: &Entity{
				ID:             "effect",
				X:              100,
				Y:              100,
				Z:              0,
				BlocksMovement: false,
			},
			field: "blocks_movement",
		},
		{
			name: "NaN X",
			entity: spatialSpawnClaimTestEntity(
				"nan",
				math.NaN(),
				100,
				0,
			),
			field: "x",
		},
		{
			name: "infinite Y",
			entity: spatialSpawnClaimTestEntity(
				"infinite",
				100,
				math.Inf(1),
				0,
			),
			field: "y",
		},
		{
			name: "fractional X",
			entity: spatialSpawnClaimTestEntity(
				"fractional",
				100.5,
				100,
				0,
			),
			field: "x",
		},
		{
			name: "positive overflow",
			entity: spatialSpawnClaimTestEntity(
				"overflow",
				float64(1<<31),
				100,
				0,
			),
			field: "x",
		},
		{
			name: "negative floor",
			entity: spatialSpawnClaimTestEntity(
				"negative-floor",
				100,
				100,
				-1,
			),
			field: "z",
		},
		{
			name: "floor overflow",
			entity: spatialSpawnClaimTestEntity(
				"floor-overflow",
				100,
				100,
				16,
			),
			field: "z",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				index := NewSpatialIndex()

				err := index.TryClaimBlockingEntity(
					testCase.entity,
				)
				if err == nil {
					t.Fatal(
						"expected validation error",
					)
				}

				var validation *SpatialSpawnClaimValidationError

				if !errors.As(err, &validation) {
					t.Fatalf(
						"error type = %T, want SpatialSpawnClaimValidationError",
						err,
					)
				}

				if validation.Field !=
					testCase.field {
					t.Fatalf(
						"validation field = %q, want %q",
						validation.Field,
						testCase.field,
					)
				}
			},
		)
	}
}

func TestSpatialSpawnClaimConcurrentSingleWinner(
	t *testing.T,
) {
	index := NewSpatialIndex()

	const contenderCount = 64

	var successes int32
	var occupiedFailures int32
	var unexpectedFailures int32

	var waitGroup sync.WaitGroup
	waitGroup.Add(contenderCount)

	for contender := 0; contender < contenderCount; contender++ {
		contender := contender

		go func() {
			defer waitGroup.Done()

			err := index.TryClaimBlockingEntity(
				spatialSpawnClaimTestEntity(
					fmtSpatialSpawnClaimID(contender),
					100,
					100,
					0,
				),
			)

			if err == nil {
				atomic.AddInt32(
					&successes,
					1,
				)
				return
			}

			var occupied *SpatialSpawnClaimOccupiedError

			if errors.As(err, &occupied) {
				atomic.AddInt32(
					&occupiedFailures,
					1,
				)
				return
			}

			atomic.AddInt32(
				&unexpectedFailures,
				1,
			)
		}()
	}

	waitGroup.Wait()

	if successes != 1 {
		t.Fatalf(
			"successful claims = %d, want 1",
			successes,
		)
	}

	if occupiedFailures !=
		contenderCount-1 {
		t.Fatalf(
			"occupied failures = %d, want %d",
			occupiedFailures,
			contenderCount-1,
		)
	}

	if unexpectedFailures != 0 {
		t.Fatalf(
			"unexpected failures = %d, want 0",
			unexpectedFailures,
		)
	}
}

func TestSpatialSpawnClaimConcurrentDifferentTiles(
	t *testing.T,
) {
	index := NewSpatialIndex()

	const contenderCount = 64

	var successes int32
	var failures int32

	var waitGroup sync.WaitGroup
	waitGroup.Add(contenderCount)

	for contender := 0; contender < contenderCount; contender++ {
		contender := contender

		go func() {
			defer waitGroup.Done()

			err := index.TryClaimBlockingEntity(
				spatialSpawnClaimTestEntity(
					fmtSpatialSpawnClaimID(contender),
					float64(100+contender),
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

	if successes != contenderCount ||
		failures != 0 {
		t.Fatalf(
			"successes=%d failures=%d, want %d/0",
			successes,
			failures,
			contenderCount,
		)
	}
}

func TestSpatialSpawnClaimIsolatedByFloor(
	t *testing.T,
) {
	index := NewSpatialIndex()

	err := index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
			"ground",
			100,
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf(
			"claim ground floor: %v",
			err,
		)
	}

	err = index.TryClaimBlockingEntity(
		spatialSpawnClaimTestEntity(
			"upper",
			100,
			100,
			1,
		),
	)
	if err != nil {
		t.Fatalf(
			"claim upper floor: %v",
			err,
		)
	}

	if !index.IsTileOccupied(
		"other",
		100,
		100,
		0,
	) {
		t.Fatal(
			"ground floor is not occupied",
		)
	}

	if !index.IsTileOccupied(
		"other",
		100,
		100,
		1,
	) {
		t.Fatal(
			"upper floor is not occupied",
		)
	}
}

func fmtSpatialSpawnClaimID(
	value int,
) string {
	return fmt.Sprintf(
		"player-%d",
		value,
	)
}
