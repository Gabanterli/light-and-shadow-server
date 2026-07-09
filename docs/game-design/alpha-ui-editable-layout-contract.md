# Alpha UI Editable Layout Contract

## Goal

Create an editable Alpha HUD path without breaking the current technical Alpha runtime.

## Current Rule

The existing `AlphaWorldEntryController` remains the active runtime authority for the Alpha screen until component binding is migrated safely.

## Component Contract

The editable Alpha HUD is split into small reusable panels:

- `AlphaTopBarPanel`
- `AlphaWorldPanel`
- `AlphaBattlePanel`
- `AlphaBackpackPanel`
- `AlphaFeedbackLogPanel`

Each component exposes explicit bind methods instead of owning gameplay state.

## Authority Rule

UI components never become authoritative for:

- position
- combat state
- target identity
- damage
- loot
- inventory
- cooldowns
- economy

They only render state provided by the controller or future client-side presentation model.

## Runtime Migration Plan

1. Add reusable component scripts.
2. Add editable scene placeholder.
3. Keep current Alpha runtime unchanged.
4. Bind current Alpha controller to components in a later task.
5. Only after runtime parity, remove duplicated UI construction from the controller.

## Editing Rule

Gabriel may edit the future Alpha HUD scene visually in Godot, but gameplay state flow must continue through explicit binding APIs.
## Editable Preview Controller

`AlphaHudLayoutController` exists only to make the editable HUD scene previewable in Godot.

It may seed fake visual data for layout editing, but it must not become gameplay authority.

Runtime data must still flow through the active Alpha controller or a future presentation model.