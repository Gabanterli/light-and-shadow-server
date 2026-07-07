# R2-R-S - Stale RuntimeEntityID Target Guard Validation Closure

## Status

Closed.

## Related Tasks

- R2-R-Q - Document stale RuntimeEntityID target guard requirements
- R2-R-R - Reject stale debug RuntimeEntityID attack targets

## Context

RuntimeEntityID is now used in the debug Orc Elite lifecycle for respawn sync, post-respawn attack targeting, and target dead sync.

The remaining immediate risk was stale RuntimeEntityID targeting.

A stale RuntimeEntityID is an old runtime id for a creature instance that is no longer the active spawned entity.

Example:

    creature:debug_orc_elite_001:1

becomes stale after respawn creates:

    creature:debug_orc_elite_001:2

## Implementation Summary

The gateway attack handler now rejects debug Orc Elite runtime targets that match the debug runtime id prefix but do not resolve to the active creature spawn state.

Accepted targets:

- Orc_Elite temporary debug fallback.
- Active RuntimeEntityID for debug_orc_elite_001.

Rejected targets:

- Stale RuntimeEntityID values for debug_orc_elite_001.
- Unknown debug runtime ids matching the stale debug prefix.

## Safe Failure Behavior

Rejected stale targets return SC_DAMAGE_EVENT with:

- damage: 0
- isCrit: false
- isHit: false
- success: false
- skillName: stale target

This prevents stale attacks from reaching CombatManager and avoids duplicate death, duplicate damage contribution, and stale loot/corpse side effects.

## Validation Recorded

### Static Validation

Required validation for backend patch:

    docker gofmt exit code: 0
    git diff --check exit code: 0
    docker compose build gateway-server exit code: 0

### Runtime Regression Validation

Manual runtime flow confirmed:

- Login with default_user succeeded.
- Character Gabriela entered the world.
- Orc Elite registered as creature:debug_orc_elite_001:1.
- Player killed runtime entity 1.
- Target dead flow succeeded.
- Scheduler respawned Orc Elite as creature:debug_orc_elite_001:2.
- Client attacked after respawn using RuntimeEntityID.
- Backend resolved requested_target creature:debug_orc_elite_001:2 to Orc_Elite.
- Second death succeeded for runtime entity 2.
- Scheduler later respawned runtime entity 3.

## Known Limitation

The normal client does not yet expose a manual stale-target injection path.

Therefore, the runtime validation confirmed that the stale guard did not regress active RuntimeEntityID targeting.

A future protocol/debug test harness can explicitly send creature:debug_orc_elite_001:1 after active runtime changes to creature:debug_orc_elite_001:2.

## Architecture Result

The backend now has a first stale RuntimeEntityID guard for the debug Orc Elite lifecycle.

This reduces the risk of stale client state causing damage, death, loot, or corpse side effects against an old creature instance.

## Next Recommended Alpha Task

Fix debug loot grant after Orc Elite death.

Reason:

Runtime logs still show Failed to grant debug loot after target death.

That issue does not invalidate RuntimeEntityID lifecycle work, but it blocks a clean Alpha PvE reward loop.
