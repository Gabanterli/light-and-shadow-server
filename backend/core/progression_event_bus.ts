export type GameEvent =
  | "ON_MONSTER_KILL"
  | "ON_DAMAGE_DEALT"
  | "ON_DAMAGE_BLOCKED"
  | "ON_SPELL_CAST"
  | "ON_SKILL_USE"
  | "ON_ELEMENTAL_DAMAGE"
  | "ON_QUEST_PROGRESS"
  | "ON_LADDER_INTERACTION"
  | "ON_ITEM_PICKUP"
  | "ON_ITEM_DROP";

export interface GameEventPayload {
  playerId: string;
  timestamp: number;
  [key: string]: any;
}

export type EventListener = (payload: GameEventPayload) => void;

class ProgressionEventBus {
  private listeners: Record<string, EventListener[]> = {};

  public on(event: GameEvent, callback: EventListener): void {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  public off(event: GameEvent, callback: EventListener): void {
    if (!this.listeners[event]) return;
    this.listeners[event] = this.listeners[event].filter(cb => cb !== callback);
  }

  public emit(event: GameEvent, payload: GameEventPayload): void {
    console.log(`[EventBus] Received event: ${event}`, payload);
    const callbacks = this.listeners[event] || [];
    for (const callback of callbacks) {
      try {
        callback(payload);
      } catch (err) {
        console.error(`[EventBus] Error executing listener for event ${event}:`, err);
      }
    }
  }
}

export const progressionEventBus = new ProgressionEventBus();
