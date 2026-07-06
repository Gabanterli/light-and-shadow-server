# Combat Protocol Technical Alpha Audit

## Status

- Audit Completed
- Purpose: To document the existing backend combat protocol contract before client-side implementation for the technical alpha.

## Backend Opcodes

The following opcodes are implemented in the backend `protocol` package:

- `3000 CS_ATTACK_REQUEST`
- `3001 CS_CAST_SKILL`
- `3002 SC_DAMAGE_EVENT`
- `3003 SC_TARGET_DEAD`

## Attack Request Contract (3000)

The `CS_ATTACK_REQUEST` payload is defined as:

- `string targetID`
- `string weaponType`

## Cast Skill Request Contract (3001)

The `CS_CAST_SKILL` payload is defined as:

- `uint32 skillID`
- `string targetID`
- `fixed32 targetX`
- `fixed32 targetY`

## Damage Event Contract (3002)

The `SC_DAMAGE_EVENT` payload is defined as:

- `string attackerID`
- `string targetID`
- `fixed32 damage`
- `bool isCrit`
- `bool isHit`
- `bool success`
- `string skillName`

## Target Dead Event Contract (3003)

The `SC_TARGET_DEAD` payload is defined as:

- `string targetID`

## Backend Readiness

- The `Gateway` has implemented handlers for `CS_ATTACK_REQUEST` and `CS_CAST_SKILL`.
- The `CombatManager` is called to process these requests authoritatively.
- The `Gateway` sends `SC_DAMAGE_EVENT` and `SC_TARGET_DEAD` packets based on the `CombatManager`'s results.
- A test enemy, `Orc_Elite`, is registered in the combat system upon player entry, providing a valid target for testing.
- The `PveManager` is active, ready for future integration with loot drops upon target death.

## Godot Client Gaps

- `BinaryProtocol.cs` does not yet have codecs for the combat opcodes (3000-3003).
- `GatewayTcpClient.cs` does not have methods like `SendAttackRequestAsync`.
- `DebugWorldEntryController.cs` does not have a UI or logic to send attack requests.
- `DebugWorldEntryController.cs` does not have handlers in its packet router to process incoming `SC_DAMAGE_EVENT` or `SC_TARGET_DEAD` packets.

## Technical Alpha Implication

The backend is ready to process and respond to basic combat actions. The client, however, is not yet equipped to participate in this loop. The next phase of work must focus on implementing the client-side protocol and UI hooks to enable end-to-end combat testing.

## Recommended Next Steps

1.  Add combat protocol codecs to `BinaryProtocol.cs`.
2.  Add `SendAttackRequestAsync` to `GatewayTcpClient.cs`.
3.  Add a debug "Attack Orc_Elite" button to the `DebugWorldEntry` scene and controller.
4.  Add packet router handlers for `SC_DAMAGE_EVENT` and `SC_TARGET_DEAD` to log the results.
5.  Perform a full runtime validation of the attack -> damage -> death loop.

## Explicit Constraints

- Do not change the backend or the existing protocol contract.
- Do not add final combat UI or art assets.
- Keep the implementation focused on the debug client and technical validation.