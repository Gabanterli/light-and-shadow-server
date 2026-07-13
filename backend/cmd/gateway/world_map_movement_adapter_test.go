package main

import (
	"errors"
	"math"
	"testing"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

func canonicalWorldMapMovementValidatorForTest(
	t *testing.T,
) movement.StaticStepValidator {
	t.Helper()

	manifestPath, residencyPath :=
		canonicalWorldMapRuntimePathsForTest(t)

	state, err := initializeWorldMapRuntime(
		worldmap.ModeProduction,
		manifestPath,
		residencyPath,
	)
	if err != nil {
		t.Fatalf(
			"initialize canonical world-map runtime: %v",
			err,
		)
	}

	validator := newWorldMapStaticStepValidator(
		state.staticCollisionIndex,
	)
	if validator == nil {
		t.Fatal("validator = nil, want production validator")
	}

	return validator
}

func TestNewWorldMapStaticStepValidatorNilIndex(
	t *testing.T,
) {
	if validator := newWorldMapStaticStepValidator(
		nil,
	); validator != nil {
		t.Fatalf(
			"validator = %#v, want nil for debug runtime",
			validator,
		)
	}
}

func TestWorldMapStaticStepValidatorCanonicalStep(
	t *testing.T,
) {
	validator :=
		canonicalWorldMapMovementValidatorForTest(t)

	err := validator.ValidateStaticStep(
		movement.StaticStep{
			WorldSpaceID: string(
				worldmap.WorldSpaceMainContinent,
			),
			FromX: 100,
			FromY: 100,
			FromZ: 0,
			ToX:   101,
			ToY:   100,
			ToZ:   0,
		},
	)
	if err != nil {
		t.Fatalf(
			"canonical static step rejected: %v",
			err,
		)
	}
}

func TestWorldMapStaticStepValidatorRejectsMalformedCoordinates(
	t *testing.T,
) {
	validator :=
		canonicalWorldMapMovementValidatorForTest(t)

	testCases := []struct {
		name          string
		value         float64
		expectedField string
		expectedText  string
	}{
		{
			name:          "fractional",
			value:         100.5,
			expectedField: "to_x",
			expectedText:  "exact tile",
		},
		{
			name:          "not a number",
			value:         math.NaN(),
			expectedField: "to_x",
			expectedText:  "finite",
		},
		{
			name:          "positive infinity",
			value:         math.Inf(1),
			expectedField: "to_x",
			expectedText:  "finite",
		},
		{
			name:          "positive signed 32-bit overflow",
			value:         float64(1 << 31),
			expectedField: "to_x",
			expectedText:  "signed 32-bit",
		},
		{
			name:          "negative signed 32-bit overflow",
			value:         float64((-1 << 31) - 1),
			expectedField: "to_x",
			expectedText:  "signed 32-bit",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateStaticStep(
				movement.StaticStep{
					WorldSpaceID: string(
						worldmap.WorldSpaceMainContinent,
					),
					FromX: 100,
					FromY: 100,
					FromZ: 0,
					ToX:   testCase.value,
					ToY:   100,
					ToZ:   0,
				},
			)

			var coordinateError *worldMapMovementCoordinateError

			if !errors.As(err, &coordinateError) {
				t.Fatalf(
					"error = %v, want worldMapMovementCoordinateError",
					err,
				)
			}

			if coordinateError.Field !=
				testCase.expectedField {
				t.Fatalf(
					"field = %q, want %q",
					coordinateError.Field,
					testCase.expectedField,
				)
			}

			if !containsText(
				coordinateError.Error(),
				testCase.expectedText,
			) {
				t.Fatalf(
					"error = %q, want text %q",
					coordinateError,
					testCase.expectedText,
				)
			}
		})
	}
}

func TestWorldMapStaticStepValidatorRejectsMalformedWorldSpace(
	t *testing.T,
) {
	validator :=
		canonicalWorldMapMovementValidatorForTest(t)

	for _, worldSpaceID := range []string{
		"",
		"   ",
		" main_continent",
		"main_continent ",
	} {
		t.Run(worldSpaceID, func(t *testing.T) {
			err := validator.ValidateStaticStep(
				movement.StaticStep{
					WorldSpaceID: worldSpaceID,
					FromX:        100,
					FromY:        100,
					FromZ:        0,
					ToX:          101,
					ToY:          100,
					ToZ:          0,
				},
			)

			var worldSpaceError *worldMapMovementWorldSpaceError

			if !errors.As(err, &worldSpaceError) {
				t.Fatalf(
					"error = %v, want worldMapMovementWorldSpaceError",
					err,
				)
			}
		})
	}
}

func TestWorldMapStaticStepValidatorPropagatesWorldMapErrors(
	t *testing.T,
) {
	validator :=
		canonicalWorldMapMovementValidatorForTest(t)

	err := validator.ValidateStaticStep(
		movement.StaticStep{
			WorldSpaceID: string(
				worldmap.WorldSpaceMainContinent,
			),
			FromX: 127,
			FromY: 127,
			FromZ: 0,
			ToX:   128,
			ToY:   128,
			ToZ:   0,
		},
	)

	var notPublished *worldmap.ChunkReferenceNotFoundError

	if !errors.As(err, &notPublished) {
		t.Fatalf(
			"error = %v, want ChunkReferenceNotFoundError",
			err,
		)
	}
}

func TestProductionMovementUsesWorldMapStaticStepValidator(
	t *testing.T,
) {
	validator :=
		canonicalWorldMapMovementValidatorForTest(t)

	spatialIndex := movement.NewSpatialIndex()
	chunkManager := movement.NewChunkManager()
	aoiManager := movement.NewAOIManager(spatialIndex)

	system :=
		movement.NewMovementSystemWithStaticStepValidator(
			spatialIndex,
			chunkManager,
			aoiManager,
			validator,
		)

	system.InitPlayerStateInWorldSpace(
		"allowed-player",
		string(worldmap.WorldSpaceMainContinent),
		100,
		100,
		0,
	)

	success, confirmedX, confirmedY, confirmedZ :=
		system.ValidateAndMove(
			"allowed-player",
			101,
			100,
			0,
			1,
		)

	if !success {
		t.Fatal("canonical production movement rejected")
	}

	if confirmedX != 101 ||
		confirmedY != 100 ||
		confirmedZ != 0 {
		t.Fatalf(
			"allowed confirmation = (%v,%v,%d), want (101,100,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}

	system.InitPlayerStateInWorldSpace(
		"unpublished-player",
		string(worldmap.WorldSpaceMainContinent),
		127,
		127,
		0,
	)

	success, confirmedX, confirmedY, confirmedZ =
		system.ValidateAndMove(
			"unpublished-player",
			128,
			128,
			0,
			1,
		)

	if success {
		t.Fatal("movement into unpublished chunk succeeded")
	}

	if confirmedX != 127 ||
		confirmedY != 127 ||
		confirmedZ != 0 {
		t.Fatalf(
			"rejected confirmation = (%v,%v,%d), want (127,127,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}

	entity, exists := spatialIndex.GetEntity(
		"unpublished-player",
	)
	if !exists {
		t.Fatal("unpublished-player missing from spatial index")
	}

	if entity.X != 127 ||
		entity.Y != 127 ||
		entity.Z != 0 {
		t.Fatalf(
			"rejected spatial position = (%v,%v,%d), want (127,127,0)",
			entity.X,
			entity.Y,
			entity.Z,
		)
	}
}

func containsText(value, expected string) bool {
	if expected == "" {
		return true
	}

	for index := 0; index+len(expected) <= len(value); index++ {
		if value[index:index+len(expected)] == expected {
			return true
		}
	}

	return false
}
