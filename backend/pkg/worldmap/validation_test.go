package worldmap

import (
	"strings"
	"testing"
)

func TestValidateWorldManifest(t *testing.T) {
	validRef := WorldSpaceReference{
		WorldSpaceID:       "main_continent",
		DisplayName:        "Main Continent",
		GeographicPosition: "central",
		ManifestFile:       "spaces/main.json",
	}

	validManifest := WorldManifest{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		WorldSpaces:   []WorldSpaceReference{validRef},
	}

	testCases := []struct {
		name          string
		modifier      func(*WorldManifest)
		expectedError string
	}{
		{"valid manifest", func(m *WorldManifest) {}, ""},
		{"invalid schema version", func(m *WorldManifest) { m.SchemaVersion = 99 }, "unsupported schema_version"},
		{"empty world id", func(m *WorldManifest) { m.WorldID = " " }, "world_id cannot be empty"},
		{"zero world version", func(m *WorldManifest) { m.WorldVersion = 0 }, "world_version must be positive"},
		{"empty world spaces", func(m *WorldManifest) { m.WorldSpaces = nil }, "world_spaces cannot be empty"},
		{"empty world space id", func(m *WorldManifest) { m.WorldSpaces[0].WorldSpaceID = "" }, "world_space_id cannot be empty"},
		{"empty display name", func(m *WorldManifest) { m.WorldSpaces[0].DisplayName = " " }, "display_name cannot be empty"},
		{"empty geo position", func(m *WorldManifest) { m.WorldSpaces[0].GeographicPosition = "" }, "geographic_position cannot be empty"},
		{"empty manifest file", func(m *WorldManifest) { m.WorldSpaces[0].ManifestFile = "" }, "path cannot be empty"},
		{"duplicate world space id", func(m *WorldManifest) { m.WorldSpaces = append(m.WorldSpaces, validRef) }, "duplicate world_space_id"},
		{"duplicate manifest file", func(m *WorldManifest) {
			ref2 := validRef
			ref2.WorldSpaceID = "other_continent"
			m.WorldSpaces = append(m.WorldSpaces, ref2)
		}, "duplicate manifest_file"},
		{"absolute manifest file", func(m *WorldManifest) { m.WorldSpaces[0].ManifestFile = "/abs/path.json" }, "path cannot be absolute"},
		{"traversal in manifest file", func(m *WorldManifest) { m.WorldSpaces[0].ManifestFile = "../secret.json" }, "path traversal is not allowed"},
		{"backslash in manifest file", func(m *WorldManifest) { m.WorldSpaces[0].ManifestFile = "spaces\\main.json" }, "path cannot contain backslashes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := validManifest
			// Create a copy to modify
			manifest.WorldSpaces = append([]WorldSpaceReference(nil), manifest.WorldSpaces...)
			tc.modifier(&manifest)

			err := ValidateWorldManifest(manifest)

			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected an error containing %q, but got nil", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error to contain %q, but got: %v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestValidateWorldSpaceManifest(t *testing.T) {
	validChunkRef := ChunkReference{ChunkX: 10, ChunkY: 20, Z: 7, File: "chunks/10_20_7.json"}

	validManifest := WorldSpaceManifest{
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
		Chunks:             []ChunkReference{validChunkRef},
	}

	testCases := []struct {
		name          string
		modifier      func(*WorldSpaceManifest)
		expectedError string
	}{
		{"valid manifest", func(m *WorldSpaceManifest) {}, ""},
		{"invalid schema version", func(m *WorldSpaceManifest) { m.SchemaVersion = 0 }, "unsupported schema_version"},
		{"empty world id", func(m *WorldSpaceManifest) { m.WorldID = "" }, "world_id cannot be empty"},
		{"invalid world version", func(m *WorldSpaceManifest) { m.WorldVersion = -1 }, "world_version must be positive"},
		{"empty world space id", func(m *WorldSpaceManifest) { m.WorldSpaceID = "" }, "world_space_id cannot be empty"},
		{"invalid width", func(m *WorldSpaceManifest) { m.WidthTiles = 0 }, "width_tiles and height_tiles must be positive"},
		{"invalid chunk size", func(m *WorldSpaceManifest) { m.ChunkSize = 64 }, "chunk_size must be 32"},
		{"invalid floor range", func(m *WorldSpaceManifest) { m.MinFloor = 8; m.MaxFloor = 7 }, "min_floor (8) cannot be greater than max_floor (7)"},
		{"duplicate chunk coord", func(m *WorldSpaceManifest) { m.Chunks = append(m.Chunks, validChunkRef) }, "duplicate chunk coordinate"},
		{"duplicate chunk file", func(m *WorldSpaceManifest) {
			ref2 := validChunkRef
			ref2.ChunkX = 11
			m.Chunks = append(m.Chunks, ref2)
		}, "duplicate file path"},
		{"negative chunk x", func(m *WorldSpaceManifest) { m.Chunks[0].ChunkX = -1 }, "chunk coordinates (-1,20) cannot be negative"},
		{"z below min floor", func(m *WorldSpaceManifest) { m.Chunks[0].Z = -1 }, "chunk z-level -1 is outside floor range"},
		{"z above max floor", func(m *WorldSpaceManifest) { m.Chunks[0].Z = 16 }, "chunk z-level 16 is outside floor range"},
		{"chunk x out of bounds", func(m *WorldSpaceManifest) { m.Chunks[0].ChunkX = 375 }, "chunk coordinate (375,20) is outside world space bounds"},
		{"chunk y out of bounds", func(m *WorldSpaceManifest) { m.Chunks[0].ChunkY = 375 }, "chunk coordinate (10,375) is outside world space bounds"},
		{"valid boundary chunk (12000)", func(m *WorldSpaceManifest) {
			m.Chunks[0].ChunkX = 374
			m.Chunks[0].ChunkY = 374
		}, ""},
		{"valid boundary chunk (16000)", func(m *WorldSpaceManifest) {
			m.WidthTiles = 16000
			m.HeightTiles = 16000
			m.Chunks[0].ChunkX = 499
			m.Chunks[0].ChunkY = 499
		}, ""},
		{"invalid boundary chunk (16000)", func(m *WorldSpaceManifest) {
			m.WidthTiles = 16000
			m.HeightTiles = 16000
			m.Chunks[0].ChunkX = 500
		}, "chunk coordinate (500,20) is outside world space bounds"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := validManifest
			manifest.Chunks = append([]ChunkReference(nil), manifest.Chunks...)
			tc.modifier(&manifest)

			err := ValidateWorldSpaceManifest(manifest)

			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected an error containing %q, but got nil", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error to contain %q, but got: %v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestValidateChunkDocument(t *testing.T) {
	validDoc := ChunkDocument{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		WorldSpaceID:  WorldSpaceMainContinent,
		ChunkX:        10,
		ChunkY:        20,
		Z:             7,
		Width:         CanonicalChunkSize,
		Height:        CanonicalChunkSize,
		TilePalette: []TileDefinition{
			{VisualID: 1, Walkable: true, MovementCost: 1},
			{VisualID: 2, Walkable: false, BlocksMovement: true},
		},
		Tiles: make([]uint16, CanonicalChunkSize*CanonicalChunkSize),
	}

	testCases := []struct {
		name          string
		modifier      func(*ChunkDocument)
		expectedError string
	}{
		{"valid document", func(d *ChunkDocument) {}, ""},
		{"invalid schema version", func(d *ChunkDocument) { d.SchemaVersion = 2 }, "unsupported schema_version"},
		{"empty world id", func(d *ChunkDocument) { d.WorldID = "" }, "world_id cannot be empty"},
		{"invalid world version", func(d *ChunkDocument) { d.WorldVersion = 0 }, "world_version must be positive"},
		{"empty world space id", func(d *ChunkDocument) { d.WorldSpaceID = "" }, "world_space_id cannot be empty"},
		{"negative chunk x", func(d *ChunkDocument) { d.ChunkX = -1 }, "chunk coordinates (-1,20) cannot be negative"},
		{"invalid width", func(d *ChunkDocument) { d.Width = 31 }, "chunk dimensions must be 32x32"},
		{"empty tile palette", func(d *ChunkDocument) { d.TilePalette = nil }, "tile_palette cannot be empty"},
		{"tile count too small", func(d *ChunkDocument) { d.Tiles = make([]uint16, 1023) }, "tiles array length is 1023"},
		{"tile count too large", func(d *ChunkDocument) { d.Tiles = make([]uint16, 1025) }, "tiles array length is 1025"},
		{"invalid palette index", func(d *ChunkDocument) { d.Tiles[5] = 2 }, "palette index 2 is out of bounds"},
		{"negative visual id", func(d *ChunkDocument) { d.TilePalette[0].VisualID = -1 }, "visual_id cannot be negative"},
		{"walkable with zero cost", func(d *ChunkDocument) { d.TilePalette[0].MovementCost = 0 }, "walkable tile must have a positive movement_cost"},
		{"walkable and blocks movement", func(d *ChunkDocument) { d.TilePalette[0].BlocksMovement = true }, "tile cannot be both walkable and block movement"},
		{"non-walkable with zero cost", func(d *ChunkDocument) { d.TilePalette[1].MovementCost = 0 }, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := validDoc
			// Deep copy to avoid test pollution
			doc.TilePalette = append([]TileDefinition(nil), doc.TilePalette...)
			doc.Tiles = append([]uint16(nil), doc.Tiles...)
			tc.modifier(&doc)

			err := ValidateChunkDocument(doc)

			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected an error containing %q, but got nil", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error to contain %q, but got: %v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestValidateAssetPath(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		expectedError string
	}{
		{"valid relative path", "chunks/0_0_0.json", ""},
		{"valid nested path", "world_spaces/main_continent.json", ""},
		{"empty path", "", "path cannot be empty"},
		{"absolute path unix", "/etc/passwd", "path cannot be absolute"},
		{"absolute path windows", "C:/Users/file.json", "path cannot be absolute"},
		{"path with backslashes", "chunks\\0_0_0.json", "path cannot contain backslashes"},
		{"path with traversal", "../secret.txt", "path traversal is not allowed"},
		{"path with traversal nested", "chunks/../secret.txt", "path traversal is not allowed"},
		{"path is dot", ".", "path cannot resolve"},
		{"path is double dot", "..", "path traversal is not allowed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAssetPath(tc.path)

			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("for path %q, expected no error, but got: %v", tc.path, err)
				}
			} else {
				if err == nil {
					t.Errorf("for path %q, expected an error containing %q, but got nil", tc.path, tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("for path %q, expected error to contain %q, but got: %v", tc.path, tc.expectedError, err)
				}
			}
		})
	}
}
func TestValidateWorldSpaceManifestAdditionalCases(t *testing.T) {
	valid := WorldSpaceManifest{
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
				ChunkX: 10,
				ChunkY: 20,
				Z:      0,
				File:   "chunks/10_20_0.json",
			},
		},
	}

	tests := []struct {
		name      string
		modify    func(*WorldSpaceManifest)
		wantError string
	}{
		{
			name: "empty display name",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.DisplayName = " "
			},
			wantError: "display_name cannot be empty",
		},
		{
			name: "empty geographic position",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.GeographicPosition = ""
			},
			wantError: "geographic_position cannot be empty",
		},
		{
			name: "invalid height",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.HeightTiles = 0
			},
			wantError: "width_tiles and height_tiles must be positive",
		},
		{
			name: "negative chunk y",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.Chunks[0].ChunkY = -1
			},
			wantError: "cannot be negative",
		},
		{
			name: "empty chunk file",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.Chunks[0].File = ""
			},
			wantError: "path cannot be empty",
		},
		{
			name: "absolute chunk file",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.Chunks[0].File = "C:/maps/chunk.json"
			},
			wantError: "path cannot be absolute",
		},
		{
			name: "chunk file traversal",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.Chunks[0].File = "chunks/../../secret.json"
			},
			wantError: "path traversal is not allowed",
		},
		{
			name: "chunk file backslash",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.Chunks[0].File = `chunks\10_20_0.json`
			},
			wantError: "path cannot contain backslashes",
		},
		{
			name: "valid Abyssia boundary chunk",
			modify: func(manifest *WorldSpaceManifest) {
				manifest.WorldSpaceID = WorldSpaceAbyssiaContinent
				manifest.DisplayName = "Abyssia"
				manifest.GeographicPosition = GeographicPositionExtremeNorth
				manifest.WidthTiles = 16000
				manifest.HeightTiles = 16000
				manifest.Chunks[0].ChunkX = 499
				manifest.Chunks[0].ChunkY = 499
				manifest.Chunks[0].File = "chunks/499_499_0.json"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := valid
			manifest.Chunks = append(
				[]ChunkReference(nil),
				valid.Chunks...,
			)

			test.modify(&manifest)

			err := ValidateWorldSpaceManifest(manifest)

			if test.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q", test.wantError)
			}

			if !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf(
					"error %q does not contain %q",
					err,
					test.wantError,
				)
			}
		})
	}
}

func TestValidateChunkDocumentAdditionalCases(t *testing.T) {
	valid := ChunkDocument{
		SchemaVersion: SupportedSchemaVersion,
		WorldID:       CanonicalWorldID,
		WorldVersion:  1,
		WorldSpaceID:  WorldSpaceMainContinent,
		ChunkX:        0,
		ChunkY:        0,
		Z:             0,
		Width:         CanonicalChunkSize,
		Height:        CanonicalChunkSize,
		TilePalette: []TileDefinition{
			{
				VisualID:     1,
				Walkable:     true,
				MovementCost: 1,
			},
		},
		Tiles: make(
			[]uint16,
			CanonicalChunkSize*CanonicalChunkSize,
		),
	}

	tests := []struct {
		name      string
		modify    func(*ChunkDocument)
		wantError string
	}{
		{
			name: "negative chunk y",
			modify: func(document *ChunkDocument) {
				document.ChunkY = -1
			},
			wantError: "cannot be negative",
		},
		{
			name: "invalid height",
			modify: func(document *ChunkDocument) {
				document.Height = CanonicalChunkSize - 1
			},
			wantError: "chunk dimensions must be 32x32",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := valid
			document.TilePalette = append(
				[]TileDefinition(nil),
				valid.TilePalette...,
			)
			document.Tiles = append(
				[]uint16(nil),
				valid.Tiles...,
			)

			test.modify(&document)

			err := ValidateChunkDocument(document)

			if err == nil {
				t.Fatalf("expected error containing %q", test.wantError)
			}

			if !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf(
					"error %q does not contain %q",
					err,
					test.wantError,
				)
			}
		})
	}
}
