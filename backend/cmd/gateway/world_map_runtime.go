package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

// worldMapRuntimeState contains the immutable world-map services assembled
// during Gateway startup. Static collision remains nil in debug mode.
type worldMapRuntimeState struct {
	provider             worldmap.Provider
	staticCollisionIndex *worldmap.StaticCollisionIndex
	residentChunkCount   int
}

// worldMapRuntimeIdentityMismatchError indicates that the operational
// residency document targets a different authoritative world or version.
type worldMapRuntimeIdentityMismatchError struct {
	Field    string
	Expected string
	Actual   string
}

// Error implements error.
func (e *worldMapRuntimeIdentityMismatchError) Error() string {
	return fmt.Sprintf(
		"initial residency %s mismatch: expected %q, got %q",
		e.Field,
		e.Expected,
		e.Actual,
	)
}

// initializeWorldMapRuntime atomically assembles the world-map runtime needed
// by the Gateway. Debug mode preserves the existing provider-only behavior.
// Production mode materializes exactly the configured resident chunks and
// builds an immutable static collision index from their authoritative data.
func initializeWorldMapRuntime(
	mode worldmap.Mode,
	manifestPath string,
	initialResidencyPath string,
) (*worldMapRuntimeState, error) {
	if mode != worldmap.ModeProduction {
		provider, err := initializeWorldMapProvider(
			mode,
			manifestPath,
		)
		if err != nil {
			return nil, err
		}

		return &worldMapRuntimeState{
			provider: provider,
		}, nil
	}

	normalizedManifestPath := strings.TrimSpace(
		manifestPath,
	)
	if normalizedManifestPath == "" {
		return nil, fmt.Errorf(
			"production world map manifest path cannot be empty",
		)
	}

	snapshot, err := worldmap.LoadManifestSnapshot(
		normalizedManifestPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load production world map manifest snapshot from %q: %w",
			normalizedManifestPath,
			err,
		)
	}

	productionProvider, err := worldmap.NewProductionProvider(
		snapshot,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create production world map provider: %w",
			err,
		)
	}

	if err := validateProductionWorldMapReady(
		productionProvider,
	); err != nil {
		return nil, fmt.Errorf(
			"validate production world map readiness: %w",
			err,
		)
	}

	residency, err := worldmap.LoadInitialResidencySnapshot(
		initialResidencyPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load production initial residency from %q: %w",
			initialResidencyPath,
			err,
		)
	}

	if residency.WorldID() != productionProvider.WorldID() {
		return nil, &worldMapRuntimeIdentityMismatchError{
			Field:    "world_id",
			Expected: productionProvider.WorldID(),
			Actual:   residency.WorldID(),
		}
	}

	if residency.Version() != productionProvider.Version() {
		return nil, &worldMapRuntimeIdentityMismatchError{
			Field: "world_version",
			Expected: strconv.Itoa(
				productionProvider.Version(),
			),
			Actual: strconv.Itoa(
				residency.Version(),
			),
		}
	}

	residentChunks := residency.ResidentChunks()

	requests := make(
		[]worldmap.ChunkDocumentRequest,
		0,
		len(residentChunks),
	)

	for index, resident := range residentChunks {
		coordinate := resident.Coordinate()

		if _, err := productionProvider.ChunkReference(
			resident.WorldSpaceID,
			coordinate,
		); err != nil {
			return nil, fmt.Errorf(
				"resolve published resident chunk %d for world space %q at chunk (%d,%d,%d): %w",
				index,
				resident.WorldSpaceID,
				coordinate.ChunkX,
				coordinate.ChunkY,
				coordinate.Z,
				err,
			)
		}

		requests = append(
			requests,
			worldmap.ChunkDocumentRequest{
				WorldSpaceID: resident.WorldSpaceID,
				Coordinate:   coordinate,
			},
		)
	}

	loader, err := worldmap.NewChunkDocumentLoader(
		normalizedManifestPath,
		snapshot,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create production chunk document loader: %w",
			err,
		)
	}

	documents, err := loader.Load(requests)
	if err != nil {
		return nil, fmt.Errorf(
			"load resident production chunk documents: %w",
			err,
		)
	}

	collisionIndex, err := worldmap.NewStaticCollisionIndex(
		productionProvider,
		documents,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build resident production static collision index: %w",
			err,
		)
	}

	return &worldMapRuntimeState{
		provider:             productionProvider,
		staticCollisionIndex: collisionIndex,
		residentChunkCount:   len(documents),
	}, nil
}
