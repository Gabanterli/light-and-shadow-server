# A5 - Debug Client vs Alpha Client Separation

## Status

Draft baseline.

## Purpose

Define the separation between the current Debug Client and the future Alpha Client experience.

The current client is valuable because it exposes real technical state.

The Alpha Client must become playable without removing the ability to validate backend systems.

## Core Decision

Do not delete the Debug Client.

Do not replace it with a purely visual mock.

Instead, evolve the client into two presentation modes:

- Debug Mode
- Alpha Mode

Both modes must use the same real backend state.

## Debug Mode

Debug Mode is for engineering validation.

It should expose technical information needed to verify MMO systems.

### Debug Mode Should Show

- packet logs
- opcode names and numbers
- packet sequence
- RuntimeEntityID
- confirmed tile coordinates
- last movement target
- move pending state
- raw combat events
- target dead packets
- respawn packets
- inventory sync state
- chunk streaming state
- backend-facing entity names
- technical warnings

### Debug Mode Should Allow

- manual movement testing
- repeated attack testing
- RuntimeEntityID validation
- stale target validation
- respawn validation
- inventory/reward validation
- packet routing inspection

### Debug Mode May Look Ugly

Debug Mode does not need final art.

It can use:

- labels
- text logs
- placeholder markers
- debug colors
- technical overlays

Its job is correctness, not presentation.

## Alpha Mode

Alpha Mode is for playable internal testing.

It should hide most raw technical details and present the game loop clearly.

### Alpha Mode Should Show

- player in the world
- visible nearby tiles
- visible creature
- selected target frame
- target alive/dead state
- attack feedback
- damage feedback
- reward feedback
- respawn feedback
- connection status
- small optional debug indicator

### Alpha Mode Should Hide Or Minimize

- full packet log
- raw payload details
- opcode spam
- packet sequence spam
- internal backend naming
- raw RuntimeEntityID by default
- verbose movement diagnostics
- backend-only debug labels

### Alpha Mode Must Still Use Real State

Alpha Mode must not fake:

- movement acceptance
- player position
- creature death
- creature respawn
- reward grant
- inventory sync
- combat result
- persistence

All visible state must come from server-confirmed or server-derived state.

## Shared Client Foundation

Both modes should share:

- GatewayClient
- DebugIncomingPacketRouter or future packet router
- BinaryProtocol
- session state
- world state snapshot
- chunk store
- movement confirmation handling
- combat event handling
- creature lifecycle handling
- reward/inventory handling

Only presentation should diverge.

## Recommended Structure

Long-term recommended split:

    scripts/
      DebugWorldEntryController.cs
      AlphaWorldController.cs
      WorldState/
      UI/
      Network/

DebugWorldEntryController remains the engineering validation screen.

AlphaWorldController becomes the playable internal Alpha screen.

Shared systems should move out of DebugWorldEntryController as they mature.

## Migration Strategy

### Step 1 - Keep Debug Client Stable

Do not break the current debug runtime flow.

Preserve:

- login to world entry
- movement validation
- attack button
- RuntimeEntityID targeting
- target death display
- creature respawn display
- reward validation

### Step 2 - Add Alpha UX Elements Inside Current Client

Before creating a separate Alpha scene, add small visible improvements:

- compact reward notification
- target frame baseline
- clearer creature marker
- clearer alive/dead state
- smaller debug log area

### Step 3 - Extract Shared State

Move reusable state logic away from the debug controller when it becomes too large.

Candidates:

- player world state
- creature world state
- target state
- reward event state
- movement state

### Step 4 - Create Alpha Scene

Only after shared state is stable, create a separate Alpha scene/controller.

The Alpha scene should consume the same packet/state systems, not duplicate protocol logic.

## Anti-Regression Rules

Alpha Mode must not bypass:

- server-authoritative movement
- movement cooldown
- anti-speedhack validation
- anti-teleport validation
- authoritative combat
- RuntimeEntityID lifecycle
- backend reward grants
- persistence

Debug visibility can be reduced, but backend truth cannot be replaced by local-only assumptions.

## Current Debug Client Responsibilities

The current debug client is still responsible for:

- proving packet flow
- proving movement confirmation
- proving combat request flow
- proving target death sync
- proving respawn sync
- proving RuntimeEntityID targeting
- proving reward feedback
- proving inventory sync

Until Alpha Mode exists, the debug client remains the main runtime test surface.

## Alpha Client Minimum Experience

The Alpha Client should allow a tester to understand the loop without reading terminal logs:

1. I logged in.
2. I entered the world.
3. I see my character.
4. I see an enemy.
5. I can attack.
6. I see damage or attack feedback.
7. I see the enemy die.
8. I see a reward.
9. I see the enemy respawn.
10. I can attack the respawned enemy.

## Not In Scope

This document does not implement:

- new scene files
- new UI nodes
- new C# controllers
- sprite art
- animations
- production HUD
- final inventory UI

## Next Recommended Work

After this document:

1. Runtime checkpoint movement cadence and PvE reward loop.
2. Add small Alpha-style reward feedback.
3. Add minimal target frame.
4. Reduce debug log dominance.
5. Prepare generic creature presentation.
6. Plan AlphaWorldController extraction.

## Architecture Decision

The project should move from:

    Debug-only playable test

to:

    Debug Mode + Alpha Mode

without losing engineering visibility.

This gives the project both:

- strong technical validation
- a path toward a real playable Alpha