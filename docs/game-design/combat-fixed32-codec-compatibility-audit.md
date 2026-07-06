# Combat Fixed32 Codec Compatibility Audit

## Status

- Audit Completed
- Purpose: To define the `fixed32` data type contract before implementing combat codecs in the Godot client.

## Backend Fixed32 Definition

The backend uses a custom `fixed32` data type to represent floating-point numbers with a fixed precision of 3 decimal places. This is achieved by scaling the float by 1000 and storing it as a 32-bit integer. This avoids floating-point inaccuracies over the network while maintaining sufficient precision for game mechanics like damage and coordinates.

## Backend Read Behavior

The Go backend's `ReadFixed32` function performs the following steps:
1.  Reads 4 bytes from the network buffer.
2.  Interprets these bytes as a Little Endian `int32`.
3.  Divides the integer value by `1000.0` to convert it back to a `float64`.

## Backend Write Behavior

The Go backend's `WriteFixed32` function performs the following steps:
1.  Takes a `float64` value.
2.  Multiplies it by `1000.0` and rounds it to the nearest whole number.
3.  Casts the result to an `int32`.
4.  Writes the `int32` value as 4 bytes in Little Endian format to the network buffer.

## Combat Fields Using Fixed32

The following fields in the combat protocol use this `fixed32` format:
- `CS_CAST_SKILL` (3001): `targetX`, `targetY`
- `SC_DAMAGE_EVENT` (3002): `damage`

## Godot Current Gap

The client's `scripts/BinaryProtocol.cs` currently has helpers for standard types like `ReadFloat64LE`, but it does not yet have the specific `ReadFixed32LE` and `WriteFixed32LE` helpers required to correctly handle these combat protocol fields. Using `ReadFloat64LE` would lead to data misinterpretation.

## Required Godot Implementation Rule

To ensure compatibility, the C# implementation must mirror the backend's logic:

- **`WriteFixed32LE(double value)`**: Must calculate `(int)Math.Round(value * 1000.0)` and write the resulting `int` to the buffer as a 4-byte Little Endian integer.
- **`ReadFixed32LE()`**: Must read a 4-byte Little Endian `int` from the buffer and return the value cast to a `double` and divided by `1000.0`.

## Compatibility Examples

- `10.5` (float) becomes `10500` (int32)
- `35.123` (float) becomes `35123` (int32)
- `0.0` (float) becomes `0` (int32)

## Technical Alpha Implication

Failure to implement these specific `fixed32` codecs will result in incorrect data being sent and received for critical combat actions. For example, damage values and skill target coordinates would be completely wrong, making the combat loop impossible to test or validate.

## Recommended Next Step

The next logical task is to implement the combat protocol codecs in `BinaryProtocol.cs`. This involves first creating the `ReadFixed32LE` and `WriteFixed32LE` helper methods and then using them to build the codecs for opcodes `3000` through `3003`.

## Constraints

- Do not change the backend Go implementation.
- Do not change the existing network protocol contract.
- Do not add final combat UI or art assets.
- Keep the implementation focused on the debug client and technical validation.