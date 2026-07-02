import { CharacterProfile } from "@/src/types";
import { progressionEventBus, GameEvent, GameEventPayload } from "./progression_event_bus";

// Virtual or real logging helper
export interface TraceLogEntry {
  event: string;
  playerId: string;
  timestamp: string;
  before: any;
  after: any;
  xpGained: number;
  skillChanges: string[];
  affinityChanges: string[];
  details: string;
}

export class ProgressionEngine {
  private activeChar: CharacterProfile | null = null;
  private onUpdate: ((char: CharacterProfile, event?: string, details?: string) => void) | null = null;
  private traceLogs: TraceLogEntry[] = [];

  constructor() {
    this.setupListeners();
  }

  public initialize(char: CharacterProfile, onUpdate: (char: CharacterProfile, event?: string, details?: string) => void) {
    this.activeChar = JSON.parse(JSON.stringify(char)); // deep copy to prevent direct client mutations
    this.onUpdate = onUpdate;
    this.hydrate(this.activeChar!);
  }

  public getActiveCharacter(): CharacterProfile | null {
    return this.activeChar;
  }

  public getTraceLogs(): TraceLogEntry[] {
    return this.traceLogs;
  }

  private hydrate(char: CharacterProfile) {
    if (!char.skills) {
      char.skills = {
        sword_fighting: { level: 10, xp: 0 },
        axe_fighting: { level: 10, xp: 0 },
        club_fighting: { level: 10, xp: 0 },
        dagger_fighting: { level: 10, xp: 0 },
        distance_fighting: { level: 10, xp: 0 },
        magic_level: { level: 1, xp: 0 },
        shielding: { level: 10, xp: 0 }
      };
    }
    if (!char.elemental_affinities) {
      char.elemental_affinities = {
        fire: { level: 1, xp: 0 },
        water: { level: 1, xp: 0 },
        earth: { level: 1, xp: 0 },
        lightning: { level: 1, xp: 0 }
      };
    }
  }

  private setupListeners() {
    progressionEventBus.on("ON_MONSTER_KILL", (payload) => this.handleMonsterKill(payload));
    progressionEventBus.on("ON_DAMAGE_DEALT", (payload) => this.handleDamageDealt(payload));
    progressionEventBus.on("ON_DAMAGE_BLOCKED", (payload) => this.handleDamageBlocked(payload));
    progressionEventBus.on("ON_SPELL_CAST", (payload) => this.handleSpellCast(payload));
    progressionEventBus.on("ON_SKILL_USE", (payload) => this.handleSkillUse(payload));
    progressionEventBus.on("ON_ELEMENTAL_DAMAGE", (payload) => this.handleElementalDamage(payload));
    progressionEventBus.on("ON_QUEST_PROGRESS", (payload) => this.handleQuestProgress(payload));
    progressionEventBus.on("ON_LADDER_INTERACTION", (payload) => this.handleLadderInteraction(payload));
    progressionEventBus.on("ON_ITEM_PICKUP", (payload) => this.handleItemPickup(payload));
    progressionEventBus.on("ON_ITEM_DROP", (payload) => this.handleItemDrop(payload));
  }

  private triggerUpdate(beforeState: CharacterProfile, xpGained = 0, skillChanges: string[] = [], affinityChanges: string[] = [], details = "", eventName: string) {
    if (!this.activeChar) return;

    const afterState = JSON.parse(JSON.stringify(this.activeChar));

    // Save trace log
    const logEntry: TraceLogEntry = {
      event: eventName,
      playerId: this.activeChar.id,
      timestamp: new Date().toISOString(),
      before: JSON.parse(JSON.stringify(beforeState)),
      after: afterState,
      xpGained,
      skillChanges,
      affinityChanges,
      details
    };
    
    this.traceLogs.unshift(logEntry);
    this.writeToTraceFile(logEntry);

    if (this.onUpdate) {
      this.onUpdate(afterState, eventName, details);
    }
  }

  private writeToTraceFile(entry: TraceLogEntry) {
    const formattedLog = `[${entry.timestamp}] EVENT: ${entry.event} | PLAYER: ${entry.playerId}
Before Level: ${entry.before.level} (XP: ${entry.before.xp})
After Level: ${entry.after.level} (XP: ${entry.after.xp})
XP Gained: ${entry.xpGained}
Skill Changes: ${entry.skillChanges.join(", ") || "None"}
Affinity Changes: ${entry.affinityChanges.join(", ") || "None"}
Details: ${entry.details}
--------------------------------------------------\n`;
    console.log("[Trace] " + formattedLog);
  }

  // --- EVENT HANDLERS ---

  private handleMonsterKill(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const expReward = payload.expReward || 12;
    this.activeChar.xp += expReward;

    const skillChanges: string[] = [];
    const affinityChanges: string[] = [];

    // Level curve calculation
    let leveledUp = false;
    let nextLvlXp = this.activeChar.level === 1 ? 110 : 400;
    while (this.activeChar.xp >= nextLvlXp) {
      this.activeChar.xp -= nextLvlXp;
      this.activeChar.level += 1;
      leveledUp = true;
      nextLvlXp = this.activeChar.level === 1 ? 110 : 400;
    }

    // Trigger quest abates if applicable
    let updatedActiveQuests = [...this.activeChar.quest_state.activeQuests];
    if (payload.monsterType === "sewer_rat") {
      updatedActiveQuests = updatedActiveQuests.map(q => {
        if (q.quest_id === "quest_tutorial_sewer_rats") {
          const objectives = q.objectives.map(obj => {
            if (obj.type === "KillMonster" && obj.target_id === "sewer_rat") {
              return { ...obj, current_qty: Math.min(obj.required_qty, obj.current_qty + 1) };
            }
            return obj;
          });
          return { ...q, objectives };
        }
        return q;
      });
      this.activeChar.quest_state.activeQuests = updatedActiveQuests;
    }

    const details = `Derrotou ${payload.monsterName || "Monstro"} (${payload.monsterType}). Ganhou +${expReward} XP.${leveledUp ? ` SUBIU PARA O LEVEL ${this.activeChar.level}! 🎉` : ""}`;
    this.triggerUpdate(before, expReward, skillChanges, affinityChanges, details, "ON_MONSTER_KILL");
  }

  private handleDamageDealt(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const dmg = payload.damage || 0;
    const weaponType = payload.weaponType || "sword";
    
    // Melee skill increases by usage
    const skillChanges: string[] = [];
    let targetSkill = "sword_fighting";
    if (weaponType === "axe") targetSkill = "axe_fighting";
    else if (weaponType === "club") targetSkill = "club_fighting";
    else if (weaponType === "dagger") targetSkill = "dagger_fighting";
    else if (weaponType === "distance") targetSkill = "distance_fighting";

    const skills = this.activeChar.skills!;
    if (skills[targetSkill]) {
      const oldLevel = skills[targetSkill].level;
      skills[targetSkill].xp += Math.floor(dmg * 0.4) + 5; // Gain XP relative to damage dealt
      
      const xpNeeded = oldLevel * 100;
      if (skills[targetSkill].xp >= xpNeeded) {
        skills[targetSkill].xp -= xpNeeded;
        skills[targetSkill].level += 1;
        skillChanges.push(`${targetSkill} level up: ${oldLevel} -> ${skills[targetSkill].level}`);
      }
    }

    this.triggerUpdate(before, 0, skillChanges, [], `Causou ${dmg} de dano com ${weaponType}.`, "ON_DAMAGE_DEALT");
  }

  private handleDamageBlocked(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const dmgBlocked = payload.damageBlocked || 0;
    const skillChanges: string[] = [];

    const skills = this.activeChar.skills!;
    if (skills.shielding) {
      const oldLevel = skills.shielding.level;
      skills.shielding.xp += Math.floor(dmgBlocked * 0.8) + 8; // Block shielding gains
      
      const xpNeeded = oldLevel * 100;
      if (skills.shielding.xp >= xpNeeded) {
        skills.shielding.xp -= xpNeeded;
        skills.shielding.level += 1;
        skillChanges.push(`shielding level up: ${oldLevel} -> ${skills.shielding.level}`);
      }
    }

    this.triggerUpdate(before, 0, skillChanges, [], `Bloqueou ${dmgBlocked} de dano com escudo.`, "ON_DAMAGE_BLOCKED");
  }

  private handleSpellCast(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const spellId = payload.spellId;
    const manaCost = payload.manaCost || 0;
    const skillChanges: string[] = [];
    const affinityChanges: string[] = [];

    // Magic level increases with mana spent
    const skills = this.activeChar.skills!;
    if (skills.magic_level) {
      const oldLevel = skills.magic_level.level;
      skills.magic_level.xp += Math.floor(manaCost * 1.5) + 6;
      
      const xpNeeded = oldLevel * 120;
      if (skills.magic_level.xp >= xpNeeded) {
        skills.magic_level.xp -= xpNeeded;
        skills.magic_level.level += 1;
        skillChanges.push(`magic_level level up: ${oldLevel} -> ${skills.magic_level.level}`);
      }
    }

    // Elemental affinity progression
    const elementalType = payload.elementalType || "";
    const affinities = this.activeChar.elemental_affinities!;
    
    let targetAffinity = "";
    if (elementalType === "fire") targetAffinity = "fire";
    else if (elementalType === "water" || elementalType === "ice") targetAffinity = "water";
    else if (elementalType === "earth" || elementalType === "nature") targetAffinity = "earth";
    else if (elementalType === "lightning" || elementalType === "holy" || elementalType === "shadow") targetAffinity = "lightning";

    if (targetAffinity && affinities[targetAffinity]) {
      const oldAff = affinities[targetAffinity].level;
      if (oldAff < 100) { // Max cap 100
        affinities[targetAffinity].xp += 20; // 20 XP per cast
        if (affinities[targetAffinity].xp >= 1000) {
          affinities[targetAffinity].xp -= 1000;
          affinities[targetAffinity].level += 1;
          affinityChanges.push(`${targetAffinity} affinity level up: ${oldAff} -> ${affinities[targetAffinity].level}`);
        }
      }
    }

    this.triggerUpdate(before, 0, skillChanges, affinityChanges, `Conjurou magia ${spellId} (consumo: ${manaCost} mana).`, "ON_SPELL_CAST");
  }

  private handleSkillUse(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const skillType = payload.skillType; // e.g. sword_fighting
    const skillChanges: string[] = [];

    const skills = this.activeChar.skills!;
    if (skills[skillType]) {
      const oldLevel = skills[skillType].level;
      skills[skillType].xp += 15; // 15 XP per swing
      
      const xpNeeded = oldLevel * 100;
      if (skills[skillType].xp >= xpNeeded) {
        skills[skillType].xp -= xpNeeded;
        skills[skillType].level += 1;
        skillChanges.push(`${skillType} level up: ${oldLevel} -> ${skills[skillType].level}`);
      }
    }

    this.triggerUpdate(before, 0, skillChanges, [], `Usou habilidade de combate: ${skillType}.`, "ON_SKILL_USE");
  }

  private handleElementalDamage(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const dmg = payload.damage || 0;
    const element = payload.element || "fire";
    const affinityChanges: string[] = [];

    // Inflicting elemental damage increases corresponding affinity
    const affinities = this.activeChar.elemental_affinities!;
    let targetAffinity = "";
    if (element === "fire") targetAffinity = "fire";
    else if (element === "water" || element === "ice") targetAffinity = "water";
    else if (element === "earth" || element === "nature") targetAffinity = "earth";
    else if (element === "lightning" || element === "holy" || element === "shadow") targetAffinity = "lightning";

    if (targetAffinity && affinities[targetAffinity]) {
      const oldAff = affinities[targetAffinity].level;
      if (oldAff < 100) {
        affinities[targetAffinity].xp += Math.floor(dmg * 0.2) + 2;
        if (affinities[targetAffinity].xp >= 1000) {
          affinities[targetAffinity].xp -= 1000;
          affinities[targetAffinity].level += 1;
          affinityChanges.push(`${targetAffinity} affinity level up: ${oldAff} -> ${affinities[targetAffinity].level}`);
        }
      }
    }

    this.triggerUpdate(before, 0, [], affinityChanges, `Causou ${dmg} de dano elemental do tipo ${element}.`, "ON_ELEMENTAL_DAMAGE");
  }

  private handleQuestProgress(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const { questId, objectiveType, targetId, quantity, isComplete, xpReward, goldReward } = payload;
    
    if (isComplete) {
      // Remove from active quests and add to completed list
      this.activeChar.quest_state.activeQuests = this.activeChar.quest_state.activeQuests.filter(q => q.quest_id !== questId);
      if (!this.activeChar.quest_state.completedQuestIds.includes(questId)) {
        this.activeChar.quest_state.completedQuestIds.push(questId);
      }

      // Award Gold
      if (goldReward) {
        this.activeChar.gold += goldReward;
        // Also check if coins exist in inventory to sync
        const bronzeIdx = this.activeChar.inventory.findIndex(i => i.id === "bronze_coin");
        if (bronzeIdx !== -1) {
          this.activeChar.inventory[bronzeIdx].qty += goldReward;
        }
      }

      // Award XP
      if (xpReward) {
        this.activeChar.xp += xpReward;
        // Level up check
        let leveledUp = false;
        let nextLvlXp = this.activeChar.level === 1 ? 110 : 400;
        while (this.activeChar.xp >= nextLvlXp) {
          this.activeChar.xp -= nextLvlXp;
          this.activeChar.level += 1;
          leveledUp = true;
          nextLvlXp = this.activeChar.level === 1 ? 110 : 400;
        }
      }

      this.triggerUpdate(
        before,
        xpReward || 0,
        [],
        [],
        `Concluiu a missão: ${questId}. Obteve +${xpReward} XP e +${goldReward} Gold gp.`,
        "ON_QUEST_PROGRESS"
      );
    } else {
      // Progress update inside active quests
      this.activeChar.quest_state.activeQuests = this.activeChar.quest_state.activeQuests.map(q => {
        if (q.quest_id === questId) {
          const objectives = q.objectives.map(obj => {
            if (obj.type === objectiveType && obj.target_id === targetId) {
              return { ...obj, current_qty: Math.min(obj.required_qty, obj.current_qty + quantity) };
            }
            return obj;
          });
          return { ...q, objectives };
        }
        return q;
      });

      this.triggerUpdate(
        before,
        0,
        [],
        [],
        `Atualizou progresso da missão ${questId}: ${objectiveType} ${targetId} (+${quantity}).`,
        "ON_QUEST_PROGRESS"
      );
    }
  }

  private handleLadderInteraction(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const ladderType = payload.ladderType; // LADDER_UP or LADDER_DOWN
    const { fromX, fromY } = payload;

    if (ladderType === "LADDER_DOWN" && fromY === 11) {
      this.activeChar.position = { x: 10, y: 13, z: -1, region: "Ironhold Sewers" };
    } else if (ladderType === "LADDER_UP" && fromY === 12) {
      this.activeChar.position = { x: 10, y: 10, z: 0, region: "Ironhold Bastion" };
    }

    this.triggerUpdate(
      before,
      0,
      [],
      [],
      `Interagiu com escada (${ladderType}) em (${fromX}, ${fromY}). Posicionado em (${this.activeChar.position.x}, ${this.activeChar.position.y}).`,
      "ON_LADDER_INTERACTION"
    );
  }

  private handleItemPickup(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const { itemId, itemName, qty, value, type } = payload;
    const inv = [...this.activeChar.inventory];
    const idx = inv.findIndex(i => i.id === itemId);
    if (idx !== -1) {
      inv[idx] = { ...inv[idx], qty: inv[idx].qty + qty };
    } else {
      inv.push({ id: itemId, name: itemName, qty, value, type });
    }
    this.activeChar.inventory = inv;

    this.triggerUpdate(before, 0, [], [], `Pegou item do chão: ${qty}x ${itemName}.`, "ON_ITEM_PICKUP");
  }

  private handleItemDrop(payload: GameEventPayload) {
    if (!this.activeChar || payload.playerId !== this.activeChar.id) return;
    const before = JSON.parse(JSON.stringify(this.activeChar));
    this.hydrate(this.activeChar);

    const { itemId, itemName, qty } = payload;
    const inv = [...this.activeChar.inventory];
    const idx = inv.findIndex(i => i.id === itemId);
    if (idx !== -1) {
      if (inv[idx].qty > qty) {
        inv[idx] = { ...inv[idx], qty: inv[idx].qty - qty };
      } else {
        inv.splice(idx, 1);
      }
    }
    this.activeChar.inventory = inv;

    this.triggerUpdate(before, 0, [], [], `Descartou item no chão: ${qty}x ${itemName}.`, "ON_ITEM_DROP");
  }
}

export const progressionEngine = new ProgressionEngine();
