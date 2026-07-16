package main

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

const (
	minAuthoritativeSpawnCoordinate = -1 << 31
	maxAuthoritativeSpawnCoordinate = 1<<31 - 1

	maxAuthoritativeSpawnFallbackSearchRadius = 64
)

type authoritativeSpawnPlacementSource string

const (
	authoritativeSpawnPlacementSourcePersisted authoritativeSpawnPlacementSource = "persisted"
	authoritativeSpawnPlacementSourceFallback  authoritativeSpawnPlacementSource = "fallback"
)

type authoritativeSpawnRelocationReason string

const (
	authoritativeSpawnReasonPersistedAccepted authoritativeSpawnRelocationReason = "persisted_position_accepted"
	authoritativeSpawnReasonDebugPassthrough  authoritativeSpawnRelocationReason = "debug_passthrough"
	authoritativeSpawnReasonInvalidPersisted  authoritativeSpawnRelocationReason = "fallback_after_invalid_persisted_position"
	authoritativeSpawnReasonStaticBlocked     authoritativeSpawnRelocationReason = "fallback_after_static_terrain_block"
	authoritativeSpawnReasonStaticUnavailable authoritativeSpawnRelocationReason = "fallback_after_static_terrain_unavailable"
	authoritativeSpawnReasonDynamicOccupied   authoritativeSpawnRelocationReason = "fallback_after_dynamic_occupancy"
)

type authoritativeSpawnPosition struct {
	WorldSpaceID worldmap.WorldSpaceID
	X            float64
	Y            float64
	Z            int
}

type authoritativeSpawnPlacementRequest struct {
	PlayerID     string
	WorldSpaceID worldmap.WorldSpaceID
	PersistedX   float64
	PersistedY   float64
	PersistedZ   float64
}

type authoritativeSpawnPlacement struct {
	Position         authoritativeSpawnPosition
	Source           authoritativeSpawnPlacementSource
	RelocationReason authoritativeSpawnRelocationReason
	Relocated        bool
}

type authoritativeSpawnStaticOccupancy interface {
	CanOccupy(
		position worldmap.WorldPosition,
	) (bool, error)
}

type authoritativeSpawnDynamicOccupancy interface {
	IsSpawnTileOccupied(
		excludeEntityID string,
		worldSpaceID worldmap.WorldSpaceID,
		x, y float64,
		z int,
	) bool
}

// isNilAuthoritativeSpawnDependency detects interfaces whose dynamic
// value is a nil pointer, map, slice, function, channel, or interface.
// Without this normalization, a typed nil pointer stored in an interface
// compares non-nil and can route debug mode into production-only logic.
func isNilAuthoritativeSpawnDependency(
	dependency interface{},
) bool {
	if dependency == nil {
		return true
	}

	value := reflect.ValueOf(dependency)

	switch value.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// spatialIndexSpawnDynamicOccupancy adapts the current SpatialIndex boundary.
// SpatialIndex is not yet partitioned by WorldSpaceID, so the parameter is
// intentionally retained by this adapter for the future B4-D world-space split.
type spatialIndexSpawnDynamicOccupancy struct {
	index *movement.SpatialIndex
}

func newSpatialIndexSpawnDynamicOccupancy(
	index *movement.SpatialIndex,
) authoritativeSpawnDynamicOccupancy {
	if index == nil {
		return nil
	}

	return &spatialIndexSpawnDynamicOccupancy{
		index: index,
	}
}

func (a *spatialIndexSpawnDynamicOccupancy) IsSpawnTileOccupied(
	excludeEntityID string,
	_ worldmap.WorldSpaceID,
	x, y float64,
	z int,
) bool {
	return a.index.IsTileOccupied(
		excludeEntityID,
		x,
		y,
		z,
	)
}

type authoritativeSpawnPlacementRequestError struct {
	Field  string
	Reason string
}

func (e *authoritativeSpawnPlacementRequestError) Error() string {
	return fmt.Sprintf(
		"invalid authoritative spawn placement request field %q: %s",
		e.Field,
		e.Reason,
	)
}

type authoritativeSpawnPlacementUnavailableError struct {
	PlayerID          string
	Fallback          authoritativeSpawnPosition
	MaxSearchRadius   int
	CheckedCandidates int
	LastStaticError   error
}

func (e *authoritativeSpawnPlacementUnavailableError) Error() string {
	message := fmt.Sprintf(
		"no authoritative spawn placement available for player %q around fallback (%v,%v,%d) in world space %q within radius %d after checking %d candidates",
		e.PlayerID,
		e.Fallback.X,
		e.Fallback.Y,
		e.Fallback.Z,
		e.Fallback.WorldSpaceID,
		e.MaxSearchRadius,
		e.CheckedCandidates,
	)

	if e.LastStaticError != nil {
		message += fmt.Sprintf(
			": last static terrain error: %v",
			e.LastStaticError,
		)
	}

	return message
}

func (e *authoritativeSpawnPlacementUnavailableError) Unwrap() error {
	return e.LastStaticError
}

type authoritativeSpawnStaticEvaluationError struct {
	Position authoritativeSpawnPosition
	Err      error
}

func (e *authoritativeSpawnStaticEvaluationError) Error() string {
	return fmt.Sprintf(
		"authoritative spawn static evaluation failed at world space %q position (%v,%v,%d): %v",
		e.Position.WorldSpaceID,
		e.Position.X,
		e.Position.Y,
		e.Position.Z,
		e.Err,
	)
}

func (e *authoritativeSpawnStaticEvaluationError) Unwrap() error {
	return e.Err
}

func isRecoverableAuthoritativeSpawnStaticError(
	err error,
) bool {
	if err == nil {
		return false
	}

	var outOfBounds *worldmap.WorldPositionOutOfBoundsError
	if errors.As(err, &outOfBounds) {
		return true
	}

	var unpublishedChunk *worldmap.ChunkReferenceNotFoundError
	if errors.As(err, &unpublishedChunk) {
		return true
	}

	var unloadedChunk *worldmap.StaticCollisionChunkNotLoadedError

	return errors.As(err, &unloadedChunk)
}

type authoritativeSpawnCandidateStatus uint8

const (
	authoritativeSpawnCandidateAvailable authoritativeSpawnCandidateStatus = iota
	authoritativeSpawnCandidateStaticBlocked
	authoritativeSpawnCandidateStaticUnavailable
	authoritativeSpawnCandidateDynamicOccupied
)

type authoritativeSpawnFallbackOffset struct {
	X int
	Y int
}

type authoritativeSpawnPlacementResolver struct {
	staticOccupancy  authoritativeSpawnStaticOccupancy
	dynamicOccupancy authoritativeSpawnDynamicOccupancy
	fallback         authoritativeSpawnPosition
	maxSearchRadius  int
	fallbackOffsets  []authoritativeSpawnFallbackOffset
}

func newAuthoritativeSpawnPlacementResolver(
	staticOccupancy authoritativeSpawnStaticOccupancy,
	dynamicOccupancy authoritativeSpawnDynamicOccupancy,
	fallback authoritativeSpawnPosition,
	maxSearchRadius int,
) (*authoritativeSpawnPlacementResolver, error) {
	if isNilAuthoritativeSpawnDependency(staticOccupancy) {
		staticOccupancy = nil
	}

	if isNilAuthoritativeSpawnDependency(dynamicOccupancy) {
		dynamicOccupancy = nil
	}

	if maxSearchRadius < 0 {
		return nil, fmt.Errorf(
			"authoritative spawn fallback search radius cannot be negative: %d",
			maxSearchRadius,
		)
	}

	if maxSearchRadius >
		maxAuthoritativeSpawnFallbackSearchRadius {
		return nil, fmt.Errorf(
			"authoritative spawn fallback search radius %d exceeds maximum %d",
			maxSearchRadius,
			maxAuthoritativeSpawnFallbackSearchRadius,
		)
	}

	if strings.TrimSpace(
		string(fallback.WorldSpaceID),
	) == "" {
		return nil, fmt.Errorf(
			"authoritative spawn fallback world space cannot be empty",
		)
	}

	fallback.WorldSpaceID = worldmap.WorldSpaceID(
		strings.TrimSpace(
			string(fallback.WorldSpaceID),
		),
	)

	normalizedFallback, err :=
		normalizeAuthoritativeSpawnTilePosition(
			fallback,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"normalize authoritative spawn fallback: %w",
			err,
		)
	}

	if staticOccupancy != nil &&
		dynamicOccupancy == nil {
		return nil, fmt.Errorf(
			"production authoritative spawn placement requires dynamic occupancy",
		)
	}

	return &authoritativeSpawnPlacementResolver{
		staticOccupancy:  staticOccupancy,
		dynamicOccupancy: dynamicOccupancy,
		fallback:         normalizedFallback,
		maxSearchRadius:  maxSearchRadius,
		fallbackOffsets: authoritativeSpawnFallbackOffsets(
			maxSearchRadius,
		),
	}, nil
}

func (r *authoritativeSpawnPlacementResolver) Resolve(
	request authoritativeSpawnPlacementRequest,
) (authoritativeSpawnPlacement, error) {
	playerID := strings.TrimSpace(request.PlayerID)
	if playerID == "" {
		return authoritativeSpawnPlacement{},
			&authoritativeSpawnPlacementRequestError{
				Field:  "player_id",
				Reason: "cannot be empty",
			}
	}

	worldSpaceID := worldmap.WorldSpaceID(
		strings.TrimSpace(
			string(request.WorldSpaceID),
		),
	)
	if worldSpaceID == "" {
		return authoritativeSpawnPlacement{},
			&authoritativeSpawnPlacementRequestError{
				Field:  "world_space_id",
				Reason: "cannot be empty",
			}
	}

	request.PlayerID = playerID
	request.WorldSpaceID = worldSpaceID

	if r.staticOccupancy == nil {
		position, err :=
			normalizeDebugAuthoritativeSpawnPosition(
				request,
			)
		if err != nil {
			return authoritativeSpawnPlacement{}, err
		}

		return authoritativeSpawnPlacement{
			Position:         position,
			Source:           authoritativeSpawnPlacementSourcePersisted,
			RelocationReason: authoritativeSpawnReasonDebugPassthrough,
			Relocated:        false,
		}, nil
	}

	persistedPosition, persistedError :=
		normalizeProductionAuthoritativeSpawnPosition(
			request,
		)

	relocationReason :=
		authoritativeSpawnReasonInvalidPersisted

	if persistedError == nil {
		status, staticError := r.evaluateCandidate(
			playerID,
			persistedPosition,
		)

		switch status {
		case authoritativeSpawnCandidateAvailable:
			return authoritativeSpawnPlacement{
				Position:         persistedPosition,
				Source:           authoritativeSpawnPlacementSourcePersisted,
				RelocationReason: authoritativeSpawnReasonPersistedAccepted,
				Relocated:        false,
			}, nil

		case authoritativeSpawnCandidateStaticBlocked:
			relocationReason =
				authoritativeSpawnReasonStaticBlocked

		case authoritativeSpawnCandidateStaticUnavailable:
			if !isRecoverableAuthoritativeSpawnStaticError(
				staticError,
			) {
				return authoritativeSpawnPlacement{},
					&authoritativeSpawnStaticEvaluationError{
						Position: persistedPosition,
						Err:      staticError,
					}
			}

			relocationReason =
				authoritativeSpawnReasonStaticUnavailable

		case authoritativeSpawnCandidateDynamicOccupied:
			relocationReason =
				authoritativeSpawnReasonDynamicOccupied
		}
	}

	fallbackPosition, err := r.resolveFallback(playerID)
	if err != nil {
		return authoritativeSpawnPlacement{}, err
	}

	return authoritativeSpawnPlacement{
		Position:         fallbackPosition,
		Source:           authoritativeSpawnPlacementSourceFallback,
		RelocationReason: relocationReason,
		Relocated:        true,
	}, nil
}

func (r *authoritativeSpawnPlacementResolver) evaluateCandidate(
	playerID string,
	position authoritativeSpawnPosition,
) (authoritativeSpawnCandidateStatus, error) {
	canOccupy, err := r.staticOccupancy.CanOccupy(
		worldmap.WorldPosition{
			WorldSpaceID: position.WorldSpaceID,
			X:            int(position.X),
			Y:            int(position.Y),
			Z:            position.Z,
		},
	)
	if err != nil {
		return authoritativeSpawnCandidateStaticUnavailable,
			err
	}

	if !canOccupy {
		return authoritativeSpawnCandidateStaticBlocked,
			nil
	}

	if r.dynamicOccupancy.IsSpawnTileOccupied(
		playerID,
		position.WorldSpaceID,
		position.X,
		position.Y,
		position.Z,
	) {
		return authoritativeSpawnCandidateDynamicOccupied,
			nil
	}

	return authoritativeSpawnCandidateAvailable, nil
}

func (r *authoritativeSpawnPlacementResolver) resolveFallback(
	playerID string,
) (authoritativeSpawnPosition, error) {
	var lastStaticError error
	checkedCandidates := 0

	for _, offset := range r.fallbackOffsets {
		candidate, valid :=
			offsetAuthoritativeSpawnPosition(
				r.fallback,
				offset,
			)
		if !valid {
			continue
		}

		checkedCandidates++

		status, err := r.evaluateCandidate(
			playerID,
			candidate,
		)
		if err != nil {
			if !isRecoverableAuthoritativeSpawnStaticError(
				err,
			) {
				return authoritativeSpawnPosition{},
					&authoritativeSpawnStaticEvaluationError{
						Position: candidate,
						Err:      err,
					}
			}

			lastStaticError = err
		}

		if status ==
			authoritativeSpawnCandidateAvailable {
			return candidate, nil
		}
	}

	return authoritativeSpawnPosition{},
		&authoritativeSpawnPlacementUnavailableError{
			PlayerID:          playerID,
			Fallback:          r.fallback,
			MaxSearchRadius:   r.maxSearchRadius,
			CheckedCandidates: checkedCandidates,
			LastStaticError:   lastStaticError,
		}
}

func normalizeProductionAuthoritativeSpawnPosition(
	request authoritativeSpawnPlacementRequest,
) (authoritativeSpawnPosition, error) {
	x, err := normalizeAuthoritativeSpawnCoordinate(
		request.PersistedX,
		"persisted_x",
		true,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	y, err := normalizeAuthoritativeSpawnCoordinate(
		request.PersistedY,
		"persisted_y",
		true,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	z, err := normalizeAuthoritativeSpawnCoordinate(
		request.PersistedZ,
		"persisted_z",
		true,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	return authoritativeSpawnPosition{
		WorldSpaceID: request.WorldSpaceID,
		X:            float64(x),
		Y:            float64(y),
		Z:            z,
	}, nil
}

func normalizeDebugAuthoritativeSpawnPosition(
	request authoritativeSpawnPlacementRequest,
) (authoritativeSpawnPosition, error) {
	if math.IsNaN(request.PersistedX) ||
		math.IsInf(request.PersistedX, 0) {
		return authoritativeSpawnPosition{},
			&authoritativeSpawnPlacementRequestError{
				Field:  "persisted_x",
				Reason: "must be finite",
			}
	}

	if math.IsNaN(request.PersistedY) ||
		math.IsInf(request.PersistedY, 0) {
		return authoritativeSpawnPosition{},
			&authoritativeSpawnPlacementRequestError{
				Field:  "persisted_y",
				Reason: "must be finite",
			}
	}

	z, err := normalizeAuthoritativeSpawnCoordinate(
		request.PersistedZ,
		"persisted_z",
		false,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	return authoritativeSpawnPosition{
		WorldSpaceID: request.WorldSpaceID,
		X:            request.PersistedX,
		Y:            request.PersistedY,
		Z:            z,
	}, nil
}

func normalizeAuthoritativeSpawnTilePosition(
	position authoritativeSpawnPosition,
) (authoritativeSpawnPosition, error) {
	x, err := normalizeAuthoritativeSpawnCoordinate(
		position.X,
		"fallback_x",
		true,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	y, err := normalizeAuthoritativeSpawnCoordinate(
		position.Y,
		"fallback_y",
		true,
	)
	if err != nil {
		return authoritativeSpawnPosition{}, err
	}

	if position.Z <
		minAuthoritativeSpawnCoordinate ||
		position.Z >
			maxAuthoritativeSpawnCoordinate {
		return authoritativeSpawnPosition{},
			&authoritativeSpawnPlacementRequestError{
				Field: "fallback_z",
				Reason: fmt.Sprintf(
					"must be within signed 32-bit range [%d,%d]",
					minAuthoritativeSpawnCoordinate,
					maxAuthoritativeSpawnCoordinate,
				),
			}
	}

	position.X = float64(x)
	position.Y = float64(y)

	return position, nil
}

func normalizeAuthoritativeSpawnCoordinate(
	value float64,
	field string,
	requireIntegral bool,
) (int, error) {
	if math.IsNaN(value) ||
		math.IsInf(value, 0) {
		return 0, &authoritativeSpawnPlacementRequestError{
			Field:  field,
			Reason: "must be finite",
		}
	}

	if requireIntegral &&
		math.Trunc(value) != value {
		return 0, &authoritativeSpawnPlacementRequestError{
			Field:  field,
			Reason: "must identify an integral tile coordinate",
		}
	}

	if value <
		float64(minAuthoritativeSpawnCoordinate) ||
		value >
			float64(maxAuthoritativeSpawnCoordinate) {
		return 0, &authoritativeSpawnPlacementRequestError{
			Field: field,
			Reason: fmt.Sprintf(
				"must be within signed 32-bit range [%d,%d]",
				minAuthoritativeSpawnCoordinate,
				maxAuthoritativeSpawnCoordinate,
			),
		}
	}

	return int(value), nil
}

func offsetAuthoritativeSpawnPosition(
	origin authoritativeSpawnPosition,
	offset authoritativeSpawnFallbackOffset,
) (authoritativeSpawnPosition, bool) {
	x := int64(origin.X) + int64(offset.X)
	y := int64(origin.Y) + int64(offset.Y)

	if x <
		int64(minAuthoritativeSpawnCoordinate) ||
		x >
			int64(maxAuthoritativeSpawnCoordinate) ||
		y <
			int64(minAuthoritativeSpawnCoordinate) ||
		y >
			int64(maxAuthoritativeSpawnCoordinate) {
		return authoritativeSpawnPosition{}, false
	}

	return authoritativeSpawnPosition{
		WorldSpaceID: origin.WorldSpaceID,
		X:            float64(x),
		Y:            float64(y),
		Z:            origin.Z,
	}, true
}

func authoritativeSpawnFallbackOffsets(
	maxRadius int,
) []authoritativeSpawnFallbackOffset {
	sideLength := 2*maxRadius + 1
	offsets := make(
		[]authoritativeSpawnFallbackOffset,
		0,
		sideLength*sideLength,
	)

	for y := -maxRadius; y <= maxRadius; y++ {
		for x := -maxRadius; x <= maxRadius; x++ {
			offsets = append(
				offsets,
				authoritativeSpawnFallbackOffset{
					X: x,
					Y: y,
				},
			)
		}
	}

	sort.Slice(
		offsets,
		func(leftIndex, rightIndex int) bool {
			left := offsets[leftIndex]
			right := offsets[rightIndex]

			leftChebyshev := maxAbsoluteInt(
				left.X,
				left.Y,
			)
			rightChebyshev := maxAbsoluteInt(
				right.X,
				right.Y,
			)

			if leftChebyshev != rightChebyshev {
				return leftChebyshev < rightChebyshev
			}

			leftManhattan :=
				absoluteInt(left.X) +
					absoluteInt(left.Y)
			rightManhattan :=
				absoluteInt(right.X) +
					absoluteInt(right.Y)

			if leftManhattan != rightManhattan {
				return leftManhattan < rightManhattan
			}

			if left.Y != right.Y {
				return left.Y < right.Y
			}

			return left.X < right.X
		},
	)

	return offsets
}

func maxAbsoluteInt(
	left int,
	right int,
) int {
	left = absoluteInt(left)
	right = absoluteInt(right)

	if left > right {
		return left
	}

	return right
}

func absoluteInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
