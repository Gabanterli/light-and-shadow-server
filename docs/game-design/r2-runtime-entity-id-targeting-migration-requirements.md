\# R2-R-I â€” RuntimeEntityID Targeting Migration Requirements



\## Status



This document defines the requirements for migrating creature targeting from the temporary debug target ID to the server-authoritative RuntimeEntityID.



Current debug target:



\* Orc\_Elite



Current runtime creature identity examples:



\* creature:debug\_orc\_elite\_001:1

\* creature:debug\_orc\_elite\_001:2

\* creature:debug\_orc\_elite\_001:3



\## Context



The R2 respawn path now supports:



\* creature spawn state registration;

\* death state tracking;

\* loot generation guard;

\* damage contribution tracking;

\* timer-based respawn;

\* scheduler-driven respawn;

\* visual respawn sync through temporary opcode 3004.



The Godot debug client can now visually mark Orc\_Elite as dead through opcode 3003 and alive again through opcode 3004.



However, attack targeting still uses the debug string:



\* Orc\_Elite



This is acceptable for the temporary debug bridge, but it is not safe for real MMO creature targeting.



\## Problem



A static debug target ID cannot uniquely identify a creature lifecycle.



The string Orc\_Elite describes a debug target name, not a specific spawned entity.



Once creatures can die and respawn, each lifecycle must have a unique identity.



Example:



\* creature:debug\_orc\_elite\_001:1 dies.

\* creature:debug\_orc\_elite\_001:2 respawns.

\* creature:debug\_orc\_elite\_001:1 must never be attackable again.

\* creature:debug\_orc\_elite\_001:2 becomes the valid target.



Without RuntimeEntityID targeting, the client can only say:



\* attack Orc\_Elite



The backend then has to infer which runtime entity is intended.



This does not scale to multiple creatures, multiple spawns, stale client state, or real loot ownership.



\## Migration Goal



The target identity used by combat requests must migrate from:



\* debug target ID



to:



\* RuntimeEntityID



The final target should be:



\* creature:debug\_orc\_elite\_001:{version}



or a generalized equivalent for every creature spawn.



\## Required Final Direction



The client must target the exact server-authoritative runtime entity.



The server must validate that:



\* the RuntimeEntityID exists;

\* the creature is alive;

\* the creature belongs to an active spawn state;

\* the RuntimeEntityID matches the current lifecycle version;

\* the creature is within range;

\* the attacking player is allowed to attack;

\* the target is visible or known through AOI;

\* stale RuntimeEntityID attacks are rejected.



\## Temporary Compatibility Requirement



During the migration, the backend may continue accepting:



\* Orc\_Elite



as a compatibility alias.



But RuntimeEntityID must become the preferred identity.



Recommended temporary resolution order:



1\. If target ID starts with `creature:`, resolve by RuntimeEntityID.

2\. Else if target ID equals `Orc\_Elite`, resolve through the debug Orc Elite bridge.

3\. Else reject as unknown target.



This allows step-by-step migration without breaking the current debug test flow.



\## Required Client State



The Godot client must store the active runtime identity for the debug creature.



Current local state:



\* `\_isOrcEliteDead`

\* `DebugTileWorldView.IsOrcEliteDead`

\* `OrcElitePosition`



Required additional local state:



\* `\_orcEliteRuntimeEntityId`



Example:



\* `\_orcEliteRuntimeEntityId = "creature:debug\_orc\_elite\_001:2"`



The client must update this value when receiving authoritative creature state packets.



\## Required Packet Evolution



The temporary respawn visual sync packet currently carries only:



\* Orc\_Elite



Future packet payload must include:



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



Minimum next safe step:



\* Add RuntimeEntityID to debug respawn visual sync packet.



This allows the client to update its active target identity without changing attack requests immediately.



\## Required Attack Request Evolution



Current attack request:



\* TargetID: Orc\_Elite

\* WeaponType: debug\_sword



Future attack request:



\* TargetID: creature:debug\_orc\_elite\_001:{version}

\* WeaponType: debug\_sword



The protocol field may remain named TargetID temporarily, but semantically it must become RuntimeEntityID-compatible.



\## Backend Combat Resolution Requirements



Backend attack handling must support two resolution paths during migration.



\### RuntimeEntityID Path



If TargetID is a RuntimeEntityID:



\* resolve spawn state by runtime entity;

\* confirm it is current;

\* confirm it is alive;

\* confirm CombatManager has matching alive stats;

\* process attack;

\* record contribution against that runtime entity.



\### Debug Alias Path



If TargetID is Orc\_Elite:



\* resolve current debug spawn;

\* use the current RuntimeEntityID internally;

\* process attack as today;

\* log that compatibility alias was used.



The debug alias path must be removed after client migration is complete.



\## Stale Target Protection



The backend must reject attacks against stale RuntimeEntityIDs.



Example stale attack:



\* client attacks creature:debug\_orc\_elite\_001:1

\* current active lifecycle is creature:debug\_orc\_elite\_001:2



Expected result:



\* reject attack;

\* do not apply damage;

\* do not record contribution;

\* do not trigger loot;

\* optionally send a target stale/dead response to the client.



Recommended log:



\* Creature attack rejected: stale runtime entity target



\## Death State Protection



The backend must reject RuntimeEntityID attacks when:



\* creature is dead;

\* creature is awaiting respawn;

\* loot was generated;

\* NextRespawn is pending.



This prevents double death and double loot.



\## Respawn State Protection



After respawn:



\* old RuntimeEntityID must remain invalid;

\* new RuntimeEntityID becomes active;

\* damage contributors reset;

\* loot guard resets;

\* HP resets;

\* client receives new RuntimeEntityID.



\## AOI Protection



Final target validation should require AOI awareness.



A player should not attack a creature if:



\* the creature is outside AOI;

\* the creature is on another floor;

\* the creature is not visible to that player;

\* the creature belongs to another instance;

\* the player has stale AOI state.



The current R2 debug bridge may skip full AOI validation, but this must remain temporary.



\## Client Visual Requirements



When the client receives a creature state or respawn packet with RuntimeEntityID:



\* update active runtime entity ID;

\* mark visual alive/dead according to server state;

\* do not infer respawn locally;

\* do not keep old RuntimeEntityID after respawn;

\* use RuntimeEntityID for future attack requests once enabled.



For Orc\_Elite debug bridge:



\* marker remains red when alive;

\* marker remains gray/black when dead;

\* active RuntimeEntityID changes on respawn.



\## Logging Requirements



Backend logs should include:



\* attack target received;

\* target resolution path;

\* RuntimeEntityID selected;

\* stale target rejection if any;

\* debug alias fallback usage;

\* contribution RuntimeEntityID.



Client logs should include:



\* RuntimeEntityID received;

\* active debug target RuntimeEntityID updated;

\* attack request target identity;

\* stale/invalid target response if implemented later.



\## Anti-Exploit Requirements



The client must not be trusted to invent RuntimeEntityIDs.



Backend must validate all RuntimeEntityIDs against authoritative spawn state.



The client cannot force:



\* respawn;

\* target lifecycle version;

\* HP reset;

\* loot eligibility;

\* contribution ownership;

\* creature visibility.



RuntimeEntityID is an identity reference, not an authority claim.



\## Loot and Contribution Requirements



Damage contribution must always be tied to the active RuntimeEntityID.



Loot eligibility must be computed against the lifecycle that died.



Example:



\* damage on creature:debug\_orc\_elite\_001:1 belongs only to lifecycle 1;

\* respawn creates creature:debug\_orc\_elite\_001:2;

\* lifecycle 2 must start with empty damage contributors;

\* loot from lifecycle 1 must not be claimable through lifecycle 2.



\## Migration Steps



Recommended implementation order:



1\. Add RuntimeEntityID to debug respawn visual sync packet.

2\. Store active RuntimeEntityID in Godot debug state.

3\. Log active RuntimeEntityID in Godot when respawn sync arrives.

4\. Continue sending Orc\_Elite as attack target temporarily.

5\. Add backend RuntimeEntityID target resolution support.

6\. Add client option to send RuntimeEntityID as attack target.

7\. Runtime validate attacks by RuntimeEntityID.

8\. Keep Orc\_Elite fallback temporarily.

9\. Add stale RuntimeEntityID rejection validation.

10\. Document closure.

11\. Remove or quarantine debug alias after generalized creature targeting exists.



\## Recommended Next Engineering Tasks



\### R2-R-J â€” Add RuntimeEntityID to debug respawn visual sync packet



Scope:



\* backend encodes RuntimeEntityID in opcode 3004 payload;

\* Godot decodes RuntimeEntityID;

\* Godot stores `\_orcEliteRuntimeEntityId`;

\* attack request still uses Orc\_Elite;

\* no backend attack resolution change yet.



\### R2-R-K â€” Store active RuntimeEntityID in Godot debug target state



Scope:



\* display/log RuntimeEntityID;

\* confirm value changes after respawn;

\* prepare attack request migration.



\### R2-R-L â€” Send attack requests using RuntimeEntityID with Orc\_Elite fallback



Scope:



\* client sends RuntimeEntityID when known;

\* fallback to Orc\_Elite only if no RuntimeEntityID is available;

\* backend initially may still need compatibility update.



\### R2-R-M â€” Backend resolves attacks by RuntimeEntityID first



Scope:



\* backend accepts RuntimeEntityID target;

\* validates it against current spawn state;

\* rejects stale target;

\* preserves debug fallback only as temporary bridge.



\## Closure Criteria



This migration is considered complete when:



\* client receives RuntimeEntityID from backend;

\* client stores active RuntimeEntityID;

\* client sends attacks using RuntimeEntityID;

\* backend resolves target by RuntimeEntityID;

\* stale RuntimeEntityID attack is rejected;

\* death/loot/contribution are tied to the correct lifecycle;

\* debug Orc\_Elite alias is no longer required for normal combat validation.



\## Final Decision



RuntimeEntityID targeting is required before real multi-creature PvE, real AOI creature sync, and real loot/corpse ownership.



The project should not advance into generalized creature spawning or real loot ownership while attack target identity remains based only on Orc\_Elite.



The safest path is to migrate targeting incrementally:



\* first enrich packets;

\* then store RuntimeEntityID in client;

\* then send it from client;

\* then validate it server-side;

\* then remove debug alias later.
