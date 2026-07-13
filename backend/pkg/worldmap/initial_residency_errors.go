package worldmap

import (
	"errors"
	"fmt"
)

var (
	// ErrInitialResidencyPathEmpty indicates that no residency manifest path
	// was supplied to the loader.
	ErrInitialResidencyPathEmpty = errors.New(
		"worldmap: initial residency path cannot be empty",
	)

	// ErrInitialResidencyFileTooLarge indicates that the residency document
	// exceeded the defensive loader size limit.
	ErrInitialResidencyFileTooLarge = errors.New(
		"worldmap: initial residency file exceeds maximum size",
	)

	// ErrInitialResidencyEmpty indicates that the residency document does not
	// declare any chunk to load.
	ErrInitialResidencyEmpty = errors.New(
		"worldmap: initial residency must contain at least one resident chunk",
	)
)

// InitialResidencyFileTooLargeError contains the observed residency file size
// and the configured defensive limit.
type InitialResidencyFileTooLargeError struct {
	Path  string
	Size  int64
	Limit int64
}

// Error implements error.
func (e *InitialResidencyFileTooLargeError) Error() string {
	return fmt.Sprintf(
		"worldmap: initial residency file %q has size %d bytes, limit is %d bytes",
		e.Path,
		e.Size,
		e.Limit,
	)
}

// Unwrap allows callers to match ErrInitialResidencyFileTooLarge.
func (e *InitialResidencyFileTooLargeError) Unwrap() error {
	return ErrInitialResidencyFileTooLarge
}

// DuplicateResidentChunkError indicates that the same world-space chunk
// coordinate was declared more than once.
type DuplicateResidentChunkError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	FirstIndex   int
	SecondIndex  int
}

// Error implements error.
func (e *DuplicateResidentChunkError) Error() string {
	return fmt.Sprintf(
		"worldmap: duplicate resident chunk %q at (%d,%d,%d) in indexes %d and %d",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
		e.FirstIndex,
		e.SecondIndex,
	)
}
