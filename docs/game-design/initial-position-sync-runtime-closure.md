# Initial Player Position Sync Runtime Closure

## Status

- Completed
- Runtime Validated

## Commits Related

- `4af7fb9 Apply own player update to debug marker`
- `04c0864 Send initial player position update after select`

## Backend Behavior

The backend `Gateway` was modified to enhance the character selection flow. After a successful `CS_CHAR_SELECT_REQUEST`, the server now performs the following actions in order:

1.  Loads the character's state from PostgreSQL, including their last saved position (`savedX`, `savedY`, `savedZ`).
2.  Sends the `SC_CHAR_SELECT_RESPONSE` (1007) to confirm the selection.
3.  **Immediately sends an `SC_PLAYER_UPDATE` (2001) packet containing the character's server-authoritative starting position.**
4.  Proceeds with sending the initial `SC_INVENTORY_SYNC` and `SC_CHUNK_DATA` packets.

This ensures the client receives its correct starting coordinates before any other world state.

## Godot Behavior

The Godot client's `DebugWorldEntryController.cs` was updated to correctly handle the initial position packet:

1.  The `OnPlayerUpdateReceived` handler now checks if the `PlayerID` in the `SC_PLAYER_UPDATE` (2001) packet matches the `Session.SelectedCharacterName`.
2.  If it's an update for the local player, the client:
    - Updates its internal `_currentConfirmedPos` state.
    - Updates the visual player marker in the `DebugTileWorldView`.
    - Updates the "Confirmed Pos" UI label to reflect the server's position.
    - Logs that the authoritative position has been applied.

## Runtime Validation Evidence

The end-to-end flow was validated in a runtime environment.

1.  After logging in and selecting the character "Gabriela", the client transitioned to the `DebugWorldEntry` scene.
2.  **Before** any manual movement was initiated, the UI correctly displayed:
    - `IsCharacterSelected: True`
    - `SelectedCharacterName: Gabriela`
    - `Inv. Sync Received: True`
    - `Chunks Received: 9`
    - **`Confirmed Pos: (104, 102, 0)`**
3.  This confirms that the initial `SC_PLAYER_UPDATE` was received and applied, overriding any local default position.
4.  Subsequent debug movement via the "Send Debug Move" button functioned correctly, starting from the new authoritative position. A move to `(105, 102, 0)` was successfully confirmed by the server.

## Architectural Decision

- The existing `SC_PLAYER_UPDATE` (2001) opcode was reused for this initial synchronization to maintain protocol simplicity. No new opcode was created.
- The backend remains fully authoritative over player position, from the initial spawn to subsequent movements.
- The Godot client acts purely as a renderer of state, applying the position data it receives from the server without question.

## What This Closes

This task successfully closes the loop on initial player spawning. The client no longer relies on a hardcoded starting position and is now correctly synchronized with the server's state from the moment it enters the world.

## What Remains Future Work

- Implementing real player movement via keyboard input (e.g., WASD) or mouse clicks.
- Migrating the `SC_PLAYER_UPDATE` payload from JSON to a more efficient binary format.
- Improving the `PacketLogTextEdit` UI for better readability and filtering.
- Replacing the debug world marker with a proper animated player sprite.
- Enhancing the persistence and respawn logic to handle more complex scenarios like logout, disconnects, and crashes.

## Next Recommended Step

The next logical phase is to build upon this synchronized foundation. Two viable paths are:

1.  **Validate Persistence:** Implement and test the autosave/logout/relogin cycle to ensure the player's position is correctly saved and restored.
2.  **Implement Real Input:** Begin implementing the client-side logic for sending movement requests based on actual player input (e.g., keyboard or mouse), still using the debug world view.