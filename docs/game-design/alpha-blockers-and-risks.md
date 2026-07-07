# Alpha Blockers And Risks

## Status

Draft baseline for Alpha blocker tracking.

## Purpose

Track the remaining blockers and risks before Light and Shadow can be considered ready for a Technical Alpha checkpoint.

## Critical Blockers

### A2-B1 - Movement Rubberband Polish

Observed runtime warning:

    Authoritative movement validation failed

Impact:

- Can make early Alpha feel unstable.
- May confuse combat positioning.
- Can cause client/server position mismatch.

Required action:

- Inspect movement tolerance.
- Inspect client movement send cadence.
- Inspect saved position and local prediction.
- Avoid weakening anti-cheat globally.

### A2-B2 - Generic Creature Lifecycle

Current creature loop still depends on debug Orc Elite aliasing.

Impact:

- Works for debug validation.
- Not scalable to real multi-creature AOI.
- CombatManager still resolves through Orc_Elite compatibility path.

Required action:

- Prepare generic creature identity path.
- Move more combat lifecycle operations toward RuntimeEntityID.
- Keep debug alias only until generic creature rendering is stable.

### A2-B3 - Client Debug World Dependency

Current gameplay loop is still debug-world oriented.

Impact:

- Useful for validation.
- Not enough for Alpha user experience.
- UI and targeting are still minimal.

Required action:

- Define minimum Alpha UI.
- Reduce dependency on manual debug logs.
- Make target/death/respawn feedback visible and reliable.

### A2-B4 - Reward Loop Still Temporary

Gold fallback works, but it is not final loot design.

Impact:

- Good Alpha-safe reward validation.
- Does not represent final corpse/loot container system.

Required action:

- Keep fallback for Alpha.
- Later implement corpse ownership and loot container flow.

## Medium Risks

### Persistence Risk

Autosave is working, but Alpha needs repeat-session verification.

Required validation:

- Login.
- Earn reward.
- Autosave.
- Logout.
- Login again.
- Confirm persisted state.

### Economy Risk

Gold fallback rewards can inflate economy if not debug-limited.

Required rule:

- Keep fallback tied to debug Orc Elite only.
- Do not generalize fallback rewards to normal PvE.

### PvE Scaling Risk

Single creature flow does not validate multiple simultaneous creature instances.

Required future validation:

- Multiple spawns.
- Multiple RuntimeEntityIDs.
- Multiple death/respawn loops.
- No cross-target loot contamination.

### Client UX Risk

If logs are required to know what happened, Alpha UX is not ready.

Required future work:

- Visible target state.
- Visible reward notification.
- Visible respawn/death state.
- Better movement correction feedback.

## Alpha Entry Recommendation

Before declaring Technical Alpha, complete:

- Movement rubberband inspection and minimal polish.
- Alpha runtime checklist run.
- Generic creature lifecycle planning.
- One clean full-session validation from login to reward persistence.

## Current Decision

Do not add new feature systems before stabilizing:

- movement
- generic creature lifecycle path
- Alpha checklist validation