package worldmap

import (
	"reflect"
	"sync"
	"testing"
)

func newManifestSnapshotForTest() *ManifestSnapshot {
	root := WorldManifest{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		WorldSpaces: []WorldSpaceReference{
			{
				WorldSpaceID:       WorldSpaceMainContinent,
				DisplayName:        "Main Continent",
				GeographicPosition: GeographicPositionCentral,
				ManifestFile:       "world_spaces/main_continent.json",
			},
		},
	}

	worldSpace := WorldSpaceManifest{
		SchemaVersion:      SupportedSchemaVersion,
		WorldID:            CanonicalWorldID,
		WorldVersion:       1,
		WorldSpaceID:       WorldSpaceMainContinent,
		DisplayName:        "Main Continent",
		GeographicPosition: GeographicPositionCentral,
		WidthTiles:         12000,
		HeightTiles:        12000,
		ChunkSize:          CanonicalChunkSize,
		MinFloor:           0,
		MaxFloor:           15,
		Chunks: []ChunkReference{
			{
				ChunkX: 1,
				ChunkY: 2,
				Z:      0,
				File:   "chunks/1_2_0.json",
			},
		},
	}

	return &ManifestSnapshot{
		rootManifest: root,
		worldSpaces: map[WorldSpaceID]WorldSpaceManifest{
			WorldSpaceMainContinent: worldSpace,
		},
		worldSpaceIDs: []WorldSpaceID{
			WorldSpaceMainContinent,
		},
	}
}

func TestManifestSnapshotDefensiveCopies(t *testing.T) {
	snapshot := newManifestSnapshotForTest()

	rootCopy := snapshot.RootManifest()
	rootCopy.WorldVersion = 99
	rootCopy.WorldSpaces[0].ManifestFile = "mutated.json"
	rootCopy.WorldSpaces = append(
		rootCopy.WorldSpaces,
		WorldSpaceReference{WorldSpaceID: "unexpected"},
	)

	rootAgain := snapshot.RootManifest()

	if rootAgain.WorldVersion != 1 {
		t.Fatalf(
			"root snapshot version mutated: got %d, want 1",
			rootAgain.WorldVersion,
		)
	}

	if got := rootAgain.WorldSpaces[0].ManifestFile; got !=
		"world_spaces/main_continent.json" {
		t.Fatalf(
			"root world-space reference mutated: got %q",
			got,
		)
	}

	if len(rootAgain.WorldSpaces) != 1 {
		t.Fatalf(
			"root world-space slice mutated: got length %d, want 1",
			len(rootAgain.WorldSpaces),
		)
	}

	idsCopy := snapshot.WorldSpaceIDs()
	idsCopy[0] = WorldSpaceFireContinent
	idsCopy = append(idsCopy, WorldSpaceIceContinent)

	idsAgain := snapshot.WorldSpaceIDs()
	wantIDs := []WorldSpaceID{WorldSpaceMainContinent}

	if !reflect.DeepEqual(idsAgain, wantIDs) {
		t.Fatalf(
			"world-space IDs mutated: got %v, want %v",
			idsAgain,
			wantIDs,
		)
	}

	worldSpaceCopy, found := snapshot.WorldSpace(
		WorldSpaceMainContinent,
	)
	if !found {
		t.Fatal("expected main_continent to exist")
	}

	worldSpaceCopy.WorldVersion = 99
	worldSpaceCopy.Chunks[0].File = "mutated.json"
	worldSpaceCopy.Chunks = append(
		worldSpaceCopy.Chunks,
		ChunkReference{File: "chunks/extra.json"},
	)

	worldSpaceAgain, found := snapshot.WorldSpace(
		WorldSpaceMainContinent,
	)
	if !found {
		t.Fatal("expected main_continent to remain available")
	}

	if worldSpaceAgain.WorldVersion != 1 {
		t.Fatalf(
			"world-space version mutated: got %d, want 1",
			worldSpaceAgain.WorldVersion,
		)
	}

	if got := worldSpaceAgain.Chunks[0].File; got !=
		"chunks/1_2_0.json" {
		t.Fatalf("chunk reference mutated: got %q", got)
	}

	if len(worldSpaceAgain.Chunks) != 1 {
		t.Fatalf(
			"chunk slice mutated: got length %d, want 1",
			len(worldSpaceAgain.Chunks),
		)
	}
}

func TestManifestSnapshotMissingWorldSpace(t *testing.T) {
	snapshot := newManifestSnapshotForTest()

	worldSpace, found := snapshot.WorldSpace(
		WorldSpaceFireContinent,
	)

	if found {
		t.Fatalf(
			"unexpected world space returned: %+v",
			worldSpace,
		)
	}
}

func TestManifestSnapshotConcurrentReads(t *testing.T) {
	snapshot := newManifestSnapshotForTest()

	const goroutineCount = 32
	const readsPerGoroutine = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for index := 0; index < goroutineCount; index++ {
		go func() {
			defer waitGroup.Done()

			for read := 0; read < readsPerGoroutine; read++ {
				if snapshot.WorldID() != CanonicalWorldID {
					t.Errorf(
						"unexpected world ID: %q",
						snapshot.WorldID(),
					)
					return
				}

				if snapshot.Version() != 1 {
					t.Errorf(
						"unexpected version: %d",
						snapshot.Version(),
					)
					return
				}

				ids := snapshot.WorldSpaceIDs()
				if len(ids) != 1 ||
					ids[0] != WorldSpaceMainContinent {
					t.Errorf(
						"unexpected world-space IDs: %v",
						ids,
					)
					return
				}

				worldSpace, found := snapshot.WorldSpace(
					WorldSpaceMainContinent,
				)
				if !found ||
					worldSpace.WorldSpaceID !=
						WorldSpaceMainContinent {
					t.Errorf(
						"unexpected world-space result: found=%v value=%+v",
						found,
						worldSpace,
					)
					return
				}
			}
		}()
	}

	waitGroup.Wait()
}
