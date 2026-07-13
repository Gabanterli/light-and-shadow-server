package worldmap

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const maxChunkDocumentBytes int64 = 1 << 20

// ChunkDocumentRequest identifies one explicitly requested published chunk.
type ChunkDocumentRequest struct {
	WorldSpaceID WorldSpaceID
	Coordinate   ChunkCoordinate
}

// ChunkDocumentLoader is an immutable, read-only loader for explicitly
// requested production chunk documents. It stores no decoded chunk cache.
type ChunkDocumentLoader struct {
	rootDirectory         string
	provider              *ProductionProvider
	worldSpaceDirectories map[WorldSpaceID]string
}

// NewChunkDocumentLoader constructs an immutable loader using an already
// validated manifest snapshot and a physical root manifest path.
func NewChunkDocumentLoader(
	rootManifestPath string,
	snapshot *ManifestSnapshot,
) (*ChunkDocumentLoader, error) {
	if strings.TrimSpace(rootManifestPath) == "" {
		return nil, ErrEmptyRootManifestPath
	}

	if snapshot == nil {
		return nil, ErrNilManifestSnapshot
	}

	resolvedRootPath, err := resolveExistingRegularFile(
		rootManifestPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve root world manifest path %q: %w",
			rootManifestPath,
			err,
		)
	}

	rootDirectory := filepath.Dir(resolvedRootPath)

	provider, err := NewProductionProvider(snapshot)
	if err != nil {
		return nil, fmt.Errorf(
			"construct production provider for chunk document loader: %w",
			err,
		)
	}

	rootManifest := snapshot.RootManifest()

	worldSpaceDirectories := make(
		map[WorldSpaceID]string,
		len(rootManifest.WorldSpaces),
	)

	for _, reference := range rootManifest.WorldSpaces {
		resolvedManifestPath, err :=
			resolvePublishedWorldSpaceManifestPath(
				rootDirectory,
				reference,
			)
		if err != nil {
			return nil, err
		}

		worldSpaceDirectories[reference.WorldSpaceID] =
			filepath.Dir(resolvedManifestPath)
	}

	return &ChunkDocumentLoader{
		rootDirectory:         rootDirectory,
		provider:              provider,
		worldSpaceDirectories: worldSpaceDirectories,
	}, nil
}

// Load atomically materializes only the explicitly requested chunks. Requests
// are copied and sorted; the caller-owned slice is never modified.
func (l *ChunkDocumentLoader) Load(
	requests []ChunkDocumentRequest,
) ([]ChunkDocument, error) {
	if len(requests) == 0 {
		return []ChunkDocument{}, nil
	}

	ordered := append(
		[]ChunkDocumentRequest(nil),
		requests...,
	)

	sort.Slice(
		ordered,
		func(left int, right int) bool {
			leftRequest := ordered[left]
			rightRequest := ordered[right]

			if leftRequest.WorldSpaceID !=
				rightRequest.WorldSpaceID {
				return leftRequest.WorldSpaceID <
					rightRequest.WorldSpaceID
			}

			if leftRequest.Coordinate.Z !=
				rightRequest.Coordinate.Z {
				return leftRequest.Coordinate.Z <
					rightRequest.Coordinate.Z
			}

			if leftRequest.Coordinate.ChunkY !=
				rightRequest.Coordinate.ChunkY {
				return leftRequest.Coordinate.ChunkY <
					rightRequest.Coordinate.ChunkY
			}

			return leftRequest.Coordinate.ChunkX <
				rightRequest.Coordinate.ChunkX
		},
	)

	for requestIndex := 1; requestIndex < len(ordered); requestIndex++ {
		previous := ordered[requestIndex-1]
		current := ordered[requestIndex]

		if previous == current {
			return nil,
				&DuplicateChunkDocumentRequestError{
					WorldSpaceID: current.WorldSpaceID,
					Coordinate:   current.Coordinate,
				}
		}
	}

	documents := make(
		[]ChunkDocument,
		0,
		len(ordered),
	)

	for _, request := range ordered {
		document, err := l.loadOne(request)
		if err != nil {
			return nil, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (l *ChunkDocumentLoader) loadOne(
	request ChunkDocumentRequest,
) (ChunkDocument, error) {
	reference, err := l.provider.ChunkReference(
		request.WorldSpaceID,
		request.Coordinate,
	)
	if err != nil {
		return ChunkDocument{}, err
	}

	worldSpaceDirectory, found :=
		l.worldSpaceDirectories[request.WorldSpaceID]
	if !found {
		return ChunkDocument{},
			&WorldSpaceNotFoundError{
				WorldSpaceID: request.WorldSpaceID,
			}
	}

	resolvedChunkPath, err :=
		l.resolvePublishedChunkPath(
			worldSpaceDirectory,
			request,
			reference.File,
		)
	if err != nil {
		return ChunkDocument{}, err
	}

	raw, err := readBoundedChunkDocument(
		resolvedChunkPath,
		request,
	)
	if err != nil {
		return ChunkDocument{}, err
	}

	if err := validatePublishedChunkContentHash(
		reference.ContentHash,
		raw,
		request,
	); err != nil {
		return ChunkDocument{}, err
	}

	var document ChunkDocument

	if err := decodeStrictJSON(raw, &document); err != nil {
		return ChunkDocument{}, fmt.Errorf(
			"decode chunk document for world space %q at chunk (%d,%d,%d) from %q: %w",
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			resolvedChunkPath,
			err,
		)
	}

	if err := ValidateChunkDocument(document); err != nil {
		return ChunkDocument{}, fmt.Errorf(
			"validate chunk document for world space %q at chunk (%d,%d,%d) from %q: %w",
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			resolvedChunkPath,
			err,
		)
	}

	if err := l.validateDocumentIdentity(
		request,
		document,
	); err != nil {
		return ChunkDocument{}, err
	}

	return document, nil
}

func (l *ChunkDocumentLoader) validateDocumentIdentity(
	request ChunkDocumentRequest,
	document ChunkDocument,
) error {
	checks := []struct {
		field    string
		expected string
		actual   string
	}{
		{
			field:    "world_id",
			expected: l.provider.WorldID(),
			actual:   document.WorldID,
		},
		{
			field: "world_version",
			expected: strconv.Itoa(
				l.provider.Version(),
			),
			actual: strconv.Itoa(
				document.WorldVersion,
			),
		},
		{
			field: "world_space_id",
			expected: string(
				request.WorldSpaceID,
			),
			actual: string(
				document.WorldSpaceID,
			),
		},
		{
			field: "chunk_x",
			expected: strconv.Itoa(
				request.Coordinate.ChunkX,
			),
			actual: strconv.Itoa(
				document.ChunkX,
			),
		},
		{
			field: "chunk_y",
			expected: strconv.Itoa(
				request.Coordinate.ChunkY,
			),
			actual: strconv.Itoa(
				document.ChunkY,
			),
		},
		{
			field: "z",
			expected: strconv.Itoa(
				request.Coordinate.Z,
			),
			actual: strconv.Itoa(
				document.Z,
			),
		},
		{
			field: "width",
			expected: strconv.Itoa(
				CanonicalChunkSize,
			),
			actual: strconv.Itoa(
				document.Width,
			),
		},
		{
			field: "height",
			expected: strconv.Itoa(
				CanonicalChunkSize,
			),
			actual: strconv.Itoa(
				document.Height,
			),
		},
	}

	for _, check := range checks {
		if check.expected == check.actual {
			continue
		}

		return &ChunkDocumentIdentityMismatchError{
			WorldSpaceID: request.WorldSpaceID,
			Coordinate:   request.Coordinate,
			Field:        check.field,
			Expected:     check.expected,
			Actual:       check.actual,
		}
	}

	return nil
}

func resolvePublishedWorldSpaceManifestPath(
	rootDirectory string,
	reference WorldSpaceReference,
) (string, error) {
	if err := validateAssetPath(
		reference.ManifestFile,
	); err != nil {
		return "", fmt.Errorf(
			"invalid world space %q manifest path %q: %w",
			reference.WorldSpaceID,
			reference.ManifestFile,
			err,
		)
	}

	candidatePath := filepath.Join(
		rootDirectory,
		filepath.FromSlash(reference.ManifestFile),
	)

	resolvedPath, err := filepath.EvalSymlinks(
		candidatePath,
	)
	if err != nil {
		return "", fmt.Errorf(
			"resolve world space %q manifest path %q: %w",
			reference.WorldSpaceID,
			candidatePath,
			err,
		)
	}

	resolvedPath, err = filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf(
			"resolve absolute world space %q manifest path %q: %w",
			reference.WorldSpaceID,
			resolvedPath,
			err,
		)
	}
	resolvedPath = filepath.Clean(resolvedPath)

	if !pathIsWithinRoot(rootDirectory, resolvedPath) {
		return "",
			&WorldSpaceManifestPathOutsideRootError{
				WorldSpaceID: reference.WorldSpaceID,
				Path:         resolvedPath,
				Root:         rootDirectory,
			}
	}

	if err := requireRegularFile(resolvedPath); err != nil {
		return "", fmt.Errorf(
			"world space %q manifest path %q: %w",
			reference.WorldSpaceID,
			resolvedPath,
			err,
		)
	}

	return resolvedPath, nil
}

func (l *ChunkDocumentLoader) resolvePublishedChunkPath(
	worldSpaceDirectory string,
	request ChunkDocumentRequest,
	assetPath string,
) (string, error) {
	if err := validateAssetPath(assetPath); err != nil {
		return "", fmt.Errorf(
			"invalid chunk asset path %q for world space %q at chunk (%d,%d,%d): %w",
			assetPath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}

	candidatePath := filepath.Join(
		worldSpaceDirectory,
		filepath.FromSlash(assetPath),
	)

	resolvedPath, err := filepath.EvalSymlinks(
		candidatePath,
	)
	if err != nil {
		return "", fmt.Errorf(
			"resolve chunk path %q for world space %q at chunk (%d,%d,%d): %w",
			candidatePath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}

	resolvedPath, err = filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf(
			"resolve absolute chunk path %q for world space %q at chunk (%d,%d,%d): %w",
			resolvedPath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}
	resolvedPath = filepath.Clean(resolvedPath)

	if !pathIsWithinRoot(
		l.rootDirectory,
		resolvedPath,
	) {
		return "",
			&ChunkDocumentPathOutsideRootError{
				WorldSpaceID: request.WorldSpaceID,
				Coordinate:   request.Coordinate,
				Path:         resolvedPath,
				Root:         l.rootDirectory,
			}
	}

	if err := requireRegularFile(resolvedPath); err != nil {
		return "", fmt.Errorf(
			"chunk path %q for world space %q at chunk (%d,%d,%d): %w",
			resolvedPath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}

	return resolvedPath, nil
}

func resolveExistingRegularFile(
	filePath string,
) (string, error) {
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	absolutePath = filepath.Clean(absolutePath)

	resolvedPath, err := filepath.EvalSymlinks(
		absolutePath,
	)
	if err != nil {
		return "", err
	}

	resolvedPath, err = filepath.Abs(resolvedPath)
	if err != nil {
		return "", err
	}
	resolvedPath = filepath.Clean(resolvedPath)

	if err := requireRegularFile(resolvedPath); err != nil {
		return "", err
	}

	return resolvedPath, nil
}

func requireRegularFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf(
			"path is not a regular file",
		)
	}

	return nil
}

func pathIsWithinRoot(
	root string,
	target string,
) bool {
	relativePath, err := filepath.Rel(
		filepath.Clean(root),
		filepath.Clean(target),
	)
	if err != nil {
		return false
	}

	if filepath.IsAbs(relativePath) {
		return false
	}

	if relativePath == ".." {
		return false
	}

	parentPrefix := ".." + string(os.PathSeparator)

	return !strings.HasPrefix(
		relativePath,
		parentPrefix,
	)
}

func readBoundedChunkDocument(
	filePath string,
	request ChunkDocumentRequest,
) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"open chunk document %q for world space %q at chunk (%d,%d,%d): %w",
			filePath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}
	defer file.Close()

	raw, err := io.ReadAll(
		io.LimitReader(
			file,
			maxChunkDocumentBytes+1,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"read chunk document %q for world space %q at chunk (%d,%d,%d): %w",
			filePath,
			request.WorldSpaceID,
			request.Coordinate.ChunkX,
			request.Coordinate.ChunkY,
			request.Coordinate.Z,
			err,
		)
	}

	if int64(len(raw)) > maxChunkDocumentBytes {
		return nil, &ChunkDocumentTooLargeError{
			WorldSpaceID: request.WorldSpaceID,
			Coordinate:   request.Coordinate,
			MaximumBytes: maxChunkDocumentBytes,
		}
	}

	return raw, nil
}

func validatePublishedChunkContentHash(
	contentHash string,
	raw []byte,
	request ChunkDocumentRequest,
) error {
	if contentHash == "" {
		return nil
	}

	if !validCanonicalChunkContentHash(
		contentHash,
	) {
		return &InvalidChunkContentHashError{
			WorldSpaceID: request.WorldSpaceID,
			Coordinate:   request.Coordinate,
			ContentHash:  contentHash,
		}
	}

	sum := sha256.Sum256(raw)
	actual := fmt.Sprintf("sha256:%x", sum)

	if actual != contentHash {
		return &ChunkContentHashMismatchError{
			WorldSpaceID: request.WorldSpaceID,
			Coordinate:   request.Coordinate,
			Expected:     contentHash,
			Actual:       actual,
		}
	}

	return nil
}

func validCanonicalChunkContentHash(
	contentHash string,
) bool {
	const prefix = "sha256:"
	const encodedLength = 64

	if len(contentHash) != len(prefix)+encodedLength {
		return false
	}

	if !strings.HasPrefix(contentHash, prefix) {
		return false
	}

	for _, character := range contentHash[len(prefix):] {
		if character >= '0' && character <= '9' {
			continue
		}

		if character >= 'a' && character <= 'f' {
			continue
		}

		return false
	}

	return true
}
