package main

import (
	"strings"
	"sync"
)

type activeCharacterSessionLease struct {
	generation uint64
}

// activeCharacterSessionRegistry owns the in-process lifecycle lease for each
// active character. A lease remains held through the entire disconnect cleanup,
// preventing a new session from activating while old state is still being
// removed from inventory, movement, AOI, combat, quest, or social systems.
type activeCharacterSessionRegistry struct {
	mu sync.RWMutex

	nextGeneration uint64
	sessions       map[string]activeCharacterSessionLease
}

func newActiveCharacterSessionRegistry() *activeCharacterSessionRegistry {
	return &activeCharacterSessionRegistry{
		sessions: make(map[string]activeCharacterSessionLease),
	}
}

// TryAcquire atomically grants a character lifecycle lease when no lease is
// active for playerID.
func (r *activeCharacterSessionRegistry) TryAcquire(
	playerID string,
) (activeCharacterSessionLease, bool) {
	playerID = strings.TrimSpace(playerID)
	if r == nil || playerID == "" {
		return activeCharacterSessionLease{}, false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[playerID]; exists {
		return activeCharacterSessionLease{}, false
	}

	r.nextGeneration++
	if r.nextGeneration == 0 {
		r.nextGeneration++
	}

	lease := activeCharacterSessionLease{
		generation: r.nextGeneration,
	}

	r.sessions[playerID] = lease

	return lease, true
}

// Owns reports whether lease is still the authoritative lifecycle owner for
// playerID.
func (r *activeCharacterSessionRegistry) Owns(
	playerID string,
	lease activeCharacterSessionLease,
) bool {
	playerID = strings.TrimSpace(playerID)
	if r == nil ||
		playerID == "" ||
		lease.generation == 0 {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	current, exists := r.sessions[playerID]
	return exists &&
		current.generation == lease.generation
}

// Release removes the active lease only when the supplied generation still
// owns playerID. A stale session can therefore never release a newer session.
func (r *activeCharacterSessionRegistry) Release(
	playerID string,
	lease activeCharacterSessionLease,
) bool {
	playerID = strings.TrimSpace(playerID)
	if r == nil ||
		playerID == "" ||
		lease.generation == 0 {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	current, exists := r.sessions[playerID]
	if !exists ||
		current.generation != lease.generation {
		return false
	}

	delete(r.sessions, playerID)
	return true
}

// IsActive reports whether playerID currently has an acquisition or active
// session whose cleanup has not completed.
func (r *activeCharacterSessionRegistry) IsActive(
	playerID string,
) bool {
	playerID = strings.TrimSpace(playerID)
	if r == nil || playerID == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.sessions[playerID]
	return exists
}
