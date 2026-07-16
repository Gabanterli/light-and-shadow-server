package main

import (
	"errors"
	"sync"
	"testing"

	"github.com/light-and-shadow/backend/pkg/combat"
)

type scriptedOverlapResponse struct {
	started chan struct{}
	release <-chan struct{}
	result  bool
}

type fakeOverlapQuerier struct {
	mu          sync.Mutex
	overlapping map[string]bool
	callCounts  map[string]int
	scripted    map[string][]scriptedOverlapResponse
}

func newFakeOverlapQuerier() *fakeOverlapQuerier {
	return &fakeOverlapQuerier{
		overlapping: make(map[string]bool),
		callCounts:  make(map[string]int),
		scripted:    make(map[string][]scriptedOverlapResponse),
	}
}

func (f *fakeOverlapQuerier) SetOverlapping(playerID string, isOverlapping bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.overlapping[playerID] = isOverlapping
}

func (f *fakeOverlapQuerier) QueueResponse(playerID string, response scriptedOverlapResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scripted[playerID] = append(f.scripted[playerID], response)
}

func (f *fakeOverlapQuerier) HasOverlappingPlayer(playerID string) bool {
	f.mu.Lock()
	f.callCounts[playerID]++
	if len(f.scripted[playerID]) > 0 {
		response := f.scripted[playerID][0]
		f.scripted[playerID] = f.scripted[playerID][1:]
		f.mu.Unlock()

		if response.started != nil {
			close(response.started)
		}
		if response.release != nil {
			<-response.release
		}
		return response.result
	}

	isOverlapping := f.overlapping[playerID]
	f.mu.Unlock()
	return isOverlapping
}

func (f *fakeOverlapQuerier) CallCount(playerID string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCounts[playerID]
}

func requireBlocked(t *testing.T, err error, shouldNotify bool) *combat.ErrOffensiveActionBlocked {
	t.Helper()

	if err == nil {
		t.Fatal("expected offensive action to be blocked, got nil")
	}

	var blocked *combat.ErrOffensiveActionBlocked
	if !errors.As(err, &blocked) {
		t.Fatalf("error type = %T, want *combat.ErrOffensiveActionBlocked: %v", err, err)
	}

	if blocked.ShouldNotify != shouldNotify {
		t.Fatalf("ShouldNotify = %v, want %v", blocked.ShouldNotify, shouldNotify)
	}

	return blocked
}
func TestLoginOverlapRegistry_Constructor(t *testing.T) {
	t.Run("rejects nil querier", func(t *testing.T) {
		_, err := newLoginOverlapCombatRestrictionRegistry(nil)
		if err == nil {
			t.Fatal("expected error for nil querier, got nil")
		}
	})

	t.Run("succeeds with valid querier", func(t *testing.T) {
		querier := newFakeOverlapQuerier()
		registry, err := newLoginOverlapCombatRestrictionRegistry(querier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if registry == nil {
			t.Fatal("registry is nil")
		}
	})
}

func TestLoginOverlapRegistry_UnmarkedPlayer(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)

	err := registry.ValidateOffensiveAction("playerA")
	if err != nil {
		t.Fatalf("unmarked player should not be blocked, got: %v", err)
	}

	if querier.CallCount("playerA") > 0 {
		t.Fatal("querier was called for an unmarked player")
	}
}

func TestLoginOverlapRegistry_TwoPlayerScenario(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)

	playerA := "playerA"
	playerB := "playerB"

	// Player B logs in on top of Player A
	querier.SetOverlapping(playerB, true)
	registry.MarkRestrictedEntrant(playerB)

	if !registry.IsMarked(playerB) {
		t.Fatal("Player B was not marked")
	}
	if registry.IsMarked(playerA) {
		t.Fatal("Player A was incorrectly marked")
	}

	// Player A (the original occupant) can still attack
	if err := registry.ValidateOffensiveAction(playerA); err != nil {
		t.Fatalf("original occupant was blocked: %v", err)
	}

	if got := querier.CallCount(playerA); got != 0 {
		t.Fatalf("occupant querier calls = %d, want 0", got)
	}

	first := requireBlocked(
		t,
		registry.ValidateOffensiveAction(playerB),
		true,
	)

	if first.Reason != loginOverlapOffensiveActionBlockedReason {
		t.Fatalf(
			"Reason = %q, want %q",
			first.Reason,
			loginOverlapOffensiveActionBlockedReason,
		)
	}

	requireBlocked(
		t,
		registry.ValidateOffensiveAction(playerB),
		false,
	)

	querier.SetOverlapping(playerB, false)

	if err := registry.ValidateOffensiveAction(playerB); err != nil {
		t.Fatalf("entrant remained blocked after separation: %v", err)
	}

	if registry.IsMarked(playerB) {
		t.Fatal("entrant remained marked after separation")
	}
}

func TestLoginOverlapRegistry_ThreePlayerScenario(t *testing.T) {
	querier := newFakeOverlapQuerier()

	registry, err :=
		newLoginOverlapCombatRestrictionRegistry(querier)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	playerA, playerB, playerC := "playerA", "playerB", "playerC"

	// B and C log in on top of A
	querier.SetOverlapping(playerB, true)
	querier.SetOverlapping(playerC, true)
	registry.MarkRestrictedEntrant(playerB)
	registry.MarkRestrictedEntrant(playerC)

	// A is not restricted
	if err := registry.ValidateOffensiveAction(playerA); err != nil {
		t.Fatalf("Player A should not be restricted: %v", err)
	}

	// B and C are restricted
	if err := registry.ValidateOffensiveAction(playerB); err == nil {
		t.Fatal("Player B should be restricted")
	}
	if err := registry.ValidateOffensiveAction(playerC); err == nil {
		t.Fatal("Player C should be restricted")
	}

	// A leaves, but B and C are still overlapping each other
	// (The querier's state for B and C remains true)
	if err := registry.ValidateOffensiveAction(playerB); err == nil {
		t.Fatal("Player B should still be restricted after A leaves")
	}

	// B moves away, leaving C alone on the tile
	querier.SetOverlapping(playerB, false) // B is no longer overlapping
	querier.SetOverlapping(playerC, false) // C is now alone, so not overlapping

	// B's next action clears their restriction
	if err := registry.ValidateOffensiveAction(playerB); err != nil {
		t.Fatalf("Player B should be free after moving: %v", err)
	}
	if registry.IsMarked(playerB) {
		t.Fatal("Player B should be unmarked")
	}

	// C's next action clears their restriction
	if err := registry.ValidateOffensiveAction(playerC); err != nil {
		t.Fatalf("Player C should be free after B moves: %v", err)
	}
	if registry.IsMarked(playerC) {
		t.Fatal("Player C should be unmarked")
	}
}

func TestLoginOverlapRegistry_Clear(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)
	playerID := "testPlayer"

	registry.MarkRestrictedEntrant(playerID)
	if !registry.IsMarked(playerID) {
		t.Fatal("Failed to mark player")
	}

	registry.Clear(playerID)
	if registry.IsMarked(playerID) {
		t.Fatal("Clear did not remove the restriction")
	}

	// Idempotency check
	registry.Clear(playerID)
	if registry.IsMarked(playerID) {
		t.Fatal("Second Clear call had an unexpected effect")
	}
}

func TestLoginOverlapRegistry_Concurrency(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)
	playerID := "concurrentPlayer"

	var wg sync.WaitGroup
	const numGoroutines = 100

	querier.SetOverlapping(playerID, true)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%3 == 0 {
				registry.MarkRestrictedEntrant(playerID)
			} else if i%3 == 1 {
				_ = registry.ValidateOffensiveAction(playerID)
			} else {
				registry.Clear(playerID)
			}
		}(i)
	}

	wg.Wait()
}

func TestLoginOverlapRegistry_NewEpisodeRestoresNotification(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)
	playerID := "player"
	querier.SetOverlapping(playerID, true)

	// Episode 1
	registry.MarkRestrictedEntrant(playerID)
	err1 := registry.ValidateOffensiveAction(playerID)
	var blockedErr *combat.ErrOffensiveActionBlocked
	if !errors.As(err1, &blockedErr) || !blockedErr.ShouldNotify {
		t.Fatal("Episode 1, first attempt should notify")
	}
	err2 := registry.ValidateOffensiveAction(playerID)
	if !errors.As(err2, &blockedErr) || blockedErr.ShouldNotify {
		t.Fatal("Episode 1, second attempt should be silent")
	}

	// Episode 2
	registry.MarkRestrictedEntrant(playerID)
	err3 := registry.ValidateOffensiveAction(playerID)
	if !errors.As(err3, &blockedErr) || !blockedErr.ShouldNotify {
		t.Fatal("Episode 2, first attempt should notify again")
	}
}

func TestLoginOverlapRegistry_StaleQueryCannotReleaseNewEpisode(t *testing.T) {
	querier := newFakeOverlapQuerier()
	registry, _ := newLoginOverlapCombatRestrictionRegistry(querier)
	playerID := "player"
	querier.SetOverlapping(playerID, true)

	// 1. Mark B as episode 1
	registry.MarkRestrictedEntrant(playerID)

	// 2. Queue a stale query result that will block
	releaseChan1 := make(chan struct{})
	startedChan1 := make(chan struct{})
	querier.QueueResponse(playerID, scriptedOverlapResponse{
		result:  false, // Stale result: player is not overlapping
		release: releaseChan1,
		started: startedChan1,
	})
	// 3. Queue the second, correct query result
	querier.QueueResponse(playerID, scriptedOverlapResponse{
		result: true, // Correct result: player is overlapping
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 4. Start validation, which will block on the first query
		_ = registry.ValidateOffensiveAction(playerID)
	}()

	// 5. Wait for the first query to start
	<-startedChan1

	// 6. Clear the old episode and mark a new one
	registry.Clear(playerID)
	registry.MarkRestrictedEntrant(playerID)

	// 7. Unblock the first query, which will return the stale 'false'
	close(releaseChan1)

	// Wait for the goroutine to finish
	wg.Wait()

	// 8. The final result should be a block, because the loop retried
	finalErr := registry.ValidateOffensiveAction(playerID)
	var blockedErr *combat.ErrOffensiveActionBlocked
	if !errors.As(finalErr, &blockedErr) {
		t.Fatalf("Expected player to be blocked, but got err: %v", finalErr)
	}

	// 9. The notification should be for the new episode (already sent in the goroutine)
	if blockedErr.ShouldNotify {
		t.Error("Expected second block to be silent, but it had notify=true")
	}

	// 10. Player should still be marked
	if !registry.IsMarked(playerID) {
		t.Error("Player should still be marked")
	}

	// 11. Querier should have been called at least twice
	if querier.CallCount(playerID) < 2 {
		t.Errorf("Expected querier to be called at least twice, got %d", querier.CallCount(playerID))
	}
}
