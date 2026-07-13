package worldmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func canonicalRootManifestForLoaderTest() WorldManifest {
	return WorldManifest{
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
			{
				WorldSpaceID:       WorldSpaceFireContinent,
				DisplayName:        "Fire Continent",
				GeographicPosition: GeographicPositionSouth,
				ManifestFile:       "world_spaces/fire_continent.json",
			},
			{
				WorldSpaceID:       WorldSpaceIceContinent,
				DisplayName:        "Ice Continent",
				GeographicPosition: GeographicPositionNorth,
				ManifestFile:       "world_spaces/ice_continent.json",
			},
			{
				WorldSpaceID:       WorldSpaceHolyContinent,
				DisplayName:        "Holy Continent",
				GeographicPosition: GeographicPositionEast,
				ManifestFile:       "world_spaces/holy_continent.json",
			},
			{
				WorldSpaceID:       WorldSpaceShadowContinent,
				DisplayName:        "Shadow Continent",
				GeographicPosition: GeographicPositionWest,
				ManifestFile:       "world_spaces/shadow_continent.json",
			},
			{
				WorldSpaceID:       WorldSpaceNatureContinent,
				DisplayName:        "Nature Continent",
				GeographicPosition: GeographicPositionIntermediate,
				ManifestFile:       "world_spaces/nature_continent.json",
			},
			{
				WorldSpaceID:       WorldSpaceAbyssiaContinent,
				DisplayName:        "Abyssia",
				GeographicPosition: GeographicPositionExtremeNorth,
				ManifestFile:       "world_spaces/abyssia_continent.json",
			},
		},
	}
}

func worldSpaceManifestForLoaderTest(
	reference WorldSpaceReference,
) WorldSpaceManifest {
	width := 12000
	height := 12000

	if reference.WorldSpaceID == WorldSpaceAbyssiaContinent {
		width = 16000
		height = 16000
	}

	return WorldSpaceManifest{
		SchemaVersion:      SupportedSchemaVersion,
		WorldID:            CanonicalWorldID,
		WorldVersion:       1,
		WorldSpaceID:       reference.WorldSpaceID,
		DisplayName:        reference.DisplayName,
		GeographicPosition: reference.GeographicPosition,
		WidthTiles:         width,
		HeightTiles:        height,
		ChunkSize:          CanonicalChunkSize,
		MinFloor:           0,
		MaxFloor:           15,
		Chunks:             []ChunkReference{},
	}
}

func writeJSONForLoaderTest(
	t *testing.T,
	filePath string,
	value interface{},
) {
	t.Helper()

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %q: %v", filePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("create directory for %q: %v", filePath, err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		t.Fatalf("write %q: %v", filePath, err)
	}
}

func writeLoaderFixture(
	t *testing.T,
) (string, WorldManifest) {
	t.Helper()

	rootDirectory := t.TempDir()
	root := canonicalRootManifestForLoaderTest()
	rootPath := filepath.Join(rootDirectory, "world_manifest.json")

	writeJSONForLoaderTest(t, rootPath, root)

	for _, reference := range root.WorldSpaces {
		childPath := filepath.Join(
			rootDirectory,
			filepath.FromSlash(reference.ManifestFile),
		)

		writeJSONForLoaderTest(
			t,
			childPath,
			worldSpaceManifestForLoaderTest(reference),
		)
	}

	return rootPath, root
}

func canonicalManifestPathForLoaderTest(
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

func TestLoadManifestSnapshotCanonical(t *testing.T) {
	snapshot, err := LoadManifestSnapshot(
		canonicalManifestPathForLoaderTest(t),
	)
	if err != nil {
		t.Fatalf("load canonical manifests: %v", err)
	}

	if snapshot.WorldID() != CanonicalWorldID {
		t.Fatalf(
			"WorldID() = %q, want %q",
			snapshot.WorldID(),
			CanonicalWorldID,
		)
	}

	if snapshot.Version() != 1 {
		t.Fatalf(
			"Version() = %d, want 1",
			snapshot.Version(),
		)
	}

	canonicalIDs := CanonicalWorldSpaceIDs()
	wantIDs := canonicalIDs[:]
	gotIDs := snapshot.WorldSpaceIDs()

	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf(
			"WorldSpaceIDs() = %v, want %v",
			gotIDs,
			wantIDs,
		)
	}

	for _, worldSpaceID := range canonicalIDs {
		worldSpace, found := snapshot.WorldSpace(worldSpaceID)
		if !found {
			t.Fatalf(
				"canonical world space %q not found",
				worldSpaceID,
			)
		}

		wantSize := 12000
		if worldSpaceID == WorldSpaceAbyssiaContinent {
			wantSize = 16000
		}

		if worldSpace.WidthTiles != wantSize ||
			worldSpace.HeightTiles != wantSize {
			t.Fatalf(
				"%q dimensions = %dx%d, want %dx%d",
				worldSpaceID,
				worldSpace.WidthTiles,
				worldSpace.HeightTiles,
				wantSize,
				wantSize,
			)
		}

		if worldSpace.ChunkSize != CanonicalChunkSize ||
			worldSpace.MinFloor != 0 ||
			worldSpace.MaxFloor != 15 {
			t.Fatalf(
				"%q has unexpected chunk/floor metadata: %+v",
				worldSpaceID,
				worldSpace,
			)
		}

		if len(worldSpace.Chunks) != 0 {
			t.Fatalf(
				"%q has %d chunk references, want 0",
				worldSpaceID,
				len(worldSpace.Chunks),
			)
		}
	}
}

func TestDecodeStrictJSON(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		var manifest WorldManifest

		err := decodeStrictJSON(
			[]byte(`{"schema_version":1,"unknown":true}`),
			&manifest,
		)

		if err == nil ||
			!strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("got error %v, want unknown-field error", err)
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		var manifest WorldManifest

		err := decodeStrictJSON(
			[]byte(`{"schema_version":1} {"schema_version":1}`),
			&manifest,
		)

		if err == nil ||
			!strings.Contains(
				err.Error(),
				"multiple JSON documents",
			) {
			t.Fatalf(
				"got error %v, want multiple-document error",
				err,
			)
		}
	})

	t.Run("invalid trailing data", func(t *testing.T) {
		var manifest WorldManifest

		err := decodeStrictJSON(
			[]byte(`{"schema_version":1} trailing`),
			&manifest,
		)

		if err == nil ||
			!strings.Contains(err.Error(), "trailing data") {
			t.Fatalf(
				"got error %v, want trailing-data error",
				err,
			)
		}
	})
}

func TestLoadManifestSnapshotFilesystemFailures(t *testing.T) {
	t.Run("missing child", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)

		missingPath := filepath.Join(
			filepath.Dir(rootPath),
			filepath.FromSlash(root.WorldSpaces[0].ManifestFile),
		)

		if err := os.Remove(missingPath); err != nil {
			t.Fatalf("remove child: %v", err)
		}

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(err.Error(), "load world space") {
			t.Fatalf(
				"got error %v, want missing-child error",
				err,
			)
		}
	})

	t.Run("extra JSON manifest", func(t *testing.T) {
		rootPath, _ := writeLoaderFixture(t)

		extraPath := filepath.Join(
			filepath.Dir(rootPath),
			"world_spaces",
			"extra.json",
		)

		if err := os.WriteFile(
			extraPath,
			[]byte(`{}`),
			0o644,
		); err != nil {
			t.Fatalf("write extra manifest: %v", err)
		}

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"unexpected world space manifest file",
			) {
			t.Fatalf(
				"got error %v, want extra-manifest error",
				err,
			)
		}
	})

	t.Run("invalid referenced path", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		root.WorldSpaces[0].ManifestFile = "../outside.json"
		writeJSONForLoaderTest(t, rootPath, root)

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"path traversal is not allowed",
			) {
			t.Fatalf(
				"got error %v, want traversal error",
				err,
			)
		}
	})
}

func TestLoadManifestSnapshotValidationFailures(t *testing.T) {
	t.Run("invalid root", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		root.WorldVersion = 0
		writeJSONForLoaderTest(t, rootPath, root)

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"world_version must be positive",
			) {
			t.Fatalf(
				"got error %v, want root validation error",
				err,
			)
		}
	})

	t.Run("non-canonical world space", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		root.WorldSpaces[0].WorldSpaceID = "unexpected_space"
		writeJSONForLoaderTest(t, rootPath, root)

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"unexpected non-canonical world space",
			) {
			t.Fatalf(
				"got error %v, want non-canonical error",
				err,
			)
		}
	})

	t.Run("invalid child", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		reference := root.WorldSpaces[0]
		child := worldSpaceManifestForLoaderTest(reference)
		child.WidthTiles = 0

		writeJSONForLoaderTest(
			t,
			filepath.Join(
				filepath.Dir(rootPath),
				filepath.FromSlash(reference.ManifestFile),
			),
			child,
		)

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"width_tiles and height_tiles must be positive",
			) {
			t.Fatalf(
				"got error %v, want child validation error",
				err,
			)
		}
	})
}

func TestCheckConsistencyMismatches(t *testing.T) {
	root := canonicalRootManifestForLoaderTest()
	reference := root.WorldSpaces[0]
	valid := worldSpaceManifestForLoaderTest(reference)

	testCases := []struct {
		name     string
		mutate   func(*WorldSpaceManifest)
		wantText string
	}{
		{
			name: "schema version",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.SchemaVersion++
			},
			wantText: "schema_version mismatch",
		},
		{
			name: "world ID",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.WorldID = "other_world"
			},
			wantText: "world_id mismatch",
		},
		{
			name: "world version",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.WorldVersion++
			},
			wantText: "world_version mismatch",
		},
		{
			name: "world space ID",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.WorldSpaceID =
					WorldSpaceFireContinent
			},
			wantText: "world_space_id mismatch",
		},
		{
			name: "display name",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.DisplayName = "Other"
			},
			wantText: "display_name mismatch",
		},
		{
			name: "geographic position",
			mutate: func(worldSpace *WorldSpaceManifest) {
				worldSpace.GeographicPosition =
					GeographicPositionNorth
			},
			wantText: "geographic_position mismatch",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			worldSpace := valid
			testCase.mutate(&worldSpace)

			err := checkConsistency(
				&root,
				&reference,
				&worldSpace,
			)

			if err == nil ||
				!strings.Contains(
					err.Error(),
					testCase.wantText,
				) {
				t.Fatalf(
					"got error %v, want text %q",
					err,
					testCase.wantText,
				)
			}
		})
	}
}

func TestManifestSnapshotIndependentFromDisk(t *testing.T) {
	rootPath, _ := writeLoaderFixture(t)

	snapshot, err := LoadManifestSnapshot(rootPath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	if err := os.RemoveAll(filepath.Dir(rootPath)); err != nil {
		t.Fatalf("remove fixture directory: %v", err)
	}

	if snapshot.WorldID() != CanonicalWorldID {
		t.Fatalf(
			"snapshot changed after disk removal: world ID %q",
			snapshot.WorldID(),
		)
	}

	worldSpace, found := snapshot.WorldSpace(
		WorldSpaceMainContinent,
	)
	if !found {
		t.Fatal(
			"main_continent disappeared after disk removal",
		)
	}

	if worldSpace.WidthTiles != 12000 {
		t.Fatalf(
			"snapshot changed after disk removal: width %d",
			worldSpace.WidthTiles,
		)
	}
}

func TestLoadManifestSnapshotStrictRootDecoding(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		rootPath, _ := writeLoaderFixture(t)

		data, err := os.ReadFile(rootPath)
		if err != nil {
			t.Fatalf("read root manifest: %v", err)
		}

		var document map[string]interface{}
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatalf("decode root fixture: %v", err)
		}

		document["unknown_field"] = true
		writeJSONForLoaderTest(t, rootPath, document)

		_, err = LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(err.Error(), "unknown field") {
			t.Fatalf(
				"got error %v, want root unknown-field error",
				err,
			)
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		rootPath, _ := writeLoaderFixture(t)

		data, err := os.ReadFile(rootPath)
		if err != nil {
			t.Fatalf("read root manifest: %v", err)
		}

		data = append(data, []byte("\n{}")...)

		if err := os.WriteFile(rootPath, data, 0o644); err != nil {
			t.Fatalf("write root manifest: %v", err)
		}

		_, err = LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"multiple JSON documents",
			) {
			t.Fatalf(
				"got error %v, want root multiple-document error",
				err,
			)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		rootPath, _ := writeLoaderFixture(t)

		if err := os.WriteFile(
			rootPath,
			[]byte(`{"schema_version":`),
			0o644,
		); err != nil {
			t.Fatalf("write invalid root manifest: %v", err)
		}

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"decode root world manifest",
			) {
			t.Fatalf(
				"got error %v, want invalid-root error",
				err,
			)
		}
	})
}

func TestLoadManifestSnapshotStrictChildDecoding(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		reference := root.WorldSpaces[0]

		childPath := filepath.Join(
			filepath.Dir(rootPath),
			filepath.FromSlash(reference.ManifestFile),
		)

		data, err := os.ReadFile(childPath)
		if err != nil {
			t.Fatalf("read child manifest: %v", err)
		}

		var document map[string]interface{}
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatalf("decode child fixture: %v", err)
		}

		document["unknown_field"] = true
		writeJSONForLoaderTest(t, childPath, document)

		_, err = LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(err.Error(), "unknown field") {
			t.Fatalf(
				"got error %v, want child unknown-field error",
				err,
			)
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		reference := root.WorldSpaces[0]

		childPath := filepath.Join(
			filepath.Dir(rootPath),
			filepath.FromSlash(reference.ManifestFile),
		)

		data, err := os.ReadFile(childPath)
		if err != nil {
			t.Fatalf("read child manifest: %v", err)
		}

		data = append(data, []byte("\n{}")...)

		if err := os.WriteFile(childPath, data, 0o644); err != nil {
			t.Fatalf("write child manifest: %v", err)
		}

		_, err = LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"multiple JSON documents",
			) {
			t.Fatalf(
				"got error %v, want child multiple-document error",
				err,
			)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		rootPath, root := writeLoaderFixture(t)
		reference := root.WorldSpaces[0]

		childPath := filepath.Join(
			filepath.Dir(rootPath),
			filepath.FromSlash(reference.ManifestFile),
		)

		if err := os.WriteFile(
			childPath,
			[]byte(`{"schema_version":`),
			0o644,
		); err != nil {
			t.Fatalf("write invalid child manifest: %v", err)
		}

		_, err := LoadManifestSnapshot(rootPath)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"decode world space",
			) {
			t.Fatalf(
				"got error %v, want invalid-child error",
				err,
			)
		}
	})
}
