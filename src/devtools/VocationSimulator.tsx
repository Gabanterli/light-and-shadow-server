import React, { useState, useEffect } from "react";
import {
  Shield,
  Sword,
  Sparkles,
  CheckCircle,
  Sliders,
  Activity,
  Award,
  BookOpen,
  RefreshCw,
  Heart,
  Zap,
  ShieldAlert,
  Feather,
  FlameKindling,
  Crosshair,
  Timer,
  Scale
} from "lucide-react";
import { motion } from "motion/react";

interface Vocation {
  id: string;
  name: string;
  description: string;
  baseHp: number;
  baseMana: number;
  manaRegen: number;
  baseArmor: number;
  baseMagicResistance: number;
  baseMoveSpeed: number;
  allowedWeapons: string[];
  allowedOffhands: { id: string; name: string; archetype: string; desc: string }[];
  hpGainPerLevel: number;
  manaGainPerLevel: number;
}

const CANONICAL_VOCATIONS: Vocation[] = [
  {
    id: "knight",
    name: "Knight",
    description: "Stoic frontline warriors built for survivability, heavy defense, and high health pools.",
    baseHp: 150,
    baseMana: 40,
    manaRegen: 2.0,
    baseArmor: 25,
    baseMagicResistance: 10,
    baseMoveSpeed: 95,
    allowedWeapons: ["Sword", "Axe", "Club"],
    allowedOffhands: [
      { id: "shield", name: "Heavy Shield", archetype: "defensive", desc: "Provides high physical mitigation and direct armor blocks." },
      { id: "sword", name: "Offhand Sword", archetype: "offensive_physical", desc: "Enables physical dual-wield combos and faster combat speed." },
      { id: "empty", name: "Empty Offhand", archetype: "none", desc: "No off-hand equipped; baseline values remain unmodified." }
    ],
    hpGainPerLevel: 15,
    manaGainPerLevel: 2
  },
  {
    id: "assassin",
    name: "Assassin",
    description: "Vanish in shadows to strike vitals with high-speed burst physical and movement mechanics.",
    baseHp: 100,
    baseMana: 60,
    manaRegen: 3.0,
    baseArmor: 10,
    baseMagicResistance: 10,
    baseMoveSpeed: 120,
    allowedWeapons: ["Dagger", "Sword"],
    allowedOffhands: [
      { id: "dagger", name: "Offhand Dagger", archetype: "offensive_physical", desc: "Enables double-dagger dual wield, optimizing strike loops." },
      { id: "shield", name: "Buckler Shield", archetype: "defensive", desc: "Enables defensive PvP setups with decent physical blocks." },
      { id: "empty", name: "Empty Offhand", archetype: "none", desc: "No off-hand equipped; baseline values remain unmodified." }
    ],
    hpGainPerLevel: 10,
    manaGainPerLevel: 4
  },
  {
    id: "archer",
    name: "Archer",
    description: "Precise marksmanship experts specializing in ranged efficiency and strategic mobility.",
    baseHp: 110,
    baseMana: 70,
    manaRegen: 3.5,
    baseArmor: 12,
    baseMagicResistance: 12,
    baseMoveSpeed: 110,
    allowedWeapons: ["Bow", "Crossbow"],
    allowedOffhands: [
      { id: "quiver", name: "Elemental Quiver", archetype: "utility", desc: "Archer-exclusive utility offhand that boosts positioning speed and tactical options." }
    ],
    hpGainPerLevel: 11,
    manaGainPerLevel: 5
  },
  {
    id: "mage",
    name: "Mage",
    description: "Commanders of raw elementals, trading physical resilience for catastrophic magical damage.",
    baseHp: 80,
    baseMana: 180,
    manaRegen: 8.0,
    baseArmor: 5,
    baseMagicResistance: 25,
    baseMoveSpeed: 100,
    allowedWeapons: ["Spell Staff"],
    allowedOffhands: [
      { id: "shield", name: "Kite Shield", archetype: "defensive", desc: "Enables robust physical mitigation at the expense of magical speed." },
      { id: "spellbook", name: "Grimoire Spellbook", archetype: "offensive_magical", desc: "Spellbooks provide magical amplification and lower physical defense than shields." }
    ],
    hpGainPerLevel: 5,
    manaGainPerLevel: 15
  },
  {
    id: "cleric",
    name: "Cleric",
    description: "Devout support channels specializing in powerful restorative magic, divine protection, and high mana sustain.",
    baseHp: 120,
    baseMana: 130,
    manaRegen: 6.0,
    baseArmor: 15,
    baseMagicResistance: 20,
    baseMoveSpeed: 98,
    allowedWeapons: ["Mace", "Dagger"],
    allowedOffhands: [
      { id: "shield", name: "Round Shield", archetype: "defensive", desc: "Increases defense to absorb front-line strikes for battle builds." },
      { id: "sacred_scepter", name: "Sacred Scepter", archetype: "offensive_magical", desc: "Sacred scepters amplify holy power, healing output, and mana levels while providing less defense than shields." }
    ],
    hpGainPerLevel: 8,
    manaGainPerLevel: 11
  }
];

export function VocationSimulator() {
  const [selectedVocation, setSelectedVocation] = useState<Vocation>(CANONICAL_VOCATIONS[0]);
  const [selectedOffhand, setSelectedOffhand] = useState<string>("shield");
  const [selectedWeapon, setSelectedWeapon] = useState<string>("Sword");
  const [statFocus, setStatFocus] = useState<"defensive" | "offensive" | "support" | "hybrid">("defensive");
  const [manaAllocation, setManaAllocation] = useState<number>(30); // mana burden slider percentage
  const [characterLevel, setCharacterLevel] = useState<number>(1);

  // Helper to calculate soft-capped level progression gains for HP or Mana based on three bands:
  // Band 1 (Levels 1–200): 100% scaling efficiency
  // Band 2 (Levels 201–500): 50% scaling efficiency
  // Band 3 (Levels 501+): 25% scaling efficiency
  const getSoftCappedLevelGain = (level: number, gainPerLevel: number) => {
    const n1 = Math.max(0, Math.min(level - 1, 199));
    const n2 = Math.max(0, Math.min(level - 1 - n1, 300));
    const n3 = Math.max(0, level - 1 - n1 - n2);
    return (n1 * gainPerLevel * 1.0) + (n2 * gainPerLevel * 0.5) + (n3 * gainPerLevel * 0.25);
  };

  // Sync selected offhand when vocation changes
  useEffect(() => {
    if (selectedVocation.allowedOffhands.length > 0) {
      const exists = selectedVocation.allowedOffhands.find(o => o.id === selectedOffhand);
      if (!exists) {
        setSelectedOffhand(selectedVocation.allowedOffhands[0].id);
      }
    }
    // Sync weapon selection
    if (selectedVocation.allowedWeapons.length > 0) {
      const defaultWeapon = getVocationWeaponDefault(selectedVocation.id);
      if (selectedVocation.allowedWeapons.includes(defaultWeapon)) {
        setSelectedWeapon(defaultWeapon);
      } else {
        setSelectedWeapon(selectedVocation.allowedWeapons[0]);
      }
    }
  }, [selectedVocation]);

  // Helper to resolve starting default weapon based on vocation
  const getVocationWeaponDefault = (vocationId: string) => {
    switch (vocationId) {
      case "knight": return "Sword";
      case "assassin": return "Dagger";
      case "archer": return "Bow";
      case "mage": return "Spell Staff";
      case "cleric": return "Mace";
      default: return "";
    }
  };

  // Resolve Equipment Modifiers Dynamically Based on Choices
  const getEquipmentModifiers = () => {
    const mods = {
      // Defensive
      HP: 0,
      Armor: 0,
      MagicResistance: 0,
      // Physical Offensive
      CritChance: 0,
      CritDamage: 0,
      AttackSpeed: 0,
      ArmorPenetration: 0,
      DistanceFighting: 0,
      // Magical Offensive
      MagicLevel: 0,
      HealingPower: 0,
      HolyDamage: 0,
      FireDamage: 0,
      IceDamage: 0,
      ShadowDamage: 0,
      NatureDamage: 0,
      // Utility
      MovementSpeed: 0,
      ManaEfficiency: 0,
      CooldownReduction: 0
    };

    // 1. Weapon Modifiers
    switch (selectedWeapon) {
      case "Sword":
        mods.CritChance += 8;
        mods.CritDamage += 15;
        mods.AttackSpeed += 10;
        break;
      case "Axe":
        mods.CritDamage += 25;
        mods.ArmorPenetration += 12;
        break;
      case "Club":
        mods.ArmorPenetration += 15;
        mods.CritChance += 5;
        break;
      case "Mace":
        mods.ArmorPenetration += 8;
        mods.HealingPower += 15;
        mods.HolyDamage += 10;
        break;
      case "Dagger":
        mods.CritChance += 12;
        mods.AttackSpeed += 15;
        mods.MovementSpeed += 10;
        break;
      case "Bow":
        mods.CritChance += 10;
        mods.CritDamage += 15;
        mods.DistanceFighting += 10;
        break;
      case "Crossbow":
        mods.CritDamage += 30;
        mods.ArmorPenetration += 15;
        mods.DistanceFighting += 5;
        break;
      case "Spell Staff":
        mods.MagicLevel += 6;
        mods.FireDamage += 15;
        mods.IceDamage += 15;
        break;
    }

    // 2. Offhand Modifiers
    switch (selectedOffhand) {
      case "shield":
        mods.HP += 80;
        mods.Armor += 45;
        mods.MagicResistance += 20;
        mods.MovementSpeed -= 5;
        break;
      case "sword":
        mods.CritChance += 8;
        mods.CritDamage += 15;
        mods.AttackSpeed += 10;
        mods.ArmorPenetration += 5;
        mods.MovementSpeed += 10;
        break;
      case "dagger":
        mods.CritChance += 10;
        mods.CritDamage += 10;
        mods.AttackSpeed += 12;
        mods.ArmorPenetration += 3;
        mods.MovementSpeed += 15;
        break;
      case "spellbook":
        mods.MagicLevel += 5;
        mods.ManaEfficiency += 10;
        mods.CooldownReduction += 8;
        mods.FireDamage += 15;
        mods.IceDamage += 15;
        mods.ShadowDamage += 15;
        mods.NatureDamage += 15;
        break;
      case "sacred_scepter":
        mods.MagicLevel += 4;
        mods.HealingPower += 25;
        mods.HolyDamage += 20;
        mods.CooldownReduction += 5;
        mods.ManaEfficiency += 8;
        break;
      case "quiver":
        mods.CritChance += 12;
        mods.CritDamage += 20;
        mods.AttackSpeed += 15;
        mods.ArmorPenetration += 10;
        mods.DistanceFighting += 8;
        mods.MovementSpeed += 12;
        break;
    }

    // 3. Focus Modifier
    switch (statFocus) {
      case "defensive":
        mods.HP += 50;
        mods.Armor += 15;
        mods.MagicResistance += 10;
        break;
      case "offensive":
        mods.CritChance += 5;
        mods.CritDamage += 10;
        mods.AttackSpeed += 5;
        mods.MovementSpeed += 5;
        break;
      case "support":
        mods.MagicLevel += 2;
        mods.HealingPower += 15;
        mods.CooldownReduction += 5;
        break;
      case "hybrid":
        mods.HP += 20;
        mods.Armor += 5;
        mods.CritChance += 3;
        mods.ManaEfficiency += 5;
        break;
    }

    // 4. Mana burden effect on cooldown/damage
    const manaFactor = manaAllocation / 50; // 0.2 to 1.8
    mods.ManaEfficiency += Math.round(manaFactor * 6);

    return mods;
  };

  const eqMods = getEquipmentModifiers();

  // Determine vector statistics dynamically based on vocation, offhand, focus, and mana
  const getRoleVectors = () => {
    const voc = selectedVocation.id;
    const off = selectedOffhand;

    // Base vectors
    let survivability = 50;
    let damage = 50;
    let utility = 50;
    let sustain = 50;

    if (voc === "knight") {
      survivability = 80;
      damage = 45;
      utility = 35;
      sustain = 25;

      if (off === "shield") {
        survivability += 15;
        utility += 10;
      } else if (off === "sword") {
        survivability -= 15;
        damage += 30;
      } else {
        survivability -= 5;
        damage += 10;
      }
    }

    else if (voc === "assassin") {
      survivability = 35;
      damage = 80;
      utility = 55;
      sustain = 20;

      if (off === "dagger") {
        damage += 15;
        utility += 5;
      } else if (off === "shield") {
        survivability += 25;
        damage -= 15;
      }
    }

    else if (voc === "archer") {
      survivability = 35;
      damage = 75;
      utility = 65;
      sustain = 30;

      if (off === "quiver") {
        damage += 15;
        utility += 10;
      }
    }

    else if (voc === "mage") {
      survivability = 20;
      damage = 85;
      utility = 45;
      sustain = 55;

      if (off === "spellbook") {
        damage += 13;
        sustain += 10;
      } else if (off === "shield") {
        survivability += 45;
        damage -= 20;
      }
    }

    else if (voc === "cleric") {
      survivability = 40;
      damage = 35;
      utility = 60;
      sustain = 80;

      if (off === "sacred_scepter") {
        sustain += 15;
        utility += 10;
      } else if (off === "shield") {
        survivability += 35;
        sustain -= 10;
      }
    }

    // Apply Build Tuning (Focus) multipliers on vectors
    if (statFocus === "defensive") {
      survivability = Math.min(100, Math.round(survivability * 1.15));
      damage = Math.round(damage * 0.9);
    } else if (statFocus === "offensive") {
      damage = Math.min(100, Math.round(damage * 1.15));
      survivability = Math.round(survivability * 0.9);
    } else if (statFocus === "support") {
      utility = Math.min(100, Math.round(utility * 1.12));
      sustain = Math.min(100, Math.round(sustain * 1.15));
    } else if (statFocus === "hybrid") {
      survivability = Math.round(survivability * 1.02);
      damage = Math.round(damage * 1.02);
      utility = Math.round(utility * 1.02);
      sustain = Math.round(sustain * 1.02);
    }

    // Apply Mana allocation effects
    const manaFactor = manaAllocation / 50; // 0.2 to 1.8
    sustain = Math.min(100, Math.max(10, Math.round(sustain + (manaFactor * 8))));
    damage = Math.min(100, Math.max(10, Math.round(damage + (manaFactor * 5))));

    return {
      survivability: Math.min(100, Math.max(5, survivability)),
      damage: Math.min(100, Math.max(5, damage)),
      utility: Math.min(100, Math.max(5, utility)),
      sustain: Math.min(100, Math.max(5, sustain))
    };
  };

  const currentVectors = getRoleVectors();

  return (
    <div className="col-span-12 grid grid-cols-1 lg:grid-cols-12 gap-6" id="vocation-simulator-root">
      
      {/* Header Banner */}
      <div className="lg:col-span-12 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-0.5 rounded-full text-[9px] font-bold bg-amber-500/20 text-amber-400 border border-amber-500/30 uppercase tracking-wider">
              Canon Vocation Bible
            </span>
            <span className="text-[11px] text-emerald-400 font-mono font-bold flex items-center gap-1">
              <CheckCircle className="w-3 h-3" /> CANON STATUS: CONSOLIDATED & LOCKED
            </span>
          </div>
          <h2 className="text-lg font-bold tracking-tight text-slate-100 mt-1">
            Vocation & Class Simulator (Locked Canon Correction)
          </h2>
          <p className="text-xs text-slate-400 mt-1 max-w-3xl">
            Simulador oficial com sistema 100% livre de atributos RPG tradicionais. Separação estrutural estrita entre os <strong className="text-amber-400">Atributos Base de Vocação</strong> (HP, Mana, etc.) e o <strong className="text-violet-400">Pool de Modificadores de Equipamentos</strong> (Elemental, Ofensivo e Utilitário).
          </p>
        </div>
        <div className="flex items-center gap-2 bg-slate-950 p-2 rounded-lg border border-slate-800 text-xs font-mono">
          <Activity className="w-4 h-4 text-violet-400" />
          <span className="text-slate-400">Recurso Ativo:</span>
          <span className="text-amber-400 font-bold">MANA APENAS</span>
        </div>
      </div>

      {/* LEFT COLUMN: Vocation Selection & Loadout */}
      <div className="lg:col-span-4 space-y-6">
        
        {/* Module 1: Class Selector */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 space-y-4 shadow-xl">
          <div className="flex items-center gap-2 text-slate-300 font-bold text-xs uppercase tracking-wider border-b border-slate-800/60 pb-2">
            <Sliders className="w-4 h-4 text-amber-500" />
            <span>1. Selecionar Vocação Canônica (Estrita)</span>
          </div>
          
          <div className="grid grid-cols-1 gap-2">
            {CANONICAL_VOCATIONS.map((voc) => {
              const isSelected = selectedVocation.id === voc.id;
              return (
                <button
                  key={voc.id}
                  id={`btn-vocation-${voc.id}`}
                  onClick={() => setSelectedVocation(voc)}
                  className={`w-full text-left p-3 rounded-xl border transition-all flex justify-between items-center ${
                    isSelected
                      ? "bg-gradient-to-r from-slate-800 to-slate-800/40 border-amber-500/80 text-white shadow-md shadow-amber-500/5"
                      : "bg-slate-950/40 border-slate-900 hover:border-slate-800 text-slate-300 hover:text-white"
                  }`}
                >
                  <div className="space-y-0.5">
                    <span className="text-xs font-bold block">{voc.name}</span>
                    <span className="text-[10px] text-slate-400 line-clamp-1 pr-4">{voc.description}</span>
                  </div>
                  {isSelected && (
                    <span className="text-xs font-bold text-amber-400 shrink-0 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20 uppercase">
                      LOCKED
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          <div className="border-t border-slate-800/60 pt-4 space-y-3">
            <div className="flex justify-between items-center text-xs font-mono">
              <span className="text-slate-300 font-bold flex items-center gap-1.5">
                <Sliders className="w-3.5 h-3.5 text-amber-500" />
                Nível do Personagem (1-9999):
              </span>
              <div className="flex items-center gap-1 bg-slate-950 px-2 py-0.5 rounded border border-slate-800">
                <input
                  type="number"
                  min="1"
                  max="9999"
                  value={characterLevel}
                  onChange={(e) => {
                    const val = parseInt(e.target.value);
                    const cleanVal = Math.min(9999, Math.max(1, isNaN(val) ? 1 : val));
                    setCharacterLevel(cleanVal);
                  }}
                  className="w-12 bg-transparent text-amber-400 font-bold text-right outline-none text-xs"
                />
              </div>
            </div>
            <input
              type="range"
              min="1"
              max="200"
              value={characterLevel > 200 ? 200 : characterLevel}
              onChange={(e) => setCharacterLevel(parseInt(e.target.value))}
              className="w-full accent-amber-500 h-1 bg-slate-950 rounded-lg cursor-pointer"
            />
            <div className="flex justify-between text-[9px] text-slate-500 font-mono">
              <span>Lvl 1</span>
              <span>Lvl 200 (Slider Cap)</span>
            </div>
          </div>

        </div>

        {/* Module 2: Equipment Loadout */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 space-y-4 shadow-xl">
          <div className="flex items-center gap-2 text-slate-300 font-bold text-xs uppercase tracking-wider border-b border-slate-800/60 pb-2">
            <Sword className="w-4 h-4 text-violet-500" />
            <span>2. Configurar Equipamentos (Loadout)</span>
          </div>

          <div className="space-y-3">
            {/* Primary Weapon */}
            <div className="space-y-1.5">
              <label className="text-[10px] uppercase font-bold text-slate-400 block font-mono">Arma Primária Permitida:</label>
              <select
                id="select-weapon"
                value={selectedWeapon}
                onChange={(e) => setSelectedWeapon(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 text-xs rounded-lg p-2 text-slate-200 font-mono focus:border-amber-500 focus:ring-0"
              >
                {selectedVocation.allowedWeapons.map(w => (
                  <option key={w} value={w}>{w}</option>
                ))}
              </select>
            </div>

            {/* Offhand Slot Selection */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center">
                <label className="text-[10px] uppercase font-bold text-slate-400 block font-mono">Slot Secundário (Off-Hand):</label>
                {selectedVocation.id === "archer" && (
                  <span className="text-[8px] bg-rose-500/10 text-rose-400 px-1 rounded font-mono font-bold">Bloqueio de Escudo</span>
                )}
              </div>
              <select
                id="select-offhand"
                value={selectedOffhand}
                onChange={(e) => setSelectedOffhand(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 text-xs rounded-lg p-2 text-slate-200 font-mono focus:border-amber-500 focus:ring-0"
              >
                {selectedVocation.allowedOffhands.map(o => (
                  <option key={o.id} value={o.id}>{o.name} ({o.archetype.toUpperCase()})</option>
                ))}
              </select>
              <p className="text-[10px] text-slate-400 font-mono leading-relaxed mt-1">
                {selectedVocation.allowedOffhands.find(o => o.id === selectedOffhand)?.desc}
              </p>
            </div>
          </div>
        </div>

        {/* Module 3: Build Tuning */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 space-y-4 shadow-xl">
          <div className="flex items-center gap-2 text-slate-300 font-bold text-xs uppercase tracking-wider border-b border-slate-800/60 pb-2">
            <Award className="w-4 h-4 text-emerald-500" />
            <span>3. Afinidade & Foco do Equipamento</span>
          </div>

          <div className="space-y-4 text-xs">
            {/* Stat Focus */}
            <div className="space-y-1.5">
              <label className="text-[10px] uppercase font-bold text-slate-400 block font-mono">Foco de Melhoria de Equipamento:</label>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { id: "defensive", label: "Defensivo (+HP/+Armor)" },
                  { id: "offensive", label: "Ofensivo (+Speed)" },
                  { id: "support", label: "Suporte (+Mana/+Regen)" },
                  { id: "hybrid", label: "Híbrido (Equilibrado)" }
                ].map(item => (
                  <button
                    key={item.id}
                    id={`btn-focus-${item.id}`}
                    onClick={() => setStatFocus(item.id as any)}
                    className={`p-2 rounded border text-[10px] font-semibold text-center transition-all ${
                      statFocus === item.id
                        ? "bg-slate-800 border-emerald-500 text-emerald-300"
                        : "bg-slate-950/40 border-slate-900 hover:border-slate-800 text-slate-400"
                    }`}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Active Mana Burden Slider */}
            <div className="space-y-2">
              <div className="flex justify-between items-center text-[10px] font-mono">
                <span className="text-slate-400">Drenagem de Mana por Segundo:</span>
                <span className="text-amber-400 font-bold">{manaAllocation}% de Carga</span>
              </div>
              <input
                type="range"
                min="10"
                max="90"
                value={manaAllocation}
                onChange={(e) => setManaAllocation(parseInt(e.target.value))}
                className="w-full accent-amber-500 h-1 bg-slate-950 rounded-lg cursor-pointer"
              />
              <span className="text-[9px] text-slate-500 block leading-relaxed italic">
                Regra Canônica: Todas as habilidades e feitiços drenam a reserva única de Mana. Rage, Stamina e Energy são abolidos.
              </span>
            </div>
          </div>
        </div>

      </div>

      {/* RIGHT COLUMN: Emergent Performance Vectors & Canonical Allowed Base Stats */}
      <div className="lg:col-span-8 space-y-6">
        
        {/* Dynamic Vector Outputs */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 shadow-xl space-y-6">
          <div className="flex justify-between items-start">
            <div className="space-y-1">
              <span className="text-[9px] uppercase font-bold text-slate-400 font-mono tracking-wider block">Vetores de Performance de Combate (Emergência de Papéis):</span>
              <h3 className="text-lg font-black text-amber-400 flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-amber-400 animate-pulse" />
                VETORES NUMÉRICOS DE PAPEL DE COMBATE
              </h3>
            </div>
            <span className="px-3 py-1 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20 text-xs font-mono font-extrabold uppercase">
              VECTOR OUTPUT MODE ONLY
            </span>
          </div>

          <p className="text-xs text-slate-300 bg-slate-950/60 p-3 rounded-lg border border-slate-900 leading-relaxed">
            Nenhum rótulo estático de classe (como Tank, DPS ou Healer) é usado. Os perfis e a aptidão de combate são mapeados exclusivamente através de tendências numéricas emergentes abaixo.
          </p>

          {/* Performance Vectors */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[
              { label: "Survivability", value: currentVectors.survivability, color: "bg-emerald-500", icon: Heart, desc: "Resiliência física e mágica geral baseada em HP e mitigação." },
              { label: "Damage Output", value: currentVectors.damage, color: "bg-rose-500", icon: FlameKindling, desc: "Poder ofensivo físico, crítico, à distância ou elemental." },
              { label: "Utility", value: currentVectors.utility, color: "bg-indigo-500", icon: Feather, desc: "Capacidade tática, mobilidade adicional e velocidade de posicionamento." },
              { label: "Sustain", value: currentVectors.sustain, color: "bg-violet-500", icon: Zap, desc: "Eficiência de cura holy, velocidade de regen e conservação de mana." }
            ].map(vector => (
              <div key={vector.label} className="space-y-1.5 bg-slate-950/50 p-3 rounded-lg border border-slate-900/50">
                <div className="flex justify-between items-center text-xs">
                  <span className="font-bold text-slate-300 flex items-center gap-1.5 font-mono">
                    <vector.icon className="w-3.5 h-3.5 text-slate-400" />
                    {vector.label}
                  </span>
                  <span className="font-mono font-black text-amber-400 text-sm">{vector.value} / 100</span>
                </div>
                <div className="w-full bg-slate-950 h-2 rounded-full overflow-hidden border border-slate-800">
                  <div className={`${vector.color} h-2 transition-all duration-300`} style={{ width: `${vector.value}%` }} />
                </div>
                <span className="text-[9.5px] text-slate-500 leading-normal block">{vector.desc}</span>
              </div>
            ))}
          </div>

          {/* CLASS CORE BASE STATS PANEL (STRICTLY LIMITED - 6 CORE STATS) */}
          <div className="border-t border-slate-800/60 pt-5 space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-[10px] font-bold uppercase text-amber-400 font-mono block">
                A) Vocation Base Stats (Strictly Limited)
              </span>
              <span className="text-[9px] bg-amber-500/10 text-amber-400 border border-amber-500/20 px-2 py-0.5 rounded font-mono font-bold">
                CLASS STRUCTURE ONLY
              </span>
            </div>
            
            <p className="text-[11px] text-slate-400 leading-normal">
              Estes parâmetros estruturais são definidos rigidamente pela classe selecionada e <strong className="text-slate-300">não possuem atributos primários (como Strength ou Dexterity)</strong> de escalonamento.
            </p>

            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-center">
              {[
                { label: "Base HP (Lvl 1)", val: selectedVocation.baseHp },
                { label: "Base Mana (Lvl 1)", val: selectedVocation.baseMana },
                { label: "Mana Regen", val: `${selectedVocation.manaRegen}/s` },
                { label: "Base Armor", val: selectedVocation.baseArmor },
                { label: "Base Magic Resistance", val: selectedVocation.baseMagicResistance },
                { label: "Base Move Speed", val: selectedVocation.baseMoveSpeed }
              ].map(item => (
                <div key={item.label} className="bg-slate-950/60 p-3 rounded-lg border border-slate-900 font-mono space-y-1">
                  <span className="text-[9px] text-slate-500 block uppercase font-bold leading-tight">{item.label}</span>
                  <span className="text-sm font-black text-slate-100">{item.val}</span>
                </div>
              ))}
            </div>
          </div>

          {/* ⭐ ACTIVE RESOURCE SCALING AUDIT PANEL (STRICT CANONICAL FORMULAS) */}
          <div className="border-t border-slate-800/60 pt-5 space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-[10px] font-bold uppercase text-emerald-400 font-mono block">
                ⭐ Active Resource Scaling Audit (HP & Mana Progression)
              </span>
              <span className="text-[9px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded font-mono font-bold">
                CANONICAL CALCULATION
              </span>
            </div>

            <p className="text-[11px] text-slate-400 leading-normal">
              Conforme as fórmulas do <strong className="text-emerald-400">Progression Bible</strong>, a vida (HP) e mana finais são calculadas pela combinação do Valor Base (Lvl 1), do Crescimento por Nível e dos Bônus de Equipamento.
            </p>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs font-mono text-slate-300 border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 text-slate-500 text-[10px] uppercase">
                    <th className="py-2 px-3">Recurso</th>
                    <th className="py-2 px-3 text-right">Valor Base (Lvl 1)</th>
                    <th className="py-2 px-3 text-right">Crescimento (Nível {characterLevel})</th>
                    <th className="py-2 px-3 text-right">Equipamentos (Loadout)</th>
                    <th className="py-2 px-3 text-right text-amber-400 font-bold">Total Final</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/40">
                  {/* HP Row */}
                  <tr>
                    <td className="py-3 px-3 font-bold flex items-center gap-1.5 text-slate-200">
                      <Heart className="w-4 h-4 text-emerald-500 shrink-0" />
                      Pontos de Vida (HP)
                    </td>
                    <td className="py-3 px-3 text-right text-slate-400">
                      {selectedVocation.baseHp}
                    </td>
                    <td className="py-3 px-3 text-right text-emerald-400">
                      +{ getSoftCappedLevelGain(characterLevel, selectedVocation.hpGainPerLevel) }
                      <span className="text-[9px] text-slate-500 block">({selectedVocation.hpGainPerLevel}/lvl base)</span>
                    </td>
                    <td className="py-3 px-3 text-right text-violet-400">
                      +{eqMods.HP}
                    </td>
                    <td className="py-3 px-3 text-right font-black text-amber-400 text-sm">
                      {selectedVocation.baseHp + getSoftCappedLevelGain(characterLevel, selectedVocation.hpGainPerLevel) + eqMods.HP}
                    </td>
                  </tr>

                  {/* Mana Row */}
                  <tr>
                    <td className="py-3 px-3 font-bold flex items-center gap-1.5 text-slate-200">
                      <Zap className="w-4 h-4 text-violet-500 shrink-0" />
                      Pontos de Mana (Mana)
                    </td>
                    <td className="py-3 px-3 text-right text-slate-400">
                      {selectedVocation.baseMana}
                    </td>
                    <td className="py-3 px-3 text-right text-emerald-400">
                      +{ getSoftCappedLevelGain(characterLevel, selectedVocation.manaGainPerLevel) }
                      <span className="text-[9px] text-slate-500 block">({selectedVocation.manaGainPerLevel}/lvl base)</span>
                    </td>
                    <td className="py-3 px-3 text-right text-violet-400">
                      +0
                    </td>
                    <td className="py-3 px-3 text-right font-black text-amber-400 text-sm">
                      {selectedVocation.baseMana + getSoftCappedLevelGain(characterLevel, selectedVocation.manaGainPerLevel)}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div className="p-3 bg-emerald-500/5 rounded-xl border border-emerald-500/10 text-[10px] text-slate-400 leading-relaxed font-mono space-y-1">
              <div className="text-emerald-400 font-bold uppercase text-[11px] flex items-center gap-1">
                <CheckCircle className="w-3.5 h-3.5" /> FÓRMULA OFICIAL DE PROGRESSÃO COM SOFT-CAP ATIVO
              </div>
              <div className="text-slate-300">Final HP = BaseHP + SoftCappedLevelHP + Equipment_HP</div>
              <div className="text-slate-300">Final Mana = BaseMana + SoftCappedLevelMana</div>
              <div className="text-[9px] text-slate-500 border-t border-slate-800/60 pt-1 mt-1">
                * Soft-Cap de Nível: Lvl 1–200 (100% de ganho), Lvl 201–500 (50% de ganho), Lvl 501+ (25% de ganho).
              </div>
            </div>
          </div>

          {/* EQUIPMENT MODIFIER SYSTEM PANEL (EXPANDED SYSTEM - OFFENSIVE / DEFENSIVE / ELEMENTAL / UTILITY) */}
          <div className="border-t border-slate-800/60 pt-5 space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-[10px] font-bold uppercase text-violet-400 font-mono block">
                B) Active Equipment Modifiers (Expanded System)
              </span>
              <span className="text-[9px] bg-violet-500/10 text-violet-400 border border-violet-500/20 px-2 py-0.5 rounded font-mono font-bold">
                GEAR-DRIVEN DYNAMICS
              </span>
            </div>

            <p className="text-[11px] text-slate-400 leading-normal">
              O equipamento ativo (Arma Primária + Off-hand + Sincronização) introduz modificadores cumulativos que definem sua identidade ofensiva e defensiva real.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs font-mono">
              {/* Defensive & Utility modifiers */}
              <div className="space-y-3 bg-slate-950/50 p-4 rounded-xl border border-slate-900">
                <h4 className="text-[10px] text-emerald-400 font-bold uppercase border-b border-slate-900 pb-1.5 flex items-center gap-1.5">
                  <Shield className="w-3.5 h-3.5" /> Defensivos & Utilidades
                </h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-slate-400">Bônus de HP:</span>
                    <span className={`font-bold ${eqMods.HP > 0 ? "text-emerald-400" : "text-slate-400"}`}>+{eqMods.HP}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Bônus de Armor:</span>
                    <span className={`font-bold ${eqMods.Armor > 0 ? "text-emerald-400" : "text-slate-400"}`}>+{eqMods.Armor}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Bônus de Magic Resist:</span>
                    <span className={`font-bold ${eqMods.MagicResistance > 0 ? "text-emerald-400" : "text-slate-400"}`}>+{eqMods.MagicResistance}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Velocidade de Movimento:</span>
                    <span className={`font-bold ${eqMods.MovementSpeed > 0 ? "text-emerald-400" : eqMods.MovementSpeed < 0 ? "text-rose-400" : "text-slate-400"}`}>
                      {eqMods.MovementSpeed > 0 ? `+${eqMods.MovementSpeed}` : eqMods.MovementSpeed}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Eficiência de Mana:</span>
                    <span className={`font-bold ${eqMods.ManaEfficiency > 0 ? "text-emerald-400" : "text-slate-400"}`}>+{eqMods.ManaEfficiency}%</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-400 font-mono">Redução de Cooldown (CDR):</span>
                    <div className="flex items-center gap-1.5">
                      {eqMods.CooldownReduction > 15 && (
                        <span className="text-[9px] bg-amber-500/10 text-amber-400 px-1 py-0.5 rounded font-mono font-bold leading-none border border-amber-500/20">
                          LIMITADO (Cap: 15%)
                        </span>
                      )}
                      <span className={`font-bold ${eqMods.CooldownReduction > 0 ? "text-emerald-400" : "text-slate-400"}`}>
                        {eqMods.CooldownReduction > 15 ? "15%" : `+${eqMods.CooldownReduction}%`}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Physical & Magical Offensive modifiers */}
              <div className="space-y-3 bg-slate-950/50 p-4 rounded-xl border border-slate-900">
                <h4 className="text-[10px] text-rose-400 font-bold uppercase border-b border-slate-900 pb-1.5 flex items-center gap-1.5">
                  <Sword className="w-3.5 h-3.5" /> Ofensivos Físicos & Mágicos
                </h4>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-slate-400">Chance de Crítico:</span>
                    <span className={`font-bold ${eqMods.CritChance > 0 ? "text-rose-400" : "text-slate-400"}`}>+{eqMods.CritChance}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Dano Crítico:</span>
                    <span className={`font-bold ${eqMods.CritDamage > 0 ? "text-rose-400" : "text-slate-400"}`}>+{eqMods.CritDamage}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Velocidade de Ataque:</span>
                    <span className={`font-bold ${eqMods.AttackSpeed > 0 ? "text-rose-400" : "text-slate-400"}`}>+{eqMods.AttackSpeed}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Penetração de Armadura:</span>
                    <span className={`font-bold ${eqMods.ArmorPenetration > 0 ? "text-rose-400" : "text-slate-400"}`}>+{eqMods.ArmorPenetration}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Nível de Magia (Magic Level):</span>
                    <span className={`font-bold ${eqMods.MagicLevel > 0 ? "text-violet-400" : "text-slate-400"}`}>+{eqMods.MagicLevel}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Healing Power:</span>
                    <span className={`font-bold ${eqMods.HealingPower > 0 ? "text-emerald-400" : "text-slate-400"}`}>+{eqMods.HealingPower}</span>
                  </div>
                </div>
              </div>

              {/* Elemental Amplification Panel */}
              <div className="md:col-span-2 space-y-2.5 bg-slate-950/50 p-4 rounded-xl border border-slate-900">
                <h4 className="text-[10px] text-amber-400 font-bold uppercase border-b border-slate-900 pb-1.5 flex items-center gap-1.5">
                  <Sparkles className="w-3.5 h-3.5" /> Amplificações de Dano Elemental de Equipamento
                </h4>
                <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 text-center">
                  {[
                    { label: "Holy Damage", val: eqMods.HolyDamage, color: "text-amber-300" },
                    { label: "Fire Damage", val: eqMods.FireDamage, color: "text-orange-400" },
                    { label: "Ice Damage", val: eqMods.IceDamage, color: "text-blue-300" },
                    { label: "Shadow Damage", val: eqMods.ShadowDamage, color: "text-purple-400" },
                    { label: "Nature Damage", val: eqMods.NatureDamage, color: "text-emerald-400" }
                  ].map(element => (
                    <div key={element.label} className="bg-slate-950 p-2 rounded-lg border border-slate-900 space-y-0.5">
                      <span className="text-[8.5px] text-slate-500 block leading-tight font-bold uppercase">{element.label}</span>
                      <span className={`text-xs font-black ${element.color}`}>+{element.val}%</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="p-2.5 bg-rose-500/5 rounded-lg border border-rose-500/10 text-[9.5px] text-rose-400 font-mono flex items-center gap-2">
              <ShieldAlert className="w-3.5 h-3.5 shrink-0" />
              <span>Conforme as regras de design canônico: Nenhum modificador de item se funde com os atributos base de classe estruturais. Os efeitos mantêm sua separação nativa.</span>
            </div>
          </div>

        </div>

        {/* Validation Board */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 space-y-3 shadow-xl">
          <div className="flex justify-between items-center border-b border-slate-800/60 pb-2">
            <div className="flex items-center gap-2 text-slate-300 font-bold text-xs uppercase tracking-wider">
              <CheckCircle className="w-4 h-4 text-emerald-400" />
              <span>Painel de Consolidação e Separação de Atributos</span>
            </div>
            <span className="text-[10px] text-emerald-400 font-mono font-extrabold uppercase">LOCKED</span>
          </div>

          <div className="space-y-2.5 text-xs font-mono">
            {/* Rule 1 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">base-stats-rule (base-stat-core-lock)</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  Classes definem rigidamente apenas 6 atributos estruturais de sobrevivência e mana base.
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>

            {/* Rule 2 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">equipment-modifier-rule (equip-modifier-rule)</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  O equipamento é a fonte primária de identidade e modifica parâmetros avançados (físicos, elementais, mágicos, etc.) sem corromper as bases de vocação.
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>

            {/* Rule 3 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">cooldown-reduction-rule (cdr-lock-rule)</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  A Redução de Cooldown (CDR) é altamente restrita a itens raros de Tier 4/Tier 5, com um limite global absoluto de 15%.
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>

            {/* Rule 4 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">infinite-level-softcap-rule</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  A progressão de nível do personagem é infinita, mas obedece a soft-caps de 100% (Lvl 1-200), 50% (Lvl 201-500) e 25% (Lvl 501+) de eficiência para evitar runaway power creep.
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>

            {/* Rule 5 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">elemental-cap-rule</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  A afinidade elemental tem um limite máximo rígido e insuperável de nível 100 para todas as especializações elementais (Fire, Ice, Holy, Shadow, Nature).
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>

            {/* Rule 6 */}
            <div className="flex items-start justify-between bg-slate-950 p-2.5 rounded border border-slate-900 gap-4">
              <div className="space-y-0.5">
                <span className="text-[10px] font-bold text-slate-300 block">power-hierarchy-rule</span>
                <p className="text-[9.5px] text-slate-500 leading-normal">
                  O poder é estruturado em quatro camadas: Nível (baseline soft-capped), Skill (mastery com diminishing), Gear (fonte primária) e Elemental (especialização cap 100).
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0 text-emerald-400 font-bold text-[10px]">
                <CheckCircle className="w-3.5 h-3.5 shrink-0" /> VERIFIED
              </div>
            </div>
          </div>
        </div>

      </div>

    </div>
  );
}
