package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

func canonicalWorldMapRuntimePathsForTest(
	t *testing.T,
) (string, string) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	backendRoot := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"..",
			"..",
		),
	)

	return filepath.Join(
			backendRoot,
			"config",
			"worldmap",
			"world_manifest.json",
		),
		filepath.Join(
			backendRoot,
			"config",
			"worldmap",
			"initial_residency.json",
		)
}

func canonicalWorldMapRuntimeResidencyForTest() worldmap.InitialResidencyManifest {
	return worldmap.InitialResidencyManifest{
		SchemaVersion: worldmap.SupportedSchemaVersion,
		WorldID:       worldmap.CanonicalWorldID,
		WorldVersion:  1,
		ResidentChunks: []worldmap.ResidentChunk{
			{
				WorldSpaceID: worldmap.WorldSpaceMainContinent,
				ChunkX:       3,
				ChunkY:       3,
				Z:            0,
			},
		},
	}
}

func writeWorldMapRuntimeResidencyForTest(
	t *testing.T,
	manifest worldmap.InitialResidencyManifest,
) string {
	t.Helper()

	data, err := json.MarshalIndent(
		manifest,
		"",
		"  ",
	)
	if err != nil {
		t.Fatalf("marshal residency fixture: %v", err)
	}

	data = append(data, '\n')

	path := filepath.Join(
		t.TempDir(),
		"initial_residency.json",
	)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write residency fixture: %v", err)
	}

	return path
}

func TestInitializeWorldMapRuntimeDebugIgnoresFiles(
	t *testing.T,
) {
	missingRoot := t.TempDir()

	state, err := initializeWorldMapRuntime(
		worldmap.ModeDebug,
		filepath.Join(missingRoot, "missing-manifest.json"),
		filepath.Join(missingRoot, "missing-residency.json"),
	)
	if err != nil {
		t.Fatalf("initialize debug runtime: %v", err)
	}

	if state == nil {
		t.Fatal("state = nil, want non-nil")
	}

	if state.provider == nil {
		t.Fatal("provider = nil, want debug provider")
	}

	if state.provider.Mode() != worldmap.ModeDebug {
		t.Fatalf(
			"provider mode = %q, want %q",
			state.provider.Mode(),
			worldmap.ModeDebug,
		)
	}

	if state.staticCollisionIndex != nil {
		t.Fatal(
			"debug static collision index must remain nil",
		)
	}

	if state.residentChunkCount != 0 {
		t.Fatalf(
			"debug resident chunk count = %d, want 0",
			state.residentChunkCount,
		)
	}
}

func TestInitializeWorldMapRuntimeCanonicalProduction(
	t *testing.T,
) {
	manifestPath, residencyPath :=
		canonicalWorldMapRuntimePathsForTest(t)

	state, err := initializeWorldMapRuntime(
		worldmap.ModeProduction,
		"  "+manifestPath+"  ",
		residencyPath,
	)
	if err != nil {
		t.Fatalf(
			"initialize canonical production runtime: %v",
			err,
		)
	}

	if state == nil {
		t.Fatal("state = nil, want non-nil")
	}

	if state.provider == nil {
		t.Fatal("provider = nil, want production provider")
	}

	if state.provider.Mode() != worldmap.ModeProduction {
		t.Fatalf(
			"provider mode = %q, want %q",
			state.provider.Mode(),
			worldmap.ModeProduction,
		)
	}

	if state.provider.WorldID() !=
		worldmap.CanonicalWorldID {
		t.Fatalf(
			"provider world ID = %q, want %q",
			state.provider.WorldID(),
			worldmap.CanonicalWorldID,
		)
	}

	if state.provider.Version() != 1 {
		t.Fatalf(
			"provider version = %d, want 1",
			state.provider.Version(),
		)
	}

	if state.staticCollisionIndex == nil {
		t.Fatal(
			"static collision index = nil, want non-nil",
		)
	}

	if state.residentChunkCount != 1 {
		t.Fatalf(
			"resident chunk count = %d, want 1",
			state.residentChunkCount,
		)
	}

	positions := []worldmap.WorldPosition{
		{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            96,
			Y:            96,
			Z:            0,
		},
		{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            112,
			Y:            108,
			Z:            0,
		},
		{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            127,
			Y:            127,
			Z:            0,
		},
	}

	for _, position := range positions {
		canOccupy, err :=
			state.staticCollisionIndex.CanOccupy(
				position,
			)
		if err != nil {
			t.Fatalf(
				"CanOccupy(%+v): %v",
				position,
				err,
			)
		}

		if !canOccupy {
			t.Fatalf(
				"canonical position %+v cannot be occupied",
				position,
			)
		}
	}

	unpublishedPosition := worldmap.WorldPosition{
		WorldSpaceID: worldmap.WorldSpaceMainContinent,
		X:            128,
		Y:            128,
		Z:            0,
	}

	_, err = state.staticCollisionIndex.CanOccupy(
		unpublishedPosition,
	)

	var notPublished *worldmap.ChunkReferenceNotFoundError

	if !errors.As(err, &notPublished) {
		t.Fatalf(
			"unpublished collision error = %v, want ChunkReferenceNotFoundError",
			err,
		)
	}

	if notPublished.Coordinate !=
		(worldmap.ChunkCoordinate{
			ChunkX: 4,
			ChunkY: 4,
			Z:      0,
		}) {
		t.Fatalf(
			"unpublished coordinate = %+v, want (4,4,0)",
			notPublished.Coordinate,
		)
	}
}

func TestInitializeWorldMapRuntimeProductionResidencyFailures(
	t *testing.T,
) {
	manifestPath, _ :=
		canonicalWorldMapRuntimePathsForTest(t)

	testCases := []struct {
		name          string
		residencyPath func(*testing.T) string
		expectedText  string
		expectedField string
	}{
		{
			name: "empty path",
			residencyPath: func(*testing.T) string {
				return ""
			},
			expectedText: "initial residency path cannot be empty",
		},
		{
			name: "missing file",
			residencyPath: func(t *testing.T) string {
				return filepath.Join(
					t.TempDir(),
					"missing.json",
				)
			},
			expectedText: "open initial residency",
		},
		{
			name: "world ID mismatch",
			residencyPath: func(t *testing.T) string {
				manifest :=
					canonicalWorldMapRuntimeResidencyForTest()

				manifest.WorldID = "other_world"

				return writeWorldMapRuntimeResidencyForTest(
					t,
					manifest,
				)
			},
			expectedText:  "world_id mismatch",
			expectedField: "world_id",
		},
		{
			name: "world version mismatch",
			residencyPath: func(t *testing.T) string {
				manifest :=
					canonicalWorldMapRuntimeResidencyForTest()

				manifest.WorldVersion = 2

				return writeWorldMapRuntimeResidencyForTest(
					t,
					manifest,
				)
			},
			expectedText:  "world_version mismatch",
			expectedField: "world_version",
		},
		{
			name: "unpublished resident chunk",
			residencyPath: func(t *testing.T) string {
				manifest :=
					canonicalWorldMapRuntimeResidencyForTest()

				manifest.ResidentChunks[0].ChunkX = 4
				manifest.ResidentChunks[0].ChunkY = 4

				return writeWorldMapRuntimeResidencyForTest(
					t,
					manifest,
				)
			},
			expectedText: "resolve published resident chunk",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			state, err := initializeWorldMapRuntime(
				worldmap.ModeProduction,
				manifestPath,
				testCase.residencyPath(t),
			)

			if state != nil {
				t.Fatalf(
					"state = %+v, want nil",
					state,
				)
			}

			if err == nil {
				t.Fatal("error = nil, want non-nil")
			}

			if !strings.Contains(
				err.Error(),
				testCase.expectedText,
			) {
				t.Fatalf(
					"error = %q, want text %q",
					err,
					testCase.expectedText,
				)
			}

			if testCase.expectedField == "" {
				return
			}

			var mismatch *worldMapRuntimeIdentityMismatchError

			if !errors.As(err, &mismatch) {
				t.Fatalf(
					"error = %v, want worldMapRuntimeIdentityMismatchError",
					err,
				)
			}

			if mismatch.Field !=
				testCase.expectedField {
				t.Fatalf(
					"mismatch field = %q, want %q",
					mismatch.Field,
					testCase.expectedField,
				)
			}
		})
	}
}

func TestInitializeWorldMapRuntimeRejectsUnsupportedMode(
	t *testing.T,
) {
	state, err := initializeWorldMapRuntime(
		worldmap.Mode("unsupported"),
		"ignored",
		"ignored",
	)

	if state != nil {
		t.Fatalf("state = %+v, want nil", state)
	}

	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}
}

func TestInitializeWorldMapRuntimeFailureIsAtomic(
	t *testing.T,
) {
	manifestPath, _ :=
		canonicalWorldMapRuntimePathsForTest(t)

	manifest :=
		canonicalWorldMapRuntimeResidencyForTest()

	manifest.ResidentChunks[0].ChunkX = 99
	manifest.ResidentChunks[0].ChunkY = 99

	state, err := initializeWorldMapRuntime(
		worldmap.ModeProduction,
		manifestPath,
		writeWorldMapRuntimeResidencyForTest(
			t,
			manifest,
		),
	)

	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}

	if state != nil {
		t.Fatalf(
			"state = %+v, want nil after failure",
			state,
		)
	}
}
