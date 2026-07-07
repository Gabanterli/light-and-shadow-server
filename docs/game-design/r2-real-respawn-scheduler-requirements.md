\# R2-R-C â€” Real Respawn Scheduler Requirements



\## Status



This document defines the requirements for the real server-side creature respawn scheduler planned after the first Orc\_Elite respawn timer validation.



R2-R-A validated the first server-authoritative respawn timer gate through CreatureSpawnManager.TryRespawnDue.



R2-R-B documented the successful runtime validation of that gate.



R2-R-C defines what must exist before replacing the current attack-triggered respawn bridge with a broader MMO-safe respawn scheduler.



\## Current Runtime State



The current Orc\_Elite debug spawn lifecycle supports:



\* Spawn registration through CreatureSpawnManager.

\* RuntimeEntityID generation.

\* Damage contribution tracking.

\* MarkDead death transition.

\* NextRespawn scheduling.

\* MarkLootGenerated loot guard.

\* TryRespawnDue timer gate.

\* ReviveRespawn debug retry fallback.

\* RuntimeEntityID version increment on respawn.

\* DamageContributors reset on respawn.

\* LootGenerated reset on respawn.



Current validated spawn:



\* SpawnID: debug\_orc\_elite\_001

\* CreatureID: orc\_elite

\* Debug target ID: Orc\_Elite

\* RuntimeEntityID examples:



&#x20; \* creature:debug\_orc\_elite\_001:1

&#x20; \* creature:debug\_orc\_elite\_001:2



\## Current Limitation



The respawn timer is currently checked only when the player attacks Orc\_Elite after death.



This is acceptable as an R2 bridge, but it is not the final MMO architecture.



The final architecture must not depend on client interaction to process respawn timing.



\## Required Final Direction



Creature respawn must become server-authoritative and scheduler-driven.



The world server or gateway-side PvE runtime must periodically evaluate dead creature spawns and respawn only those whose NextRespawn timestamp is due.



The respawn scheduler must not trust client packets, target IDs, local visuals, or retry actions.



\## Scheduler Ownership



The respawn scheduler should eventually belong to the authoritative world lifecycle layer.



Acceptable intermediate owner during R2:



\* Gateway-side PvE runtime loop, if isolated and documented.



Preferred final owner:



\* World server creature lifecycle subsystem.



The scheduler should not be owned by UI, client input, debug buttons, inventory, loot code, or combat packet handlers.



\## Required Scheduler Behavior



The scheduler must:



\* Iterate over known creature spawn states.

\* Select dead spawns only.

\* Ignore alive spawns.

\* Ignore spawns without NextRespawn set.

\* Respawn only when now >= NextRespawn.

\* Generate a new RuntimeEntityID.

\* Reset CurrentHP to MaxHP.

\* Reset DiedAt.

\* Reset NextRespawn.

\* Reset KillerPlayerID.

\* Reset LastDamagerPlayerID.

\* Reset DamageContributors.

\* Reset LootGenerated.

\* Increment Version.

\* Emit structured logs.

\* Eventually notify AOI clients.



\## Double-Respawn Guard



The scheduler must be protected against double-respawn.



A spawn can respawn only if all conditions are true:



\* Spawn exists.

\* Spawn is dead.

\* NextRespawn is not zero.

\* now >= NextRespawn.

\* Respawn transition wins the lock.

\* Spawn is still dead after lock acquisition.



If another goroutine respawns the spawn first, the scheduler must skip it safely.



\## Concurrency Requirements



CreatureSpawnManager must remain concurrency-safe.



Respawn scheduler calls must use manager methods rather than mutating spawn state directly.



The scheduler must not return internal pointers to mutable state.



Returned spawn states must remain cloned snapshots.



\## RuntimeEntityID Requirements



Every respawn must create a new RuntimeEntityID.



RuntimeEntityID is the future authoritative protocol target identity.



Current debug target ID:



\* Orc\_Elite



Future protocol target ID:



\* creature:{spawn\_id}:{version}



The scheduler must assume old RuntimeEntityIDs become invalid after respawn.



\## CombatManager Integration



Current state:



\* CombatManager owns immediate combat stats and damage processing.

\* CreatureSpawnManager owns lifecycle state.



Scheduler-driven respawn must keep both systems consistent.



When a creature respawns, the combat entity must also be alive with full health.



Intermediate R2 approach:



\* TryRespawnDue returns a respawned spawn state.

\* Gateway calls CombatManager.ReviveEntity for Orc\_Elite.



Future approach:



\* Creature runtime entity model owns both lifecycle and combat state.

\* CombatManager reads or updates through the authoritative creature runtime model.



\## AOI Requirements



The final scheduler must notify nearby players when a creature respawns.



AOI spawn broadcast must include:



\* RuntimeEntityID.

\* CreatureID.

\* SpawnID if needed for debugging.

\* Position X/Y/Z.

\* CurrentHP.

\* MaxHP.

\* Alive state.

\* Version.

\* Optional visual state.



AOI must prevent clients from seeing or targeting stale creature versions.



\## Despawn Requirements



When a creature dies, the server must eventually distinguish between:



\* Dead creature visual.

\* Corpse object.

\* Loot container.

\* Fully despawned creature.

\* Respawned new creature runtime entity.



The scheduler must not respawn a creature in a way that conflicts with an active corpse or loot container.



\## Corpse and Loot Requirements



Before full real loot integration, the scheduler must define corpse behavior.



Open decisions:



\* Does corpse remain until loot is taken?

\* Does corpse expire before respawn?

\* Can creature respawn while corpse still exists?

\* Does corpse store RuntimeEntityID of the dead creature?

\* Does corpse store SpawnID?

\* Does corpse block respawn tile occupancy?

\* Does loot ownership expire independently from corpse lifetime?



Recommended MMO-safe direction:



\* Dead creature creates corpse/container.

\* Corpse has independent lifetime.

\* Respawn may occur while corpse exists only if tile occupancy and AOI rules are safe.

\* Loot ownership is tied to dead RuntimeEntityID, not new RuntimeEntityID.



\## Loot Ownership Requirements



Future loot ownership must not be confused by respawn.



Loot from:



\* creature:debug\_orc\_elite\_001:1



must never be claimable as loot from:



\* creature:debug\_orc\_elite\_001:2



The loot system must bind generated loot to the dead runtime entity version.



\## Damage Contribution Requirements



DamageContributors must reset on respawn.



Damage contribution for a dead runtime entity must be snapshotted before reset if used for:



\* Loot eligibility.

\* Party distribution.

\* Anti-leech checks.

\* PvE analytics.

\* Kill credit.

\* Quest credit.

\* Economy audit.



Current float64 contribution totals are acceptable for debug validation.



Future loot eligibility should use deterministic integer or scaled fixed-point values.



\## Persistence Requirements



The scheduler design must define what happens during:



\* Gateway restart.

\* World server restart.

\* Crash after death before loot generation.

\* Crash after loot generation before respawn.

\* Crash after respawn before persistence.

\* Multi-shard migration.



Initial R2 can remain in-memory.



Future MMO architecture must persist enough state to avoid:



\* Duplicate loot.

\* Lost corpse state.

\* Respawn skipping.

\* Permanent dead spawns.

\* Resetting rare spawns on crash.



\## Timing Requirements



Respawn timing must use server time.



Client time must never determine respawn eligibility.



All timestamps should be UTC.



Scheduler tick interval should be configurable.



Recommended initial interval:



\* 1 second for debug and early R2 validation.



Future production interval:



\* Configurable by world region and creature type.

\* Potentially coalesced by timing wheel or priority queue for large-scale spawn counts.



\## Performance Requirements



The scheduler must scale beyond one debug creature.



Risks of naive scanning:



\* O(n) scan over all spawns every tick.

\* Lock contention in CreatureSpawnManager.

\* AOI broadcast burst if many spawns respawn at once.

\* Respawn thundering herd after server recovery.



Future optimization options:



\* Priority queue ordered by NextRespawn.

\* Timing wheel.

\* Region-local scheduler shards.

\* Chunk-local spawn managers.

\* Batch AOI updates.

\* Respawn jitter for non-boss creatures.



\## PvP and Exploit Requirements



The respawn system must prevent:



\* Client-triggered respawn farming.

\* Double-loot through respawn race conditions.

\* Attacking stale dead RuntimeEntityID.

\* Looting new creature version with old corpse rights.

\* Forcing respawn by disconnect/reconnect.

\* Resetting aggro or contribution unfairly during combat.

\* Respawn blocking or griefing through tile occupancy.

\* Pulling respawned monsters into safe zones.



\## Safe Zone Requirements



Creature respawn must respect world region rules.



A creature must not respawn inside protected city areas unless explicitly configured.



Respawn validation should eventually check:



\* Region type.

\* Tile walkability.

\* Spawn area bounds.

\* Occupancy.

\* PvP/PvE restrictions.

\* Instance/dungeon ownership.



\## Debug Retry Removal Requirements



The debug retry fallback can be removed only after:



\* Scheduler respawn is active.

\* AOI respawn broadcast works.

\* Client can visually reset creature state from server packets.

\* CombatManager and CreatureSpawnManager remain synchronized.

\* RuntimeEntityID targeting is available or a safe compatibility bridge exists.

\* Runtime validation confirms kill -> corpse/death -> respawn -> new attack loop.



Until then, debug retry may remain as a temporary bridge.



\## Required Logs



Scheduler logs should include:



\* Respawn scheduler tick started.

\* Spawn skipped because alive.

\* Spawn skipped because respawn not due.

\* Spawn respawned.

\* Combat revive failed after spawn respawn.

\* AOI respawn broadcast emitted.

\* Duplicate respawn blocked.

\* Respawn scheduler tick completed.



For production, high-volume logs must be rate-limited or debug-level.



\## Recommended Next Implementation



Recommended next engineering task:



\* R2-R-D â€” Add minimal creature respawn scheduler loop



Initial implementation scope:



\* Add manager method to list due dead spawns or try respawn by due time.

\* Add minimal gateway-side loop for debug Orc\_Elite only.

\* Keep interval simple.

\* Keep debug retry fallback.

\* Do not add real loot.

\* Do not add corpse.

\* Do not add RuntimeEntityID protocol routing yet.

\* Build validate with docker compose build gateway-server.

\* Runtime validate login -> kill -> wait -> automatic or scheduler-processed respawn behavior.



\## Out of Scope For Next Task



Do not implement in the next scheduler task:



\* Real loot tables.

\* Party loot.

\* Corpse containers.

\* Persistent spawn state.

\* Full AOI spawn packet.

\* Client-side creature list UI.

\* AI pathfinding.

\* Multi-creature spawn config migration.

\* RuntimeEntityID protocol migration.



\## Closure



R2-R-C defines the architectural requirements for moving from attack-triggered respawn validation to a real server-side creature respawn scheduler.



The current TryRespawnDue gate is accepted as a validated bridge.



The next safe step is a minimal scheduler loop with tightly controlled scope.
