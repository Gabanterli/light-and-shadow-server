# R2-R-M - RuntimeEntityID Attack Targeting Validation Closure

## Status

Closed.

## Commit Recorded

- 6645d43 Send debug attacks by RuntimeEntityID

## Context

R2-R-L migrated the debug Orc Elite attack flow to prefer the active RuntimeEntityID when available.

The previous debug flow always sent the static target id Orc_Elite.

The new flow keeps Orc_Elite as a safe fallback while allowing post-respawn attacks to target the active runtime entity id.

## Client Behavior

Godot now chooses the debug attack target as follows:

- If _orcEliteRuntimeEntityId is populated, send that RuntimeEntityID.
- If _orcEliteRuntimeEntityId is empty, fall back to Orc_Elite.

Validated post-respawn target:

    creature:debug_orc_elite_001:2

## Backend Behavior

The Gateway now resolves debug Orc Elite attack requests by RuntimeEntityID.

When the requested target matches the active spawn RuntimeEntityID for debug_orc_elite_001, backend resolves it to the current internal combat target:

    requested_target=creature:debug_orc_elite_001:2
    resolved_target=Orc_Elite

This keeps compatibility with the existing CombatManager entity id while allowing the client to move toward RuntimeEntityID-based targeting.

## Validations Recorded

### Client Build

Validated:

    dotnet build LightAndShadow.sln
    dotnet build exit code: 0
    0 warnings
    0 errors

### Backend Build

Validated:

    docker compose build gateway-server
    docker compose build gateway-server exit code: 0

### Runtime Validation

Runtime validation confirmed:

- Initial Orc Elite spawn registered as creature:debug_orc_elite_001:1.
- Player Gabriela killed Orc_Elite.
- Backend marked debug_orc_elite_001 dead with runtime_entity_id creature:debug_orc_elite_001:1.
- Backend emitted SC_TARGET_DEAD opcode 3003.
- Backend scheduler respawned Orc Elite as creature:debug_orc_elite_001:2.
- Backend emitted respawn visual sync opcode 3004 with runtime_entity_id creature:debug_orc_elite_001:2.
- Client then sent attack requests with the RuntimeEntityID target.
- Backend logged Resolved debug Orc Elite attack target by runtime entity id.
- Backend resolved creature:debug_orc_elite_001:2 to Orc_Elite.
- Second kill and debug loot generation succeeded against runtime_entity_id creature:debug_orc_elite_001:2.

## Known Limitation

Death packet opcode 3003 still emits the static visual target id:

    Orc_Elite

This is acceptable for the current task because R2-R-L only migrated attack targeting.

RuntimeEntityID death/dead-state sync should be handled in a separate task.

## Next Recommended Task

R2-R-N - Document or implement RuntimeEntityID death sync strategy.

Potential scope:

- Decide whether SC_TARGET_DEAD should carry RuntimeEntityID.
- Keep visual compatibility with Orc_Elite during debug phase.
- Avoid breaking current dead-state UI.
- Do not change loot/corpse ownership in the same task.
- Keep stale target validation as a separate task if needed.

## Architecture Decision

RuntimeEntityID is now the preferred attack target after respawn.

Orc_Elite remains a temporary debug fallback and internal combat compatibility bridge.

The project should continue migrating PvE lifecycle packets toward authoritative RuntimeEntityID targeting, while keeping compatibility layers until multiple-creature AOI is introduced.
