# Alpha Runtime Checkpoint A6 - Movement Cadence + PvE Loop

## Status

PASS WITH MINOR WARNING

## Purpose

Validate the current Alpha Technical runtime loop after the client movement cadence guard fix.

Validated loop:

login -> character selection -> world entry -> movement -> visible creature -> attack -> damage -> death -> reward -> respawn -> post-respawn attack -> persistence

## Environment

- Branch: master
- Backend service: gateway-server
- Client: Godot 4 + C#
- Backend: Go
- Persistence: PostgreSQL
- Session: Redis
- Runtime user: default_user
- Runtime character: Gabriela

## Build Validation

Backend build:

docker compose build gateway-server

Result:

docker compose build gateway-server exit code: 0

Docker runtime:

docker compose up -d gateway-server
docker compose ps

Result:

docker compose up -d gateway-server exit code: 0
docker compose ps exit code: 0

Client build:

dotnet build LightAndShadow.sln

Result:

Compilacao com exito.
0 Aviso(s)
0 Erro(s)

## Runtime Validation Summary

Validated successfully:

- Login accepted for default_user.
- Character list request worked.
- Character Gabriela selected successfully.
- Character and inventory loaded from PostgreSQL.
- Inventory loaded with items_count=30.
- Initial gold loaded as 1050.
- Player registered in AOIManager.
- Initial player position sent to client.
- Initial chunks streamed to client.
- Orc Elite debug spawn registered.
- Initial Orc Elite RuntimeEntityID was creature:debug_orc_elite_001:1.
- Basic attack packets were received.
- Orc Elite damage contributions were recorded.
- Orc Elite death was detected.
- Target dead opcode 3003 was sent to the client.
- Debug loot fallback gold was granted after target death.
- First fallback reward persisted gold from 1050 to 1075.
- Orc Elite respawn scheduler emitted new spawn.
- Respawn visual sync opcode 3004 was emitted.
- Post-respawn RuntimeEntityID was creature:debug_orc_elite_001:2.
- Post-respawn attack resolved by RuntimeEntityID.
- Second Orc Elite death was detected.
- Second debug loot fallback gold was granted.
- Logout persistence saved final gold as 1100.
- Player state cleanup ran on disconnect.
- Final git status was clean.

## Runtime Evidence

Initial character load:

Character and inventory loaded successfully from PostgreSQL
player=Gabriela
items_count=30
gold=1050

Initial creature spawn:

Registered Orc Elite creature spawn state
runtime_entity_id=creature:debug_orc_elite_001:1

First death and reward:

Marked Orc Elite creature spawn dead
runtime_entity_id=creature:debug_orc_elite_001:1

Target dead packet sent to client
opcode=3003

Debug loot fallback gold granted after target death
fallback_gold=25

First persistence:

Character state, version, gold and inventory successfully persisted to PostgreSQL
player=Gabriela
gold=1075

Respawn:

Orc Elite creature spawn scheduler respawned
runtime_entity_id=creature:debug_orc_elite_001:2

Orc Elite creature respawn visual sync packet emitted
opcode=3004

Post-respawn attack:

Resolved debug Orc Elite attack target by runtime entity id
requested_target=creature:debug_orc_elite_001:2
resolved_target=Orc_Elite

Second death and reward:

Marked Orc Elite creature spawn dead
runtime_entity_id=creature:debug_orc_elite_001:2

Debug loot fallback gold granted after target death
fallback_gold=25

Final persistence:

Character state, version, gold and inventory successfully persisted to PostgreSQL
player=Gabriela
gold=1100

## Movement Cadence Result

The client movement cadence guard was validated as mostly stable.

Observed result:

- Normal gameplay continued.
- Movement did not block combat.
- Movement did not block reward.
- Movement did not block respawn.
- Movement did not block post-respawn targeting.
- No repeated movement validation spam was observed in the submitted logs.

## Minor Warning

One isolated backend movement validation warning occurred:

Authoritative movement validation failed (Client out of sync/rubberbanded)
player=Gabriela
requested_x=165
requested_y=154
confirmed_x=164
confirmed_y=154

Assessment:

- This does not block A6.
- It was not observed as repeated spam.
- It should remain monitored during future Alpha client movement work.
- No immediate patch is recommended before small Alpha UX improvements.

## Gold Persistence Result

Gold progression during checkpoint:

Initial gold: 1050
After first Orc Elite death: 1075
Final persisted gold after second death/logout: 1100

Result:

PASS

## RuntimeEntityID Result

RuntimeEntityID progression during checkpoint:

Initial spawn: creature:debug_orc_elite_001:1
Post-respawn spawn: creature:debug_orc_elite_001:2
Next observed respawn: creature:debug_orc_elite_001:3

Result:

PASS

## Final Decision

A6 is accepted as:

PASS WITH MINOR WARNING

The Alpha Technical loop remains valid after the movement cadence guard fix.

## Follow-up

Recommended next task:

A8 - Add small Alpha-style reward feedback

Reason:

The backend loop is stable enough to begin small, safe Alpha Mode UX improvements without replacing or deleting the Debug Client.

Do not start PvP, classes, quests, economy expansion, or large content systems before Alpha client readability improves.
