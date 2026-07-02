import { progressionEventBus } from "./progression_event_bus";

export class SpellSystem {
  public static castSpell(
    playerId: string,
    spellId: string,
    manaCost: number,
    elementalType: string = ""
  ): void {
    progressionEventBus.emit("ON_SPELL_CAST", {
      playerId,
      timestamp: Date.now(),
      spellId,
      manaCost,
      elementalType
    });
  }

  public static applyElementalSpellDamage(
    playerId: string,
    damage: number,
    elementalType: string
  ): void {
    progressionEventBus.emit("ON_ELEMENTAL_DAMAGE", {
      playerId,
      timestamp: Date.now(),
      damage,
      element: elementalType.toLowerCase()
    });
  }
}
