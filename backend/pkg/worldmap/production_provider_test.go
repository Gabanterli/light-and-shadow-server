package worldmap

import (
	"errors"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"
)

func canonicalManifestPathForProductionProviderTest(
	t *testing.T,
) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	backendRoot := filepath.Clean(
		filepath.Join(filepath.Dir(currentFile), "..", ".."),
	)

	return filepath.Join(
		backendRoot,
		"config",
		"worldmap",
		"world_manifest.json",
	)
}

func loadCanonicalSnapshotForProductionProviderTest(
	t *testing.T,
) *ManifestSnapshot {
	t.Helper()

	snapshot, err := LoadManifestSnapshot(
		canonicalManifestPathForProductionProviderTest(t),
	)
	if err != nil {
		t.Fatalf("LoadManifestSnapshot failed: %v", err)
	}

	return snapshot
}

func newProductionProviderForTest(
	t *testing.T,
	snapshot *ManifestSnapshot,
) *ProductionProvider {
	t.Helper()

	provider, err := NewProductionProvider(snapshot)
	if err != nil {
		t.Fatalf("NewProductionProvider failed: %v", err)
	}

	return provider
}

func snapshotWithPublishedChunksForProviderTest(
	t *testing.T,
) *ManifestSnapshot {
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
				"canonical world space %q was not found",
				worldSpaceID,
			)
		}

		worldSpaces[worldSpaceID] = manifest
	}

	mainContinent :=
		worldSpaces[WorldSpaceMainContinent]

	mainContinent.Chunks = []ChunkReference{
		{
			ChunkX:      1,
			ChunkY:      2,
			Z:           0,
			File:        "chunks/1_2_0.json",
			ContentHash: "hash-floor-zero",
		},
		{
			ChunkX:      1,
			ChunkY:      2,
			Z:           1,
			File:        "chunks/1_2_1.json",
			ContentHash: "hash-floor-one",
		},
		{
			ChunkX: 7,
			ChunkY: 8,
			Z:      0,
			File:   "chunks/7_8_0.json",
		},
	}

	worldSpaces[WorldSpaceMainContinent] =
		mainContinent

	return &ManifestSnapshot{
		rootManifest:  rootManifest,
		worldSpaces:   worldSpaces,
		worldSpaceIDs: worldSpaceIDs,
	}
}

func TestProductionProviderImplementsProvider(t *testing.T) {
	var _ Provider = (*ProductionProvider)(nil)
}

func TestNewProductionProviderRejectsNilSnapshot(
	t *testing.T,
) {
	provider, err := NewProductionProvider(nil)

	if provider != nil {
		t.Fatal("provider is not nil after nil snapshot")
	}

	if !errors.Is(err, ErrNilManifestSnapshot) {
		t.Fatalf(
			"error = %v, want ErrNilManifestSnapshot",
			err,
		)
	}
}

func TestNewProductionProviderRejectsMissingWorldSpace(
	t *testing.T,
) {
	snapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	missingID := snapshot.worldSpaceIDs[0]
	delete(snapshot.worldSpaces, missingID)

	provider, err := NewProductionProvider(snapshot)

	if provider != nil {
		t.Fatal(
			"provider is not nil after incomplete snapshot",
		)
	}

	var target *WorldSpaceNotFoundError
	if !errors.As(err, &target) {
		t.Fatalf(
			"error type = %T, want WorldSpaceNotFoundError",
			err,
		)
	}

	if target.WorldSpaceID != missingID {
		t.Fatalf(
			"error WorldSpaceID = %q, want %q",
			target.WorldSpaceID,
			missingID,
		)
	}
}

func TestProductionProviderCanonicalMetadata(t *testing.T) {
	snapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	if provider.Mode() != ModeProduction {
		t.Fatalf(
			"Mode() = %q, want %q",
			provider.Mode(),
			ModeProduction,
		)
	}

	if provider.WorldID() != CanonicalWorldID {
		t.Fatalf(
			"WorldID() = %q, want %q",
			provider.WorldID(),
			CanonicalWorldID,
		)
	}

	if provider.Version() != 1 {
		t.Fatalf(
			"Version() = %d, want 1",
			provider.Version(),
		)
	}

	canonicalIDsArray := CanonicalWorldSpaceIDs()
	canonicalIDs := canonicalIDsArray[:]

	if !slices.Equal(
		provider.WorldSpaceIDs(),
		canonicalIDs,
	) {
		t.Fatalf(
			"WorldSpaceIDs() = %v, want %v",
			provider.WorldSpaceIDs(),
			canonicalIDs,
		)
	}
}

func TestProductionProviderCanonicalBounds(t *testing.T) {
	snapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	tests := []struct {
		name         string
		worldSpaceID WorldSpaceID
		expected     WorldBounds
	}{
		{
			name:         "main continent",
			worldSpaceID: WorldSpaceMainContinent,
			expected: WorldBounds{
				MinX: 0,
				MinY: 0,
				MaxX: 11999,
				MaxY: 11999,
				MinZ: 0,
				MaxZ: 15,
			},
		},
		{
			name:         "abyssia continent",
			worldSpaceID: WorldSpaceAbyssiaContinent,
			expected: WorldBounds{
				MinX: 0,
				MinY: 0,
				MaxX: 15999,
				MaxY: 15999,
				MinZ: 0,
				MaxZ: 15,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bounds, err :=
				provider.Bounds(test.worldSpaceID)
			if err != nil {
				t.Fatalf("Bounds failed: %v", err)
			}

			if bounds != test.expected {
				t.Fatalf(
					"Bounds = %+v, want %+v",
					bounds,
					test.expected,
				)
			}
		})
	}
}

func TestProductionProviderMissingWorldSpaceErrors(
	t *testing.T,
) {
	snapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	const missingID WorldSpaceID = "missing_space"

	assertWorldSpaceNotFound := func(
		t *testing.T,
		err error,
	) {
		t.Helper()

		var target *WorldSpaceNotFoundError
		if !errors.As(err, &target) {
			t.Fatalf(
				"error type = %T, want WorldSpaceNotFoundError",
				err,
			)
		}

		if target.WorldSpaceID != missingID {
			t.Fatalf(
				"WorldSpaceID = %q, want %q",
				target.WorldSpaceID,
				missingID,
			)
		}
	}

	t.Run("WorldSpace", func(t *testing.T) {
		_, err := provider.WorldSpace(missingID)
		assertWorldSpaceNotFound(t, err)
	})

	t.Run("Bounds", func(t *testing.T) {
		_, err := provider.Bounds(missingID)
		assertWorldSpaceNotFound(t, err)
	})

	t.Run("ChunkReferences", func(t *testing.T) {
		_, err := provider.ChunkReferences(missingID)
		assertWorldSpaceNotFound(t, err)
	})

	t.Run("ChunkReference", func(t *testing.T) {
		_, err := provider.ChunkReference(
			missingID,
			ChunkCoordinate{},
		)
		assertWorldSpaceNotFound(t, err)
	})
}

func TestProductionProviderCanonicalChunkReferences(
	t *testing.T,
) {
	snapshot :=
		loadCanonicalSnapshotForProductionProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	bootstrapCoordinate := ChunkCoordinate{
		ChunkX: 3,
		ChunkY: 3,
		Z:      0,
	}

	for _, worldSpaceID := range provider.WorldSpaceIDs() {
		t.Run(string(worldSpaceID), func(t *testing.T) {
			references, err :=
				provider.ChunkReferences(worldSpaceID)
			if err != nil {
				t.Fatalf(
					"ChunkReferences failed: %v",
					err,
				)
			}

			if references == nil {
				t.Fatal(
					"ChunkReferences returned nil slice",
				)
			}

			wantReferenceCount := 0
			if worldSpaceID == WorldSpaceMainContinent {
				wantReferenceCount = 1
			}

			if len(references) != wantReferenceCount {
				t.Fatalf(
					"reference count = %d, want %d",
					len(references),
					wantReferenceCount,
				)
			}

			if worldSpaceID == WorldSpaceMainContinent {
				reference := references[0]

				if reference.ChunkX != 3 ||
					reference.ChunkY != 3 ||
					reference.Z != 0 {
					t.Fatalf(
						"bootstrap coordinate = (%d,%d,%d), want (3,3,0)",
						reference.ChunkX,
						reference.ChunkY,
						reference.Z,
					)
				}

				if reference.File != "chunks/main_continent/3_3_0.json" {
					t.Fatalf(
						"bootstrap file = %q, want %q",
						reference.File,
						"chunks/main_continent/3_3_0.json",
					)
				}

				if reference.ContentHash != "sha256:81062edbfc2797a9b6f33e20381a2edf1169b56ef189add13992ed97c253bdea" {
					t.Fatalf(
						"bootstrap hash = %q, want canonical hash",
						reference.ContentHash,
					)
				}

				lookupReference, err := provider.ChunkReference(
					worldSpaceID,
					bootstrapCoordinate,
				)
				if err != nil {
					t.Fatalf(
						"bootstrap ChunkReference failed: %v",
						err,
					)
				}

				if lookupReference != reference {
					t.Fatalf(
						"bootstrap lookup = %+v, want %+v",
						lookupReference,
						reference,
					)
				}
			}

			missingCoordinate := ChunkCoordinate{
				ChunkX: 0,
				ChunkY: 0,
				Z:      0,
			}

			_, err = provider.ChunkReference(
				worldSpaceID,
				missingCoordinate,
			)

			var target *ChunkReferenceNotFoundError
			if !errors.As(err, &target) {
				t.Fatalf(
					"error type = %T, want ChunkReferenceNotFoundError",
					err,
				)
			}
		})
	}
}
func TestProductionProviderPublishedChunks(t *testing.T) {
	snapshot :=
		snapshotWithPublishedChunksForProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	references, err := provider.ChunkReferences(
		WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf("ChunkReferences failed: %v", err)
	}

	if len(references) != 3 {
		t.Fatalf(
			"reference count = %d, want 3",
			len(references),
		)
	}

	floorZero, err := provider.ChunkReference(
		WorldSpaceMainContinent,
		ChunkCoordinate{
			ChunkX: 1,
			ChunkY: 2,
			Z:      0,
		},
	)
	if err != nil {
		t.Fatalf(
			"floor-zero ChunkReference failed: %v",
			err,
		)
	}

	if floorZero.File != "chunks/1_2_0.json" {
		t.Fatalf(
			"floor-zero File = %q",
			floorZero.File,
		)
	}

	if floorZero.ContentHash != "hash-floor-zero" {
		t.Fatalf(
			"floor-zero ContentHash = %q",
			floorZero.ContentHash,
		)
	}

	floorOne, err := provider.ChunkReference(
		WorldSpaceMainContinent,
		ChunkCoordinate{
			ChunkX: 1,
			ChunkY: 2,
			Z:      1,
		},
	)
	if err != nil {
		t.Fatalf(
			"floor-one ChunkReference failed: %v",
			err,
		)
	}

	if floorOne.File != "chunks/1_2_1.json" {
		t.Fatalf(
			"floor-one File = %q",
			floorOne.File,
		)
	}

	coordinate := ChunkCoordinate{
		ChunkX: 99,
		ChunkY: 99,
		Z:      7,
	}

	_, err = provider.ChunkReference(
		WorldSpaceMainContinent,
		coordinate,
	)

	var target *ChunkReferenceNotFoundError
	if !errors.As(err, &target) {
		t.Fatalf(
			"error type = %T, want ChunkReferenceNotFoundError",
			err,
		)
	}

	if target.Coordinate != coordinate {
		t.Fatalf(
			"error coordinate = %+v, want %+v",
			target.Coordinate,
			coordinate,
		)
	}
}

func TestProductionProviderOwnsIndependentCopies(
	t *testing.T,
) {
	snapshot :=
		snapshotWithPublishedChunksForProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	snapshot.rootManifest.WorldID = "mutated_world"
	snapshot.rootManifest.WorldVersion = 99
	snapshot.worldSpaceIDs[0] = "mutated_space"

	mainContinent :=
		snapshot.worldSpaces[WorldSpaceMainContinent]
	mainContinent.WidthTiles = 1
	mainContinent.Chunks[0].File = "mutated.json"
	snapshot.worldSpaces[WorldSpaceMainContinent] =
		mainContinent

	if provider.WorldID() == "mutated_world" {
		t.Fatal("provider WorldID changed with snapshot")
	}

	if provider.Version() == 99 {
		t.Fatal("provider Version changed with snapshot")
	}

	if provider.WorldSpaceIDs()[0] == "mutated_space" {
		t.Fatal(
			"provider WorldSpaceIDs changed with snapshot",
		)
	}

	manifest, err :=
		provider.WorldSpace(WorldSpaceMainContinent)
	if err != nil {
		t.Fatalf("WorldSpace failed: %v", err)
	}

	if manifest.WidthTiles == 1 {
		t.Fatal(
			"provider manifest changed with snapshot",
		)
	}

	reference, err := provider.ChunkReference(
		WorldSpaceMainContinent,
		ChunkCoordinate{
			ChunkX: 1,
			ChunkY: 2,
			Z:      0,
		},
	)
	if err != nil {
		t.Fatalf("ChunkReference failed: %v", err)
	}

	if reference.File == "mutated.json" {
		t.Fatal(
			"provider chunk reference changed with snapshot",
		)
	}
}

func TestProductionProviderReturnsDefensiveCopies(
	t *testing.T,
) {
	snapshot :=
		snapshotWithPublishedChunksForProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	worldSpaceIDs := provider.WorldSpaceIDs()
	worldSpaceIDs[0] = "mutated_space"

	if provider.WorldSpaceIDs()[0] == "mutated_space" {
		t.Fatal("WorldSpaceIDs exposed internal slice")
	}

	manifest, err :=
		provider.WorldSpace(WorldSpaceMainContinent)
	if err != nil {
		t.Fatalf("WorldSpace failed: %v", err)
	}

	manifest.Chunks[0].File = "mutated.json"

	manifestAgain, err :=
		provider.WorldSpace(WorldSpaceMainContinent)
	if err != nil {
		t.Fatalf("second WorldSpace failed: %v", err)
	}

	if manifestAgain.Chunks[0].File == "mutated.json" {
		t.Fatal("WorldSpace exposed internal chunks")
	}

	references, err := provider.ChunkReferences(
		WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf("ChunkReferences failed: %v", err)
	}

	references[0].ContentHash = "mutated_hash"

	referencesAgain, err := provider.ChunkReferences(
		WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf(
			"second ChunkReferences failed: %v",
			err,
		)
	}

	if referencesAgain[0].ContentHash ==
		"mutated_hash" {
		t.Fatal(
			"ChunkReferences exposed internal slice",
		)
	}
}

func TestProductionProviderConcurrentReads(t *testing.T) {
	snapshot :=
		snapshotWithPublishedChunksForProviderTest(t)

	provider :=
		newProductionProviderForTest(t, snapshot)

	const goroutineCount = 64

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for index := 0; index < goroutineCount; index++ {
		go func() {
			defer waitGroup.Done()

			_ = provider.Mode()
			_ = provider.WorldID()
			_ = provider.Version()
			_ = provider.WorldSpaceIDs()

			_, _ = provider.WorldSpace(
				WorldSpaceMainContinent,
			)
			_, _ = provider.Bounds(
				WorldSpaceMainContinent,
			)
			_, _ = provider.ChunkReferences(
				WorldSpaceMainContinent,
			)
			_, _ = provider.ChunkReference(
				WorldSpaceMainContinent,
				ChunkCoordinate{
					ChunkX: 1,
					ChunkY: 2,
					Z:      0,
				},
			)
			_, _ = provider.ChunkReference(
				WorldSpaceMainContinent,
				ChunkCoordinate{
					ChunkX: 99,
					ChunkY: 99,
					Z:      7,
				},
			)
		}()
	}

	waitGroup.Wait()
}
