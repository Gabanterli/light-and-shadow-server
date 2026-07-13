package worldmap

const (
	// SupportedSchemaVersion is the world map schema version supported by this binary.
	SupportedSchemaVersion = 1

	// CanonicalChunkSize is the width and height of every complete map chunk.
	CanonicalChunkSize = 32

	// CanonicalWorldID identifies the single logical Light and Shadow universe.
	CanonicalWorldID = "light_and_shadow_world"
)

// WorldSpaceID identifies an authoritative continental world space.
type WorldSpaceID string

const (
	WorldSpaceMainContinent    WorldSpaceID = "main_continent"
	WorldSpaceFireContinent    WorldSpaceID = "fire_continent"
	WorldSpaceIceContinent     WorldSpaceID = "ice_continent"
	WorldSpaceHolyContinent    WorldSpaceID = "holy_continent"
	WorldSpaceShadowContinent  WorldSpaceID = "shadow_continent"
	WorldSpaceNatureContinent  WorldSpaceID = "nature_continent"
	WorldSpaceAbyssiaContinent WorldSpaceID = "abyssia_continent"
)

// CanonicalWorldSpaceIDs returns the seven canonical continental world spaces.
//
// The fixed-size array makes the canonical count explicit and returns a copy,
// preventing callers from modifying shared package state.
func CanonicalWorldSpaceIDs() [7]WorldSpaceID {
	return [7]WorldSpaceID{
		WorldSpaceMainContinent,
		WorldSpaceFireContinent,
		WorldSpaceIceContinent,
		WorldSpaceHolyContinent,
		WorldSpaceShadowContinent,
		WorldSpaceNatureContinent,
		WorldSpaceAbyssiaContinent,
	}
}

// GeographicPosition describes a continent's relative lore position.
//
// It is metadata only and does not create global coordinates between spaces.
type GeographicPosition string

const (
	GeographicPositionCentral      GeographicPosition = "central"
	GeographicPositionNorth        GeographicPosition = "north"
	GeographicPositionSouth        GeographicPosition = "south"
	GeographicPositionEast         GeographicPosition = "east"
	GeographicPositionWest         GeographicPosition = "west"
	GeographicPositionIntermediate GeographicPosition = "intermediate"
	GeographicPositionExtremeNorth GeographicPosition = "extreme_north"
)

// WorldManifest is the root document for the logical game universe.
type WorldManifest struct {
	SchemaVersion int                   `json:"schema_version"`
	WorldID       string                `json:"world_id"`
	WorldVersion  int                   `json:"world_version"`
	WorldSpaces   []WorldSpaceReference `json:"world_spaces"`
}

// WorldSpaceReference links the root manifest to a continental manifest.
type WorldSpaceReference struct {
	WorldSpaceID       WorldSpaceID       `json:"world_space_id"`
	DisplayName        string             `json:"display_name"`
	GeographicPosition GeographicPosition `json:"geographic_position"`
	ManifestFile       string             `json:"manifest_file"`
}

// WorldSpaceManifest describes one continental world space and its published
// sparse chunk set.
type WorldSpaceManifest struct {
	SchemaVersion      int                `json:"schema_version"`
	WorldID            string             `json:"world_id"`
	WorldVersion       int                `json:"world_version"`
	WorldSpaceID       WorldSpaceID       `json:"world_space_id"`
	DisplayName        string             `json:"display_name"`
	GeographicPosition GeographicPosition `json:"geographic_position"`
	WidthTiles         int                `json:"width_tiles"`
	HeightTiles        int                `json:"height_tiles"`
	ChunkSize          int                `json:"chunk_size"`
	MinFloor           int                `json:"min_floor"`
	MaxFloor           int                `json:"max_floor"`
	Chunks             []ChunkReference   `json:"chunks"`
}

// WorldPosition identifies a tile-local position inside one world space.
type WorldPosition struct {
	WorldSpaceID WorldSpaceID `json:"world_space_id"`
	X            int          `json:"x"`
	Y            int          `json:"y"`
	Z            int          `json:"z"`
}

// WorldBounds defines inclusive local coordinate bounds for a world space.
type WorldBounds struct {
	MinX int `json:"min_x"`
	MinY int `json:"min_y"`
	MaxX int `json:"max_x"`
	MaxY int `json:"max_y"`
	MinZ int `json:"min_z"`
	MaxZ int `json:"max_z"`
}

// ChunkCoordinate identifies a chunk inside a world space.
type ChunkCoordinate struct {
	ChunkX int `json:"chunk_x"`
	ChunkY int `json:"chunk_y"`
	Z      int `json:"z"`
}

// ChunkReference links a world-space manifest to a published chunk document.
type ChunkReference struct {
	ChunkX      int    `json:"chunk_x"`
	ChunkY      int    `json:"chunk_y"`
	Z           int    `json:"z"`
	File        string `json:"file"`
	ContentHash string `json:"content_hash,omitempty"`
}

// ChunkDocument contains one sparse map chunk.
type ChunkDocument struct {
	SchemaVersion int              `json:"schema_version"`
	WorldID       string           `json:"world_id"`
	WorldVersion  int              `json:"world_version"`
	WorldSpaceID  WorldSpaceID     `json:"world_space_id"`
	ChunkX        int              `json:"chunk_x"`
	ChunkY        int              `json:"chunk_y"`
	Z             int              `json:"z"`
	Width         int              `json:"width"`
	Height        int              `json:"height"`
	TilePalette   []TileDefinition `json:"tile_palette"`
	Tiles         []uint16         `json:"tiles"`
}

// TileDefinition describes visual and authoritative gameplay properties shared
// by one or more tiles through the chunk palette.
type TileDefinition struct {
	VisualID          int  `json:"visual_id"`
	Walkable          bool `json:"walkable"`
	BlocksMovement    bool `json:"blocks_movement"`
	BlocksProjectiles bool `json:"blocks_projectiles"`
	SafeZone          bool `json:"safe_zone"`
	MovementCost      int  `json:"movement_cost"`
}
