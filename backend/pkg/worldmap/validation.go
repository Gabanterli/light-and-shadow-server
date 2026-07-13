package worldmap

import (
	"fmt"
	"path"
	"strings"
)

// ValidateWorldManifest checks the structural integrity of the root world manifest.
func ValidateWorldManifest(manifest WorldManifest) error {
	if manifest.SchemaVersion != SupportedSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d, expected %d", manifest.SchemaVersion, SupportedSchemaVersion)
	}
	if strings.TrimSpace(manifest.WorldID) == "" {
		return fmt.Errorf("world_id cannot be empty")
	}
	if manifest.WorldVersion <= 0 {
		return fmt.Errorf("world_version must be positive, got %d", manifest.WorldVersion)
	}
	if len(manifest.WorldSpaces) == 0 {
		return fmt.Errorf("world_spaces cannot be empty")
	}

	seenIDs := make(map[WorldSpaceID]struct{})
	seenFiles := make(map[string]struct{})

	for i, ref := range manifest.WorldSpaces {
		if strings.TrimSpace(string(ref.WorldSpaceID)) == "" {
			return fmt.Errorf("world_spaces[%d]: world_space_id cannot be empty", i)
		}
		if strings.TrimSpace(ref.DisplayName) == "" {
			return fmt.Errorf("world_spaces[%d]: display_name cannot be empty for id %q", i, ref.WorldSpaceID)
		}
		if strings.TrimSpace(string(ref.GeographicPosition)) == "" {
			return fmt.Errorf("world_spaces[%d]: geographic_position cannot be empty for id %q", i, ref.WorldSpaceID)
		}

		if err := validateAssetPath(ref.ManifestFile); err != nil {
			return fmt.Errorf("world_spaces[%d]: invalid manifest_file for id %q: %w", i, ref.WorldSpaceID, err)
		}

		if _, exists := seenIDs[ref.WorldSpaceID]; exists {
			return fmt.Errorf("world_spaces[%d]: duplicate world_space_id %q", i, ref.WorldSpaceID)
		}
		seenIDs[ref.WorldSpaceID] = struct{}{}

		if _, exists := seenFiles[ref.ManifestFile]; exists {
			return fmt.Errorf("world_spaces[%d]: duplicate manifest_file %q", i, ref.ManifestFile)
		}
		seenFiles[ref.ManifestFile] = struct{}{}
	}

	return nil
}

// ValidateWorldSpaceManifest checks the structural integrity of a continental manifest.
func ValidateWorldSpaceManifest(manifest WorldSpaceManifest) error {
	if manifest.SchemaVersion != SupportedSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d, expected %d", manifest.SchemaVersion, SupportedSchemaVersion)
	}
	if strings.TrimSpace(manifest.WorldID) == "" {
		return fmt.Errorf("world_id cannot be empty")
	}
	if manifest.WorldVersion <= 0 {
		return fmt.Errorf("world_version must be positive, got %d", manifest.WorldVersion)
	}
	if strings.TrimSpace(string(manifest.WorldSpaceID)) == "" {
		return fmt.Errorf("world_space_id cannot be empty")
	}
	if strings.TrimSpace(manifest.DisplayName) == "" {
		return fmt.Errorf("display_name cannot be empty")
	}
	if strings.TrimSpace(string(manifest.GeographicPosition)) == "" {
		return fmt.Errorf("geographic_position cannot be empty")
	}
	if manifest.WidthTiles <= 0 || manifest.HeightTiles <= 0 {
		return fmt.Errorf("width_tiles and height_tiles must be positive, got %dx%d", manifest.WidthTiles, manifest.HeightTiles)
	}
	if manifest.ChunkSize != CanonicalChunkSize {
		return fmt.Errorf("chunk_size must be %d, got %d", CanonicalChunkSize, manifest.ChunkSize)
	}
	if manifest.MinFloor > manifest.MaxFloor {
		return fmt.Errorf("min_floor (%d) cannot be greater than max_floor (%d)", manifest.MinFloor, manifest.MaxFloor)
	}

	chunkColumns := (manifest.WidthTiles + manifest.ChunkSize - 1) / manifest.ChunkSize
	chunkRows := (manifest.HeightTiles + manifest.ChunkSize - 1) / manifest.ChunkSize

	seenCoords := make(map[ChunkCoordinate]struct{})
	seenFiles := make(map[string]struct{})

	for i, chunk := range manifest.Chunks {
		if chunk.ChunkX < 0 || chunk.ChunkY < 0 {
			return fmt.Errorf("chunks[%d]: chunk coordinates (%d,%d) cannot be negative", i, chunk.ChunkX, chunk.ChunkY)
		}
		if chunk.Z < manifest.MinFloor || chunk.Z > manifest.MaxFloor {
			return fmt.Errorf("chunks[%d]: chunk z-level %d is outside floor range [%d, %d]", i, chunk.Z, manifest.MinFloor, manifest.MaxFloor)
		}
		if chunk.ChunkX >= chunkColumns || chunk.ChunkY >= chunkRows {
			return fmt.Errorf("chunks[%d]: chunk coordinate (%d,%d) is outside world space bounds (%dx%d chunks)", i, chunk.ChunkX, chunk.ChunkY, chunkColumns, chunkRows)
		}

		if err := validateAssetPath(chunk.File); err != nil {
			return fmt.Errorf("chunks[%d]: invalid file path for chunk (%d,%d,%d): %w", i, chunk.ChunkX, chunk.ChunkY, chunk.Z, err)
		}

		coord := ChunkCoordinate{ChunkX: chunk.ChunkX, ChunkY: chunk.ChunkY, Z: chunk.Z}
		if _, exists := seenCoords[coord]; exists {
			return fmt.Errorf("chunks[%d]: duplicate chunk coordinate (%d,%d,%d)", i, chunk.ChunkX, chunk.ChunkY, chunk.Z)
		}
		seenCoords[coord] = struct{}{}

		if _, exists := seenFiles[chunk.File]; exists {
			return fmt.Errorf("chunks[%d]: duplicate file path %q", i, chunk.File)
		}
		seenFiles[chunk.File] = struct{}{}
	}

	return nil
}

// ValidateChunkDocument checks the structural integrity of a single chunk file.
func ValidateChunkDocument(doc ChunkDocument) error {
	if doc.SchemaVersion != SupportedSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d, expected %d", doc.SchemaVersion, SupportedSchemaVersion)
	}
	if strings.TrimSpace(doc.WorldID) == "" {
		return fmt.Errorf("world_id cannot be empty")
	}
	if doc.WorldVersion <= 0 {
		return fmt.Errorf("world_version must be positive, got %d", doc.WorldVersion)
	}
	if strings.TrimSpace(string(doc.WorldSpaceID)) == "" {
		return fmt.Errorf("world_space_id cannot be empty")
	}
	if doc.ChunkX < 0 || doc.ChunkY < 0 {
		return fmt.Errorf("chunk coordinates (%d,%d) cannot be negative", doc.ChunkX, doc.ChunkY)
	}
	if doc.Width != CanonicalChunkSize || doc.Height != CanonicalChunkSize {
		return fmt.Errorf("chunk dimensions must be %dx%d, got %dx%d", CanonicalChunkSize, CanonicalChunkSize, doc.Width, doc.Height)
	}
	if len(doc.TilePalette) == 0 {
		return fmt.Errorf("tile_palette cannot be empty")
	}

	expectedTileCount := doc.Width * doc.Height
	if len(doc.Tiles) != expectedTileCount {
		return fmt.Errorf("tiles array length is %d, but expected %d for a %dx%d chunk", len(doc.Tiles), expectedTileCount, doc.Width, doc.Height)
	}

	paletteLen := len(doc.TilePalette)
	for i, tileIndex := range doc.Tiles {
		if int(tileIndex) >= paletteLen {
			return fmt.Errorf("tiles[%d]: palette index %d is out of bounds for palette of length %d", i, tileIndex, paletteLen)
		}
	}

	for i, tileDef := range doc.TilePalette {
		if tileDef.VisualID < 0 {
			return fmt.Errorf("tile_palette[%d]: visual_id cannot be negative, got %d", i, tileDef.VisualID)
		}
		if tileDef.Walkable && tileDef.MovementCost <= 0 {
			return fmt.Errorf("tile_palette[%d]: walkable tile must have a positive movement_cost, got %d", i, tileDef.MovementCost)
		}
		if tileDef.Walkable && tileDef.BlocksMovement {
			return fmt.Errorf("tile_palette[%d]: tile cannot be both walkable and block movement", i)
		}
	}

	return nil
}

// validateAssetPath ensures a file path is a safe, relative, portable asset path.
func validateAssetPath(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if strings.Contains(filePath, `\`) {
		return fmt.Errorf("path cannot contain backslashes: %q", filePath)
	}

	if path.IsAbs(filePath) || hasWindowsDrivePrefix(filePath) {
		return fmt.Errorf("path cannot be absolute: %q", filePath)
	}

	for _, segment := range strings.Split(filePath, "/") {
		if segment == ".." {
			return fmt.Errorf("path traversal is not allowed: %q", filePath)
		}
	}

	cleaned := path.Clean(filePath)

	if cleaned == "." || cleaned == "" {
		return fmt.Errorf("path cannot resolve to %q", cleaned)
	}

	if cleaned != filePath {
		return fmt.Errorf(
			"path is not clean: original %q, cleaned %q",
			filePath,
			cleaned,
		)
	}

	return nil
}

func hasWindowsDrivePrefix(filePath string) bool {
	if len(filePath) < 2 || filePath[1] != ':' {
		return false
	}

	first := filePath[0]

	return (first >= 'A' && first <= 'Z') ||
		(first >= 'a' && first <= 'z')
}
