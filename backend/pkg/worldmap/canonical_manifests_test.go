package worldmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func getTestdataRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test source path")
	}
	// Navigate from backend/pkg/worldmap -> backend/
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	return filepath.Join(backendRoot, "config", "worldmap")
}

func readCanonicalJSON[T any](t *testing.T, filePath string) T {
	t.Helper()
	var result T

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", filePath, err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON from %q: %v", filePath, err)
	}

	return result
}

func TestCanonicalWorldManifest(t *testing.T) {
	worldMapRoot := getTestdataRoot(t)
	manifestPath := filepath.Join(worldMapRoot, "world_manifest.json")

	manifest := readCanonicalJSON[WorldManifest](t, manifestPath)

	if err := ValidateWorldManifest(manifest); err != nil {
		t.Fatalf("ValidateWorldManifest failed: %v", err)
	}

	if manifest.SchemaVersion != SupportedSchemaVersion {
		t.Errorf("world_manifest: schema_version = %d, want %d", manifest.SchemaVersion, SupportedSchemaVersion)
	}
	if manifest.WorldID != CanonicalWorldID {
		t.Errorf("world_manifest: world_id = %q, want %q", manifest.WorldID, CanonicalWorldID)
	}
	if manifest.WorldVersion != 1 {
		t.Errorf("world_manifest: world_version = %d, want 1", manifest.WorldVersion)
	}

	canonicalIDs := CanonicalWorldSpaceIDs()
	if len(manifest.WorldSpaces) != len(canonicalIDs) {
		t.Fatalf("world_manifest: world_spaces count = %d, want %d", len(manifest.WorldSpaces), len(canonicalIDs))
	}

	foundIDs := make(map[WorldSpaceID]bool)
	for _, ref := range manifest.WorldSpaces {
		foundIDs[ref.WorldSpaceID] = true
	}

	for _, id := range canonicalIDs {
		if !foundIDs[id] {
			t.Errorf("world_manifest: missing canonical world_space_id %q", id)
		}
	}
}

func TestCanonicalWorldSpaceManifests(t *testing.T) {
	type expectedWorldSpace struct {
		ID                 WorldSpaceID
		DisplayName        string
		GeographicPosition GeographicPosition
		WidthTiles         int
		HeightTiles        int
		ManifestFile       string
	}

	expectedSpaces := []expectedWorldSpace{
		{WorldSpaceMainContinent, "Main Continent", "central", 12000, 12000, "world_spaces/main_continent.json"},
		{WorldSpaceFireContinent, "Fire Continent", "south", 12000, 12000, "world_spaces/fire_continent.json"},
		{WorldSpaceIceContinent, "Ice Continent", "north", 12000, 12000, "world_spaces/ice_continent.json"},
		{WorldSpaceHolyContinent, "Holy Continent", "east", 12000, 12000, "world_spaces/holy_continent.json"},
		{WorldSpaceShadowContinent, "Shadow Continent", "west", 12000, 12000, "world_spaces/shadow_continent.json"},
		{WorldSpaceNatureContinent, "Nature Continent", "intermediate", 12000, 12000, "world_spaces/nature_continent.json"},
		{WorldSpaceAbyssiaContinent, "Abyssia", "extreme_north", 16000, 16000, "world_spaces/abyssia_continent.json"},
	}

	worldMapRoot := getTestdataRoot(t)
	rootManifest := readCanonicalJSON[WorldManifest](t, filepath.Join(worldMapRoot, "world_manifest.json"))

	if len(rootManifest.WorldSpaces) != len(expectedSpaces) {
		t.Fatalf("mismatch between expected spaces (%d) and manifest references (%d)", len(expectedSpaces), len(rootManifest.WorldSpaces))
	}

	for _, expected := range expectedSpaces {
		t.Run(string(expected.ID), func(t *testing.T) {
			var ref WorldSpaceReference
			foundRef := false
			for _, r := range rootManifest.WorldSpaces {
				if r.WorldSpaceID == expected.ID {
					ref = r
					foundRef = true
					break
				}
			}

			if !foundRef {
				t.Fatalf("world space reference for %q not found in root manifest", expected.ID)
			}

			if ref.ManifestFile != expected.ManifestFile {
				t.Errorf("manifest_file mismatch: got %q, want %q", ref.ManifestFile, expected.ManifestFile)
			}

			// Security check: ensure path is safe before reading
			childPath := filepath.Join(worldMapRoot, filepath.FromSlash(ref.ManifestFile))
			cleanedPath := filepath.Clean(childPath)
			if !strings.HasPrefix(cleanedPath, worldMapRoot) {
				t.Fatalf("unsafe path detected for %q: resolves to %q which is outside the worldmap root", ref.ManifestFile, cleanedPath)
			}

			manifest := readCanonicalJSON[WorldSpaceManifest](t, cleanedPath)

			if err := ValidateWorldSpaceManifest(manifest); err != nil {
				t.Fatalf("ValidateWorldSpaceManifest failed: %v", err)
			}

			// Cross-validate fields
			if manifest.WorldSpaceID != expected.ID {
				t.Errorf("world_space_id: got %q, want %q", manifest.WorldSpaceID, expected.ID)
			}
			if manifest.DisplayName != expected.DisplayName {
				t.Errorf("display_name: got %q, want %q", manifest.DisplayName, expected.DisplayName)
			}
			if manifest.GeographicPosition != expected.GeographicPosition {
				t.Errorf("geographic_position: got %q, want %q", manifest.GeographicPosition, expected.GeographicPosition)
			}
			if manifest.WidthTiles != expected.WidthTiles {
				t.Errorf("width_tiles: got %d, want %d", manifest.WidthTiles, expected.WidthTiles)
			}
			if manifest.HeightTiles != expected.HeightTiles {
				t.Errorf("height_tiles: got %d, want %d", manifest.HeightTiles, expected.HeightTiles)
			}

			// Validate common fields
			if manifest.SchemaVersion != SupportedSchemaVersion {
				t.Errorf("schema_version: got %d, want %d", manifest.SchemaVersion, SupportedSchemaVersion)
			}
			if manifest.WorldID != CanonicalWorldID {
				t.Errorf("world_id: got %q, want %q", manifest.WorldID, CanonicalWorldID)
			}
			if manifest.WorldVersion != 1 {
				t.Errorf("world_version: got %d, want 1", manifest.WorldVersion)
			}
			if manifest.ChunkSize != CanonicalChunkSize {
				t.Errorf("chunk_size: got %d, want %d", manifest.ChunkSize, CanonicalChunkSize)
			}
			if manifest.MinFloor != 0 || manifest.MaxFloor != 15 {
				t.Errorf("floor range: got %d-%d, want 0-15", manifest.MinFloor, manifest.MaxFloor)
			}
			if len(manifest.Chunks) != 0 {
				t.Errorf("chunks array should be empty, but has length %d", len(manifest.Chunks))
			}
		})
	}
}

func TestNoExtraManifests(t *testing.T) {
	worldMapRoot := getTestdataRoot(t)
	worldSpacesDir := filepath.Join(worldMapRoot, "world_spaces")

	entries, err := os.ReadDir(worldSpacesDir)
	if err != nil {
		t.Fatalf("failed to read world_spaces directory: %v", err)
	}

	if len(entries) != 7 {
		t.Errorf("expected exactly 7 manifest files in world_spaces, but found %d", len(entries))
	}
}
