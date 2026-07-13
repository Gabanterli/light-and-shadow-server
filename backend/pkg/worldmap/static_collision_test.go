package worldmap

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func staticCollisionReferenceForTest(
	chunkX int,
	chunkY int,
	z int,
) ChunkReference {
	return ChunkReference{
		ChunkX: chunkX,
		ChunkY: chunkY,
		Z:      z,
		File: fmt.Sprintf(
			"chunks/%d_%d_%d.json",
			chunkX,
			chunkY,
			z,
		),
	}
}

func staticCollisionProviderForTest(
	t *testing.T,
	references []ChunkReference,
) *ProductionProvider {
	t.Helper()

	canonicalSnapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	rootManifest := canonicalSnapshot.RootManifest()
	worldSpaceIDs := canonicalSnapshot.WorldSpaceIDs()

	worldSpaces := make(
		map[WorldSpaceID]WorldSpaceManifest,
		len(worldSpaceIDs),
	)

	for _, worldSpaceID := range worldSpaceIDs {
		manifest, found :=
			canonicalSnapshot.WorldSpace(worldSpaceID)
		if !found {
			t.Fatalf(
				"canonical world space %q not found",
				worldSpaceID,
			)
		}

		worldSpaces[worldSpaceID] = manifest
	}

	mainContinent :=
		worldSpaces[WorldSpaceMainContinent]

	mainContinent.Chunks =
		append([]ChunkReference(nil), references...)

	worldSpaces[WorldSpaceMainContinent] =
		mainContinent

	snapshot := &ManifestSnapshot{
		rootManifest:  rootManifest,
		worldSpaces:   worldSpaces,
		worldSpaceIDs: worldSpaceIDs,
	}

	return newProductionProviderForTest(t, snapshot)
}

func staticCollisionDocumentForTest(
	provider *ProductionProvider,
	coordinate ChunkCoordinate,
	blockedLocalTiles ...[2]int,
) ChunkDocument {
	tiles := make(
		[]uint16,
		CanonicalChunkSize*CanonicalChunkSize,
	)

	for _, localTile := range blockedLocalTiles {
		localX := localTile[0]
		localY := localTile[1]

		if localX < 0 ||
			localX >= CanonicalChunkSize ||
			localY < 0 ||
			localY >= CanonicalChunkSize {
			panic(
				fmt.Sprintf(
					"invalid test local tile (%d,%d)",
					localX,
					localY,
				),
			)
		}

		tiles[localY*CanonicalChunkSize+localX] = 1
	}

	return ChunkDocument{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       provider.WorldID(),
		WorldVersion:  provider.Version(),
		WorldSpaceID:  WorldSpaceMainContinent,
		ChunkX:        coordinate.ChunkX,
		ChunkY:        coordinate.ChunkY,
		Z:             coordinate.Z,
		Width:         CanonicalChunkSize,
		Height:        CanonicalChunkSize,
		TilePalette: []TileDefinition{
			{
				VisualID:       1,
				Walkable:       true,
				MovementCost:   100,
				BlocksMovement: false,
			},
			{
				VisualID:       2,
				Walkable:       false,
				MovementCost:   0,
				BlocksMovement: true,
			},
		},
		Tiles: tiles,
	}
}

func staticCollisionDocumentsForTest(
	provider *ProductionProvider,
	references []ChunkReference,
	blocked map[WorldPosition]bool,
	omitted map[ChunkCoordinate]bool,
) []ChunkDocument {
	documents := make(
		[]ChunkDocument,
		0,
		len(references),
	)

	for _, reference := range references {
		coordinate := ChunkCoordinate{
			ChunkX: reference.ChunkX,
			ChunkY: reference.ChunkY,
			Z:      reference.Z,
		}

		if omitted[coordinate] {
			continue
		}

		blockedLocals := make([][2]int, 0)

		for position, isBlocked := range blocked {
			if !isBlocked ||
				position.WorldSpaceID !=
					WorldSpaceMainContinent ||
				position.Z != coordinate.Z {
				continue
			}

			positionChunkX :=
				position.X / CanonicalChunkSize
			positionChunkY :=
				position.Y / CanonicalChunkSize

			if positionChunkX != coordinate.ChunkX ||
				positionChunkY != coordinate.ChunkY {
				continue
			}

			blockedLocals = append(
				blockedLocals,
				[2]int{
					position.X % CanonicalChunkSize,
					position.Y % CanonicalChunkSize,
				},
			)
		}

		documents = append(
			documents,
			staticCollisionDocumentForTest(
				provider,
				coordinate,
				blockedLocals...,
			),
		)
	}

	return documents
}

func staticCollisionIndexForTest(
	t *testing.T,
	references []ChunkReference,
	blocked map[WorldPosition]bool,
	omitted map[ChunkCoordinate]bool,
) (*StaticCollisionIndex, *ProductionProvider) {
	t.Helper()

	provider := staticCollisionProviderForTest(
		t,
		references,
	)

	documents := staticCollisionDocumentsForTest(
		provider,
		references,
		blocked,
		omitted,
	)

	index, err := NewStaticCollisionIndex(
		provider,
		documents,
	)
	if err != nil {
		t.Fatalf(
			"NewStaticCollisionIndex failed: %v",
			err,
		)
	}

	return index, provider
}

func staticCollisionPositionForTest(
	x int,
	y int,
	z int,
) WorldPosition {
	return WorldPosition{
		WorldSpaceID: WorldSpaceMainContinent,
		X:            x,
		Y:            y,
		Z:            z,
	}
}

func TestNewStaticCollisionIndexConstruction(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
	}

	provider := staticCollisionProviderForTest(
		t,
		references,
	)

	validDocument := staticCollisionDocumentForTest(
		provider,
		ChunkCoordinate{
			ChunkX: 0,
			ChunkY: 0,
			Z:      0,
		},
	)

	t.Run("rejects nil provider", func(t *testing.T) {
		index, err := NewStaticCollisionIndex(nil, nil)

		if index != nil {
			t.Fatal("index is not nil")
		}

		if !errors.Is(err, ErrNilProductionProvider) {
			t.Fatalf(
				"error = %v, want ErrNilProductionProvider",
				err,
			)
		}
	})

	t.Run("accepts zero documents", func(t *testing.T) {
		index, err := NewStaticCollisionIndex(
			provider,
			nil,
		)
		if err != nil {
			t.Fatalf(
				"NewStaticCollisionIndex failed: %v",
				err,
			)
		}

		if got := len(index.loadedCollisions); got != 0 {
			t.Fatalf(
				"loaded collision count = %d, want 0",
				got,
			)
		}
	})

	t.Run("accepts valid document", func(t *testing.T) {
		index, err := NewStaticCollisionIndex(
			provider,
			[]ChunkDocument{validDocument},
		)
		if err != nil {
			t.Fatalf(
				"NewStaticCollisionIndex failed: %v",
				err,
			)
		}

		if got := len(index.loadedCollisions); got != 1 {
			t.Fatalf(
				"loaded collision count = %d, want 1",
				got,
			)
		}
	})

	t.Run(
		"rejects structurally invalid document",
		func(t *testing.T) {
			document := validDocument
			document.TilePalette = nil

			index, err := NewStaticCollisionIndex(
				provider,
				[]ChunkDocument{document},
			)

			if index != nil {
				t.Fatal("index is not nil")
			}

			if err == nil {
				t.Fatal("error is nil")
			}
		},
	)

	t.Run("rejects duplicate document", func(t *testing.T) {
		index, err := NewStaticCollisionIndex(
			provider,
			[]ChunkDocument{
				validDocument,
				validDocument,
			},
		)

		if index != nil {
			t.Fatal("index is not nil")
		}

		var target *DuplicateChunkDocumentError
		if !errors.As(err, &target) {
			t.Fatalf(
				"error type = %T, want DuplicateChunkDocumentError",
				err,
			)
		}
	})

	t.Run(
		"rejects missing world space",
		func(t *testing.T) {
			document := validDocument
			document.WorldSpaceID =
				WorldSpaceID("missing_world_space")

			index, err := NewStaticCollisionIndex(
				provider,
				[]ChunkDocument{document},
			)

			if index != nil {
				t.Fatal("index is not nil")
			}

			var target *WorldSpaceNotFoundError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want WorldSpaceNotFoundError",
					err,
				)
			}
		},
	)

	t.Run(
		"rejects unpublished document",
		func(t *testing.T) {
			document :=
				staticCollisionDocumentForTest(
					provider,
					ChunkCoordinate{
						ChunkX: 1,
						ChunkY: 0,
						Z:      0,
					},
				)

			index, err := NewStaticCollisionIndex(
				provider,
				[]ChunkDocument{document},
			)

			if index != nil {
				t.Fatal("index is not nil")
			}

			var target *UnpublishedChunkDocumentError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want UnpublishedChunkDocumentError",
					err,
				)
			}
		},
	)

	t.Run(
		"rejects inconsistent world id",
		func(t *testing.T) {
			document := validDocument
			document.WorldID = "different_world"

			index, err := NewStaticCollisionIndex(
				provider,
				[]ChunkDocument{document},
			)

			if index != nil {
				t.Fatal("index is not nil")
			}

			if err == nil {
				t.Fatal("error is nil")
			}
		},
	)

	t.Run(
		"rejects inconsistent world version",
		func(t *testing.T) {
			document := validDocument
			document.WorldVersion++

			index, err := NewStaticCollisionIndex(
				provider,
				[]ChunkDocument{document},
			)

			if index != nil {
				t.Fatal("index is not nil")
			}

			if err == nil {
				t.Fatal("error is nil")
			}
		},
	)
}

func TestStaticCollisionIndexTileCollisionAndCanOccupy(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
		staticCollisionReferenceForTest(1, 0, 0),
		staticCollisionReferenceForTest(0, 1, 0),
		staticCollisionReferenceForTest(1, 1, 0),
		staticCollisionReferenceForTest(2, 0, 0),
	}

	blockedPosition :=
		staticCollisionPositionForTest(5, 6, 0)

	omittedCoordinate := ChunkCoordinate{
		ChunkX: 2,
		ChunkY: 0,
		Z:      0,
	}

	index, provider := staticCollisionIndexForTest(
		t,
		references,
		map[WorldPosition]bool{
			blockedPosition: true,
		},
		map[ChunkCoordinate]bool{
			omittedCoordinate: true,
		},
	)

	t.Run("returns free tile", func(t *testing.T) {
		position :=
			staticCollisionPositionForTest(4, 6, 0)

		collision, err := index.TileCollision(position)
		if err != nil {
			t.Fatalf(
				"TileCollision failed: %v",
				err,
			)
		}

		if collision.BlocksMovement {
			t.Fatal("free tile blocks movement")
		}

		canOccupy, err := index.CanOccupy(position)
		if err != nil {
			t.Fatalf(
				"CanOccupy failed: %v",
				err,
			)
		}

		if !canOccupy {
			t.Fatal("free tile cannot be occupied")
		}
	})

	t.Run("returns blocked tile", func(t *testing.T) {
		collision, err :=
			index.TileCollision(blockedPosition)
		if err != nil {
			t.Fatalf(
				"TileCollision failed: %v",
				err,
			)
		}

		if !collision.BlocksMovement {
			t.Fatal("blocked tile does not block movement")
		}

		canOccupy, err :=
			index.CanOccupy(blockedPosition)
		if err != nil {
			t.Fatalf(
				"CanOccupy failed: %v",
				err,
			)
		}

		if canOccupy {
			t.Fatal("blocked tile can be occupied")
		}
	})

	t.Run(
		"resolves chunk coordinate boundaries",
		func(t *testing.T) {
			positions := []WorldPosition{
				staticCollisionPositionForTest(
					0,
					0,
					0,
				),
				staticCollisionPositionForTest(
					31,
					31,
					0,
				),
				staticCollisionPositionForTest(
					32,
					0,
					0,
				),
				staticCollisionPositionForTest(
					0,
					32,
					0,
				),
				staticCollisionPositionForTest(
					32,
					32,
					0,
				),
			}

			for _, position := range positions {
				canOccupy, err :=
					index.CanOccupy(position)
				if err != nil {
					t.Fatalf(
						"CanOccupy(%+v) failed: %v",
						position,
						err,
					)
				}

				if !canOccupy {
					t.Fatalf(
						"CanOccupy(%+v) = false, want true",
						position,
					)
				}
			}
		},
	)

	t.Run(
		"distinguishes published but not loaded chunk",
		func(t *testing.T) {
			position :=
				staticCollisionPositionForTest(
					2*CanonicalChunkSize,
					0,
					0,
				)

			canOccupy, err :=
				index.CanOccupy(position)

			if canOccupy {
				t.Fatal(
					"position in unloaded chunk can be occupied",
				)
			}

			var target *StaticCollisionChunkNotLoadedError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want StaticCollisionChunkNotLoadedError",
					err,
				)
			}
		},
	)

	t.Run(
		"distinguishes unpublished chunk",
		func(t *testing.T) {
			position :=
				staticCollisionPositionForTest(
					3*CanonicalChunkSize,
					0,
					0,
				)

			canOccupy, err :=
				index.CanOccupy(position)

			if canOccupy {
				t.Fatal(
					"position in unpublished chunk can be occupied",
				)
			}

			var target *ChunkReferenceNotFoundError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want ChunkReferenceNotFoundError",
					err,
				)
			}
		},
	)

	t.Run("rejects unknown world space", func(t *testing.T) {
		position := WorldPosition{
			WorldSpaceID: WorldSpaceID("missing_world_space"),
			X:            0,
			Y:            0,
			Z:            0,
		}

		_, err := index.CanOccupy(position)

		var target *WorldSpaceNotFoundError
		if !errors.As(err, &target) {
			t.Fatalf(
				"error type = %T, want WorldSpaceNotFoundError",
				err,
			)
		}
	})

	t.Run("rejects every bounds edge", func(t *testing.T) {
		bounds, err :=
			provider.Bounds(WorldSpaceMainContinent)
		if err != nil {
			t.Fatalf("Bounds failed: %v", err)
		}

		positions := []WorldPosition{
			staticCollisionPositionForTest(
				bounds.MinX-1,
				bounds.MinY,
				0,
			),
			staticCollisionPositionForTest(
				bounds.MinX,
				bounds.MinY-1,
				0,
			),
			staticCollisionPositionForTest(
				bounds.MaxX+1,
				bounds.MinY,
				0,
			),
			staticCollisionPositionForTest(
				bounds.MinX,
				bounds.MaxY+1,
				0,
			),
			staticCollisionPositionForTest(
				bounds.MinX,
				bounds.MinY,
				bounds.MinZ-1,
			),
			staticCollisionPositionForTest(
				bounds.MinX,
				bounds.MinY,
				bounds.MaxZ+1,
			),
		}

		for _, position := range positions {
			canOccupy, err :=
				index.CanOccupy(position)

			if canOccupy {
				t.Fatalf(
					"out-of-bounds position %+v can be occupied",
					position,
				)
			}

			var target *WorldPositionOutOfBoundsError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type for %+v = %T, want WorldPositionOutOfBoundsError",
					position,
					err,
				)
			}
		}
	})
}

func TestStaticCollisionIndexValidateStepCardinalAndInvalid(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
	}

	blockedDestination :=
		staticCollisionPositionForTest(12, 10, 0)

	index, provider := staticCollisionIndexForTest(
		t,
		references,
		map[WorldPosition]bool{
			blockedDestination: true,
		},
		nil,
	)

	origin :=
		staticCollisionPositionForTest(10, 10, 0)

	validDestinations := []WorldPosition{
		staticCollisionPositionForTest(10, 9, 0),
		staticCollisionPositionForTest(10, 11, 0),
		staticCollisionPositionForTest(9, 10, 0),
		staticCollisionPositionForTest(11, 10, 0),
	}

	for _, destination := range validDestinations {
		if err := index.ValidateStep(
			origin,
			destination,
		); err != nil {
			t.Fatalf(
				"ValidateStep(%+v, %+v) failed: %v",
				origin,
				destination,
				err,
			)
		}
	}

	blockedFrom :=
		staticCollisionPositionForTest(11, 10, 0)

	err := index.ValidateStep(
		blockedFrom,
		blockedDestination,
	)

	var blockedError *StaticMovementBlockedError
	if !errors.As(err, &blockedError) {
		t.Fatalf(
			"error type = %T, want StaticMovementBlockedError",
			err,
		)
	}

	invalidSteps := []struct {
		name string
		from WorldPosition
		to   WorldPosition
	}{
		{
			name: "equal origin and destination",
			from: origin,
			to:   origin,
		},
		{
			name: "two tiles x",
			from: origin,
			to: staticCollisionPositionForTest(
				12,
				10,
				0,
			),
		},
		{
			name: "two tiles y",
			from: origin,
			to: staticCollisionPositionForTest(
				10,
				12,
				0,
			),
		},
		{
			name: "two by two",
			from: origin,
			to: staticCollisionPositionForTest(
				12,
				12,
				0,
			),
		},
		{
			name: "floor change",
			from: origin,
			to: staticCollisionPositionForTest(
				11,
				10,
				1,
			),
		},
		{
			name: "world space change",
			from: origin,
			to: WorldPosition{
				WorldSpaceID: WorldSpaceID("other_world_space"),
				X:            11,
				Y:            10,
				Z:            0,
			},
		},
	}

	for _, testCase := range invalidSteps {
		t.Run(testCase.name, func(t *testing.T) {
			err := index.ValidateStep(
				testCase.from,
				testCase.to,
			)

			var target *InvalidStaticMovementStepError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want InvalidStaticMovementStepError",
					err,
				)
			}
		})
	}

	bounds, err :=
		provider.Bounds(WorldSpaceMainContinent)
	if err != nil {
		t.Fatalf("Bounds failed: %v", err)
	}

	outOfBoundsOrigin :=
		staticCollisionPositionForTest(
			bounds.MinX-1,
			bounds.MinY,
			0,
		)

	err = index.ValidateStep(
		outOfBoundsOrigin,
		staticCollisionPositionForTest(
			bounds.MinX,
			bounds.MinY,
			0,
		),
	)

	var boundsError *WorldPositionOutOfBoundsError
	if !errors.As(err, &boundsError) {
		t.Fatalf(
			"error type = %T, want WorldPositionOutOfBoundsError",
			err,
		)
	}
}

func TestStaticCollisionIndexValidateStepDiagonalRules(
	t *testing.T,
) {
	singleChunkReferences := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
	}

	from :=
		staticCollisionPositionForTest(10, 10, 0)
	to :=
		staticCollisionPositionForTest(11, 11, 0)
	firstOrthogonal :=
		staticCollisionPositionForTest(11, 10, 0)
	secondOrthogonal :=
		staticCollisionPositionForTest(10, 11, 0)

	t.Run("allows fully open diagonal", func(t *testing.T) {
		index, _ := staticCollisionIndexForTest(
			t,
			singleChunkReferences,
			nil,
			nil,
		)

		if err := index.ValidateStep(from, to); err != nil {
			t.Fatalf(
				"ValidateStep failed: %v",
				err,
			)
		}
	})

	blockedCases := []struct {
		name     string
		position WorldPosition
	}{
		{
			name:     "blocked destination",
			position: to,
		},
		{
			name:     "blocked first orthogonal",
			position: firstOrthogonal,
		},
		{
			name:     "blocked second orthogonal",
			position: secondOrthogonal,
		},
	}

	for _, testCase := range blockedCases {
		t.Run(testCase.name, func(t *testing.T) {
			index, _ := staticCollisionIndexForTest(
				t,
				singleChunkReferences,
				map[WorldPosition]bool{
					testCase.position: true,
				},
				nil,
			)

			err := index.ValidateStep(from, to)

			var target *StaticMovementBlockedError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want StaticMovementBlockedError",
					err,
				)
			}

			if target.Position != testCase.position {
				t.Fatalf(
					"blocked position = %+v, want %+v",
					target.Position,
					testCase.position,
				)
			}
		})
	}

	t.Run("blocks when both orthogonals block", func(t *testing.T) {
		index, _ := staticCollisionIndexForTest(
			t,
			singleChunkReferences,
			map[WorldPosition]bool{
				firstOrthogonal:  true,
				secondOrthogonal: true,
			},
			nil,
		)

		err := index.ValidateStep(from, to)

		var target *StaticMovementBlockedError
		if !errors.As(err, &target) {
			t.Fatalf(
				"error type = %T, want StaticMovementBlockedError",
				err,
			)
		}
	})

	boundaryReferences := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
		staticCollisionReferenceForTest(1, 0, 0),
		staticCollisionReferenceForTest(0, 1, 0),
		staticCollisionReferenceForTest(1, 1, 0),
	}

	boundaryFrom :=
		staticCollisionPositionForTest(31, 31, 0)
	boundaryTo :=
		staticCollisionPositionForTest(32, 32, 0)

	t.Run(
		"allows diagonal across four chunk boundary",
		func(t *testing.T) {
			index, _ := staticCollisionIndexForTest(
				t,
				boundaryReferences,
				nil,
				nil,
			)

			if err := index.ValidateStep(
				boundaryFrom,
				boundaryTo,
			); err != nil {
				t.Fatalf(
					"ValidateStep failed: %v",
					err,
				)
			}
		},
	)

	t.Run(
		"preserves unloaded orthogonal chunk error",
		func(t *testing.T) {
			omittedCoordinate := ChunkCoordinate{
				ChunkX: 1,
				ChunkY: 0,
				Z:      0,
			}

			index, _ := staticCollisionIndexForTest(
				t,
				boundaryReferences,
				nil,
				map[ChunkCoordinate]bool{
					omittedCoordinate: true,
				},
			)

			err := index.ValidateStep(
				boundaryFrom,
				boundaryTo,
			)

			var target *StaticCollisionChunkNotLoadedError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want StaticCollisionChunkNotLoadedError",
					err,
				)
			}

			if target.Coordinate != omittedCoordinate {
				t.Fatalf(
					"missing coordinate = %+v, want %+v",
					target.Coordinate,
					omittedCoordinate,
				)
			}
		},
	)
}

func TestStaticCollisionIndexOwnsIndependentCopies(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
	}

	provider := staticCollisionProviderForTest(
		t,
		references,
	)

	blockedLocal := [2]int{5, 6}

	document := staticCollisionDocumentForTest(
		provider,
		ChunkCoordinate{
			ChunkX: 0,
			ChunkY: 0,
			Z:      0,
		},
		blockedLocal,
	)

	documents := []ChunkDocument{document}

	index, err := NewStaticCollisionIndex(
		provider,
		documents,
	)
	if err != nil {
		t.Fatalf(
			"NewStaticCollisionIndex failed: %v",
			err,
		)
	}

	position :=
		staticCollisionPositionForTest(5, 6, 0)

	documents[0].WorldID = "mutated"
	documents[0].WorldVersion = 999
	documents[0].Tiles[6*CanonicalChunkSize+5] = 0
	documents[0].TilePalette[1].BlocksMovement = false
	documents[0].TilePalette = nil
	documents[0].Tiles = nil

	collision, err := index.TileCollision(position)
	if err != nil {
		t.Fatalf(
			"TileCollision failed after source mutation: %v",
			err,
		)
	}

	if !collision.BlocksMovement {
		t.Fatal(
			"collision index changed after source mutation",
		)
	}
}

func TestStaticCollisionIndexRemainsSparse(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
		staticCollisionReferenceForTest(100, 100, 0),
		staticCollisionReferenceForTest(200, 200, 0),
	}

	omittedCoordinate := ChunkCoordinate{
		ChunkX: 200,
		ChunkY: 200,
		Z:      0,
	}

	index, provider := staticCollisionIndexForTest(
		t,
		references,
		nil,
		map[ChunkCoordinate]bool{
			omittedCoordinate: true,
		},
	)

	if got := len(index.loadedCollisions); got != 2 {
		t.Fatalf(
			"loaded collision count = %d, want 2",
			got,
		)
	}

	published, err := provider.ChunkReferences(
		WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf(
			"ChunkReferences failed: %v",
			err,
		)
	}

	if got := len(published); got != 3 {
		t.Fatalf(
			"provider published chunk count = %d, want 3",
			got,
		)
	}
}

func TestStaticCollisionIndexConcurrentReads(
	t *testing.T,
) {
	references := []ChunkReference{
		staticCollisionReferenceForTest(0, 0, 0),
	}

	index, _ := staticCollisionIndexForTest(
		t,
		references,
		nil,
		nil,
	)

	from :=
		staticCollisionPositionForTest(10, 10, 0)
	to :=
		staticCollisionPositionForTest(11, 11, 0)

	var waitGroup sync.WaitGroup

	for worker := 0; worker < 64; worker++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			for iteration := 0; iteration < 100; iteration++ {
				if _, err := index.TileCollision(from); err != nil {
					t.Errorf(
						"TileCollision failed: %v",
						err,
					)
					return
				}

				if _, err := index.CanOccupy(to); err != nil {
					t.Errorf(
						"CanOccupy failed: %v",
						err,
					)
					return
				}

				if err := index.ValidateStep(
					from,
					to,
				); err != nil {
					t.Errorf(
						"ValidateStep failed: %v",
						err,
					)
					return
				}
			}
		}()
	}

	waitGroup.Wait()
}
