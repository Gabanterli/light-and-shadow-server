export interface Position {
  x: number;
  y: number;
  z: number;
  region: string;
}

export interface InventoryItem {
  id: string;
  name: string;
  qty: number;
  value: number;
  type: string;
}

export interface EquippedItem {
  slot: string;
  id: string;
  name: string;
  value: number;
}

export interface QuestObjective {
  type: string;
  target_id: string;
  required_qty: number;
  current_qty: number;
}

export interface QuestState {
  quest_id: string;
  title: string;
  description: string;
  objectives: QuestObjective[];
  rewards: {
    experience: number;
    gold: number;
    items?: { item_id: string; quantity: number }[];
  };
}

export interface PlayerQuestData {
  activeQuests: QuestState[];
  completedQuestIds: string[];
}

export interface CharacterProfile {
  id: string;
  name: string;
  race?: "Human" | "Forest Elf" | "Ice Elf" | "Dwarf" | "Green Orc";
  level: number;
  xp: number;
  vocation_state: "Unassigned" | "Novice" | "Knight" | "Mage" | "Archer" | "Assassin" | "Cleric";
  position: Position;
  inventory: InventoryItem[];
  equipment: EquippedItem[];
  gold: number;
  quest_state: PlayerQuestData;
  created_at: string;
  last_login: string;
  skills?: Record<string, { level: number; xp: number }>;
  elemental_affinities?: Record<string, { level: number; xp: number }>;
}

export type SessionState =
  | "LOGIN"
  | "CHARACTER_CREATION"
  | "CHARACTER_SELECT"
  | "LOADING"
  | "IN_WORLD"
  | "TUTORIAL_ACTIVE";

export interface ServerApiLog {
  id: string;
  timestamp: string;
  method: "GET" | "POST" | "PUT" | "DELETE" | "WEBSOCKET";
  path: string;
  status: number;
  latencyMs: number;
  payload: string;
}
