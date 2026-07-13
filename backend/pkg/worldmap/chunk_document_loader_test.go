package worldmap

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

const (
	loaderTestWorldID      = "loader_test_world"
	loaderTestWorldVersion = 1
)

const loaderTestWorldSpaceID WorldSpaceID = "loader_b"

type chunkLoaderTestFixture struct {
	rootDirectory  string
	rootPath       string
	worldSpaceID   WorldSpaceID
	worldSpacePath string
	snapshot       *ManifestSnapshot
}

func defaultChunkLoaderTestReferences() []ChunkReference {
	return []ChunkReference{
		{
			ChunkX: 0,
			ChunkY: 0,
			Z:      0,
			File:   "chunks/0_0_0.json",
		},
		{
			ChunkX: 1,
			ChunkY: 0,
			Z:      0,
			File:   "chunks/1_0_0.json",
		},
		{
			ChunkX: 0,
			ChunkY: 1,
			Z:      1,
			File:   "chunks/0_1_1.json",
		},
	}
}

func newChunkLoaderTestFixture(
	t *testing.T,
	references []ChunkReference,
) *chunkLoaderTestFixture {
	t.Helper()

	if references == nil {
		references =
			defaultChunkLoaderTestReferences()
	}
	references = append(
		[]ChunkReference(nil),
		references...,
	)

	rootDirectory := t.TempDir()
	rootPath := filepath.Join(
		rootDirectory,
		"world_manifest.json",
	)
	worldSpacePath := filepath.Join(
		rootDirectory,
		"world_spaces",
		"loader_b.json",
	)

	mustWriteLoaderTestFile(
		t,
		rootPath,
		[]byte("{}"),
	)
	mustWriteLoaderTestFile(
		t,
		worldSpacePath,
		[]byte("{}"),
	)

	root := WorldManifest{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       loaderTestWorldID,
		WorldVersion:  loaderTestWorldVersion,
		WorldSpaces: []WorldSpaceReference{
			{
				WorldSpaceID: loaderTestWorldSpaceID,
				ManifestFile: "world_spaces/loader_b.json",
			},
		},
	}

	worldSpace := WorldSpaceManifest{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       loaderTestWorldID,
		WorldVersion:  loaderTestWorldVersion,
		WorldSpaceID:  loaderTestWorldSpaceID,
		DisplayName:   "Loader B",
		WidthTiles:    128,
		HeightTiles:   128,
		ChunkSize:     CanonicalChunkSize,
		MinFloor:      0,
		MaxFloor:      15,
		Chunks:        references,
	}

	snapshot := &ManifestSnapshot{
		rootManifest: root,
		worldSpaces: map[WorldSpaceID]WorldSpaceManifest{
			loaderTestWorldSpaceID: worldSpace,
		},
		worldSpaceIDs: []WorldSpaceID{
			loaderTestWorldSpaceID,
		},
	}

	return &chunkLoaderTestFixture{
		rootDirectory:  rootDirectory,
		rootPath:       rootPath,
		worldSpaceID:   loaderTestWorldSpaceID,
		worldSpacePath: worldSpacePath,
		snapshot:       snapshot,
	}
}

func (f *chunkLoaderTestFixture) manifestPath(
	t *testing.T,
	worldSpaceID WorldSpaceID,
) string {
	t.Helper()

	for _, reference := range f.snapshot.rootManifest.WorldSpaces {
		if reference.WorldSpaceID == worldSpaceID {
			return filepath.Join(
				f.rootDirectory,
				filepath.FromSlash(
					reference.ManifestFile,
				),
			)
		}
	}

	t.Fatalf(
		"world space %q has no root reference",
		worldSpaceID,
	)
	return ""
}

func (f *chunkLoaderTestFixture) chunkReference(
	t *testing.T,
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
) ChunkReference {
	t.Helper()

	worldSpace, found :=
		f.snapshot.worldSpaces[worldSpaceID]
	if !found {
		t.Fatalf(
			"world space %q not found in fixture",
			worldSpaceID,
		)
	}

	for _, reference := range worldSpace.Chunks {
		if reference.ChunkX == coordinate.ChunkX &&
			reference.ChunkY == coordinate.ChunkY &&
			reference.Z == coordinate.Z {
			return reference
		}
	}

	t.Fatalf(
		"chunk (%d,%d,%d) not found in fixture world space %q",
		coordinate.ChunkX,
		coordinate.ChunkY,
		coordinate.Z,
		worldSpaceID,
	)
	return ChunkReference{}
}

func (f *chunkLoaderTestFixture) setChunkReference(
	t *testing.T,
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
	mutate func(*ChunkReference),
) {
	t.Helper()

	worldSpace, found :=
		f.snapshot.worldSpaces[worldSpaceID]
	if !found {
		t.Fatalf(
			"world space %q not found in fixture",
			worldSpaceID,
		)
	}

	for index := range worldSpace.Chunks {
		reference := &worldSpace.Chunks[index]
		if reference.ChunkX == coordinate.ChunkX &&
			reference.ChunkY == coordinate.ChunkY &&
			reference.Z == coordinate.Z {
			mutate(reference)
			f.snapshot.worldSpaces[worldSpaceID] =
				worldSpace
			return
		}
	}

	t.Fatalf(
		"chunk (%d,%d,%d) not found for mutation",
		coordinate.ChunkX,
		coordinate.ChunkY,
		coordinate.Z,
	)
}

func (f *chunkLoaderTestFixture) addWorldSpace(
	t *testing.T,
	worldSpaceID WorldSpaceID,
	manifestFile string,
	references []ChunkReference,
) {
	t.Helper()

	manifestPath := filepath.Join(
		f.rootDirectory,
		filepath.FromSlash(manifestFile),
	)
	mustWriteLoaderTestFile(
		t,
		manifestPath,
		[]byte("{}"),
	)

	f.snapshot.rootManifest.WorldSpaces = append(
		f.snapshot.rootManifest.WorldSpaces,
		WorldSpaceReference{
			WorldSpaceID: worldSpaceID,
			ManifestFile: manifestFile,
		},
	)

	f.snapshot.worldSpaces[worldSpaceID] =
		WorldSpaceManifest{
			SchemaVersion: SupportedSchemaVersion,
			WorldID:       loaderTestWorldID,
			WorldVersion:  loaderTestWorldVersion,
			WorldSpaceID:  worldSpaceID,
			DisplayName:   string(worldSpaceID),
			WidthTiles:    128,
			HeightTiles:   128,
			ChunkSize:     CanonicalChunkSize,
			MinFloor:      0,
			MaxFloor:      15,
			Chunks: append(
				[]ChunkReference(nil),
				references...,
			),
		}

	f.snapshot.worldSpaceIDs = append(
		f.snapshot.worldSpaceIDs,
		worldSpaceID,
	)
}

func validChunkLoaderTestDocument(
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
) ChunkDocument {
	return ChunkDocument{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       loaderTestWorldID,
		WorldVersion:  loaderTestWorldVersion,
		WorldSpaceID:  worldSpaceID,
		ChunkX:        coordinate.ChunkX,
		ChunkY:        coordinate.ChunkY,
		Z:             coordinate.Z,
		Width:         CanonicalChunkSize,
		Height:        CanonicalChunkSize,
		TilePalette: []TileDefinition{
			{
				VisualID:     1,
				Walkable:     true,
				MovementCost: 1,
			},
			{
				VisualID:       2,
				BlocksMovement: true,
			},
		},
		Tiles: make(
			[]uint16,
			CanonicalChunkSize*CanonicalChunkSize,
		),
	}
}

func writeChunkLoaderTestDocument(
	t *testing.T,
	fixture *chunkLoaderTestFixture,
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
	document ChunkDocument,
) []byte {
	t.Helper()

	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf(
			"marshal chunk document: %v",
			err,
		)
	}

	writeChunkLoaderTestRaw(
		t,
		fixture,
		worldSpaceID,
		coordinate,
		raw,
	)

	return raw
}

func writeChunkLoaderTestRaw(
	t *testing.T,
	fixture *chunkLoaderTestFixture,
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
	raw []byte,
) string {
	t.Helper()

	reference := fixture.chunkReference(
		t,
		worldSpaceID,
		coordinate,
	)
	manifestPath := fixture.manifestPath(
		t,
		worldSpaceID,
	)
	chunkPath := filepath.Join(
		filepath.Dir(manifestPath),
		filepath.FromSlash(reference.File),
	)

	mustWriteLoaderTestFile(
		t,
		chunkPath,
		raw,
	)

	return chunkPath
}

func mustWriteLoaderTestFile(
	t *testing.T,
	filePath string,
	data []byte,
) {
	t.Helper()

	if err := os.MkdirAll(
		filepath.Dir(filePath),
		0o755,
	); err != nil {
		t.Fatalf(
			"create directory for %q: %v",
			filePath,
			err,
		)
	}

	if err := os.WriteFile(
		filePath,
		data,
		0o644,
	); err != nil {
		t.Fatalf(
			"write %q: %v",
			filePath,
			err,
		)
	}
}

func mustNewChunkDocumentLoader(
	t *testing.T,
	fixture *chunkLoaderTestFixture,
) *ChunkDocumentLoader {
	t.Helper()

	loader, err := NewChunkDocumentLoader(
		fixture.rootPath,
		fixture.snapshot,
	)
	if err != nil {
		t.Fatalf(
			"construct chunk document loader: %v",
			err,
		)
	}

	return loader
}

func loaderTestRequest(
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
) ChunkDocumentRequest {
	return ChunkDocumentRequest{
		WorldSpaceID: worldSpaceID,
		Coordinate:   coordinate,
	}
}

func TestNewChunkDocumentLoaderRejectsInvalidInputs(
	t *testing.T,
) {
	t.Run("empty root path", func(t *testing.T) {
		loader, err := NewChunkDocumentLoader(
			"   ",
			nil,
		)
		if loader != nil {
			t.Fatal(
				"loader must be nil for empty path",
			)
		}
		if !errors.Is(
			err,
			ErrEmptyRootManifestPath,
		) {
			t.Fatalf(
				"got error %v, want ErrEmptyRootManifestPath",
				err,
			)
		}
	})

	t.Run("nil snapshot", func(t *testing.T) {
		rootPath := filepath.Join(
			t.TempDir(),
			"world_manifest.json",
		)
		mustWriteLoaderTestFile(
			t,
			rootPath,
			[]byte("{}"),
		)

		loader, err := NewChunkDocumentLoader(
			rootPath,
			nil,
		)
		if loader != nil {
			t.Fatal(
				"loader must be nil for nil snapshot",
			)
		}
		if !errors.Is(
			err,
			ErrNilManifestSnapshot,
		) {
			t.Fatalf(
				"got error %v, want ErrNilManifestSnapshot",
				err,
			)
		}
	})

	t.Run("missing root manifest", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)

		loader, err := NewChunkDocumentLoader(
			filepath.Join(
				fixture.rootDirectory,
				"missing.json",
			),
			fixture.snapshot,
		)
		if loader != nil || err == nil {
			t.Fatalf(
				"got loader=%v error=%v, want construction failure",
				loader,
				err,
			)
		}
	})

	t.Run("root path is directory", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)

		loader, err := NewChunkDocumentLoader(
			fixture.rootDirectory,
			fixture.snapshot,
		)
		if loader != nil || err == nil {
			t.Fatalf(
				"got loader=%v error=%v, want regular-file failure",
				loader,
				err,
			)
		}
	})

	t.Run("missing world space manifest", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)

		if err := os.Remove(
			fixture.worldSpacePath,
		); err != nil {
			t.Fatalf(
				"remove world-space manifest: %v",
				err,
			)
		}

		loader, err := NewChunkDocumentLoader(
			fixture.rootPath,
			fixture.snapshot,
		)
		if loader != nil || err == nil {
			t.Fatalf(
				"got loader=%v error=%v, want missing-child failure",
				loader,
				err,
			)
		}
	})
}

func TestNewChunkDocumentLoaderValidConstruction(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)

	resolvedRoot, err := filepath.EvalSymlinks(
		fixture.rootDirectory,
	)
	if err != nil {
		t.Fatalf(
			"resolve expected root: %v",
			err,
		)
	}
	resolvedRoot, err = filepath.Abs(
		resolvedRoot,
	)
	if err != nil {
		t.Fatalf(
			"absolute expected root: %v",
			err,
		)
	}
	resolvedRoot = filepath.Clean(
		resolvedRoot,
	)

	if loader.rootDirectory != resolvedRoot {
		t.Fatalf(
			"rootDirectory = %q, want %q",
			loader.rootDirectory,
			resolvedRoot,
		)
	}

	if loader.provider == nil {
		t.Fatal("provider must not be nil")
	}

	if len(loader.worldSpaceDirectories) != 1 {
		t.Fatalf(
			"world-space directory count = %d, want 1",
			len(loader.worldSpaceDirectories),
		)
	}
}

func TestNewChunkDocumentLoaderAcceptsRootSymlink(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	linkPath := filepath.Join(
		fixture.rootDirectory,
		"root_link.json",
	)

	if err := os.Symlink(
		fixture.rootPath,
		linkPath,
	); err != nil {
		t.Skipf(
			"symlink creation unavailable: %v",
			err,
		)
	}

	loader, err := NewChunkDocumentLoader(
		linkPath,
		fixture.snapshot,
	)
	if err != nil {
		t.Fatalf(
			"construct through root symlink: %v",
			err,
		)
	}
	if loader == nil {
		t.Fatal("loader must not be nil")
	}
}

func TestNewChunkDocumentLoaderWorldSpaceSymlinks(
	t *testing.T,
) {
	t.Run("inside root", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)
		targetPath := filepath.Join(
			fixture.rootDirectory,
			"internal",
			"loader_b.json",
		)
		mustWriteLoaderTestFile(
			t,
			targetPath,
			[]byte("{}"),
		)

		if err := os.Remove(
			fixture.worldSpacePath,
		); err != nil {
			t.Fatalf(
				"remove original world-space manifest: %v",
				err,
			)
		}
		if err := os.Symlink(
			targetPath,
			fixture.worldSpacePath,
		); err != nil {
			t.Skipf(
				"symlink creation unavailable: %v",
				err,
			)
		}

		loader, err := NewChunkDocumentLoader(
			fixture.rootPath,
			fixture.snapshot,
		)
		if err != nil {
			t.Fatalf(
				"construct with internal symlink: %v",
				err,
			)
		}
		if loader == nil {
			t.Fatal("loader must not be nil")
		}
	})

	t.Run("outside root", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)
		outsidePath := filepath.Join(
			t.TempDir(),
			"outside_world_space.json",
		)
		mustWriteLoaderTestFile(
			t,
			outsidePath,
			[]byte("{}"),
		)

		if err := os.Remove(
			fixture.worldSpacePath,
		); err != nil {
			t.Fatalf(
				"remove original world-space manifest: %v",
				err,
			)
		}
		if err := os.Symlink(
			outsidePath,
			fixture.worldSpacePath,
		); err != nil {
			t.Skipf(
				"symlink creation unavailable: %v",
				err,
			)
		}

		loader, err := NewChunkDocumentLoader(
			fixture.rootPath,
			fixture.snapshot,
		)
		if loader != nil {
			t.Fatal(
				"loader must be nil for escaping symlink",
			)
		}

		var typed *WorldSpaceManifestPathOutsideRootError
		if !errors.As(err, &typed) {
			t.Fatalf(
				"got error %v, want WorldSpaceManifestPathOutsideRootError",
				err,
			)
		}
	})
}

func TestChunkDocumentLoaderZeroRequests(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)

	documents, err := loader.Load(nil)
	if err != nil {
		t.Fatalf(
			"load zero requests: %v",
			err,
		)
	}
	if documents == nil {
		t.Fatal(
			"zero requests must return a non-nil empty slice",
		)
	}
	if len(documents) != 0 {
		t.Fatalf(
			"zero-request result length = %d, want 0",
			len(documents),
		)
	}
}

func TestChunkDocumentLoaderLoadsValidDocumentWithoutHash(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	coordinate := ChunkCoordinate{}
	document := validChunkLoaderTestDocument(
		fixture.worldSpaceID,
		coordinate,
	)
	document.Tiles[10] = 1

	writeChunkLoaderTestDocument(
		t,
		fixture,
		fixture.worldSpaceID,
		coordinate,
		document,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	documents, err := loader.Load(
		[]ChunkDocumentRequest{
			loaderTestRequest(
				fixture.worldSpaceID,
				coordinate,
			),
		},
	)
	if err != nil {
		t.Fatalf(
			"load valid document: %v",
			err,
		)
	}

	if len(documents) != 1 {
		t.Fatalf(
			"document count = %d, want 1",
			len(documents),
		)
	}
	if !reflect.DeepEqual(
		documents[0],
		document,
	) {
		t.Fatalf(
			"loaded document differs from written document",
		)
	}
}

func TestChunkDocumentLoaderValidatesRawSHA256(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	coordinate := ChunkCoordinate{}
	document := validChunkLoaderTestDocument(
		fixture.worldSpaceID,
		coordinate,
	)

	raw, err := json.MarshalIndent(
		document,
		"",
		"  ",
	)
	if err != nil {
		t.Fatalf(
			"marshal indented document: %v",
			err,
		)
	}
	raw = append(raw, '\n')

	sum := sha256.Sum256(raw)
	contentHash := fmt.Sprintf(
		"sha256:%x",
		sum,
	)

	fixture.setChunkReference(
		t,
		fixture.worldSpaceID,
		coordinate,
		func(reference *ChunkReference) {
			reference.ContentHash = contentHash
		},
	)
	writeChunkLoaderTestRaw(
		t,
		fixture,
		fixture.worldSpaceID,
		coordinate,
		raw,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	documents, err := loader.Load(
		[]ChunkDocumentRequest{
			loaderTestRequest(
				fixture.worldSpaceID,
				coordinate,
			),
		},
	)
	if err != nil {
		t.Fatalf(
			"load correctly hashed document: %v",
			err,
		)
	}
	if len(documents) != 1 {
		t.Fatalf(
			"document count = %d, want 1",
			len(documents),
		)
	}
}

func TestChunkDocumentLoaderRejectsInvalidHashFormats(
	t *testing.T,
) {
	validHex := strings.Repeat("a", 64)

	testCases := []struct {
		name        string
		contentHash string
	}{
		{
			name:        "missing prefix",
			contentHash: validHex,
		},
		{
			name:        "wrong prefix",
			contentHash: "md5:" + validHex,
		},
		{
			name:        "short digest",
			contentHash: "sha256:" + validHex[:63],
		},
		{
			name:        "long digest",
			contentHash: "sha256:" + validHex + "a",
		},
		{
			name: "uppercase",
			contentHash: "sha256:" +
				strings.Repeat("A", 64),
		},
		{
			name: "invalid hexadecimal",
			contentHash: "sha256:" +
				strings.Repeat("g", 64),
		},
		{
			name:        "leading whitespace",
			contentHash: " sha256:" + validHex,
		},
		{
			name:        "trailing whitespace",
			contentHash: "sha256:" + validHex + " ",
		},
		{
			name: "duplicate prefix",
			contentHash: "sha256:sha256:" +
				validHex[:57],
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			coordinate := ChunkCoordinate{}
			document := validChunkLoaderTestDocument(
				fixture.worldSpaceID,
				coordinate,
			)
			writeChunkLoaderTestDocument(
				t,
				fixture,
				fixture.worldSpaceID,
				coordinate,
				document,
			)
			fixture.setChunkReference(
				t,
				fixture.worldSpaceID,
				coordinate,
				func(reference *ChunkReference) {
					reference.ContentHash =
						testCase.contentHash
				},
			)

			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)
			documents, err := loader.Load(
				[]ChunkDocumentRequest{
					loaderTestRequest(
						fixture.worldSpaceID,
						coordinate,
					),
				},
			)

			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}

			var typed *InvalidChunkContentHashError
			if !errors.As(err, &typed) {
				t.Fatalf(
					"got error %v, want InvalidChunkContentHashError",
					err,
				)
			}
			if typed.ContentHash !=
				testCase.contentHash {
				t.Fatalf(
					"ContentHash = %q, want %q",
					typed.ContentHash,
					testCase.contentHash,
				)
			}
		})
	}
}

func TestChunkDocumentLoaderRejectsHashMismatch(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	coordinate := ChunkCoordinate{}
	document := validChunkLoaderTestDocument(
		fixture.worldSpaceID,
		coordinate,
	)
	raw := writeChunkLoaderTestDocument(
		t,
		fixture,
		fixture.worldSpaceID,
		coordinate,
		document,
	)

	actualSum := sha256.Sum256(raw)
	actual := fmt.Sprintf(
		"sha256:%x",
		actualSum,
	)
	expected := "sha256:" +
		strings.Repeat("0", 64)

	fixture.setChunkReference(
		t,
		fixture.worldSpaceID,
		coordinate,
		func(reference *ChunkReference) {
			reference.ContentHash = expected
		},
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	documents, err := loader.Load(
		[]ChunkDocumentRequest{
			loaderTestRequest(
				fixture.worldSpaceID,
				coordinate,
			),
		},
	)

	if documents != nil {
		t.Fatalf(
			"documents = %#v, want nil",
			documents,
		)
	}

	var typed *ChunkContentHashMismatchError
	if !errors.As(err, &typed) {
		t.Fatalf(
			"got error %v, want ChunkContentHashMismatchError",
			err,
		)
	}
	if typed.Expected != expected ||
		typed.Actual != actual {
		t.Fatalf(
			"unexpected hash mismatch context: %+v",
			typed,
		)
	}
}

func TestReadBoundedChunkDocumentSizeLimit(
	t *testing.T,
) {
	request := loaderTestRequest(
		loaderTestWorldSpaceID,
		ChunkCoordinate{},
	)

	t.Run("exact limit is not too large", func(t *testing.T) {
		filePath := filepath.Join(
			t.TempDir(),
			"exact_limit.bin",
		)
		mustWriteLoaderTestFile(
			t,
			filePath,
			bytes.Repeat(
				[]byte{'a'},
				int(maxChunkDocumentBytes),
			),
		)

		raw, err := readBoundedChunkDocument(
			filePath,
			request,
		)
		if err != nil {
			t.Fatalf(
				"read exact-limit file: %v",
				err,
			)
		}
		if int64(len(raw)) !=
			maxChunkDocumentBytes {
			t.Fatalf(
				"read %d bytes, want %d",
				len(raw),
				maxChunkDocumentBytes,
			)
		}
	})

	t.Run("one byte over limit is rejected", func(t *testing.T) {
		filePath := filepath.Join(
			t.TempDir(),
			"over_limit.bin",
		)
		mustWriteLoaderTestFile(
			t,
			filePath,
			bytes.Repeat(
				[]byte{'b'},
				int(maxChunkDocumentBytes+1),
			),
		)

		raw, err := readBoundedChunkDocument(
			filePath,
			request,
		)
		if raw != nil {
			t.Fatalf(
				"raw length = %d, want nil",
				len(raw),
			)
		}

		var typed *ChunkDocumentTooLargeError
		if !errors.As(err, &typed) {
			t.Fatalf(
				"got error %v, want ChunkDocumentTooLargeError",
				err,
			)
		}
	})
}

func TestChunkDocumentLoaderStrictJSON(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{}
	baseDocument := validChunkLoaderTestDocument(
		loaderTestWorldSpaceID,
		coordinate,
	)
	baseRaw, err := json.Marshal(baseDocument)
	if err != nil {
		t.Fatalf(
			"marshal base document: %v",
			err,
		)
	}

	unknownField := append(
		[]byte(nil),
		baseRaw[:len(baseRaw)-1]...,
	)
	unknownField = append(
		unknownField,
		[]byte(`,"unknown_field":true}`)...,
	)

	wrongType := bytes.Replace(
		baseRaw,
		[]byte(`"width":32`),
		[]byte(`"width":"32"`),
		1,
	)
	if bytes.Equal(wrongType, baseRaw) {
		t.Fatal(
			"failed to prepare wrong-type fixture",
		)
	}

	testCases := []struct {
		name string
		raw  []byte
	}{
		{
			name: "unknown field",
			raw:  unknownField,
		},
		{
			name: "multiple documents",
			raw: append(
				append([]byte(nil), baseRaw...),
				[]byte("\n{}")...,
			),
		},
		{
			name: "trailing token",
			raw: append(
				append([]byte(nil), baseRaw...),
				[]byte(" trailing")...,
			),
		},
		{
			name: "truncated JSON",
			raw: []byte(
				`{"schema_version":`,
			),
		},
		{
			name: "wrong field type",
			raw:  wrongType,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			writeChunkLoaderTestRaw(
				t,
				fixture,
				fixture.worldSpaceID,
				coordinate,
				testCase.raw,
			)

			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)
			documents, err := loader.Load(
				[]ChunkDocumentRequest{
					loaderTestRequest(
						fixture.worldSpaceID,
						coordinate,
					),
				},
			)

			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}
			if err == nil {
				t.Fatal(
					"expected strict JSON failure",
				)
			}
		})
	}
}

func TestChunkDocumentLoaderStructuralValidation(
	t *testing.T,
) {
	testCases := []struct {
		name     string
		mutate   func(*ChunkDocument)
		wantText string
	}{
		{
			name: "empty palette",
			mutate: func(document *ChunkDocument) {
				document.TilePalette = nil
			},
			wantText: "tile_palette cannot be empty",
		},
		{
			name: "wrong tile count",
			mutate: func(document *ChunkDocument) {
				document.Tiles =
					document.Tiles[:100]
			},
			wantText: "tiles array length",
		},
		{
			name: "palette index out of range",
			mutate: func(document *ChunkDocument) {
				document.Tiles[0] = 99
			},
			wantText: "palette index",
		},
		{
			name: "walkable and blocking",
			mutate: func(document *ChunkDocument) {
				document.TilePalette[0].
					BlocksMovement = true
			},
			wantText: "both walkable and block movement",
		},
		{
			name: "invalid width",
			mutate: func(document *ChunkDocument) {
				document.Width = 31
			},
			wantText: "chunk dimensions must be",
		},
		{
			name: "invalid height",
			mutate: func(document *ChunkDocument) {
				document.Height = 31
			},
			wantText: "chunk dimensions must be",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			coordinate := ChunkCoordinate{}
			document := validChunkLoaderTestDocument(
				fixture.worldSpaceID,
				coordinate,
			)
			testCase.mutate(&document)
			writeChunkLoaderTestDocument(
				t,
				fixture,
				fixture.worldSpaceID,
				coordinate,
				document,
			)

			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)
			documents, err := loader.Load(
				[]ChunkDocumentRequest{
					loaderTestRequest(
						fixture.worldSpaceID,
						coordinate,
					),
				},
			)

			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}
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

			var identityError *ChunkDocumentIdentityMismatchError
			if errors.As(err, &identityError) {
				t.Fatalf(
					"structural error unexpectedly became identity error: %v",
					err,
				)
			}
		})
	}
}

func TestChunkDocumentLoaderIdentityValidation(
	t *testing.T,
) {
	testCases := []struct {
		name      string
		field     string
		mutate    func(*ChunkDocument)
		wantValue string
	}{
		{
			name:  "world id",
			field: "world_id",
			mutate: func(document *ChunkDocument) {
				document.WorldID = "other_world"
			},
			wantValue: "other_world",
		},
		{
			name:  "world version",
			field: "world_version",
			mutate: func(document *ChunkDocument) {
				document.WorldVersion = 2
			},
			wantValue: "2",
		},
		{
			name:  "world space id",
			field: "world_space_id",
			mutate: func(document *ChunkDocument) {
				document.WorldSpaceID =
					"other_space"
			},
			wantValue: "other_space",
		},
		{
			name:  "chunk x",
			field: "chunk_x",
			mutate: func(document *ChunkDocument) {
				document.ChunkX = 7
			},
			wantValue: "7",
		},
		{
			name:  "chunk y",
			field: "chunk_y",
			mutate: func(document *ChunkDocument) {
				document.ChunkY = 8
			},
			wantValue: "8",
		},
		{
			name:  "z",
			field: "z",
			mutate: func(document *ChunkDocument) {
				document.Z = 9
			},
			wantValue: "9",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			coordinate := ChunkCoordinate{}
			document := validChunkLoaderTestDocument(
				fixture.worldSpaceID,
				coordinate,
			)
			testCase.mutate(&document)
			writeChunkLoaderTestDocument(
				t,
				fixture,
				fixture.worldSpaceID,
				coordinate,
				document,
			)

			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)
			documents, err := loader.Load(
				[]ChunkDocumentRequest{
					loaderTestRequest(
						fixture.worldSpaceID,
						coordinate,
					),
				},
			)

			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}

			var typed *ChunkDocumentIdentityMismatchError
			if !errors.As(err, &typed) {
				t.Fatalf(
					"got error %v, want ChunkDocumentIdentityMismatchError",
					err,
				)
			}
			if typed.Field != testCase.field ||
				typed.Actual != testCase.wantValue {
				t.Fatalf(
					"unexpected identity context: %+v",
					typed,
				)
			}
		})
	}
}

func TestChunkDocumentLoaderPublicationErrors(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)

	t.Run("unknown world space", func(t *testing.T) {
		documents, err := loader.Load(
			[]ChunkDocumentRequest{
				loaderTestRequest(
					"missing_space",
					ChunkCoordinate{},
				),
			},
		)
		if documents != nil {
			t.Fatalf(
				"documents = %#v, want nil",
				documents,
			)
		}

		var typed *WorldSpaceNotFoundError
		if !errors.As(err, &typed) {
			t.Fatalf(
				"got error %v, want WorldSpaceNotFoundError",
				err,
			)
		}
	})

	t.Run("unpublished chunk", func(t *testing.T) {
		documents, err := loader.Load(
			[]ChunkDocumentRequest{
				loaderTestRequest(
					fixture.worldSpaceID,
					ChunkCoordinate{
						ChunkX: 99,
						ChunkY: 99,
						Z:      0,
					},
				),
			},
		)
		if documents != nil {
			t.Fatalf(
				"documents = %#v, want nil",
				documents,
			)
		}

		var typed *ChunkReferenceNotFoundError
		if !errors.As(err, &typed) {
			t.Fatalf(
				"got error %v, want ChunkReferenceNotFoundError",
				err,
			)
		}
	})
}

func TestChunkDocumentLoaderRejectsDuplicateRequests(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{}

	testCases := []struct {
		name     string
		requests []ChunkDocumentRequest
	}{
		{
			name: "adjacent",
			requests: []ChunkDocumentRequest{
				loaderTestRequest(
					loaderTestWorldSpaceID,
					coordinate,
				),
				loaderTestRequest(
					loaderTestWorldSpaceID,
					coordinate,
				),
			},
		},
		{
			name: "separated",
			requests: []ChunkDocumentRequest{
				loaderTestRequest(
					loaderTestWorldSpaceID,
					coordinate,
				),
				loaderTestRequest(
					loaderTestWorldSpaceID,
					ChunkCoordinate{
						ChunkX: 1,
					},
				),
				loaderTestRequest(
					loaderTestWorldSpaceID,
					coordinate,
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)

			documents, err := loader.Load(
				testCase.requests,
			)
			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}

			var typed *DuplicateChunkDocumentRequestError
			if !errors.As(err, &typed) {
				t.Fatalf(
					"got error %v, want DuplicateChunkDocumentRequestError",
					err,
				)
			}
		})
	}
}

func TestChunkDocumentLoaderDeterministicOrderAndInputImmutability(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)

	worldSpaceA := WorldSpaceID("loader_a")
	referenceA := ChunkReference{
		ChunkX: 2,
		ChunkY: 1,
		Z:      0,
		File:   "chunks_a/2_1_0.json",
	}
	fixture.addWorldSpace(
		t,
		worldSpaceA,
		"world_spaces/loader_a.json",
		[]ChunkReference{
			referenceA,
		},
	)

	coordinatesB := []ChunkCoordinate{
		{
			ChunkX: 0,
			ChunkY: 0,
			Z:      0,
		},
		{
			ChunkX: 1,
			ChunkY: 0,
			Z:      0,
		},
		{
			ChunkX: 0,
			ChunkY: 1,
			Z:      1,
		},
	}
	coordinateA := ChunkCoordinate{
		ChunkX: 2,
		ChunkY: 1,
		Z:      0,
	}

	for _, coordinate := range coordinatesB {
		writeChunkLoaderTestDocument(
			t,
			fixture,
			fixture.worldSpaceID,
			coordinate,
			validChunkLoaderTestDocument(
				fixture.worldSpaceID,
				coordinate,
			),
		)
	}
	writeChunkLoaderTestDocument(
		t,
		fixture,
		worldSpaceA,
		coordinateA,
		validChunkLoaderTestDocument(
			worldSpaceA,
			coordinateA,
		),
	)

	requests := []ChunkDocumentRequest{
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[2],
		),
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[1],
		),
		loaderTestRequest(
			worldSpaceA,
			coordinateA,
		),
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[0],
		),
	}
	originalRequests := append(
		[]ChunkDocumentRequest(nil),
		requests...,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	documents, err := loader.Load(requests)
	if err != nil {
		t.Fatalf(
			"load out-of-order requests: %v",
			err,
		)
	}

	if !reflect.DeepEqual(
		requests,
		originalRequests,
	) {
		t.Fatalf(
			"caller request slice was modified: got %+v, want %+v",
			requests,
			originalRequests,
		)
	}

	wantOrder := []ChunkDocumentRequest{
		loaderTestRequest(
			worldSpaceA,
			coordinateA,
		),
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[0],
		),
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[1],
		),
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinatesB[2],
		),
	}

	if len(documents) != len(wantOrder) {
		t.Fatalf(
			"document count = %d, want %d",
			len(documents),
			len(wantOrder),
		)
	}

	for index, want := range wantOrder {
		got := documents[index]
		if got.WorldSpaceID != want.WorldSpaceID ||
			got.ChunkX != want.Coordinate.ChunkX ||
			got.ChunkY != want.Coordinate.ChunkY ||
			got.Z != want.Coordinate.Z {
			t.Fatalf(
				"documents[%d] identity = %q (%d,%d,%d), want %q (%d,%d,%d)",
				index,
				got.WorldSpaceID,
				got.ChunkX,
				got.ChunkY,
				got.Z,
				want.WorldSpaceID,
				want.Coordinate.ChunkX,
				want.Coordinate.ChunkY,
				want.Coordinate.Z,
			)
		}
	}
}

func TestChunkDocumentLoaderAtomicFailure(
	t *testing.T,
) {
	references := []ChunkReference{
		{
			ChunkX: 0,
			ChunkY: 0,
			Z:      0,
			File:   "chunks/0.json",
		},
		{
			ChunkX: 1,
			ChunkY: 0,
			Z:      0,
			File:   "chunks/1.json",
		},
		{
			ChunkX: 2,
			ChunkY: 0,
			Z:      0,
			File:   "chunks/2.json",
		},
	}
	fixture := newChunkLoaderTestFixture(
		t,
		references,
	)

	for _, chunkX := range []int{0, 2} {
		coordinate := ChunkCoordinate{
			ChunkX: chunkX,
		}
		writeChunkLoaderTestDocument(
			t,
			fixture,
			fixture.worldSpaceID,
			coordinate,
			validChunkLoaderTestDocument(
				fixture.worldSpaceID,
				coordinate,
			),
		)
	}

	writeChunkLoaderTestRaw(
		t,
		fixture,
		fixture.worldSpaceID,
		ChunkCoordinate{
			ChunkX: 1,
		},
		[]byte(`{"schema_version":`),
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	documents, err := loader.Load(
		[]ChunkDocumentRequest{
			loaderTestRequest(
				fixture.worldSpaceID,
				ChunkCoordinate{
					ChunkX: 0,
				},
			),
			loaderTestRequest(
				fixture.worldSpaceID,
				ChunkCoordinate{
					ChunkX: 1,
				},
			),
			loaderTestRequest(
				fixture.worldSpaceID,
				ChunkCoordinate{
					ChunkX: 2,
				},
			),
		},
	)

	if documents != nil {
		t.Fatalf(
			"partial documents escaped: %#v",
			documents,
		)
	}
	if err == nil {
		t.Fatal(
			"expected middle-request failure",
		)
	}
}

func TestChunkDocumentLoaderRejectsUnsafeAssetPaths(
	t *testing.T,
) {
	testCases := []struct {
		name string
		path func(t *testing.T) string
	}{
		{
			name: "parent traversal",
			path: func(t *testing.T) string {
				return "../outside.json"
			},
		},
		{
			name: "absolute",
			path: func(t *testing.T) string {
				return filepath.Join(
					t.TempDir(),
					"outside.json",
				)
			},
		},
		{
			name: "backslash",
			path: func(t *testing.T) string {
				return `chunks\outside.json`
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newChunkLoaderTestFixture(
				t,
				nil,
			)
			coordinate := ChunkCoordinate{}
			fixture.setChunkReference(
				t,
				fixture.worldSpaceID,
				coordinate,
				func(reference *ChunkReference) {
					reference.File =
						testCase.path(t)
				},
			)

			loader := mustNewChunkDocumentLoader(
				t,
				fixture,
			)
			documents, err := loader.Load(
				[]ChunkDocumentRequest{
					loaderTestRequest(
						fixture.worldSpaceID,
						coordinate,
					),
				},
			)

			if documents != nil {
				t.Fatalf(
					"documents = %#v, want nil",
					documents,
				)
			}
			if err == nil {
				t.Fatal(
					"expected unsafe-path rejection",
				)
			}
		})
	}
}

func TestChunkDocumentLoaderChunkSymlinks(
	t *testing.T,
) {
	t.Run("inside root", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)
		coordinate := ChunkCoordinate{}
		reference := fixture.chunkReference(
			t,
			fixture.worldSpaceID,
			coordinate,
		)
		manifestPath := fixture.manifestPath(
			t,
			fixture.worldSpaceID,
		)
		linkPath := filepath.Join(
			filepath.Dir(manifestPath),
			filepath.FromSlash(reference.File),
		)
		targetPath := filepath.Join(
			fixture.rootDirectory,
			"internal_chunks",
			"actual.json",
		)
		document := validChunkLoaderTestDocument(
			fixture.worldSpaceID,
			coordinate,
		)
		raw, err := json.Marshal(document)
		if err != nil {
			t.Fatalf(
				"marshal document: %v",
				err,
			)
		}
		mustWriteLoaderTestFile(
			t,
			targetPath,
			raw,
		)
		if err := os.MkdirAll(
			filepath.Dir(linkPath),
			0o755,
		); err != nil {
			t.Fatalf(
				"create link directory: %v",
				err,
			)
		}
		if err := os.Symlink(
			targetPath,
			linkPath,
		); err != nil {
			t.Skipf(
				"symlink creation unavailable: %v",
				err,
			)
		}

		loader := mustNewChunkDocumentLoader(
			t,
			fixture,
		)
		documents, err := loader.Load(
			[]ChunkDocumentRequest{
				loaderTestRequest(
					fixture.worldSpaceID,
					coordinate,
				),
			},
		)
		if err != nil {
			t.Fatalf(
				"load internal chunk symlink: %v",
				err,
			)
		}
		if len(documents) != 1 {
			t.Fatalf(
				"document count = %d, want 1",
				len(documents),
			)
		}
	})

	t.Run("outside root", func(t *testing.T) {
		fixture := newChunkLoaderTestFixture(
			t,
			nil,
		)
		coordinate := ChunkCoordinate{}
		reference := fixture.chunkReference(
			t,
			fixture.worldSpaceID,
			coordinate,
		)
		manifestPath := fixture.manifestPath(
			t,
			fixture.worldSpaceID,
		)
		linkPath := filepath.Join(
			filepath.Dir(manifestPath),
			filepath.FromSlash(reference.File),
		)
		targetPath := filepath.Join(
			t.TempDir(),
			"outside.json",
		)
		document := validChunkLoaderTestDocument(
			fixture.worldSpaceID,
			coordinate,
		)
		raw, err := json.Marshal(document)
		if err != nil {
			t.Fatalf(
				"marshal document: %v",
				err,
			)
		}
		mustWriteLoaderTestFile(
			t,
			targetPath,
			raw,
		)
		if err := os.MkdirAll(
			filepath.Dir(linkPath),
			0o755,
		); err != nil {
			t.Fatalf(
				"create link directory: %v",
				err,
			)
		}
		if err := os.Symlink(
			targetPath,
			linkPath,
		); err != nil {
			t.Skipf(
				"symlink creation unavailable: %v",
				err,
			)
		}

		loader := mustNewChunkDocumentLoader(
			t,
			fixture,
		)
		documents, err := loader.Load(
			[]ChunkDocumentRequest{
				loaderTestRequest(
					fixture.worldSpaceID,
					coordinate,
				),
			},
		)
		if documents != nil {
			t.Fatalf(
				"documents = %#v, want nil",
				documents,
			)
		}

		var typed *ChunkDocumentPathOutsideRootError
		if !errors.As(err, &typed) {
			t.Fatalf(
				"got error %v, want ChunkDocumentPathOutsideRootError",
				err,
			)
		}
	})
}

func TestChunkDocumentLoaderReturnsIndependentDocuments(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	coordinate := ChunkCoordinate{}
	document := validChunkLoaderTestDocument(
		fixture.worldSpaceID,
		coordinate,
	)
	writeChunkLoaderTestDocument(
		t,
		fixture,
		fixture.worldSpaceID,
		coordinate,
		document,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	requests := []ChunkDocumentRequest{
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinate,
		),
	}

	first, err := loader.Load(requests)
	if err != nil {
		t.Fatalf(
			"first load: %v",
			err,
		)
	}
	first[0].WorldID = "mutated"
	first[0].TilePalette[0].VisualID = 999
	first[0].Tiles[0] = 1

	second, err := loader.Load(requests)
	if err != nil {
		t.Fatalf(
			"second load: %v",
			err,
		)
	}

	if second[0].WorldID != loaderTestWorldID {
		t.Fatalf(
			"second WorldID = %q, want %q",
			second[0].WorldID,
			loaderTestWorldID,
		)
	}
	if second[0].TilePalette[0].VisualID != 1 {
		t.Fatalf(
			"second palette VisualID = %d, want 1",
			second[0].TilePalette[0].VisualID,
		)
	}
	if second[0].Tiles[0] != 0 {
		t.Fatalf(
			"second tile index = %d, want 0",
			second[0].Tiles[0],
		)
	}
}

func TestChunkDocumentLoaderConcurrentLoads(
	t *testing.T,
) {
	fixture := newChunkLoaderTestFixture(
		t,
		nil,
	)
	coordinate := ChunkCoordinate{}
	document := validChunkLoaderTestDocument(
		fixture.worldSpaceID,
		coordinate,
	)
	writeChunkLoaderTestDocument(
		t,
		fixture,
		fixture.worldSpaceID,
		coordinate,
		document,
	)

	loader := mustNewChunkDocumentLoader(
		t,
		fixture,
	)
	requests := []ChunkDocumentRequest{
		loaderTestRequest(
			fixture.worldSpaceID,
			coordinate,
		),
	}

	const goroutineCount = 32
	const loadsPerGoroutine = 25

	var waitGroup sync.WaitGroup
	errorChannel := make(
		chan error,
		goroutineCount,
	)

	for goroutineIndex := 0; goroutineIndex < goroutineCount; goroutineIndex++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			for loadIndex := 0; loadIndex < loadsPerGoroutine; loadIndex++ {
				documents, err := loader.Load(
					requests,
				)
				if err != nil {
					errorChannel <- err
					return
				}
				if len(documents) != 1 {
					errorChannel <- fmt.Errorf(
						"document count = %d, want 1",
						len(documents),
					)
					return
				}
				if documents[0].WorldID !=
					loaderTestWorldID {
					errorChannel <- fmt.Errorf(
						"WorldID = %q, want %q",
						documents[0].WorldID,
						loaderTestWorldID,
					)
					return
				}
			}
		}()
	}

	waitGroup.Wait()
	close(errorChannel)

	for err := range errorChannel {
		t.Fatalf(
			"concurrent load failed: %v",
			err,
		)
	}
}
