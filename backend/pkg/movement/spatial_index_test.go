package movement

import (
	"testing"
)

func TestSpatialIndex_RegisterAndGetEntity(t *testing.T) {
	si := NewSpatialIndex()
	entity := &Entity{ID: "player1", X: 10, Y: 10, Z: 0, Type: "player"}
	si.RegisterEntity(entity)

	retrieved, exists := si.GetEntity("player1")
	if !exists {
		t.Fatal("Expected entity to exist, but it didn't")
	}
	if retrieved.ID != "player1" {
		t.Errorf("Expected entity ID to be 'player1', got '%s'", retrieved.ID)
	}
}

func TestSpatialIndex_IdempotentRegistration(t *testing.T) {
	si := NewSpatialIndex()
	entity1 := &Entity{ID: "player1", X: 10, Y: 10, Z: 0, Type: "player", BlocksMovement: true}
	si.RegisterEntity(entity1)

	if !si.IsTileOccupied("other", 10, 10, 0) {
		t.Fatal("Expected tile (10,10) to be occupied after first registration")
	}

	entity2 := &Entity{ID: "player1", X: 20, Y: 20, Z: 0, Type: "player", BlocksMovement: true}
	si.RegisterEntity(entity2)

	if si.IsTileOccupied("other", 10, 10, 0) {
		t.Error("Expected tile (10,10) to be free after re-registration")
	}
	if !si.IsTileOccupied("other", 20, 20, 0) {
		t.Error("Expected tile (20,20) to be occupied after re-registration")
	}
}

func TestSpatialIndex_RemoveEntity(t *testing.T) {
	si := NewSpatialIndex()
	entity := &Entity{ID: "player1", X: 10, Y: 10, Z: 0, Type: "player", BlocksMovement: true}
	si.RegisterEntity(entity)

	si.RemoveEntity("player1")

	_, exists := si.GetEntity("player1")
	if exists {
		t.Fatal("Expected entity to be removed, but it still exists")
	}
	if si.IsTileOccupied("other", 10, 10, 0) {
		t.Error("Expected tile (10,10) to be free after entity removal")
	}
}

func TestSpatialIndex_IsTileOccupied(t *testing.T) {
	si := NewSpatialIndex()
	blocker := &Entity{ID: "npc1", X: 15.5, Y: 15.5, Z: 7, Type: "npc", BlocksMovement: true}
	nonBlocker := &Entity{ID: "effect1", X: 16.5, Y: 16.5, Z: 7, Type: "vfx", BlocksMovement: false}
	otherFloorBlocker := &Entity{ID: "npc2", X: 15.5, Y: 15.5, Z: 8, Type: "npc", BlocksMovement: true}

	si.RegisterEntity(blocker)
	si.RegisterEntity(nonBlocker)
	si.RegisterEntity(otherFloorBlocker)

	t.Run("tile is occupied by a blocker", func(t *testing.T) {
		if !si.IsTileOccupied("player1", 15, 15, 7) {
			t.Error("Expected tile (15,15,7) to be occupied, but it was not")
		}
	})

	t.Run("tile is not occupied by a non-blocker", func(t *testing.T) {
		if si.IsTileOccupied("player1", 16, 16, 7) {
			t.Error("Expected tile (16,16,7) to be free, but it was occupied")
		}
	})

	t.Run("self-check is ignored", func(t *testing.T) {
		if si.IsTileOccupied("npc1", 15, 15, 7) {
			t.Error("Expected IsTileOccupied to ignore self, but it did not")
		}
	})

	t.Run("floor isolation", func(t *testing.T) {
		if si.IsTileOccupied("player1", 15, 15, 6) {
			t.Error("Expected tile on empty floor 6 to be free")
		}
		if !si.IsTileOccupied("player1", 15, 15, 7) {
			t.Error("Expected blocker on floor 7")
		}
		if !si.IsTileOccupied("player1", 15, 15, 8) {
			t.Error("Expected blocker on floor 8")
		}
	})

	t.Run("empty tile is not occupied", func(t *testing.T) {
		if si.IsTileOccupied("player1", 100, 100, 7) {
			t.Error("Expected empty tile (100,100,7) to be free, but it was occupied")
		}
	})
}

func TestSpatialIndex_UpdateEntityPosition(t *testing.T) {
	si := NewSpatialIndex()
	entity := &Entity{ID: "player1", X: 10, Y: 10, Z: 0, Type: "player", BlocksMovement: true}
	si.RegisterEntity(entity)

	t.Run("move within same chunk", func(t *testing.T) {
		si.UpdateEntityPosition("player1", 11, 11, 0)
		if si.IsTileOccupied("other", 10, 10, 0) {
			t.Error("Expected old tile to be free after move")
		}
		if !si.IsTileOccupied("other", 11, 11, 0) {
			t.Error("Expected new tile to be occupied after move")
		}
	})

	t.Run("move to different chunk", func(t *testing.T) {
		si.UpdateEntityPosition("player1", 40, 40, 0)
		if si.IsTileOccupied("other", 11, 11, 0) {
			t.Error("Expected old tile in old chunk to be free after move")
		}
		if !si.IsTileOccupied("other", 40, 40, 0) {
			t.Error("Expected new tile in new chunk to be occupied after move")
		}
	})

	t.Run("move to different floor", func(t *testing.T) {
		si.UpdateEntityPosition("player1", 40, 40, 1)
		if si.IsTileOccupied("other", 40, 40, 0) {
			t.Error("Expected old tile on old floor to be free after floor change")
		}
		if !si.IsTileOccupied("other", 40, 40, 1) {
			t.Error("Expected new tile on new floor to be occupied after floor change")
		}
	})
}

// B3-E: Add edge-case validation tests for spatial index.
func TestSpatialIndex_BlockingLifecycleAndEdgeCases(t *testing.T) {
	t.Run("floor isolation for blockers", func(t *testing.T) {
		si := NewSpatialIndex()

		si.RegisterEntity(&Entity{
			ID:             "floor1-blocker",
			X:              50,
			Y:              50,
			Z:              1,
			Type:           "creature",
			BlocksMovement: true,
		})

		si.RegisterEntity(&Entity{
			ID:             "floor2-blocker",
			X:              50,
			Y:              50,
			Z:              2,
			Type:           "creature",
			BlocksMovement: true,
		})

		if !si.IsTileOccupied("player", 50, 50, 1) {
			t.Error("expected blocker on floor 1")
		}

		if !si.IsTileOccupied("player", 50, 50, 2) {
			t.Error("expected blocker on floor 2")
		}

		if si.IsTileOccupied("floor1-blocker", 50, 50, 1) {
			t.Error("entity must not collide with itself")
		}
	})

	t.Run("move blocker to occupied tile fails", func(t *testing.T) {
		si := NewSpatialIndex()

		si.RegisterEntity(&Entity{
			ID:             "moving-blocker",
			X:              10,
			Y:              10,
			Z:              1,
			Type:           "creature",
			BlocksMovement: true,
		})

		si.RegisterEntity(&Entity{
			ID:             "destination-blocker",
			X:              11,
			Y:              10,
			Z:              1,
			Type:           "npc",
			BlocksMovement: true,
		})

		moved, _ := si.UpdateEntityPosition("moving-blocker", 11, 10, 1)
		if moved {
			t.Fatal("expected movement to occupied tile to fail")
		}

		if !si.IsTileOccupied("player", 10, 10, 1) {
			t.Error("origin must remain occupied after failed movement")
		}

		entity, exists := si.GetEntity("moving-blocker")
		if !exists {
			t.Fatal("moving blocker disappeared after failed movement")
		}

		if entity.X != 10 || entity.Y != 10 || entity.Z != 1 {
			t.Errorf(
				"moving blocker position changed after failed movement: got (%v,%v,%v)",
				entity.X,
				entity.Y,
				entity.Z,
			)
		}
	})

	t.Run("remove blocker frees tile", func(t *testing.T) {
		si := NewSpatialIndex()

		si.RegisterEntity(&Entity{
			ID:             "blocker1",
			X:              50,
			Y:              50,
			Z:              1,
			Type:           "creature",
			BlocksMovement: true,
		})

		si.RegisterEntity(&Entity{
			ID:             "blocker2",
			X:              50,
			Y:              50,
			Z:              2,
			Type:           "creature",
			BlocksMovement: true,
		})

		si.RemoveEntity("blocker1")

		if si.IsTileOccupied("player", 50, 50, 1) {
			t.Error("expected floor 1 tile to be free after removing blocker1")
		}

		if !si.IsTileOccupied("player", 50, 50, 2) {
			t.Error("blocker on floor 2 was incorrectly removed")
		}
	})

	t.Run("chunk boundary collision", func(t *testing.T) {
		si := NewSpatialIndex()

		si.RegisterEntity(&Entity{
			ID:             "boundary-blocker",
			X:              32,
			Y:              31,
			Z:              1,
			Type:           "npc",
			BlocksMovement: true,
		})

		if !si.IsTileOccupied("player", 32, 31, 1) {
			t.Error("expected blocker across chunk boundary to be detected")
		}

		moved, _ := si.UpdateEntityPosition("boundary-blocker", 31, 31, 1)
		if !moved {
			t.Fatal("expected movement across chunk boundary to succeed")
		}

		if si.IsTileOccupied("player", 32, 31, 1) {
			t.Error("old boundary tile remained occupied after movement")
		}

		if !si.IsTileOccupied("player", 31, 31, 1) {
			t.Error("new boundary tile was not occupied after movement")
		}
	})

	t.Run("creature lifecycle blocking", func(t *testing.T) {
		si := NewSpatialIndex()

		si.RegisterEntity(&Entity{
			ID:             "creature:orc:1",
			X:              70,
			Y:              70,
			Z:              1,
			Type:           "creature",
			BlocksMovement: true,
		})

		if !si.IsTileOccupied("player", 70, 70, 1) {
			t.Fatal("spawned creature must block its tile")
		}

		si.RemoveEntity("creature:orc:1")

		if si.IsTileOccupied("player", 70, 70, 1) {
			t.Fatal("dead creature must release its tile")
		}

		si.RegisterEntity(&Entity{
			ID:             "creature:orc:2",
			X:              70,
			Y:              70,
			Z:              1,
			Type:           "creature",
			BlocksMovement: true,
		})

		if !si.IsTileOccupied("player", 70, 70, 1) {
			t.Fatal("respawned creature must block its tile")
		}

		si.RemoveEntity("creature:orc:1")

		if !si.IsTileOccupied("player", 70, 70, 1) {
			t.Fatal("stale runtime entity removal removed the current creature")
		}
	})
}
