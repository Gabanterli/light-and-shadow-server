package main

import (
	"testing"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

type typedNilTestStaticOccupancy struct{}

func (*typedNilTestStaticOccupancy) CanOccupy(
	_ worldmap.WorldPosition,
) (bool, error) {
	return true, nil
}

type typedNilTestDynamicOccupancy struct{}

func (*typedNilTestDynamicOccupancy) IsSpawnTileOccupied(
	_ string,
	_ worldmap.WorldSpaceID,
	_, _ float64,
	_ int,
) bool {
	return false
}

func TestAuthoritativeSpawnResolverTypedNilStaticUsesDebugPassthrough(
	t *testing.T,
) {
	var staticOccupancy *worldmap.StaticCollisionIndex

	resolver, err := newAuthoritativeSpawnPlacementResolver(
		staticOccupancy,
		nil,
		authoritativeSpawnPosition{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            100,
			Y:            100,
			Z:            0,
		},
		maxAuthoritativeSpawnFallbackSearchRadius,
	)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}

	placement, err := resolver.Resolve(
		authoritativeSpawnPlacementRequest{
			PlayerID:     "Gabriela",
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			PersistedX:   220,
			PersistedY:   148,
			PersistedZ:   0,
		},
	)
	if err != nil {
		t.Fatalf("resolve debug spawn: %v", err)
	}

	if placement.Source !=
		authoritativeSpawnPlacementSourcePersisted {
		t.Fatalf(
			"source = %q, want %q",
			placement.Source,
			authoritativeSpawnPlacementSourcePersisted,
		)
	}

	if placement.RelocationReason !=
		authoritativeSpawnReasonDebugPassthrough {
		t.Fatalf(
			"reason = %q, want %q",
			placement.RelocationReason,
			authoritativeSpawnReasonDebugPassthrough,
		)
	}

	if placement.Relocated {
		t.Fatal("debug passthrough unexpectedly relocated")
	}

	if placement.Position.X != 220 ||
		placement.Position.Y != 148 ||
		placement.Position.Z != 0 {
		t.Fatalf(
			"position = (%v,%v,%d), want (220,148,0)",
			placement.Position.X,
			placement.Position.Y,
			placement.Position.Z,
		)
	}
}

func TestAuthoritativeSpawnResolverTypedNilDynamicRejectedInProduction(
	t *testing.T,
) {
	staticOccupancy := &typedNilTestStaticOccupancy{}

	var dynamicOccupancy *typedNilTestDynamicOccupancy

	_, err := newAuthoritativeSpawnPlacementResolver(
		staticOccupancy,
		dynamicOccupancy,
		authoritativeSpawnPosition{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			X:            100,
			Y:            100,
			Z:            0,
		},
		maxAuthoritativeSpawnFallbackSearchRadius,
	)
	if err == nil {
		t.Fatal(
			"production resolver accepted typed-nil dynamic occupancy",
		)
	}
}
