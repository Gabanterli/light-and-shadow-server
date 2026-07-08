# Alpha UI Layout Contract

## Status

Accepted for Alpha Technical UI planning.

## Purpose

Define the first playable Alpha UI direction for Light and Shadow without replacing or weakening the Debug Client.

The Alpha UI must feel like a minimal playable MMORPG client while still using real backend state.

The Debug Client remains available for protocol validation, packet inspection, RuntimeEntityID checks, movement debugging, reward validation, and raw technical diagnostics.

## Design Goal

The Alpha UI is not the final MMORPG HUD.

It is a playable technical Alpha shell focused on readability, real backend state, and a stable core loop.

Primary validated loop:

login -> character selection -> world entry -> movement -> visible creature -> attack -> damage -> death -> reward feedback -> respawn -> post-respawn attack -> persistence

## Layout Direction

The Alpha screen should use the following layout:

Top bar:

- Small height.
- Player frame.
- Character name.
- Level.
- HP.
- Mana.
- Basic connection/world status.

Main center:

- Large World View.
- Player visible.
- Creature visible.
- Tilemap readable.
- Combat feedback visible without covering the whole world.

Right side panel:

- Compact tabbed panel.
- Equipment tab.
- Backpack tab.
- Skills tab.
- Battle tab.
- Battle tab contains the small target frame.

Bottom panel:

- Discreet chat-style panel.
- Tabbed output.
- System tab.
- Combat tab.
- Action feedback such as attack sent, target dead, reward pending, reward sync confirmed, and respawn received.

## Visual Priority

Priority order:

1. World View
2. Player status
3. Target frame
4. Reward/combat feedback
5. Side panel tabs
6. Bottom chat/action feedback
7. Debug data only when explicitly needed

The World View must dominate the Alpha screen.

The packet log must not dominate the Alpha screen.

## Top Bar Contract

The top bar should show only minimal player state:

- Character name
- Level
- HP
- Mana
- Connection/world status

Do not show by default:

- AccountId
- IsAuthenticated
- IsCharacterSelected
- packet sequence
- raw opcodes
- chunk coordinates

Those remain Debug Client concerns.

## World View Contract

The World View should be the largest visible element.

It should show:

- Tilemap
- Player marker
- Creature marker
- Alive/dead creature state
- Future floating damage/reward feedback when available

The World View must be backed by real backend state.

No fake world state should be introduced in Alpha Mode.

## Side Panel Contract

The side panel should be compact and tabbed.

Initial tabs:

- Battle
- Backpack
- Equipment
- Skills

### Battle Tab

The Battle tab has highest priority for Alpha Technical.

It should show:

- Target name
- Target alive/dead state
- Optional target category
- Optional debug expansion in future

Do not show target HP unless HP is received from real backend state.

Do not show RuntimeEntityID by default in Alpha Mode.

RuntimeEntityID may be available later through a debug overlay or debug expansion.

### Backpack Tab

The Backpack tab can begin as a shell.

It should not invent item state.

It may initially show basic inventory sync state or item count only if backed by real inventory sync.

### Equipment Tab

The Equipment tab can begin as a shell.

It should not invent equipped items.

### Skills Tab

The Skills tab can begin as a shell.

It should not invent skill progression.

## Bottom Panel Contract

The bottom panel should behave like a small chat/action feedback area.

Initial tabs:

- System
- Combat

System examples:

- Connected
- World entered
- Respawn received
- Reward sync confirmed

Combat examples:

- Attack sent
- Damage received
- Target dead
- Reward pending
- Reward confirmed

The bottom panel must stay visually discreet.

It should not replace the Debug Client packet log.

## Debug Client Separation

The Debug Client remains intentionally technical.

Debug Client may continue showing:

- Raw packets
- Opcodes
- RuntimeEntityID
- coordinates
- movement validation state
- inventory sync boolean
- logs
- technical snapshot data

Alpha Client should not expose those by default.

## Alpha Mode Data Rule

Alpha UI must use real backend state.

Allowed:

- Real HP/Mana/Level from inventory/player sync.
- Real target dead event from opcode 3003.
- Real creature respawn event from opcode 3004.
- Real reward feedback confirmed by backend inventory/gold sync.
- Real movement confirmation.

Not allowed:

- Fake HP bars.
- Fake loot.
- Fake target state.
- Fake skill values.
- Fake equipment.
- Local-only reward simulation.

## Implementation Direction

Do not continue trying to reshape DebugWorldEntryScene into Alpha UI.

Recommended implementation path:

1. Keep DebugWorldEntryScene and DebugWorldEntryController for technical validation.
2. Add a separate AlphaWorldEntryScene.
3. Add a separate AlphaWorldEntryController.
4. Reuse networking/protocol code.
5. Reuse world view or extract shared presentation carefully.
6. Wire Alpha UI to real backend events incrementally.

## Recommended Task Breakdown

A10-R2 - Create AlphaWorldEntryScene shell

- Top bar placeholder.
- Large World View placeholder.
- Right tab panel placeholder.
- Bottom chat/action panel placeholder.
- No backend logic yet unless safely reused.

A10-R3 - Add top player frame baseline

- Character name.
- Level.
- HP.
- Mana.
- Real values only.

A10-R4 - Add large world view baseline

- Reuse or adapt existing debug world view.
- Keep world readable.
- No fake state.

A10-R5 - Add side panel tabs baseline

- Battle.
- Backpack.
- Equipment.
- Skills.
- Shell first.

A10-R6 - Add bottom action feedback tabs baseline

- System.
- Combat.
- Minimal feedback routing.

A10-R7 - Wire battle tab to real Orc Elite state

- Target name.
- Alive/dead.
- Respawn state.
- No fake HP.

## Current Decision

The previous idea of reducing debug log dominance directly inside DebugWorldEntryScene is deferred.

Reason:

The Debug Client should remain technical.

Alpha readability should be solved by creating a real Alpha UI shell instead of over-polishing the Debug Client.
