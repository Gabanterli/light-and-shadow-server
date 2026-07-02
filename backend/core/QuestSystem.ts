import { progressionEventBus } from "./progression_event_bus";

export class QuestSystem {
  public static updateObjectiveProgress(
    playerId: string,
    questId: string,
    objectiveType: string,
    targetId: string,
    quantity: number
  ): void {
    progressionEventBus.emit("ON_QUEST_PROGRESS", {
      playerId,
      timestamp: Date.now(),
      questId,
      objectiveType,
      targetId,
      quantity,
      isComplete: false
    });
  }

  public static completeQuest(
    playerId: string,
    questId: string,
    xpReward: number,
    goldReward: number
  ): void {
    progressionEventBus.emit("ON_QUEST_PROGRESS", {
      playerId,
      timestamp: Date.now(),
      questId,
      objectiveType: "Complete",
      targetId: questId,
      quantity: 1,
      isComplete: true,
      xpReward,
      goldReward
    });
  }
}
