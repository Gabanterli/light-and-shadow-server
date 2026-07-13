package worldmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadManifestSnapshot loads and validates the root manifest and all referenced
// world-space manifests, then builds an in-memory immutable snapshot.
func LoadManifestSnapshot(rootManifestPath string) (*ManifestSnapshot, error) {
	if strings.TrimSpace(rootManifestPath) == "" {
		return nil, errors.New("root manifest path cannot be empty")
	}

	absoluteRootPath, err := filepath.Abs(rootManifestPath)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve absolute root world manifest path %q: %w",
			rootManifestPath,
			err,
		)
	}
	absoluteRootPath = filepath.Clean(absoluteRootPath)
	rootDir := filepath.Dir(absoluteRootPath)

	root, err := loadAndValidateRootManifest(absoluteRootPath)
	if err != nil {
		return nil, err
	}

	if err := checkCanonicalWorldSpaces(root); err != nil {
		return nil, err
	}

	worldSpaces, err := loadAndValidateWorldSpaces(root, rootDir)
	if err != nil {
		return nil, err
	}

	if err := checkForExtraManifests(
		root,
		rootDir,
		absoluteRootPath,
	); err != nil {
		return nil, err
	}

	canonicalIDs := CanonicalWorldSpaceIDs()
	worldSpaceIDs := make([]WorldSpaceID, len(canonicalIDs))
	for index, worldSpaceID := range canonicalIDs {
		worldSpaceIDs[index] = worldSpaceID
	}

	return &ManifestSnapshot{
		rootManifest:  *root,
		worldSpaces:   worldSpaces,
		worldSpaceIDs: worldSpaceIDs,
	}, nil
}

func loadAndValidateRootManifest(
	rootManifestPath string,
) (*WorldManifest, error) {
	data, err := os.ReadFile(rootManifestPath)
	if err != nil {
		return nil, fmt.Errorf(
			"load root world manifest %q: %w",
			rootManifestPath,
			err,
		)
	}

	var manifest WorldManifest
	if err := decodeStrictJSON(data, &manifest); err != nil {
		return nil, fmt.Errorf(
			"decode root world manifest %q: %w",
			rootManifestPath,
			err,
		)
	}

	if err := ValidateWorldManifest(manifest); err != nil {
		return nil, fmt.Errorf(
			"validate root world manifest %q: %w",
			rootManifestPath,
			err,
		)
	}

	return &manifest, nil
}

func checkCanonicalWorldSpaces(root *WorldManifest) error {
	canonicalIDs := CanonicalWorldSpaceIDs()

	if len(root.WorldSpaces) != len(canonicalIDs) {
		return fmt.Errorf(
			"root manifest must contain exactly %d world spaces, found %d",
			len(canonicalIDs),
			len(root.WorldSpaces),
		)
	}

	canonicalSet := make(
		map[WorldSpaceID]struct{},
		len(canonicalIDs),
	)
	for _, worldSpaceID := range canonicalIDs {
		canonicalSet[worldSpaceID] = struct{}{}
	}

	foundIDs := make(
		map[WorldSpaceID]struct{},
		len(root.WorldSpaces),
	)

	for _, reference := range root.WorldSpaces {
		if _, canonical := canonicalSet[reference.WorldSpaceID]; !canonical {
			return fmt.Errorf(
				"unexpected non-canonical world space %q in root manifest",
				reference.WorldSpaceID,
			)
		}

		foundIDs[reference.WorldSpaceID] = struct{}{}
	}

	for _, worldSpaceID := range canonicalIDs {
		if _, found := foundIDs[worldSpaceID]; !found {
			return fmt.Errorf(
				"missing canonical world space %q in root manifest",
				worldSpaceID,
			)
		}
	}

	return nil
}

func loadAndValidateWorldSpaces(
	root *WorldManifest,
	rootDir string,
) (map[WorldSpaceID]WorldSpaceManifest, error) {
	worldSpaces := make(
		map[WorldSpaceID]WorldSpaceManifest,
		len(root.WorldSpaces),
	)

	for _, reference := range root.WorldSpaces {
		childPath, err := secureResolvePath(
			rootDir,
			reference.ManifestFile,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"world space %q manifest path: %w",
				reference.WorldSpaceID,
				err,
			)
		}

		data, err := os.ReadFile(childPath)
		if err != nil {
			return nil, fmt.Errorf(
				"load world space %q from %q: %w",
				reference.WorldSpaceID,
				childPath,
				err,
			)
		}

		var worldSpace WorldSpaceManifest
		if err := decodeStrictJSON(data, &worldSpace); err != nil {
			return nil, fmt.Errorf(
				"decode world space %q from %q: %w",
				reference.WorldSpaceID,
				childPath,
				err,
			)
		}

		if err := ValidateWorldSpaceManifest(worldSpace); err != nil {
			return nil, fmt.Errorf(
				"validate world space %q from %q: %w",
				reference.WorldSpaceID,
				childPath,
				err,
			)
		}

		if err := checkConsistency(
			root,
			&reference,
			&worldSpace,
		); err != nil {
			return nil, err
		}

		worldSpaces[worldSpace.WorldSpaceID] = worldSpace
	}

	return worldSpaces, nil
}

func decodeStrictJSON(data []byte, destination interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var trailing interface{}
	err := decoder.Decode(&trailing)

	if errors.Is(err, io.EOF) {
		return nil
	}

	if err == nil {
		return errors.New(
			"multiple JSON documents are not allowed",
		)
	}

	return fmt.Errorf(
		"invalid trailing data after JSON document: %w",
		err,
	)
}

func secureResolvePath(
	baseDir string,
	assetPath string,
) (string, error) {
	if err := validateAssetPath(assetPath); err != nil {
		return "", fmt.Errorf(
			"invalid asset path %q: %w",
			assetPath,
			err,
		)
	}

	absoluteBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf(
			"resolve absolute base directory %q: %w",
			baseDir,
			err,
		)
	}
	absoluteBase = filepath.Clean(absoluteBase)

	candidatePath := filepath.Join(
		absoluteBase,
		filepath.FromSlash(assetPath),
	)

	absoluteCandidate, err := filepath.Abs(candidatePath)
	if err != nil {
		return "", fmt.Errorf(
			"resolve absolute manifest path %q: %w",
			candidatePath,
			err,
		)
	}
	absoluteCandidate = filepath.Clean(absoluteCandidate)

	relativePath, err := filepath.Rel(
		absoluteBase,
		absoluteCandidate,
	)
	if err != nil {
		return "", fmt.Errorf(
			"compare manifest path %q with base directory %q: %w",
			absoluteCandidate,
			absoluteBase,
			err,
		)
	}

	parentPrefix := ".." + string(os.PathSeparator)
	if relativePath == ".." ||
		filepath.IsAbs(relativePath) ||
		strings.HasPrefix(relativePath, parentPrefix) {
		return "", fmt.Errorf(
			"manifest path %q escapes base directory %q",
			assetPath,
			baseDir,
		)
	}

	return absoluteCandidate, nil
}

func checkConsistency(
	root *WorldManifest,
	reference *WorldSpaceReference,
	worldSpace *WorldSpaceManifest,
) error {
	if worldSpace.SchemaVersion != root.SchemaVersion {
		return fmt.Errorf(
			"world space %q field schema_version mismatch: expected %d, got %d",
			reference.WorldSpaceID,
			root.SchemaVersion,
			worldSpace.SchemaVersion,
		)
	}

	if worldSpace.WorldID != root.WorldID {
		return fmt.Errorf(
			"world space %q field world_id mismatch: expected %q, got %q",
			reference.WorldSpaceID,
			root.WorldID,
			worldSpace.WorldID,
		)
	}

	if worldSpace.WorldVersion != root.WorldVersion {
		return fmt.Errorf(
			"world space %q field world_version mismatch: expected %d, got %d",
			reference.WorldSpaceID,
			root.WorldVersion,
			worldSpace.WorldVersion,
		)
	}

	if worldSpace.WorldSpaceID != reference.WorldSpaceID {
		return fmt.Errorf(
			"world space %q field world_space_id mismatch: expected %q, got %q",
			reference.WorldSpaceID,
			reference.WorldSpaceID,
			worldSpace.WorldSpaceID,
		)
	}

	if worldSpace.DisplayName != reference.DisplayName {
		return fmt.Errorf(
			"world space %q field display_name mismatch: expected %q, got %q",
			reference.WorldSpaceID,
			reference.DisplayName,
			worldSpace.DisplayName,
		)
	}

	if worldSpace.GeographicPosition !=
		reference.GeographicPosition {
		return fmt.Errorf(
			"world space %q field geographic_position mismatch: expected %q, got %q",
			reference.WorldSpaceID,
			reference.GeographicPosition,
			worldSpace.GeographicPosition,
		)
	}

	return nil
}

func checkForExtraManifests(
	root *WorldManifest,
	rootDir string,
	rootManifestPath string,
) error {
	referencedPaths := make(map[string]struct{})
	directoriesToCheck := make(map[string]struct{})

	for _, reference := range root.WorldSpaces {
		resolvedPath, err := secureResolvePath(
			rootDir,
			reference.ManifestFile,
		)
		if err != nil {
			return fmt.Errorf(
				"resolve referenced world space %q for extra-file check: %w",
				reference.WorldSpaceID,
				err,
			)
		}

		referencedPaths[filepath.Clean(resolvedPath)] = struct{}{}
		directoriesToCheck[filepath.Dir(resolvedPath)] = struct{}{}
	}

	absoluteRootPath, err := filepath.Abs(rootManifestPath)
	if err != nil {
		return fmt.Errorf(
			"resolve root manifest path for extra-file check %q: %w",
			rootManifestPath,
			err,
		)
	}
	absoluteRootPath = filepath.Clean(absoluteRootPath)

	directories := make([]string, 0, len(directoriesToCheck))
	for directory := range directoriesToCheck {
		directories = append(directories, directory)
	}
	sort.Strings(directories)

	for _, directory := range directories {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf(
				"read world-space manifest directory %q: %w",
				directory,
				err,
			)
		}

		for _, entry := range entries {
			if entry.IsDir() ||
				!strings.EqualFold(
					filepath.Ext(entry.Name()),
					".json",
				) {
				continue
			}

			fullPath, err := filepath.Abs(
				filepath.Join(directory, entry.Name()),
			)
			if err != nil {
				return fmt.Errorf(
					"resolve candidate manifest path %q: %w",
					entry.Name(),
					err,
				)
			}
			fullPath = filepath.Clean(fullPath)

			if fullPath == absoluteRootPath {
				continue
			}

			if _, referenced := referencedPaths[fullPath]; !referenced {
				return fmt.Errorf(
					"unexpected world space manifest file %q",
					fullPath,
				)
			}
		}
	}

	return nil
}
