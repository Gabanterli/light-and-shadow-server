\# R2-R-F â€” AOI/Client Creature Respawn Sync Requirements



\## Status



This document defines the requirements for synchronizing creature respawn events from the authoritative backend to the Godot client.



R2-R-D validated that the backend scheduler can respawn Orc\_Elite without requiring a player attack.



R2-R-E documented the runtime validation and confirmed one known limitation:



\* Backend respawn works.

\* Godot debug marker does not automatically return to alive/red.

\* Client visual state still requires interaction/debug flow to update.



R2-R-F defines the contract needed before implementing a creature respawn visual sync packet.



\## Current Runtime State



The backend currently supports:



\* CreatureSpawnManager.RegisterSpawn

\* CreatureSpawnManager.MarkDead

\* CreatureSpawnManager.MarkLootGenerated

\* CreatureSpawnManager.TryRespawnDue

\* CreatureSpawnManager.ReviveRespawn

\* DamageContributors reset on respawn

\* LootGenerated reset on respawn

\* RuntimeEntityID increment on respawn

\* Minimal gateway-side respawn scheduler



Validated runtime entity progression:



\* creature:debug\_orc\_elite\_001:1

\* creature:debug\_orc\_elite\_001:2

\* creature:debug\_orc\_elite\_001:3



\## Current Client Limitation



The Godot client currently knows how to mark Orc\_Elite as dead through:



\* SC\_TARGET\_DEAD

\* opcode 3003



However, there is no matching authoritative packet that tells the client:



\* the creature respawned;

\* the creature is alive again;

\* the marker should return to alive/red;

\* the old dead visual state is stale;

\* the new RuntimeEntityID is active.



Because of this, the backend scheduler can respawn the creature correctly while the client debug marker remains visually dead.



This is expected for the current R2 state.



\## Required Final Direction



Creature respawn must be synchronized through a server-authoritative AOI/client event.



The client must never infer respawn from:



\* clicking attack;

\* local timer;

\* local animation;

\* local cached state;

\* previous target ID only.



The server must explicitly inform the client that a creature runtime entity is alive and visible.



\## Required Packet Semantics



The project needs a creature respawn/spawn sync event.



Possible naming options:



\* SC\_CREATURE\_SPAWN

\* SC\_CREATURE\_RESPAWN

\* SC\_CREATURE\_STATE\_SYNC

\* SC\_TARGET\_ALIVE



Recommended R2 bridge name:



\* SC\_CREATURE\_RESPAWN



Reason:



\* The current task is specifically about respawn.

\* It avoids over-designing a full spawn system before AOI creature lifecycle is complete.

\* It clearly pairs with scheduler-driven respawn validation.



\## Recommended Temporary Opcode



A future opcode should be reserved for debug respawn sync.



Suggested temporary R2 opcode:



\* 3004: SC\_CREATURE\_RESPAWN



Current related opcodes:



\* 3000: CS\_ATTACK\_REQUEST

\* 3002: SC\_DAMAGE\_EVENT

\* 3003: SC\_TARGET\_DEAD

\* 4001: SC\_INVENTORY\_SYNC



Final opcode assignment can change later when the combat/creature protocol is consolidated.



\## Minimum Payload Requirements



The respawn packet should include:



\* RuntimeEntityID

\* SpawnID

\* CreatureID

\* DebugTargetID

\* X

\* Y

\* Z

\* CurrentHP

\* MaxHP

\* Alive

\* Version



For the current Orc\_Elite bridge:



\* RuntimeEntityID: creature:debug\_orc\_elite\_001:{version}

\* SpawnID: debug\_orc\_elite\_001

\* CreatureID: orc\_elite

\* DebugTargetID: Orc\_Elite



\## Payload Compatibility Requirements



Because the client still attacks using the debug target ID:



\* Orc\_Elite



the packet must include a temporary compatibility field:



\* DebugTargetID



This allows the Godot debug view to update the existing Orc\_Elite marker without requiring immediate RuntimeEntityID targeting.



Future target routing should use:



\* RuntimeEntityID



\## Backend Emit Requirements



The backend should emit the respawn packet when a creature respawns through the scheduler.



Current scheduler log:



\* Orc Elite creature spawn scheduler respawned



After this event, backend should send a respawn packet to nearby clients.



In the current R2 bridge, acceptable simplified behavior:



\* send directly to active player connections if AOI-specific creature visibility is not yet implemented;

\* or broadcast through existing AOIManager if a suitable method exists.



Final MMO behavior:



\* send only to players whose AOI contains the respawned creature.



\## AOI Requirements



AOI filtering must eventually use creature position and player visibility range.



A client should receive creature respawn only if:



\* player is connected;

\* player is in same world/shard;

\* player is on same Z/floor;

\* creature is within AOI range;

\* creature is not hidden by instancing/dungeon ownership rules;

\* region rules allow the creature to be visible.



The current debug bridge may skip full AOI filtering only if documented as temporary.



\## Godot Client Requirements



When Godot receives the respawn packet, it must:



\* find the creature marker by DebugTargetID or RuntimeEntityID;

\* mark the creature as alive;

\* clear local dead visual state;

\* set marker color back to alive/red;

\* update stored RuntimeEntityID if available;

\* update HP state if displayed;

\* avoid generating local loot or local combat state;

\* avoid trusting local timers.



For the current debug view:



\* Orc\_Elite should return from gray/black to red automatically.



\## Stale State Requirements



The client must not keep attacking stale runtime entities.



When a respawn packet arrives:



\* old RuntimeEntityID must be considered invalid;

\* new RuntimeEntityID becomes the active creature identity;

\* any future target selection should prefer the new RuntimeEntityID.



During the temporary debug bridge:



\* Orc\_Elite remains accepted as target ID;

\* backend remains authoritative and maps it to the current debug spawn lifecycle.



\## Death/Respawn Visual State Contract



Current death event:



\* SC\_TARGET\_DEAD marks target dead visually.



Future respawn event:



\* SC\_CREATURE\_RESPAWN marks target alive visually.



Expected debug visual states:



\* Alive Orc\_Elite: red marker

\* Dead Orc\_Elite: gray/black marker

\* Respawned Orc\_Elite: red marker again



\## Anti-Exploit Requirements



The client must not be able to request or force respawn.



Respawn sync is server-to-client only.



The client must not send:



\* CS\_RESPAWN\_CREATURE

\* local respawn confirm

\* client-side creature alive override



The backend must remain authoritative for:



\* respawn timing;

\* RuntimeEntityID;

\* HP;

\* position;

\* loot eligibility;

\* death state.



\## Loot Safety Requirements



Respawn visual sync must not reset corpse or loot ownership.



Loot from old RuntimeEntityID remains tied to the dead lifecycle.



Example:



\* loot from creature:debug\_orc\_elite\_001:1 must not become loot from creature:debug\_orc\_elite\_001:2.



Respawn packet must not grant loot.



Respawn packet must not clear inventory.



Respawn packet must not modify corpse ownership.



\## Logging Requirements



Backend logs should include:



\* Creature respawn packet emitted

\* RuntimeEntityID

\* SpawnID

\* CreatureID

\* DebugTargetID

\* AOI recipient count if available



Client logs should include:



\* Creature respawn packet received

\* RuntimeEntityID

\* DebugTargetID

\* visual marker reset to alive



\## Out of Scope for First Implementation



Do not implement in the first visual sync task:



\* full AOI creature spawn system;

\* corpse containers;

\* real loot tables;

\* party loot;

\* RuntimeEntityID-only targeting;

\* multi-creature config migration;

\* persistence of creature runtime state;

\* client creature list UI;

\* creature AI/pathfinding.



\## Recommended First Implementation



Recommended next engineering task:



\* R2-R-G â€” Add debug creature respawn visual sync packet



Minimum safe scope:



\* Add temporary SC\_CREATURE\_RESPAWN opcode.

\* Encode backend packet with minimal fields.

\* Emit packet when Orc\_Elite scheduler respawns.

\* Handle packet in Godot debug incoming router.

\* Set Orc\_Elite marker back to alive/red.

\* Keep attack target ID as Orc\_Elite for now.

\* Keep scheduler and retry debug unchanged.

\* Build backend.

\* Build Godot.

\* Runtime validate login -> kill -> wait 30s -> scheduler respawn -> marker turns red without clicking.



\## Runtime Validation Criteria



The implementation is successful if:



\* Orc\_Elite dies and marker turns gray/black.

\* Scheduler respawns Orc\_Elite after NextRespawn.

\* Backend emits creature respawn sync packet.

\* Godot receives the packet.

\* Orc\_Elite marker turns red automatically.

\* Player does not need to click or attack to refresh visual state.

\* Subsequent attack records damage against the new RuntimeEntityID.



\## Closure



R2-R-F defines the AOI/client creature respawn sync requirements.



The backend scheduler is already validated.



The next safe engineering step is a temporary debug respawn visual sync packet that updates the Godot marker automatically while preserving server authority and avoiding premature full AOI creature lifecycle implementation.
