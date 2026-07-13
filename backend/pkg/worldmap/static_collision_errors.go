package worldmap

import (
	"errors"
	"fmt"
)

// ErrNilProductionProvider indicates that a static collision index was
// requested without a production world-map provider.
var ErrNilProductionProvider = errors.New(
	"production provider cannot be nil",
)

// DuplicateChunkDocumentError indicates that more than one materialized chunk
// document was supplied for the same authoritative chunk coordinate.
type DuplicateChunkDocumentError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// Error implements the error interface.
func (e *DuplicateChunkDocumentError) Error() string {
	return fmt.Sprintf(
		"duplicate chunk document for world space %q at chunk (%d,%d,%d)",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}

// UnpublishedChunkDocumentError indicates that a materialized chunk document
// was supplied for a coordinate not published by the production manifest.
type UnpublishedChunkDocumentError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// Error implements the error interface.
func (e *UnpublishedChunkDocumentError) Error() string {
	return fmt.Sprintf(
		"chunk document for world space %q at chunk (%d,%d,%d) is not published",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}

// StaticCollisionChunkNotLoadedError indicates that a chunk is published by
// the production manifest but its materialized document was not supplied to
// the collision index.
type StaticCollisionChunkNotLoadedError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// Error implements the error interface.
func (e *StaticCollisionChunkNotLoadedError) Error() string {
	return fmt.Sprintf(
		"static collision chunk for world space %q at chunk (%d,%d,%d) is not loaded",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}

// WorldPositionOutOfBoundsError indicates that a local world-space position is
// outside the inclusive bounds declared by its manifest.
type WorldPositionOutOfBoundsError struct {
	Position WorldPosition
	Bounds   WorldBounds
}

// Error implements the error interface.
func (e *WorldPositionOutOfBoundsError) Error() string {
	return fmt.Sprintf(
		"position (%d,%d,%d) in world space %q is outside bounds x=[%d,%d], y=[%d,%d], z=[%d,%d]",
		e.Position.X,
		e.Position.Y,
		e.Position.Z,
		e.Position.WorldSpaceID,
		e.Bounds.MinX,
		e.Bounds.MaxX,
		e.Bounds.MinY,
		e.Bounds.MaxY,
		e.Bounds.MinZ,
		e.Bounds.MaxZ,
	)
}

// InvalidStaticMovementStepError indicates that a requested movement is not a
// valid single-tile step in one world space and on one floor.
type InvalidStaticMovementStepError struct {
	From   WorldPosition
	To     WorldPosition
	Reason string
}

// Error implements the error interface.
func (e *InvalidStaticMovementStepError) Error() string {
	return fmt.Sprintf(
		"invalid static movement step from %q:(%d,%d,%d) to %q:(%d,%d,%d): %s",
		e.From.WorldSpaceID,
		e.From.X,
		e.From.Y,
		e.From.Z,
		e.To.WorldSpaceID,
		e.To.X,
		e.To.Y,
		e.To.Z,
		e.Reason,
	)
}

// StaticMovementBlockedError indicates that an existing terrain tile
// explicitly blocks occupation.
type StaticMovementBlockedError struct {
	Position WorldPosition
}

// Error implements the error interface.
func (e *StaticMovementBlockedError) Error() string {
	return fmt.Sprintf(
		"static movement is blocked at world space %q position (%d,%d,%d)",
		e.Position.WorldSpaceID,
		e.Position.X,
		e.Position.Y,
		e.Position.Z,
	)
}

// StaticCollisionTileDataError indicates an internal inconsistency while
// resolving a tile from an otherwise valid, loaded collision chunk.
type StaticCollisionTileDataError struct {
	Position WorldPosition
	Reason   string
}

// Error implements the error interface.
func (e *StaticCollisionTileDataError) Error() string {
	return fmt.Sprintf(
		"invalid static collision tile data at world space %q position (%d,%d,%d): %s",
		e.Position.WorldSpaceID,
		e.Position.X,
		e.Position.Y,
		e.Position.Z,
		e.Reason,
	)
}
