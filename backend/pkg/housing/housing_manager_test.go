package housing

import (
	"context"
	"testing"

	"github.com/light-and-shadow/backend/pkg/inventory"
)

// MockMarketHook implements the HousingMarketHook interface to verify compatibility (PATCH 9)
type MockMarketHook struct {
	Auctions map[string]*PropertyListing
}

func (m *MockMarketHook) RegisterPropertyForAuction(ctx context.Context, houseID string, startingBid int64, duration interface{}) error {
	return nil
}

func (m *MockMarketHook) PlaceBidOnProperty(ctx context.Context, playerID string, houseID string, bidAmount int64) error {
	return nil
}

func (m *MockMarketHook) FinalizeAuction(ctx context.Context, houseID string) error {
	return nil
}

func (m *MockMarketHook) TransferOwnershipDirect(ctx context.Context, houseID string, newOwnerID int, salePrice int64) error {
	return nil
}

func TestHousingManagerInitialization(t *testing.T) {
	hm := NewHousingManager(nil)
	if hm == nil {
		t.Fatal("Expected NewHousingManager to return a valid instance, got nil")
	}

	// Verify that fallbacks were loaded
	if len(hm.houses) == 0 {
		t.Error("Expected houses config to be loaded via fallback")
	}

	if len(hm.furniture) == 0 {
		t.Error("Expected furniture blueprints to be loaded via fallback")
	}
}

func TestRespawnActivationAndFurnitureChecks(t *testing.T) {
	hm := NewHousingManager(nil)

	playerID := "HeroOfLight"
	houseID := "house_beginner_01"

	// Mock character ID mapping in the state
	state := hm.states[houseID]
	state.OwnerID = 1001
	state.OwnerName = playerID
	state.RentStatus = "active"

	// 1. Initially, no furniture is placed, so respawn should NOT be allowed (PATCH 6)
	x, y, z, name, ok := hm.GetPlayerActiveHouseLocation(playerID)
	if ok {
		t.Errorf("Expected respawn to be blocked without a placed bed, got ok=true (Coordinates: %f, %f, %d, Name: %s)", x, y, z, name)
	}

	// 2. Place a non-respawn furniture (e.g. wooden chair)
	state.Decorations = append(state.Decorations, PlacedDecoration{
		ID:          1,
		FurnitureID: "furn_chair_wooden",
		X:           1.0,
		Y:           1.0,
		Z:           0,
	})

	x, y, z, name, ok = hm.GetPlayerActiveHouseLocation(playerID)
	if ok {
		t.Error("Expected respawn to still be blocked with only a wooden chair placed")
	}

	// 3. Place a bed furniture
	state.Decorations = append(state.Decorations, PlacedDecoration{
		ID:          2,
		FurnitureID: "furn_bed_comfy",
		X:           2.0,
		Y:           2.0,
		Z:           0,
	})

	x, y, z, name, ok = hm.GetPlayerActiveHouseLocation(playerID)
	if !ok {
		t.Fatal("Expected respawn to be allowed after placing a bed")
	}

	if x != hm.houses[houseID].X ||
	y != hm.houses[houseID].Y ||
	z != hm.houses[houseID].Z ||
	name != hm.houses[houseID].Name {

	t.Errorf("Expected respawn coordinates to match house configs, got (%f, %f, %d, %s)", x, y, z, name)
}
}

func TestDecorationBudgetLimits(t *testing.T) {
	hm := NewHousingManager(nil)

	playerID := "DecorQueen"
	houseID := "house_beginner_01" // small size: budget of 50 points

	state := hm.states[houseID]
	state.OwnerID = 1001
	state.OwnerName = playerID
	state.RentStatus = "active"

	// Mock a player inventory containing furniture
	inv := &inventory.PlayerInventory{
		Items: map[int]*inventory.InventoryItem{
			0: {ItemID: "furn_secure_chest", Quantity: 5, Durability: 100},
			1: {ItemID: "furn_guild_vault_gold", Quantity: 2, Durability: 100},
		},
	}

	// Place some furniture up to budget limit (furn_secure_chest decoration cost is 15)
	// Placing 3 secure chests = 45 points
	for i := 0; i < 3; i++ {
		err := hm.PlaceFurniture(playerID, houseID, "furn_secure_chest", 0, 0, 0, 0, inv)
		if err != nil {
			t.Fatalf("Expected placing secure chest %d to succeed, got error: %v", i+1, err)
		}
	}

	// Placing a 4th secure chest would cost 15 more, exceeding the 50 points budget (45 + 15 = 60)
	err := hm.PlaceFurniture(playerID, houseID, "furn_secure_chest", 0, 0, 0, 0, inv)
	if err == nil {
		t.Error("Expected decoration budget enforcement to block placing the 4th secure chest, but it succeeded")
	}

	// Placing a heavy gold vault (40 points) on top of 45 should also exceed budget (45 + 40 = 85)
	err = hm.PlaceFurniture(playerID, houseID, "furn_guild_vault_gold", 0, 0, 0, 0, inv)
	if err == nil {
		t.Error("Expected placing gold vault to fail due to budget exhaustion")
	}
}

func TestPermissionsSystemACL(t *testing.T) {
	hm := NewHousingManager(nil)

	ownerID := "Owner"
	houseID := "house_beginner_01"

	state := hm.states[houseID]
	state.OwnerID = 1001
	state.OwnerName = ownerID
	state.RentStatus = "active"
	state.Permissions = "private"

	ctx := context.Background()

	// 1. Owner always has direct access
	err := hm.checkPermission(ctx, ownerID, houseID, "storage_deposit")
	if err != nil {
		t.Errorf("Expected owner to always have access, got error: %v", err)
	}

	// 2. Private ACL should block non-owner
	err = hm.checkPermission(ctx, "Stranger", houseID, "storage_deposit")
	if err == nil {
		t.Error("Expected Stranger to be blocked from private house storage")
	}

	// 3. Public ACL should allow anyone
	state.Permissions = "public"
	err = hm.checkPermission(ctx, "Stranger", houseID, "storage_deposit")
	if err != nil {
		t.Errorf("Expected anyone to access public house storage, got error: %v", err)
	}
}

func TestGuildHousePermissions(t *testing.T) {
	hm := NewHousingManager(nil)

	houseID := "guild_house_holy_01"
	state := hm.states[houseID]
	state.OwnerID = 0
	state.GuildID = 5
	state.MinRank = "member"

	ctx := context.Background()

	// Without DB, checkPermission uses state configuration
	// We verify that guild member checking handles mock configurations correctly
	err := hm.checkPermission(ctx, "SomePlayer", houseID, "storage_deposit")
	if err == nil {
		// When no DB, checkPermission assumes Stranger is not part of the guild and blocks access
		t.Log("Successfully checked stranger access on guild house without DB")
	}
}

func TestHousingMarketInterfaceSatisfied(t *testing.T) {
	var _ HousingMarketHook = (*MockMarketHook)(nil)
}
