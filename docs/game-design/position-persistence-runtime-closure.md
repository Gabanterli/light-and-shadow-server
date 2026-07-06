# Position Persistence Runtime Closure

## Status

- Completed
- Runtime Validated

## Related Commits

- `f05a6d5 Document initial position sync runtime closure`
- `4af7fb9 Apply own player update to debug marker`
- `04c0864 Send initial player position update after select`

## Runtime Validation Summary

This validation confirms that a player's position is correctly persisted to the database upon movement and successfully restored upon relogging. The test involved moving a character in the debug world, disconnecting the client, and then logging back in with the same character to verify that they spawned at their last confirmed position, not the default starting point.

## Backend Persistence Behavior

The backend `Gateway` orchestrates the saving of a player's state through two primary mechanisms:

1.  **Autosave:** A background ticker fires every 30 seconds, triggering a save for all active, "dirty" (modified) characters.
2.  **Save on Disconnect/Logout:** When a client's TCP connection is closed, the `handleClient` cleanup logic immediately triggers a final save for the associated character.

In both cases, the Gateway snapshots the character's current state from the `CombatManager` (including position) and the `inventories` map, then passes this snapshot to the `PersistenceManager` for transactional saving to PostgreSQL.

## Runtime Evidence with Gabriela

The following logs were captured during runtime validation with the character "Gabriela":

1.  **Initial Spawn/Relog:** Upon selecting the character, the Gateway correctly loads the last saved position and sends it to the client.

    ```
    Initial player position update sent to client
    playerID: Gabriela
    x: 108
    y: 102
    z: 0
    ```

2.  **Disconnect/Logout:** After moving, disconnecting the client triggers a successful save.

    ```
    Saving player state on disconnect / logout...
    Character state, version, gold and inventory successfully persisted to PostgreSQL
    player: Gabriela, lvl: 1, x: 108, y: 102, gold: 1000, new_version: 62
    Successfully saved character state on database
    player: Gabriela, new_version: 62
    ```

3.  **Cleanup:** The server correctly deregisters the player from runtime systems after saving.
    ```
    Player connection deregistered from AOIManager
    Cleaned up player states from systems on disconnect
    ```

## Autosave Behavior

The autosave mechanism was confirmed to be active, with logs appearing consistently every 30 seconds, ensuring periodic state persistence for active players.

```
Autosave ticker fired. Persisting all active character states...
```

## Disconnect/Logout Save Behavior

The save-on-disconnect behavior was confirmed as the primary mechanism for ensuring the very last state of a player is persisted, preventing rollbacks from unexpected logouts.

## Architectural Decision

- The **backend remains fully authoritative** over player state, including position.
- The **Godot client does not persist position** locally; it only renders the state received from the server.
- **PostgreSQL is the single source of truth** for persistent character data.
- The `CombatManager` holds the active, in-memory runtime position of all entities.
- The `Gateway` is responsible for creating a snapshot of the active state and orchestrating the save.

## What This Closes

This task closes the fundamental loop of player position persistence. The system can now reliably save a player's location and restore it upon their return to the world.

## What Remains Future Work

- Testing persistence during server crash/restart scenarios.
- Validating persistence with multiple characters on the same account and multiple concurrent players.
- Implementing real player movement input (e.g., WASD, click-to-move).
- Migrating the `SC_PLAYER_UPDATE` payload to a more efficient binary format.
- Replacing the debug world marker with a proper animated player sprite.

## Next Recommended Step

With the core persistence and synchronization loop validated, the next logical step is to **implement real-time movement input** on the client. This would involve capturing keyboard/mouse events in Godot and sending `CS_MOVE_REQUEST` packets to the server, still within the context of the debug world.