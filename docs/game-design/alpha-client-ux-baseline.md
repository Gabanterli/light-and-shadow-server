# A4 - Alpha Client UX Baseline

## Status

Draft baseline.

## Purpose

Define the minimum client experience required for the Technical Alpha of Light and Shadow.

This is not final UI, final art, or final gameplay polish.

The goal is to transform the current debug-world client into a minimum playable Alpha client while keeping the real MMO backend foundation intact.

## Context

The current client is intentionally technical and debug-oriented.

It validates real systems:

- login
- character selection
- world entry
- chunk streaming
- inventory sync
- movement requests
- movement confirmation
- combat packets
- target death
- creature respawn
- RuntimeEntityID targeting
- reward sync
- autosave persistence

However, the current presentation still feels less structured than the earlier simulator because the earlier simulator was more visual and less constrained by real server authority.

Alpha Client UX must recover structure and readability without reverting to mock-only behavior.

## UX Goal

The Alpha client must make the core loop understandable without requiring the developer to read backend logs for every action.

The player should clearly understand:

- where the character is
- where the target creature is
- whether the target is alive or dead
- when movement is accepted or delayed
- when damage happens
- when loot or gold is received
- when respawn happens
- whether the client is connected

## Alpha Screen Layout

Minimum recommended layout:

- game view in the center
- player marker or sprite centered/clear
- visible tile area around the player
- creature marker or sprite visible in world space
- target frame near the top or side
- small combat/reward feedback area
- compact connection/status area
- debug log collapsed or reduced

The debug log must not dominate the screen during Alpha gameplay.

## Player Presentation

Minimum Alpha requirement:

- player must be clearly visible
- player position must update after server confirmation
- player should appear stable even when movement is server-authoritative
- movement rejection should not feel like a crash or bug

Optional later polish:

- walking animation
- directional facing
- movement smoothing
- camera follow

## Creature Presentation

Minimum Alpha requirement:

- creature must be visible as a world entity
- creature must expose current alive/dead state
- creature must update after respawn
- creature identity should eventually come from RuntimeEntityID

Current debug Orc Elite marker is acceptable only as a temporary Alpha bridge.

## Target Frame

Minimum Alpha target frame:

- target name
- alive/dead state
- approximate HP state or debug HP value when available
- RuntimeEntityID visible only in debug mode
- selected target feedback

The player should not need to inspect logs to know if the target died.

## Combat Feedback

Minimum Alpha combat feedback:

- attack action feedback
- damage event feedback
- target death feedback
- failed attack feedback when target is dead/stale
- cooldown or pending state feedback

Damage numbers are preferred but not mandatory for first Alpha.

## Reward Feedback

Minimum Alpha reward feedback:

- visible reward notification
- gold gain display
- inventory sync confirmation
- fallback reward should be shown as gold reward, not silent backend log

Example:

    +25 gold

The player should not need backend logs to know a reward was granted.

## Movement Feedback

Minimum Alpha movement feedback:

- show movement pending state subtly
- show confirmed position in debug mode only
- avoid sending movement faster than server cadence
- rejected movement should reset target marker cleanly
- rubberband should be minimized without weakening server authority

The current 275ms client cadence guard supports the backend 250ms authoritative movement cooldown.

## Debug vs Gameplay Separation

Debug information should remain available, but not be the primary gameplay interface.

Keep visible in debug mode:

- opcode logs
- RuntimeEntityID
- confirmed tile coordinates
- packet sequence
- raw payload summaries

Hide or minimize in Alpha gameplay mode:

- full packet log
- excessive protocol details
- backend-only naming
- raw entity internals

## Technical Constraints

Alpha UX must not bypass:

- server-authoritative movement
- backend movement cooldown
- anti-speedhack validation
- anti-teleport validation
- authoritative combat result
- RuntimeEntityID lifecycle
- backend reward grant
- persistence validation

Visual polish must sit on top of real state, not fake local-only state.

## Minimum Alpha UX Acceptance Criteria

The Alpha client is acceptable when:

- player can enter the world and understand where they are
- player can move and see movement feedback
- player can identify a creature
- player can attack the creature
- player can see damage/death feedback
- player can receive reward feedback
- player can see creature respawn feedback
- player can attack the respawned creature
- debug logs are not required for normal understanding
- backend logs are only needed for validation, not gameplay

## Not In Scope

Not required for this step:

- final sprites
- final animations
- final UI skin
- full minimap
- full inventory UI
- full equipment UI
- party UI
- PvP UI
- quest tracker
- chat polish
- production HUD

## Recommended Next Steps

After this document:

1. Document Debug Client vs Alpha Client separation.
2. Runtime validate movement cadence guard.
3. Add small visible reward feedback.
4. Add target frame baseline.
5. Reduce packet log dominance.
6. Move toward generic creature presentation.

## Architecture Decision

Do not discard the current debug client.

Instead, evolve it into two modes:

- Debug Mode: technical validation, logs, packet/state visibility.
- Alpha Mode: playable minimum UX using the same real backend state.

This keeps engineering visibility while moving toward a playable internal Alpha.