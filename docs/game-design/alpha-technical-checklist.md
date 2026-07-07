# Alpha Technical Checklist

## Status

Draft baseline for technical Alpha.

## Goal

Define the minimum technical requirements for the first internal Alpha of Light and Shadow.

This Alpha is not a content Alpha, marketing Alpha, or public test.

It is a technical validation milestone for the core MMO loop.

## Required Alpha Core Loop

The Alpha Technical build must support:

- Login.
- Character selection.
- World entry.
- Authoritative movement.
- Initial AOI/chunk streaming.
- Inventory sync.
- Basic combat.
- Creature spawn.
- Creature targeting.
- Creature damage.
- Creature death.
- Reward delivery.
- Creature respawn.
- Post-respawn targeting.
- Persistence after autosave/logout.

## Current Confirmed Systems

Confirmed working:

- Auth login with default_user.
- Character selection with Gabriela.
- World entry.
- PostgreSQL character load.
- Inventory load and sync.
- Gold persistence.
- AOI manager player registration.
- Initial chunk streaming.
- Debug Orc Elite spawn state.
- RuntimeEntityID assignment.
- Basic attack packet flow.
- Damage contribution tracking.
- Target death sync opcode 3003.
- Creature respawn sync opcode 3004.
- Post-respawn RuntimeEntityID attack targeting.
- Stale RuntimeEntityID target guard.
- Debug loot fallback gold reward.
- Autosave persistence.

## Minimum Alpha Pass Criteria

The build can be called Alpha Technical only when all items below pass in one runtime session:

- Player logs in without manual backend edits.
- Player selects existing character.
- Player enters the world.
- Player moves without repeated desync loops.
- Client receives initial inventory.
- Client receives initial world state.
- Player can target a creature.
- Player can attack a creature.
- Backend validates damage authoritatively.
- Creature death is visible to client.
- Reward is granted or safely fallbacked.
- Reward persists.
- Creature respawns.
- Client receives respawn sync.
- Client attacks respawned creature using RuntimeEntityID.
- Backend rejects stale RuntimeEntityID attacks safely.
- No duplicate death packet.
- No duplicate loot grant.
- No crash on disconnect.
- Character state saves successfully.

## Known Non-Goals For Technical Alpha

Not required yet:

- Full class system UI.
- Real questline progression.
- Full corpse loot container.
- Multi-creature combat at scale.
- Final UI polish.
- Public server deployment.
- Real economy balancing.
- Party loot distribution.
- PvP Alpha.
- Full map content.

## Required Validation Before Alpha Tag

Required local validations:

- dotnet build LightAndShadow.sln exit code 0.
- docker compose build gateway-server exit code 0.
- Runtime login test.
- Runtime movement test.
- Runtime combat/death/respawn test.
- Runtime reward persistence test.
- Final git status clean.

## Current Alpha Readiness Estimate

Current technical Alpha readiness:

- Backend core loop: high.
- Client debug loop: medium.
- PvE lifecycle: medium-high.
- Reward loop: medium.
- Movement polish: medium.
- Generic creature support: low-medium.
- Alpha packaging/readiness: low.

Overall status:

Approximately 55 percent ready for Technical Alpha.