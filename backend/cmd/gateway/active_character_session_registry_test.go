package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestActiveCharacterSessionRegistryAcquireAndRelease(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	lease, acquired :=
		registry.TryAcquire("Gabriela")
	if !acquired {
		t.Fatal("first lease was not acquired")
	}

	if lease.generation == 0 {
		t.Fatal("acquired lease has zero generation")
	}

	if !registry.IsActive("Gabriela") {
		t.Fatal("acquired character is not active")
	}

	if !registry.Owns("Gabriela", lease) {
		t.Fatal("acquired lease is not recognized as owner")
	}

	if _, duplicate :=
		registry.TryAcquire("Gabriela"); duplicate {
		t.Fatal("duplicate lease was acquired")
	}

	if !registry.Release("Gabriela", lease) {
		t.Fatal("authoritative lease was not released")
	}

	if registry.IsActive("Gabriela") {
		t.Fatal("released character remains active")
	}

	if registry.Owns("Gabriela", lease) {
		t.Fatal("released lease remains owner")
	}
}

func TestActiveCharacterSessionRegistryRejectsInvalidInput(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	invalidIDs := []string{
		"",
		" ",
		"\t",
		"\r\n",
	}

	for _, playerID := range invalidIDs {
		playerID := playerID

		t.Run(
			fmt.Sprintf("id_%q", playerID),
			func(t *testing.T) {
				if _, acquired :=
					registry.TryAcquire(playerID); acquired {
					t.Fatal(
						"invalid player ID acquired a lease",
					)
				}

				if registry.IsActive(playerID) {
					t.Fatal(
						"invalid player ID reported active",
					)
				}

				if registry.Owns(
					playerID,
					activeCharacterSessionLease{
						generation: 1,
					},
				) {
					t.Fatal(
						"invalid player ID reported ownership",
					)
				}

				if registry.Release(
					playerID,
					activeCharacterSessionLease{
						generation: 1,
					},
				) {
					t.Fatal(
						"invalid player ID released a lease",
					)
				}
			},
		)
	}
}

func TestActiveCharacterSessionRegistryTrimsPlayerID(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	lease, acquired :=
		registry.TryAcquire("  Gabriela  ")
	if !acquired {
		t.Fatal("trimmed character lease was not acquired")
	}

	if !registry.IsActive("Gabriela") {
		t.Fatal("trimmed character is not active")
	}

	if _, duplicate :=
		registry.TryAcquire("Gabriela"); duplicate {
		t.Fatal(
			"canonical ID bypassed trimmed duplicate protection",
		)
	}

	if !registry.Release(" Gabriela ", lease) {
		t.Fatal(
			"trimmed release did not recognize ownership",
		)
	}
}

func TestActiveCharacterSessionRegistryStaleLeaseCannotReleaseNewOwner(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	firstLease, acquired :=
		registry.TryAcquire("Gabriela")
	if !acquired {
		t.Fatal("first lease was not acquired")
	}

	if !registry.Release("Gabriela", firstLease) {
		t.Fatal("first lease was not released")
	}

	secondLease, acquired :=
		registry.TryAcquire("Gabriela")
	if !acquired {
		t.Fatal("second lease was not acquired")
	}

	if secondLease.generation == firstLease.generation {
		t.Fatal(
			"new lease reused the previous generation",
		)
	}

	if registry.Release("Gabriela", firstLease) {
		t.Fatal(
			"stale lease released the new owner",
		)
	}

	if !registry.Owns("Gabriela", secondLease) {
		t.Fatal(
			"new owner was lost after stale release attempt",
		)
	}

	if !registry.Release("Gabriela", secondLease) {
		t.Fatal("new owner could not release its lease")
	}
}

func TestActiveCharacterSessionRegistryLeaseBlocksDuringCleanup(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	lease, acquired :=
		registry.TryAcquire("Gabriela")
	if !acquired {
		t.Fatal("lease was not acquired")
	}

	cleanupStarted := make(chan struct{})
	allowCleanupFinish := make(chan struct{})
	cleanupFinished := make(chan struct{})

	go func() {
		close(cleanupStarted)
		<-allowCleanupFinish

		registry.Release("Gabriela", lease)
		close(cleanupFinished)
	}()

	<-cleanupStarted

	if _, acquiredDuringCleanup :=
		registry.TryAcquire("Gabriela"); acquiredDuringCleanup {
		t.Fatal(
			"new session acquired while cleanup lease remained held",
		)
	}

	close(allowCleanupFinish)
	<-cleanupFinished

	if _, acquiredAfterCleanup :=
		registry.TryAcquire("Gabriela"); !acquiredAfterCleanup {
		t.Fatal(
			"new session was not acquired after cleanup released lease",
		)
	}
}

func TestActiveCharacterSessionRegistryConcurrentSingleWinner(
	t *testing.T,
) {
	registry := newActiveCharacterSessionRegistry()

	const contenderCount = 128

	start := make(chan struct{})
	results := make(
		chan activeCharacterSessionLease,
		contenderCount,
	)

	var waitGroup sync.WaitGroup
	waitGroup.Add(contenderCount)

	var winners int32

	for contender := 0; contender < contenderCount; contender++ {
		go func() {
			defer waitGroup.Done()

			<-start

			lease, acquired :=
				registry.TryAcquire("Gabriela")
			if !acquired {
				return
			}

			atomic.AddInt32(&winners, 1)
			results <- lease
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	if got := atomic.LoadInt32(&winners); got != 1 {
		t.Fatalf(
			"concurrent winners = %d, want 1",
			got,
		)
	}

	var winningLease activeCharacterSessionLease

	for lease := range results {
		winningLease = lease
	}

	if winningLease.generation == 0 {
		t.Fatal("winning lease has zero generation")
	}

	if !registry.Owns("Gabriela", winningLease) {
		t.Fatal("winning lease is not authoritative")
	}

	if !registry.Release("Gabriela", winningLease) {
		t.Fatal("winning lease could not be released")
	}
}
