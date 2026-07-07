# R2-R-P - RuntimeEntityID Target Dead Sync Validation Closure

## Status

Closed.

## Commits Recorded

- ee85bc7 Add RuntimeEntityID to target dead sync
- 99c8cab Emit RuntimeEntityID in target dead sync

## Context

R2-R-O and R2-R-O2 completed the RuntimeEntityID-aware target dead sync migration for the debug Orc Elite lifecycle.

Before this work, SC_TARGET_DEAD opcode 3003 only carried the static target id Orc_Elite.

That was sufficient for the single debug creature, but unsafe for future multiple-creature AOI because static target ids cannot uniquely identify creature instances.

## Client Implementation

Godot now decodes opcode 3003 as:

    TargetID
    RuntimeEntityID

The decoder remains backward-compatible:

- If RuntimeEntityID is present, it is decoded.
- If RuntimeEntityID is missing, it falls back to an empty string.

DebugWorldEntryController now logs RuntimeEntityID on target dead events and keeps Orc_Elite visual fallback compatibility.

## Backend Implementation

Gateway now captures the RuntimeEntityID from the authoritative creature spawn state when debug Orc Elite is marked dead.

Backend then emits SC_TARGET_DEAD opcode 3003 with:

    TargetID: Orc_Elite
    RuntimeEntityID: creature:debug_orc_elite_001:<version>

The backend implementation keeps the existing TargetID for visual compatibility while adding RuntimeEntityID for authoritative lifecycle identity.

## Validations Recorded

### Backend Formatting

Validated:

    docker gofmt exit code: 0

### Diff Validation

Validated:

    git diff --check exit code: 0

### Backend Build

Validated after commit 99c8cab:

    docker compose build gateway-server exit code: 0

### Runtime Validation

Runtime validation confirmed by manual gameplay test:

- Login succeeded with default_user.
- Character Gabriela entered the world.
- Orc Elite spawn state registered with RuntimeEntityID.
- Player killed Orc_Elite.
- Godot received SC_TARGET_DEAD opcode 3003.
- Godot target dead log included Target Dead: Orc_Elite.
- Godot target dead log included RuntimeEntityID.
- Scheduler respawned Orc Elite as creature:debug_orc_elite_001:2.
- Client attacked again using RuntimeEntityID targeting.
- Backend resolved the runtime target to Orc_Elite.
- Second death/loot flow succeeded.
- Godot target dead sync included RuntimeEntityID for the respawned runtime entity.

## Architecture Result

Creature death sync is now RuntimeEntityID-aware while preserving the current debug Orc_Elite compatibility path.

This closes the immediate RuntimeEntityID targeting/death sync gap for the R2 real respawn debug lifecycle.

## Remaining Limitations

The system still uses Orc_Elite as an internal debug combat alias.

Future tasks must eventually migrate:

- CombatManager entity identity
- damage events
- corpse ownership
- loot ownership
- AOI creature replication
- multi-creature rendering

to first-class RuntimeEntityID-based identity.

## Next Recommended Task

R2-R-Q - Document RuntimeEntityID lifecycle migration closure or start stale target guard planning.

Recommended next code direction:

- Add stale RuntimeEntityID target guard.
- Reject or safely resolve attacks against old runtime ids.
- Keep Orc_Elite fallback only for current debug phase.
- Do not mix with loot/corpse/AOI generalization.
