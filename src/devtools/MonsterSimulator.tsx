import React, { useState, useEffect } from "react";
import {
  Skull,
  Award,
  Zap,
  Compass,
  RefreshCw,
  Users,
  Eye,
  Sliders,
  CheckCircle,
  HelpCircle,
  Shield,
  Play,
  RotateCcw,
  Gift,
  Flame,
  ArrowRight,
  UserCheck
} from "lucide-react";

// Types matching the JSON templates
interface Monster {
  MonsterId: string;
  MonsterName: string;
  Biome: string;
  BaseHP: number;
  BaseDamage: number;
  AggroType: string;
  PrimaryElement: "Physical" | "Fire" | "Ice" | "Holy" | "Shadow" | "Nature";
  Vulnerabilities: string[];
  Resistances: string[];
  BaseXP: number;
  Description: string;
}

interface LootItem {
  LootId: string;
  LootName: string;
  Rarity: "Normal" | "Magic" | "Rare" | "Epic" | "Legendary";
  Category: string;
  Biomes: string[];
  DropRatePct: number;
  BaseValue: number;
}

const CANONICAL_MONSTERS: Monster[] = [
  {
    MonsterId: "plague_rat",
    MonsterName: "Plague Rat",
    Biome: "Sewers / Ruins",
    BaseHP: 450,
    BaseDamage: 25,
    AggroType: "Normal",
    PrimaryElement: "Nature",
    Vulnerabilities: ["Fire"],
    Resistances: ["Nature"],
    BaseXP: 120,
    Description: "A disease-ridden vermin that attacks with swift bites."
  },
  {
    MonsterId: "skeleton_warrior",
    MonsterName: "Skeleton Warrior",
    Biome: "Crypts / Dark Forests",
    BaseHP: 1200,
    BaseDamage: 55,
    AggroType: "Normal",
    PrimaryElement: "Physical",
    Vulnerabilities: ["Holy"],
    Resistances: ["Shadow"],
    BaseXP: 350,
    Description: "An undead soldier carrying a rusty shield and scimitar."
  },
  {
    MonsterId: "stone_golem_elite",
    MonsterName: "Stonewarded Guardian Golem",
    Biome: "Mountain Peaks",
    BaseHP: 4500,
    BaseDamage: 110,
    AggroType: "Tactical (Elite)",
    PrimaryElement: "Physical",
    Vulnerabilities: ["Physical"],
    Resistances: ["Ice", "Fire"],
    BaseXP: 1200,
    Description: "A massive elite construct carved from rock that defends mountain corridors."
  },
  {
    MonsterId: "shadow_assassin_elite",
    MonsterName: "Mana-Gorged Shadow Stalker",
    Biome: "Ancient Spires",
    BaseHP: 3200,
    BaseDamage: 140,
    AggroType: "Tactical (Elite)",
    PrimaryElement: "Shadow",
    Vulnerabilities: ["Holy"],
    Resistances: ["Shadow", "Nature"],
    BaseXP: 1500,
    Description: "An agile elite assassin that prioritizes ranged targets and disrupts spells."
  },
  {
    MonsterId: "ash_dragon_worldboss",
    MonsterName: "Ignisaur the Ash Devourer",
    Biome: "Volcanic Caldera (Open World)",
    BaseHP: 150000,
    BaseDamage: 450,
    AggroType: "Tactical (World Boss)",
    PrimaryElement: "Fire",
    Vulnerabilities: ["Ice"],
    Resistances: ["Fire", "Physical"],
    BaseXP: 25000,
    Description: "A colossal open-world boss. Spawns in volcanic territories, contested by guilds."
  },
  {
    MonsterId: "lich_king_instancedboss",
    MonsterName: "Malakar the Soul Binder",
    Biome: "Glacial Palace (Instanced)",
    BaseHP: 85000,
    BaseDamage: 350,
    AggroType: "Tactical (Instanced Boss)",
    PrimaryElement: "Ice",
    Vulnerabilities: ["Fire", "Holy"],
    Resistances: ["Ice", "Shadow"],
    BaseXP: 18000,
    Description: "A dangerous instanced boss that tests party coordination, interrupts, and phase control."
  }
];

const CANONICAL_LOOT: LootItem[] = [
  {
    LootId: "rusted_sword",
    LootName: "Rusted Knight Sword",
    Rarity: "Normal",
    Category: "Weapons",
    Biomes: ["Sewers / Ruins", "Crypts / Dark Forests"],
    DropRatePct: 45.0,
    BaseValue: 15
  },
  {
    LootId: "golem_core_fragment",
    LootName: "Stonewarded Golem Core",
    Rarity: "Rare",
    Category: "Crafting Materials",
    Biomes: ["Mountain Peaks"],
    DropRatePct: 12.5,
    BaseValue: 150
  },
  {
    LootId: "shadow_stalker_dagger",
    LootName: "Vesper Shadow Dagger",
    Rarity: "Epic",
    Category: "Weapons",
    Biomes: ["Ancient Spires"],
    DropRatePct: 2.5,
    BaseValue: 850
  },
  {
    LootId: "ignisaur_scale_mail",
    LootName: "Ash Devourer Scale Plate",
    Rarity: "Legendary",
    Category: "Armor",
    Biomes: ["Volcanic Caldera (Open World)"],
    DropRatePct: 0.5,
    BaseValue: 5000
  },
  {
    LootId: "soul_binder_staff",
    LootName: "Staff of Eternal Winter",
    Rarity: "Legendary",
    Category: "Weapons",
    Biomes: ["Glacial Palace (Instanced)"],
    DropRatePct: 1.0,
    BaseValue: 4500
  },
  {
    LootId: "elemental_mana_crystal",
    LootName: "Pure Elemental Core",
    Rarity: "Magic",
    Category: "Shards",
    Biomes: ["Mountain Peaks", "Ancient Spires", "Volcanic Caldera (Open World)", "Glacial Palace (Instanced)"],
    DropRatePct: 20.0,
    BaseValue: 120
  }
];

const BIOMES = [
  "Sewers / Ruins",
  "Crypts / Dark Forests",
  "Mountain Peaks",
  "Ancient Spires",
  "Volcanic Caldera (Open World)",
  "Glacial Palace (Instanced)"
];

const MODIFIERS = [
  { name: "None", hpMult: 1.0, dmgMult: 1.0, xpMult: 1.0, desc: "A regular monster walking the ecosystem." },
  { name: "Stonewarded", hpMult: 1.5, dmgMult: 1.0, xpMult: 1.3, desc: "+50% HP. Infused with granite armor." },
  { name: "Flamebound", hpMult: 1.1, dmgMult: 1.4, xpMult: 1.3, desc: "+40% damage. Emits random solar outbursts." },
  { name: "Mana-Gorged", hpMult: 1.2, dmgMult: 1.2, xpMult: 1.4, desc: "+20% stats. Spells have lower CD and interrupt potential." },
  { name: "Apex Predatory", hpMult: 2.0, dmgMult: 1.8, xpMult: 2.2, desc: "Dangerous mini-boss modifier. Massive stat amplification." }
];

export function MonsterSimulator() {
  const [selectedMonster, setSelectedMonster] = useState<Monster>(CANONICAL_MONSTERS[0]);
  const [selectedBiome, setSelectedBiome] = useState<string>("Sewers / Ruins");
  const [selectedModifier, setSelectedModifier] = useState(MODIFIERS[0]);
  
  // XP sliders
  const [dmgContrib, setDmgContrib] = useState<number>(40);
  const [healContrib, setHealContrib] = useState<number>(20);
  const [participationPct, setParticipationPct] = useState<number>(80);
  const [survivalMod, setSurvivalMod] = useState<number>(1.0); // 1.0 if alive, 0.5 if dead

  // Respawn state
  const [playerDensity, setPlayerDensity] = useState<number>(50); // 0-100%
  const [baseRespawnSec, setBaseRespawnSec] = useState<number>(60);

  // Boss state
  const [bossType, setBossType] = useState<"world" | "instanced">("world");
  const [bossActivePhase, setBossActivePhase] = useState<number>(1);
  const [bossSimLog, setBossSimLog] = useState<string[]>([]);

  // Simulation logs
  const [auditLogs, setAuditLogs] = useState<{ type: string; text: string }[]>([]);
  const [rolledLoot, setRolledLoot] = useState<{ lootName: string; rarity: string; value: number } | null>(null);

  useEffect(() => {
    runEcosystemCalculation();
  }, [selectedMonster, selectedBiome, selectedModifier, dmgContrib, healContrib, participationPct, survivalMod, playerDensity, baseRespawnSec]);

  // Recalculate dynamic values
  const runEcosystemCalculation = () => {
    const logs: { type: string; text: string }[] = [];

    logs.push({
      type: "info",
      text: `=== INICIANDO AUDITORIA: ECOSSISTEMA E CRIATURAS (monster-ecosystem-rule) ===`
    });

    // Base monster + modifier calculation
    const effectiveHP = Math.round(selectedMonster.BaseHP * selectedModifier.hpMult);
    const effectiveDamage = Math.round(selectedMonster.BaseDamage * selectedModifier.dmgMult);
    const modifiedXP = Math.round(selectedMonster.BaseXP * selectedModifier.xpMult);

    logs.push({
      type: "success",
      text: `1. Entidade: ${selectedMonster.MonsterName} (${selectedMonster.Biome}) | Modificador: ${selectedModifier.name}`
    });
    logs.push({
      type: "formula",
      text: `   - Vida Efetiva: ${selectedMonster.BaseHP} base * ${selectedModifier.hpMult}x = ${effectiveHP} HP.`
    });
    logs.push({
      type: "formula",
      text: `   - Dano Efetivo: ${selectedMonster.BaseDamage} base * ${selectedModifier.dmgMult}x = ${effectiveDamage} DMG.`
    });

    // AI logic trigger
    if (selectedMonster.AggroType.includes("Tactical") || selectedModifier.name === "Apex Predatory") {
      logs.push({
        type: "success",
        text: `🧠 INTELIGÊNCIA TÁTICA ATIVA (monster-ai-rule): Esta entidade executa decisões de movimentação, priorização de healers e interrupções em habilidades de alto custo.`
      });
    } else {
      logs.push({
        type: "info",
        text: `🤖 AGGRO BÁSICO ATIVO (monster-ai-rule): Comportamento padrão de perseguição e ataque físico em loop.`
      });
    }

    // Elemental soft advantages
    logs.push({
      type: "success",
      text: `⚡ ADVANTAJES ELEMENTAIS (elemental-interaction-rule): Vantagens suaves ativas.`
    });
    logs.push({
      type: "formula",
      text: `   - Vulnerabilidades (+15% Dano recebido): ${selectedMonster.Vulnerabilities.join(", ")}`
    });
    logs.push({
      type: "formula",
      text: `   - Resistências (-10% Dano recebido): ${selectedMonster.Resistances.join(", ")}`
    });

    // XP Breakdown formula
    const totalContrib = dmgContrib + healContrib;
    const contribFactor = totalContrib > 100 ? 1.0 : totalContrib / 100;
    const computedXPEarned = Math.round(
      modifiedXP * 
      (0.40 * (dmgContrib / 100) + 
       0.30 * (healContrib / 100) + 
       0.20 * (participationPct / 100) + 
       0.10 * survivalMod)
    );

    logs.push({
      type: "success",
      text: `🎯 RECOMPENSA DE XP DE PERFORMANCE (xp-performance-rule):`
    });
    logs.push({
      type: "formula",
      text: `   - XP Base Modificado: ${modifiedXP}`
    });
    logs.push({
      type: "formula",
      text: `   - Fórmula: Base * (0.4 * DmgContrib + 0.3 * HealContrib + 0.2 * Participation + 0.1 * Survival)`
    });
    logs.push({
      type: "info",
      text: `   - XP Efetivo Concedido ao Jogador: ${computedXPEarned} XP.`
    });

    // Dynamic Respawn Calculations
    // Timer is dynamically modified by player density
    // high density -> respawn rates adapt (lower timer to sustain, but diminishes yield or increases Elite hazard probability)
    const densityMult = playerDensity < 20 ? 1.5 : playerDensity > 80 ? 0.6 : 1.0;
    const dynamicRespawnSec = Math.round(baseRespawnSec * densityMult);
    const eliteChance = Math.min(Math.round((playerDensity / 100) * 25), 35);

    logs.push({
      type: "success",
      text: `🔄 SISTEMA DE RESPAWN HÍBRIDO (respawn-rule):`
    });
    logs.push({
      type: "formula",
      text: `   - Densidade de Jogadores: ${playerDensity}% (Multiplicador de tempo: ${densityMult}x)`
    });
    logs.push({
      type: "formula",
      text: `   - Tempo de Respawn Ajustado: ${dynamicRespawnSec}s (Base: ${baseRespawnSec}s).`
    });
    logs.push({
      type: "warning",
      text: `   - Probabilidade de Surgimento de Elite devido à atividade: ${eliteChance}%`
    });

    setAuditLogs(logs);
  };

  // Loot Roller
  const handleRollLoot = () => {
    // 1. Roll rarity
    const rand = Math.random();
    let rolledRarity: "Normal" | "Magic" | "Rare" | "Epic" | "Legendary" = "Normal";

    if (rand < 0.01) rolledRarity = "Legendary";
    else if (rand < 0.05) rolledRarity = "Epic";
    else if (rand < 0.15) rolledRarity = "Rare";
    else if (rand < 0.40) rolledRarity = "Magic";
    else rolledRarity = "Normal";

    // 2. Filter matching loot from canonical list
    // biome or category matching
    const matchingLoot = CANONICAL_LOOT.filter(
      l => l.Rarity === rolledRarity || l.Biomes.includes(selectedMonster.Biome)
    );

    const selection = matchingLoot.length > 0 
      ? matchingLoot[Math.floor(Math.random() * matchingLoot.length)]
      : CANONICAL_LOOT[0];

    // Calculate value based on modifier
    const finalVal = Math.round(selection.BaseValue * selectedModifier.xpMult);

    setRolledLoot({
      lootName: selection.LootName,
      rarity: selection.Rarity,
      value: finalVal
    });

    const lootLog = {
      type: "loot",
      text: `🎁 Rarity Roll: ${rolledRarity.toUpperCase()}! Drop gerado: ${selection.LootName} (${selection.Category}) avaliado em ${finalVal} moedas.`
    };
    setAuditLogs(prev => [lootLog, ...prev]);
  };

  // Boss Phase Sim
  const triggerBossPhaseSim = () => {
    const logs: string[] = [];
    if (bossType === "world") {
      logs.push("⚔️ [WORLD BOSS] Ignisaur surge no Mundo Aberto!");
      logs.push("PvPvE Ativo: Guilda 'Crimson Dawn' inicia o combate.");
      logs.push("⚠️ Ignisaur conjura 'Lava Barrage'! Guilda rival 'Midnight Elite' ataca por trás para roubar o claim!");
      logs.push("🔥 Fase de Fúria: O boss ignora aggro básico e prioriza clérigos que curam a guilda líder.");
      logs.push("🏆 Combate encerrado: Claim concedido aos jogadores com maior contribuição de DPS/Heal geral.");
    } else {
      logs.push("🏰 [INSTANCED BOSS] Malakar a Soul Binder inicia fase de progressão.");
      logs.push("Fase 1: Escudo de Almas Ativo. Jogadores precisam quebrar totens rúnicos.");
      logs.push("🛑 Alerta de Mecânica: Malakar inicia conjuração de morte instantânea 'Soul Reap' (0.5s remaining)!");
      logs.push("🛡️ Interrupt executado com sucesso pelo Cavaleiro! Fase de dano limpa iniciada.");
      logs.push("Fase 2: Malakar se divide em 3 cópias. Healers mantêm sobrevivência do grupo em alto TTK.");
    }
    setBossSimLog(logs);
  };

  return (
    <div className="col-span-12 grid grid-cols-1 lg:grid-cols-12 gap-6" id="monster-bible-simulator-tab">
      
      {/* Settings / Config (Left Side) */}
      <div className="lg:col-span-5 space-y-6">
        
        {/* Monster Selector & Modifiers */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="monster-generator-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Skull className="w-4 h-4 text-red-500 animate-pulse" />
              1. Gerador de Criaturas do Ecossistema
            </h2>
            <span className="text-[10px] bg-slate-950 px-2 py-0.5 rounded text-slate-400 font-mono border border-slate-800">
              ECOSYSTEM-RULE
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            Escolha uma criatura para analisar seus parâmetros canônicos de combate, inteligência artificial e acoplamento biome-dependente.
          </p>

          <div className="space-y-3">
            <div className="space-y-1.5">
              <label className="text-[11px] font-mono text-slate-400">Criatura Alvo:</label>
              <select
                value={selectedMonster.MonsterId}
                onChange={(e) => {
                  const found = CANONICAL_MONSTERS.find(m => m.MonsterId === e.target.value);
                  if (found) setSelectedMonster(found);
                }}
                className="w-full bg-slate-950 border border-slate-800 text-slate-200 py-1.5 px-3 rounded text-xs focus:ring-1 focus:ring-amber-500 font-mono"
              >
                {CANONICAL_MONSTERS.map(m => (
                  <option key={m.MonsterId} value={m.MonsterId}>
                    {m.MonsterName} ({m.AggroType})
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-mono text-slate-400">Modificador Elite Dinâmico:</label>
              <div className="grid grid-cols-2 gap-2">
                {MODIFIERS.map(mod => (
                  <button
                    key={mod.name}
                    onClick={() => setSelectedModifier(mod)}
                    className={`p-2 rounded text-[10px] font-mono transition text-left border ${
                      selectedModifier.name === mod.name 
                        ? "bg-red-500/10 text-red-400 border-red-500/30" 
                        : "bg-slate-950/50 text-slate-400 border-slate-900 hover:bg-slate-900"
                    }`}
                  >
                    👑 {mod.name}
                    <span className="block text-[8px] text-slate-500 mt-0.5 font-normal leading-tight">
                      {mod.desc}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* XP Performance Contribution System */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="xp-performance-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Award className="w-4 h-4 text-amber-500" />
              2. Auditoria de Distribuição de XP
            </h2>
            <span className="text-[10px] bg-slate-950 px-2 py-0.5 rounded text-amber-400 font-mono border border-amber-900">
              XP-PERFORMANCE-RULE
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            XP não é estático. Controle a participação do seu jogador e veja o cálculo de performance em tempo real.
          </p>

          <div className="space-y-4 pt-2">
            {/* Damage Contribution */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Contribuição de Dano:</span>
                <span className="text-red-400 font-bold">{dmgContrib}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                value={dmgContrib}
                onChange={(e) => setDmgContrib(parseInt(e.target.value))}
                className="w-full accent-red-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
            </div>

            {/* Healing Contribution */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Contribuição de Cura / Suporte:</span>
                <span className="text-emerald-400 font-bold">{healContrib}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                value={healContrib}
                onChange={(e) => setHealContrib(parseInt(e.target.value))}
                className="w-full accent-emerald-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
            </div>

            {/* Participation Time */}
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Tempo de Participação:</span>
                <span className="text-violet-400 font-bold">{participationPct}%</span>
              </div>
              <input
                type="range"
                min="10"
                max="100"
                value={participationPct}
                onChange={(e) => setParticipationPct(parseInt(e.target.value))}
                className="w-full accent-violet-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
            </div>

            {/* Survival Modifier */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-mono text-slate-400 block">Condição de Sobrevivência:</label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setSurvivalMod(1.0)}
                  className={`py-1.5 rounded text-xs font-mono border transition ${
                    survivalMod === 1.0 
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" 
                      : "bg-slate-950 text-slate-500 border-slate-900"
                  }`}
                >
                  🟢 Sobreviveu (1.0x)
                </button>
                <button
                  onClick={() => setSurvivalMod(0.5)}
                  className={`py-1.5 rounded text-xs font-mono border transition ${
                    survivalMod === 0.5 
                      ? "bg-red-500/10 text-red-400 border-red-500/30" 
                      : "bg-slate-950 text-slate-500 border-slate-900"
                  }`}
                >
                  💀 Morreu (0.5x Penalidade)
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Dynamic Respawn Rules Panel */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="respawn-rules-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <RefreshCw className="w-4 h-4 text-emerald-500" />
              3. Simulador de Respawn Híbrido
            </h2>
            <span className="text-[10px] bg-slate-950 px-2 py-0.5 rounded text-emerald-400 font-mono border border-emerald-950">
              RESPAWN-RULE
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            Respawn é governado de forma híbrida para evitar bots e monopólio de pontos de caça.
          </p>

          <div className="space-y-4">
            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Atividade / Densidade do Camp:</span>
                <span className="text-amber-400 font-bold">{playerDensity}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                value={playerDensity}
                onChange={(e) => setPlayerDensity(parseInt(e.target.value))}
                className="w-full accent-amber-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
              <span className="text-[9px] text-slate-500 font-mono block leading-tight">
                Mais densidade = respawn mais rápido para sustentar pressão, mas gera riscos de elites.
              </span>
            </div>

            <div className="space-y-1.5">
              <div className="flex justify-between items-center text-[11px] font-mono">
                <span className="text-slate-400">Timer Base de Respawn (Segundos):</span>
                <span className="text-slate-300 font-bold">{baseRespawnSec}s</span>
              </div>
              <input
                type="range"
                min="10"
                max="300"
                step="10"
                value={baseRespawnSec}
                onChange={(e) => setBaseRespawnSec(parseInt(e.target.value))}
                className="w-full accent-slate-500 h-1 bg-slate-950 rounded cursor-pointer"
              />
            </div>
          </div>
        </div>

      </div>

      {/* Simulator Actions & Logs (Right Side) */}
      <div className="lg:col-span-7 space-y-6">
        
        {/* Creature Core Stats & Action Bar */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="creature-display-panel">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Shield className="w-4 h-4 text-orange-400" />
              Painel de Análise Ativa da Criatura
            </h2>
            <div className="flex items-center gap-2">
              <button
                onClick={handleRollLoot}
                className="px-4 py-1.5 rounded bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-xs uppercase font-mono transition-all flex items-center gap-1.5 shadow"
              >
                <Gift className="w-3.5 h-3.5" />
                Rolar Tabela Loot
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-center">
            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Vida Efetiva</span>
              <span className="text-base font-bold text-red-400">
                {Math.round(selectedMonster.BaseHP * selectedModifier.hpMult)} HP
              </span>
            </div>

            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Ataque Físico / Mágico</span>
              <span className="text-base font-bold text-orange-400">
                {Math.round(selectedMonster.BaseDamage * selectedModifier.dmgMult)} DMG
              </span>
            </div>

            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Cognição IA</span>
              <span className="text-xs font-bold text-violet-400 uppercase mt-1 block">
                {selectedMonster.AggroType}
              </span>
            </div>

            <div className="bg-slate-950/80 p-3 rounded-xl border border-slate-800/40 font-mono">
              <span className="text-[9px] text-slate-500 uppercase block">Elemento Principal</span>
              <span className="text-base font-bold text-emerald-400 flex items-center justify-center gap-1">
                <Flame className="w-3.5 h-3.5" />
                {selectedMonster.PrimaryElement}
              </span>
            </div>
          </div>

          {/* Elemental Modifiers Display */}
          <div className="bg-slate-950 p-4 rounded-xl border border-slate-850 grid grid-cols-1 md:grid-cols-2 gap-4 text-xs font-mono">
            <div className="space-y-1.5">
              <span className="text-[11px] text-slate-400 uppercase font-bold block">Interação de Fraquezas</span>
              <div className="flex items-center gap-1.5">
                {selectedMonster.Vulnerabilities.map(v => (
                  <span key={v} className="bg-red-500/15 text-red-400 border border-red-500/20 px-2 py-0.5 rounded text-[10px]">
                    🔥 {v} (+15% DMG)
                  </span>
                ))}
              </div>
            </div>

            <div className="space-y-1.5">
              <span className="text-[11px] text-slate-400 uppercase font-bold block">Interação de Resistências</span>
              <div className="flex items-center gap-1.5">
                {selectedMonster.Resistances.map(r => (
                  <span key={r} className="bg-slate-800 text-slate-400 border border-slate-700 px-2 py-0.5 rounded text-[10px]">
                    🛡️ {r} (-10% DMG)
                  </span>
                ))}
              </div>
            </div>
          </div>

          {/* Loot Roll Result */}
          {rolledLoot && (
            <div className="bg-amber-500/10 border border-amber-500/20 p-4 rounded-xl flex items-center justify-between text-xs font-mono">
              <div className="space-y-1">
                <span className="text-amber-500 font-bold block">★ LOOT GERADO CANONICAMENTE:</span>
                <span className="text-slate-100 text-sm font-bold flex items-center gap-1">
                  🎁 {rolledLoot.lootName} ({rolledLoot.rarity})
                </span>
              </div>
              <div className="text-right">
                <span className="text-slate-400 text-[10px] block">Valor de Mercado</span>
                <span className="text-amber-400 font-bold">{rolledLoot.value} moedas</span>
              </div>
            </div>
          )}
        </div>

        {/* Live Mathematical Audit Logs */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-3" id="ecosystem-audit-logs">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Eye className="w-4 h-4 text-amber-500" />
              Cálculo de Fórmulas e Logs Ativos de Auditoria
            </h2>
            <button
              onClick={() => {
                setAuditLogs([]);
                runEcosystemCalculation();
              }}
              className="text-slate-500 hover:text-slate-300 transition-all font-mono text-[10px] flex items-center gap-1"
            >
              <RotateCcw className="w-3 h-3" /> Limpar Logs
            </button>
          </div>

          <div className="bg-slate-950 p-4 rounded-xl border border-slate-900 h-[220px] overflow-y-auto space-y-2 font-mono text-[11px] leading-relaxed">
            {auditLogs.length === 0 ? (
              <div className="text-slate-600 italic text-center pt-16">Nenhum log gerado.</div>
            ) : (
              auditLogs.map((log, index) => (
                <div
                  key={index}
                  className={`p-2 rounded border transition ${
                    log.type === "warning" ? "bg-red-500/10 text-red-300 border-red-500/20" :
                    log.type === "success" ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20 font-bold" :
                    log.type === "loot" ? "bg-amber-500/10 text-amber-400 border-amber-500/30 font-bold text-xs" :
                    "bg-slate-900/40 text-slate-300 border-slate-800/40"
                  }`}
                >
                  {log.text}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Boss Encounter Simulator Card */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="boss-simulator-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <Zap className="w-4 h-4 text-violet-400 animate-bounce" />
              4. Simulador de Arquitetura de Bosses (PvPvE)
            </h2>
            <span className="text-[10px] bg-slate-950 px-2 py-0.5 rounded text-violet-400 font-mono border border-violet-950">
              BOSS-SYSTEM-RULE
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-normal">
            Teste os dois tipos de arquitetura de bosses: World Boss (com PvP/Mundo Aberto disputado) ou Boss Instanciado (focado em mecânicas/fases).
          </p>

          <div className="grid grid-cols-2 gap-2">
            <button
              onClick={() => setBossType("world")}
              className={`p-3 rounded-xl border text-xs font-mono font-bold transition flex flex-col items-center gap-1 ${
                bossType === "world"
                  ? "bg-amber-500/20 text-amber-400 border-amber-500/40"
                  : "bg-slate-950 text-slate-500 border-slate-900 hover:bg-slate-900"
              }`}
            >
              👑 World Boss (Contestado)
              <span className="text-[9px] font-normal text-slate-500">PvPvE & Disputa de Clãs</span>
            </button>

            <button
              onClick={() => setBossType("instanced")}
              className={`p-3 rounded-xl border text-xs font-mono font-bold transition flex flex-col items-center gap-1 ${
                bossType === "instanced"
                  ? "bg-violet-500/20 text-violet-400 border-violet-500/40"
                  : "bg-slate-950 text-slate-500 border-slate-900 hover:bg-slate-900"
              }`}
            >
              🏰 Instanced Boss (Progression)
              <span className="text-[9px] font-normal text-slate-500">Party Segura & Fases</span>
            </button>
          </div>

          <div className="pt-2 flex justify-between items-center">
            <span className="text-xs text-slate-400 font-mono">Disparar Mecânicas e Fases:</span>
            <button
              onClick={triggerBossPhaseSim}
              className="px-4 py-1 bg-violet-600 hover:bg-violet-500 text-slate-100 font-bold text-xs uppercase font-mono rounded transition shadow"
            >
              Simular Encontro
            </button>
          </div>

          {bossSimLog.length > 0 && (
            <div className="bg-slate-950 p-4 rounded-xl border border-slate-900 space-y-1.5 font-mono text-[11px]">
              {bossSimLog.map((line, idx) => (
                <div key={idx} className="text-slate-300 flex gap-2">
                  <span className="text-violet-400 font-bold shrink-0">[{idx + 1}]</span>
                  <span>{line}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Canonical Rules List Checklist */}
        <div className="bg-slate-900/60 rounded-2xl border border-slate-800/80 p-5 space-y-4" id="monster-checklist-card">
          <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
            <h2 className="text-sm font-bold uppercase text-amber-400 font-mono flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-emerald-500" />
              Regras de Ecossistema & Monstros Homologadas
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {[
              { id: "monster-ecosystem-rule", title: "monster-ecosystem-rule", desc: "Monstros distribuídos dinamicamente sem Tiers fixos (emergent level)." },
              { id: "monster-ai-rule", title: "monster-ai-rule", desc: "Engine cognitiva híbrida (aggro simples vs tático para Elites/Bosses)." },
              { id: "elemental-interaction-rule", title: "elemental-interaction-rule", desc: "Interações elementais suaves (+15% vulnerável / -10% resistente)." },
              { id: "xp-performance-rule", title: "xp-performance-rule", desc: "XP calculado em tempo real com base em dano, cura e tempo de vida." },
              { id: "loot-economy-rule", title: "loot-economy-rule", desc: "Loot sustentável para alto TTK, controlando inflação de moedas e raridade." },
              { id: "respawn-rule", title: "respawn-rule", desc: "Fórmula híbrida de respawn influenciada por densidade de caçadores." },
              { id: "boss-system-rule", title: "boss-system-rule", desc: "Categorização em World Boss (público PvPvE) vs Instanciado (party fechada)." }
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
