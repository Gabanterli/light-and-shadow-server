package worldmap

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestStaticCollisionErrorsSupportErrorsIsAndAs(
	t *testing.T,
) {
	wrappedNilProvider := fmt.Errorf(
		"wrapped: %w",
		ErrNilProductionProvider,
	)

	if !errors.Is(
		wrappedNilProvider,
		ErrNilProductionProvider,
	) {
		t.Fatal(
			"errors.Is failed for ErrNilProductionProvider",
		)
	}

	position := WorldPosition{
		WorldSpaceID: WorldSpaceMainContinent,
		X:            32,
		Y:            64,
		Z:            0,
	}

	coordinate := ChunkCoordinate{
		ChunkX: 1,
		ChunkY: 2,
		Z:      0,
	}

	testCases := []struct {
		name   string
		err    error
		assert func(t *testing.T, err error)
	}{
		{
			name: "duplicate chunk document",
			err: &DuplicateChunkDocumentError{
				WorldSpaceID: WorldSpaceMainContinent,
				Coordinate:   coordinate,
			},
			assert: func(t *testing.T, err error) {
				var target *DuplicateChunkDocumentError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "unpublished chunk document",
			err: &UnpublishedChunkDocumentError{
				WorldSpaceID: WorldSpaceMainContinent,
				Coordinate:   coordinate,
			},
			assert: func(t *testing.T, err error) {
				var target *UnpublishedChunkDocumentError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "chunk not loaded",
			err: &StaticCollisionChunkNotLoadedError{
				WorldSpaceID: WorldSpaceMainContinent,
				Coordinate:   coordinate,
			},
			assert: func(t *testing.T, err error) {
				var target *StaticCollisionChunkNotLoadedError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "position out of bounds",
			err: &WorldPositionOutOfBoundsError{
				Position: position,
				Bounds: WorldBounds{
					MinX: 0,
					MinY: 0,
					MaxX: 31,
					MaxY: 31,
					MinZ: 0,
					MaxZ: 0,
				},
			},
			assert: func(t *testing.T, err error) {
				var target *WorldPositionOutOfBoundsError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "invalid movement step",
			err: &InvalidStaticMovementStepError{
				From:   position,
				To:     position,
				Reason: "test reason",
			},
			assert: func(t *testing.T, err error) {
				var target *InvalidStaticMovementStepError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "movement blocked",
			err: &StaticMovementBlockedError{
				Position: position,
			},
			assert: func(t *testing.T, err error) {
				var target *StaticMovementBlockedError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
		{
			name: "tile data",
			err: &StaticCollisionTileDataError{
				Position: position,
				Reason:   "test reason",
			},
			assert: func(t *testing.T, err error) {
				var target *StaticCollisionTileDataError
				if !errors.As(err, &target) {
					t.Fatalf(
						"errors.As failed for %T",
						err,
					)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			wrapped := fmt.Errorf(
				"wrapped: %w",
				testCase.err,
			)

			testCase.assert(t, wrapped)

			message := testCase.err.Error()

			if !strings.Contains(
				message,
				string(WorldSpaceMainContinent),
			) {
				t.Fatalf(
					"error message %q lacks world-space context",
					message,
				)
			}
		})
	}
}
