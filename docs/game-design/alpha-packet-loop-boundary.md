# Alpha Packet Loop Boundary

## Status

Accepted for Alpha Technical planning.

## Purpose

Define how the Alpha UI will consume real backend packets without copying the full DebugWorldEntryController implementation.

The immediate goal is to prepare safe wiring for real InventorySync opcode 4001 so the Alpha top bar can show real Level, HP, and Mana.

## Current State

The Alpha UI currently has:

- A dedicated AlphaWorldEntryScene.
- A dedicated AlphaWorldEntryController.
- Safe navigation from DebugAuthScene.
- Optional AuthSession and GatewayTcpClient passed into AlphaWorldEntryController.
- Real selected character name from AuthSession.
- A reusable DebugTileWorldView mounted as AlphaWorldView.
- No Alpha packet loop yet.

The Debug Client currently owns the full debug packet loop, packet router, movement handling, combat handling, chunk sync, inventory sync, and raw packet logging.

## Data Sources

### AuthSession

AuthSession is safe to read immediately in Alpha UI.

Allowed Alpha usage:

- IsAuthenticated
- IsCharacterSelected
- SelectedCharacterName

Not allowed:

- Treating AuthSession as a source for HP, Mana, Level, inventory items, target state, or world state.

### InventorySync 4001

InventorySync opcode 4001 is the first real gameplay state packet that Alpha UI should consume.

It provides:

- Item count
- Level
- Current HP
- Max HP
- Current Mana
- Max Mana
- Combat stat baseline

Alpha UI may use it for:

- Top player frame
- Backpack summary
- Future stat panel

Alpha UI must not fabricate these values before receiving real InventorySync.

## Boundary Rule

Only one packet loop may read from a GatewayTcpClient at a time.

DebugWorldEntryController and AlphaWorldEntryController must not both read from the same GatewayTcpClient concurrently.

Scene ownership must be explicit:

- DebugWorldEntryController owns the packet loop while DebugWorldEntryScene is active.
- AlphaWorldEntryController may own a packet loop only while AlphaWorldEntryScene is active.
- Leaving the scene must cancel its loop.
- Returning to DebugAuthScene must not leave a background packet listener alive.

## Do Not Copy Rule

AlphaWorldEntryController must not become a copy of DebugWorldEntryController.

Allowed reuse:

- Shared data models.
- BinaryProtocol decoders.
- GatewayTcpClient send/receive methods.
- DebugIncomingPacketRouter only if it remains generic enough.
- DebugTileWorldView as a temporary reusable visual component.
- Small extracted helpers in later tasks.

Not allowed:

- Copying the entire DebugWorldEntryController packet handling block.
- Copying debug packet log behavior into Alpha UI.
- Showing raw opcodes by default in Alpha UI.
- Exposing RuntimeEntityID by default in Alpha UI.
- Treating Debug UI labels as Alpha state.

## Recommended Alpha Packet Loop v1

Alpha packet loop v1 should be minimal.

Initial handlers:

- 4001 InventorySync

Optional later handlers:

- 2006 ChunkData
- 2005 MoveConfirm
- 2007 PlayerUpdate
- 3002 DamageEvent
- 3003 TargetDead
- 3004 CreatureRespawn

Do not add all handlers in the first Alpha packet loop task.

## Alpha Packet Loop v1 Behavior

On Alpha scene ready:

1. Verify GatewayClient is not null.
2. Verify GatewayClient.IsConnected is true.
3. Create a CancellationTokenSource.
4. Start a background receive loop.
5. Receive packets from GatewayClient.
6. Route only opcode 4001 initially.
7. Ignore or lightly log unsupported packets to Alpha system feedback without raw packet spam.
8. Use CallDeferred for UI changes.
9. Cancel the loop on _ExitTree.
10. Do not dispose GatewayClient unless Alpha explicitly owns final logout behavior in a later task.

## InventorySync UI Contract

When opcode 4001 is received:

Alpha top bar should update:

- Player: SelectedCharacterName from AuthSession
- Level: InventorySyncData.Level
- HP: InventorySyncData.Health / InventorySyncData.MaxHealth
- Mana: InventorySyncData.Mana / InventorySyncData.MaxMana

Backpack tab may update:

- Item count only

Bottom System tab may update:

- Inventory sync received
- Timestamp or simple confirmation

Bottom Combat tab should not update from InventorySync unless the sync confirms a reward after combat in a later task.

## Error Handling

If InventorySync decode fails:

- Do not crash the client.
- Show a short Alpha system message.
- Keep previous known top bar values.
- Log a concise error with GD.PrintErr.

If GatewayClient is null:

- Show client missing.
- Do not start packet loop.

If GatewayClient is disconnected:

- Show client disconnected.
- Do not start packet loop.

## Threading Rule

Packet receive loop is asynchronous.

UI writes must be deferred to the main thread.

Allowed pattern:

- Receive/decode packet in async loop.
- Store decoded values in local data object.
- CallDeferred into a UI method.

Do not mutate Godot nodes directly from a background task.

## Debug Client Separation

The Debug Client remains the authoritative diagnostic tool.

Debug Client may continue showing:

- Raw packets
- Opcode numbers
- Packet sizes
- RuntimeEntityID
- Movement internal state
- Chunk counters
- Debug inventory sync booleans
- Full packet log

Alpha UI should show only player-facing summarized state unless a later debug overlay is explicitly added.

## Implementation Plan

A10-R9 - Add Alpha InventorySync state model

- Add minimal fields to AlphaWorldEntryController.
- No receive loop yet if the implementation can be staged safely.
- Keep top bar update method ready for real data.

A10-R10 - Add Alpha packet loop for InventorySync 4001 only

- Start loop only when Alpha scene is active.
- Decode opcode 4001 only.
- Update Level, HP, Mana from real backend InventorySync.
- Cancel loop on _ExitTree.

A10-R11 - Add Alpha backpack item count summary

- Use real item count from InventorySync.
- Do not build full inventory UI yet.

A10-R12 - Add Alpha system feedback for InventorySync

- Show a discreet system message when sync arrives.
- No raw packet spam.

## Decision

Alpha UI will consume real backend packets incrementally.

The first real packet to wire is InventorySync 4001.

HP, Mana, and Level must remain pending sync until InventorySync 4001 is received by the Alpha packet loop.

The Alpha packet loop must be minimal, cancellable, scene-owned, and separate from the Debug Client diagnostic flow.
