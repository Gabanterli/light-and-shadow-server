\# R2-C-D — Creature Combat State Runtime Closure



\## Status



R2-C — Creature Combat State is closed as a runtime-validated foundation block.



This document closes the transition from the R1 debug Orc\_Elite combat loop into the R2 creature runtime state model.



\## Covered Commits



\- c41dfc2 Route Orc Elite death through spawn state

\- 6612848 Guard Orc Elite death and debug loot

\- 4e360a6 Track Orc Elite damage contributions



\## Context



During R1, Orc\_Elite existed primarily as a debug combat target registered directly in the combat flow.



During R2-C, Orc\_Elite was connected to the new CreatureSpawnManager lifecycle model. The debug target is still compatible with the old client button and target ID, but the backend now tracks creature runtime state for death, loot guard, damage contribution, revive/retry, and runtime entity versioning.



The current Orc\_Elite debug spawn is:



\- SpawnID: debug\_orc\_elite\_001

\- CreatureID: orc\_elite

\- Debug client target ID: Orc\_Elite

\- RuntimeEntityID format: creature:debug\_orc\_elite\_001:{version}



Examples observed during runtime validation:



\- creature:debug\_orc\_elite\_001:1

\- creature:debug\_orc\_elite\_001:2



\## Runtime Systems Validated



\### CreatureSpawnManager Integration



The gateway now owns a CreatureSpawnManager instance and registers the Orc\_Elite spawn state during character select.



Validated log:



\- Registered Orc Elite creature spawn state



\### Death State Transition



The Orc\_Elite death transition is now routed through CreatureSpawnManager.MarkDead.



MarkDead acts as the authoritative death-state gate for the debug Orc\_Elite flow.



Validated log:



\- Marked Orc Elite creature spawn dead



\### Double-Death Guard



SC\_TARGET\_DEAD is sent only when MarkDead accepts the transition.



Duplicate death transitions are blocked by the creature spawn state.



Validated log:



\- Orc Elite duplicate death blocked by creature spawn state



\### Debug Loot Guard



Debug loot generation is now gated by CreatureSpawnManager.MarkLootGenerated.



This prevents duplicate loot generation for the same runtime spawn lifecycle.



Validated logs:



\- Orc Elite debug loot generation guard accepted

\- Orc Elite debug loot generation blocked by guard

\- Debug loot granted after target death



\### Damage Contribution Tracking



CreatureSpawnState now tracks damage contribution by player.



Current field:



\- DamageContributors map\[string]float64



Current method:



\- AddDamageContribution(spawnID, playerID string, damage float64) (\*CreatureSpawnState, bool)



The gateway records contribution after successful attacks against Orc\_Elite when damage is greater than zero.



Validated log:



\- Recorded Orc Elite damage contribution



\### Retry Revive Flow



The debug retry flow still exists.



When the player attacks Orc\_Elite after death, the backend revives the debug combat entity and syncs the CreatureSpawnManager state through ReviveRespawn.



ReviveRespawn currently:



\- Sets Alive back to true.

\- Clears DiedAt.

\- Clears NextRespawnAt.

\- Clears KillerPlayerID.

\- Clears LastDamagerPlayerID.

\- Resets LootGenerated.

\- Increments Version.

\- Rebuilds RuntimeEntityID.

\- Resets DamageContributors.



Validated log:



\- Debug Orc Elite creature spawn state revived for retry flow



\### Client Compatibility



The client still attacks using the debug target ID:



\- Orc\_Elite



This remains temporary until final protocol routing uses RuntimeEntityID.



The current flow remains compatible with:



\- SC\_DAMAGE\_EVENT opcode 3002

\- SC\_TARGET\_DEAD opcode 3003

\- SC\_INVENTORY\_SYNC opcode 4001



\## Runtime Validation Summary



Validated end-to-end flow:



login -> character list -> character select -> world entry -> attack Orc\_Elite -> damage contribution -> death guard -> target dead packet -> debug loot guard -> inventory sync -> retry revive -> new runtime entity ID -> damage contribution reset -> second death -> second loot



Validated important logs:



\- Registered Orc Elite creature spawn state

\- Recorded Orc Elite damage contribution

\- Marked Orc Elite creature spawn dead

\- Orc Elite debug loot generation guard accepted

\- Sending target dead packet to client

\- Target dead packet sent to client

\- Debug loot granted after target death

\- Debug Orc Elite creature spawn state revived for retry flow



Validated build/runtime checks:



\- Godot C# build OK.

\- Docker gateway-server build OK.

\- Runtime kill/retry/kill OK.

\- Inventory sync continued working.

\- items\_count increased correctly after loot grants.



\## Remaining Risks



\### Debug Target ID



The flow still uses the fixed debug target ID:



\- Orc\_Elite



This is not the final MMO protocol model.



Future creature combat must route target selection and attack requests through RuntimeEntityID.



\### Parallel Combat State



CombatManager and CreatureSpawnManager are still partially parallel systems.



Current state:



\- CombatManager owns immediate combat entity behavior.

\- CreatureSpawnManager owns creature lifecycle state.



Future work must converge these responsibilities through a clean creature runtime combat adapter or authoritative creature runtime entity model.



\### Debug Respawn



Respawn is still driven by debug retry behavior.



The real respawn timer has not been implemented yet.



This must be replaced by a server-authoritative respawn scheduler.



\### Debug Loot



Loot is still fixed and debug-only:



\- sword\_t1\_rusty x1



There is no real loot table, no corpse container, no ownership, no party distribution, and no contribution-based eligibility yet.



\### Damage Contribution Precision



DamageContributors currently uses float64.



This is acceptable for debug validation, but future loot eligibility, audit logs, anti-abuse, and economy-sensitive systems should avoid raw floating-point totals.



Recommended future options:



\- Store damage in scaled integers.

\- Round contribution totals at system boundaries.

\- Keep deterministic combat audit values for loot eligibility.



\### AOI Spawn/Despawn Missing



Creature spawn/despawn is not yet integrated with AOI streaming.



Future PvE work must prevent:



\- AOI spoofing.

\- Attacking unseen creatures.

\- Looting despawned corpses.

\- Receiving stale creature state after respawn.



\### Persistence Missing



Creature runtime state is currently in-memory.



Future persistence rules must define what survives:



\- Server restart.

\- World shard migration.

\- Crash recovery.

\- Creature death during shutdown.

\- Corpse lifetime.

\- Loot claim status.



\## Architectural Decision



R2-C establishes CreatureSpawnManager as the authoritative lifecycle guard for creature death, loot generation, damage contribution, and runtime versioning.



CombatManager remains responsible for immediate combat processing for now.



The current architecture is acceptable as an R2 bridge, but the final PvE model must route creature combat by RuntimeEntityID and eventually remove the fixed Orc\_Elite debug path.



\## Closure



R2-C is closed.



The project is ready to proceed to the next R2 block.



Recommended next task:



\- R2-R-A — Add respawn timer to spawn state



Alternative next task:



\- R2-L-A — Define loot table runtime model



Architectural recommendation:



R2-R-A should happen before real loot tables, because a real loot economy should be built on top of a stable creature lifecycle and server-authoritative respawn model.

