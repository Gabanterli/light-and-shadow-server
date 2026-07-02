import { progressionEventBus } from "./progression_event_bus";

export class TibiaBattleSystem {
  public static selectTarget(playerId: string, monsterId: string, monsterName: string): void {
    console.log(`[TibiaBattleSystem] Player ${playerId} targeted monster ${monsterId} (${monsterName})`);
  }

  public static triggerMonsterKill(
    playerId: string,
    monsterId: string,
    monsterType: string,
    monsterName: string,
    expReward: number
  ): void {
    progressionEventBus.emit("ON_MONSTER_KILL", {
      playerId,
      timestamp: Date.now(),
      monsterId,
      monsterType,
      monsterName,
      expReward
    });
  }
}
