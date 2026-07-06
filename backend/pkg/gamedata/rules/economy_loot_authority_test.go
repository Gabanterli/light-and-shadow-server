package rules

import (
	"errors"
	"testing"
)

func TestOfficialItemMutationSources(t *testing.T) {
	t.Run("returns exactly 4 sources", func(t *testing.T) {
		sources := OfficialItemMutationSources()
		if len(sources) != 4 {
			t.Errorf("expected 4 official sources, but got %d", len(sources))
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		sources1 := OfficialItemMutationSources()
		if len(sources1) == 0 {
			t.Fatal("expected non-empty slice")
		}
		originalFirstSource := sources1[0]
		sources1[0] = "mutated"

		sources2 := OfficialItemMutationSources()
		if sources2[0] == "mutated" {
			t.Error("mutation of returned slice affected the internal source")
		}
		if sources2[0] != originalFirstSource {
			t.Errorf("expected first source to be '%s', but got '%s'", originalFirstSource, sources2[0])
		}
	})
}

func TestIsOfficialItemMutationSource(t *testing.T) {
	testCases := []struct {
		name     string
		source   ItemMutationSource
		expected bool
	}{
		{"loot_drop is official", ItemMutationSourceLootDrop, true},
		{"player_trade is official", ItemMutationSourcePlayerTrade, true},
		{"quest_reward is official", ItemMutationSourceQuestReward, true},
		{"inventory_move is official", ItemMutationSourceInventoryMove, true},
		{"empty string is not official", "", false},
		{"client_create is not official", "client_create", false},
		{"admin_spawn is not official", "admin_spawn", false},
		{"debug_grant is not official", "debug_grant", false},
		{"offline_macro is not official", "offline_macro", false},
		{"water is not official", "water", false},
		{"air is not official", "air", false},
		{"unknown is not official", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialItemMutationSource(tc.source); got != tc.expected {
				t.Errorf("IsOfficialItemMutationSource('%s') = %v; want %v", tc.source, got, tc.expected)
			}
		})
	}
}

func TestItemMutationSourceCreatesItem(t *testing.T) {
	testCases := []struct {
		name     string
		source   ItemMutationSource
		expected bool
	}{
		{"loot_drop creates item", ItemMutationSourceLootDrop, true},
		{"quest_reward creates item", ItemMutationSourceQuestReward, true},
		{"player_trade does not create item", ItemMutationSourcePlayerTrade, false},
		{"inventory_move does not create item", ItemMutationSourceInventoryMove, false},
		{"empty string does not create item", "", false},
		{"client_create does not create item", "client_create", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ItemMutationSourceCreatesItem(tc.source); got != tc.expected {
				t.Errorf("ItemMutationSourceCreatesItem('%s') = %v; want %v", tc.source, got, tc.expected)
			}
		})
	}
}

func TestItemMutationSourceRequiresTransaction(t *testing.T) {
	testCases := []struct {
		name     string
		source   ItemMutationSource
		expected bool
	}{
		{"player_trade requires transaction", ItemMutationSourcePlayerTrade, true},
		{"loot_drop does not require transaction", ItemMutationSourceLootDrop, false},
		{"quest_reward does not require transaction", ItemMutationSourceQuestReward, false},
		{"inventory_move does not require transaction", ItemMutationSourceInventoryMove, false},
		{"empty string does not require transaction", "", false},
		{"client_create does not require transaction", "client_create", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ItemMutationSourceRequiresTransaction(tc.source); got != tc.expected {
				t.Errorf("ItemMutationSourceRequiresTransaction('%s') = %v; want %v", tc.source, got, tc.expected)
			}
		})
	}
}

func TestCanApplyItemMutation(t *testing.T) {
	t.Run("allows valid loot drop", func(t *testing.T) {
		req := ItemMutationAuthorityRequest{
			Source:               ItemMutationSourceLootDrop,
			BackendValidated:     true,
			ClientCreatedItem:    false,
			BackendGeneratedItem: true,
			Transactional:        false, // Not required for loot
			AuditLogged:          true,
		}
		if err := CanApplyItemMutation(req); err != nil {
			t.Errorf("expected nil for valid loot drop, got %v", err)
		}
	})

	t.Run("allows valid player trade", func(t *testing.T) {
		req := ItemMutationAuthorityRequest{
			Source:               ItemMutationSourcePlayerTrade,
			BackendValidated:     true,
			ClientCreatedItem:    false,
			BackendGeneratedItem: false, // Not required for trade
			Transactional:        true,
			AuditLogged:          true,
		}
		if err := CanApplyItemMutation(req); err != nil {
			t.Errorf("expected nil for valid player trade, got %v", err)
		}
	})

	t.Run("allows valid inventory move", func(t *testing.T) {
		req := ItemMutationAuthorityRequest{
			Source:               ItemMutationSourceInventoryMove,
			BackendValidated:     true,
			ClientCreatedItem:    false,
			BackendGeneratedItem: false, // Not required for move
			Transactional:        false, // Not required for move
			AuditLogged:          true,
		}
		if err := CanApplyItemMutation(req); err != nil {
			t.Errorf("expected nil for valid inventory move, got %v", err)
		}
	})

	baseValidReq := ItemMutationAuthorityRequest{
		BackendValidated:     true,
		ClientCreatedItem:    false,
		BackendGeneratedItem: true,
		Transactional:        true,
		AuditLogged:          true,
	}

	t.Run("rejects invalid source", func(t *testing.T) {
		req := baseValidReq
		req.Source = "client_create"
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrInvalidItemMutationSource) {
			t.Errorf("expected ErrInvalidItemMutationSource, got %v", err)
		}
	})

	t.Run("rejects if not backend validated", func(t *testing.T) {
		req := baseValidReq
		req.Source = ItemMutationSourceLootDrop
		req.BackendValidated = false
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrItemMutationNotBackendValidated) {
			t.Errorf("expected ErrItemMutationNotBackendValidated, got %v", err)
		}
	})

	t.Run("rejects if client created item", func(t *testing.T) {
		req := baseValidReq
		req.Source = ItemMutationSourceLootDrop
		req.ClientCreatedItem = true
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrClientCreatedItemRejected) {
			t.Errorf("expected ErrClientCreatedItemRejected, got %v", err)
		}
	})

	t.Run("rejects if item creation source has no backend generated item", func(t *testing.T) {
		req := baseValidReq
		req.Source = ItemMutationSourceLootDrop
		req.BackendGeneratedItem = false
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrItemMutationNotBackendGenerated) {
			t.Errorf("expected ErrItemMutationNotBackendGenerated, got %v", err)
		}
	})

	t.Run("rejects if transaction source is not transactional", func(t *testing.T) {
		req := baseValidReq
		req.Source = ItemMutationSourcePlayerTrade
		req.BackendGeneratedItem = false // Not required for trade
		req.Transactional = false
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrItemMutationTransactionRequired) {
			t.Errorf("expected ErrItemMutationTransactionRequired, got %v", err)
		}
	})

	t.Run("rejects if not audit logged", func(t *testing.T) {
		req := baseValidReq
		req.Source = ItemMutationSourceInventoryMove
		req.BackendGeneratedItem = false // Not required
		req.Transactional = false        // Not required
		req.AuditLogged = false
		err := CanApplyItemMutation(req)
		if !errors.Is(err, ErrItemMutationAuditRequired) {
			t.Errorf("expected ErrItemMutationAuditRequired, got %v", err)
		}
	})

	t.Run("error order is deterministic", func(t *testing.T) {
		baseReq := ItemMutationAuthorityRequest{
			BackendValidated:     false,
			ClientCreatedItem:    true,
			BackendGeneratedItem: false,
			Transactional:        false,
			AuditLogged:          false,
		}

		t.Run("invalid source is first error", func(t *testing.T) {
			req := baseReq
			req.Source = "client_create"
			err := CanApplyItemMutation(req)
			if !errors.Is(err, ErrInvalidItemMutationSource) {
				t.Errorf("expected ErrInvalidItemMutationSource, got %v", err)
			}
		})

		t.Run("not backend validated is second error", func(t *testing.T) {
			req := baseReq
			req.Source = ItemMutationSourceLootDrop
			err := CanApplyItemMutation(req)
			if !errors.Is(err, ErrItemMutationNotBackendValidated) {
				t.Errorf("expected ErrItemMutationNotBackendValidated, got %v", err)
			}
		})

		t.Run("client created is third error", func(t *testing.T) {
			req := baseReq
			req.Source = ItemMutationSourceLootDrop
			req.BackendValidated = true
			err := CanApplyItemMutation(req)
			if !errors.Is(err, ErrClientCreatedItemRejected) {
				t.Errorf("expected ErrClientCreatedItemRejected, got %v", err)
			}
		})
	})
}

func TestAuthorityBooleans(t *testing.T) {
	if !MustValidateItemMutationsOnBackend() {
		t.Error("expected MustValidateItemMutationsOnBackend to return true")
	}
	if !MustRejectClientCreatedItems() {
		t.Error("expected MustRejectClientCreatedItems to return true")
	}
	if !MustAuditItemMutations() {
		t.Error("expected MustAuditItemMutations to return true")
	}
}
