package worldmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSchemaConstants(t *testing.T) {
	t.Parallel()

	if SupportedSchemaVersion != 1 {
		t.Fatalf(
			"SupportedSchemaVersion = %d, want 1",
			SupportedSchemaVersion,
		)
	}

	if CanonicalChunkSize != 32 {
		t.Fatalf(
			"CanonicalChunkSize = %d, want 32",
			CanonicalChunkSize,
		)
	}

	if CanonicalWorldID != "light_and_shadow_world" {
		t.Fatalf(
			"CanonicalWorldID = %q, want %q",
			CanonicalWorldID,
			"light_and_shadow_world",
		)
	}
}

func TestCanonicalWorldSpaceIDs(t *testing.T) {
	t.Parallel()

	want := [7]WorldSpaceID{
		"main_continent",
		"fire_continent",
		"ice_continent",
		"holy_continent",
		"shadow_continent",
		"nature_continent",
		"abyssia_continent",
	}

	got := CanonicalWorldSpaceIDs()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CanonicalWorldSpaceIDs() = %#v, want %#v",
			got,
			want,
		)
	}

	seen := make(map[WorldSpaceID]struct{}, len(got))

	for _, worldSpaceID := range got {
		if _, exists := seen[worldSpaceID]; exists {
			t.Fatalf("duplicate canonical world space ID %q", worldSpaceID)
		}

		seen[worldSpaceID] = struct{}{}
	}

	if len(seen) != 7 {
		t.Fatalf("canonical world space count = %d, want 7", len(seen))
	}
}

func TestWorldPositionJSONMarshal(t *testing.T) {
	t.Parallel()

	position := WorldPosition{
		WorldSpaceID: WorldSpaceMainContinent,
		X:            123,
		Y:            456,
		Z:            7,
	}

	encoded, err := json.Marshal(position)
	if err != nil {
		t.Fatalf("marshal WorldPosition: %v", err)
	}

	var document map[string]any

	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode WorldPosition JSON: %v", err)
	}

	expectedKeys := []string{
		"world_space_id",
		"x",
		"y",
		"z",
	}

	if len(document) != len(expectedKeys) {
		t.Fatalf(
			"WorldPosition JSON key count = %d, want %d: %s",
			len(document),
			len(expectedKeys),
			encoded,
		)
	}

	for _, key := range expectedKeys {
		if _, exists := document[key]; !exists {
			t.Fatalf(
				"WorldPosition JSON missing key %q: %s",
				key,
				encoded,
			)
		}
	}

	if document["world_space_id"] != "main_continent" {
		t.Fatalf(
			"world_space_id = %#v, want %q",
			document["world_space_id"],
			"main_continent",
		)
	}
}

func TestWorldSpaceManifestJSONMarshal(t *testing.T) {
	t.Parallel()

	manifest := WorldSpaceManifest{
		SchemaVersion:      SupportedSchemaVersion,
		WorldID:            CanonicalWorldID,
		WorldVersion:       1,
		WorldSpaceID:       WorldSpaceAbyssiaContinent,
		DisplayName:        "Abyssia",
		GeographicPosition: GeographicPositionExtremeNorth,
		WidthTiles:         16000,
		HeightTiles:        16000,
		ChunkSize:          CanonicalChunkSize,
		MinFloor:           0,
		MaxFloor:           15,
		Chunks:             []ChunkReference{},
	}

	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal WorldSpaceManifest: %v", err)
	}

	var document map[string]any

	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode WorldSpaceManifest JSON: %v", err)
	}

	expectedKeys := []string{
		"schema_version",
		"world_id",
		"world_version",
		"world_space_id",
		"display_name",
		"geographic_position",
		"width_tiles",
		"height_tiles",
		"chunk_size",
		"min_floor",
		"max_floor",
		"chunks",
	}

	if len(document) != len(expectedKeys) {
		t.Fatalf(
			"WorldSpaceManifest JSON key count = %d, want %d: %s",
			len(document),
			len(expectedKeys),
			encoded,
		)
	}

	for _, key := range expectedKeys {
		if _, exists := document[key]; !exists {
			t.Fatalf(
				"WorldSpaceManifest JSON missing key %q: %s",
				key,
				encoded,
			)
		}
	}
}

func TestChunkDocumentJSONRoundTrip(t *testing.T) {
	t.Parallel()

	document := ChunkDocument{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		WorldSpaceID:  WorldSpaceMainContinent,
		ChunkX:        10,
		ChunkY:        20,
		Z:             0,
		Width:         CanonicalChunkSize,
		Height:        CanonicalChunkSize,
		TilePalette: []TileDefinition{
			{
				VisualID:          101,
				Walkable:          true,
				BlocksMovement:    false,
				BlocksProjectiles: false,
				SafeZone:          false,
				MovementCost:      1,
			},
		},
		Tiles: make([]uint16, CanonicalChunkSize*CanonicalChunkSize),
	}

	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal ChunkDocument: %v", err)
	}

	var decoded ChunkDocument

	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal ChunkDocument: %v", err)
	}

	if !reflect.DeepEqual(decoded, document) {
		t.Fatalf(
			"ChunkDocument round trip mismatch:\ngot:  %#v\nwant: %#v",
			decoded,
			document,
		)
	}
}

func TestChunkReferenceOmitsEmptyContentHash(t *testing.T) {
	t.Parallel()

	reference := ChunkReference{
		ChunkX: 1,
		ChunkY: 2,
		Z:      0,
		File:   "chunks/1_2_0.json",
	}

	encoded, err := json.Marshal(reference)
	if err != nil {
		t.Fatalf("marshal ChunkReference: %v", err)
	}

	var document map[string]any

	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode ChunkReference JSON: %v", err)
	}

	if _, exists := document["content_hash"]; exists {
		t.Fatalf(
			"empty content_hash must be omitted: %s",
			encoded,
		)
	}

	if document["file"] != "chunks/1_2_0.json" {
		t.Fatalf(
			"file = %#v, want %q",
			document["file"],
			"chunks/1_2_0.json",
		)
	}
}

func TestChunkReferenceIncludesContentHash(t *testing.T) {
	t.Parallel()

	reference := ChunkReference{
		ChunkX:      1,
		ChunkY:      2,
		Z:           0,
		File:        "chunks/1_2_0.json",
		ContentHash: "sha256:example",
	}

	encoded, err := json.Marshal(reference)
	if err != nil {
		t.Fatalf("marshal ChunkReference: %v", err)
	}

	var document map[string]any

	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode ChunkReference JSON: %v", err)
	}

	if document["content_hash"] != "sha256:example" {
		t.Fatalf(
			"content_hash = %#v, want %q",
			document["content_hash"],
			"sha256:example",
		)
	}
}
