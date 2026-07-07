\# R2-R-E â€” Minimal Respawn Scheduler Runtime Validation Closure



\## Status



R2-R-D â€” Add minimal creature respawn scheduler loop is runtime-validated and closed.



This document records the successful runtime validation of the minimal server-side respawn scheduler for the Orc\_Elite debug spawn.



\## Covered Commit



\* 0b6f025 Add minimal Orc Elite respawn scheduler



\## Runtime Validation Summary



The minimal respawn scheduler was validated with the following flow:



login -> character list -> character select -> world entry -> attack Orc\_Elite -> death -> NextRespawn scheduled -> wait without attacking -> scheduler respawn -> new RuntimeEntityID -> attack again -> damage contribution on new runtime entity -> second death -> scheduler respawn again



\## Login Data Used



\* User: default\_user

\* Password: test123

\* Character: Gabriela



\## Build Validation



The gateway build was validated successfully before runtime testing.



Validated command:



\* docker compose build gateway-server



Validated result:



\* docker compose build gateway-server exit code: 0



\## Runtime Flow Validated



The backend registered the debug creature spawn:



\* SpawnID: debug\_orc\_elite\_001

\* CreatureID: orc\_elite

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:1



The player killed Orc\_Elite and the backend marked the creature dead.



Validated logs:



\* Marked Orc Elite creature spawn dead

\* Orc Elite debug loot generation guard accepted

\* Sending target dead packet to client

\* Target dead packet sent to client

\* Debug loot granted after target death



The backend scheduled the next respawn:



\* DiedAt: 2026-07-07T14:11:36.579651347Z

\* NextRespawn: 2026-07-07T14:12:06.579651347Z



After the respawn time passed, the scheduler respawned the creature without requiring a new attack packet.



Validated log:



\* Orc Elite creature spawn scheduler respawned



The runtime entity advanced to:



\* creature:debug\_orc\_elite\_001:2



After the scheduler respawn, damage contribution was recorded under the new runtime entity.



Validated log:



\* Recorded Orc Elite damage contribution



The player killed the second lifecycle successfully.



Second lifecycle:



\* RuntimeEntityID: creature:debug\_orc\_elite\_001:2

\* NextRespawn: 2026-07-07T14:13:35.15240938Z



The scheduler then respawned the creature again, even after the client disconnected.



Validated runtime entity:



\* creature:debug\_orc\_elite\_001:3



\## Result



R2-R-D passed runtime validation.



Confirmed behavior:



\* Scheduler loop starts with the gateway.

\* Scheduler checks respawn eligibility every second.

\* Scheduler skips alive creatures.

\* Scheduler waits until NextRespawn is due.

\* Scheduler calls TryRespawnDue.

\* TryRespawnDue advances RuntimeEntityID.

\* Scheduler syncs CombatManager by calling ReviveEntity.

\* Respawn happens without requiring a post-death attack.

\* Respawn still works after client disconnect.

\* DamageContributors resets between runtime entity versions.

\* LootGenerated resets between runtime entity versions.

\* Existing debug attack and loot flow remain compatible.



\## Important Client Limitation



The Godot debug marker did not automatically turn red after backend scheduler respawn.



This is expected for the current R2-R-D scope.



Reason:



The backend now respawns Orc\_Elite server-side, but the client does not yet receive a creature respawn or alive-state sync packet.



The client currently has a visual dead-state transition through:



\* SC\_TARGET\_DEAD

\* opcode 3003



But the project does not yet have the matching visual respawn packet, such as:



\* SC\_CREATURE\_SPAWN

\* SC\_CREATURE\_RESPAWN

\* SC\_TARGET\_ALIVE

\* AOI creature spawn update



Therefore, the backend state is correct, but the client debug visual remains stale until another interaction path causes the debug view to update.



This is not considered a failure of R2-R-D.



\## Remaining Risks



\### Missing AOI Respawn Broadcast



The scheduler respawns the creature in backend state, but nearby clients are not notified.



Future work must emit an AOI respawn/spawn packet to nearby players.



\### Client Visual State Not Synced



The Godot debug view does not automatically switch Orc\_Elite back to alive/red after scheduler respawn.



Future client work must consume a server-authoritative respawn packet and update the visual marker.



\### Debug Target ID Still Used



The client still attacks:



\* Orc\_Elite



Future combat protocol should target:



\* RuntimeEntityID



Example:



\* creature:debug\_orc\_elite\_001:2



\### Retry Debug Still Exists



The debug retry fallback still exists.



It should remain until:



\* scheduler respawn is stable;

\* AOI respawn broadcast exists;

\* client visual alive-state sync exists;

\* RuntimeEntityID targeting is available or safely bridged.



\### Loot Still Debug-Only



Loot is still debug-only:



\* sword\_t1\_rusty x1



Real loot tables, corpse containers, loot ownership, party rules, and contribution-based eligibility remain future work.



\### Scheduler Location Is Temporary



The scheduler currently runs gateway-side.



This is acceptable for R2 validation, but final MMO architecture should move creature lifecycle scheduling into the authoritative world lifecycle layer.



\## Architectural Decision



R2-R-D establishes the first working server-side creature respawn scheduler bridge.



This scheduler is accepted as an intermediate R2 implementation.



It proves that respawn no longer depends on a player attack to trigger TryRespawnDue.



However, it is not yet the final respawn architecture because it lacks AOI spawn broadcast, client visual sync, corpse integration, persistence, and RuntimeEntityID protocol routing.



\## Closure



R2-R-D is closed.



R2-R-E documents the runtime validation and known client sync limitation.



Recommended next task:



\* R2-R-F â€” Define AOI/client creature respawn sync requirements



Alternative next engineering task:



\* R2-R-F â€” Add debug creature respawn visual sync packet



Architectural recommendation:



Define AOI/client respawn sync requirements before implementing the packet, because respawn visual sync touches protocol, AOI, client state, stale RuntimeEntityID protection, and future multi-creature support.
