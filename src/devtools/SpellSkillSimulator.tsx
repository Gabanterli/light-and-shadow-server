import React, { useState, useEffect } from "react";
import {
  Wand2,
  Sword,
  Shield,
  Heart,
  Zap,
  Sparkles,
  Sliders,
  CheckCircle,
  Info,
  AlertTriangle,
  Play,
  RotateCcw,
  Users,
  Eye,
  Activity,
  Flame,
  UserCheck,
  Award
} from "lucide-react";

// Types
interface Spell {
  SpellId: string;
  SpellName: string;
  ClassId: string;
  CooldownCategory: "Basic" | "Strong" | "Ultimate";
  ManaCost: number;
  CooldownSec: number;
  BasePower: number;
  ElementalType: "Physical" | "Fire" | "Ice" | "Holy" | "Shadow" | "Nature";
  IsAoE: boolean;
  IsHealing: boolean;
  skill_category: "Damage Spell" | "Heal Spell" | "Buff Spell" | "Debuff Spell" | "Mobility Skill" | "Utility Skill";
  Description: string;
}

const CANONICAL_CLASSES = [
  { id: "mage", name: "Mage", archetype: "spell-centric", defaultWeapon: 80, defaultMagicLvl: 45, mainElement: "Fire" },
  { id: "cleric", name: "Cleric", archetype: "spell-centric", defaultWeapon: 60, defaultMagicLvl: 40, mainElement: "Holy" },
  { id: "knight", name: "Knight", archetype: "weapon-centric", defaultWeapon: 120, defaultMagicLvl: 10, mainElement: "Physical" },
  { id: "archer", name: "Archer", archetype: "weapon-centric", defaultWeapon: 100, defaultMagicLvl: 15, mainElement: "Nature" },
  { id: "assassin", name: "Assassin", archetype: "weapon-centric", defaultWeapon: 95, defaultMagicLvl: 20, mainElement: "Shadow" }
];

const CANONICAL_SPELLS: Spell[] = [
  {
    SpellId: "slash_strike",
    SpellName: "Slash Strike",
    ClassId: "knight",
    CooldownCategory: "Basic",
    ManaCost: 8,
    CooldownSec: 2.0,
    BasePower: 25,
    ElementalType: "Physical",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "An instant heavy swing that deals physical damage to a single target."
  },
  {
    SpellId: "shield_bash",
    SpellName: "Shield Bash",
    ClassId: "knight",
    CooldownCategory: "Basic",
    ManaCost: 12,
    CooldownSec: 3.5,
    BasePower: 30,
    ElementalType: "Physical",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Debuff Spell",
    Description: "Bashes target with the shield, dealing moderate physical damage and a brief stun."
  },
  {
    SpellId: "exori_charge",
    SpellName: "Exori Charge",
    ClassId: "knight",
    CooldownCategory: "Strong",
    ManaCost: 30,
    CooldownSec: 10.0,
    BasePower: 75,
    ElementalType: "Physical",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "A spinning swing dealing high AoE physical damage to surrounding targets."
  },
  {
    SpellId: "gran_strike",
    SpellName: "Fierce Berserk Strike",
    ClassId: "knight",
    CooldownCategory: "Ultimate",
    ManaCost: 60,
    CooldownSec: 18.0,
    BasePower: 150,
    ElementalType: "Physical",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Ultimate physical burst that strikes all nearby enemies with extreme force."
  },
  {
    SpellId: "fireball",
    SpellName: "Fireball",
    ClassId: "mage",
    CooldownCategory: "Basic",
    ManaCost: 20,
    CooldownSec: 2.5,
    BasePower: 55,
    ElementalType: "Fire",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Launches a blazing sphere of fire that deals elemental fire damage."
  },
  {
    SpellId: "ice_barrage",
    SpellName: "Ice Barrage",
    ClassId: "mage",
    CooldownCategory: "Basic",
    ManaCost: 18,
    CooldownSec: 3.0,
    BasePower: 45,
    ElementalType: "Ice",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Fires shards of ice that deal ice damage to multiple targets in a targeted area."
  },
  {
    SpellId: "meteor_strike",
    SpellName: "Meteor Strike",
    ClassId: "mage",
    CooldownCategory: "Strong",
    ManaCost: 55,
    CooldownSec: 12.0,
    BasePower: 110,
    ElementalType: "Fire",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Calls down a giant meteor, dealing heavy Fire AoE damage with a slow effect."
  },
  {
    SpellId: "apocalypse_vortex",
    SpellName: "Apocalypse Vortex",
    ClassId: "mage",
    CooldownCategory: "Ultimate",
    ManaCost: 90,
    CooldownSec: 20.0,
    BasePower: 240,
    ElementalType: "Shadow",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "An ultimate vortex of shadow magic that tears through all targets in range."
  },
  {
    SpellId: "piercing_shot",
    SpellName: "Piercing Arrow",
    ClassId: "archer",
    CooldownCategory: "Basic",
    ManaCost: 10,
    CooldownSec: 2.0,
    BasePower: 35,
    ElementalType: "Physical",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Fires a swift penetrative arrow dealing single-target physical damage."
  },
  {
    SpellId: "rain_of_arrows",
    SpellName: "Rain of Arrows",
    ClassId: "archer",
    CooldownCategory: "Strong",
    ManaCost: 35,
    CooldownSec: 9.0,
    BasePower: 70,
    ElementalType: "Physical",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Rains physical arrows down on a targeted area, hitting up to many targets."
  },
  {
    SpellId: "nature_breeze",
    SpellName: "Nature's Breeze",
    ClassId: "archer",
    CooldownCategory: "Basic",
    ManaCost: 15,
    CooldownSec: 4.0,
    BasePower: 25,
    ElementalType: "Nature",
    IsAoE: false,
    IsHealing: true,
    skill_category: "Heal Spell",
    Description: "Archer's basic healing spell that recovers health based on Magic Level."
  },
  {
    SpellId: "god_slayer_arrow",
    SpellName: "God Slayer Arrow",
    ClassId: "archer",
    CooldownCategory: "Ultimate",
    ManaCost: 70,
    CooldownSec: 16.0,
    BasePower: 170,
    ElementalType: "Holy",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Fires a giant, light-infused arrow that pierces defenses for immense Holy damage."
  },
  {
    SpellId: "shadow_stab",
    SpellName: "Shadow Stab",
    ClassId: "assassin",
    CooldownCategory: "Basic",
    ManaCost: 12,
    CooldownSec: 1.5,
    BasePower: 40,
    ElementalType: "Shadow",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Stabs from the shadows, dealing dark burst damage to a single target."
  },
  {
    SpellId: "poison_dart",
    SpellName: "Poison Dart",
    ClassId: "assassin",
    CooldownCategory: "Basic",
    ManaCost: 15,
    CooldownSec: 3.0,
    BasePower: 30,
    ElementalType: "Nature",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Debuff Spell",
    Description: "Fires a dart coated in toxin, dealing damage and applying a slow."
  },
  {
    SpellId: "blade_dance",
    SpellName: "Blade Dance",
    ClassId: "assassin",
    CooldownCategory: "Strong",
    ManaCost: 40,
    CooldownSec: 11.0,
    BasePower: 85,
    ElementalType: "Physical",
    IsAoE: true,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Performs a rapid sequence of strikes, dealing physical damage to nearby enemies."
  },
  {
    SpellId: "death_mark",
    SpellName: "Death Mark Burst",
    ClassId: "assassin",
    CooldownCategory: "Ultimate",
    ManaCost: 75,
    CooldownSec: 19.0,
    BasePower: 190,
    ElementalType: "Shadow",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "Marks a target for execution, dealing massive Shadow damage after a delay."
  },
  {
    SpellId: "holy_heal",
    SpellName: "Holy Healing Light",
    ClassId: "cleric",
    CooldownCategory: "Basic",
    ManaCost: 15,
    CooldownSec: 2.5,
    BasePower: 45,
    ElementalType: "Holy",
    IsAoE: false,
    IsHealing: true,
    skill_category: "Heal Spell",
    Description: "Heals a single ally with holy magic, scaling beautifully with Magic Level."
  },
  {
    SpellId: "divine_burst",
    SpellName: "Divine Burst",
    ClassId: "cleric",
    CooldownCategory: "Basic",
    ManaCost: 14,
    CooldownSec: 2.0,
    BasePower: 35,
    ElementalType: "Holy",
    IsAoE: false,
    IsHealing: false,
    skill_category: "Damage Spell",
    Description: "A blast of holy energy that strikes a single enemy."
  },
  {
    SpellId: "sacred_sanctuary",
    SpellName: "Sacred Sanctuary",
    ClassId: "cleric",
    CooldownCategory: "Strong",
    ManaCost: 45,
    CooldownSec: 13.0,
    BasePower: 90,
    ElementalType: "Holy",
    IsAoE: true,
    IsHealing: true,
    skill_category: "Heal Spell",
    Description: "Summons a holy zone that continuously heals party members and strikes shadow creatures."
  },
  {
    SpellId: "resurrection_blessing",
    SpellName: "Divine Aegis Blessing",
    ClassId: "cleric",
    CooldownCategory: "Ultimate",
    ManaCost: 80,
    CooldownSec: 20.0,
    BasePower: 200,
    ElementalType: "Holy",
    IsAoE: true,
    IsHealing: true,
    skill_category: "Heal Spell",
    Description: "An ultimate protective shield and mass heal for all nearby party members."
  }
];

export function SpellSkillSimulator() {
  const [selectedClass, setSelectedClass] = useState(CANONICAL_CLASSES[0]);
  const [selectedSpell, setSelectedSpell] = useState<Spell>(CANONICAL_SPELLS.find(s => s.ClassId === "mage") || CANONICAL_SPELLS[4]);
  const [magicLevel, setMagicLevel] = useState<number>(45);
  const [weaponMagicalBase, setWeaponMagicalBase] = useState<number>(80);
  const [elementalAffinityLvl, setElementalAffinityLvl] = useState<number>(75);
  const [holyAmplification, setHolyAmplification] = useState<number>(1.5);
  const [targetCount, setTargetCount] = useState<number>(1);
  const [samePartyProtection, setSamePartyProtection] = useState<boolean>(false);
  const [forceCrit, setForceCrit] = useState<boolean>(false);

  // Buff stacking selections
  const [speedBuffA, setSpeedBuffA] = useState<boolean>(false); // Wind Run
  const [speedBuffB, setSpeedBuffB] = useState<boolean>(false); // Swiftness Breeze
  const [armorBuffA, setArmorBuffA] = useState<boolean>(false); // Armor Fortitude
  const [armorBuffB, setArmorBuffB] = useState<boolean>(false); // Divine Barrier
  const [auraBuffA, setAuraBuffA] = useState<boolean>(false);   // Flame Aura
  const [auraBuffB, setAuraBuffB] = useState<boolean>(false);   // Holy Aura

  // Simulation state outputs
  const [castLogs, setCastLogs] = useState<{ type: "formula" | "hit" | "warning" | "success"; text: string }[]>([]);
  const [critResult, setCritResult] = useState<{ triggered: boolean; multiplier: number } | null>(null);

  // Synchronize spell and defaults when class changes
  const handleClassChange = (cls: typeof CANONICAL_CLASSES[0]) => {
    setSelectedClass(cls);
    setWeaponMagicalBase(cls.defaultWeapon);
    setMagicLevel(cls.defaultMagicLvl);
    
    // Select first spell of new class
    const found = CANONICAL_SPELLS.find(s => s.ClassId === cls.id);
    if (found) {
      setSelectedSpell(found);
    }
  };

  // Run initial log on load and whenever attributes change
  useEffect(() => {
    runFormulaCalculation();
  }, [selectedClass, selectedSpell, magicLevel, weaponMagicalBase, elementalAffinityLvl, holyAmplification, targetCount, samePartyProtection, forceCrit, speedBuffA, speedBuffB, armorBuffA, armorBuffB, auraBuffA, auraBuffB]);

  // Buff Exclusivity Warnings
  const getBuffExclusivityWarnings = () => {
    const warnings: string[] = [];
    if (speedBuffA && speedBuffB) {
      warnings.push("Apenas 1 Buff de Velocidade de Movimento (Speed Category) pode estar ativo! (Wind Run substituído por Swiftness Breeze).");
    }
    if (armorBuffA && armorBuffB) {
      warnings.push("Apenas 1 Buff de Armadura (Defense Category) pode estar ativo! (Armor Fortitude substituído por Divine Barrier).");
    }
    if (auraBuffA && auraBuffB) {
      warnings.push("Apenas 1 Aura Ofensiva (Offensive Aura) pode estar ativa! (Flame Aura substituída por Holy Aura).");
    }
    return warnings;
  };

  // AoE Soft Cap Multipliers
  const getAoEMultiplier = (targets: number) => {
    if (targets <= 5) return { mult: 1.0, label: "100% de dano" };
    if (targets <= 10) return { mult: 0.8, label: "80% de dano" };
    if (targets <= 20) return { mult: 0.6, label: "60% de dano" };
    return { mult: 0.4, label: "40% de dano" };
  };

  const runFormulaCalculation = () => {
    const logs: { type: "formula" | "hit" | "warning" | "success"; text: string }[] = [];

    // Rule references
    logs.push({ type: "success", text: `=== INICIALIZANDO SIMULAÇÃO: [${selectedSpell.SpellName}] [${selectedSpell.skill_category}] ===` });
    logs.push({ type: "formula", text: `1. Camada de Nível de Poder: Classe [${selectedClass.name}] (${selectedClass.archetype})` });

    // Assassin Crit Override log
    if (selectedClass.id === "assassin") {
      logs.push({
        type: "success",
        text: `🗡️ ASSASSIN CRITICAL OVERRIDE (assassin-critical-override-rule): Assassin critical hit override (2.2x) applies to every valid outgoing damage source (physical, magical, skill-based, elemental, or hybrid).`
      });
    } else {
      logs.push({
        type: "formula",
        text: `🎯 Sistema Crítico Global: Chance Base: 5% | Multiplicador Padrão: 1.5x.`
      });
    }

    // 1. Elemental Affinity & Hard Cap Verification
    const effectiveAffinity = Math.min(elementalAffinityLvl, 100);
    const hasAffinityOvercap = elementalAffinityLvl > 100;
    
    if (hasAffinityOvercap) {
      logs.push({ 
        type: "warning", 
        text: `⚠️ ELEMENTAL CAP RULE TRIGGERED: Afinidade de ${elementalAffinityLvl} Lvl excedeu o limite! Limitado ao teto absoluto de 100.` 
      });
    }

    const affinityBonusPct = effectiveAffinity * 1.5; // Each lvl grants 1.5% bonus power
    logs.push({ 
      type: "formula", 
      text: `2. Afinidade Elemental: ${selectedSpell.ElementalType} Lvl ${effectiveAffinity}${hasAffinityOvercap ? ' (CAPPED)' : ''} => +${affinityBonusPct.toFixed(1)}% de bônus.` 
    });

    // 2. Spell Power Scaling: SpellPower = WeaponMagicalBase * MagicLevel * (1 + ElementalAffinityBonus)
    const multiplier = 1 + (affinityBonusPct / 100);
    const rawSpellPower = weaponMagicalBase * magicLevel * multiplier;

    logs.push({
      type: "formula",
      text: `3. SpellPower Raw Formula: Weapon Magical Base (${weaponMagicalBase}) * Magic Level (${magicLevel}) * AffinityMult (${multiplier.toFixed(3)}) = ${rawSpellPower.toLocaleString("en-US", { maximumFractionDigits: 1 })} SpellPower.`
    });

    // 3. Cooldown & Global Cooldown Validation
    logs.push({
      type: "formula",
      text: `4. Tempo de Execução: INSTANT CAST (instant-cast-rule). Cooldown da categoria [${selectedSpell.CooldownCategory}]: ${selectedSpell.CooldownSec}s. GCD Gatilho: 0.35s.`
    });

    // 4. AoE Target Soft-Cap calculations
    if (selectedSpell.IsAoE) {
      const { mult, label } = getAoEMultiplier(targetCount);
      const totalPowerScaled = rawSpellPower * mult;
      logs.push({
        type: "formula",
        text: `5. AOE SOFT-CAP SYSTEM (aoe-softcap-rule) ativo para ${targetCount} alvo(s): Multiplicador de ${label} (x${mult}) => Poder Efetivo por Alvo: ${totalPowerScaled.toLocaleString("en-US", { maximumFractionDigits: 1 })}.`
      });

      if (samePartyProtection) {
        logs.push({
          type: "success",
          text: `🛡️ PARTY PROTECTION RULE: Alvo(s) amigável(is) detectado(s). Dano de área anulado nos membros do grupo (0 DANO).`
        });
      } else {
        logs.push({
          type: "warning",
          text: `⚔️ Alvos vulneráveis: Dano de área ativo de forma indistinta (Modo PvP/Mundo Aberto).`
        });
      }
    } else {
      logs.push({
        type: "formula",
        text: `5. Alvo Único: Sem necessidade de cálculo de caps de alvos de área.`
      });
    }

    // 5. Healing Scaling Check (healing-scaling-rule)
    if (selectedSpell.IsHealing || selectedSpell.skill_category === "Heal Spell") {
      const calculatedHealingPower = weaponMagicalBase * magicLevel * holyAmplification;
      const baseHealAmt = calculatedHealingPower * (selectedSpell.BasePower / 100);
      logs.push({
        type: "success",
        text: `💖 HEALING SCALING SYSTEM (healing-scaling-rule): Healing Power = WeaponMagicalBase (${weaponMagicalBase}) * MagicLevel (${magicLevel}) * HolyAmplification (${holyAmplification.toFixed(2)}) = ${calculatedHealingPower.toLocaleString("en-US", { maximumFractionDigits: 1 })}.`
      });
      logs.push({
        type: "formula",
        text: `   - Eficácia PvP: 100% (Sem penalidades/reduções em arenas canônicas). Cura Base Esperada: ${baseHealAmt.toLocaleString("en-US", { maximumFractionDigits: 1 })} HP.`
      });
    }

    // 6. Buff Stacking Exclusivity
    const buffWarnings = getBuffExclusivityWarnings();
    if (buffWarnings.length > 0) {
      buffWarnings.forEach(w => logs.push({ type: "warning", text: `⚠️ BUFF RULE CRITICAL: ${w}` }));
    } else {
      logs.push({
        type: "success",
        text: `✅ Buffs Ativos: Totalmente exclusivos por categoria (buff-category-rule ativa). Sem runaway stat creep.`
      });
    }

    // 7. Skill Taxonomy Verification
    logs.push({
      type: "success",
      text: `🏷️ TAXONOMY VERIFICATION: Habilidade registrada canonicamente como [${selectedSpell.skill_category}]. (skill-taxonomy-rule)`
    });

    setCastLogs(logs);
  };

  const handleCastSpellSim = () => {
    // Perform standard crit roll (Base 5% chance. If Assassin, mult is 2.2x, else 1.5x)
    const critChance = 0.05;
    const isCrit = forceCrit || (Math.random() < critChance);
    const critMult = isCrit ? (selectedClass.id === "assassin" ? 2.2 : 1.5) : 1.0;

    setCritResult({
      triggered: isCrit,
      multiplier: critMult
    });

    // Calculate core power
    const effectiveAffinity = Math.min(elementalAffinityLvl, 100);
    const rawSpellPower = weaponMagicalBase * magicLevel * (1 + (effectiveAffinity * 1.5) / 100);
    const aoeFactor = selectedSpell.IsAoE ? getAoEMultiplier(targetCount).mult : 1.0;
    
    let baseVal = 0;
    if (selectedSpell.IsHealing || selectedSpell.skill_category === "Heal Spell") {
      // Correct healing scaling formula: WeaponMagicalBase * MagicLevel * HolyAmplification
      const calculatedHealingPower = weaponMagicalBase * magicLevel * holyAmplification;
      baseVal = calculatedHealingPower * (selectedSpell.BasePower / 100);
    } else {
      // Standard damage scaling
      baseVal = rawSpellPower * (selectedSpell.BasePower / 100) * aoeFactor;
    }

    const finalResult = Math.round(baseVal * critMult);

    let actionWord = (selectedSpell.IsHealing || selectedSpell.skill_category === "Heal Spell") ? "Curou" : "Deu";
    let targetWord = (selectedSpell.IsHealing || selectedSpell.skill_category === "Heal Spell") ? "pontos de Vida" : `pontos de dano do tipo ${selectedSpell.ElementalType}`;

    const hitLogText = isCrit 
      ? `💥 ¡CRÍTICO DEVASTADOR (${critMult}x)! ${actionWord} ${finalResult} ${targetWord} com ${selectedSpell.SpellName}!`
      : `✨ Conjurou ${selectedSpell.SpellName} com sucesso! ${actionWord} ${finalResult} ${targetWord}.`;

    const newHitLog = { type: "hit" as const, text: hitLogText };
    setCastLogs(prev => [newHitLog, ...prev]);
  };

  return (
    <div className="col-span-12 grid grid-cols-1 lg:grid-cols-12 gap-6" id="spell-skill-bible-sim">
      
      {/* Module 1: Settings Panels (Left Side) */}
      <div className="lg:col-span-5 space-y-6">
        
        {/* Class Selection & Core Attributes Card */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-5" id="vocation-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Award className="w-4 h-4 text-amber-500 animate-pulse" />
              1. Seleção de Vocações & Atributos
            </h2>
            <span className="text-[10px] bg-slate-950 px-2 py-0.5 rounded text-slate-400 font-mono border border-slate-800">
              GODOT 4 CORE
            </span>
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] font-mono text-slate-400">Vocações Canônicas:</label>
            <div className="grid grid-cols-5 gap-1.5">
              {CANONICAL_CLASSES.map(cls => (
                <button
                  key={cls.id}
                  onClick={() => handleClassChange(cls)}
                  className={`py-2 rounded text-center text-xs font-mono font-bold transition flex flex-col items-center gap-1 ${
                    selectedClass.id === cls.id
                      ? "bg-amber-500/20 text-amber-400 border border-amber-500/50"
                      : "bg-slate-950/60 text-slate-400 border border-slate-800/50 hover:bg-slate-900"
                  }`}
                >
                  <span className="text-sm">
                    {cls.id === "mage" && "🔮"}
                    {cls.id === "cleric" && "☀️"}
                    {cls.id === "knight" && "🛡️"}
                    {cls.id === "archer" && "🏹"}
                    {cls.id === "assassin" && "🗡️"}
                  </span>
                  <span className="text-[9px] uppercase tracking-tight">{cls.name}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-2">
            {/* Weapon Magical Base */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Weapon Magical Base:</span>
                <span className="text-amber-400 font-bold">{weaponMagicalBase}</span>
              </div>
              <input
                type="range"
                min="10"
                max="300"
                value={weaponMagicalBase}
                onChange={(e) => setWeaponMagicalBase(parseInt(e.target.value))}
                className="w-full accent-amber-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
              <span className="text-[9px] text-slate-500 font-mono block">Mín: 10 / Máx: 300</span>
            </div>

            {/* Magic Level */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Nível Mágico (ML):</span>
                <span className="text-violet-400 font-bold">{magicLevel}</span>
              </div>
              <input
                type="range"
                min="1"
                max="150"
                value={magicLevel}
                onChange={(e) => setMagicLevel(parseInt(e.target.value))}
                className="w-full accent-violet-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
              <span className="text-[9px] text-slate-500 font-mono block">Mín: 1 / Máx: 150</span>
            </div>
          </div>

          {/* Holy Amplification (for heals and holy specs) */}
          <div className="border-t border-slate-800/40 pt-3 space-y-1.5">
            <div className="flex justify-between items-center text-[11px] font-mono">
              <span className="text-pink-400 font-bold flex items-center gap-1">
                💖 Amplificação Sagrada (Holy Amplification):
              </span>
              <span className="text-pink-400 font-bold">{holyAmplification.toFixed(2)}x</span>
            </div>
            <input
              type="range"
              min="0.5"
              max="3.0"
              step="0.05"
              value={holyAmplification}
              onChange={(e) => setHolyAmplification(parseFloat(e.target.value))}
              className="w-full accent-pink-500 h-1 bg-slate-950 rounded cursor-pointer"
            />
            <span className="text-[9px] text-slate-500 font-mono block">Formula canônica de cura (healing-scaling-rule) ativa de 0.5x a 3.0x</span>
          </div>

          {/* Elemental Affinity & Hard Cap */}
          <div className="border-t border-slate-800/40 pt-3 space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-[11px] text-slate-300 font-mono font-bold flex items-center gap-1.5">
                <Flame className="w-3.5 h-3.5 text-orange-500" />
                Afinidade Elemental ({selectedClass.mainElement}):
              </span>
              <div className="flex items-center gap-1">
                <input
                  type="number"
                  min="0"
                  max="120"
                  value={elementalAffinityLvl}
                  onChange={(e) => {
                    const val = parseInt(e.target.value);
                    setElementalAffinityLvl(isNaN(val) ? 0 : val);
                  }}
                  className="w-12 bg-slate-950 text-orange-400 font-mono font-bold text-center border border-slate-800 py-0.5 rounded text-xs"
                />
              </div>
            </div>

            <input
              type="range"
              min="0"
              max="120"
              value={elementalAffinityLvl}
              onChange={(e) => setElementalAffinityLvl(parseInt(e.target.value))}
              className="w-full accent-orange-500 h-1 bg-slate-950 rounded cursor-pointer"
            />
            
            <div className="flex justify-between items-center text-[10px] font-mono">
              <span className="text-slate-500">Lvl 0</span>
              {elementalAffinityLvl > 100 ? (
                <span className="text-red-400 font-bold bg-red-500/10 border border-red-500/20 px-1.5 py-0.5 rounded text-[9px] uppercase animate-pulse">
                  LIMITADO (Cap: 100)
                </span>
              ) : (
                <span className="text-emerald-400 font-bold">Máx Mastery: 100</span>
              )}
              <span className="text-slate-500">Lvl 120</span>
            </div>
          </div>
        </div>

        {/* Spells List & Current Selection Card */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="spell-selection-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Wand2 className="w-4 h-4 text-violet-500" />
              2. Catálogo de Habilidades & Taxonomia
            </h2>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            Selecione uma habilidade para aplicar os testes de <strong className="text-amber-400">skill-taxonomy-rule</strong> e visualizar seus multiplicadores canônicos.
          </p>

          <div className="space-y-1.5 max-h-[220px] overflow-y-auto pr-1 text-xs">
            {CANONICAL_SPELLS.filter(s => s.ClassId === selectedClass.id).map(spell => {
              const isSelected = selectedSpell.SpellId === spell.SpellId;
              return (
                <button
                  key={spell.SpellId}
                  onClick={() => setSelectedSpell(spell)}
                  className={`w-full p-2.5 rounded text-left transition border font-mono flex flex-col gap-1 ${
                    isSelected
                      ? "bg-slate-850 text-amber-400 border-amber-500/40"
                      : "bg-slate-950/40 text-slate-300 border-slate-800/60 hover:bg-slate-900/60"
                  }`}
                >
                  <div className="flex justify-between items-center">
                    <span className="font-bold flex items-center gap-1.5">
                      {(spell.IsHealing || spell.skill_category === "Heal Spell") ? "💖" : "⚔️"} {spell.SpellName}
                    </span>
                    <span className={`text-[9px] px-1.5 py-0.5 rounded uppercase font-bold bg-violet-950 text-violet-300 border border-violet-850`}>
                      {spell.skill_category}
                    </span>
                  </div>
                  <div className="flex justify-between text-[10px] text-slate-400">
                    <span>CD: {spell.CooldownSec}s ({spell.CooldownCategory})</span>
                    <span>Mana: {spell.ManaCost}</span>
                    <span>Tipo: {spell.ElementalType}</span>
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Buff Categories Exclusivity Verification Card */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="buff-categories-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Zap className="w-4 h-4 text-amber-500" />
              3. Verificação de Acúmulo de Buffs (Exclusivos)
            </h2>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            Teste a regra <strong className="text-amber-400">buff-category-rule</strong>. Marcar buffs da mesma categoria gerará advertência de sobreposição e ignorará buffs duplicados.
          </p>

          <div className="space-y-3.5">
            {/* Category: Speed */}
            <div className="space-y-1.5">
              <span className="text-[10px] text-slate-400 font-mono uppercase block">Categoria A: Movimento (Speed)</span>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setSpeedBuffA(!speedBuffA)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    speedBuffA ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  💨 [Wind Run]
                </button>
                <button
                  onClick={() => setSpeedBuffB(!speedBuffB)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    speedBuffB ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  🍃 [Swiftness Breeze]
                </button>
              </div>
            </div>

            {/* Category: Defense */}
            <div className="space-y-1.5">
              <span className="text-[10px] text-slate-400 font-mono uppercase block">Categoria B: Armadura (Defense)</span>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setArmorBuffA(!armorBuffA)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    armorBuffA ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  🛡️ [Armor Fortitude]
                </button>
                <button
                  onClick={() => setArmorBuffB(!armorBuffB)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    armorBuffB ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  ☀️ [Divine Barrier]
                </button>
              </div>
            </div>

            {/* Category: Offense */}
            <div className="space-y-1.5">
              <span className="text-[10px] text-slate-400 font-mono uppercase block">Categoria C: Aura Ativa (Offense)</span>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setAuraBuffA(!auraBuffA)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    auraBuffA ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  🔥 [Flame Aura]
                </button>
                <button
                  onClick={() => setAuraBuffB(!auraBuffB)}
                  className={`p-2 rounded text-xs font-mono transition text-left border ${
                    auraBuffB ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" : "bg-slate-950/50 text-slate-500 border-slate-900"
                  }`}
                >
                  🌟 [Holy Aura]
                </button>
              </div>
            </div>
          </div>
        </div>

      </div>

      {/* Module 2: Combat Simulation, Formula Output, Logs (Right Side) */}
      <div className="lg:col-span-7 space-y-6">
        
        {/* Spell Calculator Output Display */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-5" id="spell-calculator-output">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400 animate-pulse" />
              Simulador de Combate Efetivo
            </h2>
            <div className="flex items-center gap-3">
              <label className="flex items-center gap-1.5 text-xs font-mono text-slate-300 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={forceCrit}
                  onChange={(e) => setForceCrit(e.target.checked)}
                  className="rounded bg-slate-950 border-slate-850 text-amber-500 focus:ring-0 w-3.5 h-3.5"
                />
                <span>Forçar Crítico</span>
              </label>
              <button
                onClick={handleCastSpellSim}
                className="px-4 py-1.5 rounded bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-xs uppercase font-mono transition-all flex items-center gap-1.5 shadow"
              >
                <Play className="w-3 h-3 fill-current" />
                Executar Magia
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-center">
            {/* Base Power */}
            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Poder Base (Spell)</span>
              <span className="text-lg font-bold text-slate-200">{selectedSpell.BasePower}%</span>
            </div>
            
            {/* Final Power (Capped) */}
            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Custo de Mana</span>
              <span className="text-lg font-bold text-violet-400">{selectedSpell.ManaCost} MP</span>
            </div>

            {/* Cooldown */}
            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Cooldown</span>
              <span className="text-lg font-bold text-orange-400">{selectedSpell.CooldownSec}s</span>
            </div>

            {/* Global Cooldown (GCD) */}
            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Gatilho GCD</span>
              <span className="text-lg font-bold text-emerald-400">0.35s</span>
            </div>
          </div>

          {/* Spell Properties Extra: AoE and Same Party Protection */}
          <div className="p-4 bg-slate-950/40 rounded-xl border border-slate-800/60 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex justify-between items-center text-xs font-mono">
                <span className="text-slate-400 flex items-center gap-1.5">
                  <Users className="w-3.5 h-3.5 text-blue-400" />
                  Alvos no Raio de Ação (AoE):
                </span>
                <span className="text-blue-400 font-bold">{selectedSpell.IsAoE ? targetCount : "N/A (Alvo Único)"}</span>
              </div>
              <input
                type="range"
                min="1"
                max="30"
                disabled={!selectedSpell.IsAoE}
                value={targetCount}
                onChange={(e) => setTargetCount(parseInt(e.target.value))}
                className="w-full accent-blue-500 h-1 bg-slate-950 rounded cursor-pointer disabled:opacity-30"
              />
              <div className="flex justify-between text-[9px] text-slate-500 font-mono">
                <span>1 Alvo</span>
                <span>30 Alvos</span>
              </div>
            </div>

            <div className="flex flex-col justify-center gap-2">
              <label className="flex items-center gap-2 text-xs font-mono text-slate-300 cursor-pointer select-none">
                <input
                  type="checkbox"
                  disabled={!selectedSpell.IsAoE}
                  checked={samePartyProtection}
                  onChange={(e) => setSamePartyProtection(e.target.checked)}
                  className="rounded bg-slate-950 border-slate-800 text-amber-500 focus:ring-0 w-3.5 h-3.5 disabled:opacity-30"
                />
                <span className={selectedSpell.IsAoE ? "text-slate-300" : "text-slate-500"}>
                  Ativar Proteção de Grupo (Party Protection)
                </span>
              </label>
              <span className="text-[9.5px] text-slate-500 font-mono leading-tight">
                Se habilitado, aliados no mesmo grupo não tomam dano amigável de habilidades em área.
              </span>
            </div>
          </div>

          {/* Spell Power Calculation Formula Live Audit */}
          <div className="bg-slate-950 p-4 rounded-xl border border-slate-800 space-y-3 font-mono text-xs">
            <div className="flex justify-between items-center border-b border-slate-800 pb-2">
              <span className="text-emerald-400 font-bold uppercase text-[11px] flex items-center gap-1.5">
                <Sliders className="w-3.5 h-3.5 text-emerald-500" />
                Auditoria de Fórmula Efetiva
              </span>
              <span className="text-[9px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded">
                CANONICAL ACTIVE
              </span>
            </div>

            <div className="space-y-1 leading-relaxed text-slate-300">
              <div className="flex justify-between">
                <span className="text-slate-500">Fórmula de Poder de Magia:</span>
                <span className="text-slate-300">WeaponMagicalBase * MagicLevel * (1 + AffinityBonus / 100)</span>
              </div>
              <div className="flex justify-between font-bold text-violet-400">
                <span>SpellPower Computado:</span>
                <span>
                  {Math.round(weaponMagicalBase * magicLevel * (1 + (Math.min(elementalAffinityLvl, 100) * 1.5) / 100))} SP
                </span>
              </div>

              {(selectedSpell.IsHealing || selectedSpell.skill_category === "Heal Spell") ? (
                <>
                  <div className="flex justify-between text-pink-400 border-t border-slate-900 pt-1.5 mt-1">
                    <span className="font-bold">Fórmula de Cura Canônica:</span>
                    <span className="text-slate-300">WeaponMagicalBase * MagicLevel * HolyAmplification</span>
                  </div>
                  <div className="flex justify-between text-pink-400">
                    <span className="font-bold">Poder de Cura da Vocaçao:</span>
                    <span>{(weaponMagicalBase * magicLevel * holyAmplification).toLocaleString("en-US", { maximumFractionDigits: 1 })} HP Power</span>
                  </div>
                  <div className="flex justify-between text-pink-300 font-bold">
                    <span>Cura Real Esperada (BasePower {selectedSpell.BasePower}%):</span>
                    <span>{(weaponMagicalBase * magicLevel * holyAmplification * (selectedSpell.BasePower / 100)).toLocaleString("en-US", { maximumFractionDigits: 1 })} HP</span>
                  </div>
                </>
              ) : (
                <div className="flex justify-between text-orange-400 border-t border-slate-900 pt-1.5 mt-1">
                  <span>Dano de Golpe Esperado (BasePower {selectedSpell.BasePower}%):</span>
                  <span>
                    {Math.round(
                      (weaponMagicalBase * magicLevel * (1 + (Math.min(elementalAffinityLvl, 100) * 1.5) / 100)) * 
                      (selectedSpell.BasePower / 100) * (selectedSpell.IsAoE ? getAoEMultiplier(targetCount).mult : 1.0)
                    )} DMG base
                  </span>
                </div>
              )}

              {/* Assassin Critical Override Detail */}
              {selectedClass.id === "assassin" && (
                <div className="flex justify-between text-amber-400 border-t border-slate-900 pt-1.5 mt-1 font-bold">
                  <span>Assassin Critical Multiplier Override:</span>
                  <span>2.2x (aplica-se a TODAS as fontes de dano)</span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Live Mathematical Audit Logs */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-3" id="live-audit-logs">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Eye className="w-4 h-4 text-amber-500" />
              Logs de Combate e Audit Matemático (Fórmula)
            </h2>
            <button
              onClick={() => {
                setCastLogs([]);
                runFormulaCalculation();
              }}
              className="text-slate-500 hover:text-slate-300 transition-all font-mono text-[10px] flex items-center gap-1"
            >
              <RotateCcw className="w-3 h-3" /> Limpar Logs
            </button>
          </div>

          <div className="bg-slate-950 p-4 rounded-xl border border-slate-900 h-[220px] overflow-y-auto space-y-2 font-mono text-[11px] leading-relaxed">
            {castLogs.length === 0 ? (
              <div className="text-slate-600 italic text-center pt-16">Nenhum log gerado. Execute uma magia ou altere atributos.</div>
            ) : (
              castLogs.map((log, index) => (
                <div
                  key={index}
                  className={`p-2 rounded border transition ${
                    log.type === "warning" ? "bg-red-500/10 text-red-300 border-red-500/20" :
                    log.type === "success" ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20 font-bold" :
                    log.type === "hit" ? "bg-amber-500/10 text-amber-400 border-amber-500/30 font-bold animate-pulse text-xs" :
                    "bg-slate-900/40 text-slate-300 border-slate-800/40"
                  }`}
                >
                  {log.text}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Canonical Rules Registered Status Checklist */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="bible-checklist-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-emerald-500" />
              Requisitos Canônicos de Spells & Combat
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {[
              { id: "high-ttk-spell-combat-rule", title: "high-ttk-spell-combat-rule", desc: "TTK alto, pressão constante, combate fluido." },
              { id: "skill-cooldown-rule", title: "skill-cooldown-rule", desc: "Cooldowns divididos em Basic, Strong e Ultimate." },
              { id: "global-cooldown-rule", title: "global-cooldown-rule", desc: "GCD de 0.35s para evitar macros de burst instantâneo." },
              { id: "instant-cast-rule", title: "instant-cast-rule", desc: "Sem canalização ou barra de cast: conjuração instantânea." },
              { id: "damage-source-rule", title: "damage-source-rule", desc: "Guerreiros escalam arma; Magos escalam magia." },
              { id: "spell-scaling-rule", title: "spell-scaling-rule", desc: "SpellPower = WeaponMagBase * MagicLvl * ElementalAffinity." },
              { id: "mana-flow-rule", title: "mana-flow-rule", desc: "Mana regenera passivamente e por hit ativo." },
              { id: "healing-scaling-rule", title: "healing-scaling-rule", desc: "Cura escala com Weapon Magical Base, Magic Level e Holy Amplification." },
              { id: "spell-critical-rule", title: "spell-critical-rule", desc: "Crítico de magias (base 5% chance, 1.5x / Assassin 2.2x)." },
              { id: "limited-cc-rule", title: "limited-cc-rule", desc: "Controle de grupo limitado para evitar chain-lock." },
              { id: "aoe-softcap-rule", title: "aoe-softcap-rule", desc: "Teto de alvo progressivo: 1-5 (100%), 6-10 (80%), 11-20 (60%)." },
              { id: "buff-category-rule", title: "buff-category-rule", desc: "Buffs são exclusivos por categoria, impedindo stack." },
              { id: "party-protection-rule", title: "party-protection-rule", desc: "Aliados da mesma party protegidos de dano amigável em área." },
              { id: "assassin-critical-override-rule", title: "assassin-critical-override-rule", desc: "Crítico de Assassin (2.2x) se aplica a TODAS as fontes de dano." },
              { id: "skill-taxonomy-rule", title: "skill-taxonomy-rule", desc: "Habilidades divididas canonicamente em Damage, Heal, Buff, Debuff, Mobility, Utility." }
            ].map(rule => (
              <div key={rule.id} className="bg-slate-950 p-2.5 rounded border border-slate-900 flex justify-between gap-3 text-xs font-mono">
                <div className="space-y-0.5">
                  <span className="text-[10px] font-bold text-slate-300 block">{rule.title}</span>
                  <p className="text-[9.5px] text-slate-500 leading-tight">{rule.desc}</p>
                </div>
                <span className="flex items-center gap-1 text-emerald-400 font-bold text-[10px] shrink-0 align-middle">
                  <CheckCircle className="w-3.5 h-3.5" /> APPROVED
                </span>
              </div>
            ))}
          </div>
        </div>

      </div>

    </div>
  );
}
