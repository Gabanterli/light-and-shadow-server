# R2-R-N - RuntimeEntityID Death Sync Strategy

## Status

Closed.

## Context

R2-R-L migrated debug attack targeting to prefer the active RuntimeEntityID after respawn.

Runtime validation confirmed that the client attacked creature:debug_orc_elite_001:2 and the backend resolved that runtime target to the internal combat target Orc_Elite.

However, death sync still emits the static visual target id Orc_Elite through SC_TARGET_DEAD opcode 3003.

## Current Behavior

Current death packet behavior:

    SC_TARGET_DEAD / opcode 3003
    TargetID: Orc_Elite

This keeps the existing Godot debug dead-state working because the client currently checks:

    data.TargetID == "Orc_Elite"

## Problem

As soon as multiple creatures of the same type exist, static death targeting becomes unsafe.

Static target ids create ambiguity:

- Two or more Orc Elite instances could share the same visual/debug name.
- Client-side dead state could update the wrong creature.
- Loot/corpse ownership could become ambiguous if death sync remains name-based.
- AOI creature despawn/death replication requires a stable runtime identity.

## Architecture Decision

Future death sync must become RuntimeEntityID-aware.

The recommended migration path is additive, not destructive.

SC_TARGET_DEAD should keep TargetID for temporary visual compatibility and add RuntimeEntityID in a future protocol-safe migration.

Recommended future death payload fields:

    TargetID
    RuntimeEntityID

During the debug phase:

- TargetID remains Orc_Elite.
- RuntimeEntityID carries the authoritative runtime entity id, for example creature:debug_orc_elite_001:2.
- Godot should prefer RuntimeEntityID when available.
- Godot should fall back to TargetID only when RuntimeEntityID is empty.

## Not In Scope For This Task

This task does not modify code.

Do not change:

- SC_TARGET_DEAD packet format yet.
- BinaryProtocol.cs decoder yet.
- DebugWorldEntryController.cs dead-state logic yet.
- Backend death packet emission yet.
- Loot or corpse ownership.
- AOI creature generalization.
- Stale target validation.

## Future Implementation Task

Recommended next code task:

    R2-R-O - Add RuntimeEntityID to debug target dead sync with Orc_Elite fallback

Safe scope for R2-R-O:

- Backend emits TargetID plus RuntimeEntityID for debug Orc Elite death sync.
- Client decodes RuntimeEntityID from opcode 3003 if present.
- Client logs RuntimeEntityID on death event.
- Client keeps Orc_Elite fallback.
- Existing visual dead state must not break.
- No loot/corpse changes.
- No AOI generalization.

## Validation Requirements For Future Code Task

R2-R-O must validate:

- dotnet build LightAndShadow.sln exit code 0.
- docker compose build gateway-server exit code 0.
- Runtime kill before respawn still marks Orc_Elite dead visually.
- Runtime kill after respawn logs RuntimeEntityID for death sync.
- No regression in debug loot generation.
- No regression in scheduler respawn opcode 3004.

## Long-Term Direction

RuntimeEntityID must become the primary identity for all creature lifecycle events:

- spawn
- damage
- death
- corpse
- loot ownership
- despawn
- respawn
- AOI replication

Static ids like Orc_Elite should remain only as temporary debug compatibility aliases.
