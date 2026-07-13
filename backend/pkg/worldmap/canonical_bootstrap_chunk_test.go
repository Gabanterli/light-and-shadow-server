package worldmap

import (
	"path/filepath"
	"testing"
)

func TestCanonicalBootstrapChunk(t *testing.T) {
	t.Parallel()

	worldMapRoot := getTestdataRoot(t)
	rootManifestPath := filepath.Join(
		worldMapRoot,
		"world_manifest.json",
	)

	snapshot, err := LoadManifestSnapshot(rootManifestPath)
	if err != nil {
		t.Fatalf("load canonical manifest snapshot: %v", err)
	}

	provider, err := NewProductionProvider(snapshot)
	if err != nil {
		t.Fatalf("create canonical production provider: %v", err)
	}

	coordinate := ChunkCoordinate{
		ChunkX: 3,
		ChunkY: 3,
		Z:      0,
	}

	reference, err := provider.ChunkReference(
		WorldSpaceMainContinent,
		coordinate,
	)
	if err != nil {
		t.Fatalf("resolve canonical bootstrap chunk reference: %v", err)
	}

	if reference.File != "chunks/main_continent/3_3_0.json" {
		t.Fatalf(
			"bootstrap chunk file = %q, want %q",
			reference.File,
			"chunks/main_continent/3_3_0.json",
		)
	}

	if !validCanonicalChunkContentHash(reference.ContentHash) {
		t.Fatalf(
			"bootstrap chunk content hash is not canonical: %q",
			reference.ContentHash,
		)
	}

	loader, err := NewChunkDocumentLoader(
		rootManifestPath,
		snapshot,
	)
	if err != nil {
		t.Fatalf("create canonical chunk loader: %v", err)
	}

	documents, err := loader.Load(
		[]ChunkDocumentRequest{
			{
				WorldSpaceID: WorldSpaceMainContinent,
				Coordinate:   coordinate,
			},
		},
	)
	if err != nil {
		t.Fatalf("load canonical bootstrap chunk: %v", err)
	}

	if len(documents) != 1 {
		t.Fatalf(
			"loaded document count = %d, want 1",
			len(documents),
		)
	}

	document := documents[0]

	if document.WorldID != CanonicalWorldID {
		t.Fatalf(
			"document world ID = %q, want %q",
			document.WorldID,
			CanonicalWorldID,
		)
	}

	if document.WorldVersion != 1 {
		t.Fatalf(
			"document world version = %d, want 1",
			document.WorldVersion,
		)
	}

	if document.WorldSpaceID != WorldSpaceMainContinent {
		t.Fatalf(
			"document world space = %q, want %q",
			document.WorldSpaceID,
			WorldSpaceMainContinent,
		)
	}

	if document.ChunkX != 3 ||
		document.ChunkY != 3 ||
		document.Z != 0 {
		t.Fatalf(
			"document coordinate = (%d,%d,%d), want (3,3,0)",
			document.ChunkX,
			document.ChunkY,
			document.Z,
		)
	}

	if document.Width != CanonicalChunkSize ||
		document.Height != CanonicalChunkSize {
		t.Fatalf(
			"document dimensions = %dx%d, want %dx%d",
			document.Width,
			document.Height,
			CanonicalChunkSize,
			CanonicalChunkSize,
		)
	}

	if len(document.TilePalette) != 1 {
		t.Fatalf(
			"tile palette length = %d, want 1",
			len(document.TilePalette),
		)
	}

	tileDefinition := document.TilePalette[0]

	if tileDefinition.VisualID != 1 {
		t.Fatalf(
			"tile visual ID = %d, want 1",
			tileDefinition.VisualID,
		)
	}

	if !tileDefinition.Walkable {
		t.Fatal("canonical bootstrap tile is not walkable")
	}

	if tileDefinition.BlocksMovement {
		t.Fatal("canonical bootstrap tile blocks movement")
	}

	if tileDefinition.BlocksProjectiles {
		t.Fatal("canonical bootstrap tile blocks projectiles")
	}

	if tileDefinition.SafeZone {
		t.Fatal("canonical bootstrap tile is unexpectedly a safe zone")
	}

	if tileDefinition.MovementCost != 1 {
		t.Fatalf(
			"tile movement cost = %d, want 1",
			tileDefinition.MovementCost,
		)
	}

	expectedTileCount := CanonicalChunkSize * CanonicalChunkSize

	if len(document.Tiles) != expectedTileCount {
		t.Fatalf(
			"tile count = %d, want %d",
			len(document.Tiles),
			expectedTileCount,
		)
	}

	for index, paletteIndex := range document.Tiles {
		if paletteIndex != 0 {
			t.Fatalf(
				"tiles[%d] palette index = %d, want 0",
				index,
				paletteIndex,
			)
		}
	}

	collisionIndex, err := NewStaticCollisionIndex(
		provider,
		documents,
	)
	if err != nil {
		t.Fatalf(
			"build canonical static collision index: %v",
			err,
		)
	}

	testPositions := []WorldPosition{
		{
			WorldSpaceID: WorldSpaceMainContinent,
			X:            96,
			Y:            96,
			Z:            0,
		},
		{
			WorldSpaceID: WorldSpaceMainContinent,
			X:            112,
			Y:            108,
			Z:            0,
		},
		{
			WorldSpaceID: WorldSpaceMainContinent,
			X:            127,
			Y:            127,
			Z:            0,
		},
	}

	for _, position := range testPositions {
		canOccupy, err := collisionIndex.CanOccupy(position)
		if err != nil {
			t.Fatalf(
				"check canonical position %+v: %v",
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
}
