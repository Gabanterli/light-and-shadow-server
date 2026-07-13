package worldmap

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestChunkDocumentLoaderSentinelErrors(t *testing.T) {
	if !errors.Is(
		ErrEmptyRootManifestPath,
		ErrEmptyRootManifestPath,
	) {
		t.Fatal(
			"ErrEmptyRootManifestPath must support errors.Is",
		)
	}

	if !strings.Contains(
		ErrEmptyRootManifestPath.Error(),
		"root manifest path",
	) {
		t.Fatalf(
			"unexpected empty-root error text: %q",
			ErrEmptyRootManifestPath.Error(),
		)
	}
}

func TestDuplicateChunkDocumentRequestErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 4,
		ChunkY: 5,
		Z:      6,
	}
	original := &DuplicateChunkDocumentRequestError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *DuplicateChunkDocumentRequestError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.WorldSpaceID != "test_space" ||
		typed.Coordinate != coordinate {
		t.Fatalf(
			"unexpected duplicate-request context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"duplicate chunk document request",
	)
}

func TestWorldSpaceManifestPathOutsideRootErrorContract(
	t *testing.T,
) {
	original := &WorldSpaceManifestPathOutsideRootError{
		WorldSpaceID: "test_space",
		Path:         "/outside/world_space.json",
		Root:         "/world",
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *WorldSpaceManifestPathOutsideRootError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.WorldSpaceID != original.WorldSpaceID ||
		typed.Path != original.Path ||
		typed.Root != original.Root {
		t.Fatalf(
			"unexpected world-space path context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"outside authorized root",
	)
}

func TestChunkDocumentPathOutsideRootErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 7,
		ChunkY: 8,
		Z:      9,
	}
	original := &ChunkDocumentPathOutsideRootError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
		Path:         "/outside/chunk.json",
		Root:         "/world",
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *ChunkDocumentPathOutsideRootError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.WorldSpaceID != original.WorldSpaceID ||
		typed.Coordinate != coordinate ||
		typed.Path != original.Path ||
		typed.Root != original.Root {
		t.Fatalf(
			"unexpected chunk path context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"outside authorized root",
	)
}

func TestChunkDocumentTooLargeErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 1,
		ChunkY: 2,
		Z:      3,
	}
	original := &ChunkDocumentTooLargeError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
		MaximumBytes: maxChunkDocumentBytes,
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *ChunkDocumentTooLargeError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.MaximumBytes != maxChunkDocumentBytes ||
		typed.Coordinate != coordinate {
		t.Fatalf(
			"unexpected size-error context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"exceeds maximum size",
	)
}

func TestInvalidChunkContentHashErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 10,
		ChunkY: 11,
		Z:      12,
	}
	original := &InvalidChunkContentHashError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
		ContentHash:  "invalid",
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *InvalidChunkContentHashError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.ContentHash != "invalid" ||
		typed.Coordinate != coordinate {
		t.Fatalf(
			"unexpected invalid-hash context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"invalid content hash",
	)
}

func TestChunkContentHashMismatchErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 13,
		ChunkY: 14,
		Z:      15,
	}
	original := &ChunkContentHashMismatchError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
		Expected:     "sha256:expected",
		Actual:       "sha256:actual",
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *ChunkContentHashMismatchError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.Expected != original.Expected ||
		typed.Actual != original.Actual ||
		typed.Coordinate != coordinate {
		t.Fatalf(
			"unexpected hash-mismatch context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"content hash mismatch",
	)
}

func TestChunkDocumentIdentityMismatchErrorContract(
	t *testing.T,
) {
	coordinate := ChunkCoordinate{
		ChunkX: 16,
		ChunkY: 17,
		Z:      18,
	}
	original := &ChunkDocumentIdentityMismatchError{
		WorldSpaceID: "test_space",
		Coordinate:   coordinate,
		Field:        "world_id",
		Expected:     "expected_world",
		Actual:       "actual_world",
	}

	wrapped := fmt.Errorf("wrapped: %w", original)

	var typed *ChunkDocumentIdentityMismatchError
	if !errors.As(wrapped, &typed) {
		t.Fatalf(
			"errors.As failed for %T",
			original,
		)
	}

	if typed.Field != "world_id" ||
		typed.Expected != "expected_world" ||
		typed.Actual != "actual_world" ||
		typed.Coordinate != coordinate {
		t.Fatalf(
			"unexpected identity-error context: %+v",
			typed,
		)
	}

	assertLoaderErrorContains(
		t,
		original,
		"identity mismatch",
	)
}

func assertLoaderErrorContains(
	t *testing.T,
	err error,
	want string,
) {
	t.Helper()

	if !strings.Contains(err.Error(), want) {
		t.Fatalf(
			"error %q does not contain %q",
			err.Error(),
			want,
		)
	}
}
