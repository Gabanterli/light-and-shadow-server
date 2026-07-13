package worldmap

import (
	"errors"
	"fmt"
)

// ErrNilManifestSnapshot indicates that a production provider was requested
// without a loaded manifest snapshot.
var ErrNilManifestSnapshot = errors.New(
	"manifest snapshot cannot be nil",
)

// WorldSpaceNotFoundError indicates that a requested world space is not
// present in the provider.
type WorldSpaceNotFoundError struct {
	WorldSpaceID WorldSpaceID
}

// Error implements the error interface.
func (e *WorldSpaceNotFoundError) Error() string {
	return fmt.Sprintf(
		"world space %q not found",
		e.WorldSpaceID,
	)
}

// ChunkReferenceNotFoundError indicates that a sparse chunk reference is not
// published for a coordinate in a world space.
type ChunkReferenceNotFoundError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// Error implements the error interface.
func (e *ChunkReferenceNotFoundError) Error() string {
	return fmt.Sprintf(
		"chunk reference not found for world space %q at (%d,%d,%d)",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}
