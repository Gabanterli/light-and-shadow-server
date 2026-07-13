package movement

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type staticStepValidatorStub struct {
	err   error
	calls []StaticStep
}

func (s *staticStepValidatorStub) ValidateStaticStep(
	step StaticStep,
) error {
	s.calls = append(s.calls, step)
	return s.err
}

func newStaticStepMovementSystemForTest(
	validator StaticStepValidator,
) (*MovementSystem, *SpatialIndex) {
	spatialIndex := NewSpatialIndex()
	chunkManager := NewChunkManager()
	aoiManager := NewAOIManager(spatialIndex)

	return NewMovementSystemWithStaticStepValidator(
		spatialIndex,
		chunkManager,
		aoiManager,
		validator,
	), spatialIndex
}

func TestMovementSystemStaticValidatorReplacesLegacyChunkCollision(
	t *testing.T,
) {
	validator := &staticStepValidatorStub{}

	system, spatialIndex :=
		newStaticStepMovementSystemForTest(validator)

	system.InitPlayerStateInWorldSpace(
		"player",
		"main_continent",
		16383,
		100,
		0,
	)

	success, confirmedX, confirmedY, confirmedZ :=
		system.ValidateAndMove(
			"player",
			16384,
			100,
			0,
			1,
		)

	if !success {
		t.Fatal("movement rejected, want static validator to allow it")
	}

	if confirmedX != 16384 ||
		confirmedY != 100 ||
		confirmedZ != 0 {
		t.Fatalf(
			"confirmed position = (%v,%v,%d), want (16384,100,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}

	if len(validator.calls) != 1 {
		t.Fatalf(
			"validator call count = %d, want 1",
			len(validator.calls),
		)
	}

	step := validator.calls[0]

	if step.WorldSpaceID != "main_continent" {
		t.Fatalf(
			"world space = %q, want main_continent",
			step.WorldSpaceID,
		)
	}

	if step.FromX != 16383 ||
		step.FromY != 100 ||
		step.FromZ != 0 ||
		step.ToX != 16384 ||
		step.ToY != 100 ||
		step.ToZ != 0 {
		t.Fatalf(
			"unexpected static step: %+v",
			step,
		)
	}

	entity, exists := spatialIndex.GetEntity("player")
	if !exists {
		t.Fatal("player missing from spatial index")
	}

	if entity.X != 16384 ||
		entity.Y != 100 ||
		entity.Z != 0 {
		t.Fatalf(
			"spatial position = (%v,%v,%d), want (16384,100,0)",
			entity.X,
			entity.Y,
			entity.Z,
		)
	}
}

func TestMovementSystemStaticRejectionPreservesStateAndCooldown(
	t *testing.T,
) {
	blockedError := errors.New("static terrain blocked")

	validator := &staticStepValidatorStub{
		err: blockedError,
	}

	system, spatialIndex :=
		newStaticStepMovementSystemForTest(validator)

	system.InitPlayerStateInWorldSpace(
		"player",
		"main_continent",
		100,
		100,
		0,
	)

	success, confirmedX, confirmedY, confirmedZ :=
		system.ValidateAndMove(
			"player",
			101,
			100,
			0,
			1,
		)

	if success {
		t.Fatal("blocked static movement succeeded")
	}

	if confirmedX != 100 ||
		confirmedY != 100 ||
		confirmedZ != 0 {
		t.Fatalf(
			"rejected confirmation = (%v,%v,%d), want (100,100,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}

	entity, exists := spatialIndex.GetEntity("player")
	if !exists {
		t.Fatal("player missing from spatial index")
	}

	if entity.X != 100 ||
		entity.Y != 100 ||
		entity.Z != 0 {
		t.Fatalf(
			"spatial position changed after rejection: (%v,%v,%d)",
			entity.X,
			entity.Y,
			entity.Z,
		)
	}

	validator.err = nil

	success, confirmedX, confirmedY, confirmedZ =
		system.ValidateAndMove(
			"player",
			101,
			100,
			0,
			2,
		)

	if !success {
		t.Fatal(
			"immediate allowed retry failed; rejected movement consumed cooldown",
		)
	}

	if confirmedX != 101 ||
		confirmedY != 100 ||
		confirmedZ != 0 {
		t.Fatalf(
			"retry confirmation = (%v,%v,%d), want (101,100,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}
}

func TestMovementSystemWithoutStaticValidatorPreservesLegacyCollision(
	t *testing.T,
) {
	spatialIndex := NewSpatialIndex()
	chunkManager := NewChunkManager()
	aoiManager := NewAOIManager(spatialIndex)

	system := NewMovementSystem(
		spatialIndex,
		chunkManager,
		aoiManager,
	)

	system.InitPlayerState(
		"player",
		16383,
		100,
		0,
	)

	success, confirmedX, confirmedY, confirmedZ :=
		system.ValidateAndMove(
			"player",
			16384,
			100,
			0,
			1,
		)

	if success {
		t.Fatal(
			"legacy movement accepted a tile outside ChunkManager bounds",
		)
	}

	if confirmedX != 16383 ||
		confirmedY != 100 ||
		confirmedZ != 0 {
		t.Fatalf(
			"legacy rejection = (%v,%v,%d), want (16383,100,0)",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}
}

func TestMovementSystemRejectsUnknownPlayerState(
	t *testing.T,
) {
	validator := &staticStepValidatorStub{}

	system, spatialIndex :=
		newStaticStepMovementSystemForTest(validator)

	success, confirmedX, confirmedY, confirmedZ :=
		system.ValidateAndMove(
			"unknown-player",
			5000,
			5000,
			0,
			1,
		)

	if success {
		t.Fatal("unknown player state was accepted")
	}

	if confirmedX != 0 ||
		confirmedY != 0 ||
		confirmedZ != 0 {
		t.Fatalf(
			"unknown state confirmation = (%v,%v,%d), want zero position",
			confirmedX,
			confirmedY,
			confirmedZ,
		)
	}

	if len(validator.calls) != 0 {
		t.Fatalf(
			"validator called %d times for unknown player",
			len(validator.calls),
		)
	}

	if _, exists := spatialIndex.GetEntity(
		"unknown-player",
	); exists {
		t.Fatal(
			"unknown player was registered from untrusted movement input",
		)
	}
}

func TestMovementSystemDefaultInitializationUsesMainContinent(
	t *testing.T,
) {
	validator := &staticStepValidatorStub{}

	system, _ :=
		newStaticStepMovementSystemForTest(validator)

	system.InitPlayerState(
		"player",
		100,
		100,
		0,
	)

	success, _, _, _ := system.ValidateAndMove(
		"player",
		101,
		100,
		0,
		1,
	)

	if !success {
		t.Fatal("default initialized movement was rejected")
	}

	if len(validator.calls) != 1 {
		t.Fatalf(
			"validator call count = %d, want 1",
			len(validator.calls),
		)
	}

	if validator.calls[0].WorldSpaceID !=
		"main_continent" {
		t.Fatalf(
			"default world space = %q, want main_continent",
			validator.calls[0].WorldSpaceID,
		)
	}
}

func TestMovementSystemSerializesConcurrentMoves(
	t *testing.T,
) {
	validator := &staticStepValidatorStub{}

	system, spatialIndex :=
		newStaticStepMovementSystemForTest(validator)

	system.InitPlayerStateInWorldSpace(
		"player",
		"main_continent",
		100,
		100,
		0,
	)

	const requestCount = 64

	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	var successfulMoves int32

	for requestIndex := 0; requestIndex < requestCount; requestIndex++ {
		waitGroup.Add(1)

		go func(sequence uint32) {
			defer waitGroup.Done()
			<-start

			success, _, _, _ :=
				system.ValidateAndMove(
					"player",
					101,
					100,
					0,
					sequence,
				)

			if success {
				atomic.AddInt32(
					&successfulMoves,
					1,
				)
			}
		}(uint32(requestIndex + 1))
	}

	close(start)
	waitGroup.Wait()

	if successfulMoves != 1 {
		t.Fatalf(
			"successful concurrent moves = %d, want 1",
			successfulMoves,
		)
	}

	if len(validator.calls) != 1 {
		t.Fatalf(
			"static validator calls = %d, want 1",
			len(validator.calls),
		)
	}

	entity, exists := spatialIndex.GetEntity("player")
	if !exists {
		t.Fatal("player missing from spatial index")
	}

	if entity.X != 101 ||
		entity.Y != 100 ||
		entity.Z != 0 {
		t.Fatalf(
			"final spatial position = (%v,%v,%d), want (101,100,0)",
			entity.X,
			entity.Y,
			entity.Z,
		)
	}
}

type parallelStaticStepValidator struct {
	entered chan float64
	release chan struct{}
}

func (v *parallelStaticStepValidator) ValidateStaticStep(
	step StaticStep,
) error {
	v.entered <- step.FromX
	<-v.release

	return nil
}

func TestMovementSystemAllowsParallelDifferentPlayers(
	t *testing.T,
) {
	validator := &parallelStaticStepValidator{
		entered: make(chan float64, 2),
		release: make(chan struct{}),
	}

	system, _ :=
		newStaticStepMovementSystemForTest(validator)

	system.InitPlayerStateInWorldSpace(
		"player-a",
		"main_continent",
		100,
		100,
		0,
	)

	system.InitPlayerStateInWorldSpace(
		"player-b",
		"main_continent",
		110,
		100,
		0,
	)

	var waitGroup sync.WaitGroup
	var successfulMoves int32

	move := func(
		playerID string,
		targetX float64,
		sequence uint32,
	) {
		defer waitGroup.Done()

		success, _, _, _ :=
			system.ValidateAndMove(
				playerID,
				targetX,
				100,
				0,
				sequence,
			)

		if success {
			atomic.AddInt32(
				&successfulMoves,
				1,
			)
		}
	}

	waitGroup.Add(2)

	go move("player-a", 101, 1)
	go move("player-b", 111, 2)

	enteredOrigins := make(map[float64]bool)

	for entryIndex := 0; entryIndex < 2; entryIndex++ {
		select {
		case origin := <-validator.entered:
			enteredOrigins[origin] = true

		case <-time.After(2 * time.Second):
			close(validator.release)
			waitGroup.Wait()

			t.Fatal(
				"different players did not enter static validation concurrently",
			)
		}
	}

	close(validator.release)
	waitGroup.Wait()

	if successfulMoves != 2 {
		t.Fatalf(
			"successful parallel moves = %d, want 2",
			successfulMoves,
		)
	}

	if !enteredOrigins[100] ||
		!enteredOrigins[110] {
		t.Fatalf(
			"entered origins = %#v, want 100 and 110",
			enteredOrigins,
		)
	}

	system.playerMovementLocksMu.Lock()
	remainingLockEntries :=
		len(system.playerMovementLocks)
	system.playerMovementLocksMu.Unlock()

	if remainingLockEntries != 0 {
		t.Fatalf(
			"remaining player lock entries = %d, want 0",
			remainingLockEntries,
		)
	}
}
