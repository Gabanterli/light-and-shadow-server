package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

// worldMapStaticStepValidator adapts the neutral movement terrain contract to
// the immutable production world-map collision index.
type worldMapStaticStepValidator struct {
	index *worldmap.StaticCollisionIndex
}

// worldMapMovementCoordinateError identifies a movement coordinate that cannot
// be represented safely as one authoritative tile coordinate.
type worldMapMovementCoordinateError struct {
	Field  string
	Value  float64
	Reason string
}

// Error implements error.
func (e *worldMapMovementCoordinateError) Error() string {
	return fmt.Sprintf(
		"invalid movement coordinate %s=%v: %s",
		e.Field,
		e.Value,
		e.Reason,
	)
}

// worldMapMovementWorldSpaceError identifies malformed movement world-space
// identity before querying the production provider.
type worldMapMovementWorldSpaceError struct {
	WorldSpaceID string
	Reason       string
}

// Error implements error.
func (e *worldMapMovementWorldSpaceError) Error() string {
	return fmt.Sprintf(
		"invalid movement world space %q: %s",
		e.WorldSpaceID,
		e.Reason,
	)
}

// newWorldMapStaticStepValidator returns nil when production static collision
// is unavailable, preserving the MovementSystem legacy/debug terrain path.
func newWorldMapStaticStepValidator(
	index *worldmap.StaticCollisionIndex,
) movement.StaticStepValidator {
	if index == nil {
		return nil
	}

	return &worldMapStaticStepValidator{
		index: index,
	}
}

// ValidateStaticStep converts the neutral movement step without truncation and
// delegates all terrain rules to the immutable world-map collision index.
func (v *worldMapStaticStepValidator) ValidateStaticStep(
	step movement.StaticStep,
) error {
	if v == nil || v.index == nil {
		return fmt.Errorf(
			"production static collision index is unavailable",
		)
	}

	trimmedWorldSpaceID := strings.TrimSpace(
		step.WorldSpaceID,
	)

	if trimmedWorldSpaceID == "" {
		return &worldMapMovementWorldSpaceError{
			WorldSpaceID: step.WorldSpaceID,
			Reason:       "identity cannot be empty",
		}
	}

	if trimmedWorldSpaceID != step.WorldSpaceID {
		return &worldMapMovementWorldSpaceError{
			WorldSpaceID: step.WorldSpaceID,
			Reason: "identity cannot contain leading or " +
				"trailing whitespace",
		}
	}

	fromX, err := exactWorldMapMovementCoordinate(
		"from_x",
		step.FromX,
	)
	if err != nil {
		return err
	}

	fromY, err := exactWorldMapMovementCoordinate(
		"from_y",
		step.FromY,
	)
	if err != nil {
		return err
	}

	toX, err := exactWorldMapMovementCoordinate(
		"to_x",
		step.ToX,
	)
	if err != nil {
		return err
	}

	toY, err := exactWorldMapMovementCoordinate(
		"to_y",
		step.ToY,
	)
	if err != nil {
		return err
	}

	return v.index.ValidateStep(
		worldmap.WorldPosition{
			WorldSpaceID: worldmap.WorldSpaceID(
				step.WorldSpaceID,
			),
			X: fromX,
			Y: fromY,
			Z: step.FromZ,
		},
		worldmap.WorldPosition{
			WorldSpaceID: worldmap.WorldSpaceID(
				step.WorldSpaceID,
			),
			X: toX,
			Y: toY,
			Z: step.ToZ,
		},
	)
}

func exactWorldMapMovementCoordinate(
	field string,
	value float64,
) (int, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, &worldMapMovementCoordinateError{
			Field:  field,
			Value:  value,
			Reason: "coordinate must be finite",
		}
	}

	if math.Trunc(value) != value {
		return 0, &worldMapMovementCoordinateError{
			Field:  field,
			Value:  value,
			Reason: "coordinate must identify an exact tile",
		}
	}

	const (
		minMovementCoordinate = -1 << 31
		maxMovementCoordinate = 1<<31 - 1
	)

	if value < float64(minMovementCoordinate) ||
		value > float64(maxMovementCoordinate) {
		return 0, &worldMapMovementCoordinateError{
			Field: field,
			Value: value,
			Reason: "coordinate exceeds signed 32-bit " +
				"movement range",
		}
	}

	return int(value), nil
}
