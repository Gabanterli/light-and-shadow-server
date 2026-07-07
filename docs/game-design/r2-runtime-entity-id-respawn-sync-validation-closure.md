# R2-R-K — RuntimeEntityID Respawn Sync Validation Closure

## Status

Closed.

This document records the validation closure for the RuntimeEntityID respawn visual sync work completed in commit:

- 2a42f07 Add RuntimeEntityID to respawn visual sync

## Context

R2-R-J added RuntimeEntityID propagation to the debug creature respawn visual sync flow.

The goal was to ensure that, after the debug Orc Elite respawns, the backend emits the active runtime entity identifier to the client so future targeting can migrate away from the static debug name Orc_Elite.

This is part of the R2 Real PvE Foundation work and prepares the project for authoritative RuntimeEntityID-based creature targeting.

## Implementation Recorded

The referenced implementation updated the backend and Godot client protocol flow so that respawn sync now carries the active runtime entity id.

Relevant behavior:

- Backend respawn scheduler creates the next runtime creature entity.
- Backend emits respawn visual sync opcode 3004.
- Opcode 3004 includes runtime_entity_id.
- Godot has a decoder for the respawn event.
- Godot has CreatureRespawnEventData.
- Godot stores the received RuntimeEntityID in _orcEliteRuntimeEntityId.

## Validations Recorded

### Client Build

Validated:

    dotnet build LightAndShadow.sln
    dotnet build exit code: 0

### Backend Build

Validated:

    docker compose build gateway-server
    docker compose build gateway-server exit code: 0

### Backend Runtime

Backend runtime validation confirmed that the scheduler respawned the debug Orc Elite as:

    creature:debug_orc_elite_001:2

Backend emitted opcode 3004 with:

    runtime_entity_id="creature:debug_orc_elite_001:2"

After respawn, backend damage was registered against the correct active runtime entity:

    creature:debug_orc_elite_001:2

This confirms that the backend authoritative lifecycle is now using the respawned runtime entity after the first death/respawn cycle.

## Client-Side Validation Caveat

Godot build validation passed with the new decoder and data structure.

However, the chat did not explicitly confirm a Godot runtime log showing:

    RuntimeEntityID: creature:debug_orc_elite_001:2

from opcode 3004.

Therefore, client-side runtime logging and visual confirmation should still be treated as partially confirmed by build and protocol implementation, but not fully closed by explicit observed Godot log evidence.

## Current Targeting Limitation

Debug attack still uses the static target identifier:

    Orc_Elite

This is acceptable for the current closure because R2-R-J only covered respawn visual sync and RuntimeEntityID propagation.

The next code task must migrate debug attack targeting to prefer the active RuntimeEntityID while keeping a safe fallback to Orc_Elite.

## Next Task Prepared

Next recommended task:

    R2-R-L — Send debug attack request using active RuntimeEntityID with Orc_Elite fallback

Safe scope for R2-R-L:

- Client sends _orcEliteRuntimeEntityId when available.
- Client falls back to Orc_Elite when RuntimeEntityID is empty.
- Backend resolves target by RuntimeEntityID first.
- Backend keeps fallback support for Orc_Elite.
- Stale target validation remains a separate future task.
- Loot and corpse logic remain unchanged.
- AOI generalization remains unchanged.

## Architecture Decision

RuntimeEntityID is now the correct path for creature targeting after respawn.

The static debug name Orc_Elite should be treated only as a temporary compatibility fallback until the full R2 targeting migration is complete.
