\# R2-R-H â€” Debug Respawn Visual Sync Validation Closure



\## Status



R2-R-G â€” Add debug creature respawn visual sync packet is runtime-validated and closed.



This document records the successful validation of the temporary debug respawn visual sync bridge for Orc\_Elite.



\## Covered Commit



\* feb98a0 Add debug creature respawn visual sync



\## Previous Required Foundation



This task depends on the following completed work:



\* R2-R-D â€” Add minimal creature respawn scheduler loop

\* R2-R-E â€” Document minimal respawn scheduler runtime validation closure

\* R2-R-F â€” Define AOI/client creature respawn sync requirements

\* R2-R-G â€” Add debug creature respawn visual sync packet



\## Goal



The goal of R2-R-G was to remove the need for a player click or attack to visually refresh Orc\_Elite after backend scheduler respawn.



Before this task:



\* backend scheduler respawned Orc\_Elite correctly;

\* RuntimeEntityID advanced correctly;

\* damage contribution reset correctly;

\* loot generation reset correctly;

\* Godot debug marker stayed visually dead until another interaction path refreshed it.



After this task:



\* backend emits a temporary respawn visual sync packet;

\* Godot receives opcode 3004;

\* Godot resets Orc\_Elite visual state to alive/red;

\* subsequent attacks continue against the new backend creature lifecycle.



\## Runtime Validation Summary



Validated runtime flow:



login -> character select -> world entry -> attack Orc\_Elite -> death -> marker dead state -> scheduler wait -> backend respawn -> opcode 3004 visual sync emitted -> client visual state can reset to alive/red -> second attack -> damage recorded against new RuntimeEntityID



\## Login Data Used



\* User: default\_user

\* Password: test123

\* Character: Gabriela



\## Backend Validation Evidence



Initial creature lifecycle:



\* SpawnID: debug\_orc\_elite\_001

\* CreatureID: orc\_elite

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:1



Validated death:



\* Marked Orc Elite creature spawn dead

\* Orc Elite debug loot generation guard accepted

\* Sending target dead packet to client

\* Target dead packet sent to client

\* Debug loot granted after target death



Validated first scheduled respawn:



\* Orc Elite creature spawn scheduler respawned

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:2



Validated new visual sync event:



\* Orc Elite creature respawn visual sync packet emitted

\* Target: Orc\_Elite

\* Opcode: 3004

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:2



Validated second lifecycle damage:



\* Recorded Orc Elite damage contribution

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:2



Validated second lifecycle death:



\* Marked Orc Elite creature spawn dead

\* RuntimeEntityID: creature:debug\_orc\_elite\_001:2

\* NextRespawn scheduled correctly



\## Backend Behavior Added



When the minimal respawn scheduler successfully respawns Orc\_Elite, the gateway now emits a temporary visual sync packet:



\* Opcode: 3004

\* Meaning: SC\_CREATURE\_RESPAWN debug bridge

\* Payload: Orc\_Elite target identifier



The packet is emitted after:



\* TryRespawnDue succeeds;

\* CombatManager.ReviveEntity succeeds;

\* RuntimeEntityID advances.



\## Godot Behavior Added



The Godot debug world entry controller now registers a handler for:



\* Opcode: 3004



The handler:



\* decodes the payload using the existing target string decoder;

\* identifies the target as Orc\_Elite;

\* sets local Orc\_Elite dead state to false;

\* updates DebugTileWorldView.IsOrcEliteDead to false;

\* queues redraw;

\* logs that Orc\_Elite visual state was reset to alive/red.



Expected client-side log:



\* \[RECV] Opcode: 3004 (SC\_CREATURE\_RESPAWN)

\* Creature Respawn: Orc\_Elite

\* Orc\_Elite visual state reset to alive/red.



\## Accepted Temporary Design



The implementation intentionally reuses a minimal target string payload.



This is acceptable for the R2 debug bridge because:



\* only Orc\_Elite is involved;

\* the goal is visual validation, not final protocol design;

\* RuntimeEntityID routing is not yet implemented on the client;

\* AOI real creature lifecycle is not yet implemented;

\* the previous target-dead visual path already uses target ID compatibility.



\## Known Temporary Limitations



\### Debug Target ID Still Used



The client still recognizes:



\* Orc\_Elite



as the visual and attack target.



Future architecture must migrate target identity to:



\* RuntimeEntityID



Example:



\* creature:debug\_orc\_elite\_001:2



\### Opcode 3004 Is Temporary



Opcode 3004 is a temporary R2 bridge.



Final protocol should consolidate creature state events into a stable creature lifecycle protocol.



Possible future events:



\* SC\_CREATURE\_SPAWN

\* SC\_CREATURE\_DESPAWN

\* SC\_CREATURE\_RESPAWN

\* SC\_CREATURE\_STATE\_SYNC

\* SC\_CREATURE\_HEALTH\_SYNC



\### No Real AOI Broadcast Yet



The current bridge is not the final AOI creature spawn system.



Final MMO behavior must send creature respawn only to players whose AOI includes the creature.



\### No Multi-Creature Support Yet



The current validation covers:



\* debug\_orc\_elite\_001



Future work must generalize to all configured creature spawns.



\### No RuntimeEntityID Client Targeting Yet



The backend lifecycle now tracks RuntimeEntityID.



The client does not yet use RuntimeEntityID as the target identity.



This means stale target prevention is still incomplete from the client protocol perspective.



\### No Corpse/Loot Integration Yet



The respawn visual packet does not:



\* create corpse containers;

\* manage corpse visibility;

\* assign loot ownership;

\* clear old corpse state;

\* enforce contribution-based loot rights.



This remains part of the future real loot/corpse system.



\## Anti-Exploit Notes



The respawn visual sync is server-to-client only.



The client cannot force:



\* creature respawn;

\* RuntimeEntityID advancement;

\* HP reset;

\* loot reset;

\* visual authority on the server.



The server remains authoritative for:



\* death state;

\* respawn timer;

\* creature HP;

\* RuntimeEntityID;

\* loot generation guard;

\* damage contribution state.



\## Closure Result



R2-R-G is accepted as a successful temporary debug bridge.



The backend now has:



\* real death state;

\* loot guard;

\* damage contributors;

\* timer-based respawn;

\* scheduler-driven respawn;

\* visual respawn packet emission.



The Godot debug client now has:



\* dead visual state through opcode 3003;

\* alive visual reset through opcode 3004.



\## Remaining Architecture Path



Recommended next documentation task:



\* R2-R-I â€” Define RuntimeEntityID targeting migration requirements



Alternative next gameplay/system task:



\* R2-L-A â€” Define real loot table and corpse ownership architecture



Recommended order:



1\. Define RuntimeEntityID targeting migration.

2\. Define real loot/corpse ownership.

3\. Implement RuntimeEntityID-safe targeting.

4\. Replace debug Orc\_Elite bridge with generalized creature lifecycle protocol.

5\. Implement real AOI creature spawn/despawn/respawn sync.

6\. Implement real loot table and corpse container flow.



\## Final Closure



R2-R-H documents the successful closure of the debug respawn visual sync bridge.



The R2 respawn path is now validated from backend lifecycle to Godot debug visual update.



The implementation remains intentionally temporary and debug-scoped, but it removes the previous manual visual refresh gap and prepares the project for RuntimeEntityID-safe creature lifecycle work.
