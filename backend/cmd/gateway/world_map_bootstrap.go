package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

var errProductionWorldMapHasNoPublishedChunks = errors.New(
	"production world map is not ready: no chunk references are published",
)

func initializeWorldMapProvider(
	mode worldmap.Mode,
	manifestPath string,
) (worldmap.Provider, error) {
	switch mode {
	case worldmap.ModeDebug:
		return worldmap.NewDebugProvider(), nil

	case worldmap.ModeProduction:
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
				"load production world map manifest %q: %w",
				normalizedManifestPath,
				err,
			)
		}

		provider, err := worldmap.NewProductionProvider(
			snapshot,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"build production world map provider: %w",
				err,
			)
		}

		if err := validateProductionWorldMapReady(
			provider,
		); err != nil {
			return nil, err
		}

		return provider, nil

	default:
		return nil, fmt.Errorf(
			"unsupported world map mode %q",
			mode,
		)
	}
}

func validateProductionWorldMapReady(
	provider *worldmap.ProductionProvider,
) error {
	if provider == nil {
		return fmt.Errorf(
			"production world map provider cannot be nil",
		)
	}

	for _, worldSpaceID := range provider.WorldSpaceIDs() {
		references, err := provider.ChunkReferences(
			worldSpaceID,
		)
		if err != nil {
			return fmt.Errorf(
				"inspect production world space %q: %w",
				worldSpaceID,
				err,
			)
		}

		if len(references) > 0 {
			return nil
		}
	}

	return errProductionWorldMapHasNoPublishedChunks
}
