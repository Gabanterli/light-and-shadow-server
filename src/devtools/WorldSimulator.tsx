import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import {
  Globe,
  Map,
  Compass,
  Sliders,
  Sparkles,
  Gift,
  Sword,
  Shield,
  Activity,
  Flame,
  Snowflake,
  Leaf,
  Moon,
  Lock,
  Unlock,
  MapPin,
  Coins,
  TrendingUp,
  RefreshCw,
  Info,
  HelpCircle,
  ArrowRight
} from "lucide-react";

// --- REGISTRY DATA COPIED DIRECTLY FROM CANONICAL BIBLE CONFIGS ---

interface Biome {
  id: string;
  name: string;
  description: string;
  elementalAffinity: string;
  hazardLevel: number;
  difficultyMultiplier: number;
  encounterDensity: string;
  nativeMonsterFamilies: string[];
  environmentalHazards: string[];
  lootModifiers: {
    rarityBonus: number;
    materialsAvailable: string[];
    affixWeights: Record<string, number>;
  };
  color: string;
  bgGlow: string;
  borderColor: string;
}

const BIOMES: Biome[] = [
  {
    id: "forest",
    name: "Whispering Woodlands (Forest)",
    description: "Lush, dense temperate forests inhabited by beasts, wild plants, and druidic forces.",
    elementalAffinity: "Nature",
    hazardLevel: 0.1,
    difficultyMultiplier: 1.0,
    encounterDensity: "Medium",
    nativeMonsterFamilies: ["Beast", "Plant", "Insect", "Goblinoid"],
    environmentalHazards: ["None (Mild pollen weather)"],
    lootModifiers: {
      rarityBonus: 1.0,
      materialsAvailable: ["Treated Lumber", "Herbs", "Leather Pelts"],
      affixWeights: { vitality: 1.5, dexterity: 1.2, poison_resistance: 1.3 }
    },
    color: "text-emerald-400 bg-emerald-500/10",
    bgGlow: "rgba(16, 185, 129, 0.05)",
    borderColor: "border-emerald-500/30"
  },
  {
    id: "desert",
    name: "Shifting Sands (Desert)",
    description: "Scorching sun-baked wastes with immense dune fields and hidden subterranean ruins.",
    elementalAffinity: "Earth",
    hazardLevel: 0.4,
    difficultyMultiplier: 1.15,
    encounterDensity: "Low",
    nativeMonsterFamilies: ["Reptilian", "Elemental", "Insect", "Undead"],
    environmentalHazards: ["Heat Exhaustion (Slows health regen)", "Sandstorms (Reduces line of sight)"],
    lootModifiers: {
      rarityBonus: 1.1,
      materialsAvailable: ["Fine Sand", "Ancient Relics", "Sunstone Ore"],
      affixWeights: { fire_resistance: 1.4, accuracy: 1.3, stamina: 1.2 }
    },
    color: "text-amber-400 bg-amber-500/10",
    bgGlow: "rgba(245, 158, 11, 0.05)",
    borderColor: "border-amber-500/30"
  },
  {
    id: "frozen_tundra",
    name: "Frostbite Wastes (Frozen Tundra)",
    description: "Glacial valleys and ice-locked peaks where only the hardiest survive.",
    elementalAffinity: "Ice",
    hazardLevel: 0.5,
    difficultyMultiplier: 1.2,
    encounterDensity: "Medium",
    nativeMonsterFamilies: ["Beast", "Giant", "Elemental", "Undead"],
    environmentalHazards: ["Hypothermia (Drains stamina and slows movement speed)"],
    lootModifiers: {
      rarityBonus: 1.15,
      materialsAvailable: ["Glacial Ice Crystal", "Frost Leather", "Silver Ore"],
      affixWeights: { cold_damage: 1.5, physical_defense: 1.2, ice_resistance: 1.6 }
    },
    color: "text-sky-400 bg-sky-500/10",
    bgGlow: "rgba(14, 165, 233, 0.05)",
    borderColor: "border-sky-500/30"
  },
  {
    id: "swamp",
    name: "Mire of Murk (Swamp)",
    description: "Decaying marshes and toxic bayous filled with gas vents and treacherous sinkholes.",
    elementalAffinity: "Poison",
    hazardLevel: 0.6,
    difficultyMultiplier: 1.25,
    encounterDensity: "High",
    nativeMonsterFamilies: ["Reptilian", "Plant", "Insect", "Aberration"],
    environmentalHazards: ["Toxic Gas Miasma (Applies poison damage over time)"],
    lootModifiers: {
      rarityBonus: 1.2,
      materialsAvailable: ["Poison Sacks", "Bog Iron", "Rare Moss"],
      affixWeights: { poison_damage: 1.6, lifesteal: 1.4, acid_touch: 1.3 }
    },
    color: "text-teal-400 bg-teal-500/10",
    bgGlow: "rgba(20, 184, 166, 0.05)",
    borderColor: "border-teal-500/30"
  },
  {
    id: "volcanic",
    name: "Cinder Wastes (Volcanic Region)",
    description: "Basalt plains and magma rivers. Extreme danger, rich in prime metal veins.",
    elementalAffinity: "Fire",
    hazardLevel: 0.8,
    difficultyMultiplier: 1.4,
    encounterDensity: "High",
    nativeMonsterFamilies: ["Demon", "Elemental", "Dragonkin"],
    environmentalHazards: ["Magma Flows (Instant high fire damage)", "Ash Choke (Silences spellcasting periodically)"],
    lootModifiers: {
      rarityBonus: 1.3,
      materialsAvailable: ["Obsidian", "Sulfur Ore", "Magma Core"],
      affixWeights: { fire_damage: 1.7, strength: 1.4, critical_multiplier: 1.3 }
    },
    color: "text-rose-400 bg-rose-500/10",
    bgGlow: "rgba(244, 63, 94, 0.05)",
    borderColor: "border-rose-500/30"
  },
  {
    id: "sacred_lands",
    name: "Sanctuary of Light (Sacred Lands)",
    description: "Radiant fields of golden light, protected by celestial structures and divine guardians.",
    elementalAffinity: "Holy",
    hazardLevel: 0.2,
    difficultyMultiplier: 1.3,
    encounterDensity: "Low",
    nativeMonsterFamilies: ["Celestial", "Construct", "Humanoid"],
    environmentalHazards: ["Radiant Burn (Burns non-believers/unholy classes)"],
    lootModifiers: {
      rarityBonus: 1.4,
      materialsAvailable: ["Sacred Dust", "Light Crystal", "Pure Gold Wire"],
      affixWeights: { holy_damage: 1.8, healing_bonus: 1.5, intelligence: 1.3 }
    },
    color: "text-amber-300 bg-amber-400/10",
    bgGlow: "rgba(251, 191, 36, 0.05)",
    borderColor: "border-amber-400/30"
  },
  {
    id: "corrupted_zones",
    name: "The Abyss Scar (Corrupted Zones)",
    description: "Torn lands leaking dark void energy, twisting all native life into nightmares.",
    elementalAffinity: "Shadow",
    hazardLevel: 0.9,
    difficultyMultiplier: 1.5,
    encounterDensity: "Extreme",
    nativeMonsterFamilies: ["Undead", "Demon", "Aberration", "Voidspawn"],
    environmentalHazards: ["Sanity Drain (Decreases maximum health over time)", "Void Fissures (Erupts with shadow damage)"],
    lootModifiers: {
      rarityBonus: 1.5,
      materialsAvailable: ["Void Essence", "Corrupted Bone", "Dark Iron"],
      affixWeights: { shadow_damage: 1.8, corruption_infusion: 1.6, critical_chance: 1.4 }
    },
    color: "text-violet-400 bg-violet-500/10",
    bgGlow: "rgba(139, 92, 246, 0.05)",
    borderColor: "border-violet-500/30"
  }
];

interface WorldRegion {
  id: string;
  name: string;
  continent: string;
  primaryBiome: string;
  minLevel: number;
  maxLevel: number;
  keySettlements: string[];
  centerCoordinates: { x: number; y: number };
  pointsOfInterest: string[];
  nativeMonsterFamilies: string[];
  accessRestriction?: string;
}

const REGIONS: WorldRegion[] = [
  {
    id: "main_plains",
    name: "Eldret Farmlands",
    continent: "Main Continent",
    primaryBiome: "forest",
    minLevel: 1,
    maxLevel: 15,
    keySettlements: ["Ironhold Bastion (100,100)", "Ravenshire (1600,600)"],
    centerCoordinates: { x: 150, y: 350 },
    pointsOfInterest: ["Royal Farmland", "Vaelor Guard Post", "Trade Road Crossing"],
    nativeMonsterFamilies: ["Beast", "Goblinoid", "Insect"]
  },
  {
    id: "main_mountains",
    name: "Crags of Khaz Tirith",
    continent: "Main Continent",
    primaryBiome: "desert",
    minLevel: 15,
    maxLevel: 40,
    keySettlements: ["Stone Tirith (500,500)"],
    centerCoordinates: { x: 220, y: 280 },
    pointsOfInterest: ["Dwarven Great Forge", "Deep Mine Shafts", "Thane's Citadel"],
    nativeMonsterFamilies: ["Elemental", "Giant", "Insect"]
  },
  {
    id: "main_coast",
    name: "Blackwater Shores",
    continent: "Main Continent",
    primaryBiome: "forest",
    minLevel: 20,
    maxLevel: 50,
    keySettlements: ["Blackwater Bay (1200,1200)"],
    centerCoordinates: { x: 300, y: 360 },
    pointsOfInterest: ["Smuggler's Cove", "Blackwater Port", "Shipwreck Reef"],
    nativeMonsterFamilies: ["Reptilian", "Humanoid", "Beast"]
  },
  {
    id: "nature_elarin",
    name: "Ancient Canopy of Elarin",
    continent: "Nature Continent",
    primaryBiome: "forest",
    minLevel: 1,
    maxLevel: 60,
    keySettlements: ["Elarin"],
    centerCoordinates: { x: 450, y: 150 },
    pointsOfInterest: ["Elder Tree", "Druidic Grove", "Whispering Falls"],
    nativeMonsterFamilies: ["Plant", "Beast", "Celestial"]
  },
  {
    id: "shadow_noctharyn",
    name: "Eternal Wastes of Noctharyn",
    continent: "Shadow Continent",
    primaryBiome: "corrupted_zones",
    minLevel: 1,
    maxLevel: 60,
    keySettlements: ["Noctharyn"],
    centerCoordinates: { x: 650, y: 420 },
    pointsOfInterest: ["Ashen Necropolis", "Void Fissure Ground Zero", "Crypt of Whispers"],
    nativeMonsterFamilies: ["Undead", "Voidspawn", "Aberration"]
  },
  {
    id: "fire_central_volcano",
    name: "Central Primordial Volcano",
    continent: "Fire Continent",
    primaryBiome: "volcanic",
    minLevel: 50,
    maxLevel: 60,
    keySettlements: ["Pyra Magnus Palace"],
    centerCoordinates: { x: 820, y: 380 },
    pointsOfInterest: ["Primordial Caldera", "Basalt Core Dungeon", "Lesser General Altar"],
    nativeMonsterFamilies: ["Demon", "Elemental", "Dragonkin"]
  },
  {
    id: "fire_ash_plains",
    name: "Crimson Ash Plains",
    continent: "Fire Continent",
    primaryBiome: "volcanic",
    minLevel: 30,
    maxLevel: 50,
    keySettlements: ["Crimson Hollow"],
    centerCoordinates: { x: 770, y: 320 },
    pointsOfInterest: ["Red Orc Arena", "Sulfur Geysers", "Ash Quarry"],
    nativeMonsterFamilies: ["Beast", "Elemental", "Reptilian"]
  },
  {
    id: "fire_forge_mountains",
    name: "Forge Mountains",
    continent: "Fire Continent",
    primaryBiome: "volcanic",
    minLevel: 40,
    maxLevel: 55,
    keySettlements: ["Molten Anvil"],
    centerCoordinates: { x: 870, y: 290 },
    pointsOfInterest: ["Grand Cyclopean Forge", "Magma Rivers", "Obsidian Spire"],
    nativeMonsterFamilies: ["Giant", "Construct", "Elemental"]
  },
  {
    id: "holy_luminaar",
    name: "Sovereign Plains of Luminaar",
    continent: "Holy Continent",
    primaryBiome: "sacred_lands",
    minLevel: 50,
    maxLevel: 60,
    keySettlements: ["Luminaar"],
    centerCoordinates: { x: 550, y: 240 },
    pointsOfInterest: ["Trial of Spirits Altar", "Grand Paladin Cathedral", "Aureum Bastion"],
    nativeMonsterFamilies: ["Celestial", "Construct", "Humanoid"]
  },
  {
    id: "ice_elarisheim",
    name: "Glacial Peaks of Elarisheim",
    continent: "Ice Continent",
    primaryBiome: "frozen_tundra",
    minLevel: 1,
    maxLevel: 60,
    keySettlements: ["Elarisheim"],
    centerCoordinates: { x: 180, y: 120 },
    pointsOfInterest: ["Frozen Mountain Peaks", "Everfrost Palace", "Deep Khaz Mines"],
    nativeMonsterFamilies: ["Beast", "Giant", "Elemental"]
  },
  {
    id: "ice_ymirr_cavern",
    name: "Ymirr's Hidden Cavern",
    continent: "Ice Continent",
    primaryBiome: "frozen_tundra",
    minLevel: 55,
    maxLevel: 60,
    keySettlements: [],
    centerCoordinates: { x: 230, y: 90 },
    pointsOfInterest: ["Ymirr's Throneroom", "Sacred Jotunn Rune Wall"],
    nativeMonsterFamilies: ["Giant", "Undead"],
    accessRestriction: "RESTRICTED - Requires Forbidden Mountain pass or Khaz Tirith secret mines."
  }
];

interface LootItem {
  name: string;
  rarity: "Common" | "Magic" | "Rare" | "Epic" | "Legendary";
  type: string;
  biomeAffix: string;
  stats: string;
  ilvl: number;
  itemColor: string;
}

export function WorldSimulator() {
  // Simulator State
  const [selectedBiomeId, setSelectedBiomeId] = useState<string>("forest");
  const [selectedRegionId, setSelectedRegionId] = useState<string>("main_plains");
  const [playerActivity, setPlayerActivity] = useState<number>(5); // 0 to 20
  const [dangerLevel, setDangerLevel] = useState<number>(0.3); // 0 to 1
  const [encounterType, setEncounterType] = useState<"solo" | "pack" | "horde" | "boss_encounter">("pack");
  const [playerLevel, setPlayerLevel] = useState<number>(15);
  const [discoveredSecretCavern, setDiscoveredSecretCavern] = useState<boolean>(false);
  const [spawnerLogs, setSpawnerLogs] = useState<Array<{ id: string; text: string; type: string }>>([]);
  const [lootDrops, setLootDrops] = useState<LootItem[]>([]);
  const [simulatedEnemies, setSimulatedEnemies] = useState<Array<{ id: string; name: string; isElite: boolean; hp: number; maxHp: number; angle: number; dist: number; element: string }>>([]);

  // Active Biome & Region Lookups
  const activeBiome = BIOMES.find(b => b.id === selectedBiomeId) || BIOMES[0];
  const activeRegion = REGIONS.find(r => r.id === selectedRegionId) || REGIONS[0];

  // Derive Escalation Tiers based on activity score
  const getActivityTier = (score: number) => {
    if (score <= 2) return { tier: "Quiet", rateMod: 1.0, eliteMod: 1.0, aggro: 1.0, color: "text-slate-400 border-slate-500/20 bg-slate-500/5" };
    if (score <= 8) return { tier: "Active", rateMod: 1.3, eliteMod: 1.25, aggro: 1.1, color: "text-amber-400 border-amber-500/30 bg-amber-500/5" };
    if (score <= 15) return { tier: "Overrun", rateMod: 1.8, eliteMod: 1.75, aggro: 1.25, color: "text-rose-400 border-rose-500/30 bg-rose-500/5" };
    return { tier: "Cataclysmic", rateMod: 2.5, eliteMod: 2.5, aggro: 1.5, color: "text-violet-400 border-violet-500/35 bg-violet-500/5" };
  };

  const activityTier = getActivityTier(playerActivity);

  // Math Calculations (Strict formulas from spawn_system.json)
  const baseRate = encounterType === "solo" ? 8 : encounterType === "pack" ? 15 : encounterType === "horde" ? 30 : 60;
  const dynamicSpawnRate = (baseRate / (1.0 + (playerActivity * 0.25) + (dangerLevel * 0.15))).toFixed(1);

  const baseEliteChance = encounterType === "solo" ? 0.05 : encounterType === "pack" ? 0.15 : encounterType === "horde" ? 0.10 : 0.80;
  const eliteProbabilityValue = baseEliteChance * (1.0 + (playerActivity * 0.50)) * (1.0 + (activeBiome.hazardLevel * 0.30));
  const eliteProbability = Math.min(100, Math.max(0, eliteProbabilityValue * 100)).toFixed(1);

  const threatMultiplier = (1.0 + (activeBiome.difficultyMultiplier - 1.0) * 0.5 + (encounterType === "solo" ? 0.0 : encounterType === "pack" ? 0.15 : encounterType === "horde" ? 0.30 : 0.60)).toFixed(2);

  // Underdog XP calculation
  const recommendedLevel = activeRegion.minLevel;
  const underdogXPMultiplier = playerLevel < recommendedLevel
    ? Math.min(4.0, Math.pow(recommendedLevel / playerLevel, 1.5)).toFixed(2)
    : "1.00";

  // Trigger Spawner simulation whenever biome, encounterType, activity or danger changes
  useEffect(() => {
    simulateSpawn();
  }, [selectedBiomeId, selectedRegionId, playerActivity, dangerLevel, encounterType]);

  const simulateSpawn = () => {
    // Generate simulated enemies in circle
    const count = encounterType === "solo" ? 1 : encounterType === "pack" ? 3 : encounterType === "horde" ? 8 : 2;
    const newEnemies = [];
    const mobFamilies = activeBiome.nativeMonsterFamilies;

    for (let i = 0; i < count; i++) {
      const family = mobFamilies[Math.floor(Math.random() * mobFamilies.length)];
      const isElite = Math.random() * 100 < parseFloat(eliteProbability);
      const isBoss = encounterType === "boss_encounter" && i === 0;

      const baseHp = isBoss ? 50000 : isElite ? 4500 : 800;
      const scaledHp = Math.round(baseHp * parseFloat(threatMultiplier));

      const name = isBoss
        ? `[BOSS] Primordial ${family} Colossus`
        : isElite
        ? `Elite ${family} Marauder`
        : `Wild ${family}`;

      newEnemies.push({
        id: Math.random().toString(),
        name,
        isElite: isElite || isBoss,
        hp: scaledHp,
        maxHp: scaledHp,
        angle: (360 / count) * i + Math.random() * 15,
        dist: 45 + Math.random() * 30,
        element: activeBiome.elementalAffinity
      });
    }

    setSimulatedEnemies(newEnemies);

    // Add logging message
    const msg = `Spawn triggered in ${activeBiome.name} (${activeRegion.name}). Generated ${count} entities with ${eliteProbability}% elite modifier. Threat multiplier: ${threatMultiplier}x.`;
    setSpawnerLogs(prev => [{ id: Math.random().toString(), text: msg, type: "info" }, ...prev.slice(0, 49)]);
  };

  // Roll Loot Drop
  const rollLoot = () => {
    const rarityPool: Array<"Common" | "Magic" | "Rare" | "Epic" | "Legendary"> = ["Common", "Magic", "Rare", "Epic", "Legendary"];
    const weights = [0.55, 0.28, 0.12, 0.04, 0.01];

    // Modify weights based on hazard level & regional modifiers
    const rarityBonus = activeBiome.lootModifiers.rarityBonus * (1 + dangerLevel * 0.5);
    const modifiedWeights = weights.map((w, idx) => {
      if (idx >= 2) return w * rarityBonus; // Boost rare and above
      return w;
    });
    // Re-normalize weights
    const sum = modifiedWeights.reduce((a, b) => a + b, 0);
    const normalizedWeights = modifiedWeights.map(w => w / sum);

    // Choose rarity
    const rand = Math.random();
    let cumulative = 0;
    let chosenRarity = rarityPool[0];
    for (let i = 0; i < normalizedWeights.length; i++) {
      cumulative += normalizedWeights[i];
      if (rand <= cumulative) {
        chosenRarity = rarityPool[i];
        break;
      }
    }

    // Material & Affixes
    const materials = activeBiome.lootModifiers.materialsAvailable;
    const coreMaterial = materials[Math.floor(Math.random() * materials.length)];
    const affixWeights = activeBiome.lootModifiers.affixWeights;
    const topAffix = Object.keys(affixWeights)[0];

    // Item configurations
    const items = ["Greatsword", "Bastion Shield", "Oak Staff", "Tunic", "Greaves", "Signet Ring", "Offhand Catalyst"];
    const baseItem = items[Math.floor(Math.random() * items.length)];

    let itemColor = "text-slate-400 border-slate-700 bg-slate-900/40";
    if (chosenRarity === "Magic") itemColor = "text-sky-400 border-sky-800 bg-sky-950/20";
    if (chosenRarity === "Rare") itemColor = "text-violet-400 border-violet-800 bg-violet-950/20";
    if (chosenRarity === "Epic") itemColor = "text-amber-400 border-amber-800 bg-amber-950/20";
    if (chosenRarity === "Legendary") itemColor = "text-orange-400 border-orange-800 bg-orange-950/20 animate-pulse";

    // Affix description
    const affixName = topAffix.charAt(0).toUpperCase() + topAffix.slice(1).replace("_", " ");
    const statRoll = Math.round((chosenRarity === "Legendary" ? 75 : chosenRarity === "Epic" ? 45 : chosenRarity === "Rare" ? 25 : chosenRarity === "Magic" ? 12 : 5) * (1 + dangerLevel));

    const newLoot: LootItem = {
      name: `${chosenRarity} ${baseItem} of the ${activeBiome.elementalAffinity}`,
      rarity: chosenRarity,
      type: baseItem,
      biomeAffix: `+${statRoll} ${affixName} (Weighted)`,
      stats: `Crafted from ${coreMaterial}. Base Protection scale: ${Math.round(45 * activeBiome.difficultyMultiplier)}`,
      ilvl: Math.round(playerLevel + (dangerLevel * 10) + (chosenRarity === "Legendary" ? 10 : 2)),
      itemColor
    };

    setLootDrops(prev => [newLoot, ...prev.slice(0, 19)]);
    setSpawnerLogs(prev => [{
      id: Math.random().toString(),
      text: `Rolled dynamic item loot drop: [${newLoot.rarity}] ${newLoot.name} (iLvl ${newLoot.ilvl})`,
      type: "loot"
    }, ...prev.slice(0, 49)]);
  };

  // Region clicking updates selected biome automatically
  const handleRegionClick = (region: WorldRegion) => {
    if (region.accessRestriction && !discoveredSecretCavern && region.id === "ice_ymirr_cavern") {
      setSpawnerLogs(prev => [{
        id: Math.random().toString(),
        text: `Cannot travel directly to Ymirr's Hidden Cavern. Movement filter: ACCESS BLOCKED.`,
        type: "error"
      }, ...prev.slice(0, 49)]);
      return;
    }
    setSelectedRegionId(region.id);
    setSelectedBiomeId(region.primaryBiome);
    // Adjust danger level depending on min recommended level
    setDangerLevel(Math.min(1.0, Math.max(0.1, region.minLevel / 60)));
  };

  return (
    <div className="w-full flex flex-col gap-6" id="world-foundation-container">
      
      {/* Title & Philosophy Banner */}
      <div className="bg-slate-900/60 rounded-2xl p-6 border border-slate-800/80 relative overflow-hidden" id="world-bible-banner">
        <div className="absolute top-0 right-0 w-64 h-64 bg-amber-500/5 rounded-full blur-3xl pointer-events-none" />
        <div className="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4">
          <div className="flex items-center gap-3">
            <Globe className="w-8 h-8 text-amber-400" />
            <div>
              <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2">
                World Foundation Bible
                <span className="text-xs bg-amber-500/20 text-amber-400 font-mono px-2.5 py-0.5 rounded-full border border-amber-500/20">
                  PRE-WORLD GENERATION LAYER
                </span>
              </h2>
              <p className="text-sm text-slate-400 max-w-2xl mt-1">
                Authoritative world registry mapping geographical hierarchies, biome-specific loot weighting, and real-time adaptive mob spawns.
              </p>
            </div>
          </div>
          <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 max-w-sm">
            <span className="text-xs text-amber-400/80 font-semibold block mb-1">Canonical Directive:</span>
            <p className="text-[11px] text-slate-400 leading-normal italic">
              "The world is a dynamic ecosystem where gameplay systems emerge from biome, encounter density, and systemic interactions rather than static map design."
            </p>
          </div>
        </div>
      </div>

      {/* Abstract World Travel Map */}
      <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 flex flex-col gap-4" id="world-travel-map-card">
        <div className="flex justify-between items-center border-b border-slate-800/60 pb-3">
          <div className="flex items-center gap-2">
            <Map className="w-5 h-5 text-amber-400" />
            <h3 className="font-semibold text-sm text-slate-200">Interactive Traversal & Continent Registry</h3>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => {
                setDiscoveredSecretCavern(prev => !prev);
                setSpawnerLogs(prev => [{
                  id: Math.random().toString(),
                  text: !discoveredSecretCavern
                    ? "Discovered Ancient Jötunn Runes. Ymirr's Hidden Cavern [RESTRICTED] is now traversable."
                    : "Cleared discoveries. Ymirr's Hidden Cavern is now RESTRICTED and hidden again.",
                  type: "warning"
                }, ...prev.slice(0, 49)]);
              }}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${
                discoveredSecretCavern
                  ? "bg-emerald-500/20 text-emerald-400 border-emerald-500/30"
                  : "bg-slate-900 text-slate-400 border-slate-800 hover:text-slate-300"
              }`}
            >
              {discoveredSecretCavern ? <Unlock className="w-3 h-3" /> : <Lock className="w-3 h-3" />}
              {discoveredSecretCavern ? "Uncovered Secret Passage" : "Unlock Secret Passage"}
            </button>
            <span className="text-xs text-slate-500 font-mono">No Hard level-gates enforced</span>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
          {/* Map canvas simulation */}
          <div className="lg:col-span-8 bg-slate-950 rounded-xl border border-slate-800 relative min-h-[380px] p-4 flex flex-col justify-between overflow-hidden">
            {/* Grid background */}
            <div className="absolute inset-0 bg-[linear-gradient(to_right,#1e293b_1px,transparent_1px),linear-gradient(to_bottom,#1e293b_1px,transparent_1px)] bg-[size:32px_32px] opacity-[0.15]" />
            
            {/* Continent Boundaries */}
            <div className="absolute top-8 left-10 text-[10px] text-slate-600 font-mono tracking-widest border-l border-slate-800/80 pl-2">ICE CONTINENT (ELARISHEIM)</div>
            <div className="absolute top-8 right-10 text-[10px] text-slate-600 font-mono tracking-widest border-r border-slate-800/80 pr-2">NATURE CONTINENT</div>
            <div className="absolute bottom-8 left-10 text-[10px] text-slate-600 font-mono tracking-widest border-l border-slate-800/80 pl-2">MAIN CONTINENT (SANDBOX)</div>
            <div className="absolute bottom-8 right-10 text-[10px] text-slate-600 font-mono tracking-widest border-r border-slate-800/80 pr-2">FIRE CONTINENT (VOLCANIC)</div>

            {/* Travel network SVG overlay */}
            <svg className="absolute inset-0 w-full h-full pointer-events-none z-0">
              {/* Main Continent routes */}
              <line x1="15%" y1="70%" x2="22%" y2="56%" stroke="#475569" strokeWidth="1" strokeDasharray="3 3" />
              <line x1="22%" y1="56%" x2="30%" y2="72%" stroke="#475569" strokeWidth="1" strokeDasharray="3 3" />
              {/* Continental links */}
              <line x1="22%" y1="56%" x2="18%" y2="24%" stroke="#334155" strokeWidth="1" />
              <line x1="30%" y1="72%" x2="45%" y2="30%" stroke="#334155" strokeWidth="1" />
              <line x1="45%" y1="30%" x2="55%" y2="48%" stroke="#334155" strokeWidth="1" />
              <line x1="55%" y1="48%" x2="77%" y2="64%" stroke="#334155" strokeWidth="1" />
              <line x1="77%" y1="64%" x2="82%" y2="76%" stroke="#334155" strokeWidth="1" />
              <line x1="82%" y1="76%" x2="87%" y2="58%" stroke="#334155" strokeWidth="1" />
              <line x1="18%" y1="24%" x2="23%" y2="18%" stroke="#0284c7" strokeWidth="1.5" strokeDasharray="4 2" opacity={discoveredSecretCavern ? 1 : 0.2} />
            </svg>

            {/* Region Interactive Nodes */}
            <div className="relative w-full h-full min-h-[300px] z-10">
              {REGIONS.map((region) => {
                const isSelected = selectedRegionId === region.id;
                const isRestricted = !!region.accessRestriction;
                const showCavern = !isRestricted || discoveredSecretCavern;

                if (isRestricted && !showCavern) return null;

                let nodeStyle = "bg-slate-800 border-slate-600 text-slate-300";
                if (isSelected) nodeStyle = "bg-amber-500 border-amber-300 text-slate-950 scale-110 shadow-lg shadow-amber-500/20";
                else if (isRestricted) nodeStyle = "bg-sky-500/20 border-sky-400/60 text-sky-300";

                return (
                  <button
                    key={region.id}
                    onClick={() => handleRegionClick(region)}
                    className={`absolute flex flex-col items-center group transition-all duration-300 ${nodeStyle}`}
                    style={{ left: `${region.centerCoordinates.x / 10}%`, top: `${region.centerCoordinates.y / 5}%` }}
                    id={`map-node-${region.id}`}
                  >
                    <div className="p-1.5 rounded-full border bg-inherit flex items-center justify-center">
                      <MapPin className="w-3.5 h-3.5" />
                    </div>
                    <span className="absolute top-7 bg-slate-900/90 border border-slate-800 text-[9px] font-semibold tracking-wider px-1.5 py-0.5 rounded shadow-md whitespace-nowrap opacity-80 group-hover:opacity-100 transition-opacity z-20">
                      {region.name} <span className="text-slate-400">Lv.{region.minLevel}+</span>
                    </span>
                  </button>
                );
              })}
            </div>

            <div className="flex justify-between items-center text-[10px] text-slate-500 font-mono mt-2 pt-2 border-t border-slate-900">
              <span>BOUNDS: [Min 0,0] to [Max 9999,9999]</span>
              <span>Click nodes to travel and set active biome modifiers</span>
            </div>
          </div>

          {/* Region and Continent Details */}
          <div className="lg:col-span-4 flex flex-col gap-4">
            <div className="bg-slate-950 rounded-xl border border-slate-800/80 p-4 h-full flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between text-xs text-slate-500 font-mono mb-2">
                  <span>ACTIVE REGION SUMMARY</span>
                  <span className="text-amber-400 font-semibold uppercase">{activeRegion.continent}</span>
                </div>
                <h4 className="text-lg font-bold text-slate-100 flex items-center gap-2">
                  {activeRegion.name}
                  {activeRegion.accessRestriction && (
                    <span className="text-[9px] bg-sky-500/20 text-sky-300 px-1.5 py-0.5 rounded font-mono font-bold uppercase border border-sky-500/30">
                      Hidden
                    </span>
                  )}
                </h4>
                <p className="text-xs text-slate-400 mt-2 italic leading-relaxed">
                  "Emergency danger ratings on {activeRegion.name} are dynamic. High-density threat groups patrol POIs like the {activeRegion.pointsOfInterest[0]}."
                </p>

                <div className="grid grid-cols-2 gap-3 mt-4 pt-4 border-t border-slate-900">
                  <div className="bg-slate-900/60 p-2.5 rounded-lg border border-slate-800">
                    <span className="text-[10px] text-slate-500 block">Level Bracket</span>
                    <span className="text-sm font-bold text-slate-200">
                      Lv.{activeRegion.minLevel} - {activeRegion.maxLevel}
                    </span>
                  </div>
                  <div className="bg-slate-900/60 p-2.5 rounded-lg border border-slate-800">
                    <span className="text-[10px] text-slate-500 block">Grid Center</span>
                    <span className="text-sm font-mono font-bold text-slate-200">
                      X: {activeRegion.centerCoordinates.x * 10} Y: {activeRegion.centerCoordinates.y * 10}
                    </span>
                  </div>
                </div>

                <div className="mt-4">
                  <span className="text-[10px] text-slate-500 block mb-1">Local Points of Interest</span>
                  <div className="flex flex-wrap gap-1.5">
                    {activeRegion.pointsOfInterest.map((poi, index) => (
                      <span key={index} className="text-[10px] bg-slate-900 text-slate-300 px-2 py-1 rounded-md border border-slate-800">
                        📍 {poi}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="mt-4">
                  <span className="text-[10px] text-slate-500 block mb-1">Local Families (Monster Bible)</span>
                  <div className="flex flex-wrap gap-1.5">
                    {activeRegion.nativeMonsterFamilies.map((fam, index) => (
                      <span key={index} className="text-[10px] bg-slate-900 text-slate-400 px-2.5 py-0.5 rounded-full border border-slate-800">
                        {fam}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              {activeRegion.accessRestriction && (
                <div className="bg-sky-500/10 border border-sky-500/20 rounded-lg p-3 text-[11px] text-sky-400 mt-4 leading-normal flex gap-2">
                  <Info className="w-4 h-4 shrink-0 mt-0.5" />
                  <span>
                    <strong>Movement Gate:</strong> This region is hidden. Access requires navigating frozen crags or underground passages.
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Main Grid: Biomes, Adaptive Spawn, Loot Roll */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6" id="world-bible-main-grid">
        
        {/* Column 1: Biomes Selector (3 Cols) */}
        <div className="lg:col-span-3 flex flex-col gap-4">
          <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 h-full flex flex-col gap-3">
            <div className="flex items-center gap-2 border-b border-slate-800/60 pb-3">
              <Compass className="w-5 h-5 text-amber-400" />
              <h3 className="font-semibold text-sm text-slate-200">Biome Environment Selector</h3>
            </div>

            <div className="flex flex-col gap-2 overflow-y-auto max-h-[500px] pr-1">
              {BIOMES.map((biome) => {
                const isSelected = selectedBiomeId === biome.id;
                return (
                  <button
                    key={biome.id}
                    onClick={() => setSelectedBiomeId(biome.id)}
                    className={`text-left p-3 rounded-xl border transition-all duration-200 relative group overflow-hidden ${
                      isSelected
                        ? `${biome.borderColor} bg-slate-900/80 ring-1 ring-amber-500/30`
                        : "border-slate-800/80 hover:border-slate-700 hover:bg-slate-900/20"
                    }`}
                    id={`biome-btn-${biome.id}`}
                  >
                    {/* Color Glow accent */}
                    {isSelected && (
                      <div className="absolute top-0 right-0 w-16 h-16 rounded-full blur-xl opacity-30 pointer-events-none" style={{ backgroundColor: biome.elementalAffinity === "Fire" ? "#ef4444" : biome.elementalAffinity === "Ice" ? "#0ea5e9" : biome.elementalAffinity === "Shadow" ? "#8b5cf6" : "#10b981" }} />
                    )}

                    <div className="flex justify-between items-center">
                      <span className="text-xs font-bold text-slate-100">{biome.name}</span>
                      <span className={`text-[9px] px-1.5 py-0.5 rounded font-mono font-bold uppercase ${biome.color}`}>
                        {biome.elementalAffinity}
                      </span>
                    </div>
                    <p className="text-[11px] text-slate-400 mt-1 line-clamp-2 leading-relaxed">
                      {biome.description}
                    </p>

                    <div className="flex justify-between items-center text-[9px] text-slate-500 font-mono mt-2">
                      <span>HAZARD: {(biome.hazardLevel * 100).toFixed(0)}%</span>
                      <span>TTK MULT: {biome.difficultyMultiplier}x</span>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        {/* Column 2: Adaptive Spawn Ecosystem Engine (5 Cols) */}
        <div className="lg:col-span-5 flex flex-col gap-4">
          <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 flex flex-col gap-4 h-full">
            <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
              <div className="flex items-center gap-2">
                <Sliders className="w-5 h-5 text-amber-400" />
                <h3 className="font-semibold text-sm text-slate-200">Adaptive Spawn & Encounter Engine</h3>
              </div>
              <button
                onClick={simulateSpawn}
                className="p-1 bg-slate-950 border border-slate-800 rounded-lg hover:border-slate-700 text-slate-400 hover:text-slate-200 transition-all"
                title="Force Respawn"
              >
                <RefreshCw className="w-3.5 h-3.5" />
              </button>
            </div>

            {/* Sliders Console */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                <div className="flex justify-between text-xs text-slate-400 mb-1">
                  <span className="flex items-center gap-1">
                    <Activity className="w-3 h-3 text-rose-400" />
                    Player Activity Score
                  </span>
                  <span className="font-mono text-amber-400 font-bold">{playerActivity}</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="20"
                  value={playerActivity}
                  onChange={(e) => setPlayerActivity(parseInt(e.target.value))}
                  className="w-full accent-amber-500 h-1 bg-slate-800 rounded-lg appearance-none cursor-pointer"
                  id="activity-slider"
                />
                <div className={`text-[10px] font-mono px-2 py-0.5 rounded border text-center font-bold mt-2 ${activityTier.color}`}>
                  Activity: {activityTier.tier}
                </div>
              </div>

              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                <div className="flex justify-between text-xs text-slate-400 mb-1">
                  <span>Regional Danger Level</span>
                  <span className="font-mono text-amber-400 font-bold">{(dangerLevel * 100).toFixed(0)}%</span>
                </div>
                <input
                  type="range"
                  min="10"
                  max="100"
                  value={dangerLevel * 100}
                  onChange={(e) => setDangerLevel(parseFloat(e.target.value) / 100)}
                  className="w-full accent-amber-500 h-1 bg-slate-800 rounded-lg appearance-none cursor-pointer"
                  id="danger-slider"
                />
                <div className="flex justify-between text-[9px] text-slate-500 font-mono mt-1">
                  <span>0.1 (Starter)</span>
                  <span>1.0 (Endgame)</span>
                </div>
              </div>
            </div>

            <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 flex flex-col gap-2">
              <span className="text-[10px] text-slate-500 block font-mono">ENCOUNTER COMPOSTION SELECTOR</span>
              <div className="grid grid-cols-4 gap-2">
                {(["solo", "pack", "horde", "boss_encounter"] as const).map((type) => (
                  <button
                    key={type}
                    onClick={() => setEncounterType(type)}
                    className={`py-1.5 px-1 rounded-lg text-[10px] font-bold border transition-all ${
                      encounterType === type
                        ? "bg-amber-500/10 text-amber-400 border-amber-500/40"
                        : "bg-slate-900 text-slate-400 border-slate-800 hover:text-slate-300"
                    }`}
                    id={`encounter-btn-${type}`}
                  >
                    {type === "solo" && "Solo 1x"}
                    {type === "pack" && "Pack 3-5x"}
                    {type === "horde" && "Horde 10x"}
                    {type === "boss_encounter" && "Raid Boss"}
                  </button>
                ))}
              </div>
            </div>

            {/* Computed Math Stats Grid */}
            <div className="grid grid-cols-3 gap-3">
              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800 text-center">
                <span className="text-[9px] text-slate-500 block uppercase font-mono">Spawn Timer</span>
                <span className="text-sm font-bold text-slate-200 mt-1 block">{dynamicSpawnRate}s</span>
                <span className="text-[8px] text-slate-500 block">Dynamic Rate</span>
              </div>
              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800 text-center">
                <span className="text-[9px] text-slate-500 block uppercase font-mono">Elite Ratio</span>
                <span className="text-sm font-bold text-amber-400 mt-1 block">{eliteProbability}%</span>
                <span className="text-[8px] text-slate-500 block">Promotion Chance</span>
              </div>
              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800 text-center">
                <span className="text-[9px] text-slate-500 block uppercase font-mono">Threat Coef</span>
                <span className="text-sm font-bold text-rose-400 mt-1 block">{threatMultiplier}x</span>
                <span className="text-[8px] text-slate-500 block">Dmg/HP Multiplier</span>
              </div>
            </div>

            {/* Visual Arena Map */}
            <div className="bg-slate-950 rounded-xl border border-slate-800 p-4 relative min-h-[160px] flex items-center justify-center overflow-hidden">
              <div className="absolute inset-0 bg-[radial-gradient(#1e293b_1px,transparent_1px)] bg-[size:16px_16px] opacity-10" />
              
              {/* Active Biome glow ring */}
              <div className="absolute w-28 h-28 rounded-full border border-slate-800 bg-slate-900/20 flex items-center justify-center animate-pulse">
                <span className="text-[10px] text-slate-600 uppercase tracking-widest font-mono">Aggro Core</span>
              </div>

              {/* Render simulated mobs */}
              {simulatedEnemies.map((mob) => {
                const x = Math.cos((mob.angle * Math.PI) / 180) * mob.dist;
                const y = Math.sin((mob.angle * Math.PI) / 180) * mob.dist;

                return (
                  <motion.div
                    key={mob.id}
                    initial={{ scale: 0, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    className="absolute z-10 flex flex-col items-center pointer-events-none"
                    style={{ transform: `translate(${x}px, ${y}px)` }}
                  >
                    <div className={`w-3.5 h-3.5 rounded-full flex items-center justify-center border text-[8px] ${
                      mob.isElite
                        ? "bg-amber-500 border-amber-300 shadow-md shadow-amber-500/30"
                        : "bg-slate-600 border-slate-400"
                    }`}>
                      {mob.isElite ? "★" : "●"}
                    </div>
                    <span className="text-[8px] bg-slate-900/90 text-slate-300 px-1 rounded border border-slate-800 mt-1 whitespace-nowrap scale-90">
                      {mob.name}
                    </span>
                  </motion.div>
                );
              })}

              {/* Active environmental hazards indicators */}
              <div className="absolute bottom-2 left-2 flex flex-col gap-1 z-20">
                {activeBiome.environmentalHazards.map((haz, idx) => (
                  <span key={idx} className="text-[8px] bg-red-500/10 text-red-400 border border-red-500/20 px-1.5 py-0.5 rounded font-semibold font-mono">
                    ⚠️ {haz}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Column 3: Loot Weighting & Progression (4 Cols) */}
        <div className="lg:col-span-4 flex flex-col gap-4">
          <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 flex flex-col justify-between h-full">
            <div className="flex flex-col gap-4">
              <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
                <div className="flex items-center gap-2">
                  <Gift className="w-5 h-5 text-amber-400" />
                  <h3 className="font-semibold text-sm text-slate-200">Dynamic Biome Loot Engine</h3>
                </div>
                <button
                  onClick={rollLoot}
                  className="bg-amber-500 text-slate-950 font-bold text-[10px] px-3 py-1.5 rounded-lg border border-amber-400 hover:bg-amber-400 transition-all flex items-center gap-1 cursor-pointer"
                  id="roll-loot-btn"
                >
                  <Sword className="w-3 h-3" />
                  Roll Drop
                </button>
              </div>

              {/* Current Loot pool bias list */}
              <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 text-[11px] text-slate-400 leading-normal flex flex-col gap-2">
                <span className="text-[10px] text-slate-500 font-mono block">ACTIVE LOOT WEIGHTS FOR {activeBiome.name.toUpperCase()}</span>
                <div className="flex flex-col gap-1">
                  <div className="flex justify-between border-b border-slate-900 py-1">
                    <span>Base Rarity Coefficient</span>
                    <span className="text-amber-400 font-mono font-semibold">{(activeBiome.lootModifiers.rarityBonus * (1 + dangerLevel * 0.5)).toFixed(2)}x</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-900 py-1">
                    <span>Weighted Core Affix</span>
                    <span className="text-emerald-400 font-mono font-semibold">
                      {Object.keys(activeBiome.lootModifiers.affixWeights)[0].replace("_", " ")}
                    </span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span>Target Materials Pool</span>
                    <span className="text-slate-200 font-semibold">{activeBiome.lootModifiers.materialsAvailable[0]}</span>
                  </div>
                </div>
              </div>

              {/* Rolled Items List */}
              <div className="flex flex-col gap-2 max-h-[220px] overflow-y-auto pr-1">
                {lootDrops.length === 0 ? (
                  <div className="text-center py-8 text-xs text-slate-500 italic">
                    Click "Roll Drop" to generate item templates simulated in this biome...
                  </div>
                ) : (
                  <AnimatePresence>
                    {lootDrops.map((loot, idx) => (
                      <motion.div
                        key={idx}
                        initial={{ opacity: 0, y: -5 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0 }}
                        className={`p-2.5 rounded-lg border flex justify-between items-center ${loot.itemColor}`}
                      >
                        <div>
                          <div className="flex items-center gap-1.5">
                            <span className="text-xs font-bold leading-none">{loot.name}</span>
                            <span className="text-[8px] uppercase tracking-wider font-mono font-bold bg-slate-950/40 px-1 py-0.5 rounded text-slate-400">
                              iLvl {loot.ilvl}
                            </span>
                          </div>
                          <span className="text-[10px] text-slate-400 block mt-1">{loot.stats}</span>
                        </div>
                        <div className="text-right">
                          <span className="text-[10px] text-amber-400 font-mono font-semibold">{loot.biomeAffix}</span>
                        </div>
                      </motion.div>
                    ))}
                  </AnimatePresence>
                )}
              </div>
            </div>

            {/* Quick Progression Integration view */}
            <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 mt-3 flex flex-col gap-2">
              <div className="flex justify-between items-center border-b border-slate-900 pb-1.5 mb-1.5">
                <span className="text-[10px] text-slate-500 font-mono uppercase">Progression Bible Sync</span>
                <span className="text-[9px] text-slate-500 font-mono">NON-LINEAR XP</span>
              </div>
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-1.5">
                  <span className="text-xs text-slate-400">Player Level</span>
                  <input
                    type="number"
                    min="1"
                    max="60"
                    value={playerLevel}
                    onChange={(e) => setPlayerLevel(Math.max(1, Math.min(60, parseInt(e.target.value) || 1)))}
                    className="bg-slate-900 border border-slate-800 rounded w-10 text-center text-xs text-amber-400 font-bold py-0.5"
                    id="player-level-input"
                  />
                </div>
                <div className="text-right">
                  <span className="text-[9px] text-slate-500 block uppercase">Underdog XP Bonus</span>
                  <span className="text-xs font-mono font-bold text-emerald-400">{underdogXPMultiplier}x XP</span>
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>

      {/* Spawner Logs console */}
      <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 flex flex-col gap-3" id="spawner-logs-card">
        <div className="flex justify-between items-center border-b border-slate-800/60 pb-3">
          <h3 className="font-semibold text-sm text-slate-200">World Engine Event Logs</h3>
          <button
            onClick={() => setSpawnerLogs([])}
            className="text-[10px] text-slate-500 hover:text-slate-300 transition-colors uppercase font-mono"
          >
            Clear logs
          </button>
        </div>
        <div className="bg-slate-950 rounded-xl border border-slate-900 p-3 h-32 overflow-y-auto font-mono text-xs flex flex-col gap-1 pr-1">
          {spawnerLogs.length === 0 ? (
            <div className="text-slate-600 italic">Logs are quiet... Interact with the controls or map to generate world event streams.</div>
          ) : (
            spawnerLogs.map((log) => (
              <div key={log.id} className="flex gap-2">
                <span className="text-slate-600">[Engine]</span>
                <span className={log.type === "error" ? "text-red-400" : log.type === "warning" ? "text-amber-400" : log.type === "loot" ? "text-violet-400" : "text-slate-400"}>
                  {log.text}
                </span>
              </div>
            ))
          )}
        </div>
      </div>

      {/* System Integration Graph and Architecture Map */}
      <div className="bg-slate-900/40 rounded-2xl border border-slate-800/80 p-5 flex flex-col gap-5" id="world-systems-integration">
        <div className="flex items-center gap-2 border-b border-slate-800/60 pb-3">
          <TrendingUp className="w-5 h-5 text-amber-400" />
          <h3 className="font-semibold text-sm text-slate-200">Emergent Level Difficulty Scaling & XP Multiplier Graph</h3>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* SVG Line Graph */}
          <div className="lg:col-span-7 bg-slate-950 rounded-xl border border-slate-800 p-4 flex flex-col justify-between">
            <span className="text-[10px] text-slate-500 font-mono block mb-3">CURVE PLOT: Underdog XP Boost vs Player level (Target recommended: Lv.{recommendedLevel})</span>
            
            <div className="relative w-full h-[180px]">
              {/* Plot graph lines using SVG path */}
              <svg className="w-full h-full overflow-visible">
                {/* Horizontal grid lines */}
                <line x1="0" y1="20%" x2="100%" y2="20%" stroke="#1e293b" strokeWidth="1" />
                <line x1="0" y1="50%" x2="100%" y2="50%" stroke="#1e293b" strokeWidth="1" />
                <line x1="0" y1="80%" x2="100%" y2="80%" stroke="#1e293b" strokeWidth="1" />
                
                {/* Curve generator */}
                {(() => {
                  const points = [];
                  for (let i = 1; i <= 60; i++) {
                    const mult = i < recommendedLevel ? Math.min(4.0, Math.pow(recommendedLevel / i, 1.5)) : 1.0;
                    // Map i (1-60) to X (0-100%)
                    const xPct = ((i - 1) / 59) * 100;
                    // Map mult (1-4) to Y (80% to 10%)
                    const yPct = 80 - ((mult - 1) / 3.0) * 70;
                    points.push(`${xPct}%,${yPct}%`);
                  }
                  return (
                    <>
                      {/* Underdog multiplier curve fill */}
                      <path
                        d={`M 0,100% ${points.map((pt, idx) => `L ${pt}`).join(" ")} L 100%,100% Z`}
                        fill="rgba(16, 185, 129, 0.05)"
                      />
                      {/* Underdog multiplier curve line */}
                      <polyline
                        points={points.map(pt => pt.replace(/%/g, "")).join(" ")}
                        fill="none"
                        stroke="#10b981"
                        strokeWidth="2"
                        className="transition-all duration-300"
                        style={{ transform: "scale(1, 1)" }}
                      />
                    </>
                  );
                })()}

                {/* Current Player Indicator */}
                {(() => {
                  const mult = playerLevel < recommendedLevel ? Math.min(4.0, Math.pow(recommendedLevel / playerLevel, 1.5)) : 1.0;
                  const xPct = ((playerLevel - 1) / 59) * 100;
                  const yPct = 80 - ((mult - 1) / 3.0) * 70;
                  return (
                    <circle
                      cx={`${xPct}%`}
                      cy={`${yPct}%`}
                      r="6"
                      fill="#f59e0b"
                      stroke="#1e293b"
                      strokeWidth="2"
                    />
                  );
                })()}
              </svg>
              
              <div className="absolute top-1 left-2 text-[9px] text-slate-500 font-mono">4.0x XP Boost</div>
              <div className="absolute bottom-1 left-2 text-[9px] text-slate-500 font-mono">1.0x XP (No Boost)</div>
              <div className="absolute bottom-1 right-2 text-[9px] text-slate-500 font-mono">Lv.60</div>
              <div className="absolute bottom-1 left-12 text-[9px] text-slate-500 font-mono">Lv.1</div>
            </div>

            <div className="flex justify-between items-center text-[10px] text-slate-500 font-mono mt-2 pt-2 border-t border-slate-900">
              <span>GREEN: UNDERDOG XP CURVE</span>
              <span>AMBER DOT: CURRENT PLAYER (Lv.{playerLevel})</span>
            </div>
          </div>

          {/* System Integration matrix flow */}
          <div className="lg:col-span-5 bg-slate-950 rounded-xl border border-slate-800 p-4">
            <span className="text-[10px] text-slate-500 font-mono block mb-3">CANONICAL XP FLOW SCHEMATIC</span>
            <div className="flex flex-col gap-3 text-xs leading-normal">
              
              <div className="flex items-center gap-2">
                <div className="w-5 h-5 rounded bg-emerald-500/10 text-emerald-400 flex items-center justify-center font-mono font-bold text-[10px] border border-emerald-500/20">
                  1
                </div>
                <div>
                  <span className="font-semibold text-slate-200">Monster Slay Encounter</span>
                  <p className="text-[10px] text-slate-400">Combat Bible determines baseline TTK and active player performance shares.</p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="w-5 h-5 rounded bg-sky-500/10 text-sky-400 flex items-center justify-center font-mono font-bold text-[10px] border border-sky-500/20">
                  2
                </div>
                <div>
                  <span className="font-semibold text-slate-200">Biome Modifiers & Scaling</span>
                  <p className="text-[10px] text-slate-400">Biome Bible injects difficulty factor, elite probabilities, and regional threat scaling.</p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="w-5 h-5 rounded bg-amber-500/10 text-amber-400 flex items-center justify-center font-mono font-bold text-[10px] border border-amber-500/20">
                  3
                </div>
                <div>
                  <span className="font-semibold text-slate-200">Underdog Multiplier Applied</span>
                  <p className="text-[10px] text-slate-400">Progression Bible scales the XP payout using recommended level delta calculations.</p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="w-5 h-5 rounded bg-violet-500/10 text-violet-400 flex items-center justify-center font-mono font-bold text-[10px] border border-violet-500/20">
                  4
                </div>
                <div>
                  <span className="font-semibold text-slate-200">Loot Rarity Roll Evaluation</span>
                  <p className="text-[10px] text-slate-400">Loot Bible draws items from custom table biased by regional affixes and activity coefficients.</p>
                </div>
              </div>

            </div>
          </div>
        </div>
      </div>

    </div>
  );
}
