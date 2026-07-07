\# R2-R-B — Orc Elite Respawn Timer Runtime Validation Closure



\## Status



R2-R-A — Add respawn timer to spawn state is runtime-validated.



This document closes the first runtime validation of server-authoritative creature respawn timing for the R2 real PvE foundation.



\## Covered Commit



\- 730ceb5 Add Orc Elite respawn timer gate



\## Runtime Flow Validated



Validated flow:



login -> character list -> character select -> world entry -> attack Orc\_Elite -> damage contribution -> death -> NextRespawn set -> debug loot -> wait more than 30 seconds -> attack Orc\_Elite again -> timer respawn gate accepted -> new RuntimeEntityID -> damage contribution reset -> second kill -> second loot



\## Login Data Used



\- User: default\_user

\- Password: test123

\- Character: Gabriela



\## First Spawn Lifecycle



Initial registered spawn:



\- SpawnID: debug\_orc\_elite\_001

\- CreatureID: orc\_elite

\- RuntimeEntityID: creature:debug\_orc\_elite\_001:1



Validated log:



\- Registered Orc Elite creature spawn state



First lifecycle damage contribution accumulated under:



\- creature:debug\_orc\_elite\_001:1



Observed contribution total before death:



\- 84.35000000000001



First death was recorded at:



\- 2026-07-07T13:54:08.606962341Z



Next respawn was scheduled for:



\- 2026-07-07T13:54:38.606962341Z



Validated logs:



\- Marked Orc Elite creature spawn dead

\- Orc Elite debug loot generation guard accepted

\- Sending target dead packet to client

\- Target dead packet sent to client

\- Debug loot granted after target death



Inventory sync after first loot:



\- items\_count: 16 -> 17



\## Timer Respawn Validation



After waiting more than 30 seconds, the next attack triggered the timer respawn path.



Validated log:



\- Orc Elite creature spawn timer respawned



Respawned runtime entity:



\- creature:debug\_orc\_elite\_001:2



Respawn timestamp:



\- 2026-07-07T13:55:41.882264639Z



This confirms that TryRespawnDue accepted the due respawn window and advanced the creature lifecycle into a new alive runtime entity.



\## Second Spawn Lifecycle



After timer respawn, damage contribution reset and began accumulating under:



\- creature:debug\_orc\_elite\_001:2



Observed contribution accumulation:



\- 16.21

\- 30.76

\- 44.42

\- 65.21000000000001

\- 81.41000000000001



Second death was recorded at:



\- 2026-07-07T13:55:51.534454278Z



Next respawn was scheduled for:



\- 2026-07-07T13:56:21.534454278Z



Validated logs:



\- Marked Orc Elite creature spawn dead

\- Orc Elite debug loot generation guard accepted

\- Target dead packet sent to client

\- Debug loot granted after target death



Inventory sync after second loot:



\- items\_count: 17 -> 18



\## Runtime Result



The runtime validation passed.



Confirmed behavior:



\- MarkDead schedules NextRespawn.

\- TryRespawnDue rejects early respawn implicitly by requiring the due timestamp.

\- TryRespawnDue accepts respawn after the due timestamp.

\- RuntimeEntityID advances from version 1 to version 2.

\- DamageContributors resets on respawn.

\- LootGenerated resets on respawn.

\- Debug loot guard still works per lifecycle.

\- Inventory sync still works.

\- Existing debug attack flow remains compatible.



\## Non-Blocking Observations



Some movement rubberband warnings still appeared during validation.



These warnings are not considered blockers for R2-R-B because the combat, death, respawn timer, damage contribution, loot guard, and inventory sync flow completed successfully.



Observed range rejection also occurred once:



\- alvo fora de alcance para Espada



This is expected server-authoritative combat validation and is not a respawn failure.



\## Remaining Risks



\### Respawn Trigger Path



The respawn timer is currently checked during the Orc\_Elite attack path.



This validates the timer gate, but it is not the final MMO respawn architecture.



Future work should move respawn checks into a server-side world lifecycle scheduler.



\### AOI Spawn Broadcast Missing



The client does not yet receive a real creature spawn/despawn AOI event for the respawned runtime entity.



Future work must broadcast creature spawn state to nearby players.



\### Debug Retry Still Present



The debug retry fallback still exists.



This is acceptable temporarily, but it must be removed once the respawn scheduler and AOI spawn broadcast are stable.



\### Debug Target ID Still Present



The client still attacks using:



\- Orc\_Elite



Future protocol work must route attacks through RuntimeEntityID.



\### Loot Still Debug-Only



Loot remains fixed debug loot:



\- sword\_t1\_rusty x1



Real loot tables, corpse containers, ownership, and contribution-based eligibility are still not implemented.



\## Architectural Decision



R2-R-A successfully introduced a server-authoritative respawn timer gate into CreatureSpawnManager.



The current implementation is accepted as an intermediate R2 bridge.



The final MMO architecture should move from attack-triggered respawn checks to a world lifecycle scheduler that periodically evaluates dead spawns and emits AOI spawn events.



\## Closure



R2-R-B is closed.



Recommended next task:



\- R2-R-C — Add documentation for remaining real respawn scheduler requirements



Alternative next engineering task:



\- R2-R-C — Add creature respawn scheduler loop



Architectural recommendation:



Proceed with a documentation closure for the remaining scheduler requirements before implementing a broader respawn loop, because scheduler-driven respawn affects AOI, combat targeting, corpse lifetime, and future persistence boundaries.

