package main

import (
	"errors"
	"math"
	"testing"

	"github.com/light-and-shadow/backend/pkg/worldmap"
)

type authoritativeSpawnStaticResult struct {
	canOccupy bool
	err       error
}

type authoritativeSpawnStaticStub struct {
	results    map[worldmap.WorldPosition]authoritativeSpawnStaticResult
	defaultCan bool
	calls      []worldmap.WorldPosition
}

func (s *authoritativeSpawnStaticStub) CanOccupy(
	position worldmap.WorldPosition,
) (bool, error) {
	s.calls = append(s.calls, position)

	if result, exists := s.results[position]; exists {
		return result.canOccupy, result.err
	}

	return s.defaultCan, nil
}

type authoritativeSpawnDynamicCall struct {
	excludeEntityID string
	worldSpaceID    worldmap.WorldSpaceID
	x               float64
	y               float64
	z               int
}

type authoritativeSpawnDynamicStub struct {
	occupied map[worldmap.WorldPosition]bool
	calls    []authoritativeSpawnDynamicCall
}

func (s *authoritativeSpawnDynamicStub) IsSpawnTileOccupied(
	excludeEntityID string,
	worldSpaceID worldmap.WorldSpaceID,
	x, y float64,
	z int,
) bool {
	s.calls = append(
		s.calls,
		authoritativeSpawnDynamicCall{
			excludeEntityID: excludeEntityID,
			worldSpaceID:    worldSpaceID,
			x:               x,
			y:               y,
			z:               z,
		},
	)

	return s.occupied[worldmap.WorldPosition{
		WorldSpaceID: worldSpaceID,
		X:            int(x),
		Y:            int(y),
		Z:            z,
	}]
}

func authoritativeSpawnTestPosition(
	x, y, z int,
) worldmap.WorldPosition {
	return worldmap.WorldPosition{
		WorldSpaceID: worldmap.WorldSpaceMainContinent,
		X:            x,
		Y:            y,
		Z:            z,
	}
}

func authoritativeSpawnTestFallback() authoritativeSpawnPosition {
	return authoritativeSpawnPosition{
		WorldSpaceID: worldmap.WorldSpaceMainContinent,
		X:            100,
		Y:            100,
		Z:            0,
	}
}

func authoritativeSpawnTestRequest(
	x, y, z float64,
) authoritativeSpawnPlacementRequest {
	return authoritativeSpawnPlacementRequest{
		PlayerID:     "player",
		WorldSpaceID: worldmap.WorldSpaceMainContinent,
		PersistedX:   x,
		PersistedY:   y,
		PersistedZ:   z,
	}
}

func newProductionAuthoritativeSpawnResolverForTest(
	static *authoritativeSpawnStaticStub,
	dynamic *authoritativeSpawnDynamicStub,
	maxRadius int,
) *authoritativeSpawnPlacementResolver {
	resolver, err := newAuthoritativeSpawnPlacementResolver(
		static,
		dynamic,
		authoritativeSpawnTestFallback(),
		maxRadius,
	)
	if err != nil {
		panic(err)
	}

	return resolver
}

func TestAuthoritativeSpawnPlacementResolverDebugPreservesPersistedPosition(
	t *testing.T,
) {
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver, err := newAuthoritativeSpawnPlacementResolver(
		nil,
		dynamic,
		authoritativeSpawnTestFallback(),
		2,
	)
	if err != nil {
		t.Fatalf("create debug resolver: %v", err)
	}

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			12.5,
			33.25,
			2,
		),
	)
	if err != nil {
		t.Fatalf("resolve debug placement: %v", err)
	}

	if placement.Position.X != 12.5 ||
		placement.Position.Y != 33.25 ||
		placement.Position.Z != 2 {
		t.Fatalf(
			"debug position = %+v, want (12.5,33.25,2)",
			placement.Position,
		)
	}

	if placement.Relocated {
		t.Fatal("debug placement unexpectedly relocated")
	}

	if placement.RelocationReason !=
		authoritativeSpawnReasonDebugPassthrough {
		t.Fatalf(
			"debug reason = %q, want %q",
			placement.RelocationReason,
			authoritativeSpawnReasonDebugPassthrough,
		)
	}

	if len(dynamic.calls) != 0 {
		t.Fatalf(
			"debug dynamic calls = %d, want 0",
			len(dynamic.calls),
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverProductionAcceptsPersistedPosition(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results:    make(map[worldmap.WorldPosition]authoritativeSpawnStaticResult),
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			2,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			101,
			102,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve persisted placement: %v", err)
	}

	if placement.Position.X != 101 ||
		placement.Position.Y != 102 ||
		placement.Position.Z != 0 {
		t.Fatalf(
			"persisted position = %+v, want (101,102,0)",
			placement.Position,
		)
	}

	if placement.Relocated {
		t.Fatal("valid persisted placement unexpectedly relocated")
	}

	if placement.Source !=
		authoritativeSpawnPlacementSourcePersisted {
		t.Fatalf(
			"source = %q, want persisted",
			placement.Source,
		)
	}

	if len(dynamic.calls) != 1 ||
		dynamic.calls[0].excludeEntityID != "player" ||
		dynamic.calls[0].worldSpaceID !=
			worldmap.WorldSpaceMainContinent {
		t.Fatalf(
			"dynamic calls = %+v",
			dynamic.calls,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverFallsBackFromBlockedPersistedPosition(
	t *testing.T,
) {
	persisted := authoritativeSpawnTestPosition(
		101,
		100,
		0,
	)

	static := &authoritativeSpawnStaticStub{
		results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
			persisted: {
				canOccupy: false,
			},
		},
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			2,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			101,
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve blocked persisted placement: %v", err)
	}

	if placement.Position.X != 100 ||
		placement.Position.Y != 100 {
		t.Fatalf(
			"fallback position = %+v, want (100,100,0)",
			placement.Position,
		)
	}

	if !placement.Relocated {
		t.Fatal("blocked persisted position was not relocated")
	}

	if placement.RelocationReason !=
		authoritativeSpawnReasonStaticBlocked {
		t.Fatalf(
			"reason = %q, want %q",
			placement.RelocationReason,
			authoritativeSpawnReasonStaticBlocked,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverFallsBackFromStaticError(
	t *testing.T,
) {
	staticFailure :=
		&worldmap.StaticCollisionChunkNotLoadedError{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			Coordinate: worldmap.ChunkCoordinate{
				ChunkX: 0,
				ChunkY: 0,
				Z:      0,
			},
		}

	static := &authoritativeSpawnStaticStub{
		results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
			authoritativeSpawnTestPosition(
				10,
				10,
				0,
			): {
				err: staticFailure,
			},
		},
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			2,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			10,
			10,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve static error placement: %v", err)
	}

	if !placement.Relocated ||
		placement.RelocationReason !=
			authoritativeSpawnReasonStaticUnavailable {
		t.Fatalf(
			"placement = %+v",
			placement,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverFallsBackFromDynamicOccupancy(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results:    make(map[worldmap.WorldPosition]authoritativeSpawnStaticResult),
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: map[worldmap.WorldPosition]bool{
			authoritativeSpawnTestPosition(
				101,
				100,
				0,
			): true,
		},
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			2,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			101,
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve dynamic occupancy placement: %v", err)
	}

	if !placement.Relocated ||
		placement.RelocationReason !=
			authoritativeSpawnReasonDynamicOccupied {
		t.Fatalf(
			"placement = %+v",
			placement,
		)
	}

	if placement.Position.X != 100 ||
		placement.Position.Y != 100 {
		t.Fatalf(
			"fallback position = %+v, want center fallback",
			placement.Position,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverUsesDeterministicFallbackOrder(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
			authoritativeSpawnTestPosition(
				100,
				99,
				0,
			): {
				canOccupy: true,
			},
		},
		defaultCan: false,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			1,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			0,
			0,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve deterministic fallback: %v", err)
	}

	if placement.Position.X != 100 ||
		placement.Position.Y != 99 {
		t.Fatalf(
			"deterministic fallback = %+v, want north tile (100,99,0)",
			placement.Position,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverSkipsUnavailableFallbackCandidates(
	t *testing.T,
) {
	unavailable :=
		&worldmap.StaticCollisionChunkNotLoadedError{
			WorldSpaceID: worldmap.WorldSpaceMainContinent,
			Coordinate: worldmap.ChunkCoordinate{
				ChunkX: 3,
				ChunkY: 3,
				Z:      0,
			},
		}

	static := &authoritativeSpawnStaticStub{
		results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
			authoritativeSpawnTestPosition(
				100,
				100,
				0,
			): {
				err: unavailable,
			},
			authoritativeSpawnTestPosition(
				99,
				100,
				0,
			): {
				canOccupy: true,
			},
		},
		defaultCan: false,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			1,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			0,
			0,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve fallback after unavailable candidate: %v", err)
	}

	if placement.Position.X != 99 ||
		placement.Position.Y != 100 {
		t.Fatalf(
			"fallback position = %+v, want west tile (99,100,0)",
			placement.Position,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverRejectsWhenNoFallbackTileExists(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results:    make(map[worldmap.WorldPosition]authoritativeSpawnStaticResult),
		defaultCan: false,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			1,
		)

	_, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			0,
			0,
			0,
		),
	)
	if err == nil {
		t.Fatal("expected unavailable placement error")
	}

	var unavailableError *authoritativeSpawnPlacementUnavailableError

	if !errors.As(err, &unavailableError) {
		t.Fatalf(
			"error type = %T, want authoritativeSpawnPlacementUnavailableError",
			err,
		)
	}

	if unavailableError.CheckedCandidates != 9 {
		t.Fatalf(
			"checked candidates = %d, want 9",
			unavailableError.CheckedCandidates,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverFallsBackFromInvalidPersistedCoordinates(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results:    make(map[worldmap.WorldPosition]authoritativeSpawnStaticResult),
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			1,
		)

	placement, err := resolver.Resolve(
		authoritativeSpawnTestRequest(
			math.NaN(),
			100,
			0,
		),
	)
	if err != nil {
		t.Fatalf("resolve invalid persisted coordinates: %v", err)
	}

	if !placement.Relocated ||
		placement.RelocationReason !=
			authoritativeSpawnReasonInvalidPersisted {
		t.Fatalf(
			"placement = %+v",
			placement,
		)
	}

	if placement.Position.X != 100 ||
		placement.Position.Y != 100 {
		t.Fatalf(
			"fallback position = %+v, want (100,100,0)",
			placement.Position,
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverValidatesConfigurationAndRequest(
	t *testing.T,
) {
	static := &authoritativeSpawnStaticStub{
		results:    make(map[worldmap.WorldPosition]authoritativeSpawnStaticResult),
		defaultCan: true,
	}
	dynamic := &authoritativeSpawnDynamicStub{
		occupied: make(map[worldmap.WorldPosition]bool),
	}

	t.Run(
		"production requires dynamic occupancy",
		func(t *testing.T) {
			_, err := newAuthoritativeSpawnPlacementResolver(
				static,
				nil,
				authoritativeSpawnTestFallback(),
				1,
			)
			if err == nil {
				t.Fatal("expected missing dynamic occupancy error")
			}
		},
	)

	t.Run(
		"rejects negative radius",
		func(t *testing.T) {
			_, err := newAuthoritativeSpawnPlacementResolver(
				static,
				dynamic,
				authoritativeSpawnTestFallback(),
				-1,
			)
			if err == nil {
				t.Fatal("expected negative radius error")
			}
		},
	)

	t.Run(
		"rejects excessive radius",
		func(t *testing.T) {
			_, err := newAuthoritativeSpawnPlacementResolver(
				static,
				dynamic,
				authoritativeSpawnTestFallback(),
				maxAuthoritativeSpawnFallbackSearchRadius+1,
			)
			if err == nil {
				t.Fatal("expected excessive radius error")
			}
		},
	)

	t.Run(
		"rejects fractional fallback",
		func(t *testing.T) {
			fallback := authoritativeSpawnTestFallback()
			fallback.X = 100.5

			_, err := newAuthoritativeSpawnPlacementResolver(
				static,
				dynamic,
				fallback,
				1,
			)
			if err == nil {
				t.Fatal("expected fractional fallback error")
			}
		},
	)

	resolver :=
		newProductionAuthoritativeSpawnResolverForTest(
			static,
			dynamic,
			1,
		)

	t.Run(
		"rejects empty player ID",
		func(t *testing.T) {
			request :=
				authoritativeSpawnTestRequest(
					100,
					100,
					0,
				)
			request.PlayerID = " "

			_, err := resolver.Resolve(request)
			if err == nil {
				t.Fatal("expected empty player ID error")
			}
		},
	)

	t.Run(
		"rejects empty world space",
		func(t *testing.T) {
			request :=
				authoritativeSpawnTestRequest(
					100,
					100,
					0,
				)
			request.WorldSpaceID = ""

			_, err := resolver.Resolve(request)
			if err == nil {
				t.Fatal("expected empty world space error")
			}
		},
	)

	t.Run(
		"debug rejects non-finite coordinate",
		func(t *testing.T) {
			debugResolver, err :=
				newAuthoritativeSpawnPlacementResolver(
					nil,
					dynamic,
					authoritativeSpawnTestFallback(),
					1,
				)
			if err != nil {
				t.Fatalf(
					"create debug resolver: %v",
					err,
				)
			}

			_, err = debugResolver.Resolve(
				authoritativeSpawnTestRequest(
					math.Inf(1),
					100,
					0,
				),
			)
			if err == nil {
				t.Fatal("expected non-finite debug coordinate error")
			}
		},
	)
}

func TestAuthoritativeSpawnPlacementResolverPropagatesFatalStaticErrors(
	t *testing.T,
) {
	t.Run(
		"persisted tile data corruption",
		func(t *testing.T) {
			position :=
				authoritativeSpawnTestPosition(
					101,
					100,
					0,
				)

			fatalError :=
				&worldmap.StaticCollisionTileDataError{
					Position: position,
					Reason:   "corrupt collision data",
				}

			static := &authoritativeSpawnStaticStub{
				results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
					position: {
						err: fatalError,
					},
				},
				defaultCan: true,
			}

			dynamic := &authoritativeSpawnDynamicStub{
				occupied: make(
					map[worldmap.WorldPosition]bool,
				),
			}

			resolver :=
				newProductionAuthoritativeSpawnResolverForTest(
					static,
					dynamic,
					2,
				)

			_, err := resolver.Resolve(
				authoritativeSpawnTestRequest(
					101,
					100,
					0,
				),
			)
			if err == nil {
				t.Fatal(
					"expected fatal persisted static error",
				)
			}

			var evaluationError *authoritativeSpawnStaticEvaluationError

			if !errors.As(err, &evaluationError) {
				t.Fatalf(
					"error type = %T, want authoritativeSpawnStaticEvaluationError",
					err,
				)
			}

			if !errors.Is(err, fatalError) {
				t.Fatalf(
					"error does not wrap fatal tile data error: %v",
					err,
				)
			}

			if len(static.calls) != 1 {
				t.Fatalf(
					"static calls = %d, want 1",
					len(static.calls),
				)
			}
		},
	)

	t.Run(
		"fallback tile data corruption",
		func(t *testing.T) {
			fallback :=
				authoritativeSpawnTestPosition(
					100,
					100,
					0,
				)

			fatalError :=
				&worldmap.StaticCollisionTileDataError{
					Position: fallback,
					Reason:   "corrupt fallback collision data",
				}

			static := &authoritativeSpawnStaticStub{
				results: map[worldmap.WorldPosition]authoritativeSpawnStaticResult{
					fallback: {
						err: fatalError,
					},
				},
				defaultCan: true,
			}

			dynamic := &authoritativeSpawnDynamicStub{
				occupied: make(
					map[worldmap.WorldPosition]bool,
				),
			}

			resolver :=
				newProductionAuthoritativeSpawnResolverForTest(
					static,
					dynamic,
					2,
				)

			_, err := resolver.Resolve(
				authoritativeSpawnTestRequest(
					math.NaN(),
					100,
					0,
				),
			)
			if err == nil {
				t.Fatal(
					"expected fatal fallback static error",
				)
			}

			var evaluationError *authoritativeSpawnStaticEvaluationError

			if !errors.As(err, &evaluationError) {
				t.Fatalf(
					"error type = %T, want authoritativeSpawnStaticEvaluationError",
					err,
				)
			}

			if !errors.Is(err, fatalError) {
				t.Fatalf(
					"error does not wrap fatal fallback error: %v",
					err,
				)
			}

			if len(static.calls) != 1 {
				t.Fatalf(
					"static calls = %d, want 1",
					len(static.calls),
				)
			}
		},
	)
}

func TestAuthoritativeSpawnPlacementResolverValidatesProductionPersistedCoordinates(
	t *testing.T,
) {
	testCases := []struct {
		name string
		x    float64
		y    float64
		z    float64
	}{
		{
			name: "fractional x",
			x:    100.5,
			y:    100,
			z:    0,
		},
		{
			name: "fractional y",
			x:    100,
			y:    100.5,
			z:    0,
		},
		{
			name: "fractional z",
			x:    100,
			y:    100,
			z:    0.5,
		},
		{
			name: "positive x overflow",
			x:    float64(1 << 31),
			y:    100,
			z:    0,
		},
		{
			name: "negative y overflow",
			x:    100,
			y:    float64((-1 << 31) - 1),
			z:    0,
		},
		{
			name: "positive z overflow",
			x:    100,
			y:    100,
			z:    float64(1 << 31),
		},
		{
			name: "infinite x",
			x:    math.Inf(1),
			y:    100,
			z:    0,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				static :=
					&authoritativeSpawnStaticStub{
						results: make(
							map[worldmap.WorldPosition]authoritativeSpawnStaticResult,
						),
						defaultCan: true,
					}

				dynamic :=
					&authoritativeSpawnDynamicStub{
						occupied: make(
							map[worldmap.WorldPosition]bool,
						),
					}

				resolver :=
					newProductionAuthoritativeSpawnResolverForTest(
						static,
						dynamic,
						1,
					)

				placement, err := resolver.Resolve(
					authoritativeSpawnTestRequest(
						testCase.x,
						testCase.y,
						testCase.z,
					),
				)
				if err != nil {
					t.Fatalf(
						"resolve invalid persisted position: %v",
						err,
					)
				}

				if !placement.Relocated {
					t.Fatal(
						"invalid production position was not relocated",
					)
				}

				if placement.RelocationReason !=
					authoritativeSpawnReasonInvalidPersisted {
					t.Fatalf(
						"reason = %q, want %q",
						placement.RelocationReason,
						authoritativeSpawnReasonInvalidPersisted,
					)
				}

				if placement.Position.X != 100 ||
					placement.Position.Y != 100 ||
					placement.Position.Z != 0 {
					t.Fatalf(
						"fallback position = %+v, want (100,100,0)",
						placement.Position,
					)
				}

				if len(static.calls) != 1 {
					t.Fatalf(
						"static calls = %d, want fallback only",
						len(static.calls),
					)
				}
			},
		)
	}
}

func TestAuthoritativeSpawnPlacementResolverTrimsFallbackWorldSpace(
	t *testing.T,
) {
	fallback := authoritativeSpawnTestFallback()
	fallback.WorldSpaceID = " main_continent "

	resolver, err :=
		newAuthoritativeSpawnPlacementResolver(
			nil,
			nil,
			fallback,
			1,
		)
	if err != nil {
		t.Fatalf(
			"create resolver with padded world space: %v",
			err,
		)
	}

	if resolver.fallback.WorldSpaceID !=
		worldmap.WorldSpaceMainContinent {
		t.Fatalf(
			"fallback world space = %q, want %q",
			resolver.fallback.WorldSpaceID,
			worldmap.WorldSpaceMainContinent,
		)
	}
}
