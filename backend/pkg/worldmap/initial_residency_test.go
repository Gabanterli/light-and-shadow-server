package worldmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func validInitialResidencyManifestForTest() InitialResidencyManifest {
	return InitialResidencyManifest{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		ResidentChunks: []ResidentChunk{
			{
				WorldSpaceID: WorldSpaceMainContinent,
				ChunkX:       3,
				ChunkY:       3,
				Z:            0,
			},
		},
	}
}

func cloneInitialResidencyManifestForTest(
	manifest InitialResidencyManifest,
) InitialResidencyManifest {
	cloned := manifest
	cloned.ResidentChunks = append(
		[]ResidentChunk(nil),
		manifest.ResidentChunks...,
	)

	return cloned
}

func canonicalInitialResidencyPathForTest(
	t *testing.T,
) string {
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
		"initial_residency.json",
	)
}

func writeInitialResidencyBytesForTest(
	t *testing.T,
	data []byte,
) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"initial_residency.json",
	)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write residency fixture: %v", err)
	}

	return path
}

func writeInitialResidencyManifestForTest(
	t *testing.T,
	manifest InitialResidencyManifest,
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

	return writeInitialResidencyBytesForTest(
		t,
		data,
	)
}

func TestInitialResidencyCanonicalConfiguration(
	t *testing.T,
) {
	snapshot, err := LoadInitialResidencySnapshot(
		canonicalInitialResidencyPathForTest(t),
	)
	if err != nil {
		t.Fatalf("load canonical initial residency: %v", err)
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

	manifest := snapshot.Manifest()

	if manifest.SchemaVersion != SupportedSchemaVersion {
		t.Fatalf(
			"SchemaVersion = %d, want %d",
			manifest.SchemaVersion,
			SupportedSchemaVersion,
		)
	}

	residents := snapshot.ResidentChunks()

	if len(residents) != 1 {
		t.Fatalf(
			"resident chunk count = %d, want 1",
			len(residents),
		)
	}

	resident := residents[0]

	if resident.WorldSpaceID != WorldSpaceMainContinent {
		t.Fatalf(
			"world_space_id = %q, want %q",
			resident.WorldSpaceID,
			WorldSpaceMainContinent,
		)
	}

	wantCoordinate := ChunkCoordinate{
		ChunkX: 3,
		ChunkY: 3,
		Z:      0,
	}

	if resident.Coordinate() != wantCoordinate {
		t.Fatalf(
			"coordinate = %+v, want %+v",
			resident.Coordinate(),
			wantCoordinate,
		)
	}
}

func TestInitialResidencyValidationFailures(
	t *testing.T,
) {
	testCases := []struct {
		name   string
		mutate func(*InitialResidencyManifest)
		want   string
	}{
		{
			name: "unsupported schema",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.SchemaVersion++
			},
			want: "schema_version",
		},
		{
			name: "empty world ID",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.WorldID = ""
			},
			want: "world_id cannot be empty",
		},
		{
			name: "whitespace world ID",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.WorldID = "   "
			},
			want: "world_id cannot be empty",
		},
		{
			name: "zero world version",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.WorldVersion = 0
			},
			want: "world_version must be positive",
		},
		{
			name: "negative world version",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.WorldVersion = -1
			},
			want: "world_version must be positive",
		},
		{
			name: "nil resident chunks",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks = nil
			},
			want: ErrInitialResidencyEmpty.Error(),
		},
		{
			name: "empty resident chunks",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks = []ResidentChunk{}
			},
			want: ErrInitialResidencyEmpty.Error(),
		},
		{
			name: "empty world space ID",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks[0].WorldSpaceID = ""
			},
			want: "world_space_id cannot be empty",
		},
		{
			name: "whitespace world space ID",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks[0].WorldSpaceID = "  "
			},
			want: "world_space_id cannot be empty",
		},
		{
			name: "negative chunk X",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks[0].ChunkX = -1
			},
			want: "chunk_x cannot be negative",
		},
		{
			name: "negative chunk Y",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks[0].ChunkY = -1
			},
			want: "chunk_y cannot be negative",
		},
		{
			name: "negative floor",
			mutate: func(manifest *InitialResidencyManifest) {
				manifest.ResidentChunks[0].Z = -1
			},
			want: ".z cannot be negative",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			manifest := validInitialResidencyManifestForTest()
			testCase.mutate(&manifest)

			err := ValidateInitialResidencyManifest(manifest)
			if err == nil {
				t.Fatalf("expected validation error")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf(
					"error = %q, want text %q",
					err,
					testCase.want,
				)
			}
		})
	}
}

func TestInitialResidencyDuplicateValidation(
	t *testing.T,
) {
	t.Run("consecutive duplicate", func(t *testing.T) {
		manifest := validInitialResidencyManifestForTest()
		manifest.ResidentChunks = append(
			manifest.ResidentChunks,
			manifest.ResidentChunks[0],
		)

		err := ValidateInitialResidencyManifest(manifest)

		var duplicateError *DuplicateResidentChunkError

		if !errors.As(err, &duplicateError) {
			t.Fatalf(
				"error = %v, want DuplicateResidentChunkError",
				err,
			)
		}

		if duplicateError.FirstIndex != 0 ||
			duplicateError.SecondIndex != 1 {
			t.Fatalf(
				"duplicate indexes = %d,%d, want 0,1",
				duplicateError.FirstIndex,
				duplicateError.SecondIndex,
			)
		}
	})

	t.Run("non-consecutive duplicate", func(t *testing.T) {
		manifest := validInitialResidencyManifestForTest()

		first := manifest.ResidentChunks[0]

		manifest.ResidentChunks = []ResidentChunk{
			first,
			{
				WorldSpaceID: WorldSpaceFireContinent,
				ChunkX:       1,
				ChunkY:       2,
				Z:            0,
			},
			first,
		}

		err := ValidateInitialResidencyManifest(manifest)

		var duplicateError *DuplicateResidentChunkError

		if !errors.As(err, &duplicateError) {
			t.Fatalf(
				"error = %v, want DuplicateResidentChunkError",
				err,
			)
		}

		if duplicateError.FirstIndex != 0 ||
			duplicateError.SecondIndex != 2 {
			t.Fatalf(
				"duplicate indexes = %d,%d, want 0,2",
				duplicateError.FirstIndex,
				duplicateError.SecondIndex,
			)
		}
	})

	t.Run("same coordinate in different spaces is valid", func(t *testing.T) {
		manifest := validInitialResidencyManifestForTest()

		manifest.ResidentChunks = append(
			manifest.ResidentChunks,
			ResidentChunk{
				WorldSpaceID: WorldSpaceFireContinent,
				ChunkX:       3,
				ChunkY:       3,
				Z:            0,
			},
		)

		if err := ValidateInitialResidencyManifest(manifest); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})
}

func TestInitialResidencyValidationDoesNotMutate(
	t *testing.T,
) {
	manifest := validInitialResidencyManifestForTest()

	manifest.ResidentChunks = []ResidentChunk{
		{
			WorldSpaceID: WorldSpaceMainContinent,
			ChunkX:       9,
			ChunkY:       1,
			Z:            0,
		},
		{
			WorldSpaceID: WorldSpaceFireContinent,
			ChunkX:       1,
			ChunkY:       9,
			Z:            0,
		},
	}

	before := cloneInitialResidencyManifestForTest(manifest)

	if err := ValidateInitialResidencyManifest(manifest); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if !reflect.DeepEqual(manifest, before) {
		t.Fatalf(
			"validation mutated manifest: got %+v, want %+v",
			manifest,
			before,
		)
	}
}

func TestInitialResidencyStrictJSON(
	t *testing.T,
) {
	validDocument := `{
        "schema_version": 1,
        "world_id": "light_and_shadow_world",
        "world_version": 1,
        "resident_chunks": [
            {
                "world_space_id": "main_continent",
                "chunk_x": 3,
                "chunk_y": 3,
                "z": 0
            }
        ]
    }`

	testCases := []struct {
		name string
		data string
		want string
	}{
		{
			name: "unknown root field",
			data: `{
                "schema_version": 1,
                "world_id": "light_and_shadow_world",
                "world_version": 1,
                "unknown": true,
                "resident_chunks": [{
                    "world_space_id": "main_continent",
                    "chunk_x": 3,
                    "chunk_y": 3,
                    "z": 0
                }]
            }`,
			want: "unknown field",
		},
		{
			name: "unknown resident field",
			data: `{
                "schema_version": 1,
                "world_id": "light_and_shadow_world",
                "world_version": 1,
                "resident_chunks": [{
                    "world_space_id": "main_continent",
                    "chunk_x": 3,
                    "chunk_y": 3,
                    "z": 0,
                    "unknown": true
                }]
            }`,
			want: "unknown field",
		},
		{
			name: "multiple documents",
			data: validDocument + "\n{}",
			want: "multiple JSON documents",
		},
		{
			name: "invalid trailing data",
			data: validDocument + "\ntrailing",
			want: "trailing data",
		},
		{
			name: "malformed JSON",
			data: `{"schema_version":`,
			want: "decode initial residency",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path := writeInitialResidencyBytesForTest(
				t,
				[]byte(testCase.data),
			)

			snapshot, err := LoadInitialResidencySnapshot(path)

			if snapshot != nil {
				t.Fatalf(
					"snapshot = %+v, want nil",
					snapshot,
				)
			}

			if err == nil {
				t.Fatal("expected loader error")
			}

			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf(
					"error = %q, want text %q",
					err,
					testCase.want,
				)
			}
		})
	}
}

func TestInitialResidencyFilesystemFailures(
	t *testing.T,
) {
	t.Run("empty path", func(t *testing.T) {
		snapshot, err := LoadInitialResidencySnapshot("")

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if !errors.Is(err, ErrInitialResidencyPathEmpty) {
			t.Fatalf(
				"error = %v, want ErrInitialResidencyPathEmpty",
				err,
			)
		}
	})

	t.Run("whitespace path", func(t *testing.T) {
		snapshot, err := LoadInitialResidencySnapshot("   ")

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if !errors.Is(err, ErrInitialResidencyPathEmpty) {
			t.Fatalf(
				"error = %v, want ErrInitialResidencyPathEmpty",
				err,
			)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		path := filepath.Join(
			t.TempDir(),
			"missing.json",
		)

		snapshot, err := LoadInitialResidencySnapshot(path)

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if err == nil {
			t.Fatal("expected missing-file error")
		}
	})

	t.Run("directory", func(t *testing.T) {
		snapshot, err := LoadInitialResidencySnapshot(
			t.TempDir(),
		)

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if err == nil ||
			!strings.Contains(err.Error(), "directory") {
			t.Fatalf(
				"error = %v, want directory error",
				err,
			)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := writeInitialResidencyBytesForTest(
			t,
			nil,
		)

		snapshot, err := LoadInitialResidencySnapshot(path)

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if err == nil {
			t.Fatal("expected empty-file error")
		}
	})

	t.Run("file above limit", func(t *testing.T) {
		data := bytes.Repeat(
			[]byte{'x'},
			int(maxInitialResidencyBytes)+1,
		)

		path := writeInitialResidencyBytesForTest(
			t,
			data,
		)

		snapshot, err := LoadInitialResidencySnapshot(path)

		if snapshot != nil {
			t.Fatalf("snapshot = %+v, want nil", snapshot)
		}

		if !errors.Is(
			err,
			ErrInitialResidencyFileTooLarge,
		) {
			t.Fatalf(
				"error = %v, want ErrInitialResidencyFileTooLarge",
				err,
			)
		}

		var sizeError *InitialResidencyFileTooLargeError

		if !errors.As(err, &sizeError) {
			t.Fatalf(
				"error = %v, want InitialResidencyFileTooLargeError",
				err,
			)
		}

		if sizeError.Limit != maxInitialResidencyBytes {
			t.Fatalf(
				"limit = %d, want %d",
				sizeError.Limit,
				maxInitialResidencyBytes,
			)
		}
	})
}

func TestInitialResidencySnapshotIndependentFromDisk(
	t *testing.T,
) {
	path := writeInitialResidencyManifestForTest(
		t,
		validInitialResidencyManifestForTest(),
	)

	snapshot, err := LoadInitialResidencySnapshot(path)
	if err != nil {
		t.Fatalf("load residency fixture: %v", err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove residency fixture: %v", err)
	}

	if snapshot.WorldID() != CanonicalWorldID {
		t.Fatalf(
			"WorldID() after removal = %q, want %q",
			snapshot.WorldID(),
			CanonicalWorldID,
		)
	}

	if len(snapshot.ResidentChunks()) != 1 {
		t.Fatalf(
			"resident count after removal = %d, want 1",
			len(snapshot.ResidentChunks()),
		)
	}
}

func TestInitialResidencyCanonicalOrdering(
	t *testing.T,
) {
	manifest := validInitialResidencyManifestForTest()

	manifest.ResidentChunks = []ResidentChunk{
		{
			WorldSpaceID: "space_b",
			ChunkX:       0,
			ChunkY:       0,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       0,
			ChunkY:       0,
			Z:            1,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       0,
			ChunkY:       2,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       2,
			ChunkY:       1,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       1,
			ChunkY:       1,
			Z:            0,
		},
	}

	path := writeInitialResidencyManifestForTest(
		t,
		manifest,
	)

	snapshot, err := LoadInitialResidencySnapshot(path)
	if err != nil {
		t.Fatalf("load residency fixture: %v", err)
	}

	got := snapshot.ResidentChunks()

	want := []ResidentChunk{
		{
			WorldSpaceID: "space_a",
			ChunkX:       1,
			ChunkY:       1,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       2,
			ChunkY:       1,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       0,
			ChunkY:       2,
			Z:            0,
		},
		{
			WorldSpaceID: "space_a",
			ChunkX:       0,
			ChunkY:       0,
			Z:            1,
		},
		{
			WorldSpaceID: "space_b",
			ChunkX:       0,
			ChunkY:       0,
			Z:            0,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"ordered residents = %+v, want %+v",
			got,
			want,
		)
	}
}

func TestInitialResidencyDefensiveCopies(
	t *testing.T,
) {
	path := writeInitialResidencyManifestForTest(
		t,
		validInitialResidencyManifestForTest(),
	)

	snapshot, err := LoadInitialResidencySnapshot(path)
	if err != nil {
		t.Fatalf("load residency fixture: %v", err)
	}

	firstResidents := snapshot.ResidentChunks()
	firstResidents[0].ChunkX = 999

	secondResidents := snapshot.ResidentChunks()

	if secondResidents[0].ChunkX != 3 {
		t.Fatalf(
			"ResidentChunks mutation leaked: got %d, want 3",
			secondResidents[0].ChunkX,
		)
	}

	firstManifest := snapshot.Manifest()
	firstManifest.ResidentChunks[0].ChunkY = 999

	secondManifest := snapshot.Manifest()

	if secondManifest.ResidentChunks[0].ChunkY != 3 {
		t.Fatalf(
			"Manifest mutation leaked: got %d, want 3",
			secondManifest.ResidentChunks[0].ChunkY,
		)
	}

	secondResidents[0].Z = 99

	thirdResidents := snapshot.ResidentChunks()

	if thirdResidents[0].Z != 0 {
		t.Fatalf(
			"independent slice mutation leaked: got %d, want 0",
			thirdResidents[0].Z,
		)
	}
}

func TestInitialResidencyConcurrentReads(
	t *testing.T,
) {
	path := writeInitialResidencyManifestForTest(
		t,
		validInitialResidencyManifestForTest(),
	)

	snapshot, err := LoadInitialResidencySnapshot(path)
	if err != nil {
		t.Fatalf("load residency fixture: %v", err)
	}

	const goroutineCount = 16
	const readsPerGoroutine = 100

	failures := make(
		chan error,
		goroutineCount,
	)

	var waitGroup sync.WaitGroup

	for worker := 0; worker < goroutineCount; worker++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			for read := 0; read < readsPerGoroutine; read++ {
				if snapshot.WorldID() != CanonicalWorldID {
					failures <- fmt.Errorf(
						"unexpected world ID %q",
						snapshot.WorldID(),
					)
					return
				}

				if snapshot.Version() != 1 {
					failures <- fmt.Errorf(
						"unexpected version %d",
						snapshot.Version(),
					)
					return
				}

				if len(snapshot.Manifest().ResidentChunks) != 1 {
					failures <- errors.New(
						"unexpected manifest resident count",
					)
					return
				}

				if len(snapshot.ResidentChunks()) != 1 {
					failures <- errors.New(
						"unexpected resident chunk count",
					)
					return
				}
			}
		}()
	}

	waitGroup.Wait()
	close(failures)

	for failure := range failures {
		t.Error(failure)
	}
}

func TestInitialResidencyConcurrentLoads(
	t *testing.T,
) {
	path := writeInitialResidencyManifestForTest(
		t,
		validInitialResidencyManifestForTest(),
	)

	const goroutineCount = 16

	failures := make(
		chan error,
		goroutineCount,
	)

	var waitGroup sync.WaitGroup

	for worker := 0; worker < goroutineCount; worker++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			snapshot, err := LoadInitialResidencySnapshot(path)
			if err != nil {
				failures <- err
				return
			}

			residents := snapshot.ResidentChunks()

			if len(residents) != 1 ||
				residents[0].WorldSpaceID != WorldSpaceMainContinent {
				failures <- fmt.Errorf(
					"unexpected resident chunks: %+v",
					residents,
				)
			}
		}()
	}

	waitGroup.Wait()
	close(failures)

	for failure := range failures {
		t.Error(failure)
	}
}

func TestInitialResidencyErrorContracts(
	t *testing.T,
) {
	if ErrInitialResidencyPathEmpty.Error() == "" ||
		ErrInitialResidencyFileTooLarge.Error() == "" ||
		ErrInitialResidencyEmpty.Error() == "" {
		t.Fatal("sentinel error message cannot be empty")
	}

	duplicateError := &DuplicateResidentChunkError{
		WorldSpaceID: WorldSpaceMainContinent,
		Coordinate: ChunkCoordinate{
			ChunkX: 3,
			ChunkY: 3,
			Z:      0,
		},
		FirstIndex:  1,
		SecondIndex: 4,
	}

	if !strings.Contains(
		duplicateError.Error(),
		"main_continent",
	) {
		t.Fatalf(
			"duplicate error message = %q",
			duplicateError.Error(),
		)
	}

	sizeError := &InitialResidencyFileTooLargeError{
		Path:  "residency.json",
		Size:  maxInitialResidencyBytes + 1,
		Limit: maxInitialResidencyBytes,
	}

	if !errors.Is(
		sizeError,
		ErrInitialResidencyFileTooLarge,
	) {
		t.Fatal(
			"size error does not unwrap to ErrInitialResidencyFileTooLarge",
		)
	}
}

func TestInitialResidencyLoaderAtomicity(
	t *testing.T,
) {
	testCases := []struct {
		name string
		path func(*testing.T) string
	}{
		{
			name: "malformed",
			path: func(t *testing.T) string {
				return writeInitialResidencyBytesForTest(
					t,
					[]byte(`{"schema_version":`),
				)
			},
		},
		{
			name: "invalid manifest",
			path: func(t *testing.T) string {
				manifest := validInitialResidencyManifestForTest()
				manifest.ResidentChunks = nil

				return writeInitialResidencyManifestForTest(
					t,
					manifest,
				)
			},
		},
		{
			name: "duplicate resident",
			path: func(t *testing.T) string {
				manifest := validInitialResidencyManifestForTest()

				manifest.ResidentChunks = append(
					manifest.ResidentChunks,
					manifest.ResidentChunks[0],
				)

				return writeInitialResidencyManifestForTest(
					t,
					manifest,
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			snapshot, err := LoadInitialResidencySnapshot(
				testCase.path(t),
			)

			if snapshot != nil {
				t.Fatalf(
					"snapshot = %+v, want nil",
					snapshot,
				)
			}

			if err == nil {
				t.Fatal("error = nil, want non-nil")
			}
		})
	}
}
