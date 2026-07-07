# R2-L-C - Debug Loot Gold Fallback Validation

## Status

Closed.

## Related Commit

- 0283d24 Add debug loot gold fallback

## Context

Runtime validation showed Orc Elite death and respawn lifecycle working correctly, but debug loot item grant could fail after target death.

The observed log was:

    Failed to grant debug loot after target death

Inspection showed the debug reward attempted to grant:

    sword_t1_rusty

That item is non-stackable. The current backpack range is slot 0 through slot 29. Runtime logs showed the character inventory loading with 30 items, so AddItem can fail when the backpack is full.

## Implementation Summary

The gateway debug loot block now keeps the original item reward attempt first.

If AddItem("sword_t1_rusty", 1) succeeds:

- inventory is marked dirty
- inventory sync is sent
- item reward log is emitted

If AddItem fails:

- backend attempts playerInv.AddGold(25)
- inventory sync is sent
- fallback gold reward log is emitted

Fallback log:

    Debug loot fallback gold granted after target death

## Architecture Decision

This is a temporary Alpha-safe debug fallback.

It avoids changing backpack size, database state, seed inventory, item stack rules, or persistence schema.

Gold fallback preserves the PvE reward loop even when the debug character inventory is full.

## Validation Recorded

Static validation:

    git diff --check exit code: 0

Backend build validation:

    docker compose build gateway-server exit code: 0

Git validation:

    git commit exit code: 0
    git push origin master exit code: 0
    git status exit code: 0

## Runtime Checkpoint Required

The next runtime checkpoint should confirm:

- Login with default_user.
- Character Gabriela enters the world.
- Orc Elite can be killed.
- If sword_t1_rusty cannot be inserted because inventory is full, gold fallback is granted.
- Log includes Debug loot fallback gold granted after target death.
- Log no longer includes Failed to grant debug loot after target death for the normal full-inventory case.

## Not In Scope

This task does not implement corpse loot.

This task does not change inventory capacity.

This task does not change item stackability.

This task does not modify PostgreSQL schema.

## Next Recommended Task

Run one runtime checkpoint for the R2 RuntimeEntityID + debug loot reward loop before moving deeper into Alpha technical checklist.
## Runtime Checkpoint Completed

Runtime checkpoint completed after commit 0283d24.

Observed runtime flow:

- Login with default_user succeeded.
- Character Gabriela entered the world.
- Inventory loaded with 30 items.
- Gold loaded as 1000.
- Orc Elite registered as creature:debug_orc_elite_001:1.
- Runtime entity 1 was killed.
- Backend emitted SC_TARGET_DEAD opcode 3003.
- Debug loot item insertion could not use a free backpack slot.
- Gold fallback was granted with fallback_gold 25.
- Autosave persisted gold as 1025.
- Scheduler respawned Orc Elite as creature:debug_orc_elite_001:2.
- Client attacked using RuntimeEntityID creature:debug_orc_elite_001:2.
- Backend resolved requested_target creature:debug_orc_elite_001:2 to Orc_Elite.
- Runtime entity 2 was killed.
- Gold fallback was granted again with fallback_gold 25.
- Autosave persisted gold as 1050.

Validation result:

- RuntimeEntityID respawn targeting still works.
- Target dead sync still works.
- Debug loot reward loop no longer fails when backpack is full.
- Gold fallback persists through autosave.
