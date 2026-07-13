package worldmap

import (
	"errors"
	"fmt"
)

// StaticCollisionIndex provides immutable, sparse, read-only terrain
// collision lookups for materialized production chunks.
type StaticCollisionIndex struct {
	provider         *ProductionProvider
	loadedCollisions map[staticCollisionChunkKey]staticCollisionChunk
}

type staticCollisionChunkKey struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

type staticCollisionChunk struct {
	width   int
	height  int
	blocked []bool
}

// StaticTileCollision contains the authoritative static movement property for
// one existing tile.
type StaticTileCollision struct {
	BlocksMovement bool
}

// NewStaticCollisionIndex builds an immutable collision index containing only
// the supplied materialized chunks. Published chunk metadata remains owned by
// the immutable production provider.
func NewStaticCollisionIndex(
	provider *ProductionProvider,
	documents []ChunkDocument,
) (*StaticCollisionIndex, error) {
	if provider == nil {
		return nil, ErrNilProductionProvider
	}

	index := &StaticCollisionIndex{
		provider: provider,
		loadedCollisions: make(
			map[staticCollisionChunkKey]staticCollisionChunk,
			len(documents),
		),
	}

	for documentIndex := range documents {
		document := documents[documentIndex]

		if err := ValidateChunkDocument(document); err != nil {
			return nil, fmt.Errorf(
				"validate chunk document %d for world space %q at chunk (%d,%d,%d): %w",
				documentIndex,
				document.WorldSpaceID,
				document.ChunkX,
				document.ChunkY,
				document.Z,
				err,
			)
		}

		coordinate := ChunkCoordinate{
			ChunkX: document.ChunkX,
			ChunkY: document.ChunkY,
			Z:      document.Z,
		}

		key := staticCollisionChunkKey{
			WorldSpaceID: document.WorldSpaceID,
			Coordinate:   coordinate,
		}

		if _, exists := index.loadedCollisions[key]; exists {
			return nil, &DuplicateChunkDocumentError{
				WorldSpaceID: document.WorldSpaceID,
				Coordinate:   coordinate,
			}
		}

		if _, err := provider.WorldSpace(
			document.WorldSpaceID,
		); err != nil {
			return nil, err
		}

		if document.WorldID != provider.WorldID() {
			return nil, fmt.Errorf(
				"chunk document for world space %q at chunk (%d,%d,%d) has world_id %q, expected %q",
				document.WorldSpaceID,
				document.ChunkX,
				document.ChunkY,
				document.Z,
				document.WorldID,
				provider.WorldID(),
			)
		}

		if document.WorldVersion != provider.Version() {
			return nil, fmt.Errorf(
				"chunk document for world space %q at chunk (%d,%d,%d) has world_version %d, expected %d",
				document.WorldSpaceID,
				document.ChunkX,
				document.ChunkY,
				document.Z,
				document.WorldVersion,
				provider.Version(),
			)
		}

		if _, err := provider.ChunkReference(
			document.WorldSpaceID,
			coordinate,
		); err != nil {
			var notFound *ChunkReferenceNotFoundError

			if errors.As(err, &notFound) {
				return nil, &UnpublishedChunkDocumentError{
					WorldSpaceID: document.WorldSpaceID,
					Coordinate:   coordinate,
				}
			}

			return nil, fmt.Errorf(
				"resolve published chunk for world space %q at chunk (%d,%d,%d): %w",
				document.WorldSpaceID,
				document.ChunkX,
				document.ChunkY,
				document.Z,
				err,
			)
		}

		blocked := make([]bool, len(document.Tiles))

		for tileOffset, paletteIndex := range document.Tiles {
			blocked[tileOffset] =
				document.TilePalette[paletteIndex].
					BlocksMovement
		}

		index.loadedCollisions[key] = staticCollisionChunk{
			width:   document.Width,
			height:  document.Height,
			blocked: blocked,
		}
	}

	return index, nil
}

// TileCollision resolves one local world-space position to its immutable
// static movement property.
func (i *StaticCollisionIndex) TileCollision(
	position WorldPosition,
) (StaticTileCollision, error) {
	if _, err := i.validatePositionBounds(position); err != nil {
		return StaticTileCollision{}, err
	}

	chunkCoordinate := ChunkCoordinate{
		ChunkX: position.X / CanonicalChunkSize,
		ChunkY: position.Y / CanonicalChunkSize,
		Z:      position.Z,
	}

	if _, err := i.provider.ChunkReference(
		position.WorldSpaceID,
		chunkCoordinate,
	); err != nil {
		return StaticTileCollision{}, err
	}

	key := staticCollisionChunkKey{
		WorldSpaceID: position.WorldSpaceID,
		Coordinate:   chunkCoordinate,
	}

	chunk, loaded := i.loadedCollisions[key]
	if !loaded {
		return StaticTileCollision{},
			&StaticCollisionChunkNotLoadedError{
				WorldSpaceID: position.WorldSpaceID,
				Coordinate:   chunkCoordinate,
			}
	}

	localX := position.X % CanonicalChunkSize
	localY := position.Y % CanonicalChunkSize

	if localX < 0 ||
		localY < 0 ||
		localX >= chunk.width ||
		localY >= chunk.height {
		return StaticTileCollision{},
			&StaticCollisionTileDataError{
				Position: position,
				Reason: fmt.Sprintf(
					"local coordinate (%d,%d) is outside loaded chunk dimensions %dx%d",
					localX,
					localY,
					chunk.width,
					chunk.height,
				),
			}
	}

	tileOffset := localY*chunk.width + localX

	if tileOffset < 0 ||
		tileOffset >= len(chunk.blocked) {
		return StaticTileCollision{},
			&StaticCollisionTileDataError{
				Position: position,
				Reason: fmt.Sprintf(
					"tile offset %d is outside collision data length %d",
					tileOffset,
					len(chunk.blocked),
				),
			}
	}

	return StaticTileCollision{
		BlocksMovement: chunk.blocked[tileOffset],
	}, nil
}

// CanOccupy returns true only when authoritative static data exists and the
// tile does not block movement.
func (i *StaticCollisionIndex) CanOccupy(
	position WorldPosition,
) (bool, error) {
	collision, err := i.TileCollision(position)
	if err != nil {
		return false, err
	}

	return !collision.BlocksMovement, nil
}

// ValidateStep validates one cardinal or diagonal single-tile step against
// static terrain only.
func (i *StaticCollisionIndex) ValidateStep(
	from WorldPosition,
	to WorldPosition,
) error {
	if from.WorldSpaceID != to.WorldSpaceID {
		return &InvalidStaticMovementStepError{
			From:   from,
			To:     to,
			Reason: "world space changes are not allowed",
		}
	}

	if from.Z != to.Z {
		return &InvalidStaticMovementStepError{
			From:   from,
			To:     to,
			Reason: "floor changes are not allowed",
		}
	}

	deltaX := to.X - from.X
	deltaY := to.Y - from.Y

	absoluteX := absoluteStaticStepDelta(deltaX)
	absoluteY := absoluteStaticStepDelta(deltaY)

	if absoluteX == 0 && absoluteY == 0 {
		return &InvalidStaticMovementStepError{
			From:   from,
			To:     to,
			Reason: "origin and destination are equal",
		}
	}

	if absoluteX > 1 || absoluteY > 1 {
		return &InvalidStaticMovementStepError{
			From:   from,
			To:     to,
			Reason: "movement exceeds one tile",
		}
	}

	if _, err := i.validatePositionBounds(from); err != nil {
		return err
	}

	if err := i.validateOccupableStepPosition(to); err != nil {
		return err
	}

	if absoluteX == 1 && absoluteY == 1 {
		firstOrthogonal := WorldPosition{
			WorldSpaceID: from.WorldSpaceID,
			X:            from.X + deltaX,
			Y:            from.Y,
			Z:            from.Z,
		}

		if err := i.validateOccupableStepPosition(
			firstOrthogonal,
		); err != nil {
			return err
		}

		secondOrthogonal := WorldPosition{
			WorldSpaceID: from.WorldSpaceID,
			X:            from.X,
			Y:            from.Y + deltaY,
			Z:            from.Z,
		}

		if err := i.validateOccupableStepPosition(
			secondOrthogonal,
		); err != nil {
			return err
		}
	}

	return nil
}

func (i *StaticCollisionIndex) validatePositionBounds(
	position WorldPosition,
) (WorldBounds, error) {
	bounds, err := i.provider.Bounds(
		position.WorldSpaceID,
	)
	if err != nil {
		return WorldBounds{}, err
	}

	if position.X < bounds.MinX ||
		position.X > bounds.MaxX ||
		position.Y < bounds.MinY ||
		position.Y > bounds.MaxY ||
		position.Z < bounds.MinZ ||
		position.Z > bounds.MaxZ {
		return WorldBounds{},
			&WorldPositionOutOfBoundsError{
				Position: position,
				Bounds:   bounds,
			}
	}

	return bounds, nil
}

func (i *StaticCollisionIndex) validateOccupableStepPosition(
	position WorldPosition,
) error {
	canOccupy, err := i.CanOccupy(position)
	if err != nil {
		return err
	}

	if !canOccupy {
		return &StaticMovementBlockedError{
			Position: position,
		}
	}

	return nil
}

func absoluteStaticStepDelta(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
