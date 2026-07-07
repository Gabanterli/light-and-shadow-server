# R2-R-Q - Stale RuntimeEntityID Target Guard Requirements

## Status

Closed.

## Context

The R2 RuntimeEntityID migration now supports respawn sync, attack targeting, and target dead sync for the debug Orc Elite lifecycle.

Current validated flow:

- Client receives RuntimeEntityID on creature respawn opcode 3004.
- Client sends RuntimeEntityID for debug attacks after respawn.
- Backend resolves active RuntimeEntityID to the internal Orc_Elite combat alias.
- Target dead sync opcode 3003 can include RuntimeEntityID.

## Problem

The next lifecycle risk is stale RuntimeEntityID targeting.

A stale RuntimeEntityID is an old creature runtime id that is no longer the active live entity for a spawn.

Example:

    creature:debug_orc_elite_001:1

becomes stale after the spawn respawns as:

    creature:debug_orc_elite_001:2

If the backend accepts stale RuntimeEntityID targets, players could attack dead, despawned, or replaced creature instances.

## Exploit And Consistency Risks

Stale targeting can cause:

- Attacks against dead entities.
- Duplicate damage contribution records.
- Incorrect loot eligibility.
- Incorrect corpse ownership.
- Client/server desync.
- Future multi-creature AOI ambiguity.
- Possible economy exploits if stale death/loot flows are triggered.

## Architecture Decision

Backend must treat RuntimeEntityID as authoritative and versioned.

The active RuntimeEntityID for a spawn is the only valid runtime target for that spawn.

Stale RuntimeEntityID targets must not silently resolve to the current live creature.

Recommended behavior:

- If target is Orc_Elite, allow temporary debug fallback.
- If target is the active RuntimeEntityID for debug_orc_elite_001, resolve to Orc_Elite.
- If target looks like a stale RuntimeEntityID for debug_orc_elite_001, reject safely.
- If target is unknown, reject safely.

## Safe Rejection Behavior

Rejected stale target attacks should return a normal failed damage event instead of crashing or disconnecting the client.

Recommended error result:

    damage=0
    isCrit=false
    isHit=false
    success=false
    skillName=stale target

The response should preserve the requested target id so the client can log what failed.

## Not In Scope

This document does not implement code.

Do not change in this task:

- CombatManager entity identity.
- Loot ownership.
- Corpse generation.
- AOI replication.
- Multiple creature rendering.
- Spawn table generalization.
- Client retargeting UI.

## Future Code Task

Recommended next code task:

    R2-R-R - Reject stale debug RuntimeEntityID attack targets

Safe scope for R2-R-R:

- Detect RuntimeEntityID targets for debug_orc_elite_001.
- Resolve only the active RuntimeEntityID.
- Reject stale RuntimeEntityID values safely.
- Keep Orc_Elite fallback for debug compatibility.
- Return a failed SC_DAMAGE_EVENT for rejected stale targets.
- Do not change loot, corpse, AOI, or combat formulas.

## Validation Requirements For R2-R-R

Required validations:

- git status clean before task.
- gofmt via Docker.
- git diff --check exit code 0.
- docker compose build gateway-server exit code 0.
- Runtime normal Orc_Elite fallback attack still works.
- Runtime active RuntimeEntityID attack still works after respawn.
- Runtime stale RuntimeEntityID attack is rejected safely.
- No duplicate loot.
- No duplicate death packet.
- Working tree clean after push.

## Long-Term Direction

RuntimeEntityID must become the primary identity for all creature lifecycle operations.

Static ids like Orc_Elite should remain only as temporary debug aliases until CombatManager, AOI replication, corpse ownership, loot ownership, and client creature rendering are fully RuntimeEntityID-based.
