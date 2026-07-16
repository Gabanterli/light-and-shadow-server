package main

import (
	"errors"
	"strings"
	"sync"

	"github.com/light-and-shadow/backend/pkg/combat"
)

const loginOverlapOffensiveActionBlockedReason = "Você precisa sair de cima de outro jogador para usar ações ofensivas."

type playerOverlapQuerier interface {
	HasOverlappingPlayer(playerID string) bool
}

type restrictionEpisode struct {
	generation   uint64
	feedbackSent bool
}

// loginOverlapCombatRestrictionRegistry tracks only players who entered an
// already-occupied tile through the authoritative login overlap exception.
type loginOverlapCombatRestrictionRegistry struct {
	mu sync.RWMutex

	querier        playerOverlapQuerier
	nextGeneration uint64
	entrants       map[string]restrictionEpisode
}

func newLoginOverlapCombatRestrictionRegistry(
	querier playerOverlapQuerier,
) (*loginOverlapCombatRestrictionRegistry, error) {
	if querier == nil {
		return nil, errors.New("overlap querier cannot be nil")
	}

	return &loginOverlapCombatRestrictionRegistry{
		querier:  querier,
		entrants: make(map[string]restrictionEpisode),
	}, nil
}

// MarkRestrictedEntrant starts a new restriction episode for playerID.
func (r *loginOverlapCombatRestrictionRegistry) MarkRestrictedEntrant(
	playerID string,
) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextGeneration++
	if r.nextGeneration == 0 {
		r.nextGeneration++
	}

	r.entrants[playerID] = restrictionEpisode{
		generation: r.nextGeneration,
	}
}

// Clear removes the current episode without resetting the monotonic generation.
func (r *loginOverlapCombatRestrictionRegistry) Clear(playerID string) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.entrants, playerID)
}

// IsMarked reports whether playerID currently has a restriction episode.
func (r *loginOverlapCombatRestrictionRegistry) IsMarked(
	playerID string,
) bool {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.entrants[playerID]
	return exists
}

// ValidateOffensiveAction blocks only marked entrants while their overlap
// remains. World queries are intentionally executed without the registry lock.
func (r *loginOverlapCombatRestrictionRegistry) ValidateOffensiveAction(
	playerID string,
) error {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil
	}

	for {
		r.mu.RLock()
		episode, marked := r.entrants[playerID]
		r.mu.RUnlock()

		if !marked {
			return nil
		}

		overlapping := r.querier.HasOverlappingPlayer(playerID)

		r.mu.Lock()

		current, stillMarked := r.entrants[playerID]
		if !stillMarked {
			r.mu.Unlock()
			return nil
		}

		if current.generation != episode.generation {
			r.mu.Unlock()
			continue
		}

		if !overlapping {
			delete(r.entrants, playerID)
			r.mu.Unlock()
			return nil
		}

		shouldNotify := !current.feedbackSent
		if shouldNotify {
			current.feedbackSent = true
			r.entrants[playerID] = current
		}

		r.mu.Unlock()

		return &combat.ErrOffensiveActionBlocked{
			Reason:       loginOverlapOffensiveActionBlockedReason,
			ShouldNotify: shouldNotify,
		}
	}
}
