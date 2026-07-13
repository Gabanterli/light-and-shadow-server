package movement

import (
	"fmt"
	"math"
	"strings"
)

const (
	minSpatialSpawnClaimCoordinate = -1 << 31
	maxSpatialSpawnClaimCoordinate = 1<<31 - 1
)

// SpatialSpawnClaimValidationError reports an invalid claim before the
// SpatialIndex is mutated.
type SpatialSpawnClaimValidationError struct {
	Field  string
	Reason string
}

func (e *SpatialSpawnClaimValidationError) Error() string {
	return fmt.Sprintf(
		"invalid spatial spawn claim field %q: %s",
		e.Field,
		e.Reason,
	)
}

// SpatialSpawnClaimEntityExistsError reports that an entity ID is already
// registered. Replacing an active entity during character entry is forbidden.
type SpatialSpawnClaimEntityExistsError struct {
	EntityID string
}

func (e *SpatialSpawnClaimEntityExistsError) Error() string {
	return fmt.Sprintf(
		"spatial spawn claim entity %q is already registered",
		e.EntityID,
	)
}

// SpatialSpawnClaimOccupiedError reports that another blocking entity already
// owns the requested tile.
type SpatialSpawnClaimOccupiedError struct {
	EntityID string
	X        float64
	Y        float64
	Z        int
}

func (e *SpatialSpawnClaimOccupiedError) Error() string {
	return fmt.Sprintf(
		"spatial spawn claim for entity %q rejected because tile (%v,%v,%d) is occupied",
		e.EntityID,
		e.X,
		e.Y,
		e.Z,
	)
}

// TryClaimBlockingEntity atomically checks dynamic occupancy and registers one
// blocking entity while holding the same SpatialIndex write lock.
//
// This closes the time-of-check/time-of-use window that would otherwise allow
// two concurrent character entries to select and register on the same tile.
//
// SpatialIndex is not yet partitioned by WorldSpaceID. World-space isolation
// remains the responsibility of the caller until the later index-partitioning
// task.
func (si *SpatialIndex) TryClaimBlockingEntity(
	entity *Entity,
) error {
	normalized, err :=
		validateSpatialSpawnClaimEntity(
			entity,
		)
	if err != nil {
		return err
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

	if si.isTileOccupiedLocked(
		normalized.ID,
		normalized.X,
		normalized.Y,
		normalized.Z,
	) {
		return &SpatialSpawnClaimOccupiedError{
			EntityID: normalized.ID,
			X:        normalized.X,
			Y:        normalized.Y,
			Z:        normalized.Z,
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

func validateSpatialSpawnClaimEntity(
	entity *Entity,
) (Entity, error) {
	if entity == nil {
		return Entity{},
			&SpatialSpawnClaimValidationError{
				Field:  "entity",
				Reason: "cannot be nil",
			}
	}

	normalized := *entity
	normalized.ID = strings.TrimSpace(
		normalized.ID,
	)

	if normalized.ID == "" {
		return Entity{},
			&SpatialSpawnClaimValidationError{
				Field:  "entity_id",
				Reason: "cannot be empty",
			}
	}

	if !normalized.BlocksMovement {
		return Entity{},
			&SpatialSpawnClaimValidationError{
				Field:  "blocks_movement",
				Reason: "atomic spawn claims are reserved for blocking entities",
			}
	}

	var err error

	normalized.X, err =
		validateSpatialSpawnClaimCoordinate(
			normalized.X,
			"x",
		)
	if err != nil {
		return Entity{}, err
	}

	normalized.Y, err =
		validateSpatialSpawnClaimCoordinate(
			normalized.Y,
			"y",
		)
	if err != nil {
		return Entity{}, err
	}

	if normalized.Z < 0 ||
		normalized.Z >= 16 {
		return Entity{},
			&SpatialSpawnClaimValidationError{
				Field: "z",
				Reason: fmt.Sprintf(
					"must be within supported floor range [0,15], got %d",
					normalized.Z,
				),
			}
	}

	return normalized, nil
}

func validateSpatialSpawnClaimCoordinate(
	value float64,
	field string,
) (float64, error) {
	if math.IsNaN(value) ||
		math.IsInf(value, 0) {
		return 0,
			&SpatialSpawnClaimValidationError{
				Field:  field,
				Reason: "must be finite",
			}
	}

	if math.Trunc(value) != value {
		return 0,
			&SpatialSpawnClaimValidationError{
				Field:  field,
				Reason: "must identify an integral tile coordinate",
			}
	}

	if value <
		float64(minSpatialSpawnClaimCoordinate) ||
		value >
			float64(maxSpatialSpawnClaimCoordinate) {
		return 0,
			&SpatialSpawnClaimValidationError{
				Field: field,
				Reason: fmt.Sprintf(
					"must be within signed 32-bit range [%d,%d]",
					minSpatialSpawnClaimCoordinate,
					maxSpatialSpawnClaimCoordinate,
				),
			}
	}

	return value, nil
}
