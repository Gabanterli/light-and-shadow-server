package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/persistence"
	"github.com/light-and-shadow/backend/pkg/worldmap"
)

type fakeAuthoritativeLoginSpawnResolver struct {
	placements []authoritativeSpawnPlacement
	calls      int
}

func (f *fakeAuthoritativeLoginSpawnResolver) Resolve(
	_ authoritativeSpawnPlacementRequest,
) (authoritativeSpawnPlacement, error) {
	index := f.calls
	f.calls++

	if index >= len(f.placements) {
		return authoritativeSpawnPlacement{},
			errors.New(
				"fake resolver exhausted",
			)
	}

	return f.placements[index], nil
}

type fakeAuthoritativeLoginPositionRelocator struct {
	calls   int
	request persistence.CharacterPositionRelocationRequest
	result  persistence.CharacterPositionRelocationResult
	err     error
}

func (f *fakeAuthoritativeLoginPositionRelocator) RelocateCharacterPosition(
	_ context.Context,
	request persistence.CharacterPositionRelocationRequest,
) (persistence.CharacterPositionRelocationResult, error) {
	f.calls++
	f.request = request

	if f.err != nil {
		return persistence.CharacterPositionRelocationResult{},
			f.err
	}

	return f.result, nil
}

type retryAuthoritativeLoginSpawnClaimer struct {
	index       *movement.SpatialIndex
	firstError  error
	claimCalls  int
	removeCalls int
}

func (f *retryAuthoritativeLoginSpawnClaimer) TryClaimPlayerLoginEntity(
	entity *movement.Entity,
) error {
	f.claimCalls++

	if f.claimCalls == 1 &&
		f.firstError != nil {
		return f.firstError
	}

	return f.index.TryClaimPlayerLoginEntity(
		entity,
	)
}

func (f *retryAuthoritativeLoginSpawnClaimer) RemoveEntity(
	id string,
) {
	f.removeCalls++
	f.index.RemoveEntity(id)
}

func authoritativeLoginSpawnTestPlacement(
	x, y float64,
	z int,
	relocated bool,
) authoritativeSpawnPlacement {
	source :=
		authoritativeSpawnPlacementSourcePersisted

	reason :=
		authoritativeSpawnReasonPersistedAccepted

	if relocated {
		source =
			authoritativeSpawnPlacementSourceFallback

		reason =
			authoritativeSpawnReasonDynamicOccupied
	}

	return authoritativeSpawnPlacement{
		Position: authoritativeSpawnPosition{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            x,
			Y:            y,
			Z:            z,
		},
		Source:           source,
		RelocationReason: reason,
		Relocated:        relocated,
	}
}

func TestAuthoritativeLoginSpawnAcceptsPersistedPosition(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	resolver :=
		&fakeAuthoritativeLoginSpawnResolver{
			placements: []authoritativeSpawnPlacement{
				authoritativeLoginSpawnTestPlacement(
					10,
					20,
					1,
					false,
				),
			},
		}

	relocator :=
		&fakeAuthoritativeLoginPositionRelocator{}

	activation, err :=
		resolveClaimAndPersistAuthoritativeLoginSpawn(
			context.Background(),
			resolver,
			index,
			relocator,
			authoritativeSpawnPlacementRequest{
				PlayerID:     "player",
				WorldSpaceID: worldmap.WorldSpaceMainContinent,
				PersistedX:   10,
				PersistedY:   20,
				PersistedZ:   1,
			},
			7,
		)
	if err != nil {
		t.Fatalf(
			"prepare persisted spawn: %v",
			err,
		)
	}

	if activation.Version != 7 ||
		activation.ClaimAttempts != 1 {
		t.Fatalf(
			"unexpected activation: %+v",
			activation,
		)
	}

	if relocator.calls != 0 {
		t.Fatalf(
			"relocation calls = %d, want 0",
			relocator.calls,
		)
	}

	entity, exists := index.GetEntity("player")

	if !exists ||
		entity == nil ||
		entity.X != 10 ||
		entity.Y != 20 ||
		entity.Z != 1 ||
		!entity.BlocksMovement ||
		entity.Type != "player" {
		t.Fatalf(
			"unexpected claimed player: %+v",
			entity,
		)
	}
}

func TestAuthoritativeLoginSpawnPersistsRelocation(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	resolver :=
		&fakeAuthoritativeLoginSpawnResolver{
			placements: []authoritativeSpawnPlacement{
				authoritativeLoginSpawnTestPlacement(
					100,
					101,
					0,
					true,
				),
			},
		}

	relocator :=
		&fakeAuthoritativeLoginPositionRelocator{
			result: persistence.CharacterPositionRelocationResult{
				NewVersion: 8,
			},
		}

	activation, err :=
		resolveClaimAndPersistAuthoritativeLoginSpawn(
			context.Background(),
			resolver,
			index,
			relocator,
			authoritativeSpawnPlacementRequest{
				PlayerID:     "player",
				WorldSpaceID: worldmap.WorldSpaceMainContinent,
				PersistedX:   999,
				PersistedY:   999,
				PersistedZ:   0,
			},
			7,
		)
	if err != nil {
		t.Fatalf(
			"prepare relocated spawn: %v",
			err,
		)
	}

	if activation.Version != 8 {
		t.Fatalf(
			"activation version = %d, want 8",
			activation.Version,
		)
	}

	if relocator.calls != 1 ||
		relocator.request.PlayerID != "player" ||
		relocator.request.X != 100 ||
		relocator.request.Y != 101 ||
		relocator.request.Z != 0 ||
		relocator.request.ExpectedVersion != 7 {
		t.Fatalf(
			"unexpected relocation request: %+v",
			relocator.request,
		)
	}
}

func TestAuthoritativeLoginSpawnRetriesBlockedClaim(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	resolver :=
		&fakeAuthoritativeLoginSpawnResolver{
			placements: []authoritativeSpawnPlacement{
				authoritativeLoginSpawnTestPlacement(
					100,
					100,
					0,
					false,
				),
				authoritativeLoginSpawnTestPlacement(
					101,
					100,
					0,
					true,
				),
			},
		}

	claimer :=
		&retryAuthoritativeLoginSpawnClaimer{
			index: index,
			firstError: &movement.SpatialPlayerLoginClaimBlockedError{
				EntityID:    "player",
				BlockerID:   "creature",
				BlockerType: "creature",
				X:           100,
				Y:           100,
				Z:           0,
			},
		}

	relocator :=
		&fakeAuthoritativeLoginPositionRelocator{
			result: persistence.CharacterPositionRelocationResult{
				NewVersion: 6,
			},
		}

	activation, err :=
		resolveClaimAndPersistAuthoritativeLoginSpawn(
			context.Background(),
			resolver,
			claimer,
			relocator,
			authoritativeSpawnPlacementRequest{
				PlayerID:     "player",
				WorldSpaceID: worldmap.WorldSpaceMainContinent,
				PersistedX:   100,
				PersistedY:   100,
				PersistedZ:   0,
			},
			5,
		)
	if err != nil {
		t.Fatalf(
			"retry blocked claim: %v",
			err,
		)
	}

	if activation.ClaimAttempts != 2 ||
		resolver.calls != 2 ||
		claimer.claimCalls != 2 {
		t.Fatalf(
			"unexpected retry state: activation=%+v resolver=%d claims=%d",
			activation,
			resolver.calls,
			claimer.claimCalls,
		)
	}
}

func TestAuthoritativeLoginSpawnRollsBackPersistenceFailure(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	resolver :=
		&fakeAuthoritativeLoginSpawnResolver{
			placements: []authoritativeSpawnPlacement{
				authoritativeLoginSpawnTestPlacement(
					100,
					100,
					0,
					true,
				),
			},
		}

	relocator :=
		&fakeAuthoritativeLoginPositionRelocator{
			err: errors.New(
				"database unavailable",
			),
		}

	_, err :=
		resolveClaimAndPersistAuthoritativeLoginSpawn(
			context.Background(),
			resolver,
			index,
			relocator,
			authoritativeSpawnPlacementRequest{
				PlayerID:     "player",
				WorldSpaceID: worldmap.WorldSpaceMainContinent,
				PersistedX:   999,
				PersistedY:   999,
				PersistedZ:   0,
			},
			5,
		)
	if err == nil {
		t.Fatal(
			"expected persistence failure",
		)
	}

	if _, exists := index.GetEntity("player"); exists {
		t.Fatal(
			"failed relocation left player claim registered",
		)
	}
}

func TestPlayerLoginSpawnDynamicOccupancyIgnoresPlayers(
	t *testing.T,
) {
	index := movement.NewSpatialIndex()

	if err := index.TryClaimPlayerLoginEntity(
		&movement.Entity{
			ID:             "existing-player",
			X:              100,
			Y:              100,
			Z:              0,
			Type:           "player",
			BlocksMovement: true,
		},
	); err != nil {
		t.Fatalf(
			"claim existing player: %v",
			err,
		)
	}

	occupancy :=
		newSpatialIndexPlayerLoginSpawnDynamicOccupancy(
			index,
		)

	if occupancy.IsSpawnTileOccupied(
		"new-player",
		worldmap.WorldSpaceMainContinent,
		100,
		100,
		0,
	) {
		t.Fatal(
			"existing player incorrectly blocked login",
		)
	}

	index.RegisterEntity(
		&movement.Entity{
			ID:             "npc",
			X:              100,
			Y:              100,
			Z:              0,
			Type:           "npc",
			BlocksMovement: true,
		},
	)

	if !occupancy.IsSpawnTileOccupied(
		"new-player",
		worldmap.WorldSpaceMainContinent,
		100,
		100,
		0,
	) {
		t.Fatal(
			"blocking NPC did not block login",
		)
	}
}

func TestAuthoritativeLoginSpawnMainIntegrationContract(
	t *testing.T,
) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf(
			"read main.go: %v",
			err,
		)
	}

	mainText := string(content)

	requiredFragments := []string{
		"authoritativeSpawnResolver",
		"*authoritativeSpawnPlacementResolver",
		"newSpatialIndexPlayerLoginSpawnDynamicOccupancy(",
		"prepareAuthoritativeLoginSpawn(",
		`"character_spawn_failed"`,
		"resolvedX",
		"resolvedY",
		"resolvedZ",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(
			mainText,
			fragment,
		) {
			t.Fatalf(
				"main.go missing fragment %q",
				fragment,
			)
		}
	}

	selectStart := strings.Index(
		mainText,
		"case protocol.CS_CHAR_SELECT_REQUEST:",
	)

	selectEnd := strings.Index(
		mainText,
		"case protocol.CS_INVENTORY_REQUEST:",
	)

	if selectStart < 0 ||
		selectEnd <= selectStart {
		t.Fatal(
			"character select boundaries not found",
		)
	}

	selectBlock :=
		mainText[selectStart:selectEnd]

	prepareIndex := strings.Index(
		selectBlock,
		"prepareAuthoritativeLoginSpawn(",
	)

	activationIndex := strings.Index(
		selectBlock,
		"playerID = characterID",
	)

	if prepareIndex < 0 ||
		activationIndex < 0 ||
		prepareIndex > activationIndex {
		t.Fatal(
			"player activates before spawn preparation",
		)
	}

	afterActivation :=
		selectBlock[activationIndex:]

	for _, forbidden := range []string{
		"savedX",
		"savedY",
		"savedZ",
	} {
		if strings.Contains(
			afterActivation,
			forbidden,
		) {
			t.Fatalf(
				"character select still uses %s after activation",
				forbidden,
			)
		}
	}
}
