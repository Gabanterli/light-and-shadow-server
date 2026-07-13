package worldmap

// ManifestSnapshot provides an immutable, thread-safe snapshot of the entire
// world manifest configuration. It is created by LoadManifestSnapshot and is
// guaranteed to be valid and consistent.
type ManifestSnapshot struct {
	rootManifest  WorldManifest
	worldSpaces   map[WorldSpaceID]WorldSpaceManifest
	worldSpaceIDs []WorldSpaceID
}

// WorldID returns the canonical world ID.
func (s *ManifestSnapshot) WorldID() string {
	return s.rootManifest.WorldID
}

// Version returns the canonical world version.
func (s *ManifestSnapshot) Version() int {
	return s.rootManifest.WorldVersion
}

// RootManifest returns a defensive copy of the root WorldManifest.
// Modifying the returned value will not affect the snapshot.
func (s *ManifestSnapshot) RootManifest() WorldManifest {
	// Create a copy of the root manifest.
	rootCopy := s.rootManifest
	// Create a deep copy of the WorldSpaces slice.
	rootCopy.WorldSpaces = make([]WorldSpaceReference, len(s.rootManifest.WorldSpaces))
	copy(rootCopy.WorldSpaces, s.rootManifest.WorldSpaces)
	return rootCopy
}

// WorldSpaceIDs returns a defensive copy of the list of all canonical WorldSpaceIDs.
// Modifying the returned slice will not affect the snapshot.
func (s *ManifestSnapshot) WorldSpaceIDs() []WorldSpaceID {
	idsCopy := make([]WorldSpaceID, len(s.worldSpaceIDs))
	copy(idsCopy, s.worldSpaceIDs)
	return idsCopy
}

// WorldSpace returns a defensive copy of the manifest for a specific world space.
// The second return value is false if the world space does not exist.
// Modifying the returned value will not affect the snapshot.
func (s *ManifestSnapshot) WorldSpace(id WorldSpaceID) (WorldSpaceManifest, bool) {
	ws, ok := s.worldSpaces[id]
	if !ok {
		return WorldSpaceManifest{}, false
	}

	// Create a copy of the world space manifest.
	wsCopy := ws
	// Create a deep copy of the Chunks slice.
	wsCopy.Chunks = make([]ChunkReference, len(ws.Chunks))
	copy(wsCopy.Chunks, ws.Chunks)

	return wsCopy, true
}
