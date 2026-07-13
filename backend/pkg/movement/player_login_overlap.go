package movement

import (
	"fmt"
	"math"
	"strings"
)

// SpatialPlayerLoginClaimBlockedError reports that a non-player blocking
// entity occupies the requested login tile. Existing players are intentionally
// ignored because player-on-player overlap is allowed only during login.
type SpatialPlayerLoginClaimBlockedError struct {
	EntityID    string
	BlockerID   string
	BlockerType string
	X           float64
	Y           float64
	Z           int
}

func (e *SpatialPlayerLoginClaimBlockedError) Error() string {
	return fmt.Sprintf(
		"player login claim for entity %q rejected because blocking %s entity %q occupies tile (%v,%v,%d)",
		e.EntityID,
		e.BlockerType,
		e.BlockerID,
		e.X,
		e.Y,
		e.Z,
	)
}

// IsPlayerLoginTileBlocked reports whether a login tile contains a blocking
// non-player entity. Blocking players are intentionally ignored.
//
// Static terrain remains the responsibility of the authoritative spawn
// placement resolver. This method evaluates only dynamic SpatialIndex state.
func (si *SpatialIndex) IsPlayerLoginTileBlocked(
	excludeEntityID string,
	x, y float64,
	z int,
) bool {
	if si == nil {
		return true
	}

	si.mu.RLock()
	defer si.mu.RUnlock()

	return si.playerLoginBlockingEntityLocked(
		excludeEntityID,
		x,
		y,
		z,
	) != nil
}

// TryClaimPlayerLoginEntity atomically registers a blocking player on a login
// tile while allowing that tile to contain other players.
//
// NPCs, creatures, and every other blocking non-player entity still reject the
// claim. The claimed player remains blocking for all movement after login.
func (si *SpatialIndex) TryClaimPlayerLoginEntity(
	entity *Entity,
) error {
	normalized, err :=
		validateSpatialSpawnClaimEntity(
			entity,
		)
	if err != nil {
		return err
	}

	normalized.Type = strings.ToLower(
		strings.TrimSpace(normalized.Type),
	)

	if normalized.Type != "player" {
		return &SpatialSpawnClaimValidationError{
			Field: "entity_type",
			Reason: fmt.Sprintf(
				"player login claim requires entity type %q, got %q",
				"player",
				normalized.Type,
			),
		}
	}

	if si == nil {
		return &SpatialSpawnClaimValidationError{
			Field:  "spatial_index",
			Reason: "cannot be nil",
		}
	}

	si.mu.Lock()
	defer si.mu.Unlock()

	if si.entities == nil {
		si.entities = make(map[string]*Entity)
	}

	if _, exists := si.entities[normalized.ID]; exists {
		return &SpatialSpawnClaimEntityExistsError{
			EntityID: normalized.ID,
		}
	}

	blocker := si.playerLoginBlockingEntityLocked(
		normalized.ID,
		normalized.X,
		normalized.Y,
		normalized.Z,
	)
	if blocker != nil {
		return &SpatialPlayerLoginClaimBlockedError{
			EntityID:    normalized.ID,
			BlockerID:   blocker.ID,
			BlockerType: blocker.Type,
			X:           normalized.X,
			Y:           normalized.Y,
			Z:           normalized.Z,
		}
	}

	if si.floors[normalized.Z] == nil {
		si.floors[normalized.Z] =
			make(map[uint64]map[string]*Entity)
	}

	key := getChunkKey(
		normalized.X,
		normalized.Y,
	)

	if si.floors[normalized.Z][key] == nil {
		si.floors[normalized.Z][key] =
			make(map[string]*Entity)
	}

	claimed := normalized

	si.entities[claimed.ID] = &claimed
	si.floors[claimed.Z][key][claimed.ID] = &claimed

	return nil
}

// HasOverlappingPlayer reports whether the requested player currently shares a
// tile and floor with another player.
//
// SpatialIndex is not yet partitioned by WorldSpaceID. The current production
// player population is restricted to main_continent until world-space
// partitioning is implemented.
func (si *SpatialIndex) HasOverlappingPlayer(
	playerID string,
) bool {
	if si == nil {
		return false
	}

	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return false
	}

	si.mu.RLock()
	defer si.mu.RUnlock()

	player, exists := si.entities[playerID]
	if !exists ||
		player == nil ||
		!strings.EqualFold(
			strings.TrimSpace(player.Type),
			"player",
		) {
		return false
	}

	if player.Z < 0 ||
		player.Z >= len(si.floors) {
		return false
	}

	key := getChunkKey(
		player.X,
		player.Y,
	)

	chunkEntities := si.floors[player.Z][key]
	if chunkEntities == nil {
		return false
	}

	playerTileX := int(math.Floor(player.X))
	playerTileY := int(math.Floor(player.Y))

	for _, other := range chunkEntities {
		if other == nil ||
			other.ID == playerID ||
			!strings.EqualFold(
				strings.TrimSpace(other.Type),
				"player",
			) {
			continue
		}

		otherTileX := int(math.Floor(other.X))
		otherTileY := int(math.Floor(other.Y))

		if otherTileX == playerTileX &&
			otherTileY == playerTileY {
			return true
		}
	}

	return false
}

func (si *SpatialIndex) playerLoginBlockingEntityLocked(
	excludeEntityID string,
	x, y float64,
	z int,
) *Entity {
	if z < 0 ||
		z >= len(si.floors) {
		return nil
	}

	key := getChunkKey(x, y)
	chunkEntities := si.floors[z][key]
	if chunkEntities == nil {
		return nil
	}

	targetTileX := int(math.Floor(x))
	targetTileY := int(math.Floor(y))

	for _, entity := range chunkEntities {
		if entity == nil ||
			entity.ID == excludeEntityID ||
			!entity.BlocksMovement {
			continue
		}

		entityTileX := int(math.Floor(entity.X))
		entityTileY := int(math.Floor(entity.Y))

		if entityTileX != targetTileX ||
			entityTileY != targetTileY {
			continue
		}

		if strings.EqualFold(
			strings.TrimSpace(entity.Type),
			"player",
		) {
			continue
		}

		return entity
	}

	return nil
}
