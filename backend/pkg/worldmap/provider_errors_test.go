package worldmap

import (
	"errors"
	"strings"
	"testing"
)

func TestErrNilManifestSnapshotSupportsErrorsIs(t *testing.T) {
	if !errors.Is(ErrNilManifestSnapshot, ErrNilManifestSnapshot) {
		t.Fatal("errors.Is did not recognize ErrNilManifestSnapshot")
	}
}

func TestWorldSpaceNotFoundError(t *testing.T) {
	const worldSpaceID WorldSpaceID = "missing_space"

	err := &WorldSpaceNotFoundError{
		WorldSpaceID: worldSpaceID,
	}

	var target *WorldSpaceNotFoundError
	if !errors.As(err, &target) {
		t.Fatal("errors.As did not recognize WorldSpaceNotFoundError")
	}

	if target.WorldSpaceID != worldSpaceID {
		t.Fatalf(
			"WorldSpaceID = %q, want %q",
			target.WorldSpaceID,
			worldSpaceID,
		)
	}

	if !strings.Contains(err.Error(), string(worldSpaceID)) {
		t.Fatalf(
			"error message %q does not contain world space ID",
			err.Error(),
		)
	}
}

func TestChunkReferenceNotFoundError(t *testing.T) {
	const worldSpaceID WorldSpaceID = "missing_space"

	coordinate := ChunkCoordinate{
		ChunkX: 17,
		ChunkY: 23,
		Z:      5,
	}

	err := &ChunkReferenceNotFoundError{
		WorldSpaceID: worldSpaceID,
		Coordinate:   coordinate,
	}

	var target *ChunkReferenceNotFoundError
	if !errors.As(err, &target) {
		t.Fatal(
			"errors.As did not recognize ChunkReferenceNotFoundError",
		)
	}

	if target.WorldSpaceID != worldSpaceID {
		t.Fatalf(
			"WorldSpaceID = %q, want %q",
			target.WorldSpaceID,
			worldSpaceID,
		)
	}

	if target.Coordinate != coordinate {
		t.Fatalf(
			"Coordinate = %+v, want %+v",
			target.Coordinate,
			coordinate,
		)
	}

	for _, expected := range []string{
		string(worldSpaceID),
		"17",
		"23",
		"5",
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf(
				"error message %q does not contain %q",
				err.Error(),
				expected,
			)
		}
	}
}
