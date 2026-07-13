package worldmap

import (
	"errors"
	"fmt"
)

// ErrEmptyRootManifestPath indicates that a chunk document loader was
// requested without a physical root manifest path.
var ErrEmptyRootManifestPath = errors.New(
	"root manifest path cannot be empty",
)

// DuplicateChunkDocumentRequestError indicates that the same authoritative
// chunk was requested more than once in a single atomic load operation.
type DuplicateChunkDocumentRequestError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// Error implements the error interface.
func (e *DuplicateChunkDocumentRequestError) Error() string {
	return fmt.Sprintf(
		"duplicate chunk document request for world space %q at chunk (%d,%d,%d)",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}

// WorldSpaceManifestPathOutsideRootError indicates that a published
// WorldSpace manifest resolves outside the authorized world-map root.
type WorldSpaceManifestPathOutsideRootError struct {
	WorldSpaceID WorldSpaceID
	Path         string
	Root         string
}

// Error implements the error interface.
func (e *WorldSpaceManifestPathOutsideRootError) Error() string {
	return fmt.Sprintf(
		"world space %q manifest path %q resolves outside authorized root %q",
		e.WorldSpaceID,
		e.Path,
		e.Root,
	)
}

// ChunkDocumentPathOutsideRootError indicates that a published chunk path
// resolves outside the authorized world-map root.
type ChunkDocumentPathOutsideRootError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	Path         string
	Root         string
}

// Error implements the error interface.
func (e *ChunkDocumentPathOutsideRootError) Error() string {
	return fmt.Sprintf(
		"chunk document path %q for world space %q at chunk (%d,%d,%d) resolves outside authorized root %q",
		e.Path,
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
		e.Root,
	)
}

// ChunkDocumentTooLargeError indicates that a chunk file exceeds the bounded
// maximum accepted size.
type ChunkDocumentTooLargeError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	MaximumBytes int64
}

// Error implements the error interface.
func (e *ChunkDocumentTooLargeError) Error() string {
	return fmt.Sprintf(
		"chunk document for world space %q at chunk (%d,%d,%d) exceeds maximum size of %d bytes",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
		e.MaximumBytes,
	)
}

// InvalidChunkContentHashError indicates that a published ContentHash does not
// use the canonical sha256:<64 lowercase hexadecimal characters> format.
type InvalidChunkContentHashError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	ContentHash  string
}

// Error implements the error interface.
func (e *InvalidChunkContentHashError) Error() string {
	return fmt.Sprintf(
		"invalid content hash %q for world space %q at chunk (%d,%d,%d)",
		e.ContentHash,
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
	)
}

// ChunkContentHashMismatchError indicates that the raw bytes read from a chunk
// file do not match its published SHA-256 digest.
type ChunkContentHashMismatchError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	Expected     string
	Actual       string
}

// Error implements the error interface.
func (e *ChunkContentHashMismatchError) Error() string {
	return fmt.Sprintf(
		"content hash mismatch for world space %q at chunk (%d,%d,%d): expected %q, got %q",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
		e.Expected,
		e.Actual,
	)
}

// ChunkDocumentIdentityMismatchError indicates that a structurally valid
// document does not identify the published chunk that was requested.
type ChunkDocumentIdentityMismatchError struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
	Field        string
	Expected     string
	Actual       string
}

// Error implements the error interface.
func (e *ChunkDocumentIdentityMismatchError) Error() string {
	return fmt.Sprintf(
		"chunk document identity mismatch for world space %q at chunk (%d,%d,%d): field %s expected %q, got %q",
		e.WorldSpaceID,
		e.Coordinate.ChunkX,
		e.Coordinate.ChunkY,
		e.Coordinate.Z,
		e.Field,
		e.Expected,
		e.Actual,
	)
}
