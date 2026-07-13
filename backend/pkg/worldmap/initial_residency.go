package worldmap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxInitialResidencyBytes int64 = 1 << 20

// ResidentChunk identifies a chunk that must be loaded during the initial
// production residency bootstrap.
type ResidentChunk struct {
	WorldSpaceID WorldSpaceID `json:"world_space_id"`
	ChunkX       int          `json:"chunk_x"`
	ChunkY       int          `json:"chunk_y"`
	Z            int          `json:"z"`
}

// Coordinate returns the chunk-local coordinate without retaining references
// to the residency document.
func (r ResidentChunk) Coordinate() ChunkCoordinate {
	return ChunkCoordinate{
		ChunkX: r.ChunkX,
		ChunkY: r.ChunkY,
		Z:      r.Z,
	}
}

// InitialResidencyManifest is the persisted operational configuration that
// selects the chunks loaded during a future production bootstrap.
type InitialResidencyManifest struct {
	SchemaVersion  int             `json:"schema_version"`
	WorldID        string          `json:"world_id"`
	WorldVersion   int             `json:"world_version"`
	ResidentChunks []ResidentChunk `json:"resident_chunks"`
}

// InitialResidencySnapshot is an immutable in-memory copy of a validated
// initial residency manifest.
type InitialResidencySnapshot struct {
	manifest InitialResidencyManifest
}

// ValidateInitialResidencyManifest validates only the residency document.
// Publication and world-manifest consistency are intentionally validated by
// the future production bootstrap.
func ValidateInitialResidencyManifest(
	manifest InitialResidencyManifest,
) error {
	if manifest.SchemaVersion != SupportedSchemaVersion {
		return fmt.Errorf(
			"initial residency schema_version = %d, want %d",
			manifest.SchemaVersion,
			SupportedSchemaVersion,
		)
	}

	if strings.TrimSpace(manifest.WorldID) == "" {
		return fmt.Errorf(
			"initial residency world_id cannot be empty",
		)
	}

	if manifest.WorldVersion <= 0 {
		return fmt.Errorf(
			"initial residency world_version must be positive, got %d",
			manifest.WorldVersion,
		)
	}

	if len(manifest.ResidentChunks) == 0 {
		return ErrInitialResidencyEmpty
	}

	type residentChunkKey struct {
		worldSpaceID WorldSpaceID
		coordinate   ChunkCoordinate
	}

	seen := make(
		map[residentChunkKey]int,
		len(manifest.ResidentChunks),
	)

	for index, resident := range manifest.ResidentChunks {
		if strings.TrimSpace(string(resident.WorldSpaceID)) == "" {
			return fmt.Errorf(
				"initial residency resident_chunks[%d].world_space_id cannot be empty",
				index,
			)
		}

		if resident.ChunkX < 0 {
			return fmt.Errorf(
				"initial residency resident_chunks[%d].chunk_x cannot be negative: %d",
				index,
				resident.ChunkX,
			)
		}

		if resident.ChunkY < 0 {
			return fmt.Errorf(
				"initial residency resident_chunks[%d].chunk_y cannot be negative: %d",
				index,
				resident.ChunkY,
			)
		}

		if resident.Z < 0 {
			return fmt.Errorf(
				"initial residency resident_chunks[%d].z cannot be negative: %d",
				index,
				resident.Z,
			)
		}

		key := residentChunkKey{
			worldSpaceID: resident.WorldSpaceID,
			coordinate:   resident.Coordinate(),
		}

		if firstIndex, found := seen[key]; found {
			return &DuplicateResidentChunkError{
				WorldSpaceID: resident.WorldSpaceID,
				Coordinate:   resident.Coordinate(),
				FirstIndex:   firstIndex,
				SecondIndex:  index,
			}
		}

		seen[key] = index
	}

	return nil
}

// LoadInitialResidencySnapshot loads one explicitly selected residency
// manifest and returns an immutable, canonically ordered snapshot.
func LoadInitialResidencySnapshot(
	residencyPath string,
) (*InitialResidencySnapshot, error) {
	if strings.TrimSpace(residencyPath) == "" {
		return nil, ErrInitialResidencyPathEmpty
	}

	absolutePath, err := filepath.Abs(residencyPath)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve initial residency path %q: %w",
			residencyPath,
			err,
		)
	}

	absolutePath = filepath.Clean(absolutePath)

	file, err := os.Open(absolutePath)
	if err != nil {
		return nil, fmt.Errorf(
			"open initial residency %q: %w",
			absolutePath,
			err,
		)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf(
			"stat initial residency %q: %w",
			absolutePath,
			err,
		)
	}

	if info.IsDir() {
		return nil, fmt.Errorf(
			"initial residency path %q points to a directory",
			absolutePath,
		)
	}

	if info.Size() > maxInitialResidencyBytes {
		return nil, &InitialResidencyFileTooLargeError{
			Path:  absolutePath,
			Size:  info.Size(),
			Limit: maxInitialResidencyBytes,
		}
	}

	data, err := io.ReadAll(
		io.LimitReader(
			file,
			maxInitialResidencyBytes+1,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"read initial residency %q: %w",
			absolutePath,
			err,
		)
	}

	if int64(len(data)) > maxInitialResidencyBytes {
		return nil, &InitialResidencyFileTooLargeError{
			Path:  absolutePath,
			Size:  int64(len(data)),
			Limit: maxInitialResidencyBytes,
		}
	}

	var manifest InitialResidencyManifest

	if err := decodeStrictJSON(data, &manifest); err != nil {
		return nil, fmt.Errorf(
			"decode initial residency %q: %w",
			absolutePath,
			err,
		)
	}

	if err := ValidateInitialResidencyManifest(manifest); err != nil {
		return nil, fmt.Errorf(
			"validate initial residency %q: %w",
			absolutePath,
			err,
		)
	}

	residentChunks := append(
		[]ResidentChunk(nil),
		manifest.ResidentChunks...,
	)

	sort.Slice(
		residentChunks,
		func(left int, right int) bool {
			leftChunk := residentChunks[left]
			rightChunk := residentChunks[right]

			if leftChunk.WorldSpaceID != rightChunk.WorldSpaceID {
				return leftChunk.WorldSpaceID < rightChunk.WorldSpaceID
			}

			if leftChunk.Z != rightChunk.Z {
				return leftChunk.Z < rightChunk.Z
			}

			if leftChunk.ChunkY != rightChunk.ChunkY {
				return leftChunk.ChunkY < rightChunk.ChunkY
			}

			return leftChunk.ChunkX < rightChunk.ChunkX
		},
	)

	manifest.ResidentChunks = residentChunks

	return &InitialResidencySnapshot{
		manifest: manifest,
	}, nil
}

// WorldID returns the logical world identifier copied during loading.
func (s *InitialResidencySnapshot) WorldID() string {
	return s.manifest.WorldID
}

// Version returns the world content version copied during loading.
func (s *InitialResidencySnapshot) Version() int {
	return s.manifest.WorldVersion
}

// Manifest returns a defensive copy of the complete residency manifest.
func (s *InitialResidencySnapshot) Manifest() InitialResidencyManifest {
	manifestCopy := s.manifest
	manifestCopy.ResidentChunks = append(
		[]ResidentChunk(nil),
		s.manifest.ResidentChunks...,
	)

	return manifestCopy
}

// ResidentChunks returns a defensive copy in canonical deterministic order.
func (s *InitialResidencySnapshot) ResidentChunks() []ResidentChunk {
	return append(
		[]ResidentChunk(nil),
		s.manifest.ResidentChunks...,
	)
}
