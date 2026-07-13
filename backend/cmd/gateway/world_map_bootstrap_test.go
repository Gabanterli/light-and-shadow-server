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

func canonicalManifestPathForGatewayBootstrapTest(
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
		"world_manifest.json",
	)
}

func writeJSONForGatewayBootstrapTest(
	t *testing.T,
	path string,
	value any,
) {
	t.Helper()

	if err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	); err != nil {
		t.Fatalf("os.MkdirAll failed: %v", err)
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

func readyProductionManifestPathForGatewayTest(
	t *testing.T,
) string {
	t.Helper()

	canonicalSnapshot, err := worldmap.LoadManifestSnapshot(
		canonicalManifestPathForGatewayBootstrapTest(t),
	)
	if err != nil {
		t.Fatalf(
			"LoadManifestSnapshot canonical failed: %v",
			err,
		)
	}

	rootManifest := canonicalSnapshot.RootManifest()
	fixtureRoot := t.TempDir()

	for _, reference := range rootManifest.WorldSpaces {
		manifest, found := canonicalSnapshot.WorldSpace(
			reference.WorldSpaceID,
		)
		if !found {
			t.Fatalf(
				"canonical world space %q not found",
				reference.WorldSpaceID,
			)
		}

		if reference.WorldSpaceID ==
			worldmap.WorldSpaceMainContinent {
			manifest.Chunks = []worldmap.ChunkReference{
				{
					ChunkX:      0,
					ChunkY:      0,
					Z:           0,
					File:        "chunks/main_0_0_0.json",
					ContentHash: "bootstrap-test-hash",
				},
			}
		}

		writeJSONForGatewayBootstrapTest(
			t,
			filepath.Join(
				fixtureRoot,
				filepath.FromSlash(
					reference.ManifestFile,
				),
			),
			manifest,
		)
	}

	rootPath := filepath.Join(
		fixtureRoot,
		"world_manifest.json",
	)

	writeJSONForGatewayBootstrapTest(
		t,
		rootPath,
		rootManifest,
	)

	return rootPath
}

func TestInitializeWorldMapProviderDebugIgnoresManifest(
	t *testing.T,
) {
	missingPath := filepath.Join(
		t.TempDir(),
		"missing-world-manifest.json",
	)

	provider, err := initializeWorldMapProvider(
		worldmap.ModeDebug,
		missingPath,
	)
	if err != nil {
		t.Fatalf(
			"initializeWorldMapProvider debug failed: %v",
			err,
		)
	}

	if provider.Mode() != worldmap.ModeDebug {
		t.Fatalf(
			"provider Mode() = %q, want %q",
			provider.Mode(),
			worldmap.ModeDebug,
		)
	}

	if _, ok := provider.(*worldmap.DebugProvider); !ok {
		t.Fatalf(
			"provider type = %T, want *worldmap.DebugProvider",
			provider,
		)
	}
}

func TestInitializeWorldMapProviderRejectsUnsupportedMode(
	t *testing.T,
) {
	provider, err := initializeWorldMapProvider(
		worldmap.Mode("unsupported"),
		"",
	)

	if provider != nil {
		t.Fatalf(
			"provider type = %T, want nil",
			provider,
		)
	}

	if err == nil {
		t.Fatal("expected unsupported mode error")
	}

	if !strings.Contains(
		err.Error(),
		"unsupported world map mode",
	) {
		t.Fatalf(
			"error = %q, want unsupported mode context",
			err,
		)
	}
}

func TestInitializeWorldMapProviderProductionRejectsEmptyPath(
	t *testing.T,
) {
	provider, err := initializeWorldMapProvider(
		worldmap.ModeProduction,
		"   ",
	)

	if provider != nil {
		t.Fatalf(
			"provider type = %T, want nil",
			provider,
		)
	}

	if err == nil {
		t.Fatal("expected empty manifest path error")
	}
}

func TestInitializeWorldMapProviderProductionRejectsMissingFile(
	t *testing.T,
) {
	missingPath := filepath.Join(
		t.TempDir(),
		"missing-world-manifest.json",
	)

	provider, err := initializeWorldMapProvider(
		worldmap.ModeProduction,
		missingPath,
	)

	if provider != nil {
		t.Fatalf(
			"provider type = %T, want nil",
			provider,
		)
	}

	if err == nil {
		t.Fatal("expected missing manifest error")
	}

	if !strings.Contains(
		err.Error(),
		"load production world map manifest",
	) {
		t.Fatalf(
			"error = %q, want loader context",
			err,
		)
	}
}

func TestInitializeWorldMapProviderAcceptsCanonicalPublishedWorld(
	t *testing.T,
) {
	provider, err := initializeWorldMapProvider(
		worldmap.ModeProduction,
		canonicalManifestPathForGatewayBootstrapTest(t),
	)
	if err != nil {
		t.Fatalf(
			"initialize canonical production world map: %v",
			err,
		)
	}

	if provider == nil {
		t.Fatal("canonical production provider is nil")
	}

	if provider.Mode() != worldmap.ModeProduction {
		t.Fatalf(
			"provider Mode() = %q, want %q",
			provider.Mode(),
			worldmap.ModeProduction,
		)
	}

	productionProvider, ok :=
		provider.(*worldmap.ProductionProvider)
	if !ok {
		t.Fatalf(
			"provider type = %T, want *worldmap.ProductionProvider",
			provider,
		)
	}

	references, err := productionProvider.ChunkReferences(
		worldmap.WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf(
			"main continent ChunkReferences failed: %v",
			err,
		)
	}

	if len(references) != 1 {
		t.Fatalf(
			"main continent reference count = %d, want 1",
			len(references),
		)
	}

	reference := references[0]

	if reference.ChunkX != 3 ||
		reference.ChunkY != 3 ||
		reference.Z != 0 {
		t.Fatalf(
			"canonical chunk coordinate = (%d,%d,%d), want (3,3,0)",
			reference.ChunkX,
			reference.ChunkY,
			reference.Z,
		)
	}

	if reference.File != "chunks/main_continent/3_3_0.json" {
		t.Fatalf(
			"canonical chunk file = %q, want %q",
			reference.File,
			"chunks/main_continent/3_3_0.json",
		)
	}

	if reference.ContentHash !=
		"sha256:81062edbfc2797a9b6f33e20381a2edf1169b56ef189add13992ed97c253bdea" {
		t.Fatalf(
			"canonical chunk hash = %q, want canonical hash",
			reference.ContentHash,
		)
	}
}

func TestInitializeWorldMapProviderRejectsWorldWithoutPublishedChunks(
	t *testing.T,
) {
	canonicalSnapshot, err := worldmap.LoadManifestSnapshot(
		canonicalManifestPathForGatewayBootstrapTest(t),
	)
	if err != nil {
		t.Fatalf(
			"load canonical snapshot for empty fixture: %v",
			err,
		)
	}

	rootManifest := canonicalSnapshot.RootManifest()
	fixtureRoot := t.TempDir()

	for _, reference := range rootManifest.WorldSpaces {
		manifest, found := canonicalSnapshot.WorldSpace(
			reference.WorldSpaceID,
		)
		if !found {
			t.Fatalf(
				"canonical world space %q not found",
				reference.WorldSpaceID,
			)
		}

		manifest.Chunks = []worldmap.ChunkReference{}

		writeJSONForGatewayBootstrapTest(
			t,
			filepath.Join(
				fixtureRoot,
				filepath.FromSlash(reference.ManifestFile),
			),
			manifest,
		)
	}

	rootManifestPath := filepath.Join(
		fixtureRoot,
		"world_manifest.json",
	)

	writeJSONForGatewayBootstrapTest(
		t,
		rootManifestPath,
		rootManifest,
	)

	provider, err := initializeWorldMapProvider(
		worldmap.ModeProduction,
		rootManifestPath,
	)

	if provider != nil {
		t.Fatalf(
			"provider type = %T, want nil",
			provider,
		)
	}

	if !errors.Is(
		err,
		errProductionWorldMapHasNoPublishedChunks,
	) {
		t.Fatalf(
			"error = %v, want no-published-chunks error",
			err,
		)
	}
}
func TestInitializeWorldMapProviderProductionReadyFixture(
	t *testing.T,
) {
	manifestPath :=
		readyProductionManifestPathForGatewayTest(t)

	provider, err := initializeWorldMapProvider(
		worldmap.ModeProduction,
		"  "+manifestPath+"  ",
	)
	if err != nil {
		t.Fatalf(
			"initializeWorldMapProvider production failed: %v",
			err,
		)
	}

	if provider.Mode() != worldmap.ModeProduction {
		t.Fatalf(
			"provider Mode() = %q, want %q",
			provider.Mode(),
			worldmap.ModeProduction,
		)
	}

	productionProvider, ok :=
		provider.(*worldmap.ProductionProvider)
	if !ok {
		t.Fatalf(
			"provider type = %T, want *worldmap.ProductionProvider",
			provider,
		)
	}

	references, err := productionProvider.ChunkReferences(
		worldmap.WorldSpaceMainContinent,
	)
	if err != nil {
		t.Fatalf("ChunkReferences failed: %v", err)
	}

	if len(references) != 1 {
		t.Fatalf(
			"published chunk count = %d, want 1",
			len(references),
		)
	}

	if references[0].ContentHash !=
		"bootstrap-test-hash" {
		t.Fatalf(
			"ContentHash = %q, want %q",
			references[0].ContentHash,
			"bootstrap-test-hash",
		)
	}
}

func TestValidateProductionWorldMapReadyRejectsNilProvider(
	t *testing.T,
) {
	err := validateProductionWorldMapReady(nil)

	if err == nil {
		t.Fatal("expected nil provider error")
	}

	if !strings.Contains(
		err.Error(),
		"provider cannot be nil",
	) {
		t.Fatalf(
			"error = %q, want nil provider context",
			err,
		)
	}
}
