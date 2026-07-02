import { progressionEventBus } from "./progression_event_bus";

export class CombatEngine {
  public static processWeaponStrike(
    playerId: string,
    damage: number,
    weaponType: string,
    element: string = "Physical"
  ): void {
    let skillType = "sword_fighting";
    if (weaponType === "axe") skillType = "axe_fighting";
    else if (weaponType === "club") skillType = "club_fighting";
    else if (weaponType === "dagger") skillType = "dagger_fighting";
    else if (weaponType === "distance") skillType = "distance_fighting";

    // Trigger skill use event
    progressionEventBus.emit("ON_SKILL_USE", {
      playerId,
      timestamp: Date.now(),
      skillType
    });

    // Trigger damage dealt event
    progressionEventBus.emit("ON_DAMAGE_DEALT", {
      playerId,
      timestamp: Date.now(),
      damage,
      weaponType
    });

    // Trigger elemental damage event if applicable
    if (element && element !== "Physical") {
      progressionEventBus.emit("ON_ELEMENTAL_DAMAGE", {
        playerId,
        timestamp: Date.now(),
        damage,
        element: element.toLowerCase()
      });
    }
  }

  public static processShieldBlock(playerId: string, damageBlocked: number): void {
    progressionEventBus.emit("ON_DAMAGE_BLOCKED", {
      playerId,
      timestamp: Date.now(),
      damageBlocked
    });
  }
}
