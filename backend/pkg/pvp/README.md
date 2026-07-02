# PvP System v1: Amulet of Loss & Blessings Interaction

This document specifies the authoritative interaction rules between the **Amulet of Loss (AoL)** and **Blessings** upon a player's death, implemented in the canonical patch `aol_bless_interaction_patch`.

---

## Canonical Core Rules

1. **Blessings** are the primary mechanism for XP-loss mitigation.
2. **Amulet of Loss** is primarily an item-loss protection system.
3. **AoL XP mitigation** applies **ONLY** when the player has zero (0) active blessings.

---

## Canonical Death Calculations

### 1. No Protection
* **Trigger**: Active blessings = `0` and Amulet of Loss is **not** equipped.
* **Item Loss**: Enabled (standard roll based on equipment/backpack chances).
* **XP Loss**: `10%` of the experience required for the player's next level.

### 2. Amulet of Loss Only
* **Trigger**: Active blessings = `0` and Amulet of Loss **is** equipped.
* **Item Loss**: Disabled. All equipped items and backpack items are fully protected.
* **XP Loss**: Reduced by `50%` to exactly `5%` of the experience required for the next level.
* **AoL Consumption**: The Amulet of Loss is consumed on death.

### 3. Blessings Active
* **Trigger**: Active blessings > `0` (1 to 7 blessings) and Amulet of Loss is **not** equipped.
* **Item Loss**: Disabled. Active blessings prevent all item losses.
* **XP Loss**: Reduced by `12%` per active blessing.
  * Formula: `BaseLoss (10%) * (1.0 - 0.12 * blessingsCount)`
  * With exactly 7 blessings: `1.6%` XP loss.
* **Blessings Consumption**: All active blessings are consumed on death.

### 4. Amulet of Loss + Blessings
* **Trigger**: Active blessings > `0` (1 to 7 blessings) and Amulet of Loss **is** equipped.
* **Item Loss**: Disabled. Protected by both active blessings and AoL.
* **XP Loss**: Determined **exclusively** by the number of active blessings. The AoL does **not** provide any further XP mitigation.
  * With 3 blessings and AoL equipped: XP loss is `6.4%` (matching blessings-only calculation).
* **Consumption**: The Amulet of Loss is consumed, and all active blessings are consumed.

---

## Implementation Details

* **XP & Item Protection Manager**: `ApplyDeathPenalties` in `backend/pkg/lifecycle/death_penalty_manager.go`.
* **Amulet of Loss Checks & Consumption**: Implemented thread-safely in `backend/pkg/inventory/inventory_manager.go` via `IsAoLEquipped` and `ConsumeAoL`.
* **Active Blessings Count**: Retrieved thread-safely via `GetActiveBlessingsCount` in `backend/pkg/blessing/blessing_manager.go`.
