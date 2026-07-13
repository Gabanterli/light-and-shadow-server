package worldmap

// ProductionProvider provides immutable, read-only access to validated
// production world-map manifest data.
type ProductionProvider struct {
	worldID       string
	version       int
	worldSpaceIDs []WorldSpaceID
	worldSpaces   map[WorldSpaceID]productionWorldSpaceData
}

type productionWorldSpaceData struct {
	manifest        WorldSpaceManifest
	bounds          WorldBounds
	chunkReferences []ChunkReference
	chunkIndex      map[ChunkCoordinate]ChunkReference
}

var _ Provider = (*ProductionProvider)(nil)

// NewProductionProvider builds a provider-owned immutable representation of a
// previously loaded and validated manifest snapshot.
func NewProductionProvider(
	snapshot *ManifestSnapshot,
) (*ProductionProvider, error) {
	if snapshot == nil {
		return nil, ErrNilManifestSnapshot
	}

	worldSpaceIDs := snapshot.WorldSpaceIDs()
	providerWorldSpaceIDs := make(
		[]WorldSpaceID,
		len(worldSpaceIDs),
	)
	copy(providerWorldSpaceIDs, worldSpaceIDs)

	worldSpaces := make(
		map[WorldSpaceID]productionWorldSpaceData,
		len(providerWorldSpaceIDs),
	)

	for _, worldSpaceID := range providerWorldSpaceIDs {
		manifest, found := snapshot.WorldSpace(worldSpaceID)
		if !found {
			return nil, &WorldSpaceNotFoundError{
				WorldSpaceID: worldSpaceID,
			}
		}

		manifestCopy := copyWorldSpaceManifest(manifest)
		chunkReferences := copyChunkReferences(
			manifestCopy.Chunks,
		)

		chunkIndex := make(
			map[ChunkCoordinate]ChunkReference,
			len(chunkReferences),
		)

		for _, reference := range chunkReferences {
			coordinate := ChunkCoordinate{
				ChunkX: reference.ChunkX,
				ChunkY: reference.ChunkY,
				Z:      reference.Z,
			}

			chunkIndex[coordinate] = reference
		}

		worldSpaces[worldSpaceID] = productionWorldSpaceData{
			manifest: manifestCopy,
			bounds: WorldBounds{
				MinX: 0,
				MinY: 0,
				MaxX: manifestCopy.WidthTiles - 1,
				MaxY: manifestCopy.HeightTiles - 1,
				MinZ: manifestCopy.MinFloor,
				MaxZ: manifestCopy.MaxFloor,
			},
			chunkReferences: chunkReferences,
			chunkIndex:      chunkIndex,
		}
	}

	return &ProductionProvider{
		worldID:       snapshot.WorldID(),
		version:       snapshot.Version(),
		worldSpaceIDs: providerWorldSpaceIDs,
		worldSpaces:   worldSpaces,
	}, nil
}

// Mode returns the production map mode.
func (p *ProductionProvider) Mode() Mode {
	return ModeProduction
}

// WorldID returns the logical universe identifier copied during construction.
func (p *ProductionProvider) WorldID() string {
	return p.worldID
}

// Version returns the world manifest version copied during construction.
func (p *ProductionProvider) Version() int {
	return p.version
}

// WorldSpaceIDs returns a defensive copy in snapshot order.
func (p *ProductionProvider) WorldSpaceIDs() []WorldSpaceID {
	worldSpaceIDs := make(
		[]WorldSpaceID,
		len(p.worldSpaceIDs),
	)
	copy(worldSpaceIDs, p.worldSpaceIDs)

	return worldSpaceIDs
}

// WorldSpace returns a defensive copy of one world-space manifest.
func (p *ProductionProvider) WorldSpace(
	worldSpaceID WorldSpaceID,
) (WorldSpaceManifest, error) {
	worldSpace, found := p.worldSpaces[worldSpaceID]
	if !found {
		return WorldSpaceManifest{},
			&WorldSpaceNotFoundError{
				WorldSpaceID: worldSpaceID,
			}
	}

	return copyWorldSpaceManifest(worldSpace.manifest), nil
}

// Bounds returns precomputed inclusive bounds for a world space.
func (p *ProductionProvider) Bounds(
	worldSpaceID WorldSpaceID,
) (WorldBounds, error) {
	worldSpace, found := p.worldSpaces[worldSpaceID]
	if !found {
		return WorldBounds{},
			&WorldSpaceNotFoundError{
				WorldSpaceID: worldSpaceID,
			}
	}

	return worldSpace.bounds, nil
}

// ChunkReferences returns defensive copies of the sparse published chunk
// references for a world space.
func (p *ProductionProvider) ChunkReferences(
	worldSpaceID WorldSpaceID,
) ([]ChunkReference, error) {
	worldSpace, found := p.worldSpaces[worldSpaceID]
	if !found {
		return nil,
			&WorldSpaceNotFoundError{
				WorldSpaceID: worldSpaceID,
			}
	}

	return copyChunkReferences(
		worldSpace.chunkReferences,
	), nil
}

// ChunkReference returns one published sparse chunk reference using an O(1)
// coordinate lookup.
func (p *ProductionProvider) ChunkReference(
	worldSpaceID WorldSpaceID,
	coordinate ChunkCoordinate,
) (ChunkReference, error) {
	worldSpace, found := p.worldSpaces[worldSpaceID]
	if !found {
		return ChunkReference{},
			&WorldSpaceNotFoundError{
				WorldSpaceID: worldSpaceID,
			}
	}

	reference, found := worldSpace.chunkIndex[coordinate]
	if !found {
		return ChunkReference{},
			&ChunkReferenceNotFoundError{
				WorldSpaceID: worldSpaceID,
				Coordinate:   coordinate,
			}
	}

	return reference, nil
}

func copyWorldSpaceManifest(
	manifest WorldSpaceManifest,
) WorldSpaceManifest {
	manifestCopy := manifest
	manifestCopy.Chunks = copyChunkReferences(
		manifest.Chunks,
	)

	return manifestCopy
}

func copyChunkReferences(
	references []ChunkReference,
) []ChunkReference {
	referencesCopy := make(
		[]ChunkReference,
		len(references),
	)
	copy(referencesCopy, references)

	return referencesCopy
}
