import React, { useState, useEffect, useRef, Suspense } from "react";
import {
  Folder,
  FolderOpen,
  FileCode,
  FileText,
  Settings,
  Terminal,
  Copy,
  Check,
  Play,
  ArrowRight,
  ChevronRight,
  Wifi,
  WifiOff,
  UserCheck,
  CheckCircle,
  Gamepad2,
  LogOut,
  Database,
  Activity,
  Download,
  BookOpen,
  Info,
  Shield,
  RefreshCw,
  Layers,
  Sparkles,
  Network,
  Sword,
  Map,
  Compass,
  Clock,
  AlertTriangle,
  Gift,
  XCircle,
  Flame,
  Snowflake,
  Moon,
  Leaf,
  Lock,
  Unlock,
  MapPin,
  Coins,
  Hammer,
  Users,
  Home,
  UserPlus,
  ArrowLeftRight,
  Landmark,
  Ban,
  Scale,
  Sliders,
  Globe,
  Wand2
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { codeFiles, FOLDER_STRUCTURE_TEXT, CodeFile } from "./codeTemplates";
import { GameBootstrapRoot } from "./core/components/GameBootstrapRoot";

// Global Switch: "CORE" (Gameplay de Produção) | "DEV" (Ferramentas de Debug & Simuladores)
export const GAME_MODE: "CORE" | "DEV" = "DEV";

// Lazy-loaded simulators to avoid importing them statically in CORE mode
const VocationSimulator = React.lazy(() => import("./devtools/VocationSimulator").then(m => ({ default: m.VocationSimulator })));
const SpellSkillSimulator = React.lazy(() => import("./devtools/SpellSkillSimulator").then(m => ({ default: m.SpellSkillSimulator })));
const WorldSimulator = React.lazy(() => import("./devtools/WorldSimulator").then(m => ({ default: m.WorldSimulator })));
const MonsterSimulator = React.lazy(() => import("./devtools/MonsterSimulator").then(m => ({ default: m.MonsterSimulator })));
const GameEntrySimulator = React.lazy(() => import("./devtools/GameEntrySimulator").then(m => ({ default: m.GameEntrySimulator })));

// Define local types for simulated logs
interface LogLine {
  id: string;
  timestamp: string;
  type: "info" | "warning" | "error";
  text: string;
}

// Map states to color schemes and icons for visual representation
const STATE_DETAILS = {
  Boot: {
    title: "Boot",
    color: "from-zinc-500 to-slate-600",
    border: "border-zinc-500",
    text: "text-zinc-400",
    bg: "bg-zinc-500/10",
    shadow: "shadow-zinc-500/10",
    icon: Settings,
    desc: "Inicialização primária de motores e carregamento de configurações locais."
  },
  Loading: {
    title: "Loading",
    color: "from-indigo-500 to-blue-600",
    border: "border-indigo-500",
    text: "text-indigo-400",
    bg: "bg-indigo-500/10",
    shadow: "shadow-indigo-500/10",
    icon: RefreshCw,
    desc: "Carregamento assíncrono de assets pesados e cenas de transição na thread de background."
  },
  Menu: {
    title: "Menu",
    color: "from-violet-500 to-purple-600",
    border: "border-violet-500",
    text: "text-violet-400",
    bg: "bg-violet-500/10",
    shadow: "shadow-violet-500/10",
    icon: BookOpen,
    desc: "Menu Principal ativo. Aguardando interação do usuário ou inputs de credenciais."
  },
  Connecting: {
    title: "Connecting",
    color: "from-sky-500 to-blue-600",
    border: "border-sky-500",
    text: "text-sky-400",
    bg: "bg-sky-500/10",
    shadow: "shadow-sky-500/10",
    icon: Network,
    desc: "Abertura de sockets TCP assíncronas e handshake inicial de autenticação de pacotes."
  },
  CharacterSelection: {
    title: "CharacterSelection",
    color: "from-fuchsia-500 to-pink-600",
    border: "border-fuchsia-500",
    text: "text-fuchsia-400",
    bg: "bg-fuchsia-500/10",
    shadow: "shadow-fuchsia-500/10",
    icon: UserCheck,
    desc: "Exibição de avatares disponíveis. Sincronização de metadados do personagem selecionado."
  },
  InGame: {
    title: "InGame",
    color: "from-amber-500 to-yellow-600",
    border: "border-amber-500",
    text: "text-amber-400",
    bg: "bg-amber-500/10",
    shadow: "shadow-amber-500/20",
    icon: Gamepad2,
    desc: "Loop de gameplay ativo. Sincronização posicional em tempo real e trocas de keepalive."
  },
  Disconnected: {
    title: "Disconnected",
    color: "from-rose-500 to-red-600",
    border: "border-rose-500",
    text: "text-rose-400",
    bg: "bg-rose-500/10",
    shadow: "shadow-rose-500/10",
    icon: WifiOff,
    desc: "Interrupção de soquete ou falha de conexão de rede. Executa buffers de limpeza e avisa a UI."
  },
  Shutdown: {
    title: "Shutdown",
    color: "from-neutral-700 to-zinc-800",
    border: "border-zinc-700",
    text: "text-zinc-500",
    bg: "bg-zinc-700/10",
    shadow: "shadow-zinc-700/5",
    icon: LogOut,
    desc: "Persistência final de preferências locais (JSON), encerramento de rede e liberação do motor Godot."
  }
} as const;

type AppStateType = keyof typeof STATE_DETAILS;

// ==========================================
// ITEMIZATION EXPANSION SIMULATOR COMPONENT
// ==========================================
function ItemizationSimulator() {
  const [selectedWeapon, setSelectedWeapon] = useState("sword");
  const [selectedArmor, setSelectedArmor] = useState("heavy");

  // Dual Wield Simulator States
  const [dwClass, setDwClass] = useState<"Knight" | "Assassin">("Knight");
  const [dwMode, setDwMode] = useState<"single" | "dual">("single");
  const [mainHandPower, setMainHandPower] = useState(100);
  const [offHandPower, setOffHandPower] = useState(80);

  // Affix Simulator States
  const [selectedTier, setSelectedTier] = useState<1 | 2 | 3 | 4 | 5>(3);
  const [rolledItemName, setRolledItemName] = useState("Pyre-Forged Falchion");
  const [rolledItemType, setRolledItemType] = useState("Sword");
  const [rolledItemArchetype, setRolledItemArchetype] = useState("Slashing");
  const [rolledAffixes, setRolledAffixes] = useState<Array<{ key: string; name: string; category: string; value: string }>>([
    { key: "atk", name: "ATK", category: "Offensive", value: "+35 Flat Power" },
    { key: "fire_attack", name: "Fire Attack", category: "Elemental Offense", value: "+18 Fire Damage" },
    { key: "cooldown_reduction", name: "Cooldown Reduction", category: "Resource", value: "4.5% CDR" }
  ]);
  const [rollHistory, setRollHistory] = useState<string[]>([]);

  // ==========================================
  // COMBAT FORMULA BIBLE SIMULATOR DECLARATIONS
  // ==========================================
  const CANONICAL_WEAPONS = [
    { id: "sword_t1_rusty", name: "Rusty Recruit Sword (T1)", baseDamage: 15.0, critBonus: 0.00, classes: ["Knight", "Assassin", "Archer"] },
    { id: "sword_t2_steel", name: "Tempered Steel Broadsword (T2)", baseDamage: 35.0, critBonus: 0.02, classes: ["Knight", "Assassin", "Archer"] },
    { id: "sword_t3_flame", name: "Pyre-Forged Falchion (T3)", baseDamage: 58.0, critBonus: 0.04, classes: ["Knight", "Assassin", "Archer"] },
    { id: "sword_t4_sunshard", name: "Sunshard Gladius (T4)", baseDamage: 102.0, critBonus: 0.06, classes: ["Knight", "Assassin", "Archer"] },
    { id: "sword_t5_calamity", name: "Aegis-Breaker Calamity (T5)", baseDamage: 165.0, critBonus: 0.12, classes: ["Knight"] },
    { id: "axe_t2_battle", name: "Iron Battleaxe (T2)", baseDamage: 42.0, critBonus: 0.00, classes: ["Knight"] },
    { id: "axe_t4_volcanic", name: "Magma Core Decapitator (T4)", baseDamage: 135.0, critBonus: 0.04, classes: ["Knight"] },
    { id: "bow_t1_short", name: "Oak Shortbow (T1)", baseDamage: 12.0, critBonus: 0.03, classes: ["Archer"] },
    { id: "bow_t3_composite", name: "Reinforced Composite Bow (T3)", baseDamage: 52.0, critBonus: 0.07, classes: ["Archer"] },
    { id: "bow_t5_starsong", name: "Astral Starsong (T5)", baseDamage: 145.0, critBonus: 0.18, classes: ["Archer"] },
    { id: "dagger_t2_stiletto", name: "Poisoner Stiletto (T2)", baseDamage: 26.0, critBonus: 0.08, classes: ["Assassin"] },
    { id: "dagger_t4_nightshade", name: "Nightshade Carver (T4)", baseDamage: 85.0, critBonus: 0.15, classes: ["Assassin"] },
    { id: "staff_t1_apprentice", name: "Apprentice Ash Staff (T1)", baseDamage: 8.0, critBonus: 0.01, classes: ["Mage"] },
    { id: "staff_t3_glacial", name: "Glacial Sceptre (T3)", baseDamage: 35.0, critBonus: 0.03, classes: ["Mage"] },
    { id: "staff_t5_voidlord", name: "Voidlord Greatstaff (T5)", baseDamage: 90.0, critBonus: 0.08, classes: ["Mage"] }
  ];

  const CANONICAL_SPELLS = [
    { id: "slash_strike", name: "Slash Strike", classId: "Knight", basePower: 25, elementalType: "Physical", manaCost: 10, cooldown: 2.0 },
    { id: "shield_bash", name: "Shield Bash", classId: "Knight", basePower: 35, elementalType: "Physical", manaCost: 15, cooldown: 4.0 },
    { id: "fireball", name: "Fireball", classId: "Mage", basePower: 60, elementalType: "Fire", manaCost: 25, cooldown: 3.0 },
    { id: "ice_barrage", name: "Ice Barrage", classId: "Mage", basePower: 45, elementalType: "Ice", manaCost: 20, cooldown: 2.5 },
    { id: "piercing_shot", name: "Piercing Shot", classId: "Archer", basePower: 35, elementalType: "Physical", manaCost: 12, cooldown: 2.5 },
    { id: "rain_of_arrows", name: "Rain of Arrows", classId: "Archer", basePower: 55, elementalType: "Physical", manaCost: 25, cooldown: 6.0 },
    { id: "shadow_stab", name: "Shadow Stab", classId: "Assassin", basePower: 40, elementalType: "Shadow", manaCost: 15, cooldown: 2.0 },
    { id: "poison_dart", name: "Poison Dart", classId: "Assassin", basePower: 25, elementalType: "Nature", manaCost: 18, cooldown: 4.0 },
    { id: "holy_heal", name: "Holy Heal", classId: "Cleric", basePower: 50, elementalType: "Holy", manaCost: 20, cooldown: 3.0 },
    { id: "divine_burst", name: "Divine Burst", classId: "Cleric", basePower: 30, elementalType: "Holy", manaCost: 15, cooldown: 2.5 }
  ];

  const [combClass, setCombClass] = useState<"Knight" | "Assassin" | "Archer" | "Mage" | "Cleric">("Knight");
  const [combWeapon, setCombWeapon] = useState<string>("sword_t3_flame");
  const [combSpell, setCombSpell] = useState<string>("slash_strike");
  const [combDualWield, setCombDualWield] = useState<boolean>(false);
  const [combAttackerAffinity, setCombAttackerAffinity] = useState<number>(15);
  const [combCritForced, setCombCritForced] = useState<boolean>(false);

  const [combArmor, setCombArmor] = useState<number>(350);
  const [combFireRes, setCombFireRes] = useState<number>(15);
  const [combIceRes, setCombIceRes] = useState<number>(15);
  const [combHolyRes, setCombHolyRes] = useState<number>(15);
  const [combShadowRes, setCombShadowRes] = useState<number>(15);
  const [combNatureRes, setCombNatureRes] = useState<number>(15);
  const [combAffinityDefense, setCombAffinityDefense] = useState<boolean>(true);

  const [combatLogs, setCombatLogs] = useState<Array<{ text: string; type: "info" | "success" | "warning" | "formula" }>>([]);

  const simulateHit = () => {
    const weapon = CANONICAL_WEAPONS.find(w => w.id === combWeapon) || CANONICAL_WEAPONS[0];
    const spell = CANONICAL_SPELLS.find(s => s.id === combSpell) || CANONICAL_SPELLS[0];

    const logs: Array<{ text: string; type: "info" | "success" | "warning" | "formula" }> = [];
    logs.push({ text: `⚔️ [COMBAT INICIADO] Atacante (${combClass}) usando arma [${weapon.name}] e habilidade [${spell.name}].`, type: "info" });

    let weaponDmg = weapon.baseDamage;
    logs.push({ text: `[Arma Primária] Dano Base da Arma: ${weaponDmg}`, type: "info" });

    if (combDualWield && (combClass === "Knight" || combClass === "Assassin")) {
      const offhandDmg = weaponDmg;
      const offhandEffective = offhandDmg * 0.75;
      weaponDmg = weaponDmg + offhandEffective;
      logs.push({ text: `[Dual Wield] Regra de empunhadura dupla ativa. Offhand contribui com 75% (${offhandDmg} × 0.75 = ${offhandEffective.toFixed(1)}). Dano de Arma Total: ${weaponDmg.toFixed(1)}`, type: "warning" });
    } else if (combDualWield) {
      logs.push({ text: `[Dual Wield] Classe ${combClass} não possui autorização canônica para Dual Wield. Ignorando offhand.`, type: "warning" });
    }

    const skillPower = spell.basePower;
    logs.push({ text: `[Habilidade] Base Power da Habilidade [${spell.name}]: ${skillPower}`, type: "info" });

    const rawOutput = weaponDmg + skillPower;
    logs.push({ text: `[Fórmula Primária] Raw Damage = (Weapon Damage + Skill Damage) = (${weaponDmg.toFixed(1)} + ${skillPower}) = ${rawOutput.toFixed(1)}`, type: "formula" });

    const element = spell.elementalType;
    let defenderRes = 0;
    if (element === "Fire") defenderRes = combFireRes;
    else if (element === "Ice") defenderRes = combIceRes;
    else if (element === "Holy") defenderRes = combHolyRes;
    else if (element === "Shadow") defenderRes = combShadowRes;
    else if (element === "Nature") defenderRes = combNatureRes;

    const affinityDefenseBonusActive = combAffinityDefense && element !== "Physical";
    const effectiveDefenderRes = defenderRes + (affinityDefenseBonusActive ? 4 : 0);
    const attackerAffinityBonus = combAttackerAffinity;

    const elementalModifier = element === "Physical" ? 1.0 : (1.0 + (attackerAffinityBonus - effectiveDefenderRes) * 0.01);
    let elementalDamage = rawOutput * (elementalModifier);

    if (element !== "Physical") {
      logs.push({
        text: `[Elemento: ${element}] Afinidade Atacante: +${attackerAffinityBonus}%, Resistência Defensor: ${defenderRes}%${affinityDefenseBonusActive ? " + 4% (Affinity Defense awakened)" : ""}.`,
        type: "info"
      });
      logs.push({
        text: `[Elemento: ${element}] Elemental Modifier = 1.0 + (AttackerAffinity - DefenderResistance) * 0.01 = 1.0 + (${attackerAffinityBonus} - ${effectiveDefenderRes}) * 0.01 = ${elementalModifier.toFixed(3)}x`,
        type: "formula"
      });
      logs.push({
        text: `[Dano Elemental] Dano Amplificado Elementalmente: ${rawOutput.toFixed(1)} × ${elementalModifier.toFixed(3)} = ${elementalDamage.toFixed(1)}`,
        type: "info"
      });
    } else {
      logs.push({ text: `[Elemento: Physical] Sem bônus de afinidade elemental ou resistências aplicados. Modifier: 1.000x`, type: "info" });
    }

    const baseCritChance = 5;
    const weaponCritBonus = weapon.critBonus * 100;
    const totalCritChance = baseCritChance + weaponCritBonus;

    const isCrit = combCritForced || (Math.random() * 100 < totalCritChance);
    const critMultiplier = combClass === "Assassin" ? 2.2 : 1.5;

    let critDamage = elementalDamage;
    if (isCrit) {
      critDamage = elementalDamage * critMultiplier;
      logs.push({ text: `✨ [CRITICAL HIT!] Chance de Crítico: ${totalCritChance}% (${baseCritChance}% Base + ${weaponCritBonus.toFixed(1)}% Arma). Multiplicador de Classe (${combClass === "Assassin" ? "Assassin Override" : "Normal"}): ${critMultiplier}x.`, type: "success" });
      logs.push({ text: `[Multiplicador Crítico] Dano Crítico = ${elementalDamage.toFixed(1)} × ${critMultiplier} = ${critDamage.toFixed(1)}`, type: "formula" });
    } else {
      logs.push({ text: `[Critical Roll] Chance de Crítico: ${totalCritChance}% (${baseCritChance}% Base + ${weaponCritBonus.toFixed(1)}% Arma). Golpe normal desferido.`, type: "info" });
    }

    const k = 250;
    const armorMitigation = combArmor / (combArmor + k);
    const afterArmorDamage = critDamage * (1 - armorMitigation);

    // Explicit Structured Logs as per canonical requirements
    logs.push({
      text: `💥 [Raw Damage] Dano Bruto do Atacante: ${critDamage.toFixed(1)} HP (Arma: ${weaponDmg.toFixed(1)} + Habilidade: ${skillPower.toFixed(1)}${isCrit ? `, Crítico ${critMultiplier}x aplicado` : ""})`,
      type: "formula"
    });

    logs.push({
      text: `🛡️ [Armor Mitigation %] Armadura Física Total: ${combArmor} (K=250) ➔ Redução Física: ${(armorMitigation * 100).toFixed(2)}%`,
      type: "formula"
    });

    logs.push({
      text: `🛡️ [Mitigated Damage] Dano após Mitigação Física: ${afterArmorDamage.toFixed(1)} HP (Dano mitigado: ${(critDamage * armorMitigation).toFixed(1)} HP)`,
      type: "info"
    });

    const hasElement = element !== "Physical";
    const eleResPercent = hasElement ? effectiveDefenderRes : 0;
    logs.push({
      text: `🔮 [Elemental Resistance] Resistência Elemental a ${element}: ${eleResPercent}% (${hasElement ? `${defenderRes}% base + ${affinityDefenseBonusActive ? "4% Afinidade Awakened" : "0%"}` : "N/A - Dano Físico Puro"})`,
      type: "info"
    });

    let finalDamage = afterArmorDamage;
    if (hasElement) {
      const totalEleResFraction = effectiveDefenderRes / 100;
      finalDamage = afterArmorDamage * (1 - totalEleResFraction);
    }

    logs.push({
      text: `⚔️ [Final Damage Taken] Dano Líquido Final Recebido: ${finalDamage.toFixed(1)} HP (TTK verificado)`,
      type: "success"
    });

    logs.push({ text: `💥 GOLPE SIMULADO CONCLUÍDO: O defensor recebeu exatamente ${finalDamage.toFixed(1)} de dano líquido.`, type: "success" });

    setCombatLogs(logs);
  };

  useEffect(() => {
    const validWeapons = CANONICAL_WEAPONS.filter(w => w.classes.includes(combClass));
    if (validWeapons.length > 0) {
      setCombWeapon(validWeapons[0].id);
    }
    const validSpells = CANONICAL_SPELLS.filter(s => s.classId === combClass);
    if (validSpells.length > 0) {
      setCombSpell(validSpells[0].id);
    }
    if (combClass !== "Knight" && combClass !== "Assassin") {
      setCombDualWield(false);
    }
  }, [combClass]);

  useEffect(() => {
    simulateHit();
  }, [
    combClass, combWeapon, combSpell, combDualWield, combAttackerAffinity,
    combCritForced, combArmor, combFireRes, combIceRes, combHolyRes,
    combShadowRes, combNatureRes, combAffinityDefense
  ]);

  const WEAPON_ARCHETYPES = [
    { id: "sword", name: "Sword", weaponType: "Sword", damageArchetype: "Slashing", type: "Melee", desc: "Balanced physical weapon with versatile offensive profile.", classes: ["Knight", "Assassin", "Archer"], speed: "Medium" },
    { id: "axe", name: "Axe", weaponType: "Axe", damageArchetype: "Slashing", type: "Melee", desc: "Slow heavy weapon with high burst damage.", classes: ["Knight"], speed: "Slow" },
    { id: "mace", name: "Mace", weaponType: "Mace", damageArchetype: "Bludgeoning", type: "Melee", desc: "Blunt weapon specialized in stagger and anti-heavy-armor combat.", classes: ["Knight"], speed: "Slow" },
    { id: "spear", name: "Spear", weaponType: "Spear", damageArchetype: "Piercing", type: "Melee", desc: "Extended melee reach and zoning capability.", classes: ["Knight", "Archer"], speed: "Medium" },
    { id: "dagger", name: "Dagger", weaponType: "Dagger", damageArchetype: "Piercing", type: "Melee", desc: "Very high attack speed and crit-oriented burst.", classes: ["Assassin"], speed: "Very Fast" },
    { id: "bow", name: "Bow", weaponType: "Bow", damageArchetype: "Ranged Physical", type: "Ranged", desc: "Long-range sustained physical DPS.", classes: ["Archer"], speed: "Fast" },
    { id: "crossbow", name: "Crossbow", weaponType: "Crossbow", damageArchetype: "Ranged Physical", type: "Ranged", desc: "Slower than bow but higher burst and armor penetration.", classes: ["Archer"], speed: "Slow" },
    { id: "staff", name: "Staff", weaponType: "Staff", damageArchetype: "Magical", type: "Magic", desc: "Primary spell-power weapon with strong AoE scaling.", classes: ["Mage", "Cleric"], speed: "Slow" },
    { id: "wand", name: "Wand", weaponType: "Wand", damageArchetype: "Magical", type: "Magic", desc: "Fast-casting magical weapon with mana efficiency.", classes: ["Mage", "Cleric"], speed: "Fast" },
    { id: "tome", name: "Tome / Relic", weaponType: "Tome / Relic", damageArchetype: "Magical", type: "Magic", desc: "Support-oriented magical focus for healing and utility scaling.", classes: ["Mage", "Cleric"], speed: "Medium" }
  ];

  const ARMOR_ARCHETYPES = [
    { id: "heavy", name: "Heavy Armor", desc: "Provides the highest physical defense, low dodge, moderate to low magic resistance, and prioritizes survivability in frontline combat.", classes: ["Knight"], stats: "High Physical Def, Low Dodge, Moderate Magic Resist" },
    { id: "light", name: "Light Armor", desc: "Provides medium physical defense, high dodge, mobility-oriented bonuses, and offensive utility.", classes: ["Archer", "Assassin"], stats: "Medium Physical Def, High Dodge, Offensive Speed" },
    { id: "cloth", name: "Cloth Armor", desc: "Provides low physical defense, high magic resistance, high mana synergy, and strong spell-oriented scaling.", classes: ["Mage", "Cleric"], stats: "Low Physical Def, High Magic Resist, High Mana Scaling" }
  ];

  const AFFIX_DATA = {
    offensive: [
      { key: "atk", name: "ATK", suffix: "Flat Power" },
      { key: "crit_chance", name: "Crit Chance", suffix: "% Chance" },
      { key: "crit_damage", name: "Crit Damage", suffix: "% Modifier" },
      { key: "armor_penetration", name: "Armor Penetration", suffix: "% Penetration" },
      { key: "attack_speed", name: "Attack Speed", suffix: "% Speed" },
      { key: "skill_power", name: "Skill Power", suffix: "% Amplification" }
    ],
    defensive: [
      { key: "armor", name: "Armor", suffix: "Defense Pts" },
      { key: "magic_resist", name: "Magic Resist", suffix: "Resistance Pts" },
      { key: "dodge", name: "Dodge", suffix: "% Evasion" },
      { key: "block_chance", name: "Block Chance", suffix: "% Shield Block" },
      { key: "max_hp", name: "Max HP", suffix: "Hit Points" }
    ],
    resource: [
      { key: "max_mana", name: "Max Mana", suffix: "Mana Pts" },
      { key: "mana_regeneration", name: "Mana Regeneration", suffix: "per Sec" },
      { key: "hp_regeneration", name: "HP Regeneration", suffix: "per Sec" },
      { key: "cooldown_reduction", name: "Cooldown Reduction", suffix: "% CDR" }
    ],
    elemental_offense: [
      { key: "fire_attack", name: "Fire Attack", suffix: "Fire Damage" },
      { key: "ice_attack", name: "Ice Attack", suffix: "Ice Damage" },
      { key: "holy_attack", name: "Holy Attack", suffix: "Holy Damage" },
      { key: "shadow_attack", name: "Shadow Attack", suffix: "Shadow Damage" },
      { key: "nature_attack", name: "Nature Attack", suffix: "Nature Damage" }
    ],
    elemental_defense: [
      { key: "fire_resist", name: "Fire Resist", suffix: "% Fire Ward" },
      { key: "ice_resist", name: "Ice Resist", suffix: "% Ice Ward" },
      { key: "holy_resist", name: "Holy Resist", suffix: "% Holy Ward" },
      { key: "shadow_resist", name: "Shadow Resist", suffix: "% Shadow Ward" },
      { key: "nature_resist", name: "Nature Resist", suffix: "% Nature Ward" }
    ]
  };

  // Determine affix count rules per tier
  const getTierAffixRange = (t: number) => {
    switch (t) {
      case 1: return { min: 1, max: 2, desc: "1 to 2 minor affixes" };
      case 2: return { min: 2, max: 3, desc: "2 to 3 affixes" };
      case 3: return { min: 3, max: 4, desc: "3 to 4 moderate affixes" };
      case 4: return { min: 4, max: 5, desc: "4 to 5 powerful affixes" };
      case 5: return { min: 5, max: 6, desc: "5 to 6 major affixes (Artifact)" };
      default: return { min: 1, max: 2, desc: "1 to 2 minor affixes" };
    }
  };

  const currentWeaponObj = WEAPON_ARCHETYPES.find(w => w.id === selectedWeapon) || WEAPON_ARCHETYPES[0];
  const currentArmorObj = ARMOR_ARCHETYPES.find(a => a.id === selectedArmor) || ARMOR_ARCHETYPES[0];

  // Dual Wield calculation
  const dwEffectivePower = dwMode === "dual" 
    ? (mainHandPower * 1.0) + (offHandPower * 0.75) 
    : mainHandPower;

  // Roll Affixes Handler
  const handleRollAffixes = () => {
    const range = getTierAffixRange(selectedTier);
    const count = Math.floor(Math.random() * (range.max - range.min + 1)) + range.min;

    // Pick random weapon archetype name for flavor
    const weaponsFlavor = [
      { prefix: "Grievous", base: "Edge", type: "Sword", archetype: "Slashing" },
      { prefix: "Void-Fused", base: "Calamity", type: "Sword", archetype: "Slashing" },
      { prefix: "Volcanic", base: "Cleaver", type: "Axe", archetype: "Slashing" },
      { prefix: "Hallowed", base: "Mace", type: "Mace", archetype: "Bludgeoning" },
      { prefix: "Sunshard", base: "Halberd", type: "Spear", archetype: "Piercing" },
      { prefix: "Nightshade", base: "Carver", type: "Dagger", archetype: "Piercing" },
      { prefix: "Astral", base: "Starsong", type: "Bow", archetype: "Ranged Physical" },
      { prefix: "Heavy Steel", base: "Ballista", type: "Crossbow", archetype: "Ranged Physical" },
      { prefix: "Glacial", base: "Spire", type: "Staff", archetype: "Magical" },
      { prefix: "Zephyr", base: "Conduit", type: "Wand", archetype: "Magical" },
      { prefix: "Sacred", base: "Reliquary", type: "Tome / Relic", archetype: "Magical" }
    ];

    const randomWeapon = weaponsFlavor[Math.floor(Math.random() * weaponsFlavor.length)];
    const rollName = `${randomWeapon.prefix} ${randomWeapon.base} (T${selectedTier})`;
    setRolledItemName(rollName);
    setRolledItemType(randomWeapon.type);
    setRolledItemArchetype(randomWeapon.archetype);

    // Roll random affixes
    const categories = Object.keys(AFFIX_DATA) as Array<keyof typeof AFFIX_DATA>;
    const rolled: Array<{ key: string; name: string; category: string; value: string }> = [];
    const chosenKeys = new Set<string>();

    for (let i = 0; i < count; i++) {
      // Find a non-duplicate affix
      let attempts = 0;
      while (attempts < 50) {
        const cat = categories[Math.floor(Math.random() * categories.length)];
        const pool = AFFIX_DATA[cat];
        const affix = pool[Math.floor(Math.random() * pool.length)];

        if (!chosenKeys.has(affix.key)) {
          chosenKeys.add(affix.key);
          
          // Generate value based on tier
          let val = "";
          const tierMult = selectedTier;
          if (affix.suffix.includes("%")) {
            const min = tierMult * 1.5;
            const max = tierMult * 3.0;
            const floatVal = (Math.random() * (max - min) + min).toFixed(1);
            val = `${floatVal}${affix.suffix}`;
          } else {
            const min = tierMult * 15;
            const max = tierMult * 35;
            const intVal = Math.floor(Math.random() * (max - min) + min);
            val = `+${intVal} ${affix.suffix}`;
          }

          rolled.push({
            key: affix.key,
            name: affix.name,
            category: cat.replace("_", " ").toUpperCase(),
            value: val
          });
          break;
        }
        attempts++;
      }
    }

    setRolledAffixes(rolled);
    setRollHistory(prev => [
      `Rolled T${selectedTier} ${randomWeapon.type} with ${rolled.length} affixes`,
      ...prev.slice(0, 4)
    ]);
  };

  return (
    <div className="xl:col-span-12 grid grid-cols-1 xl:grid-cols-12 gap-6">
      
      {/* Top Banner: Canonical Declarations */}
      <div className="xl:col-span-12 bg-slate-900/60 border border-slate-800/60 p-6 rounded-2xl backdrop-blur-sm shadow-xl space-y-4">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="bg-amber-500/10 text-amber-400 text-[10px] font-bold font-mono px-2.5 py-1 rounded border border-amber-500/20 uppercase tracking-wider">
                Itemization Expansion Patch • Canonizado v2.0
              </span>
              <span className="bg-indigo-500/10 text-indigo-400 text-[10px] font-bold font-mono px-2.5 py-1 rounded border border-indigo-500/20 uppercase tracking-wider">
                10 Weapons + 3 Armors
              </span>
            </div>
            <h2 className="text-2xl font-extrabold tracking-tight bg-gradient-to-r from-slate-100 via-amber-200 to-slate-300 bg-clip-text text-transparent mt-2">
              Itemization Expansion Bible & Simulator
            </h2>
            <p className="text-xs text-slate-400 max-w-4xl leading-relaxed mt-1">
              Validate and play with weapon archetypes, armor specializations, dual wielding formulas, and the global affix pool.
              No artificial upgrades exist (<code className="text-slate-300">item-upgrade-rule</code>) to protect item scarcity and organic market trade.
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs font-mono bg-slate-950/40 p-3 rounded-lg border border-slate-900/60">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
            <span className="text-slate-400">Canonical Rules Compliance Verified</span>
          </div>
        </div>
      </div>

      {/* Row 1: Weapon & Armor Archetype Explorers */}
      <div className="xl:col-span-6 bg-slate-950/80 border border-slate-850 rounded-2xl p-5 space-y-5 shadow-2xl relative">
        <div className="border-b border-slate-900 pb-3 flex items-center justify-between">
          <div>
            <span className="text-[10px] uppercase font-bold text-amber-500 font-mono tracking-wider">Canonical Archetypes</span>
            <h3 className="text-base font-bold text-slate-200 flex items-center gap-2 mt-0.5">
              <Sword className="w-5 h-5 text-amber-400" />
              <span>10 Weapon Archetypes Bible</span>
            </h3>
          </div>
          <span className="text-[11px] font-mono text-slate-500">weapon-archetype-rule</span>
        </div>

        <p className="text-xs text-slate-400 leading-relaxed font-mono">
          Weapons define combat identity, scaling patterns, attack profiles, and build diversity. Select a canonical archetype to view its master profile:
        </p>

        {/* Weapons Grid Selectors */}
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
          {WEAPON_ARCHETYPES.map((w) => (
            <button
              key={w.id}
              onClick={() => setSelectedWeapon(w.id)}
              className={`p-2.5 rounded-xl border flex flex-col items-center justify-center text-center transition-all ${
                selectedWeapon === w.id
                  ? "bg-amber-500/20 border-amber-500/40 text-amber-400 shadow shadow-amber-500/10"
                  : "bg-slate-900/40 border-slate-900 text-slate-400 hover:text-slate-200 hover:bg-slate-900/80"
              }`}
            >
              <span className="text-base mb-1">
                {w.id === "sword" ? "⚔️" : w.id === "axe" ? "🪓" : w.id === "mace" ? "🔨" : w.id === "spear" ? "🔱" : w.id === "dagger" ? "🗡️" : w.id === "bow" ? "🏹" : w.id === "crossbow" ? "🏹" : w.id === "staff" ? "🔮" : w.id === "wand" ? "✨" : "📖"}
              </span>
              <span className="text-[10.5px] font-bold tracking-tight truncate w-full">{w.name}</span>
              <span className="text-[8px] font-mono text-slate-500 uppercase mt-0.5">{w.type}</span>
            </button>
          ))}
        </div>

        {/* Active Weapon Card */}
        {currentWeaponObj && (
          <div className="bg-slate-900/40 border border-slate-900 rounded-xl p-4 space-y-3 font-mono text-xs">
            <div className="flex items-center justify-between">
              <span className="text-[10px] uppercase font-bold text-amber-400 bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded">
                {currentWeaponObj.type} Weapon Profile
              </span>
              <span className="text-[10px] text-slate-500">Speed: <strong className="text-slate-300">{currentWeaponObj.speed}</strong></span>
            </div>
            <div>
              <h4 className="text-sm font-bold text-slate-100">{currentWeaponObj.name} Profile</h4>
              <p className="text-[11px] text-slate-400 mt-1 leading-relaxed">{currentWeaponObj.desc}</p>
            </div>
            <div className="border-t border-slate-900/80 pt-2 grid grid-cols-2 gap-2 text-[11px]">
              <div>
                <span className="text-slate-500 block text-[9px] uppercase font-bold">Layer 1: Weapon Type</span>
                <span className="text-amber-400 font-extrabold">{currentWeaponObj.weaponType}</span>
              </div>
              <div>
                <span className="text-slate-500 block text-[9px] uppercase font-bold">Layer 2: Damage Archetype</span>
                <span className="text-indigo-400 font-extrabold">{currentWeaponObj.damageArchetype}</span>
              </div>
            </div>
            <div className="border-t border-slate-900/80 pt-2 flex flex-wrap gap-2">
              <span className="text-[9px] text-slate-500">Compatible Classes:</span>
              {currentWeaponObj.classes.map((c) => (
                <span key={c} className="text-[9px] bg-slate-950 px-2 py-0.5 rounded text-slate-300 border border-slate-850">
                  {c}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="xl:col-span-6 bg-slate-950/80 border border-slate-850 rounded-2xl p-5 space-y-5 shadow-2xl relative">
        <div className="border-b border-slate-900 pb-3 flex items-center justify-between">
          <div>
            <span className="text-[10px] uppercase font-bold text-indigo-500 font-mono tracking-wider">Canonical Archetypes</span>
            <h3 className="text-base font-bold text-slate-200 flex items-center gap-2 mt-0.5">
              <Shield className="w-5 h-5 text-indigo-400" />
              <span>3 Armor Archetypes Bible</span>
            </h3>
          </div>
          <span className="text-[11px] font-mono text-slate-500">armor-archetype-rule</span>
        </div>

        <p className="text-xs text-slate-400 leading-relaxed font-mono">
          Armor is divided into archetypes that define defensive identity and playstyle specialization without breaking class balance. Select an archetype:
        </p>

        {/* Armor Selectors */}
        <div className="grid grid-cols-3 gap-3">
          {ARMOR_ARCHETYPES.map((a) => (
            <button
              key={a.id}
              onClick={() => setSelectedArmor(a.id)}
              className={`p-3 rounded-xl border flex flex-col items-center justify-center text-center transition-all ${
                selectedArmor === a.id
                  ? "bg-indigo-500/20 border-indigo-500/40 text-indigo-400 shadow shadow-indigo-500/10"
                  : "bg-slate-900/40 border-slate-900 text-slate-400 hover:text-slate-200 hover:bg-slate-900/80"
              }`}
            >
              <span className="text-xl mb-1">
                {a.id === "heavy" ? "🛡️" : a.id === "light" ? "🥋" : "🧙"}
              </span>
              <span className="text-xs font-bold tracking-tight">{a.name}</span>
            </button>
          ))}
        </div>

        {/* Active Armor Card */}
        {currentArmorObj && (
          <div className="bg-slate-900/40 border border-slate-900 rounded-xl p-4 space-y-3 font-mono text-xs">
            <div className="flex items-center justify-between">
              <span className="text-[10px] uppercase font-bold text-indigo-400 bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 rounded">
                {currentArmorObj.name} Profile
              </span>
              <span className="text-[10px] text-slate-500">Alignment: <strong className="text-slate-300">{currentArmorObj.classes.join(", ")}</strong></span>
            </div>
            <div>
              <p className="text-[11px] text-slate-400 leading-relaxed">{currentArmorObj.desc}</p>
            </div>
            <div className="border-t border-slate-900/80 pt-2 space-y-1">
              <span className="text-[9px] text-slate-500 block uppercase font-bold">Standard Stat Profile:</span>
              <span className="text-[11px] text-indigo-300 block font-semibold">{currentArmorObj.stats}</span>
            </div>
          </div>
        )}
      </div>

      {/* Row 2: Dual Wield Simulator & Affix Generator */}
      <div className="xl:col-span-6 bg-slate-950/80 border border-slate-850 rounded-2xl p-5 space-y-5 shadow-2xl relative">
        <div className="border-b border-slate-900 pb-3 flex items-center justify-between">
          <div>
            <span className="text-[10px] uppercase font-bold text-pink-500 font-mono tracking-wider">Class Mechanics</span>
            <h3 className="text-base font-bold text-slate-200 flex items-center gap-2 mt-0.5">
              <ArrowLeftRight className="w-5 h-5 text-pink-400" />
              <span>Dual Wield Bible & Calculator</span>
            </h3>
          </div>
          <span className="text-[11px] font-mono text-slate-500">dual-wield-rule</span>
        </div>

        <div className="bg-pink-500/5 rounded-xl border border-pink-500/10 p-3 text-[11px] leading-relaxed text-pink-300 font-mono">
          Dual wield is an exclusive class mechanic only available to explicitly authorized classes and weapon combinations:
          <strong className="block text-slate-200 mt-1">Knight: Sword + Shield OR Dual Sword</strong>
          <strong className="block text-slate-200">Assassin: Dagger + Offhand OR Dual Dagger</strong>
        </div>

        {/* Configurations Selector */}
        <div className="grid grid-cols-2 gap-3 font-mono">
          <div>
            <label className="text-[9px] uppercase font-bold text-slate-500 block mb-1">Select Class Profile:</label>
            <div className="flex gap-2">
              <button
                onClick={() => { setDwClass("Knight"); setDwMode("single"); }}
                className={`flex-1 p-2 rounded-lg border text-xs font-bold transition-all ${
                  dwClass === "Knight" ? "bg-pink-500/20 text-pink-400 border-pink-500/40" : "bg-slate-900/50 border-slate-900 text-slate-500 hover:text-slate-300"
                }`}
              >
                🛡️ Knight
              </button>
              <button
                onClick={() => { setDwClass("Assassin"); setDwMode("single"); }}
                className={`flex-1 p-2 rounded-lg border text-xs font-bold transition-all ${
                  dwClass === "Assassin" ? "bg-pink-500/20 text-pink-400 border-pink-500/40" : "bg-slate-900/50 border-slate-900 text-slate-500 hover:text-slate-300"
                }`}
              >
                🗡️ Assassin
              </button>
            </div>
          </div>

          <div>
            <label className="text-[9px] uppercase font-bold text-slate-500 block mb-1">Equipped Configuration:</label>
            <div className="flex gap-2">
              <button
                onClick={() => setDwMode("single")}
                className={`flex-1 p-2 rounded-lg border text-xs font-bold transition-all ${
                  dwMode === "single" ? "bg-pink-500/20 text-pink-400 border-pink-500/40" : "bg-slate-900/50 border-slate-900 text-slate-500"
                }`}
              >
                {dwClass === "Knight" ? "Sword + Shield" : "Dagger + Offhand"}
              </button>
              <button
                onClick={() => setDwMode("dual")}
                className={`flex-1 p-2 rounded-lg border text-xs font-bold transition-all ${
                  dwMode === "dual" ? "bg-pink-500/20 text-pink-400 border-pink-500/40" : "bg-slate-900/50 border-slate-900 text-slate-500"
                }`}
              >
                {dwClass === "Knight" ? "Dual Sword" : "Dual Dagger"}
              </button>
            </div>
          </div>
        </div>

        {/* Sliders for weapon powers */}
        <div className="space-y-3 font-mono text-xs border-t border-slate-900 pt-3">
          <div className="space-y-1">
            <div className="flex justify-between text-[10px] uppercase font-bold text-slate-400">
              <span>Main Hand Weapon Power:</span>
              <span className="text-pink-400 font-bold">{mainHandPower}</span>
            </div>
            <input
              type="range"
              min="20"
              max="200"
              value={mainHandPower}
              onChange={(e) => setMainHandPower(parseInt(e.target.value))}
              className="w-full accent-pink-500 bg-slate-900"
            />
          </div>

          {dwMode === "dual" && (
            <div className="space-y-1">
              <div className="flex justify-between text-[10px] uppercase font-bold text-slate-400">
                <span>Off Hand Weapon Power (75% Scaling):</span>
                <span className="text-pink-400 font-bold">{offHandPower} (Effective: {Math.round(offHandPower * 0.75)})</span>
              </div>
              <input
                type="range"
                min="20"
                max="200"
                value={offHandPower}
                onChange={(e) => setOffHandPower(parseInt(e.target.value))}
                className="w-full accent-pink-500 bg-slate-900"
              />
            </div>
          )}
        </div>

        {/* Calculation Result */}
        <div className="bg-slate-900/60 border border-slate-900 rounded-xl p-4 font-mono space-y-3">
          <div className="flex justify-between items-center">
            <span className="text-[10px] uppercase font-bold text-slate-500">Damage Formula:</span>
            <span className="text-[10px] text-pink-400 bg-pink-500/10 border border-pink-500/20 px-2 py-0.5 rounded font-bold">
              dual-wield-final-rule
            </span>
          </div>
          <div className="text-xs text-slate-300 leading-relaxed">
            Main hand contributes <strong className="text-white">100%</strong> of weapon power. Off hand contributes exactly <strong className="text-white">75%</strong>.
            <div className="text-amber-400 font-bold text-sm mt-1.5 bg-slate-950 p-2.5 rounded border border-slate-850/60 text-center">
              Effective Weapon Power = {mainHandPower} + ({dwMode === "dual" ? `${offHandPower} × 0.75` : "0"}) = <span className="text-white text-lg">{dwEffectivePower.toFixed(1)}</span>
            </div>
          </div>

          {/* Tradeoff Details */}
          <div className="border-t border-slate-950 pt-2.5 space-y-1.5 text-[11px]">
            <span className="text-[9px] uppercase font-bold text-slate-500 block">Class Setup Tradeoffs:</span>
            {dwClass === "Knight" && dwMode === "dual" && (
              <div className="text-rose-400">⚔️ Knight Dual Sword: Higher burst & DPS, but loses Shield defense / lower survivability.</div>
            )}
            {dwClass === "Knight" && dwMode === "single" && (
              <div className="text-emerald-400">🛡️ Knight Sword + Shield: High armor block, high survivability, moderate DPS.</div>
            )}
            {dwClass === "Assassin" && dwMode === "dual" && (
              <div className="text-rose-400">💥 Assassin Dual Dagger: Faster burst combos & crit frequency, but lower utility / extremely fragile.</div>
            )}
            {dwClass === "Assassin" && dwMode === "single" && (
              <div className="text-emerald-400">🗡️ Assassin Single + Offhand: Better utility scaling, safer pacing, moderate burst.</div>
            )}
          </div>
        </div>
      </div>

      <div className="xl:col-span-6 bg-slate-950/80 border border-slate-850 rounded-2xl p-5 space-y-5 shadow-2xl relative">
        <div className="border-b border-slate-900 pb-3 flex items-center justify-between">
          <div>
            <span className="text-[10px] uppercase font-bold text-emerald-500 font-mono tracking-wider">Item Affixes</span>
            <h3 className="text-base font-bold text-slate-200 flex items-center gap-2 mt-0.5">
              <Sparkles className="w-5 h-5 text-emerald-400" />
              <span>Affix Pool Generator</span>
            </h3>
          </div>
          <span className="text-[11px] font-mono text-slate-500">affix-pool-rule</span>
        </div>

        <p className="text-xs text-slate-400 leading-relaxed font-mono">
          Roll attributes from the global canonical affix pool. Item tier determines the exact count and magnitude of affixes (<code className="text-slate-300">tier-affix-scaling-rule</code>).
        </p>

        {/* Tier Selector & Roll Button */}
        <div className="grid grid-cols-6 gap-2 items-center font-mono">
          <div className="col-span-4 flex items-center gap-2 bg-slate-900/80 border border-slate-800 p-1.5 rounded-lg">
            <span className="text-[9px] text-slate-500 uppercase font-bold pl-1.5">Tier:</span>
            {[1, 2, 3, 4, 5].map((t) => (
              <button
                key={t}
                onClick={() => setSelectedTier(t as any)}
                className={`flex-1 py-1 rounded text-xs font-extrabold transition-all ${
                  selectedTier === t
                    ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                T{t}
              </button>
            ))}
          </div>
          <button
            onClick={handleRollAffixes}
            className="col-span-2 bg-gradient-to-r from-emerald-500 to-teal-600 text-slate-950 font-bold text-xs py-2.5 px-3 rounded-lg hover:brightness-110 active:scale-[0.98] transition-all flex items-center justify-center gap-1 shadow shadow-emerald-500/10 font-mono"
          >
            <RefreshCw className="w-3.5 h-3.5 animate-spin-slow" />
            <span>Roll Item</span>
          </button>
        </div>

        {/* Rolled Item Preview Card */}
        <div className="bg-slate-900/60 border border-slate-900 rounded-xl p-4 space-y-3 font-mono">
          <div className="flex justify-between items-center border-b border-slate-950 pb-2">
            <div>
              <span className="text-[10px] text-slate-500 uppercase font-bold">Type: {rolledItemType} | Archetype: {rolledItemArchetype} • Tier {selectedTier}</span>
              <h4 className="text-sm font-extrabold text-slate-100 mt-0.5">{rolledItemName}</h4>
            </div>
            <span className={`text-[10px] font-mono px-2 py-0.5 rounded font-extrabold uppercase ${
              selectedTier === 5 ? "bg-amber-500/20 text-amber-300 border border-amber-500/30" :
              selectedTier === 4 ? "bg-violet-500/20 text-violet-300 border border-violet-500/30" :
              selectedTier === 3 ? "bg-blue-500/20 text-blue-300 border border-blue-500/30" :
              selectedTier === 2 ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30" :
              "bg-slate-800 text-slate-400 border border-slate-750"
            }`}>
              {selectedTier === 5 ? "Artifact" : `T${selectedTier}`}
            </span>
          </div>

          {/* Affixes List */}
          <div className="space-y-1.5 py-1">
            {rolledAffixes.map((aff, idx) => (
              <div key={idx} className="flex justify-between items-center text-xs p-1.5 bg-slate-950/40 rounded border border-slate-900/80">
                <div className="flex items-center gap-1.5">
                  <span className={`text-[8.5px] font-bold px-1.5 py-0.2 rounded font-mono uppercase tracking-wider ${
                    aff.category.includes("ELEMENTAL OFFENSE") ? "bg-orange-500/20 text-orange-300 border border-orange-500/20" :
                    aff.category.includes("ELEMENTAL DEFENSE") ? "bg-teal-500/20 text-teal-300 border border-teal-500/20" :
                    aff.category.includes("OFFENSIVE") ? "bg-rose-500/20 text-rose-300 border border-rose-500/20" :
                    aff.category.includes("DEFENSIVE") ? "bg-sky-500/20 text-sky-300 border border-sky-500/20" :
                    "bg-violet-500/20 text-violet-300 border border-violet-500/20"
                  }`}>
                    {aff.category}
                  </span>
                  <span className="text-slate-300 font-semibold">{aff.name}</span>
                </div>
                <span className="text-emerald-400 font-bold font-mono">{aff.value}</span>
              </div>
            ))}
          </div>

          {/* Tier Validator Box */}
          <div className="border-t border-slate-950 pt-2.5 flex items-center justify-between text-[11px]">
            <span className="text-slate-500">Tier Rule Validation:</span>
            <div className="flex items-center gap-1.5 text-emerald-400 font-bold">
              <CheckCircle className="w-4 h-4" />
              <span>{getTierAffixRange(selectedTier).desc} Confirmed</span>
            </div>
          </div>
        </div>

        {/* Roll History */}
        <div className="bg-slate-950/40 border border-slate-900 rounded-lg p-2.5 font-mono text-[9px] text-slate-500 space-y-1">
          <span className="uppercase font-bold tracking-wider block text-slate-600">Loot Roll Engine Logs:</span>
          {rollHistory.length === 0 ? (
            <span className="italic">No items rolled yet. Click Roll Item above.</span>
          ) : (
            rollHistory.map((h, i) => (
              <div key={i} className="flex items-center gap-1">
                <span className="text-emerald-500">▶</span>
                <span>{h}</span>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Row 3: Combat Formula & Defense Simulator */}
      <div className="xl:col-span-12 bg-slate-900/60 border border-slate-800/60 p-6 rounded-2xl backdrop-blur-sm shadow-xl space-y-6">
        <div>
          <span className="bg-amber-500/10 text-amber-400 text-[10px] font-bold font-mono px-2.5 py-1 rounded border border-amber-500/20 uppercase tracking-wider">
            FÓRMULA DE COMBATE & DEFESA • CANONIZADO
          </span>
          <h3 className="text-xl font-extrabold text-slate-100 mt-2 flex items-center gap-2">
            <Sliders className="w-5 h-5 text-amber-400" />
            <span>Simulador Canônico de Combate e TTK</span>
          </h3>
          <p className="text-xs text-slate-400 leading-relaxed mt-1">
            Simule o fluxo de dano completo aplicando as regras autoritativas de mitigação de armadura física, amplificação elemental e modificadores de acerto crítico de classe.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          
          {/* Column 1: Configuração do Ataque e Defesa (Inputs) */}
          <div className="lg:col-span-7 space-y-6">
            
            {/* Bloco 1: Offensive Calculator */}
            <div className="bg-slate-950/80 border border-slate-800/80 p-5 rounded-xl space-y-4">
              <div className="flex items-center gap-2 text-amber-400 font-bold text-xs uppercase tracking-wider border-b border-slate-900 pb-2">
                <Sword className="w-4 h-4" />
                <span>1. Calculadora Ofensiva (Atacante)</span>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 font-mono text-xs text-slate-300">
                
                <div>
                  <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1">Classe do Atacante:</label>
                  <select
                    value={combClass}
                    onChange={(e) => setCombClass(e.target.value as any)}
                    className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-xs text-slate-200 outline-none focus:border-amber-500/40"
                  >
                    <option value="Knight">🛡️ Knight</option>
                    <option value="Assassin">🗡️ Assassin</option>
                    <option value="Archer">🏹 Archer</option>
                    <option value="Mage">🔮 Mage</option>
                    <option value="Cleric">✨ Cleric</option>
                  </select>
                </div>

                <div>
                  <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1">Arma Equipada:</label>
                  <select
                    value={combWeapon}
                    onChange={(e) => setCombWeapon(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-xs text-slate-200 outline-none focus:border-amber-500/40"
                  >
                    {CANONICAL_WEAPONS.filter(w => w.classes.includes(combClass)).map(w => (
                      <option key={w.id} value={w.id}>{w.name} (Dmg: {w.baseDamage})</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1">Habilidade / Spell:</label>
                  <select
                    value={combSpell}
                    onChange={(e) => setCombSpell(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1.5 text-xs text-slate-200 outline-none focus:border-amber-500/40"
                  >
                    {CANONICAL_SPELLS.filter(s => s.classId === combClass).map(s => (
                      <option key={s.id} value={s.id}>{s.name} (Pwr: {s.basePower} | {s.elementalType})</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1.5">Mecânicas Especiais:</label>
                  <div className="flex flex-col gap-2">
                    <label className="flex items-center gap-2 cursor-pointer text-[11px] text-slate-400 hover:text-slate-200">
                      <input
                        type="checkbox"
                        checked={combDualWield}
                        disabled={combClass !== "Knight" && combClass !== "Assassin"}
                        onChange={(e) => setCombDualWield(e.target.checked)}
                        className="rounded border-slate-800 text-amber-500 focus:ring-0 focus:ring-offset-0 bg-slate-900"
                      />
                      <span>Dual Wielding (Regra 75%)</span>
                    </label>
                    <label className="flex items-center gap-2 cursor-pointer text-[11px] text-slate-400 hover:text-slate-200">
                      <input
                        type="checkbox"
                        checked={combCritForced}
                        onChange={(e) => setCombCritForced(e.target.checked)}
                        className="rounded border-slate-800 text-amber-500 focus:ring-0 focus:ring-offset-0 bg-slate-900"
                      />
                      <span>Forçar Golpe Crítico</span>
                    </label>
                  </div>
                </div>

              </div>

              {/* Slider Afinidade Elemental do Atacante */}
              {(() => {
                const spell = CANONICAL_SPELLS.find(s => s.id === combSpell);
                if (spell && spell.elementalType !== "Physical") {
                  return (
                    <div className="space-y-1 pt-2 border-t border-slate-900 font-mono text-xs">
                      <div className="flex justify-between text-[10px] uppercase font-bold text-slate-400">
                        <span>Afinidade Elemental Ativa do Atacante ({spell.elementalType}):</span>
                        <span className="text-amber-400 font-bold">+{combAttackerAffinity}%</span>
                      </div>
                      <input
                        type="range"
                        min="0"
                        max="50"
                        value={combAttackerAffinity}
                        onChange={(e) => setCombAttackerAffinity(parseInt(e.target.value))}
                        className="w-full accent-amber-500 bg-slate-900 h-1.5 rounded"
                      />
                    </div>
                  );
                }
                return null;
              })()}
            </div>

            {/* Bloco 2: Defense Calculator */}
            <div className="bg-slate-950/80 border border-slate-800/80 p-5 rounded-xl space-y-4">
              <div className="flex items-center gap-2 text-emerald-400 font-bold text-xs uppercase tracking-wider border-b border-slate-900 pb-2">
                <Shield className="w-4 h-4" />
                <span>2. Calculadora de Defesa (Defensor)</span>
              </div>

              <div className="space-y-4 font-mono text-xs">
                
                {/* Armadura Física */}
                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] uppercase font-bold text-slate-400">
                    <span>Armadura Física Total:</span>
                    <span className="text-emerald-400 font-bold">{combArmor} Def (Mitigação: {((combArmor / (combArmor + 250)) * 100).toFixed(2)}%)</span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="1500"
                    step="5"
                    value={combArmor}
                    onChange={(e) => setCombArmor(parseInt(e.target.value))}
                    className="w-full accent-emerald-500 bg-slate-900 h-1.5 rounded"
                  />
                </div>

                {/* Benchmark Presets & High-TTK Validation */}
                <div className="space-y-2 pt-2 border-t border-slate-900">
                  <span className="text-[10px] uppercase font-bold text-slate-500 block tracking-wider">Benchmarks de Mitigação Canônica (K=250):</span>
                  <div className="grid grid-cols-3 gap-2">
                    <button
                      type="button"
                      onClick={() => setCombArmor(75)}
                      className={`p-2 rounded text-left border transition-all cursor-pointer flex flex-col gap-0.5 ${
                        combArmor === 75
                          ? "bg-sky-500/10 border-sky-500 text-sky-200 animate-pulse"
                          : "bg-slate-900/60 border-slate-800/80 text-slate-400 hover:border-slate-700 hover:text-slate-300"
                      }`}
                    >
                      <span className="font-extrabold text-[10px] uppercase tracking-wider block">🍃 Light Build</span>
                      <span className="font-mono text-[9px] block">75 Def (~23.1%)</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setCombArmor(200)}
                      className={`p-2 rounded text-left border transition-all cursor-pointer flex flex-col gap-0.5 ${
                        combArmor === 200
                          ? "bg-amber-500/10 border-amber-500 text-amber-200 animate-pulse"
                          : "bg-slate-900/60 border-slate-800/80 text-slate-400 hover:border-slate-700 hover:text-slate-300"
                      }`}
                    >
                      <span className="font-extrabold text-[10px] uppercase tracking-wider block">⚖️ Hybrid Build</span>
                      <span className="font-mono text-[9px] block">200 Def (~44.4%)</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setCombArmor(500)}
                      className={`p-2 rounded text-left border transition-all cursor-pointer flex flex-col gap-0.5 ${
                        combArmor === 500
                          ? "bg-rose-500/10 border-rose-500 text-rose-200 animate-pulse"
                          : "bg-slate-900/60 border-slate-800/80 text-slate-400 hover:border-slate-700 hover:text-slate-300"
                      }`}
                    >
                      <span className="font-extrabold text-[10px] uppercase tracking-wider block">🛡️ Tank Build</span>
                      <span className="font-mono text-[9px] block">500 Def (~66.7%)</span>
                    </button>
                  </div>

                  {/* Benchmark Validation references */}
                  <div className="bg-slate-900/30 border border-slate-800/40 p-2.5 rounded text-[10px] text-slate-400 space-y-1">
                    <span className="font-bold text-slate-300 block uppercase tracking-wide text-[9px]">Gabarito de Validação (Fórmula Oficial):</span>
                    <div className="grid grid-cols-2 gap-x-4 gap-y-1 font-mono text-[9.5px]">
                      <div className="flex justify-between border-b border-slate-900 pb-0.5">
                        <span>• 100 Armor:</span>
                        <span className="text-emerald-400 font-bold">28.57%</span>
                      </div>
                      <div className="flex justify-between border-b border-slate-900 pb-0.5">
                        <span>• 250 Armor:</span>
                        <span className="text-emerald-400 font-bold">50.00%</span>
                      </div>
                      <div className="flex justify-between border-b border-slate-900 pb-0.5">
                        <span>• 500 Armor:</span>
                        <span className="text-emerald-400 font-bold text-rose-400">66.67% (Tank Anchor)</span>
                      </div>
                      <div className="flex justify-between border-b border-slate-900 pb-0.5">
                        <span>• 700 Armor:</span>
                        <span className="text-emerald-400 font-bold">73.68%</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Resistências Elementais */}
                {(() => {
                  const spell = CANONICAL_SPELLS.find(s => s.id === combSpell);
                  if (spell && spell.elementalType !== "Physical") {
                    const ele = spell.elementalType;
                    const resValue = ele === "Fire" ? combFireRes : ele === "Ice" ? combIceRes : ele === "Holy" ? combHolyRes : ele === "Shadow" ? combShadowRes : combNatureRes;
                    const setResValue = ele === "Fire" ? setCombFireRes : ele === "Ice" ? setCombIceRes : ele === "Holy" ? setCombHolyRes : ele === "Shadow" ? setCombShadowRes : setCombNatureRes;
                    return (
                      <div className="space-y-3 pt-2 border-t border-slate-900">
                        <div className="space-y-1">
                          <div className="flex justify-between text-[10px] uppercase font-bold text-slate-400">
                            <span>Resistência Elemental do Defensor a {ele}:</span>
                            <span className="text-emerald-400 font-bold">{resValue}%</span>
                          </div>
                          <input
                            type="range"
                            min="0"
                            max="75"
                            value={resValue}
                            onChange={(e) => setResValue(parseInt(e.target.value))}
                            className="w-full accent-emerald-500 bg-slate-900 h-1.5 rounded"
                          />
                        </div>

                        {/* Affinity Defense Bonus Toggle */}
                        <div className="flex items-center justify-between bg-slate-900/60 p-2.5 rounded border border-slate-800">
                          <div className="space-y-0.5">
                            <span className="text-[10px] uppercase font-bold text-slate-300 block">Affinity Defense (Nvl 100 Awakened)</span>
                            <span className="text-[9px] text-slate-500 block">Ativa +4% de resistência bônus contra o elemento de afinidade matching.</span>
                          </div>
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input
                              type="checkbox"
                              checked={combAffinityDefense}
                              onChange={(e) => setCombAffinityDefense(e.target.checked)}
                              className="sr-only peer"
                            />
                            <div className="w-9 h-5 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-slate-400 after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-600 peer-checked:after:bg-slate-100"></div>
                          </label>
                        </div>
                      </div>
                    );
                  }
                  return (
                    <div className="p-3 bg-slate-900/40 border border-slate-850 rounded text-center text-slate-500 text-[11px] italic">
                      Habilidade ativa causa dano Físico Puro. Resistências elementais e bônus de afinidade defensiva estão dormentes.
                    </div>
                  );
                })()}

              </div>
            </div>

          </div>

          {/* Column 2: Live Calculations & Real-Time Logs */}
          <div className="lg:col-span-5 flex flex-col gap-4">
            
            {/* Bloco 3: Critical Simulator Indicators & Direct Damage Indicator */}
            {(() => {
              const weapon = CANONICAL_WEAPONS.find(w => w.id === combWeapon) || CANONICAL_WEAPONS[0];
              const totalCrit = 5 + (weapon.critBonus * 100);
              const critMult = combClass === "Assassin" ? 2.2 : 1.5;
              return (
                <div className="bg-slate-950/80 border border-slate-800/80 p-5 rounded-xl space-y-4 shadow-xl">
                  <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                    <span className="text-[10px] uppercase font-bold text-violet-400 font-mono tracking-wider">Métricas e Probabilidades</span>
                    <span className="text-[10px] text-slate-500 font-mono">critical-hit-rule</span>
                  </div>

                  <div className="grid grid-cols-2 gap-3 text-center font-mono">
                    <div className="bg-slate-900 p-2.5 rounded border border-slate-800/60">
                      <span className="text-[9px] text-slate-500 uppercase font-bold block">Crit Chance</span>
                      <span className="text-base font-black text-violet-400">{totalCrit}%</span>
                      <span className="text-[8px] text-slate-600 block mt-0.5">5% Base + {(weapon.critBonus*100).toFixed(0)}% Weapon</span>
                    </div>

                    <div className="bg-slate-900 p-2.5 rounded border border-slate-800/60">
                      <span className="text-[9px] text-slate-500 uppercase font-bold block">Crit Multiplier</span>
                      <span className="text-base font-black text-violet-400">{critMult}x</span>
                      <span className="text-[8px] text-slate-600 block mt-0.5">{combClass === "Assassin" ? "Assassin Override" : "Normal Vocation"}</span>
                    </div>
                  </div>

                  <div className="bg-gradient-to-br from-slate-950 to-slate-900 border border-slate-850 p-4 rounded-xl flex flex-col gap-1 text-center shadow-inner relative overflow-hidden group">
                    <div className="absolute top-0 right-0 w-16 h-16 bg-amber-500/5 rounded-full blur-xl pointer-events-none" />
                    <span className="text-[10px] uppercase font-bold text-amber-500 font-mono tracking-wider">Dano de Impacto Esperado</span>
                    {(() => {
                      const finalLog = combatLogs[combatLogs.length - 2];
                      const valText = finalLog ? finalLog.text.split("= ")[1]?.split(" HP")[0] || "0.0" : "0.0";
                      return (
                        <div className="flex items-baseline justify-center gap-1.5 my-1">
                          <span className="text-3xl font-black font-sans tracking-tight text-white">{valText}</span>
                          <span className="text-xs font-bold text-slate-500 font-mono">HP</span>
                        </div>
                      );
                    })()}
                    <button
                      onClick={simulateHit}
                      className="w-full mt-2 py-2 bg-gradient-to-r from-amber-500 to-amber-600 hover:brightness-110 active:scale-[0.98] text-slate-950 font-black text-xs uppercase rounded transition-all shadow-md shadow-amber-950/30 flex items-center justify-center gap-2 font-mono"
                    >
                      <Sparkles className="w-4 h-4 animate-pulse" />
                      <span>Desferir Golpe Simulador</span>
                    </button>
                  </div>
                </div>
              );
            })()}

            {/* Bloco 4: Real-Time Combat Logs Terminal */}
            <div className="bg-slate-950/80 border border-slate-800/80 p-5 rounded-xl flex flex-col gap-3 flex-1 min-h-[300px]">
              <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                <div className="flex items-center gap-2">
                  <Terminal className="w-4 h-4 text-amber-400" />
                  <span className="text-xs font-bold text-slate-200 uppercase tracking-wider font-mono">Log Matemático da Transação</span>
                </div>
                <span className="text-[9px] font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.2 rounded uppercase">
                  Real-time
                </span>
              </div>

              {/* Terminal Logs View */}
              <div className="bg-slate-950 border border-slate-900 rounded-lg p-3 font-mono text-[10.5px] leading-relaxed flex-1 overflow-y-auto max-h-[380px] space-y-2 select-text">
                {combatLogs.map((log, idx) => {
                  let textCol = "text-slate-400";
                  let bgCol = "bg-transparent";
                  if (log.type === "success") {
                    textCol = "text-emerald-400 font-bold";
                    bgCol = "bg-emerald-500/5 border border-emerald-500/10 rounded px-1.5 py-0.5";
                  } else if (log.type === "warning") {
                    textCol = "text-amber-400 font-bold";
                    bgCol = "bg-amber-500/5 border border-amber-500/10 rounded px-1.5 py-0.5";
                  } else if (log.type === "formula") {
                    textCol = "text-indigo-300 font-semibold italic";
                    bgCol = "bg-indigo-500/5 border border-indigo-500/10 rounded px-1.5 py-0.5";
                  }
                  return (
                    <div key={idx} className={`${bgCol} ${textCol}`}>
                      {log.text}
                    </div>
                  );
                })}
              </div>
            </div>

          </div>

        </div>
      </div>

    </div>
  );
}


// ==========================================
// PROGRESSION BIBLE SIMULATOR COMPONENT
// ==========================================
function ProgressionSimulator() {
  const [skills, setSkills] = useState<Record<string, { level: number; xp: number }>>({
    sword_fighting: { level: 120, xp: 4500 },
    axe_fighting: { level: 35, xp: 120 },
    club_fighting: { level: 20, xp: 800 },
    dagger_fighting: { level: 80, xp: 1500 },
    distance_fighting: { level: 15, xp: 50 },
    magic_level: { level: 42, xp: 900 },
    shielding: { level: 75, xp: 3200 }
  });

  const [activeSkill, setActiveSkill] = useState<string>("sword_fighting");
  const [mobDifficulty, setMobDifficulty] = useState<"low" | "standard" | "elite" | "boss" | "pvp">("standard");
  const [selectedMilestone, setSelectedMilestone] = useState<string>("none");
  const [combatFrequency, setCombatFrequency] = useState<number>(2.5); // actions per second
  const [combatLog, setCombatLog] = useState<Array<{ id: string; text: string; type: "train" | "level" | "info" | "warning" }>>([
    { id: "init-1", text: "⚙️ [Progression] Inicializando Banco de Dados de Habilidades Canônicas (shielding-skill-rule).", type: "info" },
    { id: "init-2", text: "🛡️ [Shielding] Shielding é agora uma proficiência de combate completa e segue as mesmas regras de XP e prestígio.", type: "info" },
    { id: "init-3", text: "⚖️ [Progression] Os atributos primários (Força, Destreza, Inteligência, Vitalidade, Agilidade e Sorte) foram totalmente desativados (attribute-less-progression-rule).", type: "warning" }
  ]);
  const [autoTrain, setAutoTrain] = useState<boolean>(false);

  // Combat Calculator states
  const [selectedWeapon, setSelectedWeapon] = useState<string>("sword_t3");
  const [weaponBaseDmg, setWeaponBaseDmg] = useState<number>(58);
  const [isCritForced, setIsCritForced] = useState<boolean>(false);
  const [isAssassinClass, setIsAssassinClass] = useState<boolean>(false);
  const [elementalMultiplier, setElementalMultiplier] = useState<number>(1.25);
  const [defenderArmor, setDefenderArmor] = useState<number>(250);

  const skillNames: Record<string, string> = {
    sword_fighting: "Sword Fighting",
    axe_fighting: "Axe Fighting",
    club_fighting: "Club Fighting",
    dagger_fighting: "Dagger Fighting",
    distance_fighting: "Distance Fighting",
    magic_level: "Magic Level",
    shielding: "Shielding"
  };

  const getXpForNextLevel = (lvl: number) => {
    return Math.floor(100 * Math.pow(1.08, lvl));
  };

  const getEffectiveCombatSkill = (skillLevel: number): number => {
    if (skillLevel <= 150) {
      return skillLevel;
    } else if (skillLevel <= 250) {
      return 150 + (skillLevel - 150) * 0.75;
    } else {
      return 225 + (skillLevel - 250) * 0.50;
    }
  };

  const getPrestigeBandDetails = (skillLevel: number) => {
    if (skillLevel <= 150) {
      return {
        band: "Band 1 (1–150)",
        mult: "100% Eficácia",
        desc: "Escalonamento linear total sem perdas.",
        color: "text-emerald-400 border-emerald-500/20 bg-emerald-500/5"
      };
    } else if (skillLevel <= 250) {
      return {
        band: "Band 2 (151–250)",
        mult: "75% Eficácia",
        desc: "Curva de amortecimento médio para controle de poder.",
        color: "text-amber-400 border-amber-500/20 bg-amber-500/5"
      };
    } else {
      return {
        band: "Band 3 (251+)",
        mult: "50% Eficácia",
        desc: "Curva de amortecimento alto. Progressão infinita com segurança de sandbox.",
        color: "text-rose-400 border-rose-500/20 bg-rose-500/5"
      };
    }
  };

  // Combat Math Calculations
  const currentSkillState = skills[activeSkill];
  const effectiveSkillVal = getEffectiveCombatSkill(currentSkillState.level);
  const prestigeInfo = getPrestigeBandDetails(currentSkillState.level);

  const rawBaseDamage = weaponBaseDmg + effectiveSkillVal * 0.5;
  const criticalMultiplier = isAssassinClass ? 2.2 : 1.5;
  const currentCritDmg = isCritForced ? rawBaseDamage * criticalMultiplier : rawBaseDamage;
  const elementalDamage = currentCritDmg * elementalMultiplier;

  // Armor Mitigation (K=250 Formula)
  const K = 250;
  const armorMitigationPercent = defenderArmor / (defenderArmor + K);
  
  // Real liquid damage calculations
  const rawDamageValue = elementalDamage;
  const mitigatedDamageValue = elementalDamage * armorMitigationPercent;
  const liquidDamageValue = elementalDamage * (1 - armorMitigationPercent);

  // 3-Layer XP Model
  // Layer 1: Action XP (Guaranteed Baseline)
  const actionXpValue = activeSkill === "shielding" ? 15 : 12;

  // Layer 2: Damage Contribution XP (primarily scales with liquid damage)
  const damageXpValue = Number((liquidDamageValue * 0.25 + mitigatedDamageValue * 0.10).toFixed(1));

  // Layer 3: Combat Outcome Bonus XP (Significant milestones)
  let outcomeBonusXpValue = 0;
  if (selectedMilestone === "elite_kill") outcomeBonusXpValue = 50;
  else if (selectedMilestone === "miniboss_kill") outcomeBonusXpValue = 150;
  else if (selectedMilestone === "boss_kill") outcomeBonusXpValue = 500;
  else if (selectedMilestone === "pvp_kill") outcomeBonusXpValue = 300;

  const totalXpBeforePenalty = actionXpValue + damageXpValue + outcomeBonusXpValue;

  // Anti-Macro Penalty
  const isAntiMacroTriggered = mobDifficulty === "low";
  const penaltyRate = 0.93; // 93% penalty
  const penaltyAmountValue = isAntiMacroTriggered ? totalXpBeforePenalty * penaltyRate : 0;
  const finalXpValue = Math.max(1, Math.round(isAntiMacroTriggered ? totalXpBeforePenalty * (1 - penaltyRate) : totalXpBeforePenalty));

  // Auto-train loop
  useEffect(() => {
    let timer: any;
    if (autoTrain) {
      timer = setInterval(() => {
        handleTrainClick();
      }, Math.max(200, 1000 / combatFrequency));
    }
    return () => clearInterval(timer);
  }, [autoTrain, activeSkill, mobDifficulty, combatFrequency, finalXpValue, selectedMilestone]);

  const handleTrainClick = () => {
    setSkills(prev => {
      const current = prev[activeSkill];
      const xpGain = finalXpValue;

      let newXp = current.xp + xpGain;
      let newLevel = current.level;
      let leveledUp = false;

      while (newXp >= getXpForNextLevel(newLevel)) {
        newXp -= getXpForNextLevel(newLevel);
        newLevel += 1;
        leveledUp = true;
      }

      const skillName = skillNames[activeSkill];
      let logText = "";
      if (leveledUp) {
        logText = `⭐ LEVEL UP! Seu '${skillName}' subiu para o nível ${newLevel}! (${newXp}/${getXpForNextLevel(newLevel)} XP)`;
      } else {
        if (activeSkill === "shielding") {
          logText = `🛡️ [Shielding Block] Bloqueou ataque! Ganhou +${xpGain} XP (+${actionXpValue} Action, +${damageXpValue} Damage, +${outcomeBonusXpValue} Outcome). [Progresso: ${newXp}/${getXpForNextLevel(newLevel)} XP]`;
        } else {
          logText = `⚔️ [Combat Action] Executou ação de combate! Ganhou +${xpGain} XP (+${actionXpValue} Action, +${damageXpValue} Damage, +${outcomeBonusXpValue} Outcome). [Progresso: ${newXp}/${getXpForNextLevel(newLevel)} XP]`;
        }
        if (isAntiMacroTriggered) {
          logText += ` ⚠️ (anti-macro-hard-penalty-rule: -93% de XP aplicado)`;
        }
      }

      setCombatLog(old => [
        {
          id: `${Date.now()}-${Math.random()}`,
          text: logText,
          type: leveledUp ? "level" : isAntiMacroTriggered ? "warning" : "train"
        },
        ...old.slice(0, 39)
      ]);

      return {
        ...prev,
        [activeSkill]: { level: newLevel, xp: newXp }
      };
    });
  };

  // Quick level editor
  const adjustSkillLevel = (skillId: string, delta: number) => {
    setSkills(prev => {
      const current = prev[skillId];
      const nextLevel = Math.max(1, current.level + delta);
      return {
        ...prev,
        [skillId]: { level: nextLevel, xp: 0 }
      };
    });
    setCombatLog(old => [
      {
        id: `${Date.now()}-${Math.random()}`,
        text: `✏️ [Ajuste Direto] Ajustado ${skillNames[skillId]} para o nível ${Math.max(1, skills[skillId].level + delta)}. (no-hard-skill-cap-rule)`,
        type: "info"
      },
      ...old.slice(0, 39)
    ]);
  };

  return (
    <div className="lg:col-span-12 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl space-y-6 text-slate-100">
      
      {/* Title Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center border-b border-slate-800/80 pb-4 gap-4">
        <div>
          <div className="flex items-center gap-2">
            <Sliders className="w-5 h-5 text-amber-400" />
            <h3 className="font-extrabold text-lg tracking-tight bg-gradient-to-r from-amber-400 to-violet-400 bg-clip-text text-transparent">
              Progression Bible & Combat Simulator (Final Canon)
            </h3>
          </div>
          <p className="text-xs text-slate-400 mt-1">
            Simulador e validador em tempo real do sistema de progressão híbrido de <code className="text-amber-300">Light and Shadow</code>.
          </p>
        </div>
        <div className="flex flex-wrap gap-2 text-[10px] font-mono bg-slate-950 p-2 rounded-lg border border-slate-800/80">
          <span className="text-emerald-400">• Zero Attributes</span>
          <span className="text-blue-400">• K = 250 Armor</span>
          <span className="text-violet-400">• 3-Layer Hybrid XP</span>
          <span className="text-amber-400">• -93% Anti-Macro</span>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        
        {/* Left Column: Skill Panel & Progression (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 space-y-4">
            <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <UserCheck className="w-4 h-4 text-emerald-400" />
              Painel de Habilidades (Skill Panel)
            </h4>
            <div className="space-y-2.5">
              {Object.keys(skills).map((id) => {
                const state = skills[id];
                const isActive = activeSkill === id;
                return (
                  <div
                    key={id}
                    onClick={() => setActiveSkill(id)}
                    className={`p-3 rounded-lg border transition-all cursor-pointer relative group flex flex-col gap-2 ${
                      isActive
                        ? "bg-amber-500/5 border-amber-500/40 text-amber-200 animate-pulse-subtle"
                        : "bg-slate-900/40 border-slate-800/60 hover:bg-slate-800/30 hover:border-slate-700 text-slate-300"
                    }`}
                  >
                    <div className="flex justify-between items-center">
                      <span className="font-extrabold text-xs tracking-wide uppercase flex items-center gap-1">
                        {id === "shielding" ? "🛡️" : id === "magic_level" ? "🔮" : "⚔️"} {skillNames[id]}
                        {id === "shielding" && (
                          <span className="text-[8px] bg-sky-950 border border-sky-800/60 text-sky-400 px-1 rounded uppercase tracking-widest scale-90">Defensive Canon</span>
                        )}
                      </span>
                      <span className="font-mono text-xs font-bold bg-slate-950 px-2 py-0.5 rounded border border-slate-800">
                        Nível {state.level}
                      </span>
                    </div>

                    {/* XP Progress Bar */}
                    <div className="space-y-1">
                      <div className="flex justify-between text-[9px] font-mono text-slate-500">
                        <span>XP: {state.xp} / {getXpForNextLevel(state.level)}</span>
                        <span>{((state.xp / getXpForNextLevel(state.level)) * 100).toFixed(0)}%</span>
                      </div>
                      <div className="w-full bg-slate-950 rounded-full h-1 border border-slate-900/60 overflow-hidden">
                        <div
                          className={`h-full transition-all duration-300 ${id === "shielding" ? "bg-gradient-to-r from-sky-500 to-sky-400" : "bg-gradient-to-r from-amber-500 to-amber-400"}`}
                          style={{ width: `${Math.min(100, (state.xp / getXpForNextLevel(state.level)) * 100)}%` }}
                        />
                      </div>
                    </div>

                    {/* Quick Editors */}
                    <div className="flex justify-end gap-1 pt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onClick={(e) => { e.stopPropagation(); adjustSkillLevel(id, -10); }}
                        className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[9px] font-mono text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                        title="Reduzir 10 níveis"
                      >
                        -10
                      </button>
                      <button
                        type="button"
                        onClick={(e) => { e.stopPropagation(); adjustSkillLevel(id, -1); }}
                        className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[9px] font-mono text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                        title="Reduzir 1 nível"
                      >
                        -1
                      </button>
                      <button
                        type="button"
                        onClick={(e) => { e.stopPropagation(); adjustSkillLevel(id, 1); }}
                        className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[9px] font-mono text-emerald-400 hover:bg-emerald-950 hover:text-emerald-300"
                        title="Aumentar 1 nível"
                      >
                        +1
                      </button>
                      <button
                        type="button"
                        onClick={(e) => { e.stopPropagation(); adjustSkillLevel(id, 10); }}
                        className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[9px] font-mono text-emerald-400 hover:bg-emerald-950 hover:text-emerald-300"
                        title="Aumentar 10 níveis"
                      >
                        +10
                      </button>
                      <button
                        type="button"
                        onClick={(e) => { e.stopPropagation(); adjustSkillLevel(id, 100); }}
                        className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[9px] font-mono text-emerald-400 hover:bg-emerald-950 hover:text-emerald-300"
                        title="Aumentar 100 níveis (Simular Prestígio)"
                      >
                        +100
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right Column: Simulator Actions, Curves, and Combat Output (7 cols) */}
        <div className="lg:col-span-7 space-y-6">

          {/* Prestige & Curve Visualizer */}
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 space-y-4">
            <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <Scale className="w-4 h-4 text-amber-400" />
              Prestige Band & Effective Combat Scaling (Authoritative)
            </h4>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-slate-900/40 border border-slate-800/60 p-3 rounded-lg text-center">
                <span className="text-[10px] text-slate-400 uppercase tracking-wider">Displayed Skill</span>
                <span className="block text-2xl font-extrabold text-slate-100 font-mono mt-1">
                  {currentSkillState.level}
                </span>
                <span className="text-[9px] text-slate-500 font-mono block mt-0.5">Nível Visível</span>
              </div>
              <div className="bg-slate-900/40 border border-slate-800/60 p-3 rounded-lg text-center">
                <span className="text-[10px] text-slate-400 uppercase tracking-wider">Effective Combat Skill</span>
                <span className="block text-2xl font-extrabold text-amber-400 font-mono mt-1">
                  {effectiveSkillVal.toFixed(1)}
                </span>
                <span className="text-[9px] text-slate-500 font-mono block mt-0.5">Para Fórmulas de Combate</span>
              </div>
              <div className={`border p-3 rounded-lg text-center flex flex-col justify-center items-center ${prestigeInfo.color}`}>
                <span className="text-[10px] uppercase tracking-wider opacity-90">Prestige Band Active</span>
                <span className="block text-base font-extrabold font-mono mt-1">
                  {prestigeInfo.band}
                </span>
                <span className="text-[9px] font-mono block mt-0.5 opacity-80">{prestigeInfo.mult}</span>
              </div>
            </div>

            {/* Prestige Bands Guide */}
            <div className="text-[11px] space-y-1 bg-slate-900/30 p-3 rounded-lg border border-slate-800/40">
              <span className="font-bold text-slate-300 block uppercase tracking-wide text-[9.5px]">Regras de Curva de Amortecimento de Combate:</span>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-2 text-[10px] font-mono mt-1">
                <div className="border border-emerald-500/20 bg-emerald-500/5 p-2 rounded">
                  <span className="font-bold text-emerald-400 block">Band 1: 1 a 150</span>
                  <span className="text-slate-400 block">Multiplicador: 100%</span>
                  <span className="text-slate-500 text-[9px]">Garante combate inicial nítido.</span>
                </div>
                <div className="border border-amber-500/20 bg-amber-500/5 p-2 rounded">
                  <span className="font-bold text-amber-400 block">Band 2: 151 a 250</span>
                  <span className="text-slate-400 block">Multiplicador: 75%</span>
                  <span className="text-slate-500 text-[9px]">Controle de crescimento médio.</span>
                </div>
                <div className="border border-rose-500/20 bg-rose-500/5 p-2 rounded">
                  <span className="font-bold text-rose-400 block">Band 3: 251+ (Infinito)</span>
                  <span className="text-slate-400 block">Multiplicador: 50%</span>
                  <span className="text-slate-500 text-[9px]">Prevenção de inflação de poder.</span>
                </div>
              </div>
              <div className="text-[9px] text-slate-400 leading-normal font-mono pt-1">
                <strong>Fórmula de Conversão de Prestígio:</strong> S se ≤ 150; 150 + (S-150)*0.75 se ≤ 250; 225 + (S-250)*0.50 se &gt; 250.
                <br />
                <em>Nota: O jogador sempre vê o valor total infinito ({currentSkillState.level}) em sua ficha!</em>
              </div>
            </div>
          </div>

          {/* Three-Layer XP Engine & Training Simulator */}
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 space-y-4">
            <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <Activity className="w-4 h-4 text-violet-400" />
              Simulador de Ganho Híbrido (3-Layer XP Engine)
            </h4>

            {/* Anti-Macro Warning Box */}
            {isAntiMacroTriggered && (
              <div className="bg-rose-950/40 border border-rose-900/60 p-3 rounded-lg text-rose-300 text-xs flex items-start gap-2 animate-pulse">
                <span className="text-base">⚠️</span>
                <div>
                  <strong className="font-bold uppercase tracking-wider">Low-risk repetitive combat detected</strong>
                  <p className="text-[11px] text-rose-400/90 mt-0.5">
                    O mecanismo anti-macro (<code className="text-rose-300">anti-macro-hard-penalty-rule</code>) aplicou uma penalidade severa de <strong>-93% de XP</strong> para evitar trapaças em alvos estáticos ou dummies.
                  </p>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Left Form: Inputs */}
              <div className="space-y-3 text-xs">
                <div className="space-y-1">
                  <label className="text-slate-400 font-bold block uppercase text-[10px]">Cenário de Combate / Ameaça:</label>
                  <select
                    value={mobDifficulty}
                    onChange={(e) => setMobDifficulty(e.target.value as any)}
                    className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-slate-200 font-semibold focus:outline-none focus:border-amber-500"
                  >
                    <option value="low">Dummy Treinador (Risco Nulo / Anti-Macro Ativo)</option>
                    <option value="standard">Monstro Comum (Risco Médio PvE)</option>
                    <option value="elite">Elite Tier Mob (Risco Alto PvE)</option>
                    <option value="boss">World Boss Encounter (Ameaça Extrema)</option>
                    <option value="pvp">Combate PvP Ativo (Hostilidade Máxima)</option>
                  </select>
                </div>

                <div className="space-y-1">
                  <label className="text-slate-400 font-bold block uppercase text-[10px]">Marco de Combate (Layer 3 Outcome):</label>
                  <select
                    value={selectedMilestone}
                    onChange={(e) => setSelectedMilestone(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-slate-200 font-semibold focus:outline-none focus:border-amber-500"
                  >
                    <option value="none">Nenhum / Ação de Combate Simples</option>
                    <option value="elite_kill">Eliminação de Monstro Elite (+50 XP)</option>
                    <option value="miniboss_kill">Derrota de Miniboss (+150 XP)</option>
                    <option value="boss_kill">Morte de Boss do Calabouço (+500 XP)</option>
                    <option value="pvp_kill">Assistência / Abate em PvP (+300 XP)</option>
                  </select>
                </div>

                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] text-slate-400 uppercase font-bold">
                    <span>Frequência de Ação:</span>
                    <span className="text-amber-400 font-mono">{combatFrequency.toFixed(1)} ações/seg</span>
                  </div>
                  <input
                    type="range"
                    min="1"
                    max="5"
                    step="0.5"
                    value={combatFrequency}
                    onChange={(e) => setCombatFrequency(parseFloat(e.target.value))}
                    className="w-full accent-amber-500 bg-slate-900 h-1.5 rounded"
                  />
                </div>

                <div className="flex gap-2 pt-2">
                  <button
                    type="button"
                    onClick={handleTrainClick}
                    className="flex-1 bg-gradient-to-r from-amber-600 to-amber-500 hover:from-amber-500 hover:to-amber-400 text-slate-950 font-extrabold text-xs py-2 px-3 rounded shadow transition-all duration-200 cursor-pointer text-center uppercase tracking-wider flex items-center justify-center gap-1.5"
                  >
                    {activeSkill === "shielding" ? "🛡️ Bloquear Golpe" : "⚔️ Executar Ataque"}
                  </button>
                  <button
                    type="button"
                    onClick={() => setAutoTrain(!autoTrain)}
                    className={`px-3 py-2 text-xs font-bold rounded uppercase border transition-all duration-200 cursor-pointer ${
                      autoTrain
                        ? "bg-emerald-500/10 border-emerald-500 text-emerald-400 animate-pulse"
                        : "bg-slate-900 border-slate-800 text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                    }`}
                  >
                    {autoTrain ? "Auto On" : "Auto Off"}
                  </button>
                </div>
              </div>

              {/* Right: Progression Exponential Curves */}
              <div className="space-y-2 bg-slate-900/30 border border-slate-800/40 p-3 rounded-lg flex flex-col justify-between">
                <span className="text-[10px] font-bold text-slate-300 uppercase tracking-wide block border-b border-slate-850 pb-1">XP Curve Constants & Formulas:</span>
                <div className="space-y-1.5 text-[10px] font-mono text-slate-400">
                  <div className="flex justify-between">
                    <span>• Lvl 1 a 30:</span>
                    <span className="text-emerald-400 font-bold">Rápido (~100 XP)</span>
                  </div>
                  <div className="flex justify-between">
                    <span>• Lvl 31 a 60:</span>
                    <span className="text-sky-400 font-bold">Moderado (~1k-10k)</span>
                  </div>
                  <div className="flex justify-between">
                    <span>• Lvl 61 a 90:</span>
                    <span className="text-indigo-400 font-bold">Lento (~11k-90k)</span>
                  </div>
                  <div className="flex justify-between">
                    <span>• Lvl 91 a 150:</span>
                    <span className="text-amber-400 font-bold">Muito Lento (&gt;100k)</span>
                  </div>
                  <div className="flex justify-between">
                    <span>• Lvl 151+:</span>
                    <span className="text-rose-400 font-bold">Prestigioso</span>
                  </div>
                </div>
                <div className="border-t border-slate-900 pt-1.5 text-[9px] text-slate-500 font-mono leading-normal">
                  <strong className="text-slate-400 font-bold uppercase tracking-wider block text-[8px] mb-0.5">Exponential formula:</strong>
                  <code>XP_Req(L) = Math.floor(100 * Math.pow(1.08, L))</code>
                </div>
              </div>
            </div>
          </div>

          {/* Liquid Damage Validator & Combat Output Preview */}
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 space-y-4">
            <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <Sword className="w-4 h-4 text-rose-500" />
              Validador de Dano Líquido & Combat Output (Authoritative)
            </h4>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
              <div className="space-y-2.5">
                {/* Weapon selection */}
                <div className="space-y-1">
                  <label className="text-slate-400 font-bold text-[9px] uppercase tracking-wide block">Wielded Weapon Profile:</label>
                  <select
                    value={selectedWeapon}
                    onChange={(e) => {
                      const val = e.target.value;
                      setSelectedWeapon(val);
                      if (val === "sword_t5") setWeaponBaseDmg(165);
                      else if (val === "sword_t3") setWeaponBaseDmg(58);
                      else if (val === "sword_t1") setWeaponBaseDmg(15);
                      else if (val === "axe_t4") setWeaponBaseDmg(135);
                      else if (val === "bow_t5") setWeaponBaseDmg(145);
                    }}
                    className="w-full bg-slate-900 border border-slate-800 rounded p-1.5 text-slate-200 font-semibold focus:outline-none"
                  >
                    <option value="sword_t1">Rusty Recruit Sword [T1] - 15 Base Dmg</option>
                    <option value="sword_t3">Pyre-Forged Falchion [T3] - 58 Base Dmg</option>
                    <option value="sword_t5">Aegis-Breaker Calamity [T5] - 165 Base Dmg</option>
                    <option value="axe_t4">Magma Core Decapitator [T4] - 135 Base Dmg</option>
                    <option value="bow_t5">Astral Starsong [T5] - 145 Base Dmg</option>
                  </select>
                </div>

                {/* Combat Controls */}
                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1">
                    <label className="text-slate-400 font-bold text-[9px] uppercase block">Armadura Alvo (Def):</label>
                    <input
                      type="number"
                      min="0"
                      max="1500"
                      value={defenderArmor}
                      onChange={(e) => setDefenderArmor(Math.max(0, parseInt(e.target.value) || 0))}
                      className="w-full bg-slate-900 border border-slate-800 rounded p-1 text-slate-200 font-semibold text-center focus:outline-none focus:border-amber-500"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-slate-400 font-bold text-[9px] uppercase block">Mult. Elemental:</label>
                    <input
                      type="number"
                      min="1.0"
                      max="3.0"
                      step="0.05"
                      value={elementalMultiplier}
                      onChange={(e) => setElementalMultiplier(Math.max(1.0, parseFloat(e.target.value) || 1.0))}
                      className="w-full bg-slate-900 border border-slate-800 rounded p-1 text-slate-200 font-semibold text-center focus:outline-none focus:border-amber-500"
                    />
                  </div>
                </div>

                {/* Toggles */}
                <div className="flex flex-col gap-1 text-[10px] text-slate-400 font-mono">
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={isCritForced}
                      onChange={(e) => setIsCritForced(e.target.checked)}
                      className="accent-amber-500 rounded"
                    />
                    Forçar Golpe Crítico (Critical Strike)
                  </label>
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={isAssassinClass}
                      onChange={(e) => setIsAssassinClass(e.target.checked)}
                      className="accent-amber-500 rounded"
                    />
                    Usar Multiplicador de Assassino (2.2x vs 1.5x)
                  </label>
                </div>
              </div>

              {/* Combat Outputs and Math step list */}
              <div className="bg-slate-900/50 border border-slate-850 p-3.5 rounded-lg space-y-2.5">
                <span className="text-[10px] uppercase font-extrabold text-amber-400 block tracking-wider border-b border-slate-800 pb-0.5">Damage Step Validation:</span>
                <div className="space-y-1 text-[9.5px] font-mono text-slate-300">
                  <div className="flex justify-between border-b border-slate-950 pb-0.5">
                    <span>1. Base Weapon Dmg:</span>
                    <span>{weaponBaseDmg}</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-950 pb-0.5">
                    <span>2. Effective Combat Skill ({skillNames[activeSkill]}):</span>
                    <span>{effectiveSkillVal.toFixed(1)} Lvl</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-950 pb-0.5">
                    <span>3. Raw pre-mitigation Damage:</span>
                    <span className="text-sky-300 font-bold">{rawBaseDamage.toFixed(1)} Dmg</span>
                  </div>
                  {isCritForced && (
                    <div className="flex justify-between border-b border-slate-950 pb-0.5 text-amber-400">
                      <span>• Critical Hit Multiplier ({criticalMultiplier}x):</span>
                      <span>{currentCritDmg.toFixed(1)} Dmg</span>
                    </div>
                  )}
                  <div className="flex justify-between border-b border-slate-950 pb-0.5">
                    <span>4. Elemental-Scale Damage ({elementalMultiplier}x):</span>
                    <span>{rawDamageValue.toFixed(1)} Dmg</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-950 pb-0.5 text-emerald-400">
                    <span>5. Mitigated Damage (K=250):</span>
                    <span>{mitigatedDamageValue.toFixed(1)} Dmg ({(armorMitigationPercent * 100).toFixed(1)}%)</span>
                  </div>
                  <div className="flex justify-between pt-1 text-[11px] font-extrabold text-rose-400 border-t border-slate-800">
                    <span>6. LIQUID DAMAGE OUTCOME:</span>
                    <span className="bg-rose-950/20 px-1 rounded border border-rose-900/20">{liquidDamageValue.toFixed(1)} HP</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Combat XP Breakdown Panel */}
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 space-y-4">
            <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <Layers className="w-4 h-4 text-teal-400" />
              Detalhamento de Ganho de XP de Combate (Hybrid XP Breakdown)
            </h4>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div className="bg-slate-900/30 border border-slate-850 p-2.5 rounded text-center">
                <span className="text-[9px] uppercase font-semibold text-slate-400 block tracking-wide">Layer 1: Action XP</span>
                <span className="text-lg font-bold text-teal-400 font-mono block mt-1">+{actionXpValue}</span>
                <span className="text-[8px] text-slate-500 font-mono block mt-0.5">Garantido por Ação</span>
              </div>
              <div className="bg-slate-900/30 border border-slate-850 p-2.5 rounded text-center">
                <span className="text-[9px] uppercase font-semibold text-slate-400 block tracking-wide">Layer 2: Damage XP</span>
                <span className="text-lg font-bold text-violet-400 font-mono block mt-1">+{damageXpValue}</span>
                <span className="text-[8px] text-slate-500 font-mono block mt-0.5">Liquid * 0.25 + Mitig * 0.10</span>
              </div>
              <div className="bg-slate-900/30 border border-slate-850 p-2.5 rounded text-center">
                <span className="text-[9px] uppercase font-semibold text-slate-400 block tracking-wide">Layer 3: Outcome XP</span>
                <span className="text-lg font-bold text-amber-400 font-mono block mt-1">+{outcomeBonusXpValue}</span>
                <span className="text-[8px] text-slate-500 font-mono block mt-0.5">Mastery Milestones</span>
              </div>
              <div className="bg-slate-900/30 border border-slate-850 p-2.5 rounded text-center relative overflow-hidden">
                <span className="text-[9px] uppercase font-semibold text-slate-400 block tracking-wide">Final Net Gain</span>
                <span className="text-lg font-bold text-emerald-400 font-mono block mt-1">+{finalXpValue}</span>
                {isAntiMacroTriggered && (
                  <span className="text-[8px] text-rose-400 font-mono block font-bold mt-0.5 animate-pulse">-93% Penalty</span>
                )}
                {!isAntiMacroTriggered && (
                  <span className="text-[8px] text-emerald-500 font-mono block mt-0.5">100% Efficiency</span>
                )}
              </div>
            </div>
          </div>

          {/* Training Logs */}
          <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 space-y-2">
            <div className="flex justify-between items-center border-b border-slate-900 pb-2">
              <span className="text-[10px] uppercase font-bold text-slate-400 tracking-wider">Histórico de Treinamento de Habilidades:</span>
              <button
                type="button"
                onClick={() => setCombatLog([])}
                className="text-[9px] text-slate-500 hover:text-slate-300 font-mono"
              >
                Limpar Logs
              </button>
            </div>
            <div className="h-[120px] overflow-y-auto bg-slate-950 p-2 rounded border border-slate-900/60 font-mono text-[9.5px] space-y-1 scrollbar-thin">
              {combatLog.length === 0 ? (
                <div className="text-slate-600 italic text-center pt-8">Nenhum log de treino recente. Clique em "Atacar" para iniciar.</div>
              ) : (
                combatLog.map((log) => {
                  let color = "text-slate-400";
                  if (log.type === "level") color = "text-emerald-400 font-bold bg-emerald-950/20 px-1 py-0.5 rounded border border-emerald-900/20 animate-pulse";
                  else if (log.type === "info") color = "text-sky-300";
                  else if (log.type === "warning") color = "text-rose-400";
                  return (
                    <div key={log.id} className={`${color} leading-relaxed`}>
                      {log.text}
                    </div>
                  );
                })
              )}
            </div>
          </div>

        </div>

      </div>

    </div>
  );
}


export default function App() {
  const [activeTab, setActiveTab] = useState<string>("workspace");
  const [selectedFile, setSelectedFile] = useState<CodeFile>(codeFiles[0]);
  const [copied, setCopied] = useState(false);
  const [expandedFolders, setExpandedFolders] = useState<Record<string, boolean>>({
    root: true,
    src: true,
    core: true,
    stateMachine: true,
    states: true,
    ui: true,
    network: true,
    backend: true,
    cmd: true,
    gateway: false,
    auth: false,
    world: false,
    pkg: false,
    protocol: false,
    db: false,
    config: false,
    migrations: false,
    movement: false,
  });

  // State Machine Simulator state
  const [simState, setSimState] = useState<AppStateType>("Boot");
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [logFilter, setLogFilter] = useState<"all" | "info" | "warning" | "error">("all");
  const [simUsername, setSimUsername] = useState("Gabriela_Paladin");
  const [isSimulatingLoading, setIsSimulatingLoading] = useState(false);
  const [loadingPercent, setLoadingPercent] = useState(0);
  const [isSimulatingTcp, setIsSimulatingTcp] = useState(false);

  const consoleEndRef = useRef<HTMLDivElement>(null);

  // --- SPRINT 4 TASK 2 STATES ---
  const [dungeonId, setDungeonId] = useState<"crypt_of_shadows" | "dragon_lair">("crypt_of_shadows");
  const [dungeonMode, setDungeonMode] = useState<"solo" | "party" | "raid">("solo");
  const [dungeonActive, setDungeonActive] = useState(false);
  const [dungeonTimeLeft, setDungeonTimeLeft] = useState(1800); // 30 mins
  const [dungeonCheckpoint, setDungeonCheckpoint] = useState("Entrance");
  const [dungeonOffset, setDungeonOffset] = useState(2000.0);

  // --- SPRINT 3 TASK 5 PROGRESSION STATES ---
  const [progLevel, setProgLevel] = useState(1);
  const [progClass, setProgClass] = useState("Novice");
  const [progSubclass, setProgSubclass] = useState("");
  
  // Canonical player affinity schema: independent levels and XP
  const [affinityFireLvl, setAffinityFireLvl] = useState(1);
  const [affinityFireXp, setAffinityFireXp] = useState(0);
  
  const [affinityIceLvl, setAffinityIceLvl] = useState(1);
  const [affinityIceXp, setAffinityIceXp] = useState(0);
  
  const [affinityHolyLvl, setAffinityHolyLvl] = useState(1);
  const [affinityHolyXp, setAffinityHolyXp] = useState(0);
  
  const [affinityShadowLvl, setAffinityShadowLvl] = useState(1);
  const [affinityShadowXp, setAffinityShadowXp] = useState(0);
  
  const [affinityNatureLvl, setAffinityNatureLvl] = useState(1);
  const [affinityNatureXp, setAffinityNatureXp] = useState(0);

  const [awakenedAffinity, setAwakenedAffinity] = useState(""); // permanent lock field: "fire" | "ice" | "holy" | "shadow" | "nature" | ""

  // Keep virtual mapped values to prevent broken legacy imports/selectors
  const progAffinityFire = affinityFireLvl;
  const progAffinityIce = affinityIceLvl;
  const progAffinityHoly = affinityHolyLvl;
  const progAffinityShadow = affinityShadowLvl;
  const progAffinityNature = affinityNatureLvl;

  const [progX, setProgX] = useState(100.0);
  const [progY, setProgY] = useState(100.0);
  const [lastAffinityTime, setLastAffinityTime] = useState(0);
  const [spamWarning, setSpamWarning] = useState(false);

  // States for the interactive Gear Compatibility Simulator / Stat Calculator
  const [selectedPlayerAffinity, setSelectedPlayerAffinity] = useState("neutral");
  const [selectedGearElement, setSelectedGearElement] = useState("fire");

  // Keep player affinity selection synchronized with awakenedAffinity state dynamically
  useEffect(() => {
    if (awakenedAffinity) {
      setSelectedPlayerAffinity(awakenedAffinity);
    } else {
      setSelectedPlayerAffinity("neutral");
    }
  }, [awakenedAffinity]);

  // Authoritative gear compatibility formula
  const getGearEfficiency = (playerAff: string, gearAff: string) => {
    if (playerAff === "neutral") {
      return { pct: 75, type: "Neutral", desc: "Seu personagem está no estado Neutro (Nenhuma Subclasse desperta). Todos os equipamentos de qualquer elemento operam com 75% de eficiência." };
    }
    
    if (playerAff === gearAff) {
      return { pct: 100, type: "Matching", desc: "Afinidade Idêntica! O elemento do equipamento coincide perfeitamente com a sua alma elemental desperta (100% de eficácia total)." };
    }
    
    // Opposed pairs: Fire <-> Ice, Holy <-> Shadow
    const opposedPairs: Record<string, string> = {
      fire: "ice",
      ice: "fire",
      holy: "shadow",
      shadow: "holy"
    };
    
    if (opposedPairs[playerAff] === gearAff) {
      return { pct: 50, type: "Opposed", desc: "ELEMENTOS OPOSTOS! Equipar um item deste elemento gera um conflito de energias destrutivo (Fire ↔ Ice, Holy ↔ Shadow). Penalidade severa de 50% de eficiência aplicada aos atributos do item!" };
    }
    
    // Otherwise, Nature has no opposition, or other non-opposed combination (75% efficiency)
    return { pct: 75, type: "Neutral", desc: "Elementos Neutros/Compatíveis. O equipamento elemental difere da sua afinidade desperta, mas não gera conflito de alma. Opera com 75% de eficácia padrão." };
  };

  // Canonical Experience Curve formulas & reward scaling (CANONICAL EXPERIENCE CURVE PATCH)
  const getRequiredXpForLevel = (lvl: number): number => {
    if (lvl <= 1) return 100;
    if (lvl <= 100) {
      return lvl * lvl * 100;
    } else if (lvl <= 200) {
      const base = lvl * lvl * 100;
      return base * (1 + (lvl - 100) * 0.05);
    } else if (lvl <= 400) {
      const base = lvl * lvl * 100;
      return base * 6 * (1 + (lvl - 200) * 0.5);
    } else {
      const base = lvl * lvl * 100;
      return base * 606 * (1 + (lvl - 400) * 25);
    }
  };

  const getMonsterXpForLevel = (lvl: number): number => {
    if (lvl <= 100) {
      return 10 * lvl;
    } else if (lvl <= 200) {
      return 1000 + (lvl - 100) * 15;
    } else if (lvl <= 400) {
      return 2500 + (lvl - 200) * 2.5;
    } else {
      return Math.min(5000, 3000 + (lvl - 400) * 0.2);
    }
  };

  const getQuestXpForLevel = (lvl: number): number => {
    if (lvl <= 100) {
      return 200 * lvl;
    } else if (lvl <= 200) {
      return 20000 + (lvl - 100) * 300;
    } else if (lvl <= 400) {
      return 50000 + (lvl - 200) * 125;
    } else {
      return Math.min(100000, 75000 + (lvl - 400) * 5);
    }
  };

  const getDungeonXpPercentForLevel = (lvl: number): string => {
    if (lvl <= 100) {
      return (Math.max(1, 15 - lvl * 0.1)).toFixed(1) + "%";
    } else if (lvl <= 200) {
      return (Math.max(0.01, 5 - (lvl - 100) * 0.048)).toFixed(2) + "%";
    } else if (lvl <= 400) {
      const val = 0.2 * Math.pow(0.95, lvl - 200);
      return val.toFixed(4) + "%";
    } else {
      const val = 0.0001 * Math.pow(0.99, lvl - 400);
      return val.toFixed(8) + "%";
    }
  };

  const getCombatPowerBreakdown = (lvl: number) => {
    if (lvl <= 100) {
      return { level: 70, gear: 10, skill: 20 };
    } else if (lvl <= 200) {
      const t = (lvl - 100) / 100;
      return {
        level: Math.round(70 - t * 30),
        gear: Math.round(10 + t * 20),
        skill: Math.round(20 + t * 10)
      };
    } else if (lvl <= 400) {
      const t = (lvl - 200) / 200;
      return {
        level: Math.round(40 - t * 25),
        gear: Math.round(30 + t * 20),
        skill: Math.round(30 + t * 5)
      };
    } else {
      const t = Math.min(1, (lvl - 400) / 600);
      return {
        level: Math.round(15 - t * 13),
        gear: Math.round(50 + t * 10),
        skill: Math.round(35 + t * 3)
      };
    }
  };

  const getPhaseInfo = (lvl: number) => {
    if (lvl <= 100) {
      return {
        phase: 1,
        title: "Early Game (Fase Inicial)",
        badge: "FAST",
        color: "text-emerald-400 border-emerald-500/20 bg-emerald-500/10",
        desc: "Fase acelerada de introdução. O jogador aprende mecânicas de classe, escolhe vocação (Lvl 10) e busca o Despertar Elemental (Lvl 100 com Afinidade 100)."
      };
    } else if (lvl <= 200) {
      return {
        phase: 2,
        title: "Mid Game (Fase Intermediária)",
        badge: "MODERATE",
        color: "text-blue-400 border-blue-500/20 bg-blue-500/10",
        desc: "Especialização do personagem. Slowdown de XP visível. Foco em refinar a subclasse escolhida, dominar sinergias elementais e enfrentar masmorras regionais."
      };
    } else if (lvl <= 400) {
      return {
        phase: 3,
        title: "Advanced Game (Fase Avançada)",
        badge: "SLOW",
        color: "text-amber-400 border-amber-500/20 bg-amber-500/10",
        desc: "Cada nível ganho é um marco de prestígio. Progresso migra gradativamente de níveis puros para gear optimization, caça de elites e masmorras avançadas."
      };
    } else {
      return {
        phase: 4,
        title: "Endurance Game (Modo Lenda)",
        badge: "VERY SLOW",
        color: "text-purple-400 border-purple-500/20 bg-purple-500/10",
        desc: "Progresso sem limites rígidos. O nível 9999 é um teto teórico. Ganhar níveis é extremamente raro; o foco é total na maestria PvE, itens Lendários e coleções."
      };
    }
  };

  const formatXpNumber = (num: number): string => {
    if (num >= 1e15) return (num / 1e15).toFixed(2) + "Q (Quatrilhão)";
    if (num >= 1e12) return (num / 1e12).toFixed(2) + "T (Trilhão)";
    if (num >= 1e9) return (num / 1e9).toFixed(2) + "B (Bilhão)";
    if (num >= 1e6) return (num / 1e6).toFixed(2) + "M (Milhão)";
    if (num >= 1000) return num.toLocaleString("pt-BR");
    return num.toString();
  };

  // Canonical Monster AI Bible Helpers (CANONICAL MONSTER AI BIBLE PATCH)
  const getAggroFocusScore = (archetype: string, player: string, rawThreat: number) => {
    let instinctScore = 0;
    if (archetype === "BRUISER_TANK") {
      instinctScore = player === "Gabriela_Paladin" ? 100 : 20;
    } else if (archetype === "PREDATOR_ASSASSIN") {
      if (player === "AI_Bot_Priest") instinctScore = 100;
      else if (player === "AI_Bot_Mage") instinctScore = 80;
      else instinctScore = 15;
    } else if (archetype === "RANGED_HUNTER") {
      instinctScore = player !== "Gabriela_Paladin" ? 95 : 30;
    } else if (archetype === "CASTER") {
      instinctScore = player === "AI_Bot_Mage" || player === "AI_Bot_Priest" ? 90 : 40;
    } else if (archetype === "SUPPORT") {
      instinctScore = player === "AI_Bot_Priest" ? 100 : 50;
    } else if (archetype === "ELITE_COMMANDER") {
      instinctScore = 75;
    }

    let threatWeight = 0.5;
    let instinctWeight = 0.5;
    if (archetype === "BRUISER_TANK") {
      threatWeight = 0.8;
      instinctWeight = 0.2;
    } else if (archetype === "PREDATOR_ASSASSIN") {
      threatWeight = 0.4;
      instinctWeight = 0.6;
    }

    const maxThreatInMap = Math.max(...(Object.values(aiThreatMap) as number[]), 1);
    const normalizedThreat = (rawThreat / maxThreatInMap) * 100;
    const finalScore = normalizedThreat * threatWeight + instinctScore * instinctWeight;
    return Math.round(finalScore);
  };

  const handleSimulateSoloPull = () => {
    setSoloPullStatus("Simulando... Puxando Lobo #1 do bando.");
    addSimLogs([
      { type: "info", text: "[PackAI] Você atacou Lobo #1 (Solo Aggro Ativo)." },
      { type: "warning", text: "[PackAI] Lobo #1 entrou em combate com você. Distância: 12 metros." },
      { type: "info", text: "[PackAI] Lobo #2 e Lobo #3 permanecem neutros (Filosofia Solo Aggro validada)." }
    ]);
    setTimeout(() => {
      setSoloPullStatus("Sucesso! Apenas Lobo #1 foi aggroado (Zero aggro social passivo).");
    }, 1200);
  };

  const evaluateSurvivalBehavior = (species: string, hpPercent: number) => {
    if (species === "dragon") {
      if (hpPercent <= 25) {
        return {
          mode: "Draconic Rage (Fúria Dracônica)",
          description: "Abaixo de 25% HP! +50% Vel de Ataque, +100% Frequência de Sopro, +40% Intensidade Arcana. Nunca recua, nunca foge!",
          badge: "FOGO OU MORTE",
          color: "text-red-400 border-red-500/20 bg-red-950/40"
        };
      }
      return {
        mode: "Dragon Survival Law (Lei do Dragão)",
        description: "Regra autoritativa ativa: Um dragão nunca recua após o início do combate. O resultado deve ser a morte do jogador ou do dragão.",
        badge: "IMORTALIDADE",
        color: "text-rose-400 border-rose-500/20 bg-rose-950/20"
      };
    }
    
    if (species === "demon") {
      return {
        mode: "Death Before Retreat (Morte antes de Recuar)",
        description: "Demônios nunca recuam a menos que explicitamente ordenados por sua hierarquia superior. Lutará até a morte.",
        badge: "DEATH_BEFORE_RETREAT",
        color: "text-purple-400 border-purple-500/20 bg-purple-950/20"
      };
    }

    if (species === "beast") {
      if (hpPercent <= 40) {
        return {
          mode: "Frenzy Under Death (Frenesi de Sangue)",
          description: "Ativou Frenesi! Ataque e Velocidade aumentados massivamente, Defesa reduzida a zero.",
          badge: "FRENZY_ACTIVE",
          color: "text-amber-400 border-amber-500/20 bg-amber-950/20"
        };
      }
      return {
        mode: "Comportamento Normal",
        description: "Monstro de tipo Fera saudável. Lutando normalmente com garras e dentes.",
        badge: "NORMAL",
        color: "text-slate-400 border-slate-800 bg-slate-900/40"
      };
    }

    if (species === "assassin") {
      if (hpPercent <= 30) {
        return {
          mode: "Tactical Retreat (Retirada Tática)",
          description: "Recuando de forma ágil para as sombras! Buscando reposicionamento e mantendo distância temporária.",
          badge: "REPOSITIONING",
          color: "text-blue-400 border-blue-500/20 bg-blue-950/20"
        };
      }
      return {
        mode: "Comportamento Predador",
        description: "Focado em assassinar alvos frágeis com alto dano de rajada.",
        badge: "STALKING",
        color: "text-slate-400 border-slate-800 bg-slate-900/40"
      };
    }

    if (hpPercent <= 30) {
      return {
        mode: "Cowardly Flee (Fuga Covarde)",
        description: "Goblins e saqueadores fracos entram em pânico total e começam a fugir desordenadamente do combate!",
        badge: "PANIC_FLEE",
        color: "text-emerald-400 border-emerald-500/20 bg-emerald-950/20"
      };
    }
    return {
      mode: "Comportamento Territorial",
      description: "Patrulhando e defendendo agressivamente seu território imediato de intrusos.",
      badge: "ALERT",
      color: "text-slate-400 border-slate-800 bg-slate-900/40"
    };
  };

  const getLeashStatus = (distance: number, tier: number) => {
    let maxRange = 0;
    if (tier === 1) maxRange = 12;
    else if (tier === 2) maxRange = 30;
    else if (tier === 3) maxRange = 80;
    else if (tier === 4) maxRange = 200;
    else maxRange = Infinity;

    if (distance > maxRange) {
      return {
        leashed: true,
        text: `HARD LEASH RESET! Distância simulada (${distance}m) ultrapassa o limite de ${maxRange === Infinity ? "Sem Limite" : maxRange + "m"}. O monstro perde o aggro e retorna ao spawn com vida cheia.`,
        color: "text-rose-400 border-rose-500/20 bg-rose-950/40"
      };
    }
    return {
      leashed: false,
      text: `DENTRO DO LIMITE. Distância simulada (${distance}m) está segura dentro do raio de leash de ${maxRange === Infinity ? "Sem Limite" : maxRange + "m"}. O monstro perseguirá você ativamente.`,
      color: "text-emerald-400 border-emerald-500/20 bg-emerald-950/40"
    };
  };

  const [selectedContinent, setSelectedContinent] = useState<"main_continent" | "nature" | "shadow" | "fire_continent" | "holy" | "ice" | "abyssia">("main_continent");
  const [activeSettlement, setActiveSettlement] = useState<string | null>("ironhold_bastion");
  const [holyQuestCompleted, setHolyQuestCompleted] = useState(false);
  const [abyssiQuestCompleted, setAbyssiQuestCompleted] = useState(false);
  const [abyssiPermissionFlag, setAbyssiPermissionFlag] = useState(false);
  const [abyssStepVoidTerror, setAbyssStepVoidTerror] = useState<"idle" | "running" | "completed">("idle");
  const [abyssStepVoidFracture, setAbyssStepVoidFracture] = useState<"idle" | "running" | "completed">("idle");
  const [pillarCourage, setPillarCourage] = useState<"idle" | "running" | "completed">("idle");
  const [pillarWisdom, setPillarWisdom] = useState<"idle" | "riddle" | "completed">("idle");
  const [pillarPurity, setPillarPurity] = useState<"idle" | "running" | "completed">("idle");
  const [riddleAnswer, setRiddleAnswer] = useState<string>("");
  const [riddleFeedback, setRiddleFeedback] = useState<string>("");
  
  // Boss state
  const [bossHp, setBossHp] = useState(5000);
  const [bossMaxHp, setBossMaxHp] = useState(5000);
  const [bossPhase, setBossPhase] = useState(1);
  const [bossEnraged, setBossEnraged] = useState(false);
  const [bossCombatSeconds, setBossCombatSeconds] = useState(0);
  const [bossInCombat, setBossInCombat] = useState(false);
  const [bossTelegraphActive, setBossTelegraphActive] = useState(false);
  const [bossTelegraphTimer, setBossTelegraphTimer] = useState(0); // countdown
  const [bossAdds, setBossAdds] = useState<{ id: string; hp: number; maxHp: number }[]>([]);
  const [bossResetTimer, setBossResetTimer] = useState<number | null>(null); // anti-reset exploit
  
  // World Boss state
  const [worldBossActive, setWorldBossActive] = useState(true);
  const [worldBossHp, setWorldBossHp] = useState(50000);
  const [worldBossMaxHp, setWorldBossMaxHp] = useState(50000);
  const [worldBossThreat, setWorldBossThreat] = useState<Record<string, number>>({
    "Gabriela_Paladin": 0,
    "AI_Bot_Mage": 0,
    "AI_Bot_Priest": 0
  });

  // Loot reservations
  const [lootReservations, setLootReservations] = useState<{ id: string; itemID: string; qty: number; timer: number; claimed: boolean }[]>([]);
  
  // Persistence database states
  const [persistentLogs, setPersistentLogs] = useState<{ id: string; type: string; details: string; timestamp: string }[]>([
    { id: "1", type: "Lockout", details: "Raid Lockout checked: No active lockouts for Gabriela_Paladin", timestamp: "14:39:00" },
    { id: "2", type: "Checkpoint", details: "Dungeon checkpoints initialized for character: Gabriela_Paladin", timestamp: "14:39:01" }
  ]);

  // Monster AI Bible state variables (CANONICAL MONSTER AI BIBLE PATCH)
  const [selectedAiArchetype, setSelectedAiArchetype] = useState<"BRUISER_TANK" | "PREDATOR_ASSASSIN" | "RANGED_HUNTER" | "CASTER" | "SUPPORT" | "ELITE_COMMANDER">("BRUISER_TANK");
  const [aiThreatMap, setAiThreatMap] = useState<Record<string, number>>({
    "Gabriela_Paladin": 300,
    "AI_Bot_Mage": 200,
    "AI_Bot_Priest": 150
  });
  const [fleeTestHp, setFleeTestHp] = useState<number>(100);
  const [selectedSurvivalSpecies, setSelectedSurvivalSpecies] = useState<"goblin" | "assassin" | "beast" | "demon" | "dragon">("goblin");
  const [leashDistance, setLeashDistance] = useState<number>(10);
  const [selectedLeashTier, setSelectedLeashTier] = useState<1 | 2 | 3 | 4 | 5>(2);
  const [soloPullStatus, setSoloPullStatus] = useState<string>("Inativo. Clique em 'Simular Pull Individual' para testar.");

  // Canonical Bestiary Bible state variables (CANONICAL BESTIARY BIBLE PATCH)
  const [selectedBestiaryFamily, setSelectedBestiaryFamily] = useState<string>("humanoids");
  const [spawnCategory, setSpawnCategory] = useState<"Standard Spawn" | "Dynamic Ecosystem Spawn" | "Rare Spawn" | "Legendary Spawn">("Standard Spawn");
  const [playerDensity, setPlayerDensity] = useState<number>(5); // 0 to 50 players nearby
  const [huntingPressure, setHuntingPressure] = useState<"low" | "medium" | "high">("medium");
  const [spawnSimResult, setSpawnSimResult] = useState<string>("");
  const [demonSummoner, setDemonSummoner] = useState<string>("Commanders");
  const [demonSummonTarget, setDemonSummonTarget] = useState<string>("Common Demons");
  const [demonRegion, setDemonRegion] = useState<"Main Continent" | "Volcanic Demon Zone" | "Deep Shadow Territory">("Main Continent");
  const [demonSummonLog, setDemonSummonLog] = useState<string>("Pronto para simulação.");
  const [dragonOne, setDragonOne] = useState<string>("void_dragon_lord");
  const [dragonTwo, setDragonTwo] = useState<string>("volcanic_dragon");
  const [dragonTerritory, setDragonTerritory] = useState<string>("Ancient Volcano Peak");
  const [dragonCombatLog, setDragonCombatLog] = useState<string>("Pronto para simular coexistência territorial.");

  // Loot System Bible & Death Penalty Bible states
  const [playerInventory, setPlayerInventory] = useState<{ id: string; name: string; qty: number; value: number; type: string }[]>([
    { id: "bronze_coin", name: "Moeda de Bronze", qty: 250, value: 1, type: "currency" },
    { id: "silver_coin", name: "Moeda de Prata", qty: 15, value: 100, type: "currency" },
    { id: "gold_coin", name: "Moeda de Ouro", qty: 2, value: 10000, type: "currency" },
    { id: "iron_ore", name: "Minério de Ferro", qty: 10, value: 5, type: "material" },
    { id: "potion_heal", name: "Poção de Cura", qty: 5, value: 15, type: "consumable" }
  ]);
  const [playerEquipped, setPlayerEquipped] = useState<{ slot: string; id: string; name: string; value: number }[]>([
    { slot: "Arma", id: "sword_basic", name: "Espada Básica", value: 100 },
    { slot: "Armadura", id: "armor_leather", name: "Armadura de Couro", value: 200 },
    { slot: "Acessório", id: "ring_crit", name: "Anel Crítico", value: 500 },
    { slot: "Escudo", id: "shield_wooden", name: "Escudo de Madeira", value: 150 }
  ]);
   const [activeBlessingsCount, setActiveBlessingsCount] = useState<number>(7);
  // Guild & Social State
  const [simulatedGuildName, setSimulatedGuildName] = useState<string>("Patrulheiros da Alvorada");
  const [guildMembersCount, setGuildMembersCount] = useState<number>(12);
  const [hasGuildHouse, setHasGuildHouse] = useState<boolean>(false);
  const [guildGoldBalance, setGuildGoldBalance] = useState<number>(150); // Starts with 150 gold, enough to buy house costing 100 gold
  
  // Trade & Market State
  const [carriedCoins, setCarriedCoins] = useState<number>(450); // Carried Gold Coins (unprotected, subject to death loss)
  const [bankedCoins, setBankedCoins] = useState<number>(1200); // Banked Gold Coins (fully protected from all death penalties)
  const [activeMarketListings, setActiveMarketListings] = useState<Array<{ id: string; name: string; price: number; seller: string }>>([
    { id: "weapon_t3_listed", name: "Lâmina Rúnica da Penumbra [T3]", price: 350, seller: "Eldrin_Mage" },
    { id: "armor_t2_listed", name: "Cota de Malha Pesada [T2]", price: 180, seller: "Kaelen_Hunter" },
    { id: "shield_wooden_listed", name: "Escudo de Carvalho Reforçado [T1]", price: 45, seller: "Aria_Tracker" }
  ]);
  const [selfOffer, setSelfOffer] = useState<string>("");
  const [selfOfferCurrency, setSelfOfferCurrency] = useState<number>(0);
  const [otherOffer, setOtherOffer] = useState<string>("Fragmento de Relíquia Abissal");
  const [otherOfferCurrency, setOtherOfferCurrency] = useState<number>(120);
  const [secureTradeStatus, setSecureTradeStatus] = useState<"idle" | "initiated" | "offered" | "confirmed_self" | "confirmed_both" | "completed">("idle");
  
  // Race & Character Creation State
  const [createdCharName, setCreatedCharName] = useState<string>("Valen_Sunblade");
  const [createdCharRace, setCreatedCharRace] = useState<"Forest Elf" | "Human" | "Dwarf" | "Ice Elf" | "Green Orc">("Human");
  const [createdCharClass, setCreatedCharClass] = useState<string>("Knight");
  const [createdCharGender, setCreatedCharGender] = useState<"Male" | "Female">("Male");
  const [characterCreationLog, setCharacterCreationLog] = useState<string>("Pronto para inicialização do herói.");

  // Class & Vocation State
  const [vocationLevel, setVocationLevel] = useState<number>(1);
  const [selectedVocationClass, setSelectedVocationClass] = useState<"Novice" | "Knight" | "Mage" | "Archer" | "Assassin" | "Cleric">("Novice");
  const [vocationConsoleLog, setVocationConsoleLog] = useState<string>("Novice inicializado. Suba o nível do herói até 10 para escolher uma classe definitiva!");
  const [vocationEquippedWeapon, setVocationEquippedWeapon] = useState<string>("Espada Básica [T1]");

  // Spell & Skill Simulator State
  const [spellHp, setSpellHp] = useState<number>(120);
  const [spellMana, setSpellMana] = useState<number>(40);
  const [skillSwordProficiency, setSkillSwordProficiency] = useState<number>(5);
  const [skillMagicProficiency, setSkillMagicProficiency] = useState<number>(5);
  const [skillHealingProficiency, setSkillHealingProficiency] = useState<number>(5);
  const [spellSkillConsoleLog, setSpellSkillConsoleLog] = useState<string>("Inicie o combate automático ou conjure magias instantâneas.");
  const [spellCooldowns, setSpellCooldowns] = useState<Record<string, number>>({});
  const [scalingEquipMultiplier, setScalingEquipMultiplier] = useState<number>(1.0);
  const [scalingElementalMod, setScalingElementalMod] = useState<number>(1.0);

  // World Content Simulator State
  const [selectedBiome, setSelectedBiome] = useState<string>("barren_desert");
  const [encounterDensity, setEncounterDensity] = useState<number>(75);
  const [activeRiskTier, setActiveRiskTier] = useState<"Minimal" | "Moderate" | "Severe" | "Extreme">("Moderate");
  const [worldContentLogs, setWorldContentLogs] = useState<string[]>([
    "[Inicialização] Sistema Canônico de Geração de Conteúdo carregado.",
    "[Ambiente] Escolha um bioma e clique em 'Simular Spawn Dinâmico' ou 'Desencadear Evento Global'."
  ]);

  // World Scale Bible Simulator State
  const [selectedMountMultiplier, setSelectedMountMultiplier] = useState<number>(1.0); // 1.0, 1.25, 1.40
  const [selectedScaleContinent, setSelectedScaleContinent] = useState<string>("main_continent");




  
  const [selectedActivity, setSelectedActivity] = useState<"HUNT" | "EXPLORATION" | "QUEST" | "BOSS">("HUNT");
  const [questTab, setQuestTab] = useState<"QUESTS" | "CONTRACTS">("QUESTS");
  const [activeQuestId, setActiveQuestId] = useState<string>("main_1");
  const [selectedCraftItem, setSelectedCraftItem] = useState<string>("weapon_t1");
  const [lootFilterType, setLootFilterType] = useState<string>("all");
  const [lootFilterName, setLootFilterName] = useState<string>("");
  const [lootFilterMinVal, setLootFilterMinVal] = useState<number>(0);
  const [inventoryCapacity, setInventoryCapacity] = useState<number>(10); // current slots filled vs max 12
  const [monsterCorpseLoot, setMonsterCorpseLoot] = useState<{ id: string; name: string; qty: number; value: number; type: string }[]>([]);
  const [monsterCorpseTimer, setMonsterCorpseTimer] = useState<number>(0); // countdown in seconds
  const [playerCorpseLoot, setPlayerCorpseLoot] = useState<{ id: string; name: string; qty: number; value: number; type: string }[]>([]);
  const [playerCorpseEquipped, setPlayerCorpseEquipped] = useState<{ slot: string; id: string; name: string; value: number }[]>([]);
  const [playerCorpseExists, setPlayerCorpseExists] = useState<boolean>(false);
  const [bossEligibleSecs, setBossEligibleSecs] = useState<number>(35); // default passes 30s threshold
  const [bossEligibleContr, setBossEligibleContr] = useState<number>(1.5); // default passes 1% threshold
  const [safeZoneChestLoot, setSafeZoneChestLoot] = useState<{ id: string; name: string; qty: number }[]>([]);
  const [safeZoneChestSpawned, setSafeZoneChestSpawned] = useState<boolean>(false);
  const [bibleLogs, setBibleLogs] = useState<{ type: "success" | "error" | "info" | "warning"; text: string }[]>([
    { type: "info", text: "Registros autoritativos do Loot System Bible e Death Penalty Bible carregados com sucesso!" }
  ]);

  const addBibleLog = (type: "success" | "error" | "info" | "warning", text: string) => {
    setBibleLogs(prev => [{ type, text }, ...prev.slice(0, 49)]);
  };



  // Timers Effect
  useEffect(() => {
    const timer = setInterval(() => {
      // Dungeon timer
      if (dungeonActive) {
        setDungeonTimeLeft(prev => {
          if (prev <= 1) {
            setDungeonActive(false);
            addSimLogs([{ type: "error", text: "[DungeonManager] Instância expirou! Jogadores foram teleportados de volta para fora." }]);
            return 1800;
          }
          return prev - 1;
        });
      }

      // Boss Combat timer (Enrage Timer - PATCH 5)
      if (dungeonActive && bossHp > 0 && bossInCombat) {
        setBossCombatSeconds(prev => {
          const nextSec = prev + 1;
          if (nextSec === 30 && !bossEnraged) { // Enrage mais rápido para visualização em 30 segundos
            setBossEnraged(true);
            addSimLogs([{ type: "warning", text: `[BossAI] ${dungeonId === "crypt_of_shadows" ? "Lorde das Sombras" : "Rei Dragão"} ENFURECEU! Dano aumentado em 500%! Pulso periódico de dano ativo.` }]);
          }
          return nextSec;
        });
      }

      // Boss Reset Exploit Timer (PATCH 2)
      if (dungeonActive && bossHp > 0 && !bossInCombat && bossHp < bossMaxHp) {
        setBossResetTimer(prev => {
          if (prev === null) return 10;
          if (prev <= 1) {
            // Reset boss
            setBossHp(bossMaxHp);
            setBossPhase(1);
            setBossEnraged(false);
            setBossCombatSeconds(0);
            setBossAdds([]);
            addSimLogs([{ type: "info", text: "[BossAI] Anti-reset timer esgotou. Boss retornou para a vida cheia e resetou fase." }]);
            return null;
          }
          return prev - 1;
        });
      } else {
        setBossResetTimer(null);
      }

      // Loot reservation timer (PATCH 3: Loot idempotency)
      setLootReservations(prev => {
        return prev.map(res => {
          if (res.timer > 0 && !res.claimed) {
            return { ...res, timer: res.timer - 1 };
          }
          return res;
        }).filter(res => res.timer > 0 || res.claimed);
      });

      // Monster corpse countdown (Loot System Bible: Corpse Persistence 12-minute rule)
      setMonsterCorpseTimer(prev => {
        if (prev > 1) {
          return prev - 1;
        } else if (prev === 1) {
          setMonsterCorpseLoot([]);
          return 0;
        }
        return 0;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [dungeonActive, bossHp, bossInCombat, bossEnraged, dungeonId, bossMaxHp]);

  // Spell & Skill Cooldown Decrementor
  useEffect(() => {
    const activeCooldownsExist = Object.keys(spellCooldowns).some(key => {
      const val = spellCooldowns[key];
      return typeof val === "number" && val > 0;
    });
    if (!activeCooldownsExist) return;

    const timer = setInterval(() => {
      setSpellCooldowns(prev => {
        const next = { ...prev };
        let changed = false;
        for (const key of Object.keys(next)) {
          const val = next[key];
          if (typeof val === "number" && val > 0) {
            next[key] = Math.max(0, parseFloat((val - 0.1).toFixed(1)));
            changed = true;
          }
        }
        return changed ? next : prev;
      });
    }, 100);

    return () => clearInterval(timer);
  }, [spellCooldowns]);

  // Boss telegraph timer tick (runs faster for precision)
  useEffect(() => {
    if (!bossTelegraphActive) return;
    const interval = setInterval(() => {
      setBossTelegraphTimer(prev => {
        if (prev <= 1) {
          setBossTelegraphActive(false);
          // Aplica dano do telegrafo
          const dmg = bossEnraged ? 800 : 180;
          addSimLogs([
            { type: "info", text: `[Network] Pacote Recebido: SC_DAMAGE_EVENT (Opcode: 3004) - Alvo atingido por Terremoto Sombrio.` },
            { type: "error", text: `[BossAI] Impacto de ataque em área telegrafado! Você sofreu ${dmg} de dano!` }
          ]);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [bossTelegraphActive, bossEnraged]);

  // Handlers de Ações PvE
  const handleEnterDungeonSim = () => {
    // Simulação do handshake oficial Little Endian por rede
    const byteSize = 4 + simUsername.length + dungeonId.length + dungeonMode.length;
    addSimLogs([
      { type: "info", text: `[Network] Enviando Pacote: CS_DUNGEON_ENTER (Opcode: 9000), Sequence: ${Math.floor(Math.random() * 100)}, Payload: ${byteSize} bytes (${dungeonId}, ${dungeonMode})` },
      { type: "info", text: `[DungeonManager] Handshake de isolamento espacial ativo. Criando instância com offset coordenado.` }
    ]);

    // Resgata se há checkpoints
    const existingCp = persistentLogs.find(l => l.type === "Checkpoint" && l.details.includes(dungeonId));
    const cp = existingCp ? existingCp.details.split("checkpoint: ")[1] || "Entrance" : "Entrance";

    // Set Max HP de acordo com modo
    let baseMax = dungeonId === "crypt_of_shadows" ? 5000 : 12000;
    if (dungeonMode === "party") baseMax *= 3;
    if (dungeonMode === "raid") baseMax *= 8;

    setTimeout(() => {
      setDungeonActive(true);
      setDungeonTimeLeft(dungeonId === "crypt_of_shadows" ? 1800 : 3600);
      setDungeonCheckpoint(cp);
      setDungeonOffset(Math.floor(Math.random() * 5) * 2000.0);
      setBossHp(baseMax);
      setBossMaxHp(baseMax);
      setBossPhase(1);
      setBossEnraged(false);
      setBossCombatSeconds(0);
      setBossInCombat(false);
      setBossAdds([]);
      setBossTelegraphActive(false);

      addSimLogs([
        { type: "info", text: `[Network] Pacote Recebido: SC_DUNGEON_STATE (Opcode: 9001), Payload: inst_${dungeonId}_1, checkpoint: ${cp}, time: ${dungeonId === "crypt_of_shadows" ? 1800 : 3600}s, bossAlive: 1` },
        { type: "info", text: `[Client] Teleportado para coordenadas isoladas de masmorra (Z-axis isolado por offset posicional).` }
      ]);
    }, 500);
  };

  // --- BESTIARY SIMULATOR HANDLERS (CANONICAL BESTIARY BIBLE PATCH) ---
  const handleSimulateSpawn = () => {
    let respawnTimeText = "";
    if (spawnCategory === "Standard Spawn") {
      const baseTimer = 30 + Math.floor(Math.random() * 270); // 30s to 5m
      respawnTimeText = `[Spawn Engine] Categoria: Standard Spawn. Spawn Fixo ativado. Tempo de respawn estimado: ${baseTimer} segundos. (Regra: Goblins, Wolves, Skeletons, Spiders utilizam tempo estático)`;
    } else if (spawnCategory === "Dynamic Ecosystem Spawn") {
      // Depende de player density, hunting pressure, alive population
      const densityModifier = Math.max(0.2, 1 - playerDensity / 50); // mais players = respawn mais rápido
      const pressureModifier = huntingPressure === "high" ? 0.5 : huntingPressure === "medium" ? 1.0 : 1.5;
      const baseTime = 120 * densityModifier * pressureModifier; // base 2 minutos
      respawnTimeText = `[Spawn Engine] Categoria: Dynamic Ecosystem. Fatores calculados: Jogadores=${playerDensity}, Pressão de Caça=${huntingPressure}. Tempo de respawn recalculado dinamicamente: ${Math.round(baseTime)} segundos.`;
    } else if (spawnCategory === "Rare Spawn") {
      const baseMinutes = 30 + Math.floor(Math.random() * 1410); // 30m to 24h
      respawnTimeText = `[Spawn Engine] Categoria: Rare Spawn. Spawn semi-aleatório com múltiplas localizações e baixa frequência. Tempo de janela estimado: ${baseMinutes} minutos.`;
    } else {
      respawnTimeText = `[Spawn Engine] Categoria: Legendary Spawn. Respawn estático desabilitado. Gatilhos ativos: Eventos mundiais, condições ocultas, progressão de missões.`;
    }
    setSpawnSimResult(respawnTimeText);
    addSimLogs([{ type: "info", text: respawnTimeText }]);
  };

  const handleSimulateDemonSummon = () => {
    let log = "";
    // Verificação de Região do Continente Principal
    if (demonRegion === "Main Continent") {
      log = `❌ [Invocação Proibida] De acordo com a Seção 4 da Lei Demôniaca, "Natural demon spawning cannot occur in the Main Continent." Invocação falhou por falta de influência de Primordiais naturais. Permissão apenas via Invasão de Grande Porte ou Cataclismo Narrativo.`;
      setDemonSummonLog(log);
      addSimLogs([{ type: "error", text: log }]);
      return;
    }

    // Regra Hierárquica: "Higher demons may manifest lower demons. Lower demons cannot summon higher hierarchy."
    const hierarchyLevel = (h: string) => {
      if (h === "Common Demons") return 1;
      if (h === "Minor Demons") return 2;
      if (h === "Commanders") return 3;
      if (h === "Archegenerals") return 4;
      return 0;
    };

    const sLvl = hierarchyLevel(demonSummoner);
    const tLvl = hierarchyLevel(demonSummonTarget);

    if (sLvl > tLvl) {
      log = `✅ [Invocação Sucesso] O invocador (${demonSummoner}, Nível ${sLvl}) manifestou com sucesso o demônio de hierarquia inferior (${demonSummonTarget}, Nível ${tLvl}). A autoridade hierárquica flui de cima para baixo.`;
      addSimLogs([{ type: "info", text: log }]);
    } else {
      log = `❌ [Invocação Negada] Erro Hierárquico: Invocador (${demonSummoner}, Nível ${sLvl}) não possui autoridade para conjurar demônios de hierarquia igual ou superior (${demonSummonTarget}, Nível ${tLvl}).`;
      addSimLogs([{ type: "warning", text: log }]);
    }

    setDemonSummonLog(log);
  };

  const handleSimulateDragonTerritory = () => {
    let log = "";
    // Territorial Law: "Two apex dragons rarely tolerate the same territory."
    const isSovereign = (d: string) => d === "void_dragon_lord";
    const isLesser = (d: string) => d === "wyvern";

    if (dragonOne === dragonTwo) {
      log = `ℹ️ Dragão Único: O espécime está patrulhando o território de lendas em ${dragonTerritory} de forma soberana.`;
    } else if (isSovereign(dragonOne) && isSovereign(dragonTwo)) {
      log = `💥 [Cataclismo Territorial] Duas entidades Soberanas colidiram em ${dragonTerritory}! Consequência da Lei Territorial de Apex: "Dragon Lords avoid coexistence". O choque abalou a escala do servidor inteiro!`;
    } else if (isLesser(dragonOne) || isLesser(dragonTwo)) {
      const boss = isLesser(dragonOne) ? dragonTwo : dragonOne;
      const minion = isLesser(dragonOne) ? dragonOne : dragonTwo;
      log = `✅ [Coexistência Tolerada] O dragão menor (${minion}) aceitou a soberania do dragão apex dominador (${boss}) no território ${dragonTerritory}. Subordinação confirmada pela Seção 5 da Lei de Territorialidade.`;
    } else {
      log = `⚡ [Duelo Territorial] Combate de Titãs iniciado em ${dragonTerritory}! Os dragões apex recusam coexistência territorial. O combate resolverá por dominação ou afastamento de um deles.`;
    }

    setDragonCombatLog(log);
    addSimLogs([{ type: "info", text: log }]);
  };

  const handleLeaveDungeonSim = () => {
    addSimLogs([
      { type: "info", text: `[Network] Enviando Pacote: CS_DUNGEON_LEAVE (Opcode: 9006). Removendo jogador da lista de instâncias.` },
      { type: "info", text: `[Client] Teleportado de volta ao ponto inicial do mundo.` }
    ]);
    setDungeonActive(false);
    setBossInCombat(false);
  };

  const handleAttackBossSim = () => {
    if (bossHp <= 0) return;
    setBossInCombat(true);
    setBossResetTimer(null);

    const dmg = Math.floor(Math.random() * 250) + 300;
    const isCrit = Math.random() < 0.25;
    const finalDmg = isCrit ? Math.floor(dmg * 1.8) : dmg;

    // Se houver Adds vivos, ataca o add primeiro para cumprir a mecânica
    if (bossAdds.length > 0) {
      const targetAdd = bossAdds[0];
      const nextAddHp = Math.max(0, targetAdd.hp - finalDmg);
      addSimLogs([
        { type: "info", text: `[Combat] Você atacou Servo de IA do Boss. Dano causado: ${finalDmg} ${isCrit ? "(CRÍTICO!)" : ""}` }
      ]);
      if (nextAddHp <= 0) {
        setBossAdds(prev => prev.slice(1));
        addSimLogs([{ type: "info", text: "[Combat] Servo de IA foi derrotado de forma autoritativa!" }]);
      } else {
        setBossAdds(prev => [{ ...prev[0], hp: nextAddHp }, ...prev.slice(1)]);
      }
      return;
    }

    const nextHp = Math.max(0, bossHp - finalDmg);
    setBossHp(nextHp);

    addSimLogs([
      { type: "info", text: `[Combat] Você atacou o Boss Principal. Dano causado: ${finalDmg} ${isCrit ? "(CRÍTICO!)" : ""}` }
    ]);

    // Phase transitions (PATCH AI / Telegraphed AoE)
    const currentPct = nextHp / bossMaxHp;
    if (bossPhase === 1 && currentPct <= 0.70) {
      setBossPhase(2);
      // Invoca Adds
      setBossAdds([
        { id: "add_1", hp: 1000, maxHp: 1000 },
        { id: "add_2", hp: 1000, maxHp: 1000 }
      ]);
      addSimLogs([
        { type: "warning", text: "[BossAI] Transição de Fase! Boss entrou na Fase 2. Invocou 2 Servos de IA para protegê-lo!" },
        { type: "info", text: "[Network] Pacote Recebido: SC_BOSS_AI_PHASE (Opcode: 9003), Phase: 2" }
      ]);
    } else if (bossPhase === 2 && currentPct <= 0.35) {
      setBossPhase(3);
      // Invoca mais Adds e inicia Telegraph
      setBossAdds([
        { id: "add_3", hp: 1500, maxHp: 1500 }
      ]);
      setBossTelegraphActive(true);
      setBossTelegraphTimer(3);
      addSimLogs([
        { type: "warning", text: "[BossAI] Transição de Fase! Boss entrou na Fase 3. Iniciando canalização de ataque devastador telegrafado!" },
        { type: "info", text: "[Network] Pacote Recebido: SC_BOSS_AI_TELEGRAPH (Opcode: 9002), delay: 3s, radius: 8m" }
      ]);
    }

    // Boss derrotado de forma autoritativa
    if (nextHp <= 0) {
      setBossInCombat(false);
      setBossAdds([]);
      setBossTelegraphActive(false);

      // Persiste LOCKOUT, CHECKPOINT e KILL STATE no PostgreSQL simulado (PATCH 6)
      const nowStr = new Date().toTimeString().split(" ")[0];
      const cpNext = dungeonId === "crypt_of_shadows" ? "Throne Room" : "Completed";
      
      const newLogs = [
        { id: Math.random().toString(), type: "Lockout", details: `Raid Lockout salvo: Gabriela_Paladin bloqueado temporariamente para o boss: shadow_lord`, timestamp: nowStr },
        { id: Math.random().toString(), type: "Checkpoint", details: `Progresso de masmorra salvo para ${dungeonId}: checkpoint: ${cpNext}`, timestamp: nowStr },
        { id: Math.random().toString(), type: "Kill State", details: `Estado de abate salvo: ${dungeonId === "crypt_of_shadows" ? "Lorde das Sombras" : "Rei Dragão"} derrotado`, timestamp: nowStr },
        { id: Math.random().toString(), type: "Audit Log", details: `Log de Auditoria: Gabriela_Paladin contribuiu com 100% (Dano: ${bossMaxHp}) no boss ${dungeonId}. Item recompensado: epic_shadow_shard`, timestamp: nowStr }
      ];

      setPersistentLogs(prev => [...prev, ...newLogs]);

      // Reserva de Loot Idempotente com expiração (PATCH 3)
      setLootReservations([
        { id: "loot_1", itemID: "epic_shadow_shard", qty: 2, timer: 30, claimed: false },
        { id: "loot_2", itemID: "legendary_void_essence", qty: 1, timer: 30, claimed: false }
      ]);

      addSimLogs([
        { type: "info", text: "[DungeonManager] Boss principal derrotado de forma autoritativa no servidor!" },
        { type: "info", text: "[Network] Pacote Recebido: SC_LOOT_NOTIFICATION (Opcode: 9005). Reserva de loot disponível por 30s." }
      ]);
    }
  };

  const handleClaimLootSim = (id: string, itemID: string, qty: number) => {
    // Envia pacote de resgate idempotente
    addSimLogs([
      { type: "info", text: `[Network] Enviando Pacote: CS_DUNGEON_CLAIM_LOOT (Opcode: 9004), itemID: ${itemID}, qty: ${qty}` },
      { type: "info", text: `[DungeonManager] Validando resgate com auditoria contra exploits (Loot Idempotency).` }
    ]);

    setTimeout(() => {
      setLootReservations(prev => prev.map(res => res.id === id ? { ...res, claimed: true } : res));
      addSimLogs([
        { type: "info", text: `[Network] Pacote Recebido: SC_INVENTORY_SYNC (Opcode: 3005). Sincronizado: +${qty}x ${itemID}` },
        { type: "info", text: `[Client] Adicionado ao inventário do jogador com integridade transacional no PostgreSQL.` }
      ]);
    }, 400);
  };

  // Loot and Death Penalty Bible simulation handlers
  const handleSimulateLootDrop = (monsterType: string) => {
    addBibleLog("info", `[LootEngine] Rolando loot para abate de: ${monsterType.replace("_", " ")}`);
    
    // Drop frequency rule: 50-60% average. Elite/Minor Demon: 70-90%.
    const isElite = ["minor_demon", "commander_demon", "archegeneral_demon"].includes(monsterType);
    const dropChance = isElite ? 0.80 : 0.55;
    const rolled = Math.random();
    
    if (rolled > dropChance) {
      addBibleLog("warning", `[LootEngine] Sem itens sorteados (Sorteio: ${(rolled * 100).toFixed(0)}% > Chance: ${dropChance * 100}%). Sparse but Meaningful loot.`);
      return;
    }

    // Determine drops
    const potentialDrops: { id: string; name: string; qty: number; value: number; type: string }[] = [];
    if (monsterType === "common_demon") {
      potentialDrops.push({ id: "demonic_ash", name: "Cinza Demoníaca", qty: Math.floor(Math.random() * 3) + 1, value: 12, type: "material" });
      if (Math.random() < 0.5) potentialDrops.push({ id: "corrupted_blood", name: "Sangue Corrompido", qty: 1, value: 25, type: "material" });
      if (Math.random() < 0.2) potentialDrops.push({ id: "low_demonic_essence", name: "Essência Demoníaca Inferior", qty: 1, value: 80, type: "material" });
    } else if (monsterType === "minor_demon") {
      potentialDrops.push({ id: "demon_core_low", name: "Núcleo Demoníaco Inferior", qty: 1, value: 150, type: "material" });
      potentialDrops.push({ id: "infernal_fragment", name: "Fragmento Infernal", qty: Math.floor(Math.random() * 2) + 1, value: 60, type: "material" });
    } else if (monsterType === "commander_demon") {
      potentialDrops.push({ id: "high_demon_core", name: "Núcleo Demoníaco Superior", qty: 1, value: 600, type: "material" });
      if (Math.random() < 0.6) potentialDrops.push({ id: "primordial_shard", name: "Fragmento Primordial", qty: 1, value: 250, type: "material" });
    } else if (monsterType === "archegeneral_demon") {
      potentialDrops.push({ id: "primordial_relic", name: "Relíquia Primordial", qty: 1, value: 2000, type: "material" });
      if (Math.random() < 0.4) potentialDrops.push({ id: "abyssal_artifact", name: "Artefato Abissal", qty: 1, value: 5000, type: "material" });
    } else if (monsterType === "civilized_cultist") {
      // Civilized carrying coins directly
      potentialDrops.push({ id: "silver_coin", name: "Moeda de Prata", qty: Math.floor(Math.random() * 4) + 1, value: 100, type: "currency" });
      potentialDrops.push({ id: "bronze_coin", name: "Moeda de Bronze", qty: Math.floor(Math.random() * 50) + 10, value: 1, type: "currency" });
      if (Math.random() < 0.1) potentialDrops.push({ id: "potion_heal", name: "Poção de Cura", qty: 1, value: 15, type: "consumable" });
    } else {
      // Beast (biological creature rarely carries currency)
      potentialDrops.push({ id: "raw_hide", name: "Couro Bruto", qty: 1, value: 8, type: "material" });
      if (Math.random() < 0.3) potentialDrops.push({ id: "beast_claw", name: "Garra de Fera", qty: 2, value: 12, type: "material" });
    }

    // Filter validation & auto-loot or overflow
    const autoLooted: string[] = [];
    const overflowed: typeof potentialDrops = [];

    potentialDrops.forEach(item => {
      // Custom filter logic
      const matchesType = lootFilterType === "all" || item.type === lootFilterType;
      const matchesName = lootFilterName === "" || item.name.toLowerCase().includes(lootFilterName.toLowerCase());
      const matchesValue = item.value >= lootFilterMinVal;

      if (!matchesType || !matchesName || !matchesValue) {
        addBibleLog("info", `[Filtro] Item ${item.name} ignorado pelo filtro personalizado.`);
        return;
      }

      // Inventory capacity check
      const isFull = playerInventory.length >= 12 && !playerInventory.some(i => i.id === item.id);
      if (!isFull) {
        // Direct inventory auto-loot
        setPlayerInventory(prev => {
          const idx = prev.findIndex(i => i.id === item.id);
          if (idx > -1) {
            const updated = [...prev];
            updated[idx].qty += item.qty;
            return updated;
          }
          return [...prev, item];
        });
        autoLooted.push(`+${item.qty}x ${item.name}`);
      } else {
        // Inventory Overflow Law -> remains in corpse!
        overflowed.push(item);
      }
    });

    if (autoLooted.length > 0) {
      addBibleLog("success", `[AutoLoot] Coletado diretamente: ${autoLooted.join(", ")}`);
    }

    if (overflowed.length > 0) {
      setMonsterCorpseLoot(prev => [...prev, ...overflowed]);
      setMonsterCorpseTimer(720); // 12 minutes persistence default
      addBibleLog("error", `[Overflow] Inventário Cheio! ${overflowed.map(i => `${i.qty}x ${i.name}`).join(", ")} permanecem no cadáver do monstro (Persistência: 12 min).`);
    }
  };

  const handleLootMonsterCorpse = () => {
    if (monsterCorpseLoot.length === 0) return;
    
    const stillOverflowed: typeof monsterCorpseLoot = [];
    const looted: string[] = [];

    monsterCorpseLoot.forEach(item => {
      const isFull = playerInventory.length >= 12 && !playerInventory.some(i => i.id === item.id);
      if (!isFull) {
        setPlayerInventory(prev => {
          const idx = prev.findIndex(i => i.id === item.id);
          if (idx > -1) {
            const updated = [...prev];
            updated[idx].qty += item.qty;
            return updated;
          }
          return [...prev, item];
        });
        looted.push(`+${item.qty}x ${item.name}`);
      } else {
        stillOverflowed.push(item);
      }
    });

    setMonsterCorpseLoot(stillOverflowed);
    if (stillOverflowed.length === 0) {
      setMonsterCorpseTimer(0);
    }

    if (looted.length > 0) {
      addBibleLog("success", `[CorpseLoot] Você saqueou manualmente do cadáver: ${looted.join(", ")}`);
    } else {
      addBibleLog("error", "[CorpseLoot] Falha ao saquear: seu inventário continua completamente cheio!");
    }
  };

  const handleSimulatePlayerDeath = (isPvP = false) => {
    const originLabel = isPvP ? "Morte PvP" : "Morte PvE";
    const causeText = isPvP 
      ? "O jogador foi derrotado por outro jogador em combate de mundo aberto!" 
      : "O jogador foi derrotado por uma criatura hostil no mundo aberto!";
    
    addBibleLog("warning", `[${originLabel}] ${causeText}`);
    
    const lostEquipped: typeof playerEquipped = [];
    const keptEquipped: typeof playerEquipped = [];
    
    const lostInventory: typeof playerInventory = [];
    const keptInventory: typeof playerInventory = [];
 
    // Canonical linear formula: Final Penalty = Base Penalty * (1 - Active Blessings / 7)
    const mitigationMultiplier = 1 - (activeBlessingsCount / 7);
    const dropRateEquipped = 0.50 * mitigationMultiplier;
    const dropRateInventory = 0.80 * mitigationMultiplier;
 
    playerEquipped.forEach(item => {
      if (Math.random() < dropRateEquipped) {
        lostEquipped.push(item);
      } else {
        keptEquipped.push(item);
      }
    });
 
    playerInventory.forEach(item => {
      if (Math.random() < dropRateInventory) {
        lostInventory.push(item);
      } else {
        keptInventory.push(item);
      }
    });
 
    setPlayerEquipped(keptEquipped);
    setPlayerInventory(keptInventory);
 
    const blessingsConsumed = activeBlessingsCount;
    // Blessing Consumption Law: Both PvP and PvE death removes ALL active blessings simultaneously
    if (activeBlessingsCount > 0) {
      setActiveBlessingsCount(0);
      addBibleLog("warning", `[Leis de Morte] Consumido: Todas as ${blessingsConsumed} bênçãos ativas foram purgadas simultaneamente!`);
    }

    // Carried Currency Risk Law & Bank Security Law integration
    const coinLossRatio = 0.50 * mitigationMultiplier; // Lose up to 50% of carried coins under full risk
    const coinsLost = Math.round(carriedCoins * coinLossRatio);
    if (coinsLost > 0) {
      setCarriedCoins(prev => prev - coinsLost);
      addBibleLog("error", `[${originLabel}] Moedas Perdidas: ${coinsLost} Moedas de Ouro (da bolsa de carried_currency) caíram no seu corpo (carried-currency-risk-rule).`);
    } else {
      addBibleLog("info", `[${originLabel}] Moedas Protegidas: Nenhuma moeda de sua bolsa foi perdida devido à proteção de bênção.`);
    }
    addBibleLog("success", `[Banco de Ouro] Suas ${bankedCoins} Moedas de Ouro no Banco de Ravenshire permanecem 100% Intactas e Protegidas (bank-security-rule).`);
 
    if (lostEquipped.length > 0 || lostInventory.length > 0) {
      setPlayerCorpseLoot(lostInventory);
      setPlayerCorpseEquipped(lostEquipped);
      setPlayerCorpseExists(true);
      
      addBibleLog("error", `[${originLabel}] Itens Perdidos: ${[
        ...lostEquipped.map(e => `${e.name} (${e.slot})`),
        ...lostInventory.map(i => `${i.qty}x ${i.name}`)
      ].join(", ")} dropped on player corpse. Open Loot enabled!`);
      
      const protectionPct = ((1 - mitigationMultiplier) * 100).toFixed(3);
      addBibleLog("info", `[Mitigação] Mitigação de Morte aplicada: ${protectionPct}% de proteção linear com ${blessingsConsumed} Bênçãos.`);
    } else {
      const protectionPct = ((1 - mitigationMultiplier) * 100).toFixed(3);
      addBibleLog("success", `[${originLabel}] Sem perda de itens! Proteção de ${protectionPct}% com ${blessingsConsumed} Bênçãos garantiu segurança total.`);
    }
  };

  const handleRetrievePlayerCorpse = () => {
    if (!playerCorpseExists) return;

    const remainingLoot: typeof playerCorpseLoot = [];
    const lootedItems: string[] = [];

    playerCorpseLoot.forEach(item => {
      const isFull = playerInventory.length >= 12 && !playerInventory.some(i => i.id === item.id);
      if (!isFull) {
        setPlayerInventory(prev => {
          const idx = prev.findIndex(i => i.id === item.id);
          if (idx > -1) {
            const updated = [...prev];
            updated[idx].qty += item.qty;
            return updated;
          }
          return [...prev, item];
        });
        lootedItems.push(`${item.qty}x ${item.name}`);
      } else {
        remainingLoot.push(item);
      }
    });

    const remainingEquipped: typeof playerCorpseEquipped = [];
    playerCorpseEquipped.forEach(item => {
      const isSlotTaken = playerEquipped.some(e => e.slot === item.slot);
      if (!isSlotTaken) {
        setPlayerEquipped(prev => [...prev, item]);
        lootedItems.push(`${item.name} (Reequipado)`);
      } else {
        const isFull = playerInventory.length >= 12 && !playerInventory.some(i => i.id === item.id);
        if (!isFull) {
          setPlayerInventory(prev => [
            ...prev,
            { id: item.id, name: item.name, qty: 1, value: item.value, type: "gear" }
          ]);
          lootedItems.push(`${item.name} (Movido para Inventário)`);
        } else {
          remainingEquipped.push(item);
        }
      }
    });

    setPlayerCorpseLoot(remainingLoot);
    setPlayerCorpseEquipped(remainingEquipped);

    if (remainingLoot.length === 0 && remainingEquipped.length === 0) {
      setPlayerCorpseExists(false);
      addBibleLog("success", `[CorpseRecover] Corpo limpo! Todos os pertences foram recuperados: ${lootedItems.join(", ")}`);
    } else {
      addBibleLog("warning", `[CorpseRecover] Recuperação parcial devido a inventário cheio: ${lootedItems.join(", ")}`);
    }
  };

  const handleOpportunisticLootPlayerCorpse = () => {
    if (!playerCorpseExists) return;
    setPlayerCorpseLoot([]);
    setPlayerCorpseEquipped([]);
    setPlayerCorpseExists(false);
    addBibleLog("error", "[OpenLoot] Emergent Gameplay! Um ladrão oportunista passou por seu corpo e levou todos os itens caídos!");
  };

  const handleSimulateBossProtectedChest = () => {
    addBibleLog("info", "[BossReward] Validando elegibilidade para recompensas de chefe de masmorra...");
    
    // Hybrid Eligibility: Participation duration (>=30s) + Contribution threshold (>=1.0%)
    const durationPassed = bossEligibleSecs >= 30;
    const contributionPassed = bossEligibleContr >= 1.0;

    if (!durationPassed || !contributionPassed) {
      addBibleLog("error", `[AntiExploit] Elegibilidade Rejeitada! Tempo de combate: ${bossEligibleSecs}s/30s, Contribuição: ${bossEligibleContr}%/1.0%.`);
      return;
    }

    // Pass: Personal loot delivery through protected reward chests in safe zones
    const drops = [
      { id: "epic_shadow_shard", name: "Fragmento de Sombra Épico", qty: 2 },
      { id: "legendary_void_essence", name: "Essência do Vazio Lendária", qty: 1 }
    ];

    setSafeZoneChestLoot(drops);
    setSafeZoneChestSpawned(true);
    
    addBibleLog("success", `[BossReward] Elegibilidade Aprovada! Saque pessoal gerado de forma segura e depositado no Baú Protegido (Zona de Segurança).`);
    addBibleLog("info", "[ProtectedBossChestLaw] Conforme as leis canônicas, recompensas pessoais nunca caem diretamente na zona de combate.");
  };

  const handleClaimSafeZoneChest = () => {
    if (safeZoneChestLoot.length === 0) return;

    const remainingChest: typeof safeZoneChestLoot = [];
    const looted: string[] = [];

    safeZoneChestLoot.forEach(item => {
      const isFull = playerInventory.length >= 12 && !playerInventory.some(i => i.id === item.id);
      if (!isFull) {
        setPlayerInventory(prev => {
          const idx = prev.findIndex(i => i.id === item.id);
          if (idx > -1) {
            const updated = [...prev];
            updated[idx].qty += item.qty;
            return updated;
          }
          return [...prev, { id: item.id, name: item.name, qty: item.qty, value: 500, type: "material" }];
        });
        looted.push(`${item.qty}x ${item.name}`);
      } else {
        remainingChest.push(item);
      }
    });

    setSafeZoneChestLoot(remainingChest);
    if (remainingChest.length === 0) {
      setSafeZoneChestSpawned(false);
      addBibleLog("success", `[ProtectedChest] Baú Protegido esvaziado com segurança: ${looted.join(", ")}`);
    } else {
      addBibleLog("warning", `[ProtectedChest] Coleta parcial devido a inventário cheio: ${looted.join(", ")}`);
    }
  };

  const handleSpawnWorldBoss = () => {
    // Força spawn manual
    addSimLogs([
      { type: "info", text: `[Network] Enviando Pacote: CS_WORLD_BOSS_SPAWN_REQ (Opcode: 9007)` },
      { type: "warning", text: "[DungeonManager] Spawn de World Boss forçado! Behemoth surgiu no mundo público!" }
    ]);
    setWorldBossActive(true);
    setWorldBossHp(50000);
    setWorldBossThreat({
      "Gabriela_Paladin": 0,
      "AI_Bot_Mage": 0,
      "AI_Bot_Priest": 0
    });
  };

  const handleAttackWorldBoss = () => {
    if (worldBossHp <= 0) return;
    const dmg = Math.floor(Math.random() * 800) + 500;
    const nextHp = Math.max(0, worldBossHp - dmg);
    setWorldBossHp(nextHp);

    // Incrementa threat
    setWorldBossThreat(prev => ({
      ...prev,
      "Gabriela_Paladin": (prev["Gabriela_Paladin"] || 0) + dmg,
      "AI_Bot_Mage": (prev["AI_Bot_Mage"] || 0) + Math.floor(Math.random() * 400),
      "AI_Bot_Priest": (prev["AI_Bot_Priest"] || 0) + Math.floor(Math.random() * 300)
    }));

    addSimLogs([
      { type: "info", text: `[Combat] Você atacou o World Boss Behemoth! Dano causado: ${dmg}` }
    ]);

    // Boss derrotado
    if (nextHp <= 0) {
      setWorldBossActive(false);

      // Calcula contribuição rankeada
      const total = worldBossThreat["Gabriela_Paladin"] + worldBossThreat["AI_Bot_Mage"] + worldBossThreat["AI_Bot_Priest"];
      const userPct = ((worldBossThreat["Gabriela_Paladin"] + dmg) / total) * 100;

      const nowStr = new Date().toTimeString().split(" ")[0];
      const auditLog = {
        id: Math.random().toString(),
        type: "Audit Log",
        details: `Log de Auditoria: Gabriela_Paladin contribuiu com ${userPct.toFixed(1)}% no World Boss Behemoth. Item recompensado: rare_behemoth_scale`,
        timestamp: nowStr
      };
      setPersistentLogs(prev => [...prev, auditLog]);

      // Spawns Loot Reservation
      setLootReservations([
        { id: "world_loot_1", itemID: "rare_behemoth_scale", qty: 3, timer: 60, claimed: false }
      ]);

      addSimLogs([
        { type: "info", text: "[DungeonManager] World Boss Behemoth derrotado! Calculando contribuição e gerando drops ranqueados." },
        { type: "info", text: `[Network] Pacote Recebido: SC_LOOT_NOTIFICATION (Opcode: 9005). Reserva de loot disponível por 60s.` }
      ]);
    }
  };

  // --- SPRINT 3 TASK 5 HANDLERS ---
  const handleChooseVocationSim = (vocation: string) => {
    if (progLevel < 10) {
      addSimLogs([{ type: "error", text: `[Progression] Erro ao selecionar classe: Nível insuficiente! Requer Nível 10 (Nível atual: ${progLevel}).` }]);
      return;
    }
    if (progClass !== "Novice") {
      addSimLogs([{ type: "error", text: `[Progression] Erro: Seleção de classe já realizada! Classe atual: ${progClass} (Decisão irreversível).` }]);
      return;
    }

    addSimLogs([
      { type: "info", text: `[Network] Enviando pacote binário CS_CHOOSE_VOCATION (Opcode: 9100) pedindo vocação: ${vocation}.` },
      { type: "info", text: `[Progression] Transação atômica concluída no PostgreSQL (SET class = '${vocation}').` },
      { type: "info", text: `[Network] Pacote Recebido: SC_CHOOSE_VOCATION_RESP (Opcode: 9101). Vocação autorizada com sucesso!` }
    ]);

    setProgClass(vocation);

    const nowStr = new Date().toTimeString().split(" ")[0];
    setPersistentLogs(prev => [...prev, {
      id: Math.random().toString(),
      type: "Audit Log",
      details: `Novo Registro: Personagem escolheu irreversivelmente a vocação base: ${vocation}.`,
      timestamp: nowStr
    }]);
  };

  const handleAddAffinitySim = (element: "fire" | "ice" | "holy" | "shadow" | "nature", points: number, actionName: string) => {
    const now = Date.now();
    if (now - lastAffinityTime < 1500) {
      setSpamWarning(true);
      addSimLogs([{ type: "warning", text: `[Spam Shield] BLOQUEADO: Acúmulo de afinidade muito rápido! Rate-limit ativo (mínimo 1.5s entre ações).` }]);
      return;
    }
    
    setSpamWarning(false);
    setLastAffinityTime(now);

    // Points translate to substantial XP gains (e.g., each point is 100 XP)
    const xpGain = points * 100;

    addSimLogs([
      { type: "info", text: `[Progression] Ação '${actionName}' executada. Ganhando +${xpGain} XP de afinidade para o elemento ${element.toUpperCase()}.` }
    ]);

    switch (element) {
      case "fire": {
        setAffinityFireXp(prev => {
          let newXp = prev + xpGain;
          let newLvl = affinityFireLvl;
          while (newXp >= 1000 && newLvl < 100) {
            newXp -= 1000;
            newLvl += 1;
            addSimLogs([{ type: "info", text: `[Progression] LEVEL UP de Afinidade de Fogo! Novo nível: ${newLvl}.` }]);
          }
          if (newLvl === 100) newXp = 0;
          setAffinityFireLvl(newLvl);
          return newXp;
        });
        break;
      }
      case "ice": {
        setAffinityIceXp(prev => {
          let newXp = prev + xpGain;
          let newLvl = affinityIceLvl;
          while (newXp >= 1000 && newLvl < 100) {
            newXp -= 1000;
            newLvl += 1;
            addSimLogs([{ type: "info", text: `[Progression] LEVEL UP de Afinidade de Gelo! Novo nível: ${newLvl}.` }]);
          }
          if (newLvl === 100) newXp = 0;
          setAffinityIceLvl(newLvl);
          return newXp;
        });
        break;
      }
      case "holy": {
        setAffinityHolyXp(prev => {
          let newXp = prev + xpGain;
          let newLvl = affinityHolyLvl;
          while (newXp >= 1000 && newLvl < 100) {
            newXp -= 1000;
            newLvl += 1;
            addSimLogs([{ type: "info", text: `[Progression] LEVEL UP de Afinidade Sagrada! Novo nível: ${newLvl}.` }]);
          }
          if (newLvl === 100) newXp = 0;
          setAffinityHolyLvl(newLvl);
          return newXp;
        });
        break;
      }
      case "shadow": {
        setAffinityShadowXp(prev => {
          let newXp = prev + xpGain;
          let newLvl = affinityShadowLvl;
          while (newXp >= 1000 && newLvl < 100) {
            newXp -= 1000;
            newLvl += 1;
            addSimLogs([{ type: "info", text: `[Progression] LEVEL UP de Afinidade Sombria! Novo nível: ${newLvl}.` }]);
          }
          if (newLvl === 100) newXp = 0;
          setAffinityShadowLvl(newLvl);
          return newXp;
        });
        break;
      }
      case "nature": {
        setAffinityNatureXp(prev => {
          let newXp = prev + xpGain;
          let newLvl = affinityNatureLvl;
          while (newXp >= 1000 && newLvl < 100) {
            newXp -= 1000;
            newLvl += 1;
            addSimLogs([{ type: "info", text: `[Progression] LEVEL UP de Afinidade de Natureza! Novo nível: ${newLvl}.` }]);
          }
          if (newLvl === 100) newXp = 0;
          setAffinityNatureLvl(newLvl);
          return newXp;
        });
        break;
      }
    }
  };

  const handleUnlockSubclassSim = () => {
    if (progLevel < 100) {
      addSimLogs([{ type: "error", text: `[Progression] Erro ao despertar: Seu nível de personagem (${progLevel}) precisa ser no mínimo 100!` }]);
      return;
    }
    if (progClass === "Novice") {
      addSimLogs([{ type: "error", text: `[Progression] Erro: Escolha sua vocação base de nível 10 antes de realizar o Despertar.` }]);
      return;
    }
    if (progSubclass !== "" || awakenedAffinity !== "") {
      addSimLogs([{ type: "error", text: `[Progression] Erro: O Despertar Elemental já foi realizado e é permanente! Subclasse: ${progSubclass} (awakened_affinity = ${awakenedAffinity.toUpperCase()}).` }]);
      return;
    }

    // Calcula dominante elemental baseada nos scores e níveis de afinidade
    const affinities = [
      { name: "Holy", level: affinityHolyLvl, xp: affinityHolyXp, id: "holy" },
      { name: "Fire", level: affinityFireLvl, xp: affinityFireXp, id: "fire" },
      { name: "Ice", level: affinityIceLvl, xp: affinityIceXp, id: "ice" },
      { name: "Shadow", level: affinityShadowLvl, xp: affinityShadowXp, id: "shadow" },
      { name: "Nature", level: affinityNatureLvl, xp: affinityNatureXp, id: "nature" }
    ];

    let dominant = affinities[0];
    for (const aff of affinities) {
      if (aff.level > dominant.level || (aff.level === dominant.level && aff.xp > dominant.xp)) {
        dominant = aff;
      }
    }

    if (dominant.level < 100) {
      addSimLogs([{ type: "error", text: `[Progression] Erro: Sua afinidade dominante (${dominant.name}) está no nível ${dominant.level}. Requer Nível de Afinidade = 100 para despertar!` }]);
      return;
    }

    const calculatedSubclass = `${dominant.name} ${progClass}`;
    const awakenedKey = dominant.id;

    addSimLogs([
      { type: "info", text: `[Network] Enviando pacote binário CS_UNLOCK_SUBCLASS (Opcode: 9102) com bloqueio FOR UPDATE.` },
      { type: "info", text: `[Progression] Transação atômica iniciada com nível de isolamento Repeatable Read.` },
      { type: "info", text: `[Progression] Afinidades independentes: Holy(Lvl ${affinityHolyLvl}), Fire(Lvl ${affinityFireLvl}), Ice(Lvl ${affinityIceLvl}), Shadow(Lvl ${affinityShadowLvl}), Nature(Lvl ${affinityNatureLvl}). Dominante: ${dominant.name} (Lvl 100).` },
      { type: "info", text: `[Progression] Persistindo no banco: awakened_affinity = '${awakenedKey}' (LOCK DE AWAKENING DEFINITIVO ATIVADO).` },
      { type: "info", text: `[Network] Pacote Recebido: SC_UNLOCK_SUBCLASS_RESP (Opcode: 9103). Despertar completo! Bônus passivo de +15% de dano elemental ativo!` }
    ]);

    setProgSubclass(calculatedSubclass);
    setAwakenedAffinity(awakenedKey);

    const nowStr = new Date().toTimeString().split(" ")[0];
    setPersistentLogs(prev => [...prev, {
      id: Math.random().toString(),
      type: "Audit Log",
      details: `Novo Registro: Personagem ascendeu à subclasse: ${calculatedSubclass} (awakened_affinity = ${awakenedKey}). Este despertar é permanente e irrevogável.`,
      timestamp: nowStr
    }]);
  };

  const handleTravelSim = (city: string, tx: number, ty: number) => {
    addSimLogs([
      { type: "info", text: `[Movement] Solicitando movimento para ${city} coordenadas [X: ${tx}, Y: ${ty}] (CS_MOVE_REQUEST).` }
    ]);

    // Check if traveling to Ymirr's Hidden Cavern (blocked)
    const isYmirrCavern = tx === 5980 && ty === 5965;
    if (isYmirrCavern) {
      addSimLogs([
        { type: "error", text: `[Region Gate] ACESSO REJEITADO à Caverna de Ymirr: Região oculta e inacessível por viagem direta de navio!` },
        { type: "warning", text: `[Movement] Erro: Esta caverna só pode ser alcançada através de futuras linhas de missão, travessia manual das montanhas ou masmorras.` }
      ]);
      return;
    }

    // Check if traveling to Fire Continent
    const isTargetFire = tx >= 2000 && tx <= 2199 && ty >= 2000 && ty <= 2199;
    // Check if traveling to Ice Continent
    const isTargetIce = tx >= 5800 && tx <= 6000 && ty >= 5800 && ty <= 6000;
    // Check if traveling to Holy Continent
    const isTargetHoly = tx >= 4800 && tx <= 5000 && ty >= 4800 && ty <= 5000;
    // Check if traveling to Abyssia
    const isTargetAbyssia = tx >= 3000 && tx <= 3799 && ty >= 3000 && ty <= 3799;

    if (isTargetAbyssia) {
      if (progLevel < 150) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO a Abyssia: Nível ${progLevel} é insuficiente. Nível mínimo 150 requerido!` },
          { type: "warning", text: `[Movement] Rubberbanding forçado! Teletransportado de volta para as coordenadas iniciais [X: 100, Y: 100] (SC_MOVE_CONFIRM).` }
        ]);
        setProgX(100);
        setProgY(100);
      } else if (!abyssiQuestCompleted) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO: Missão de fim de jogo 'abyssal_endgame_questline' não concluída!` },
          { type: "warning", text: `[Movement] O barco abissal recusou sua entrada. Complete a difícil jornada de fim de jogo primeiro. Retornando para [X: 100, Y: 100].` }
        ]);
        setProgX(100);
        setProgY(100);
      } else if (!abyssiPermissionFlag) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO: Falta permissão de trânsito em sua conta (Abyssi Access Permission Flag)!` },
          { type: "warning", text: `[Movement] O oficial da resistência barrou a rota. Obtenha a permissão de trânsito em sua conta. Retornando para [X: 100, Y: 100].` }
        ]);
        setProgX(100);
        setProgY(100);
      } else {
        addSimLogs([
          { type: "info", text: `[Region Gate] ACESSO ABISSAL AUTORIZADO: Barco liberado com sucesso para herói supremo nível ${progLevel}!` },
          { type: "info", text: `[Movement] Posição confirmada no servidor em Last Bastion [X: ${tx}, Y: ${ty}] (SC_MOVE_CONFIRM). Bem-vindo à escuridão de Abyssia!` }
        ]);
        setProgX(tx);
        setProgY(ty);
      }
    } else if (isTargetFire) {
      if (progLevel < 50) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO ao Continente de Fogo: Jogador nível ${progLevel} tentou cruzar o portal.` },
          { type: "warning", text: `[Movement] Rubberbanding forçado! Nível mínimo 50 requerido. Retornando para [X: 100, Y: 100] (SC_MOVE_CONFIRM).` }
        ]);
        setProgX(100);
        setProgY(100);
      } else {
        addSimLogs([
          { type: "info", text: `[Region Gate] ACESSO AO CONTINENTE DE FOGO AUTORIZADO: Portas de Pyra Magnus e Molten Anvil liberadas para herói nível ${progLevel}!` },
          { type: "info", text: `[Economy] Taxas de porto e passes de travessia térmica processados pelo tesouro de Pyra Magnus.` },
          { type: "info", text: `[Movement] Posição confirmada no servidor [X: ${tx}, Y: ${ty}] (SC_MOVE_CONFIRM). Bem-vindo ao continente gélido/ardente em ${city}!` }
        ]);
        setProgX(tx);
        setProgY(ty);
      }
    } else if (isTargetIce) {
      addSimLogs([
        { type: "info", text: `[Region Gate] ACESSO AO CONTINENTE GELADO AUTORIZADO: Rota de navio disponível para todos os níveis (Sem restrições de nível).` },
        { type: "info", text: `[Economy] Taxa de viagem marítima (100 moedas de ouro) cobrada com sucesso pelo Mestre de Porto.` },
        { type: "info", text: `[Movement] Posição confirmada no servidor [X: ${tx}, Y: ${ty}] (SC_MOVE_CONFIRM). Bem-vindo ao continente gélido em ${city}!` }
      ]);
      setProgX(tx);
      setProgY(ty);
    } else if (isTargetHoly) {
      if (progLevel < 50) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO ao Continente Sagrado: Jogador nível ${progLevel} tentou cruzar o portal.` },
          { type: "warning", text: `[Movement] Rubberbanding forçado! Nível mínimo 50 requerido. Retornando para [X: 100, Y: 100] (SC_MOVE_CONFIRM).` }
        ]);
        setProgX(100);
        setProgY(100);
      } else if (!holyQuestCompleted) {
        addSimLogs([
          { type: "error", text: `[Region Gate] ACESSO REJEITADO: Missão 'holy_continent_access_trial' não concluída!` },
          { type: "warning", text: `[Movement] O barco sagrado recusou sua entrada. Complete as Provações Espirituais primeiro. Retornando para [X: 100, Y: 100].` }
        ]);
        setProgX(100);
        setProgY(100);
      } else {
        addSimLogs([
          { type: "info", text: `[Region Gate] ACESSO SAGRADO AUTORIZADO: Portais abertos e barco liberado para herói nível ${progLevel}!` },
          { type: "info", text: `[Movement] Posição confirmada no servidor [X: ${tx}, Y: ${ty}] (SC_MOVE_CONFIRM). Bem-vindo ao continente sagrado em ${city}!` }
        ]);
        setProgX(tx);
        setProgY(ty);
      }
    } else {
      addSimLogs([
        { type: "info", text: `[Region Gate] ACESSO AUTORIZADO para herói nível ${progLevel}.` },
        { type: "info", text: `[Movement] Posição confirmada no servidor [X: ${tx}, Y: ${ty}] (SC_MOVE_CONFIRM). Bem-vindo a ${city}!` }
      ]);
      setProgX(tx);
      setProgY(ty);
    }
  };

  // Initialize with Boot logs
  useEffect(() => {
    addSimLogs([
      { type: "info", text: "[GameManager] Inicializando MMORPG Client Bootstrap..." },
      { type: "info", text: "[ConfigManager] Tentando carregar arquivo de config em: user://client_config.json" },
      { type: "warning", text: "[ConfigManager] Arquivo de config não encontrado. Salvando configurações padrões." },
      { type: "info", text: "[ConfigManager] Configurações padrões gravadas com sucesso em: user://client_config.json" },
      { type: "info", text: "[State: Boot] Inicializando sistemas básicos do MMORPG Light and Shadow..." },
      { type: "info", text: "[State: Boot] Motores de Configuração, Barramento de Eventos e Rede prontos." }
    ]);
  }, []);

  // Scroll simulator console to bottom
  useEffect(() => {
    consoleEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  // Helper to add logs with timestamp
  const addSimLogs = (newLogs: { type: "info" | "warning" | "error"; text: string }[]) => {
    const formatted = newLogs.map((l, i) => {
      const date = new Date();
      const timeStr = date.toTimeString().split(" ")[0] + "." + String(date.getMilliseconds()).padStart(3, "0");
      return {
        id: Math.random().toString(36).substring(2, 9) + "-" + i,
        timestamp: timeStr,
        type: l.type,
        text: l.text
      };
    });
    setLogs(prev => [...prev, ...formatted]);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(selectedFile.code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const toggleFolder = (folderKey: string) => {
    setExpandedFolders(prev => ({
      ...prev,
      [folderKey]: !prev[folderKey]
    }));
  };

  // Syntax highlighting logic
  const highlightCode = (code: string, language: string) => {
    if (language === "go") {
      let escaped = code
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");

      // String literals & backticks
      escaped = escaped.replace(/(&quot;[^\n]*?&quot;)/g, '<span class="text-emerald-400 font-medium">$1</span>');
      escaped = escaped.replace(/(\`[^\n]*?\`)/g, '<span class="text-emerald-400 font-medium">$1</span>');

      // Single line comments
      escaped = escaped.replace(/(\/\/.*)/g, '<span class="text-slate-500 font-normal italic">$1</span>');

      // Keywords
      const goKeywords = [
        "package", "import", "type", "struct", "func", "return", "var", "const", "if", "else", "switch",
        "case", "default", "select", "chan", "go", "defer", "map", "range", "for", "interface", "nil", "true", "false",
        "make", "append", "copy", "close", "panic", "recover", "string", "int", "uint16", "uint32", "byte", "error", "any"
      ];

      goKeywords.forEach(keyword => {
        const regex = new RegExp(`\\b(${keyword})\\b`, "g");
        escaped = escaped.replace(regex, '<span class="text-violet-400 font-semibold">$1</span>');
      });

      // Special types and libraries
      const goTypes = [
        "Config", "Packet", "PostgresPool", "RedisClient", "GatewayServer", "AuthServer", "WorldServer",
        "slog", "net", "http", "sql", "redis", "binary", "context", "sync", "time", "s", "db"
      ];

      goTypes.forEach(type => {
        const regex = new RegExp(`\\b(${type})\\b`, "g");
        escaped = escaped.replace(regex, '<span class="text-amber-400 font-medium">$1</span>');
      });

      return escaped;
    }

    if (language === "ini") {
      let escaped = code
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
      escaped = escaped.replace(/(;.*)/g, '<span class="text-slate-500 font-normal">$1</span>');
      escaped = escaped.replace(/^(\[.*\])/gm, '<span class="text-amber-500 font-bold">$1</span>');
      escaped = escaped.replace(/^([a-zA-Z0-9_\/]+)\s*=/gm, '<span class="text-violet-400">$1</span> =');
      return escaped;
    }

    if (language === "sql") {
      let escaped = code
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
      escaped = escaped.replace(/(--.*)/g, '<span class="text-slate-500 font-normal italic">$1</span>');
      
      const sqlKeywords = [
        "CREATE", "TABLE", "IF", "NOT", "EXISTS", "PRIMARY", "KEY", "SERIAL", "UNIQUE", 
        "VARCHAR", "NOT", "NULL", "TIMESTAMP", "WITH", "TIME", "ZONE", "DEFAULT", 
        "CURRENT_TIMESTAMP", "INDEX", "ON", "INT", "REFERENCES", "CASCADE", "RESTRICT",
        "FLOAT", "BIGINT"
      ];

      sqlKeywords.forEach(keyword => {
        const regex = new RegExp(`\\b(${keyword})\\b`, "gi");
        escaped = escaped.replace(regex, '<span class="text-violet-400 font-semibold">$1</span>');
      });

      return escaped;
    }
    if (language === "xml") {
      let escaped = code
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
      escaped = escaped.replace(/(&lt;\/?[a-zA-Z0-9_\.:-]+&gt;)/g, '<span class="text-amber-500">$1</span>');
      escaped = escaped.replace(/([a-zA-Z0-9_-]+)=/g, '<span class="text-violet-400">$1</span>=');
      return escaped;
    }

    // CSharp
    let escaped = code
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");

    // String literals
    escaped = escaped.replace(/(&quot;[^\n]*?&quot;)/g, '<span class="text-emerald-400 font-medium">$1</span>');
    escaped = escaped.replace(/(\'[^\n]*?\')/g, '<span class="text-emerald-400 font-medium">$1</span>');

    // Single line comments
    escaped = escaped.replace(/(\/\/.*)/g, '<span class="text-slate-500 font-normal italic">$1</span>');

    // Annotations
    escaped = escaped.replace(/(\[[a-zA-Z0-9_]+\])/g, '<span class="text-cyan-400 font-semibold">$1</span>');

    // Keywords
    const keywords = [
      "using", "namespace", "public", "class", "interface", "async", "await", "Task",
      "override", "void", "string", "int", "double", "float", "bool", "enum",
      "private", "readonly", "get", "set", "new", "null", "throw", "try", "catch",
      "lock", "return", "switch", "case", "break", "event", "Action", "delegate",
      "partial", "const", "ushort", "byte", "static", "var", "out", "in", "if", "else", "while", "foreach"
    ];

    keywords.forEach(keyword => {
      const regex = new RegExp(`\\b(${keyword})\\b`, "g");
      escaped = escaped.replace(regex, '<span class="text-violet-400 font-semibold">$1</span>');
    });

    // Custom types & classes
    const types = [
      "GameManager", "ConfigManager", "EventBus", "SceneManager", "NetworkManager", "AppStateMachine",
      "IAppState", "BootState", "LoadingState", "MenuState", "ConnectingState", "CharacterSelectionState",
      "InGameState", "DisconnectedState", "ShutdownState", "GD", "Node", "ProgressBar", "Label",
      "LineEdit", "Button", "Control", "TcpClient", "NetworkStream", "GamePacket", "PacketOpcode",
      "Array", "Dictionary", "Delegate", "Error", "PackedScene", "Viewport", "ResourceLoader",
      "DisplayServer", "Variant", "Json"
    ];

    types.forEach(type => {
      const regex = new RegExp(`\\b(${type})\\b`, "g");
      escaped = escaped.replace(regex, '<span class="text-amber-400 font-medium">$1</span>');
    });

    return escaped;
  };

  // State Machine simulation controls
  const handleTriggerBoot = () => {
    setSimState("Boot");
    addSimLogs([
      { type: "info", text: "[StateMachine] Comutando estado manual para -> Boot" },
      { type: "info", text: "[State: Boot] Inicializando sistemas básicos do MMORPG Light and Shadow..." },
      { type: "info", text: "[State: Boot] Motores de Configuração, Barramento de Eventos e Rede prontos." },
      { type: "info", text: "[StateMachine] Transição automática ativada: Boot -> Loading" }
    ]);
    setTimeout(() => {
      handleTriggerLoading();
    }, 1000);
  };

  const handleTriggerLoading = () => {
    setSimState("Loading");
    setIsSimulatingLoading(true);
    setLoadingPercent(0);
    addSimLogs([
      { type: "info", text: "[StateMachine] Iniciando transição: Boot -> Loading" },
      { type: "info", text: "[SceneManager] Iniciando carregamento assíncrono para: res://scenes/bootstrap.tscn" },
      { type: "info", text: "[State: Loading] Entrou no estado de carregamento de Assets e UI..." }
    ]);

    // Simulate progress updates
    const interval = setInterval(() => {
      setLoadingPercent(prev => {
        const next = prev + 25;
        if (next >= 100) {
          clearInterval(interval);
          setIsSimulatingLoading(false);
          addSimLogs([
            { type: "info", text: "[SceneManager] Carregando: 100.0%" },
            { type: "info", text: "[SceneManager] Recurso de Bootstrap carregado! Instanciando cena..." },
            { type: "info", text: "[SceneManager] Cena comutada com sucesso. Nova cena ativa: BootstrapUI" },
            { type: "info", text: "[State: Loading] Recursos de interface carregados. Redirecionando ao Menu Principal." }
          ]);
          setTimeout(() => {
            handleTriggerMenu();
          }, 800);
          return 100;
        } else {
          addSimLogs([
            { type: "info", text: `[SceneManager] Carregando: ${next}.0%` }
          ]);
          return next;
        }
      });
    }, 400);
  };

  const handleTriggerMenu = () => {
    setSimState("Menu");
    addSimLogs([
      { type: "info", text: "[StateMachine] Iniciando transição: Loading -> Menu" },
      { type: "info", text: "[SceneManager] Iniciando carregamento assíncrono para: res://scenes/main_menu.tscn" },
      { type: "info", text: "[SceneManager] Recurso carregado! Instanciando nova cena..." },
      { type: "info", text: "[SceneManager] Cena comutada com sucesso. Nova cena ativa: MainMenuUI" },
      { type: "info", text: "[MainMenuUI] Eventos de botões do menu principal configurados." },
      { type: "info", text: "[State: Menu] Exibindo Menu Principal..." }
    ]);
  };

  const handleTriggerConnect = () => {
    setSimState("Connecting");
    setIsSimulatingTcp(true);
    addSimLogs([
      { type: "info", text: `[MainMenuUI] Botão login pressionado. Publicando OnLoginAttempted com credencial '${simUsername}'` },
      { type: "info", text: `[State: Menu] Tentativa de login recebida para credencial: ${simUsername}` },
      { type: "info", text: "[StateMachine] Iniciando transição: Menu -> Connecting" },
      { type: "info", text: "[State: Connecting] Tentando conectar aos servidores do MMORPG..." },
      { type: "info", text: "[NetworkManager] Conectando ao servidor em 127.0.0.1:8080..." }
    ]);
  };

  const handleConnectSuccess = () => {
    setIsSimulatingTcp(false);
    addSimLogs([
      { type: "info", text: "[NetworkManager] Conectado e ouvindo barramento de rede do servidor." },
      { type: "info", text: "[State: Connecting] Conexão TCP estabelecida com sucesso!" },
      { type: "info", text: "[StateMachine] Iniciando transição: Connecting -> CharacterSelection" },
      { type: "info", text: "[SceneManager] Iniciando carregamento assíncrono para: res://scenes/char_selection.tscn" },
      { type: "info", text: "[SceneManager] Recurso carregado! Instanciando nova cena..." },
      { type: "info", text: "[SceneManager] Cena comutada com sucesso. Nova cena ativa: CharacterSelectionUI" },
      { type: "info", text: "[State: CharacterSelection] Entrou no fluxo de seleção de personagens." }
    ]);
    setSimState("CharacterSelection");
  };

  const handleConnectFailure = () => {
    setIsSimulatingTcp(false);
    addSimLogs([
      { type: "error", text: "[NetworkManager] Falha na conexão: Tempo Limite Excedido (Timeout de soquete)." },
      { type: "error", text: "[State: Connecting] Falha crítica ao se conectar com o servidor." },
      { type: "info", text: "[StateMachine] Iniciando transição: Connecting -> Disconnected" }
    ]);
    setSimState("Disconnected");

    // Auto returns to menu after 2s like the C# state machine does
    setTimeout(() => {
      addSimLogs([
        { type: "info", text: "[State: Disconnected] Retornando ao menu principal..." }
      ]);
      handleTriggerMenu();
    }, 2000);
  };

  const handleSelectCharacter = () => {
    addSimLogs([
      { type: "info", text: `[State: CharacterSelection] Personagem selecionado: ${simUsername}. Entrando no mundo de Light & Shadow!` },
      { type: "info", text: "[StateMachine] Iniciando transição: CharacterSelection -> InGame" },
      { type: "info", text: "[SceneManager] Iniciando carregamento assíncrono para: res://scenes/game_world.tscn" },
      { type: "info", text: "[SceneManager] Recurso carregado! Instanciando nova cena..." },
      { type: "info", text: "[SceneManager] Cena comutada com sucesso. Nova cena ativa: GameWorld" },
      { type: "info", text: "[State: InGame] Sincronização do estado do jogador com o servidor completa. Bem-vindo ao mundo!" }
    ]);
    setSimState("InGame");
  };

  const handleSendHeartbeat = () => {
    addSimLogs([
      { type: "info", text: "[NetworkManager] Enviando pacote de keep-alive (Heartbeat) de rotina." },
      { type: "info", text: "[NetworkManager] Recebidos 4 bytes do servidor (Heartbeat ACK)." }
    ]);
  };

  const handleNetworkLoss = () => {
    addSimLogs([
      { type: "error", text: "[NetworkManager] Servidor fechou a conexão de forma abrupta." },
      { type: "warning", text: "[NetworkManager] Rede desconectada localmente e buffers limpos." },
      { type: "error", text: "[State: InGame] Conexão com o servidor perdida repentinamente!" },
      { type: "info", text: "[StateMachine] Iniciando transição: InGame -> Disconnected" },
      { type: "info", text: "[State: Disconnected] Conexão terminada. Limpando dados do mapa." }
    ]);
    setSimState("Disconnected");

    setTimeout(() => {
      addSimLogs([
        { type: "info", text: "[State: Disconnected] Retornando ao menu principal..." }
      ]);
      handleTriggerMenu();
    }, 2000);
  };

  const handleTriggerShutdown = () => {
    addSimLogs([
      { type: "info", text: `[GameManager] Pedido de fechamento detectado pelo OS. Evitando fechamento abrupto.` },
      { type: "info", text: `[StateMachine] Iniciando transição: ${simState} -> Shutdown` },
      { type: "info", text: "[State: Shutdown] Desligando o cliente do jogo de forma segura e liberando buffers de hardware..." },
      { type: "info", text: "[ConfigManager] Configurações de preferências gravadas em disco (Volume, Fullscreen, etc.)" },
      { type: "info", text: "[NetworkManager] Soquetes fechados e liberados com segurança." },
      { type: "info", text: "[State: Shutdown] Conexão encerrada, configurações salvas. Finalizando loop do motor Godot..." }
    ]);
    setSimState("Shutdown");
  };

  const handleSimulateCrash = (recoverable: boolean) => {
    if (recoverable) {
      addSimLogs([
        { type: "error", text: "[NetworkManager] EXCEÇÃO NO LOOP DE LEITURA DE REDE: System.IO.IOException: Falha na stream de socket." },
        { type: "warning", text: "[CrashHandler] Interceptada UnobservedTaskException de Tarefa Assíncrona!" },
        { type: "warning", text: "[CrashHandler] Gravando relatório em: user://logs/crash_rec_2026.log" },
        { type: "info", text: "[CrashHandler] Formato do Log Gravado: Timestamp | Thread ID | Exception Type | Message | Stack Trace" },
        { type: "info", text: "[CrashHandler] Tentando recuperação automatizada (Recovery Attempt)..." },
        { type: "info", text: "[ServiceRegistry] Resolvendo EventBus de forma thread-safe para disparo de sinal..." },
        { type: "info", text: "[EventBus] Publicando evento OnNetworkDisconnect para forçar auto-reconexão graciosa." },
        { type: "info", text: "[CrashHandler] Recuperação bem-sucedida! Estado original de gameplay restaurado." }
      ]);
    } else {
      addSimLogs([
        { type: "error", text: "[GameManager] EXCEÇÃO FATAL CRÍTICA NÃO TRATADA: System.NullReferenceException: Object reference not set to an instance of an object." },
        { type: "warning", text: "[CrashHandler] Interceptada UnhandledException em AppDomain principal!" },
        { type: "warning", text: "[CrashHandler] Gravando relatório estrito em: user://logs/crash_fatal_2026.log" },
        { type: "info", text: "[CrashHandler] Formato do Log Gravado: Timestamp | Thread ID | Exception Type | Message | Stack Trace" },
        { type: "error", text: "[CrashHandler] Falha irrecuperável. Iniciando Graceful Shutdown via StateMachine..." },
        { type: "info", text: "[ServiceRegistry] Resolvendo AppStateMachine dinamicamente para forçar transição segura..." },
        { type: "info", text: `[StateMachine] Iniciando transição: ${simState} -> Shutdown` },
        { type: "info", text: "[State: Shutdown] Desligando o cliente do jogo de forma segura e liberando buffers de hardware..." },
        { type: "info", text: "[ConfigManager] Configurações de preferências gravadas em disco (Volume, Fullscreen, etc.)" },
        { type: "info", text: "[NetworkManager] Soquetes fechados e liberados com segurança." },
        { type: "info", text: "[State: Shutdown] Conexão encerrada, configurações salvas. Finalizando loop do motor Godot..." }
      ]);
      setSimState("Shutdown");
    }
  };

  const handleInspectPacket = () => {
    addSimLogs([
      { type: "info", text: `[GamePacket] Criando pacote CS_LOGIN_REQUEST para usuário: '${simUsername}'` },
      { type: "info", text: "[GamePacket] Layout do Cabeçalho Oficial (Little Endian, 8 bytes):" },
      { type: "info", text: `  - size (ushort, 2 bytes): 24 bytes [ HEADER_SIZE (8) + payload (16) ]` },
      { type: "info", text: "  - opcode (ushort, 2 bytes): 1002 [ CS_LOGIN_REQUEST ]" },
      { type: "info", text: "  - sequence (uint, 4 bytes): 103 [ Autoincrement thread-safe ]" },
      { type: "info", text: `  - payload (byte[]): '${simUsername}' (UTF-8, 16 bytes)` },
      { type: "info", text: "[GamePacket] Executando Serialize() binário completo..." },
      { type: "info", text: "[GamePacket] Executando Validate() do cabeçalho pré-envio..." },
      { type: "info", text: "[GamePacket] Validação: Size (24 <= 16384) OK | Opcode (1002) OK | Payload correspondente OK." },
      { type: "info", text: "[NetworkManager] Despachando array de 24 bytes estruturado pelo soquete TCP stream." }
    ]);
  };

  const handleRegistryQuery = () => {
    addSimLogs([
      { type: "info", text: "[ServiceRegistry] Consultando registro de serviços ativos via Resolve<T>()..." },
      { type: "info", text: "[ServiceRegistry] - GameManager: RESOLVIDO (Instância única de Autoload vinculada)" },
      { type: "info", text: "[ServiceRegistry] - ConfigManager: RESOLVIDO (Configurações locais de rede/áudio)" },
      { type: "info", text: "[ServiceRegistry] - EventBus: RESOLVIDO (Barramento desacoplado com listeners ativos)" },
      { type: "info", text: "[ServiceRegistry] - NetworkManager: RESOLVIDO (Motor de conexões TCP)" },
      { type: "info", text: "[ServiceRegistry] - SceneManager: RESOLVIDO (Controle de carregamento assíncrono)" },
      { type: "info", text: "[ServiceRegistry] - AppStateMachine: RESOLVIDO (Motor de controle de fluxo de estados)" },
      { type: "info", text: "[ServiceRegistry] Sucesso: Todos os singletons desacoplados e resolvidos de forma thread-safe via DI leve." }
    ]);
  };

  const handleClearLogs = () => {
    setLogs([]);
  };

  const filteredLogs = logs.filter(l => {
    if (logFilter === "all") return true;
    return l.type === logFilter;
  });

  if (GAME_MODE === "CORE") {
    return (
      <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex flex-col selection:bg-amber-500/30 selection:text-amber-200 relative">
        {/* Ambient Top Glows */}
        <div className="absolute top-0 left-1/4 w-96 h-96 bg-amber-500/5 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute top-0 right-1/4 w-96 h-96 bg-violet-500/5 rounded-full blur-3xl pointer-events-none" />

        {/* Minimalist production header */}
        <header className="border-b border-slate-800/80 bg-slate-900/40 backdrop-blur-md px-6 py-4 flex flex-col sm:flex-row justify-between items-center gap-4 z-10 sticky top-0">
          <div className="flex items-center gap-3">
            <div className="relative flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-violet-600 shadow-lg shadow-violet-500/20">
              <Sparkles className="w-5 h-5 text-slate-950" />
            </div>
            <div>
              <h1 className="text-xl font-bold tracking-tight bg-gradient-to-r from-amber-200 via-slate-100 to-violet-200 bg-clip-text text-transparent">
                Light and Shadow MMORPG
              </h1>
              <p className="text-xs text-slate-400 font-medium">
                Official Production Client (Core Mode)
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-[10px] text-slate-500 font-mono select-none">GAME_MODE: CORE</span>
            <div className="flex items-center gap-2">
              <span className="flex h-2 w-2 relative">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              <span className="text-xs text-slate-400 font-mono">SERVERS_ONLINE</span>
            </div>
          </div>
        </header>

        {/* Main gameplay area */}
        <main className="flex-1 p-6 flex flex-col gap-6 z-10">
          <GameBootstrapRoot
            syncInventory={setPlayerInventory}
            syncEquipment={setPlayerEquipped}
            syncLevel={setProgLevel}
            syncClass={setProgClass}
            syncCoins={setCarriedCoins}
            syncCharName={setCreatedCharName}
          />
        </main>

        <footer className="border-t border-slate-800/80 bg-slate-950 py-4 px-6 flex flex-col sm:flex-row justify-between items-center text-[11px] text-slate-500 gap-2">
          <p className="font-semibold uppercase tracking-wider text-slate-600">
            Light & Shadow MMORPG • Production Core Mode
          </p>
          <p className="text-slate-600 font-mono">
            V1.0.0-BETA
          </p>
        </footer>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex flex-col selection:bg-amber-500/30 selection:text-amber-200">
      
      {/* Ambient Top Glows */}
      <div className="absolute top-0 left-1/4 w-96 h-96 bg-amber-500/5 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute top-0 right-1/4 w-96 h-96 bg-violet-500/5 rounded-full blur-3xl pointer-events-none" />

      {/* Header */}
      <header className="border-b border-slate-800/80 bg-slate-900/40 backdrop-blur-md px-6 py-4 flex flex-col md:flex-row justify-between items-center gap-4 z-10 sticky top-0">
        <div className="flex items-center gap-3">
          <div className="relative flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-violet-600 shadow-lg shadow-violet-500/20">
            <Sparkles className="w-5 h-5 text-slate-950 animate-pulse" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight bg-gradient-to-r from-amber-200 via-slate-100 to-violet-200 bg-clip-text text-transparent">
              Light and Shadow MMORPG
            </h1>
            <p className="text-xs text-slate-400 font-medium">
              Client Bootstrap Framework • Godot 4 + C# • Clean Architecture
            </p>
          </div>
        </div>

        {/* Tab Navigation */}
        <nav className="flex bg-slate-950/80 p-1 rounded-xl border border-slate-800/80 overflow-x-auto max-w-full">
          {[
            { id: "workspace", label: "Área de Trabalho", icon: Layers },
            { id: "game_entry_sim", label: "Game Entry Sim", icon: Network },
            { id: "world", label: "World Foundation", icon: Globe },
            { id: "progression", label: "Progressão & Combat", icon: Sliders },
            { id: "vocation", label: "Classes & Offhands", icon: Shield },
            { id: "spells", label: "Spells & Skills", icon: Wand2 },
            { id: "bestiary", label: "Bestiário Canônico", icon: Sparkles },
            { id: "bestiary_sim", label: "Monster AI Sim", icon: Sword },
            { id: "loot_death", label: "Loot & Penalidades", icon: Gift },
            { id: "itemization", label: "Itemização & Builds", icon: Sword },
            { id: "activity", label: "Matriz de Atividades", icon: Compass },
            { id: "config", label: "Autoload Godot", icon: Settings },
            { id: "exec", label: "Guia de Execução", icon: Play },
            { id: "arch", label: "Arquitetura Clean", icon: BookOpen }
          ].map(tab => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all duration-200 ${
                  isActive
                    ? "bg-slate-800 text-amber-400 shadow-sm border border-slate-700/50"
                    : "text-slate-400 hover:text-slate-200 hover:bg-slate-900/50"
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                {tab.label}
              </button>
            );
          })}
        </nav>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 p-6 flex flex-col gap-6 z-10">

        <AnimatePresence mode="wait">
          {activeTab === "game_entry_sim" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="flex-1 flex flex-col text-slate-100"
            >
              <Suspense fallback={<div className="bg-slate-900 border border-slate-800/80 p-6 rounded-2xl text-center text-xs text-slate-400 font-mono animate-pulse">Carregando Game Entry Simulator...</div>}>
                <GameEntrySimulator
                  syncInventory={setPlayerInventory}
                  syncEquipment={setPlayerEquipped}
                  syncLevel={setProgLevel}
                  syncClass={setProgClass}
                  syncCoins={setCarriedCoins}
                  syncCharName={setCreatedCharName}
                />
              </Suspense>
            </motion.div>
          )}

          {activeTab === "bestiary_sim" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="flex-1 flex flex-col text-slate-100"
            >
              <Suspense fallback={<div className="bg-slate-900 border border-slate-800/80 p-6 rounded-2xl text-center text-xs text-slate-400 font-mono animate-pulse">Carregando Monster AI Simulator...</div>}>
                <MonsterSimulator />
              </Suspense>
            </motion.div>
          )}

          {activeTab === "workspace" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 xl:grid-cols-12 gap-6"
            >
              {/* Left Panel: File Explorer */}
              <div className="xl:col-span-3 flex flex-col gap-4">
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 h-[620px] flex flex-col backdrop-blur-sm shadow-xl">
                  <div className="flex items-center justify-between border-b border-slate-800/80 pb-3 mb-3">
                    <div className="flex items-center gap-2 text-slate-300 font-semibold text-xs tracking-wider uppercase">
                      <Folder className="w-4 h-4 text-amber-500" />
                      Estrutura do Projeto
                    </div>
                    <span className="text-[10px] bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full border border-slate-700/40">
                      Godot 4.2+
                    </span>
                  </div>

                  {/* Interactive Tree */}
                  <div className="flex-1 overflow-y-auto text-xs space-y-2 pr-1 custom-scrollbar">
                    
                    {/* Project Root Folder */}
                    <div>
                      <div
                        onClick={() => toggleFolder("root")}
                        className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                      >
                        {expandedFolders.root ? (
                          <FolderOpen className="w-4 h-4 text-amber-500/80 shrink-0" />
                        ) : (
                          <Folder className="w-4 h-4 text-amber-500/80 shrink-0" />
                        )}
                        <span className="font-semibold text-slate-100">LightAndShadowClient/</span>
                      </div>

                      {expandedFolders.root && (
                        <div className="pl-4 border-l border-slate-800/60 ml-3.5 mt-1 space-y-1.5">
                          
                          {/* project.godot */}
                          <div
                            onClick={() => setSelectedFile(codeFiles.find(f => f.name === "project.godot")!)}
                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                              selectedFile.name === "project.godot"
                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                : "text-slate-300 hover:bg-slate-800/30"
                            }`}
                          >
                            <Settings className="w-3.5 h-3.5 text-violet-400" />
                            <span>project.godot</span>
                          </div>

                          {/* LightAndShadow.csproj */}
                          <div
                            onClick={() => setSelectedFile(codeFiles.find(f => f.name === "LightAndShadow.csproj")!)}
                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                              selectedFile.name === "LightAndShadow.csproj"
                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                : "text-slate-300 hover:bg-slate-800/30"
                            }`}
                          >
                            <FileCode className="w-3.5 h-3.5 text-blue-400" />
                            <span>LightAndShadow.csproj</span>
                          </div>

                          {/* scenes/ */}
                          <div>
                            <div className="flex items-center gap-2 py-1 px-2 text-slate-400">
                              <Folder className="w-3.5 h-3.5 text-violet-500/80 shrink-0" />
                              <span className="font-semibold text-slate-300">scenes/</span>
                            </div>
                            <div className="pl-4 border-l border-slate-800/40 ml-1.5 space-y-1">
                              <div className="flex items-center gap-2 py-1 px-2 text-slate-400 hover:text-slate-200">
                                <FileText className="w-3.5 h-3.5 text-amber-500/60" />
                                <span className="italic">bootstrap.tscn</span>
                              </div>
                              <div className="flex items-center gap-2 py-1 px-2 text-slate-400 hover:text-slate-200">
                                <FileText className="w-3.5 h-3.5 text-violet-500/60" />
                                <span className="italic">main_menu.tscn</span>
                              </div>
                            </div>
                          </div>

                          {/* src/ */}
                          <div>
                            <div
                              onClick={() => toggleFolder("src")}
                              className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                            >
                              {expandedFolders.src ? (
                                <FolderOpen className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                              ) : (
                                <Folder className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                              )}
                              <span className="font-semibold">src/</span>
                            </div>

                            {expandedFolders.src && (
                              <div className="pl-4 border-l border-slate-800/60 ml-1.5 mt-1 space-y-1.5">
                                
                                {/* src/Core/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("core")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.core ? (
                                      <FolderOpen className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                                    ) : (
                                      <Folder className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                                    )}
                                    <span className="text-amber-200 font-medium">Core/</span>
                                  </div>

                                  {expandedFolders.core && (
                                    <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                      {["GameManager.cs", "ServiceRegistry.cs", "CrashHandler.cs", "EventBus.cs", "SceneManager.cs", "ConfigManager.cs", "NetworkManager.cs"].map(name => {
                                        const file = codeFiles.find(f => f.name === name)!;
                                        return (
                                          <div
                                            key={name}
                                            onClick={() => setSelectedFile(file)}
                                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                              selectedFile.name === name
                                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                                : "text-slate-300 hover:bg-slate-800/30"
                                            }`}
                                          >
                                            <FileCode className="w-3.5 h-3.5 text-amber-400/80" />
                                            <span>{name}</span>
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>

                                {/* src/StateMachine/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("stateMachine")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.stateMachine ? (
                                      <FolderOpen className="w-3.5 h-3.5 text-violet-500/80 shrink-0" />
                                    ) : (
                                      <Folder className="w-3.5 h-3.5 text-violet-500/80 shrink-0" />
                                    )}
                                    <span className="text-violet-200 font-medium">StateMachine/</span>
                                  </div>

                                  {expandedFolders.stateMachine && (
                                    <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1.5">
                                      {/* IAppState.cs */}
                                      <div
                                        onClick={() => setSelectedFile(codeFiles.find(f => f.name === "IAppState.cs")!)}
                                        className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                          selectedFile.name === "IAppState.cs"
                                            ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                            : "text-slate-300 hover:bg-slate-800/30"
                                        }`}
                                      >
                                        <FileCode className="w-3.5 h-3.5 text-violet-400/80" />
                                        <span>IAppState.cs</span>
                                      </div>

                                      {/* AppStateMachine.cs */}
                                      <div
                                        onClick={() => setSelectedFile(codeFiles.find(f => f.name === "AppStateMachine.cs")!)}
                                        className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                          selectedFile.name === "AppStateMachine.cs"
                                            ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                            : "text-slate-300 hover:bg-slate-800/30"
                                        }`}
                                      >
                                        <FileCode className="w-3.5 h-3.5 text-violet-400/80" />
                                        <span>AppStateMachine.cs</span>
                                      </div>

                                      {/* src/StateMachine/States/ */}
                                      <div>
                                        <div
                                          onClick={() => toggleFolder("states")}
                                          className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                        >
                                          {expandedFolders.states ? (
                                            <FolderOpen className="w-3 h-3 text-pink-500 shrink-0" />
                                          ) : (
                                            <Folder className="w-3 h-3 text-pink-500 shrink-0" />
                                          )}
                                          <span className="text-pink-300">States/</span>
                                        </div>

                                        {expandedFolders.states && (
                                          <div className="pl-4 border-l border-slate-800/30 ml-1 mt-1">
                                            <div
                                              onClick={() => setSelectedFile(codeFiles.find(f => f.name === "States.cs")!)}
                                              className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                                selectedFile.name === "States.cs"
                                                  ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                                  : "text-slate-300 hover:bg-slate-800/30"
                                              }`}
                                            >
                                              <FileCode className="w-3.5 h-3.5 text-pink-400/80" />
                                              <span>States.cs</span>
                                            </div>
                                          </div>
                                        )}
                                      </div>
                                    </div>
                                  )}
                                </div>

                                {/* src/UI/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("ui")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.ui ? (
                                      <FolderOpen className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                                    ) : (
                                      <Folder className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                                    )}
                                    <span className="text-emerald-200">UI/</span>
                                  </div>

                                  {expandedFolders.ui && (
                                    <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                      {["BootstrapUI.cs", "MainMenuUI.cs"].map(name => {
                                        const file = codeFiles.find(f => f.name === name)!;
                                        return (
                                          <div
                                            key={name}
                                            onClick={() => setSelectedFile(file)}
                                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                              selectedFile.name === name
                                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                                : "text-slate-300 hover:bg-slate-800/30"
                                            }`}
                                          >
                                            <FileCode className="w-3.5 h-3.5 text-emerald-400/80" />
                                            <span>{name}</span>
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>

                                {/* src/Network/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("network")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.network ? (
                                      <FolderOpen className="w-3.5 h-3.5 text-blue-500 shrink-0" />
                                    ) : (
                                      <Folder className="w-3.5 h-3.5 text-blue-500 shrink-0" />
                                    )}
                                    <span className="text-blue-200">Network/</span>
                                  </div>

                                  {expandedFolders.network && (
                                    <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                      {["GamePacket.cs", "PacketOpcode.cs"].map(name => {
                                        const file = codeFiles.find(f => f.name === name)!;
                                        return (
                                          <div
                                            key={name}
                                            onClick={() => setSelectedFile(file)}
                                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                              selectedFile.name === name
                                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                                : "text-slate-300 hover:bg-slate-800/30"
                                            }`}
                                          >
                                            <FileCode className="w-3.5 h-3.5 text-blue-400/80" />
                                            <span>{name}</span>
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>

                                {/* src/Movement/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("movement")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.movement ? (
                                      <FolderOpen className="w-3.5 h-3.5 text-orange-500 shrink-0" />
                                    ) : (
                                      <Folder className="w-3.5 h-3.5 text-orange-500 shrink-0" />
                                    )}
                                    <span className="text-orange-200">Movement/</span>
                                  </div>

                                  {expandedFolders.movement && (
                                    <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                      {["WorldManager.cs", "ChunkManager.cs", "MovementController.cs", "Pathfinding.cs"].map(name => {
                                        const file = codeFiles.find(f => f.name === name)!;
                                        return (
                                          <div
                                            key={name}
                                            onClick={() => setSelectedFile(file)}
                                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                              selectedFile && selectedFile.name === name
                                                ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                                : "text-slate-300 hover:bg-slate-800/30"
                                            }`}
                                          >
                                            <FileCode className="w-3.5 h-3.5 text-orange-400/80" />
                                            <span>{name}</span>
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>

                              </div>
                            )}
                          </div>

                        </div>
                      )}
                    </div>

                    {/* Backend Root Folder */}
                    <div className="mt-4 pt-4 border-t border-slate-800/60">
                      <div
                        onClick={() => toggleFolder("backend")}
                        className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                      >
                        {expandedFolders.backend ? (
                          <FolderOpen className="w-4 h-4 text-emerald-500/80 shrink-0" />
                        ) : (
                          <Folder className="w-4 h-4 text-emerald-500/80 shrink-0" />
                        )}
                        <span className="font-bold text-emerald-400">LightAndShadowBackend/</span>
                      </div>

                      {expandedFolders.backend && (
                        <div className="pl-4 border-l border-emerald-950/60 ml-3.5 mt-1 space-y-1.5">
                          
                          {/* go.mod */}
                          <div
                            onClick={() => setSelectedFile(codeFiles.find(f => f.name === "go.mod")!)}
                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                              selectedFile && selectedFile.name === "go.mod"
                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                : "text-slate-300 hover:bg-slate-800/30"
                            }`}
                          >
                            <Settings className="w-3.5 h-3.5 text-emerald-400" />
                            <span>go.mod</span>
                          </div>

                          {/* docker-compose.yml */}
                          <div
                            onClick={() => setSelectedFile(codeFiles.find(f => f.name === "docker-compose.yml")!)}
                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                              selectedFile && selectedFile.name === "docker-compose.yml"
                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                : "text-slate-300 hover:bg-slate-800/30"
                            }`}
                          >
                            <Settings className="w-3.5 h-3.5 text-blue-400" />
                            <span>docker-compose.yml</span>
                          </div>

                          {/* cmd/ */}
                          <div>
                            <div
                              onClick={() => toggleFolder("cmd")}
                              className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                            >
                              {expandedFolders.cmd ? (
                                <FolderOpen className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              ) : (
                                <Folder className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              )}
                              <span className="font-semibold text-slate-300">cmd/</span>
                            </div>

                            {expandedFolders.cmd && (
                              <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                {[
                                  { n: "gateway/main.go", label: "gateway/main.go" },
                                  { n: "auth/main.go", label: "auth/main.go" },
                                  { n: "world/main.go", label: "world/main.go" }
                                ].map(item => {
                                  const file = codeFiles.find(f => f.name === item.n)!;
                                  return (
                                    <div
                                      key={item.n}
                                      onClick={() => setSelectedFile(file)}
                                      className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                        selectedFile && selectedFile.name === item.n
                                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium"
                                          : "text-slate-300 hover:bg-slate-800/30"
                                      }`}
                                    >
                                      <FileCode className="w-3.5 h-3.5 text-emerald-400/80" />
                                      <span>{item.label}</span>
                                    </div>
                                  );
                                })}
                              </div>
                            )}
                          </div>

                          {/* config/ */}
                          <div>
                            <div
                              onClick={() => toggleFolder("config")}
                              className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                            >
                              {expandedFolders.config ? (
                                <FolderOpen className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              ) : (
                                <Folder className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              )}
                              <span className="font-semibold text-slate-300">config/</span>
                            </div>

                            {expandedFolders.config && (
                              <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                {[{ n: "config.go", label: "config.go" }].map(item => {
                                  const file = codeFiles.find(f => f.name === item.n)!;
                                  return (
                                    <div
                                      key={item.n}
                                      onClick={() => setSelectedFile(file)}
                                      className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                        selectedFile && selectedFile.name === item.n
                                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium"
                                          : "text-slate-300 hover:bg-slate-800/30"
                                      }`}
                                    >
                                      <FileCode className="w-3.5 h-3.5 text-emerald-400/80" />
                                      <span>{item.label}</span>
                                    </div>
                                  );
                                })}
                              </div>
                            )}
                          </div>

                          {/* migrations/ */}
                          <div>
                            <div
                              onClick={() => toggleFolder("migrations")}
                              className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                            >
                              {expandedFolders.migrations ? (
                                <FolderOpen className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                              ) : (
                                <Folder className="w-3.5 h-3.5 text-amber-500/80 shrink-0" />
                              )}
                              <span className="font-semibold text-slate-300">migrations/</span>
                            </div>

                            {expandedFolders.migrations && (
                              <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1">
                                {[
                                  { n: "0001_create_accounts.up.sql", label: "0001_create_accounts.up.sql" },
                                  { n: "0002_create_characters.up.sql", label: "0002_create_characters.up.sql" },
                                  { n: "0003_create_inventories.up.sql", label: "0003_create_inventories.up.sql" },
                                  { n: "0004_create_guilds.up.sql", label: "0004_create_guilds.up.sql" }
                                ].map(item => {
                                  const file = codeFiles.find(f => f.name === item.n)!;
                                  return (
                                    <div
                                      key={item.n}
                                      onClick={() => setSelectedFile(file)}
                                      className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                        selectedFile && selectedFile.name === item.n
                                          ? "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-medium"
                                          : "text-slate-300 hover:bg-slate-800/30"
                                      }`}
                                    >
                                      <FileText className="w-3.5 h-3.5 text-amber-500/80" />
                                      <span>{item.label}</span>
                                    </div>
                                  );
                                })}
                              </div>
                            )}
                          </div>

                          {/* pkg/ */}
                          <div>
                            <div
                              onClick={() => toggleFolder("pkg")}
                              className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                            >
                              {expandedFolders.pkg ? (
                                <FolderOpen className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              ) : (
                                <Folder className="w-3.5 h-3.5 text-emerald-500/80 shrink-0" />
                              )}
                              <span className="font-semibold text-slate-300">pkg/</span>
                            </div>

                            {expandedFolders.pkg && (
                              <div className="pl-4 border-l border-slate-800/40 ml-1.5 mt-1 space-y-1.5">
                                
                                {/* pkg/db/ */}
                                <div>
                                  <div
                                    onClick={() => toggleFolder("db")}
                                    className="flex items-center gap-2 py-1 px-2 hover:bg-slate-800/30 rounded cursor-pointer transition-colors text-slate-200"
                                  >
                                    {expandedFolders.db ? (
                                      <FolderOpen className="w-3 h-3 text-emerald-500/80 shrink-0" />
                                    ) : (
                                      <Folder className="w-3 h-3 text-emerald-500/80 shrink-0" />
                                    )}
                                    <span className="text-emerald-300">db/</span>
                                  </div>

                                  {expandedFolders.db && (
                                    <div className="pl-4 border-l border-slate-800/30 ml-1 mt-1 space-y-1">
                                      {[
                                        { n: "postgres.go", label: "postgres.go" },
                                        { n: "redis.go", label: "redis.go" }
                                      ].map(item => {
                                        const file = codeFiles.find(f => f.name === item.n)!;
                                        return (
                                          <div
                                            key={item.n}
                                            onClick={() => setSelectedFile(file)}
                                            className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                              selectedFile && selectedFile.name === item.n
                                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium"
                                                : "text-slate-300 hover:bg-slate-800/30"
                                            }`}
                                          >
                                            <FileCode className="w-3 h-3 text-emerald-400/80" />
                                            <span>{item.label}</span>
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>

                                {/* pkg/lifecycle/ */}
                                <div
                                  onClick={() => setSelectedFile(codeFiles.find(f => f.name === "lifecycle.go")!)}
                                  className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                    selectedFile && selectedFile.name === "lifecycle.go"
                                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                      : "text-slate-300 hover:bg-slate-800/30"
                                  }`}
                                >
                                  <FileCode className="w-3.5 h-3.5 text-emerald-400" />
                                  <span>lifecycle.go</span>
                                </div>

                                {/* pkg/logger/ */}
                                <div
                                  onClick={() => setSelectedFile(codeFiles.find(f => f.name === "logger.go")!)}
                                  className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                    selectedFile && selectedFile.name === "logger.go"
                                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                      : "text-slate-300 hover:bg-slate-800/30"
                                  }`}
                                >
                                  <FileCode className="w-3.5 h-3.5 text-emerald-400" />
                                  <span>logger.go</span>
                                </div>

                                {/* pkg/protocol/ */}
                                <div
                                  onClick={() => setSelectedFile(codeFiles.find(f => f.name === "protocol.go")!)}
                                  className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                    selectedFile && selectedFile.name === "protocol.go"
                                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                      : "text-slate-300 hover:bg-slate-800/30"
                                  }`}
                                >
                                  <FileCode className="w-3.5 h-3.5 text-emerald-400" />
                                  <span>protocol.go</span>
                                </div>

                                {/* pkg/messaging/ */}
                                <div
                                  onClick={() => setSelectedFile(codeFiles.find(f => f.name === "messaging.go")!)}
                                  className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                    selectedFile && selectedFile.name === "messaging.go"
                                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                      : "text-slate-300 hover:bg-slate-800/30"
                                  }`}
                                >
                                  <FileCode className="w-3.5 h-3.5 text-emerald-400" />
                                  <span>messaging/messaging.go</span>
                                </div>

                                {/* pkg/scheduler/ */}
                                <div
                                  onClick={() => setSelectedFile(codeFiles.find(f => f.name === "scheduler.go")!)}
                                  className={`flex items-center gap-2 py-1 px-2 rounded cursor-pointer transition-all ${
                                    selectedFile && selectedFile.name === "scheduler.go"
                                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                      : "text-slate-300 hover:bg-slate-800/30"
                                  }`}
                                >
                                  <FileCode className="w-3.5 h-3.5 text-emerald-400" />
                                  <span>scheduler/scheduler.go</span>
                                </div>

                              </div>
                            )}
                          </div>

                        </div>
                      )}
                    </div>
                  </div>

                  {/* Summary Footer */}
                  <div className="mt-4 p-3 bg-slate-950/60 border border-slate-800/80 rounded-xl">
                    <p className="text-[10px] text-slate-400 leading-relaxed font-semibold">
                      💡 Alterne entre o Cliente Godot (C#) e o Backend Bootstrap (Go) para inspecionar os códigos de produção de ponta a ponta.
                    </p>
                  </div>
                </div>
              </div>

              {/* Right Panel: Code Viewer */}
              <div className="xl:col-span-9 flex flex-col gap-6">
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl h-[620px] flex flex-col shadow-xl overflow-hidden backdrop-blur-sm">
                  {/* Code Header */}
                  <div className="bg-slate-900/80 border-b border-slate-800/80 px-6 py-3 flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <FileCode className="w-4 h-4 text-amber-400" />
                        <span className="text-xs font-mono font-bold text-slate-200">
                          {selectedFile.path}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-400 mt-0.5">
                        {selectedFile.description}
                      </p>
                    </div>

                    <button
                      onClick={handleCopy}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[10px] font-semibold bg-slate-800 hover:bg-slate-700 hover:text-white transition-all text-slate-300 border border-slate-700/60"
                    >
                      {copied ? (
                        <>
                          <Check className="w-3.5 h-3.5 text-emerald-400" />
                          <span>Copiado!</span>
                        </>
                      ) : (
                        <>
                          <Copy className="w-3.5 h-3.5 text-slate-400" />
                          <span>Copiar Código</span>
                        </>
                      )}
                    </button>
                  </div>

                  {/* Code Editor Body */}
                  <div className="flex-1 overflow-auto bg-slate-950/90 font-mono text-xs p-6 custom-scrollbar">
                    <pre className="relative overflow-visible leading-relaxed">
                      <code
                        className="block select-text"
                        dangerouslySetInnerHTML={{
                          __html: highlightCode(selectedFile.code, selectedFile.language)
                        }}
                      />
                    </pre>
                  </div>
                </div>
              </div>

              {/* Simulator Panel (Takes Full Width inside workspace tab) */}
              <div className="xl:col-span-12 grid grid-cols-1 lg:grid-cols-12 gap-6 mt-2">
                
                {/* Visualizer Flowchart */}
                <div className="lg:col-span-7 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex flex-col justify-between">
                  <div>
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-2">
                        <Activity className="w-4 h-4 text-amber-500 animate-pulse" />
                        <h3 className="font-bold text-sm tracking-wide text-slate-200 uppercase">
                          Simulador de Ciclo de Vida do Cliente (C#)
                        </h3>
                      </div>
                      <span className="text-[10px] text-slate-400 bg-slate-800 border border-slate-700/40 px-2 py-0.5 rounded">
                        Estado Ativo: <strong className="text-amber-400 font-mono">{simState}</strong>
                      </span>
                    </div>

                    <p className="text-xs text-slate-300 leading-relaxed mb-6">
                      Visualize a execução sequencial da Máquina de Estados da aplicação. Clique nos estados para transicionar ou utilize os gatilhos rápidos para simular comportamentos específicos do MMORPG:
                    </p>

                    {/* Nodes Matrix Layout */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
                      {(Object.keys(STATE_DETAILS) as AppStateType[]).map(stateKey => {
                        const isCurrent = simState === stateKey;
                        const stateConf = STATE_DETAILS[stateKey];
                        const StateIcon = stateConf.icon;

                        return (
                          <div
                            key={stateKey}
                            className={`relative rounded-xl p-3 border transition-all duration-300 select-none ${
                              isCurrent
                                ? `bg-gradient-to-b ${stateConf.color} border-transparent text-slate-950 scale-[1.03] shadow-md ${stateConf.shadow}`
                                : "bg-slate-950/80 border-slate-800 hover:border-slate-700 text-slate-400"
                            }`}
                          >
                            <div className="flex justify-between items-start mb-2">
                              <StateIcon className={`w-5 h-5 ${isCurrent ? "text-slate-950" : "text-slate-500"}`} />
                              {isCurrent && (
                                <span className="absolute -top-1.5 -right-1.5 flex h-3.5 w-3.5">
                                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
                                  <span className="relative inline-flex rounded-full h-3.5 w-3.5 bg-amber-500"></span>
                                </span>
                              )}
                            </div>
                            <h4 className={`text-xs font-bold font-mono ${isCurrent ? "text-slate-950" : "text-slate-200"}`}>
                              {stateConf.title}
                            </h4>
                            <p className={`text-[9px] leading-tight mt-1 ${isCurrent ? "text-slate-900 font-medium" : "text-slate-500"}`}>
                              {stateConf.desc.substring(0, 48)}...
                            </p>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Interactive Controls Panel */}
                  <div className="border-t border-slate-800/80 pt-4 bg-slate-900/30 -mx-6 -mb-6 p-6 rounded-b-2xl">
                    <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider mb-3 flex items-center gap-1.5">
                      <Terminal className="w-3.5 h-3.5 text-amber-500" />
                      Gatilhos de Eventos & Sinais
                    </h4>
                    
                    <div className="flex flex-wrap gap-2.5">
                      
                      {simState === "Boot" && (
                        <button
                          onClick={handleTriggerLoading}
                          className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-amber-500 hover:bg-amber-400 text-slate-950 transition-all flex items-center gap-1.5 shadow"
                        >
                          <Play className="w-3.5 h-3.5 fill-current" />
                          Iniciar Boot ➔ Carregar
                        </button>
                      )}

                      {simState === "Loading" && (
                        <button
                          onClick={() => {}}
                          disabled={isSimulatingLoading}
                          className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 border border-slate-700 text-slate-300 flex items-center gap-2"
                        >
                          <RefreshCw className={`w-3.5 h-3.5 ${isSimulatingLoading ? "animate-spin" : ""}`} />
                          {isSimulatingLoading ? `Carregando Assets... (${loadingPercent}%)` : "Carregando..."}
                        </button>
                      )}

                      {simState === "Menu" && (
                        <div className="flex items-center gap-2 w-full sm:w-auto">
                          <input
                            type="text"
                            value={simUsername}
                            onChange={(e) => setSimUsername(e.target.value)}
                            placeholder="Nome do Personagem"
                            className="bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1 text-xs text-amber-300 font-mono focus:outline-none focus:border-amber-500/50 w-36"
                          />
                          <button
                            onClick={handleTriggerConnect}
                            className="px-3 py-1 text-xs font-semibold bg-indigo-500 hover:bg-indigo-400 text-slate-950 transition-all rounded-lg shadow flex items-center gap-1"
                          >
                            <Wifi className="w-3.5 h-3.5" />
                            Simular Login
                          </button>
                        </div>
                      )}

                      {simState === "Connecting" && (
                        <div className="flex items-center gap-2">
                          <button
                            onClick={handleConnectSuccess}
                            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-all flex items-center gap-1 shadow"
                          >
                            <Check className="w-3.5 h-3.5" />
                            Sucesso TCP (Conectado)
                          </button>
                          <button
                            onClick={handleConnectFailure}
                            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-500 hover:bg-rose-400 text-slate-950 transition-all flex items-center gap-1 shadow"
                          >
                            <WifiOff className="w-3.5 h-3.5" />
                            Falha TCP (Offline)
                          </button>
                        </div>
                      )}

                      {simState === "CharacterSelection" && (
                        <button
                          onClick={handleSelectCharacter}
                          className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-pink-500 hover:bg-pink-400 text-slate-950 transition-all flex items-center gap-1.5 shadow"
                        >
                          <UserCheck className="w-3.5 h-3.5" />
                          Confirmar Seleção ({simUsername})
                        </button>
                      )}

                      {simState === "InGame" && (
                        <div className="flex flex-wrap gap-2">
                          <button
                            onClick={handleSendHeartbeat}
                            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-amber-500 hover:bg-amber-400 text-slate-950 transition-all flex items-center gap-1.5 shadow"
                          >
                            <Activity className="w-3.5 h-3.5 animate-pulse" />
                            Enviar Heartbeat (KeepAlive)
                          </button>
                          <button
                            onClick={handleNetworkLoss}
                            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-500 hover:bg-rose-400 text-slate-950 transition-all flex items-center gap-1.5 shadow"
                          >
                            <WifiOff className="w-3.5 h-3.5" />
                            Interromper Rede
                          </button>
                        </div>
                      )}

                      {simState === "Disconnected" && (
                        <span className="text-xs text-rose-400 italic font-mono flex items-center gap-2 py-1">
                          <WifiOff className="w-4 h-4 animate-bounce" />
                          Simulando buffer de desconexão... Aguarde redirecionamento para o Menu.
                        </span>
                      )}

                      {simState === "Shutdown" && (
                        <button
                          onClick={handleTriggerBoot}
                          className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-zinc-700 hover:bg-zinc-600 text-white transition-all flex items-center gap-1.5"
                        >
                          <RefreshCw className="w-3.5 h-3.5" />
                          Reiniciar Jogo (Boot)
                        </button>
                      )}

                      {/* Global shutdown bypass */}
                      {simState !== "Shutdown" && simState !== "Disconnected" && (
                        <button
                          onClick={handleTriggerShutdown}
                          className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 hover:text-rose-400 border border-slate-700 text-slate-400 transition-all flex items-center gap-1.5"
                        >
                          <LogOut className="w-3.5 h-3.5" />
                          Fechar Jogo (Shutdown)
                        </button>
                      )}

                    </div>

                    {/* Architectural Patches Interactive Tests */}
                    <div className="mt-5 pt-4 border-t border-slate-800/60">
                      <h5 className="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-3 flex items-center gap-1.5">
                        <Shield className="w-3.5 h-3.5 text-violet-400" />
                        Testes Interativos dos Patches C# (DI, Crash & Packet)
                      </h5>
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                        <button
                          onClick={handleRegistryQuery}
                          className="px-2.5 py-1.5 rounded-lg text-[10px] font-bold bg-slate-800/80 hover:bg-slate-700 hover:text-amber-400 text-slate-300 transition-all border border-slate-700/40 flex items-center justify-center gap-1"
                        >
                          <Database className="w-3.5 h-3.5 text-amber-500" />
                          Consultar DI Registry
                        </button>
                        <button
                          onClick={handleInspectPacket}
                          className="px-2.5 py-1.5 rounded-lg text-[10px] font-bold bg-slate-800/80 hover:bg-slate-700 hover:text-amber-400 text-slate-300 transition-all border border-slate-700/40 flex items-center justify-center gap-1"
                        >
                          <Network className="w-3.5 h-3.5 text-sky-400" />
                          Inspecionar Packet
                        </button>
                        <button
                          onClick={() => handleSimulateCrash(true)}
                          className="px-2.5 py-1.5 rounded-lg text-[10px] font-bold bg-slate-800/80 hover:bg-slate-700 hover:text-amber-400 text-slate-300 transition-all border border-slate-700/40 flex items-center justify-center gap-1"
                        >
                          <Activity className="w-3.5 h-3.5 text-emerald-400 animate-pulse" />
                          Crash Recuperável
                        </button>
                        <button
                          onClick={() => handleSimulateCrash(false)}
                          className="px-2.5 py-1.5 rounded-lg text-[10px] font-bold bg-slate-800/80 hover:bg-slate-700 hover:text-rose-400 text-slate-300 transition-all border border-slate-700/40 flex items-center justify-center gap-1"
                        >
                          <WifiOff className="w-3.5 h-3.5 text-rose-500" />
                          Crash Fatal (Shutdown)
                        </button>
                      </div>
                    </div>

                  </div>
                </div>

                {/* Simulated Godot Terminal Log Stream */}
                <div className="lg:col-span-5 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 backdrop-blur-sm shadow-xl flex flex-col h-[400px]">
                  <div className="flex items-center justify-between border-b border-slate-800 pb-3 mb-3">
                    <div className="flex items-center gap-2">
                      <Terminal className="w-4 h-4 text-emerald-400" />
                      <h3 className="font-bold text-xs tracking-wider uppercase text-slate-300">
                        Console Godot Output (GD.Print)
                      </h3>
                    </div>
                    
                    {/* Log filter toggles */}
                    <div className="flex items-center gap-1">
                      <button
                        onClick={handleClearLogs}
                        className="text-[10px] hover:text-white text-slate-400 bg-slate-800 px-2 py-1 rounded hover:bg-slate-700 transition"
                      >
                        Limpar
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 mb-2 bg-slate-950/80 p-1.5 rounded-lg border border-slate-800">
                    <span className="text-[10px] text-slate-400 font-semibold px-2">Filtros:</span>
                    {(["all", "info", "warning", "error"] as const).map(type => (
                      <button
                        key={type}
                        onClick={() => setLogFilter(type)}
                        className={`text-[9px] px-2 py-0.5 rounded uppercase font-semibold tracking-wider transition ${
                          logFilter === type
                            ? "bg-slate-800 text-amber-400 border border-slate-700/50"
                            : "text-slate-500 hover:text-slate-300"
                        }`}
                      >
                        {type === "all" ? "Todos" : type}
                      </button>
                    ))}
                  </div>

                  {/* Logs stream body */}
                  <div className="flex-1 bg-slate-950/95 rounded-xl border border-slate-900 p-3 overflow-y-auto font-mono text-[10px] leading-relaxed space-y-1.5 custom-scrollbar">
                    {filteredLogs.length === 0 ? (
                      <div className="h-full flex items-center justify-center text-slate-500 italic">
                        Nenhum log gravado neste filtro. Interaja com o simulador.
                      </div>
                    ) : (
                      filteredLogs.map(log => {
                        let textClass = "text-slate-300";
                        if (log.type === "warning") textClass = "text-yellow-400/90";
                        if (log.type === "error") textClass = "text-rose-400 font-semibold";
                        return (
                          <div key={log.id} className="flex gap-2 items-start border-b border-slate-900/50 pb-1">
                            <span className="text-slate-500 select-none shrink-0">{log.timestamp}</span>
                            <span className="text-slate-600 font-semibold shrink-0 select-none">
                              {log.type === "error" ? "[ERR]" : log.type === "warning" ? "[WARN]" : "[INFO]"}
                            </span>
                            <span className={`${textClass} break-words`}>{log.text}</span>
                          </div>
                        );
                      })
                    )}
                    <div ref={consoleEndRef} />
                  </div>
                </div>

              </div>

              {/* SPRINT 4 TASK 2 - ENDGAME PVE & DUNGEONS PANEL */}
              <div className="xl:col-span-12 bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 backdrop-blur-md shadow-xl flex flex-col gap-6 mt-4">
                <div className="flex flex-col md:flex-row items-start md:items-center justify-between border-b border-slate-800/80 pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-rose-500/10 rounded-xl border border-rose-500/20">
                      <Sword className="w-6 h-6 text-rose-500" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-slate-100 tracking-tight">
                        Sistema de PvE Endgame, Masmorras e World Bosses
                      </h3>
                      <p className="text-xs text-slate-400">
                        Painel interativo de validação das regras autoritativas e do protocolo binário do MMORPG.
                      </p>
                    </div>
                  </div>
                  <span className="text-[10px] bg-rose-500/10 text-rose-400 border border-rose-500/20 px-3 py-1 rounded-full font-mono font-bold tracking-wider mt-2 md:mt-0 uppercase">
                    Sprint 4 — Task 2
                  </span>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Dungeon Matchmaking & Entry / World Boss Setup */}
                  <div className="lg:col-span-4 flex flex-col gap-4">
                    {/* Instanced Dungeons Control Card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4">
                      <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                        <Map className="w-4 h-4 text-emerald-400" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Masmorras Instanciadas</h4>
                      </div>

                      <div className="flex flex-col gap-2">
                        <label className="text-[11px] text-slate-400 font-semibold">Selecione a Masmorra:</label>
                        <select 
                          value={dungeonId} 
                          onChange={(e) => setDungeonId(e.target.value as any)}
                          disabled={dungeonActive}
                          className="bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 outline-none focus:border-emerald-500/40"
                        >
                          <option value="crypt_of_shadows">Cripta das Sombras (Level 45 - Solo/Party)</option>
                          <option value="dragon_lair">Covil do Dragão (Level 50 - Raid)</option>
                        </select>
                      </div>

                      <div className="flex flex-col gap-2">
                        <label className="text-[11px] text-slate-400 font-semibold">Modo de Jogo:</label>
                        <div className="grid grid-cols-3 gap-2">
                          {(["solo", "party", "raid"] as const).map(m => (
                            <button
                              key={m}
                              disabled={dungeonActive}
                              onClick={() => setDungeonMode(m)}
                              className={`py-1.5 rounded-lg text-[10px] font-bold uppercase transition border ${
                                dungeonMode === m 
                                  ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30" 
                                  : "bg-slate-900 text-slate-400 border-slate-800 hover:border-slate-700"
                              }`}
                            >
                              {m}
                            </button>
                          ))}
                        </div>
                      </div>

                      {!dungeonActive ? (
                        <button
                          onClick={handleEnterDungeonSim}
                          className="w-full mt-2 py-2 bg-emerald-500 hover:bg-emerald-400 text-slate-950 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 shadow-lg"
                        >
                          <Play className="w-3.5 h-3.5 fill-current" />
                          Entrar na Masmorra (CS_DUNGEON_ENTER)
                        </button>
                      ) : (
                        <button
                          onClick={handleLeaveDungeonSim}
                          className="w-full mt-2 py-2 bg-rose-500 hover:bg-rose-400 text-slate-950 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2"
                        >
                          <XCircle className="w-3.5 h-3.5" />
                          Sair da Masmorra (CS_DUNGEON_LEAVE)
                        </button>
                      )}
                    </div>

                    {/* World Boss Launcher Card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Compass className="w-4 h-4 text-rose-500" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">World Boss Público</h4>
                        </div>
                        <span className={`text-[9px] font-bold px-2 py-0.5 rounded ${worldBossActive ? "bg-red-500/10 text-red-400 border border-red-500/20" : "bg-slate-800 text-slate-400"}`}>
                          {worldBossActive ? "ATIVO" : "AGUARDANDO"}
                        </span>
                      </div>

                      <p className="text-[11px] text-slate-400 leading-relaxed">
                        World Bosses nascem em áreas abertas e possuem tabelas de agressividade e recompensas ranqueadas.
                      </p>

                      {worldBossActive ? (
                        <div className="flex flex-col gap-2">
                          <div className="flex justify-between items-center text-[10px] text-slate-300 font-mono">
                            <span>Behemoth do Caos</span>
                            <span>{worldBossHp} / {worldBossMaxHp} HP</span>
                          </div>
                          <div className="w-full bg-slate-900 h-2 rounded-full overflow-hidden border border-slate-800">
                            <div 
                              className="bg-red-600 h-full transition-all duration-300"
                              style={{ width: `${(worldBossHp / worldBossMaxHp) * 100}%` }}
                            />
                          </div>

                          <button
                            onClick={handleAttackWorldBoss}
                            className="w-full mt-2 py-2 bg-rose-500/20 hover:bg-rose-500/30 text-rose-300 border border-rose-500/30 hover:border-rose-500/50 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5"
                          >
                            <Sword className="w-3.5 h-3.5 animate-pulse" />
                            Atacar World Boss
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={handleSpawnWorldBoss}
                          className="w-full py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5 border border-slate-700"
                        >
                          <Activity className="w-3.5 h-3.5 text-amber-500 animate-spin" />
                          Forçar Spawn do Boss (CS_WORLD_BOSS_SPAWN_REQ)
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Middle Column: Active Dungeon Combat & Boss AI Phases */}
                  <div className="lg:col-span-5 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 flex flex-col gap-4 flex-1">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
                        <div className="flex items-center gap-2">
                          <Activity className="w-4 h-4 text-rose-500 animate-pulse" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Instância Ativa</h4>
                        </div>
                        {dungeonActive && (
                          <span className="text-[11px] text-slate-300 font-mono flex items-center gap-1.5 bg-slate-900 px-2 py-0.5 rounded border border-slate-800">
                            <Clock className="w-3 h-3 text-amber-400" />
                            {Math.floor(dungeonTimeLeft / 60)}:{(dungeonTimeLeft % 60).toString().padStart(2, "0")}
                          </span>
                        )}
                      </div>

                      {dungeonActive ? (
                        <div className="flex flex-col gap-4">
                          {/* Metadata row */}
                          <div className="grid grid-cols-3 gap-2 bg-slate-900/60 p-2.5 rounded-lg border border-slate-900 text-center text-[10px] font-mono">
                            <div>
                              <span className="text-slate-500 block">ID Instância</span>
                              <span className="text-slate-300 font-bold">inst_{dungeonId}_1</span>
                            </div>
                            <div>
                              <span className="text-slate-500 block">Offset Z</span>
                              <span className="text-amber-400 font-bold">{dungeonOffset.toFixed(1)}m</span>
                            </div>
                            <div>
                              <span className="text-slate-500 block">Checkpoint</span>
                              <span className="text-emerald-400 font-bold">{dungeonCheckpoint}</span>
                            </div>
                          </div>

                          {/* Boss Health and Info */}
                          <div className="bg-slate-900/30 p-4 rounded-xl border border-slate-800/50 flex flex-col gap-3 relative overflow-hidden">
                            {/* Boss Enraged Highlight */}
                            {bossEnraged && (
                              <div className="absolute inset-0 bg-red-950/25 border border-red-500/30 animate-pulse pointer-events-none" />
                            )}

                            <div className="flex justify-between items-start">
                              <div>
                                <h5 className="font-bold text-xs text-slate-200">
                                  {dungeonId === "crypt_of_shadows" ? "Lorde das Sombras (Boss)" : "Rei Dragão (Boss)"}
                                </h5>
                                <span className="text-[10px] text-slate-400 font-mono">
                                  Fase Ativa: <strong className="text-amber-400">{bossPhase}</strong> de 3
                                </span>
                              </div>
                              <div className="text-right">
                                <span className="text-xs font-mono text-slate-200 block">
                                  {bossHp} / {bossMaxHp} HP
                                </span>
                                {bossEnraged && (
                                  <span className="text-[9px] bg-red-500 text-slate-950 font-bold px-1.5 py-0.5 rounded uppercase animate-pulse">
                                    ENFURECIDO
                                  </span>
                                )}
                              </div>
                            </div>

                            {/* Boss HP Bar */}
                            <div className="w-full bg-slate-950 h-3 rounded-full overflow-hidden border border-slate-800 relative">
                              <div 
                                className={`h-full transition-all duration-300 ${bossEnraged ? "bg-red-600" : "bg-rose-500"}`}
                                style={{ width: `${(bossHp / bossMaxHp) * 100}%` }}
                              />
                            </div>

                            {/* Telegraph Warning */}
                            {bossTelegraphActive && (
                              <div className="bg-amber-950/40 border border-amber-500/40 rounded-lg p-2.5 flex items-center justify-between text-amber-300 animate-pulse">
                                <div className="flex items-center gap-2">
                                  <AlertTriangle className="w-4 h-4 text-amber-500 animate-bounce" />
                                  <span className="text-[10px] font-bold uppercase tracking-wider">Ataque Telegrafado Área: Terremoto Sombrio!</span>
                                </div>
                                <span className="text-xs font-mono font-bold bg-amber-500 text-slate-950 px-2 py-0.5 rounded">
                                  {bossTelegraphTimer}s
                                </span>
                              </div>
                            )}

                            {/* Adds spawned row */}
                            {bossAdds.length > 0 && (
                              <div className="mt-2 flex flex-col gap-1.5 border-t border-slate-800/40 pt-2">
                                <span className="text-[9px] font-bold text-rose-400 uppercase tracking-widest">Servos de Proteção Ativos:</span>
                                {bossAdds.map((add, idx) => (
                                  <div key={add.id} className="flex justify-between items-center bg-slate-950 px-2.5 py-1.5 rounded border border-slate-900 text-[10px]">
                                    <span className="text-slate-400 font-mono">Servo #{idx+1}</span>
                                    <div className="flex items-center gap-2">
                                      <span className="text-slate-300 font-mono">{add.hp}/{add.maxHp} HP</span>
                                      <div className="w-16 bg-slate-900 h-1.5 rounded-full overflow-hidden border border-slate-800">
                                        <div className="bg-rose-500 h-full" style={{ width: `${(add.hp / add.maxHp) * 100}%` }} />
                                      </div>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            )}

                            {/* Combat Controls */}
                            {bossHp > 0 ? (
                              <div className="grid grid-cols-2 gap-2 mt-2 z-10">
                                <button
                                  onClick={handleAttackBossSim}
                                  className="py-2 bg-rose-600 hover:bg-rose-500 text-slate-950 font-bold rounded-lg text-xs transition flex items-center justify-center gap-1.5 shadow"
                                >
                                  <Sword className="w-3.5 h-3.5" />
                                  {bossAdds.length > 0 ? "Atacar Servo" : "Atacar Boss"}
                                </button>
                                <div className="bg-slate-900/80 rounded-lg border border-slate-800/50 px-3 py-1 text-center flex flex-col justify-center">
                                  <span className="text-[9px] text-slate-500 uppercase font-bold font-mono">Enfurece Em</span>
                                  <span className="text-xs font-mono font-bold text-amber-500">
                                    {bossEnraged ? "ATIVO (5x Dano)" : `${30 - bossCombatSeconds}s`}
                                  </span>
                                </div>
                              </div>
                            ) : (
                              <div className="bg-emerald-950/20 border border-emerald-500/20 rounded-lg p-3 text-center text-emerald-400 text-xs font-bold">
                                🎉 Masmorra concluída! Chefe derrotado. Veja o loot abaixo!
                              </div>
                            )}

                            {/* Anti-Reset exploit notification */}
                            {bossResetTimer !== null && (
                              <div className="bg-indigo-950/40 border border-indigo-500/30 rounded-lg p-2 text-center text-indigo-300 text-[10px] font-mono">
                                🛡️ Anti-Reset Ativo: Boss restaurará vida cheia em {bossResetTimer}s se sair do combate.
                              </div>
                            )}
                          </div>
                        </div>
                      ) : (
                        <div className="flex-1 flex flex-col items-center justify-center text-center p-6 text-slate-500">
                          <Compass className="w-12 h-12 text-slate-700 mb-2 animate-pulse" />
                          <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Nenhuma Masmorra Ativa</span>
                          <p className="text-[11px] text-slate-600 leading-relaxed mt-1 max-w-[240px]">
                            Selecione uma masmorra e modo à esquerda e clique em Entrar para iniciar a simulação autoritativa.
                          </p>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Right Column: Loot reservation and DB audit logs */}
                  <div className="lg:col-span-3 flex flex-col gap-4">
                    {/* Loot Reservation Claim Card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                        <Gift className="w-4 h-4 text-amber-400" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Loot Reservado (PATCH 3)</h4>
                      </div>

                      {lootReservations.length === 0 ? (
                        <p className="text-[11px] text-slate-500 italic text-center py-4">
                          Nenhuma recompensa ativa para resgatar. Derrote chefes para dropar loots temporários.
                        </p>
                      ) : (
                        <div className="flex flex-col gap-2">
                          {lootReservations.map(res => (
                            <div 
                              key={res.id} 
                              className={`p-2.5 rounded-lg border text-xs flex flex-col gap-2 transition-all ${
                                res.claimed 
                                  ? "bg-slate-900/40 border-slate-800/40 opacity-60" 
                                  : "bg-slate-900 border-slate-800"
                              }`}
                            >
                              <div className="flex justify-between items-center font-mono">
                                <span className={`font-bold ${res.itemID.includes("epic") ? "text-purple-400" : "text-amber-400"}`}>
                                  {res.qty}x {res.itemID.replace(/_/g, " ").toUpperCase()}
                                </span>
                                {!res.claimed && (
                                  <span className="text-[10px] text-amber-500 font-bold bg-amber-500/10 border border-amber-500/20 px-1.5 py-0.5 rounded">
                                    {res.timer}s
                                  </span>
                                )}
                              </div>
                              
                              {!res.claimed ? (
                                <button
                                  onClick={() => handleClaimLootSim(res.id, res.itemID, res.qty)}
                                  className="w-full py-1 bg-amber-500 hover:bg-amber-400 text-slate-950 rounded font-bold text-[10px] uppercase transition"
                                >
                                  Reivindicar Drop (CS_DUNGEON_CLAIM_LOOT)
                                </button>
                              ) : (
                                <span className="text-center text-[10px] text-emerald-400 font-bold uppercase flex items-center justify-center gap-1">
                                  <Check className="w-3.5 h-3.5" /> Reivindicado
                                </span>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>

                    {/* PostgreSQL state viewer card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-2 flex-1">
                      <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                        <Database className="w-4 h-4 text-violet-400" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Persistência PostgreSQL</h4>
                      </div>

                      <div className="flex-1 overflow-y-auto max-h-[160px] text-[10px] font-mono space-y-2 pr-1 custom-scrollbar">
                        {persistentLogs.map(log => {
                          let labelColor = "bg-slate-800 text-slate-400 border-slate-700";
                          if (log.type === "Lockout") labelColor = "bg-rose-500/10 text-rose-400 border-rose-500/20";
                          if (log.type === "Checkpoint") labelColor = "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
                          if (log.type === "Audit Log") labelColor = "bg-purple-500/10 text-purple-400 border-purple-500/20";
                          return (
                            <div key={log.id} className="border border-slate-900 p-2 rounded bg-slate-900/40 flex flex-col gap-1">
                              <div className="flex justify-between items-center text-[9px]">
                                <span className={`px-1.5 py-0.2 rounded border font-bold uppercase ${labelColor}`}>
                                  {log.type}
                                </span>
                                <span className="text-slate-500">{log.timestamp}</span>
                              </div>
                              <span className="text-slate-300 leading-normal">{log.details}</span>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* CANONICAL MONSTER AI BIBLE SIMULATOR PANEL (MONSTER AI BIBLE PATCH) */}
              <div className="xl:col-span-12 bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 backdrop-blur-md shadow-xl flex flex-col gap-6 mt-4">
                <div className="flex flex-col md:flex-row items-start md:items-center justify-between border-b border-slate-800/80 pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-rose-500/10 rounded-xl border border-rose-500/20">
                      <Terminal className="w-6 h-6 text-rose-400 animate-pulse" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-slate-100 tracking-tight">
                        Bíblia de Inteligência Artificial de Monstros (Complexity Tier B)
                      </h3>
                      <p className="text-xs text-slate-400">
                        Simulador interativo autoritativo das diretrizes de Ameaça Híbrida, Modificadores de Instinto, Leash Dinâmico e Solo Pull.
                      </p>
                    </div>
                  </div>
                  <span className="text-[10px] bg-rose-500/10 text-rose-400 border border-rose-500/20 px-3 py-1 rounded-full font-mono font-bold tracking-wider mt-2 md:mt-0 uppercase">
                    Monster AI Patch
                  </span>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Archetypes & Hybrid Threat Focus */}
                  <div className="lg:col-span-4 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Layers className="w-4 h-4 text-rose-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Arquétipos e Instinto</h4>
                        </div>
                      </div>

                      <div className="space-y-2">
                        <label className="text-[11px] text-slate-400 font-semibold">Selecione o Arquétipo:</label>
                        <select
                          value={selectedAiArchetype}
                          onChange={(e) => setSelectedAiArchetype(e.target.value as any)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 outline-none focus:border-rose-500/40 border-slate-700"
                        >
                          <option value="BRUISER_TANK">Bruiser/Tank (80% Threat / 20% Instinct)</option>
                          <option value="PREDATOR_ASSASSIN">Predator/Assassin (40% Threat / 60% Instinct)</option>
                          <option value="RANGED_HUNTER">Ranged Hunter (50% Threat / 50% Instinct)</option>
                          <option value="CASTER">Caster (50% Threat / 50% Instinct)</option>
                          <option value="SUPPORT">Support (50% Threat / 50% Instinct)</option>
                          <option value="ELITE_COMMANDER">Elite Commander (50% Threat / 50% Instinct)</option>
                        </select>
                      </div>

                      {/* Display Archetype Info */}
                      <div className="p-3 bg-slate-900/40 rounded-lg border border-slate-800/50 space-y-1.5 text-xs leading-normal">
                        <div className="flex justify-between items-center text-[10px] font-mono">
                          <span className="text-rose-400 uppercase font-bold">Distribuição de Peso:</span>
                          <span className="text-slate-300">
                            {selectedAiArchetype === "BRUISER_TANK" ? "80% Threat | 20% Instinct" : 
                             selectedAiArchetype === "PREDATOR_ASSASSIN" ? "40% Threat | 60% Instinct" : 
                             "50% Threat | 50% Instinct"}
                          </span>
                        </div>
                        <p className="text-[11px] text-slate-300">
                          <strong>Comportamento:</strong> {
                            selectedAiArchetype === "BRUISER_TANK" ? "Foca agressivamente no tanque de linha de frente, absorvendo dano e bloqueando a mobilidade dos atacantes." :
                            selectedAiArchetype === "PREDATOR_ASSASSIN" ? "Ignora parcialmente o aggro do tanque para caçar alvos com baixa HP, armadura fina ou conjuradores na retaguarda." :
                            selectedAiArchetype === "RANGED_HUNTER" ? "Mantém distância ideal de 6 a 12 metros, kitando atacantes e reposicionando-se sempre que ameaçado." :
                            selectedAiArchetype === "CASTER" ? "Executa rotações de feitiços de área, focando em agrupamentos e alvos com baixa resistência mágica." :
                            selectedAiArchetype === "SUPPORT" ? "Prioriza conjurar curas, escudos e buffs em aliados de alta prioridade (Comandante > Bruiser > Próximo)." :
                            "Coordena os monstros ao redor, aplicando auras ativas de velocidade, ferocidade ou terror na área."
                          }
                        </p>
                      </div>

                      {/* Threat Adjustment Map */}
                      <div className="space-y-3 pt-2">
                        <div className="flex justify-between items-center">
                          <span className="text-[11px] text-slate-400 font-semibold">Tabela de Ameaça Híbrida:</span>
                          <button
                            onClick={() => setAiThreatMap({ "Gabriela_Paladin": 300, "AI_Bot_Mage": 200, "AI_Bot_Priest": 150 })}
                            className="text-[9px] text-slate-500 hover:text-slate-300 font-mono transition"
                          >
                            Resetar Ameaça
                          </button>
                        </div>

                        <div className="space-y-2">
                          {[
                            { id: "Gabriela_Paladin", label: "Gabriela (Paladin/Tank)", role: "Frontline Tank", baseColor: "border-sky-500/20" },
                            { id: "AI_Bot_Mage", label: "AI Mage (Caster)", role: "Ranged Damage", baseColor: "border-purple-500/20" },
                            { id: "AI_Bot_Priest", label: "AI Priest (Healer)", role: "Ranged Support", baseColor: "border-emerald-500/20" }
                          ].map(p => {
                            const rawThreat = aiThreatMap[p.id] || 0;
                            const score = getAggroFocusScore(selectedAiArchetype, p.id, rawThreat);
                            
                            // Determine highest focus score to highlight Active Target
                            const allScores = [
                              getAggroFocusScore(selectedAiArchetype, "Gabriela_Paladin", aiThreatMap["Gabriela_Paladin"] || 0),
                              getAggroFocusScore(selectedAiArchetype, "AI_Bot_Mage", aiThreatMap["AI_Bot_Mage"] || 0),
                              getAggroFocusScore(selectedAiArchetype, "AI_Bot_Priest", aiThreatMap["AI_Bot_Priest"] || 0)
                            ];
                            const maxScore = Math.max(...allScores);
                            const isActiveTarget = score === maxScore;

                            return (
                              <div
                                key={p.id}
                                className={`p-2.5 rounded-lg border transition bg-slate-900 flex flex-col gap-1 ${
                                  isActiveTarget ? "border-rose-500/50 ring-1 ring-rose-500/20 shadow-lg shadow-rose-950/20" : "border-slate-800"
                                }`}
                              >
                                <div className="flex justify-between items-center text-xs">
                                  <div className="flex flex-col">
                                    <span className="font-bold text-slate-200">{p.label}</span>
                                    <span className="text-[10px] text-slate-500 font-mono">{p.role}</span>
                                  </div>
                                  <div className="flex items-center gap-2">
                                    <div className="text-right">
                                      <span className="text-[9px] text-slate-500 block font-mono">Foco de IA</span>
                                      <span className={`font-mono font-bold ${isActiveTarget ? "text-rose-400" : "text-slate-400"}`}>{score} pts</span>
                                    </div>
                                    {isActiveTarget && (
                                      <span className="text-[9px] bg-rose-500/10 text-rose-400 border border-rose-500/20 px-1.5 py-0.5 rounded font-bold uppercase animate-pulse">
                                        ALVO ATIVO
                                      </span>
                                    )}
                                  </div>
                                </div>

                                <div className="flex gap-1.5 mt-1.5">
                                  <button
                                    onClick={() => setAiThreatMap(prev => ({ ...prev, [p.id]: (prev[p.id] || 0) + 100 }))}
                                    className="flex-1 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 text-[9px] rounded font-bold border border-slate-700 transition"
                                  >
                                    +100 Dano (+100 Threat)
                                  </button>
                                  {p.id === "AI_Bot_Priest" && (
                                    <button
                                      onClick={() => setAiThreatMap(prev => ({ ...prev, [p.id]: (prev[p.id] || 0) + 200 }))}
                                      className="flex-1 py-1 bg-emerald-950/40 hover:bg-emerald-900/40 text-emerald-300 text-[9px] rounded font-bold border border-emerald-800/40 transition"
                                    >
                                      +400 Cura (+200 Threat)
                                    </button>
                                  )}
                                  {p.id === "Gabriela_Paladin" && (
                                    <button
                                      onClick={() => setAiThreatMap(prev => ({ ...prev, [p.id]: (prev[p.id] || 0) + 300 }))}
                                      className="flex-1 py-1 bg-sky-950/40 hover:bg-sky-900/40 text-sky-300 text-[9px] rounded font-bold border border-sky-800/40 transition"
                                    >
                                      Provocar (+300 Threat)
                                    </button>
                                  )}
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Middle Column: Leash & Pack Solo Pull Systems */}
                  <div className="lg:col-span-5 flex flex-col gap-4">
                    {/* Leash Simulator Card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Compass className="w-4 h-4 text-rose-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Leash Dinâmico e Desengajamento</h4>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-[10px] text-slate-400 font-semibold block">Categoria de Leash:</label>
                          <select
                            value={selectedLeashTier}
                            onChange={(e) => setSelectedLeashTier(parseInt(e.target.value) as any)}
                            className="w-full bg-slate-900 border border-slate-800 rounded-lg p-2 text-[11px] text-slate-200 outline-none focus:border-rose-500/40"
                          >
                            <option value={1}>Tier 1 — Passive (6-12m)</option>
                            <option value={2}>Tier 2 — Standard (15-30m)</option>
                            <option value={3}>Tier 3 — Hunter (40-80m)</option>
                            <option value={4}>Tier 4 — Commander (100-200m)</option>
                            <option value={5}>Tier 5 — Apex (Sem limite)</option>
                          </select>
                        </div>

                        <div className="space-y-1">
                          <label className="text-[10px] text-slate-400 font-semibold block">Distância do Spawn: <span className="text-rose-400 font-bold">{leashDistance}m</span></label>
                          <input
                            type="range"
                            min="0"
                            max="120"
                            value={leashDistance}
                            onChange={(e) => setLeashDistance(parseInt(e.target.value))}
                            className="w-full mt-2 accent-rose-500"
                          />
                        </div>
                      </div>

                      {/* Display Leash Evaluation */}
                      {(() => {
                        const evaluation = getLeashStatus(leashDistance, selectedLeashTier);
                        return (
                          <div className={`p-3 rounded-lg border text-xs leading-normal flex flex-col gap-1.5 ${evaluation.color}`}>
                            <div className="flex justify-between items-center font-bold">
                              <span className="uppercase tracking-wider font-mono text-[10px]">
                                Estado do Leash:
                              </span>
                              <span>{evaluation.leashed ? "LEASH BREAK" : "SOB PERSEGUIÇÃO"}</span>
                            </div>
                            <p className="text-[11px] opacity-90">{evaluation.text}</p>
                          </div>
                        );
                      })()}
                    </div>

                    {/* Pack Aggro / Solo Pull Card */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                        <Activity className="w-4 h-4 text-emerald-400" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Pack Aggro: Filosofia Solo Pull</h4>
                      </div>

                      <p className="text-[11px] text-slate-400 leading-relaxed">
                        Ao contrário de jogos com social aggro passivo em massa, em <strong>Light and Shadow</strong> a agressividade é individual (<strong>Solo Aggro</strong>). Atrair uma criatura não aciona as vizinhas, permitindo combates significativos com baixa densidade de inimigos.
                      </p>

                      <div className="bg-slate-900 rounded-lg border border-slate-800/60 p-3 flex flex-col gap-2">
                        <div className="flex justify-between items-center text-[10px] font-mono">
                          <span className="text-emerald-400 uppercase font-bold">Lobo de Caça do Vazio:</span>
                          <span className="text-slate-500">Membros do Bando: 3</span>
                        </div>

                        <button
                          onClick={handleSimulateSoloPull}
                          className="w-full py-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5"
                        >
                          <Sword className="w-3.5 h-3.5" />
                          Simular Pull Individual
                        </button>

                        <div className="text-[10px] font-mono text-slate-400 text-center italic">
                          {soloPullStatus}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Right Column: Survival Flee & Special Laws */}
                  <div className="lg:col-span-3 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4 flex-1">
                      <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                        <Flame className="w-4 h-4 text-amber-500 animate-pulse" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Modos de Sobrevivência</h4>
                      </div>

                      <div className="space-y-2">
                        <label className="text-[11px] text-slate-400 font-semibold block">Selecione a Espécie / Classe:</label>
                        <select
                          value={selectedSurvivalSpecies}
                          onChange={(e) => setSelectedSurvivalSpecies(e.target.value as any)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1.5 text-xs text-slate-200 outline-none focus:border-rose-500/40 border-slate-700"
                        >
                          <option value="goblin">Goblin / Saqueador (Cowardly)</option>
                          <option value="assassin">Assassino das Sombras (Tactical Retreat)</option>
                          <option value="beast">Fera do Vazio (Frenzy Under Death)</option>
                          <option value="demon">Demônio Infernal (Death Before Retreat)</option>
                          <option value="dragon">Void Dragon Lord (Dragon Survival Law)</option>
                        </select>
                      </div>

                      <div className="space-y-1">
                        <div className="flex justify-between items-center">
                          <label className="text-[10px] text-slate-400 font-semibold">HP Atual do Monstro:</label>
                          <span className="text-xs font-mono font-bold text-rose-400">{fleeTestHp}%</span>
                        </div>
                        <input
                          type="range"
                          min="1"
                          max="100"
                          value={fleeTestHp}
                          onChange={(e) => setFleeTestHp(parseInt(e.target.value))}
                          className="w-full accent-rose-500"
                        />
                      </div>

                      {/* Display Survival Mode Results */}
                      {(() => {
                        const result = evaluateSurvivalBehavior(selectedSurvivalSpecies, fleeTestHp);
                        return (
                          <div className={`p-3 rounded-lg border text-xs leading-normal flex flex-col gap-2 ${result.color} flex-1`}>
                            <div className="flex justify-between items-center">
                              <span className="text-[10px] uppercase font-mono font-bold tracking-wider">
                                {result.mode}
                              </span>
                              <span className="text-[9px] border px-1.5 py-0.2 rounded font-bold font-mono border-rose-500/20">
                                {result.badge}
                              </span>
                            </div>
                            <p className="text-[11px] text-slate-300 leading-relaxed">
                              {result.description}
                            </p>
                          </div>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              </div>

              {/* SPRINT 3 TASK 5 - CHARACTER PROGRESSION & WORLD PROGRESSION SYSTEM */}
              <div className="xl:col-span-12 bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 backdrop-blur-md shadow-xl flex flex-col gap-6 mt-4">
                <div className="flex flex-col md:flex-row items-start md:items-center justify-between border-b border-slate-800/80 pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-amber-500/10 rounded-xl border border-amber-500/20">
                      <Sparkles className="w-6 h-6 text-amber-500" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-slate-100 tracking-tight">
                        Progressão de Personagem, Vocações & Portões Regionais
                      </h3>
                      <p className="text-xs text-slate-400">
                        Painel interativo de validação das regras autoritativas de vocações, subclasses e portais geográficos baseados em nível.
                      </p>
                    </div>
                  </div>
                  <span className="text-[10px] bg-amber-500/10 text-amber-400 border border-amber-500/20 px-3 py-1 rounded-full font-mono font-bold tracking-wider mt-2 md:mt-0 uppercase">
                    Sprint 3 — Task 5
                  </span>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Level Slider and Vocation Unlock */}
                  <div className="lg:col-span-4 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Settings className="w-4 h-4 text-amber-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Ajuste de Nível do Jogador</h4>
                        </div>
                        <span className="text-xs text-amber-400 font-mono font-bold bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded-md">
                          Nível {progLevel}
                        </span>
                      </div>

                      <div className="flex flex-col gap-3">
                        <p className="text-[11px] text-slate-400">
                          Seu personagem pode evoluir de <strong className="text-amber-400">Nível 1 a 9999</strong> (Progresso de Longo Prazo Sem Teto Rígido). Insira o nível diretamente ou mova o controle.
                        </p>
                        
                        <div className="flex items-center gap-2">
                          <input
                            type="number"
                            min="1"
                            max="9999"
                            value={progLevel}
                            onChange={(e) => {
                              let val = parseInt(e.target.value) || 1;
                              if (val < 1) val = 1;
                              if (val > 9999) val = 9999;
                              setProgLevel(val);
                              if (val < 10) {
                                setProgClass("Novice");
                              }
                              if (val < 100) {
                                setProgSubclass("");
                                setAwakenedAffinity("");
                              }
                            }}
                            className="w-20 bg-slate-900 border border-slate-800 rounded px-2.5 py-1 text-xs text-amber-400 font-mono text-center outline-none focus:border-amber-500/40"
                          />
                          <input
                            type="range"
                            min="1"
                            max="9999"
                            value={progLevel}
                            onChange={(e) => {
                              const val = parseInt(e.target.value);
                              setProgLevel(val);
                              // Reseta classe caso volte para nível < 10
                              if (val < 10) {
                                setProgClass("Novice");
                              }
                              // Reseta subclasse caso volte para nível < 100
                              if (val < 100) {
                                setProgSubclass("");
                                setAwakenedAffinity("");
                              }
                            }}
                            className="flex-1 accent-amber-500 cursor-pointer h-1.5 bg-slate-800 rounded-lg appearance-none"
                          />
                        </div>

                        {/* Quick Set Levels */}
                        <div className="flex flex-wrap gap-1 mt-1 border-t border-slate-800/50 pt-2">
                          <span className="text-[9px] text-slate-500 font-bold uppercase block w-full mb-1">Setar Nível de Personagem:</span>
                          {[1, 10, 50, 100, 150, 9999].map(lvl => (
                            <button
                              key={lvl}
                              onClick={() => {
                                setProgLevel(lvl);
                                if (lvl < 10) setProgClass("Novice");
                                if (lvl < 100) {
                                  setProgSubclass("");
                                  setAwakenedAffinity("");
                                }
                              }}
                              className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold transition ${progLevel === lvl ? "bg-amber-500 text-slate-950" : "bg-slate-900 text-slate-400 hover:text-slate-200 hover:bg-slate-850"}`}
                            >
                              Nvl {lvl}
                            </button>
                          ))}
                        </div>

                        {/* Developer Shortcuts for Affinity */}
                        <div className="flex flex-wrap gap-1 mt-1 border-t border-slate-800/50 pt-2">
                          <span className="text-[9px] text-slate-500 font-bold uppercase block w-full mb-1">Atalhos Admin de Afinidade (Setar Lvl 100):</span>
                          <button
                            onClick={() => { setAffinityFireLvl(100); setAffinityFireXp(0); }}
                            className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-red-950/40 border border-red-900 text-red-300 hover:bg-red-900/30"
                          >
                            🔥 Fogo 100
                          </button>
                          <button
                            onClick={() => { setAffinityIceLvl(100); setAffinityIceXp(0); }}
                            className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-cyan-950/40 border border-cyan-900 text-cyan-300 hover:bg-cyan-900/30"
                          >
                            ❄ Gelo 100
                          </button>
                          <button
                            onClick={() => { setAffinityHolyLvl(100); setAffinityHolyXp(0); }}
                            className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-amber-950/40 border border-amber-900 text-amber-300 hover:bg-amber-900/30"
                          >
                            ✨ Santo 100
                          </button>
                          <button
                            onClick={() => { setAffinityShadowLvl(100); setAffinityShadowXp(0); }}
                            className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-purple-950/40 border border-purple-900 text-purple-300 hover:bg-purple-900/30"
                          >
                            💀 Sombra 100
                          </button>
                          <button
                            onClick={() => { setAffinityNatureLvl(100); setAffinityNatureXp(0); }}
                            className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-emerald-950/40 border border-emerald-900 text-emerald-300 hover:bg-emerald-900/30"
                          >
                            🍃 Natureza 100
                          </button>
                        </div>

                        <div className="flex justify-between text-[10px] text-slate-500 font-mono mt-1">
                          <span>Nvl 1 (Novice)</span>
                          <span>Nvl 10 (Vocation)</span>
                          <span>Nvl 100 (Unlock)</span>
                          <span>Nvl 9999 (Cap)</span>
                        </div>
                      </div>
                    </div>

                    {/* Vocation Selection (Level 10) */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Lock className="w-4 h-4 text-violet-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Desbloqueio de Vocação (Level 10)</h4>
                        </div>
                        {progClass !== "Novice" ? (
                          <span className="text-[10px] text-emerald-400 font-bold bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-full uppercase tracking-wider">
                            Confirmado
                          </span>
                        ) : progLevel >= 10 ? (
                          <span className="text-[10px] text-amber-400 font-bold bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded-full uppercase tracking-wider">
                            Disponível
                          </span>
                        ) : (
                          <span className="text-[10px] text-slate-500 font-bold bg-slate-500/10 border border-slate-800 px-2 py-0.5 rounded-full uppercase tracking-wider">
                            Bloqueado
                          </span>
                        )}
                      </div>

                      {progLevel < 10 ? (
                        <div className="flex flex-col items-center justify-center py-6 text-center text-slate-500">
                          <Lock className="w-10 h-10 text-slate-700 mb-2" />
                          <p className="text-[11px] text-slate-400">
                            Fase de Aprendiz Ativa (Novice Phase).<br />Seu personagem precisa atingir o <strong className="text-amber-500">Nível 10</strong> para escolher uma vocação base.
                          </p>
                        </div>
                      ) : progClass !== "Novice" ? (
                        <div className="bg-slate-900/60 border border-slate-800 p-4 rounded-xl flex flex-col gap-3">
                          <div className="flex items-center gap-3">
                            <div className="p-2 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-emerald-400 font-bold text-sm">
                              {progClass[0]}
                            </div>
                            <div>
                              <h5 className="font-bold text-sm text-slate-200">{progClass}</h5>
                              <p className="text-[10px] text-slate-400">Seleção definitiva persistida no PostgreSQL.</p>
                            </div>
                          </div>
                          
                          <div className="border-t border-slate-800/80 pt-2.5">
                            <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block mb-1">Passivos Aplicados:</span>
                            <ul className="text-[11px] text-emerald-400 space-y-1 font-mono">
                              {progClass === "Knight" && (
                                <>
                                  <li>✦ +100 MaxHP (Constituição Forte)</li>
                                  <li>✦ +15 Defesa de Armadura</li>
                                </>
                              )}
                              {progClass === "Mage" && (
                                <>
                                  <li>✦ +80 MaxMana (Mente Brilhante)</li>
                                  <li>✦ +10% Dano de Ataque Elemental</li>
                                </>
                              )}
                              {progClass === "Archer" && (
                                <>
                                  <li>✦ +5% Chance de Acerto Crítico</li>
                                  <li>✦ +10 de Precisão com Armas</li>
                                </>
                              )}
                              {progClass === "Assassin" && (
                                <>
                                  <li>✦ +10 de Evasão (Passo Sombrio)</li>
                                  <li>✦ +5% Chance de Acerto Crítico</li>
                                </>
                              )}
                              {progClass === "Cleric" && (
                                <>
                                  <li>✦ +50 MaxHP (Proteção Sagrada)</li>
                                  <li>✦ +10 de Resistência Mágica</li>
                                  <li>✦ +5% Defesa Elemental</li>
                                </>
                              )}
                            </ul>
                          </div>
                        </div>
                      ) : (
                        <div className="flex flex-col gap-3">
                          <p className="text-[11px] text-slate-400 leading-relaxed">
                            Atenção: A escolha de vocação base é <strong className="text-rose-400 uppercase">irreversível</strong>. Escolha sua jornada com sabedoria:
                          </p>
                          <div className="grid grid-cols-1 gap-2">
                            <button
                              onClick={() => handleChooseVocationSim("Knight")}
                              className="w-full text-left p-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:bg-amber-500/10 hover:border-amber-500/40 transition flex justify-between items-center"
                            >
                              <div>
                                <span className="font-bold text-xs text-slate-200 block">Knight (Guerreiro)</span>
                                <span className="text-[10px] text-slate-400 block mt-0.5">Foco em defesa física extrema e maior vida base.</span>
                              </div>
                              <span className="text-[9px] bg-amber-500/10 text-amber-400 font-bold px-2 py-0.5 rounded border border-amber-500/20 uppercase font-mono">+Def & HP</span>
                            </button>
                            <button
                              onClick={() => handleChooseVocationSim("Mage")}
                              className="w-full text-left p-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:bg-amber-500/10 hover:border-amber-500/40 transition flex justify-between items-center"
                            >
                              <div>
                                <span className="font-bold text-xs text-slate-200 block">Mage (Mago)</span>
                                <span className="text-[10px] text-slate-400 block mt-0.5">Destruição mágica com dano elemental e alta mana.</span>
                              </div>
                              <span className="text-[9px] bg-blue-500/10 text-blue-400 font-bold px-2 py-0.5 rounded border border-blue-500/20 uppercase font-mono">+Elem & Mana</span>
                            </button>
                            <button
                              onClick={() => handleChooseVocationSim("Archer")}
                              className="w-full text-left p-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:bg-amber-500/10 hover:border-amber-500/40 transition flex justify-between items-center"
                            >
                              <div>
                                <span className="font-bold text-xs text-slate-200 block">Archer (Arqueiro)</span>
                                <span className="text-[10px] text-slate-400 block mt-0.5">Dano rápido à distância com alta precisão e crítico.</span>
                              </div>
                              <span className="text-[9px] bg-green-500/10 text-green-400 font-bold px-2 py-0.5 rounded border border-green-500/20 uppercase font-mono">+Crit & Acc</span>
                            </button>
                            <button
                              onClick={() => handleChooseVocationSim("Assassin")}
                              className="w-full text-left p-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:bg-amber-500/10 hover:border-amber-500/40 transition flex justify-between items-center"
                            >
                              <div>
                                <span className="font-bold text-xs text-slate-200 block">Assassin (Assassino)</span>
                                <span className="text-[10px] text-slate-400 block mt-0.5">Velocidade, golpes furtivos críticos e alta esquiva.</span>
                              </div>
                              <span className="text-[9px] bg-purple-500/10 text-purple-400 font-bold px-2 py-0.5 rounded border border-purple-500/20 uppercase font-mono">+Crit & Eva</span>
                            </button>
                            <button
                              onClick={() => handleChooseVocationSim("Cleric")}
                              className="w-full text-left p-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:bg-amber-500/10 hover:border-amber-500/40 transition flex justify-between items-center"
                            >
                              <div>
                                <span className="font-bold text-xs text-slate-200 block">Cleric (Clérigo)</span>
                                <span className="text-[10px] text-slate-400 block mt-0.5">Suporte espiritual, cura, resistência mágica e elemental.</span>
                              </div>
                              <span className="text-[9px] bg-emerald-500/10 text-emerald-400 font-bold px-2 py-0.5 rounded border border-emerald-500/20 uppercase font-mono">+Res & DefEl</span>
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Middle Column: Elemental Affinity & Subclasses */}
                  <div className="lg:col-span-5 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-4">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Activity className="w-4 h-4 text-orange-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Afinidade Elemental Oculta</h4>
                        </div>
                        {spamWarning && (
                          <span className="text-[9px] text-rose-400 font-bold animate-pulse bg-rose-500/10 border border-rose-500/20 px-2 py-0.5 rounded uppercase font-mono">
                            Spam Shield Ativo
                          </span>
                        )}
                      </div>

                      <div className="space-y-3">
                        <p className="text-[11px] text-slate-400 leading-relaxed">
                          Habilidades mágicas, combates e quests no mundo acumulam silenciosamente scores de afinidade. A afinidade possui um progresso independente (Nvl 1-100) do nível de personagem (Nvl 1-9999). O despertar da subclasse elemental é permanente e requer <strong>Nível de Personagem ≥ 100</strong> e <strong>Nível de Afinidade Dominante = 100</strong>.
                        </p>

                        <div className="space-y-3">
                          {/* Fire */}
                          <div className="bg-slate-900/40 p-2 rounded-lg border border-slate-900/60">
                            <div className="flex justify-between text-xs font-mono text-slate-300 mb-1">
                              <span className="flex items-center gap-1 font-bold"><Flame className="w-3.5 h-3.5 text-orange-500" /> Fogo (Fire)</span>
                              <span className="text-[11px] text-orange-400 font-semibold">Nvl {affinityFireLvl} <span className="text-[9px] text-slate-500">({affinityFireXp}/1000 XP)</span></span>
                            </div>
                            <div className="w-full bg-slate-950 border border-slate-900 rounded-full h-2 overflow-hidden flex">
                              <div className="bg-gradient-to-r from-orange-600 to-amber-500 h-full transition-all duration-300" style={{ width: `${affinityFireLvl === 100 ? 100 : (affinityFireXp / 10)}%` }}></div>
                            </div>
                          </div>

                          {/* Ice */}
                          <div className="bg-slate-900/40 p-2 rounded-lg border border-slate-900/60">
                            <div className="flex justify-between text-xs font-mono text-slate-300 mb-1">
                              <span className="flex items-center gap-1 font-bold"><Snowflake className="w-3.5 h-3.5 text-cyan-400" /> Gelo (Ice)</span>
                              <span className="text-[11px] text-cyan-400 font-semibold">Nvl {affinityIceLvl} <span className="text-[9px] text-slate-500">({affinityIceXp}/1000 XP)</span></span>
                            </div>
                            <div className="w-full bg-slate-950 border border-slate-900 rounded-full h-2 overflow-hidden flex">
                              <div className="bg-gradient-to-r from-cyan-500 to-blue-500 h-full transition-all duration-300" style={{ width: `${affinityIceLvl === 100 ? 100 : (affinityIceXp / 10)}%` }}></div>
                            </div>
                          </div>

                          {/* Holy */}
                          <div className="bg-slate-900/40 p-2 rounded-lg border border-slate-900/60">
                            <div className="flex justify-between text-xs font-mono text-slate-300 mb-1">
                              <span className="flex items-center gap-1 font-bold"><Sparkles className="w-3.5 h-3.5 text-amber-400" /> Sagrado (Holy)</span>
                              <span className="text-[11px] text-amber-400 font-semibold">Nvl {affinityHolyLvl} <span className="text-[9px] text-slate-500">({affinityHolyXp}/1000 XP)</span></span>
                            </div>
                            <div className="w-full bg-slate-950 border border-slate-900 rounded-full h-2 overflow-hidden flex">
                              <div className="bg-gradient-to-r from-amber-400 to-yellow-300 h-full transition-all duration-300" style={{ width: `${affinityHolyLvl === 100 ? 100 : (affinityHolyXp / 10)}%` }}></div>
                            </div>
                          </div>

                          {/* Shadow */}
                          <div className="bg-slate-900/40 p-2 rounded-lg border border-slate-900/60">
                            <div className="flex justify-between text-xs font-mono text-slate-300 mb-1">
                              <span className="flex items-center gap-1 font-bold"><Moon className="w-3.5 h-3.5 text-purple-400" /> Sombra (Shadow)</span>
                              <span className="text-[11px] text-purple-400 font-semibold">Nvl {affinityShadowLvl} <span className="text-[9px] text-slate-500">({affinityShadowXp}/1000 XP)</span></span>
                            </div>
                            <div className="w-full bg-slate-950 border border-slate-900 rounded-full h-2 overflow-hidden flex">
                              <div className="bg-gradient-to-r from-purple-600 to-indigo-700 h-full transition-all duration-300" style={{ width: `${affinityShadowLvl === 100 ? 100 : (affinityShadowXp / 10)}%` }}></div>
                            </div>
                          </div>

                          {/* Nature */}
                          <div className="bg-slate-900/40 p-2 rounded-lg border border-slate-900/60">
                            <div className="flex justify-between text-xs font-mono text-slate-300 mb-1">
                              <span className="flex items-center gap-1 font-bold"><Leaf className="w-3.5 h-3.5 text-emerald-400" /> Natural (Nature)</span>
                              <span className="text-[11px] text-emerald-400 font-semibold">Nvl {affinityNatureLvl} <span className="text-[9px] text-slate-500">({affinityNatureXp}/1000 XP)</span></span>
                            </div>
                            <div className="w-full bg-slate-950 border border-slate-900 rounded-full h-2 overflow-hidden flex">
                              <div className="bg-gradient-to-r from-emerald-500 to-teal-500 h-full transition-all duration-300" style={{ width: `${affinityNatureLvl === 100 ? 100 : (affinityNatureXp / 10)}%` }}></div>
                            </div>
                          </div>
                        </div>

                        {/* Fast Simulated Gain Buttons */}
                        <div className="border-t border-slate-800/80 pt-3">
                          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block mb-2">Simular Acúmulo de Afinidade (Proteção Anti-Spam de 1.5s):</span>
                          <div className="grid grid-cols-2 gap-2 text-[10.5px]">
                            <button
                              onClick={() => handleAddAffinitySim("holy", 3, "Conjurar Slash")}
                              className="py-1.5 px-2 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded font-mono font-bold text-slate-300 text-left flex justify-between items-center animate-fade-in"
                            >
                              <span>⚔ Slash</span>
                              <span className="text-amber-400 text-[10px] font-semibold">+3 Holy</span>
                            </button>
                            <button
                              onClick={() => handleAddAffinitySim("fire", 3, "Conjurar Fireball")}
                              className="py-1.5 px-2 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded font-mono font-bold text-slate-300 text-left flex justify-between items-center"
                            >
                              <span>🔥 Fireball</span>
                              <span className="text-orange-400 text-[10px] font-semibold">+3 Fire</span>
                            </button>
                            <button
                              onClick={() => handleAddAffinitySim("ice", 3, "Conjurar Spear Thrust")}
                              className="py-1.5 px-2 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded font-mono font-bold text-slate-300 text-left flex justify-between items-center"
                            >
                              <span>❄ Thrust</span>
                              <span className="text-cyan-400 text-[10px] font-semibold">+3 Ice</span>
                            </button>
                            <button
                              onClick={() => handleAddAffinitySim("shadow", 3, "Matar Servo das Sombras")}
                              className="py-1.5 px-2 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded font-mono font-bold text-slate-300 text-left flex justify-between items-center"
                            >
                              <span>💀 Shadow Kill</span>
                              <span className="text-purple-400 text-[10px] font-semibold">+3 Shadow</span>
                            </button>
                            <button
                              onClick={() => handleAddAffinitySim("nature", 5, "Completar Quest Floresta")}
                              className="col-span-2 py-1.5 px-2.5 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded font-mono font-bold text-slate-300 text-left flex justify-between items-center"
                            >
                              <span>📜 Quest: Proteção do Bosque</span>
                              <span className="text-emerald-400 text-[10px] font-semibold">+5 Nature</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Subclass Unlock Card (Level 100) */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-violet-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Subclasse Elemental (Level 100)</h4>
                        </div>
                        {progSubclass !== "" ? (
                          <span className="text-[10px] text-violet-400 font-bold bg-violet-500/10 border border-violet-500/20 px-2 py-0.5 rounded-full uppercase tracking-wider">
                            Ativa
                          </span>
                        ) : progLevel >= 100 ? (
                          <span className="text-[10px] text-amber-400 font-bold bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded-full uppercase tracking-wider animate-pulse">
                            Despertar
                          </span>
                        ) : (
                          <span className="text-[10px] text-slate-500 font-bold bg-slate-500/10 border border-slate-800 px-2 py-0.5 rounded-full uppercase tracking-wider">
                            Requer Nvl 100
                          </span>
                        )}
                      </div>

                      {progLevel < 100 ? (
                        <div className="flex flex-col gap-2 py-1 text-center">
                          <p className="text-[11px] text-slate-400">
                            A evolução de subclasse elemental está trancada. Evolua seu personagem até o nível 100 para despertar seu poder espiritual.
                          </p>
                          <div className="w-full bg-slate-900 rounded-full h-1.5 overflow-hidden">
                            <div className="bg-violet-500 h-full transition-all" style={{ width: `${progLevel}%` }}></div>
                          </div>
                        </div>
                      ) : progSubclass !== "" ? (
                        <div className="bg-gradient-to-r from-violet-950/60 to-slate-950 border border-violet-800/60 p-4 rounded-xl flex flex-col gap-2 shadow-inner">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              <div className="p-2 bg-violet-500/10 border border-violet-500/20 rounded-lg text-violet-400">
                                <Sparkles className="w-5 h-5 animate-spin-slow" />
                              </div>
                              <div>
                                <h5 className="font-bold text-xs text-violet-300">SUBCLASSE DESPERTA</h5>
                                <h4 className="font-extrabold text-base text-slate-100 font-sans tracking-tight">{progSubclass}</h4>
                              </div>
                            </div>
                            <span className="text-[10px] text-emerald-400 font-extrabold bg-emerald-500/10 border border-emerald-500/20 px-2.5 py-0.5 rounded">
                              +15% Dano Ativo
                            </span>
                          </div>
                          <p className="text-[11px] text-slate-400 italic">
                            O bônus passivo de 15% de dano elemental e 10% de mitigação foram aplicados com sucesso através da fórmula de combate autoritativa.
                          </p>
                        </div>
                      ) : (
                        <div className="flex flex-col gap-3">
                          <p className="text-[11px] text-slate-400 leading-relaxed">
                            Seu personagem atingiu o nível 100! O sistema calcula o elemento dominante do seu histórico espiritual para forjar a sua subclasse final:
                          </p>
                          <button
                            onClick={handleUnlockSubclassSim}
                            className="w-full py-2.5 bg-violet-600 hover:bg-violet-500 text-slate-100 rounded font-bold text-xs uppercase transition shadow-lg shadow-violet-950/40 flex items-center justify-center gap-2"
                          >
                            <Unlock className="w-4 h-4" /> Desbloquear Subclasse Elemental (CS_UNLOCK_SUBCLASS)
                          </button>
                        </div>
                      )}
                    </div>

                    {/* Calculadora de Eficiência de Equipamento / Compatibilidade */}
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Shield className="w-4 h-4 text-emerald-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Simulador de Compatibilidade de Equipamento</h4>
                        </div>
                        <span className="text-[10px] text-emerald-400 font-mono font-bold bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-md">
                          Patched V2
                        </span>
                      </div>

                      <p className="text-[11px] text-slate-400 leading-relaxed">
                        Verifique a eficiência dos atributos de seus equipamentos com base na sua afinidade elemental. Itens de elementos opostos (<strong className="text-red-400 font-mono">Fogo ↔ Gelo</strong> ou <strong className="text-purple-400 font-mono">Sagrado ↔ Sombra</strong>) sofrem penalidade de 50%. Itens neutros operam com 75%, e itens idênticos têm 100% de eficiência.
                      </p>

                      <div className="grid grid-cols-2 gap-3 mt-1">
                        <div>
                          <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1">Afinidade Ativa:</label>
                          <select
                            value={selectedPlayerAffinity}
                            onChange={(e) => setSelectedPlayerAffinity(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2 py-1.5 text-xs text-slate-300 outline-none focus:border-amber-500/40 font-mono"
                          >
                            <option value="neutral">Neutro (Nenhuma)</option>
                            <option value="fire">🔥 Fogo (Fire)</option>
                            <option value="ice">❄ Gelo (Ice)</option>
                            <option value="holy">✨ Sagrado (Holy)</option>
                            <option value="shadow">💀 Sombra (Shadow)</option>
                            <option value="nature">🍃 Natural (Nature)</option>
                          </select>
                        </div>

                        <div>
                          <label className="text-[10px] text-slate-500 font-bold uppercase block mb-1">Elemento do Item:</label>
                          <select
                            value={selectedGearElement}
                            onChange={(e) => setSelectedGearElement(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2 py-1.5 text-xs text-slate-300 outline-none focus:border-amber-500/40 font-mono"
                          >
                            <option value="fire">🔥 Item de Fogo</option>
                            <option value="ice">❄ Item de Gelo</option>
                            <option value="holy">✨ Item Sagrado</option>
                            <option value="shadow">💀 Item Sombrio</option>
                            <option value="nature">🍃 Item de Natureza</option>
                          </select>
                        </div>
                      </div>

                      {/* Display calculations */}
                      {(() => {
                        const eff = getGearEfficiency(selectedPlayerAffinity, selectedGearElement);
                        const badgeColor = eff.pct === 100 
                          ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-400" 
                          : eff.pct === 75 
                            ? "bg-amber-500/10 border-amber-500/30 text-amber-400" 
                            : "bg-rose-500/10 border-rose-500/30 text-rose-400";
                        return (
                          <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-3 flex flex-col gap-2 mt-1">
                            <div className="flex justify-between items-center border-b border-slate-800/40 pb-1.5">
                              <span className="text-[11px] text-slate-400 font-semibold uppercase">Fórmula de Eficiência:</span>
                              <span className={`text-xs font-mono font-black border px-2 py-0.5 rounded ${badgeColor}`}>
                                {eff.pct}% Eficiência
                              </span>
                            </div>
                            <p className="text-[11px] text-slate-300 leading-normal">
                              {eff.desc}
                            </p>
                            <div className="w-full bg-slate-950 rounded-full h-1.5 overflow-hidden flex mt-1">
                              <div 
                                className={`h-full transition-all duration-300 ${
                                  eff.pct === 100 ? "bg-emerald-500" : eff.pct === 75 ? "bg-amber-500" : "bg-rose-500"
                                }`} 
                                style={{ width: `${eff.pct}%` }}
                              ></div>
                            </div>
                          </div>
                        );
                      })()}
                    </div>
                  </div>

                  {/* Right Column: Geographic Lock / Cities Gates (Level 50) */}
                  <div className="lg:col-span-3 flex flex-col gap-4">
                    <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3 flex-1">
                      <div className="flex items-center justify-between border-b border-slate-800/60 pb-2">
                        <div className="flex items-center gap-2">
                          <Map className="w-4 h-4 text-amber-400" />
                          <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Navegador do Mundo</h4>
                        </div>
                        <span className="text-[10px] text-slate-400 font-mono">
                          X:{Math.round(progX)}, Y:{Math.round(progY)}
                        </span>
                      </div>

                      {/* Continent Selection Tabs */}
                      <div className="grid grid-cols-7 gap-0.5 bg-slate-900/60 p-1 rounded-lg border border-slate-800">
                        {[
                          { id: "main_continent", label: "Principal", icon: "🗺️" },
                          { id: "nature", label: "Nature", icon: "🍃" },
                          { id: "shadow", label: "Sombras", icon: "🌙" },
                          { id: "fire_continent", label: "Fogo", icon: "🔥" },
                          { id: "holy", label: "Sagrado", icon: "✨" },
                          { id: "ice", label: "Gelo", icon: "❄️" },
                          { id: "abyssia", label: "Abyssia", icon: "🌌" }
                        ].map(c => (
                          <button
                            key={c.id}
                            onClick={() => {
                              setSelectedContinent(c.id as any);
                              // Auto-select first settlement of that continent
                              const defaultSettlements: Record<string, string> = {
                                main_continent: "ironhold_bastion",
                                nature: "elarin",
                                shadow: "noctharyn",
                                fire_continent: "pyra_magnus",
                                holy: "luminaar",
                                ice: "elarisheim",
                                abyssia: "last_bastion"
                              };
                              setActiveSettlement(defaultSettlements[c.id]);
                            }}
                            className={`flex flex-col items-center gap-0.5 py-1 px-0.5 rounded text-[9px] font-medium transition ${
                              selectedContinent === c.id
                                ? "bg-amber-600/20 border border-amber-500/30 text-amber-300 font-bold"
                                : "hover:bg-slate-800 text-slate-400 border border-transparent"
                            }`}
                          >
                            <span className="text-xs">{c.icon}</span>
                            <span className="scale-[0.85] origin-center tracking-tight leading-none">{c.label}</span>
                          </button>
                        ))}
                      </div>

                      {/* Map Container & Interactive SVG */}
                      {(() => {
                        const continentsInfo = {
                          main_continent: {
                            name: "Continente Principal (Main)",
                            minLevel: 1,
                            color: "amber",
                            desc: "Starter zone (Lvl 1-10), Midgame progression (Lvl 10-50), and global economic center.",
                            bgGrid: "radial-gradient(circle, rgba(245,158,11,0.05) 1px, transparent 1px)",
                            settlements: [
                              { id: "ironhold_bastion", name: "Ironhold Bastion", x: 100, y: 100, svgX: 15, svgY: 15, type: "Starter Military Bastion", leader: "Prefeito George & Guarda Will", desc: "A pacata fortaleza militar de iniciantes protegida de ameaças. Sede do comércio inicial." },
                              { id: "stone_tirith", name: "Stone Tirith", x: 500, y: 500, svgX: 35, svgY: 35, type: "Dwarf Mountain City", leader: "Thane Thromgar & Marechal Varian", desc: "Grande fortaleza dos anões nas montanhas, centro de forja, mineração e criação de armas." },
                              { id: "blackwater_bay", name: "Blackwater Bay", x: 1200, y: 1200, svgX: 70, svgY: 70, type: "Coastal Trade & Pirate Hub", leader: "Capitão Redbeard & Port Master Drake", desc: "O maior porto marítimo global, polo de contrabando, pirataria e rotas livres." },
                              { id: "ravenshire", name: "Ravenshire", x: 1600, y: 600, svgX: 85, svgY: 35, type: "Capital Real & Sede Comercial", leader: "Rei Aldren Vaelor & Lady Genevieve", desc: "Capital política e comercial oficial do continente. Hospeda a corte real, focando em diplomacia, alta aristocracia, casas nobres e guildas influentes (estritamente não militar)." },
                              { id: "thornwall", name: "Thornwall", x: 800, y: 800, svgX: 50, svgY: 50, type: "Forest Hybrid City", leader: "Chieftain Garak & Elder Elenwe", desc: "Cidade híbrida diplomática entre humanos, elfos e orcs verdes nas densas matas do continente." }
                            ]
                          },
                          nature: {
                            name: "Continente da Natureza",
                            minLevel: 50,
                            color: "emerald",
                            desc: "Florestas densas e segredos druídicos de Elarin [Nvl 50+].",
                            bgGrid: "radial-gradient(circle, rgba(16,185,129,0.05) 1px, transparent 1px)",
                            settlements: [
                              { id: "elarin", name: "Elarin", x: 2900, y: 2950, svgX: 100, svgY: 50, type: "Capital Real Elfa", leader: "Rei Thalindir", desc: "A majestosa e antiga capital elfa esculpida nas árvores do Norte." },
                              { id: "sylvaris", name: "Sylvaris", x: 2950, y: 2900, svgX: 160, svgY: 100, type: "Fortaleza Militar", leader: "Comandante Sylas", desc: "Posto militar avançado vigiando as fronteiras do Leste." },
                              { id: "oakenspire", name: "Oakenspire", x: 2900, y: 2850, svgX: 100, svgY: 150, type: "Santuário Druida", leader: "Ancião Faelar", desc: "Ponto de encontro dos druidas e estudantes de alquimia no Sul." },
                              { id: "grunhold", name: "Grunhold", x: 2850, y: 2900, svgX: 40, svgY: 100, type: "Forte de Orcs Verdes", leader: "Gorgok o Verde", desc: "Refúgio pacífico de orcs refugiados nas matas do Oeste." }
                            ]
                          },
                          shadow: {
                            name: "Continente das Sombras",
                            minLevel: 50,
                            color: "purple",
                            desc: "Escuridão feudal primordial, névoa densa e rivalidades eternas [Nvl 50+].",
                            bgGrid: "radial-gradient(circle, rgba(168,85,247,0.06) 1px, transparent 1px)",
                            settlements: [
                              { id: "noctharyn", name: "Noctharyn", x: 3900, y: 3950, svgX: 100, svgY: 45, type: "Império de Sangue (Norte)", leader: "Conde Valdrak Nocthar", desc: "Suntuosa capital gótica governada por vampiros aristocratas e intelectuais." },
                              { id: "grimharbor", name: "Grimharbor", x: 3850, y: 3925, svgX: 45, svgY: 70, type: "Porto de Comércio Livre", leader: "Capitão Dregan", desc: "Ponto de entrada costeiro onde contrabandos, mercenários e o mercado negro prosperam." },
                              { id: "kar_goth", name: "Kar'goth", x: 3825, y: 3900, svgX: 25, svgY: 100, type: "Bastião Orc Negro", leader: "Warchief Gor'mak", desc: "Fortaleza encravada nas montanhas dedicada à metalurgia de armas de guerra." },
                              { id: "necrathis", name: "Necrathis", x: 3900, y: 3850, svgX: 100, svgY: 155, type: "Império Necromante (Sul)", leader: "Morthak o Eterno", desc: "Necrópole monumental de ossos habitada por legiões de mortos-vivos dominadas pelo Lich." },
                              { id: "vel_sharum", name: "Vel'Sharum", x: 3950, y: 3875, svgX: 165, svgY: 120, type: "Santuário dos Cultos", leader: "High Cultist Xerath", desc: "Catedral de rituais dedicada a conjurações sombrias e ao estudo de tomos proibidos." }
                            ]
                          },
                          fire_continent: {
                            name: "Continente de Fogo",
                            minLevel: 50,
                            color: "red",
                            desc: "Vulcão primordial ativo, indústria de forja ciclópea, Red Orcs e perigos magmáticos [Nvl 50+].",
                            bgGrid: "radial-gradient(circle, rgba(239,68,68,0.06) 1px, transparent 1px)",
                            settlements: [
                              { id: "pyra_magnus", name: "Pyra Magnus", x: 2100, y: 2100, svgX: 100, svgY: 100, type: "Capital do Fogo (Centro)", leader: "Ignis Rex (Rei Mago)", desc: "A monumental capital de pedra vulcânica, centro político governado pelo Rei Mago de Fogo." },
                              { id: "crimson_hollow", name: "Crimson Hollow", x: 2050, y: 2150, svgX: 50, svgY: 50, type: "Bastião Orc Vermelho (Noroeste)", leader: "Warlord Gar'thok", desc: "Fortaleza orc militarizada focada na cultura de guerra e arenas de combate." },
                              { id: "molten_anvil", name: "Molten Anvil", x: 2150, y: 2050, svgX: 150, svgY: 150, type: "Forja de Ciclopes (Sudeste)", leader: "Forge Master Brokk", desc: "Cidade-forja dos Ciclopes de Fogo. Sede dos melhores sistemas de crafting do jogo." },
                              { id: "primordial_volcano", name: "Primordial Volcano", x: 2100, y: 2120, svgX: 100, svgY: 60, type: "Dungeon de Alto Perigo (Norte)", leader: "General Malgazar", desc: "O colossal vulcão central habitado por demônios de fogo. Prisão natural do Primordial de Fogo.", hidden: false }
                            ]
                          },
                          holy: {
                            name: "Continente Sagrado",
                            minLevel: 50,
                            color: "amber",
                            desc: "Nobreza rígida, ordem medieval pura e a emanação da luz divina primordial [Nvl 50+].",
                            bgGrid: "radial-gradient(circle, rgba(245,158,11,0.06) 1px, transparent 1px)",
                            settlements: [
                              { id: "luminaar", name: "Luminaar", x: 4900, y: 4950, svgX: 100, svgY: 45, type: "Capital Real (Norte)", leader: "Conselho Triuno da Luz", desc: "Sede do Conselho Triuno governado por Aurelius (defesa), Aelthir (sabedoria) e Seraphiel (espiritual) com autoridade idêntica. Aurelius é primus inter pares para urgências de defesa." },
                              { id: "lunareth", name: "Lunareth", x: 4950, y: 4900, svgX: 165, svgY: 100, type: "Cidadela Élfica (Leste)", leader: "Keeper Aelthir Luminar", desc: "Antiga capital elfa oculta em florestas sagradas, centro do saber sagrado primordial." },
                              { id: "sunwall", name: "Sunwall", x: 4850, y: 4900, svgX: 35, svgY: 100, type: "Muralha de Paladinos (Oeste)", leader: "Sunwall Commander", desc: "Muralha colossal protegendo o continente, comandada por paladinos inflexíveis." },
                              { id: "heart_of_light", name: "Heart of Light", x: 4900, y: 4900, svgX: 100, svgY: 120, type: "Coração Divino (Centro)", leader: "Cristal Primordial", desc: "Zona central restrita onde repousa o indestrutível Cristal Primordial, fonte sagrada da terra." }
                            ]
                          },
                          ice: {
                            name: "Continente de Gelo",
                            minLevel: 1,
                            color: "sky",
                            desc: "Frio primordial eterno e sobrevivência nórdica severa [Nvl 1+].",
                            bgGrid: "radial-gradient(circle, rgba(56,189,248,0.06) 1px, transparent 1px)",
                            settlements: [
                              { id: "elarisheim", name: "Elarisheim", x: 5900, y: 5950, svgX: 100, svgY: 45, type: "Capital Real Elfa (Centro)", leader: "Rei Kaelthar Elaris", desc: "Sede da realeza dos Elfos de Gelo. Abriga o suntuoso palácio de gelo e jardins cristalinos." },
                              { id: "frosthaven", name: "Frosthaven", x: 5900, y: 5850, svgX: 100, svgY: 145, type: "Porto Humano (Sul)", leader: "Port Master", desc: "Movimentado porto comercial e principal colônia humana nas costas congeladas." },
                              { id: "khaz_tirith", name: "Khaz Tirith", x: 5950, y: 5900, svgX: 165, svgY: 100, type: "Fortaleza Anã (Leste)", leader: "Forge Master", desc: "Cidadela de pedra esculpida sob as montanhas geladas. Portal para a Caverna de Ymirr." },
                              { id: "ymirr_hidden_cavern", name: "Caverna de Ymirr", x: 5980, y: 5965, svgX: 195, svgY: 30, type: "Santuário Oculto (Montanhas)", leader: "Ymirr Stonefrost", desc: "A lendária e perigosa caverna subterrânea profunda onde habita o antigo Jötunn.", hidden: true }
                            ]
                          },
                          abyssia: {
                            name: "Abyssia (Abyssi)",
                            minLevel: 150,
                            color: "violet",
                            desc: "O abismo corrompido pós-apocalíptico do fim do jogo [Nvl 150+]. Rota marítima restrita.",
                            bgGrid: "radial-gradient(circle, rgba(139,92,246,0.06) 1px, transparent 1px)",
                            settlements: [
                              { id: "last_bastion", name: "Last Bastion", x: 3400, y: 3100, svgX: 100, svgY: 145, type: "Bastião Humano (Sul)", leader: "Grand Marshal Varian", desc: "O único e último bastião de resistência humana. Zona segura, respawn, consertos, mercadores e base de missões." },
                              { id: "void_rift_ruins", name: "Ruínas do Vazio", x: 3200, y: 3300, svgX: 50, svgY: 70, type: "Ruínas Antigas", leader: "Nenhum", desc: "Ruínas flutuantes desmoronadas repletas de horrores do vazio." },
                              { id: "crystalline_corruption", name: "Zona Cristalina Corrompida", x: 3600, y: 3500, svgX: 150, svgY: 50, type: "Zona Corrompida", leader: "Nenhum", desc: "Região abissal tomada por cristais de energia do vazio que drenam a vitalidade." }
                            ]
                          }
                        };

                        const activeCont = continentsInfo[selectedContinent];
                        const activeSet = activeCont.settlements.find(s => s.id === activeSettlement) || activeCont.settlements[0];
                        const isLocked = progLevel < activeCont.minLevel;

                        return (
                          <div className="flex flex-col gap-3 flex-1">
                            {/* Header Info */}
                            <div className="bg-slate-900/40 border border-slate-800/40 p-2.5 rounded-lg">
                              <div className="flex justify-between items-center mb-1">
                                <span className="font-bold text-[11px] text-slate-200">{activeCont.name}</span>
                                <span className={`text-[9px] px-1.5 py-0.5 rounded font-mono font-bold ${
                                  isLocked ? "bg-rose-500/10 text-rose-400 border border-rose-500/20" : "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                }`}>
                                  Requer Nvl {activeCont.minLevel}+
                                </span>
                              </div>
                              <p className="text-[10px] text-slate-400 leading-normal">{activeCont.desc}</p>
                            </div>

                            {/* SVG Visual Map */}
                            <div 
                              className="relative h-[160px] w-full bg-slate-950 border border-slate-800 rounded-lg overflow-hidden flex items-center justify-center cursor-crosshair shadow-inner"
                              style={{ 
                                backgroundImage: activeCont.bgGrid,
                                backgroundSize: "16px 16px" 
                              }}
                            >
                              {/* Continent Border Line / Map Outline */}
                              <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-40">
                                {selectedContinent === "main_continent" && (
                                  <>
                                    {/* Main Continent outline */}
                                    <path d="M 20 40 Q 50 15 100 25 T 180 50 T 170 140 T 90 170 T 30 110 Z" fill="none" stroke="#f59e0b" strokeWidth="1.2" strokeDasharray="4 2" />
                                    {/* Major rivers or roads */}
                                    <path d="M 100 25 Q 80 80 120 120 T 70 170" fill="none" stroke="#3b82f6" strokeWidth="0.8" strokeDasharray="2 2" opacity="0.5" />
                                    <text x="110" y="35" fill="#f59e0b" className="text-[7px] font-mono opacity-80 font-bold uppercase">Main Lands</text>
                                  </>
                                )}
                                {selectedContinent === "shadow" && (
                                  <>
                                    {/* Coastlines */}
                                    <path d="M 20 60 Q 40 40 80 30 T 150 40 T 180 100 T 170 150 T 120 180 T 50 160 T 20 110 Z" fill="none" stroke="#6b21a8" strokeWidth="1" strokeDasharray="4 2" />
                                    {/* Division / Border line North and South */}
                                    <line x1="20" y1="100" x2="180" y2="100" stroke="#4a044e" strokeWidth="0.8" strokeDasharray="3 3" />
                                    <text x="145" y="95" fill="#a855f7" className="text-[7px] font-mono opacity-80 font-bold">ZONA NORTE</text>
                                    <text x="145" y="110" fill="#ec4899" className="text-[7px] font-mono opacity-80 font-bold">ZONA SUL</text>
                                    {/* Ashen Wastes center area indicator */}
                                    <circle cx="100" cy="100" r="22" fill="rgba(147,151,168,0.06)" stroke="rgba(147,151,168,0.2)" strokeWidth="1" strokeDasharray="2 1" />
                                    <text x="100" y="103" textAnchor="middle" fill="#94a3b8" className="text-[7px] font-mono tracking-wider opacity-90 font-bold">ASHEN WASTES</text>
                                  </>
                                )}
                                {selectedContinent === "nature" && (
                                  <path d="M 30 50 Q 100 20 170 50 T 170 150 T 30 150 Z" fill="none" stroke="#047857" strokeWidth="1" strokeDasharray="4 2" />
                                )}
                                {selectedContinent === "fire_continent" && (
                                  <>
                                    {/* Jagged Volcanic Coastline */}
                                    <path d="M 30 40 L 90 20 L 165 35 L 180 120 L 135 165 L 50 145 Z" fill="none" stroke="#ef4444" strokeWidth="1.2" strokeDasharray="3 3" />
                                    {/* Magma Streams */}
                                    <path d="M 100 80 Q 70 110 50 50 M 100 80 Q 130 110 150 150" fill="none" stroke="#f97316" strokeWidth="1" strokeDasharray="1 1" />
                                    {/* Primordial Volcano Crater Circle */}
                                    <circle cx="100" cy="80" r="16" fill="rgba(239,68,68,0.04)" stroke="rgba(239,68,68,0.3)" strokeWidth="1" strokeDasharray="2 2" />
                                    <text x="100" y="83" textAnchor="middle" fill="#ef4444" className="text-[6.5px] font-mono tracking-widest opacity-95 font-bold uppercase">Vulcão Primordial</text>
                                  </>
                                )}
                                {selectedContinent === "holy" && (
                                  <>
                                    {/* Sacred coast outline */}
                                    <path d="M 40 30 Q 100 10 160 30 T 170 120 T 100 150 T 30 100 Z" fill="none" stroke="#d97706" strokeWidth="1" strokeDasharray="3 3" />
                                    {/* Sacred rivers */}
                                    <path d="M 100 120 Q 90 90 100 45" fill="none" stroke="#60a5fa" strokeWidth="0.8" strokeDasharray="1 1" />
                                    {/* Central core protected circle indicator */}
                                    <circle cx="100" cy="120" r="14" fill="rgba(245,158,11,0.04)" stroke="rgba(245,158,11,0.2)" strokeWidth="0.8" strokeDasharray="2 2" />
                                  </>
                                )}
                                {selectedContinent === "ice" && (
                                  <>
                                    {/* Frozen coast outline */}
                                    <path d="M 25 35 Q 100 15 175 35 T 160 135 T 100 160 T 40 125 Z" fill="none" stroke="#38bdf8" strokeWidth="1.2" strokeDasharray="4 3" />
                                    {/* Glacial crevasse and snow winds */}
                                    <path d="M 45 130 Q 100 70 155 50" fill="none" stroke="#93c5fd" strokeWidth="0.8" strokeDasharray="2 2" />
                                    {/* Ancient peaks warning circle */}
                                    <circle cx="150" cy="80" r="22" fill="rgba(56,189,248,0.02)" stroke="rgba(56,189,248,0.2)" strokeWidth="0.8" strokeDasharray="1 2" />
                                    <text x="150" y="83" textAnchor="middle" fill="#38bdf8" className="text-[6px] font-mono tracking-wider opacity-80 font-bold uppercase">Cumes Proibidos</text>
                                  </>
                                )}
                                {selectedContinent === "abyssia" && (
                                  <>
                                    {/* Eerie void fractured coastline */}
                                    <path d="M 15 50 L 40 30 L 95 20 L 160 45 L 185 110 L 140 165 L 75 170 L 30 135 Z" fill="none" stroke="#8b5cf6" strokeWidth="1.2" strokeDasharray="3 4" />
                                    {/* Fractures and Rifts */}
                                    <path d="M 50 100 Q 100 80 150 110 M 70 60 Q 100 120 130 140" fill="none" stroke="#c084fc" strokeWidth="0.8" strokeDasharray="2 2" />
                                    {/* Southern Last Bastion safe zone indicator */}
                                    <circle cx="100" cy="145" r="16" fill="rgba(34,197,94,0.04)" stroke="rgba(34,197,94,0.3)" strokeWidth="1" strokeDasharray="2 2" />
                                    <text x="100" y="148" textAnchor="middle" fill="#22c55e" className="text-[6px] font-mono tracking-widest opacity-95 font-bold uppercase">Área Segura (Last Bastion)</text>
                                  </>
                                )}
                              </svg>

                              {/* Plotted Points as Interactive Elements */}
                              {activeCont.settlements.filter(s => !s.hidden).map(s => {
                                const isCurrentActive = activeSettlement === s.id;
                                return (
                                  <button
                                    key={s.id}
                                    onClick={() => setActiveSettlement(s.id)}
                                    className="absolute transform -translate-x-1/2 -translate-y-1/2 group transition-all duration-300"
                                    style={{ left: `${s.svgX}%`, top: `${s.svgY}%` }}
                                  >
                                    {/* Pulsing indicator if currently selected */}
                                    {isCurrentActive && (
                                      <span className="absolute -inset-2.5 bg-amber-500/20 rounded-full animate-ping pointer-events-none"></span>
                                    )}
                                    <div className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] shadow-lg transition border ${
                                      isCurrentActive 
                                        ? "bg-amber-500 border-amber-300 text-slate-950 font-bold scale-125" 
                                        : "bg-slate-900 hover:bg-slate-800 border-slate-700 text-slate-400 hover:text-slate-200"
                                    }`}>
                                      {s.id === "noctharyn" ? "🏰" : s.id === "grimharbor" ? "⚓" : s.id === "kar_goth" ? "🛡️" : s.id === "necrathis" ? "💀" : s.id === "vel_sharum" ? "🔮" : s.id === "elarin" ? "🧝" : s.id === "sylvaris" ? "🏹" : s.id === "oakenspire" ? "🍃" : s.id === "grunhold" ? "🐗" : s.id === "luminaar" ? "🏰" : s.id === "lunareth" ? "🧝" : s.id === "sunwall" ? "🛡️" : s.id === "heart_of_light" ? "✨" : s.id === "elarisheim" ? "❄️" : s.id === "frosthaven" ? "⚓" : s.id === "khaz_tirith" ? "🧱" : s.id === "ymirr_hidden_cavern" ? "🏔️" : s.id === "pyra_magnus" ? "🌋" : s.id === "crimson_hollow" ? "👹" : s.id === "molten_anvil" ? "⚒️" : s.id === "primordial_volcano" ? "🔥" : s.id === "last_bastion" ? "🛡️" : s.id === "void_rift_ruins" ? "🏛️" : s.id === "crystalline_corruption" ? "🔮" : "📍"}
                                    </div>
                                    {/* Label visible on map */}
                                    <span className={`absolute top-6 left-1/2 transform -translate-x-1/2 bg-slate-950/90 text-slate-300 border border-slate-800/80 rounded px-1 text-[8px] font-sans whitespace-nowrap opacity-60 group-hover:opacity-100 transition shadow ${
                                      isCurrentActive ? "border-amber-500/30 text-amber-300 font-bold opacity-100" : ""
                                    }`}>
                                      {s.name}
                                    </span>
                                  </button>
                                );
                              })}

                              {/* Active position crosshair overlay */}
                              {selectedContinent === "fire_continent" && progX >= 2000 && progX <= 2200 && progY >= 2000 && progY <= 2200 && (
                                <div 
                                  className="absolute transform -translate-x-1/2 -translate-y-1/2 pointer-events-none"
                                  style={{ 
                                    left: `${((progX - 2000) / 200) * 100}%`, 
                                    top: `${(1 - (progY - 2000) / 200) * 100}%` 
                                  }}
                                >
                                  <div className="w-4 h-4 border border-red-500 rounded-full flex items-center justify-center animate-pulse">
                                    <div className="w-1.5 h-1.5 bg-red-500 rounded-full"></div>
                                  </div>
                                  <span className="absolute -top-4 left-4 bg-red-950/90 text-red-300 border border-red-500/20 text-[7px] font-mono font-bold rounded px-1 shadow uppercase whitespace-nowrap">
                                    VOCÊ
                                  </span>
                                </div>
                              )}

                              {selectedContinent === "shadow" && progX >= 3800 && progX <= 4000 && progY >= 3800 && progY <= 4000 && (
                                <div 
                                  className="absolute transform -translate-x-1/2 -translate-y-1/2 pointer-events-none"
                                  style={{ 
                                    left: `${((progX - 3800) / 200) * 100}%`, 
                                    top: `${(1 - (progY - 3800) / 200) * 100}%` 
                                  }}
                                >
                                  <div className="w-4 h-4 border border-emerald-400 rounded-full flex items-center justify-center animate-pulse">
                                    <div className="w-1.5 h-1.5 bg-emerald-400 rounded-full"></div>
                                  </div>
                                  <span className="absolute -top-4 left-4 bg-emerald-950/90 text-emerald-300 border border-emerald-500/20 text-[7px] font-mono font-bold rounded px-1 shadow uppercase whitespace-nowrap">
                                    VOCÊ
                                  </span>
                                </div>
                              )}

                              {selectedContinent === "holy" && progX >= 4800 && progX <= 5000 && progY >= 4800 && progY <= 5000 && (
                                <div 
                                  className="absolute transform -translate-x-1/2 -translate-y-1/2 pointer-events-none"
                                  style={{ 
                                    left: `${((progX - 4800) / 200) * 100}%`, 
                                    top: `${(1 - (progY - 4800) / 200) * 100}%` 
                                  }}
                                >
                                  <div className="w-4 h-4 border border-amber-400 rounded-full flex items-center justify-center animate-pulse">
                                    <div className="w-1.5 h-1.5 bg-amber-400 rounded-full"></div>
                                  </div>
                                  <span className="absolute -top-4 left-4 bg-amber-950/90 text-amber-300 border border-amber-500/20 text-[7px] font-mono font-bold rounded px-1 shadow uppercase whitespace-nowrap">
                                    VOCÊ
                                  </span>
                                </div>
                              )}

                              {selectedContinent === "ice" && progX >= 5800 && progX <= 6000 && progY >= 5800 && progY <= 6000 && (
                                <div 
                                  className="absolute transform -translate-x-1/2 -translate-y-1/2 pointer-events-none"
                                  style={{ 
                                    left: `${((progX - 5800) / 200) * 100}%`, 
                                    top: `${(1 - (progY - 5800) / 200) * 100}%` 
                                  }}
                                >
                                  <div className="w-4 h-4 border border-sky-400 rounded-full flex items-center justify-center animate-pulse">
                                    <div className="w-1.5 h-1.5 bg-sky-400 rounded-full"></div>
                                  </div>
                                  <span className="absolute -top-4 left-4 bg-sky-950/90 text-sky-300 border border-sky-500/20 text-[7px] font-mono font-bold rounded px-1 shadow uppercase whitespace-nowrap">
                                    VOCÊ
                                  </span>
                                </div>
                              )}
                            </div>

                            {/* Holy Quest Trials Card for Holy Continent (Level 50+) */}
                            {selectedContinent === "holy" && !holyQuestCompleted && (
                              <div className="bg-slate-900/90 border border-amber-500/20 rounded-lg p-3 flex flex-col gap-2 mt-1">
                                <div className="flex items-center justify-between border-b border-slate-800 pb-1.5">
                                  <div className="flex items-center gap-1.5 text-amber-400 font-bold text-[10px]">
                                    <Sparkles className="w-3.5 h-3.5 animate-pulse" />
                                    <span>JULGAMENTO ESPIRITUAL</span>
                                  </div>
                                  <span className="text-[8px] bg-amber-500/10 text-amber-400 px-1.5 py-0.2 rounded font-mono font-bold">
                                    Quest Ativa
                                  </span>
                                </div>

                                <p className="text-[10px] text-slate-300 leading-normal">
                                  Complete os três pilares espirituais para provar o mérito de sua alma perante o Conselho da Luz:
                                </p>

                                {progLevel < 50 ? (
                                  <div className="bg-slate-950 p-2.5 rounded border border-slate-800 text-center text-rose-400 text-[10px] font-medium flex flex-col items-center gap-1">
                                    <Lock className="w-4 h-4 text-rose-500" />
                                    <span>Requer Nível 50 para iniciar as Provações Espirituais.</span>
                                  </div>
                                ) : (
                                  <div className="flex flex-col gap-2">
                                    {/* Courage Pillar */}
                                    <div className="flex flex-col gap-1 bg-slate-950 p-2 rounded border border-slate-800/60">
                                      <div className="flex items-center justify-between">
                                        <span className="text-[10px] font-bold text-slate-300">1. Pilar da Coragem:</span>
                                        {pillarCourage === "completed" ? (
                                          <span className="text-[9px] text-emerald-400 font-bold">✓ Concluído</span>
                                        ) : (
                                          <span className="text-[9px] text-amber-500 font-medium">Pendente</span>
                                        )}
                                      </div>
                                      <p className="text-[9px] text-slate-400 leading-normal">Enfrente e elimine o Espectro Herético no Portal para provar sua virtude combativa.</p>
                                      {pillarCourage === "idle" && (
                                        <button
                                          onClick={() => {
                                            setPillarCourage("running");
                                            addSimLogs([{ type: "info", text: "[Quest] Iniciando combate com o Espectro Herético no portal..." }]);
                                            setTimeout(() => {
                                              setPillarCourage("completed");
                                              addSimLogs([{ type: "info", text: "[Quest] Espectro derrotado! Pilar da Coragem provado e sacramentado." }]);
                                            }, 2000);
                                          }}
                                          className="w-full mt-1 py-1 bg-slate-900 hover:bg-slate-800 border border-slate-700 hover:border-amber-500/30 text-slate-200 text-[9px] font-bold rounded transition"
                                        >
                                          Combater Espectro (Coragem)
                                        </button>
                                      )}
                                      {pillarCourage === "running" && (
                                        <div className="text-[9px] text-amber-400 text-center py-1 font-bold animate-pulse">
                                          Combatendo espectro... ⚔️
                                        </div>
                                      )}
                                    </div>

                                    {/* Wisdom Pillar */}
                                    <div className="flex flex-col gap-1 bg-slate-950 p-2 rounded border border-slate-800/60">
                                      <div className="flex items-center justify-between">
                                        <span className="text-[10px] font-bold text-slate-300">2. Pilar da Sabedoria:</span>
                                        {pillarWisdom === "completed" ? (
                                          <span className="text-[9px] text-emerald-400 font-bold">✓ Concluído</span>
                                        ) : (
                                          <span className="text-[9px] text-amber-500 font-medium">Pendente</span>
                                        )}
                                      </div>
                                      <p className="text-[9px] text-slate-400 leading-normal">Quem é o representante dos Elfos no Conselho Triuno da Luz?</p>
                                      {pillarWisdom !== "completed" && (
                                        <div className="flex flex-col gap-1 mt-1">
                                          {[
                                            { ans: "Aurelius Dawnbringer", label: "Aurelius Dawnbringer" },
                                            { ans: "Aelthir Luminar", label: "Aelthir Luminar" },
                                            { ans: "Seraphiel", label: "Seraphiel" }
                                          ].map(opt => (
                                            <button
                                              key={opt.ans}
                                              onClick={() => {
                                                if (opt.ans === "Aelthir Luminar") {
                                                  setPillarWisdom("completed");
                                                  setRiddleFeedback("Correto! Aelthir Luminar é o guardião élfico da sabedoria sagrada.");
                                                  addSimLogs([{ type: "info", text: "[Quest] Resposta correta! Pilar da Sabedoria provado com glória." }]);
                                                } else {
                                                  setRiddleFeedback("Incorreto! Tente novamente.");
                                                  addSimLogs([{ type: "warning", text: "[Quest] Resposta incorreta ao enigma divino." }]);
                                                }
                                              }}
                                              className="w-full text-left py-1 px-2 bg-slate-900 hover:bg-slate-800 border border-slate-800 text-[9px] text-slate-300 hover:text-slate-100 rounded transition"
                                            >
                                              {opt.label}
                                            </button>
                                          ))}
                                          {riddleFeedback && (
                                            <span className={`text-[8px] font-bold mt-1 block ${riddleFeedback.includes("Correto") ? "text-emerald-400" : "text-rose-400"}`}>
                                              {riddleFeedback}
                                            </span>
                                          )}
                                        </div>
                                      )}
                                    </div>

                                    {/* Purity Pillar */}
                                    <div className="flex flex-col gap-1 bg-slate-950 p-2 rounded border border-slate-800/60">
                                      <div className="flex items-center justify-between">
                                        <span className="text-[10px] font-bold text-slate-300">3. Pilar da Pureza:</span>
                                        {pillarPurity === "completed" ? (
                                          <span className="text-[9px] text-emerald-400 font-bold">✓ Concluído</span>
                                        ) : (
                                          <span className="text-[9px] text-amber-500 font-medium">Pendente</span>
                                        )}
                                      </div>
                                      <p className="text-[9px] text-slate-400 leading-normal">Purifique sua essência espiritual no altar da luz nativa.</p>
                                      {pillarPurity === "idle" && (
                                        <button
                                          onClick={() => {
                                            setPillarPurity("running");
                                            addSimLogs([{ type: "info", text: "[Quest] Iniciando rito de purificação da alma no altar sagrado..." }]);
                                            setTimeout(() => {
                                              setPillarPurity("completed");
                                              addSimLogs([{ type: "info", text: "[Quest] Alma purificada com sucesso! Pilar da Pureza provado." }]);
                                            }, 2000);
                                          }}
                                          className="w-full mt-1 py-1 bg-slate-900 hover:bg-slate-800 border border-slate-700 hover:border-amber-500/30 text-slate-200 text-[9px] font-bold rounded transition"
                                        >
                                          Purificar Espírito (Pureza)
                                        </button>
                                      )}
                                      {pillarPurity === "running" && (
                                        <div className="w-full bg-slate-900 rounded h-1.5 mt-1 overflow-hidden">
                                          <div className="bg-amber-400 h-full animate-pulse" style={{ width: "100%" }}></div>
                                        </div>
                                      )}
                                    </div>

                                    {/* Complete Quest Button */}
                                    {pillarCourage === "completed" && pillarWisdom === "completed" && pillarPurity === "completed" && (
                                      <button
                                        onClick={() => {
                                          setHolyQuestCompleted(true);
                                          addSimLogs([
                                            { type: "info", text: "🏆 [Quest] Missão 'holy_continent_access_trial' concluída com absoluto sucesso!" },
                                            { type: "info", text: "🏆 [Reward] Recompensa de Missão: Medalhão Sagrado (holy_medallion) adicionado!" },
                                            { type: "info", text: "🏆 [Travel] Rotas de navios autorizadas. Você agora tem passe livre para o Continente Sagrado!" }
                                          ]);
                                        }}
                                        className="w-full mt-2 py-2 bg-gradient-to-r from-amber-500 to-yellow-500 hover:from-amber-400 hover:to-yellow-400 text-slate-950 font-extrabold text-[10px] tracking-widest uppercase rounded shadow transition duration-300 animate-pulse"
                                      >
                                        Concluir Provação Espiritual
                                      </button>
                                    )}
                                  </div>
                                )}
                              </div>
                            )}

                            {/* Abyssi Endgame Questline Card for Abyssia Continent (Level 150+) */}
                            {selectedContinent === "abyssia" && (!abyssiQuestCompleted || !abyssiPermissionFlag) && (
                              <div className="bg-slate-900/90 border border-violet-500/20 rounded-lg p-3 flex flex-col gap-2 mt-1 shadow-lg shadow-violet-950/20">
                                <div className="flex items-center justify-between border-b border-slate-800 pb-1.5">
                                  <div className="flex items-center gap-1.5 text-violet-400 font-bold text-[10px]">
                                    <Sparkles className="w-3.5 h-3.5 animate-pulse text-violet-400" />
                                    <span>PROVAÇÃO DO ABISMO (ENDGAME)</span>
                                  </div>
                                  <span className="text-[8px] bg-violet-500/10 text-violet-400 px-1.5 py-0.2 rounded font-mono font-bold uppercase">
                                    Nível 150+
                                  </span>
                                </div>

                                <p className="text-[10px] text-slate-300 leading-normal">
                                  A entrada para Abyssia é restrita a lendas. Complete a jornada de fim de jogo e obtenha o visto militar do Grande Marechal:
                                </p>

                                {progLevel < 150 ? (
                                  <div className="bg-slate-950 p-2.5 rounded border border-slate-800 text-center text-rose-400 text-[10px] font-medium flex flex-col items-center gap-1">
                                    <Lock className="w-4 h-4 text-rose-500" />
                                    <span>Requer Nível 150 para desbloquear a jornada abissal. (Seu nível: {progLevel})</span>
                                    <p className="text-[9px] text-slate-500 mt-0.5">Use o painel de ajuste de nível acima para subir seu nível.</p>
                                  </div>
                                ) : (
                                  <div className="flex flex-col gap-2">
                                    {/* Sub-quest 1: Slay the Void Terror */}
                                    <div className="flex flex-col gap-1 bg-slate-950 p-2 rounded border border-slate-800/60">
                                      <div className="flex items-center justify-between">
                                        <span className="text-[10px] font-bold text-slate-300">1. Derrotar o Terror do Vazio:</span>
                                        {abyssStepVoidTerror === "completed" ? (
                                          <span className="text-[9px] text-emerald-400 font-bold">✓ Derrotado</span>
                                        ) : (
                                          <span className="text-[9px] text-violet-400 font-medium">Pendente</span>
                                        )}
                                      </div>
                                      <p className="text-[9px] text-slate-400 leading-normal">Confronte o Terror do Vazio nas profundezas cósmicas para abrir caminho para os navios.</p>
                                      {abyssStepVoidTerror === "idle" && (
                                        <button
                                          onClick={() => {
                                            setAbyssStepVoidTerror("running");
                                            addSimLogs([{ type: "info", text: "[Quest] Desafiando o Terror do Vazio em um combate de fim de jogo heróico..." }]);
                                            setTimeout(() => {
                                              setAbyssStepVoidTerror("completed");
                                              addSimLogs([{ type: "info", text: "[Quest] Terror do Vazio aniquilado! O caminho espacial/marítimo foi limpo." }]);
                                            }, 2000);
                                          }}
                                          className="w-full mt-1 py-1 bg-slate-900 hover:bg-slate-800 border border-slate-700 hover:border-violet-500/30 text-slate-200 text-[9px] font-bold rounded transition"
                                        >
                                          Batalhar com o Terror do Vazio
                                        </button>
                                      )}
                                      {abyssStepVoidTerror === "running" && (
                                        <div className="text-[9px] text-violet-400 text-center py-1 font-bold animate-pulse">
                                          Enfrentando horror abissal... 👾
                                        </div>
                                      )}
                                    </div>

                                    {/* Sub-quest 2: Seal Abyssal Fracture */}
                                    <div className="flex flex-col gap-1 bg-slate-950 p-2 rounded border border-slate-800/60">
                                      <div className="flex items-center justify-between">
                                        <span className="text-[10px] font-bold text-slate-300">2. Selar Fratura Abissal:</span>
                                        {abyssStepVoidFracture === "completed" ? (
                                          <span className="text-[9px] text-emerald-400 font-bold">✓ Selado</span>
                                        ) : (
                                          <span className="text-[9px] text-violet-400 font-medium">Pendente</span>
                                        )}
                                      </div>
                                      <p className="text-[9px] text-slate-400 leading-normal">Realize o rito de fechamento da brecha espacial para estabilizar as correntes de marés de Abyssia.</p>
                                      {abyssStepVoidFracture === "idle" && (
                                        <button
                                          onClick={() => {
                                            setAbyssStepVoidFracture("running");
                                            addSimLogs([{ type: "info", text: "[Quest] Iniciando o ritual divino de selamento cósmico da Fratura Abissal..." }]);
                                            setTimeout(() => {
                                              setAbyssStepVoidFracture("completed");
                                              addSimLogs([{ type: "info", text: "[Quest] Fratura Cósmica selada com glória! As marés do abismo estão navegáveis." }]);
                                            }, 2000);
                                          }}
                                          className="w-full mt-1 py-1 bg-slate-900 hover:bg-slate-800 border border-slate-700 hover:border-violet-500/30 text-slate-200 text-[9px] font-bold rounded transition"
                                        >
                                          Executar Ritual de Selamento
                                        </button>
                                      )}
                                      {abyssStepVoidFracture === "running" && (
                                        <div className="w-full bg-slate-900 rounded h-1.5 mt-1 overflow-hidden">
                                          <div className="bg-violet-500 h-full animate-pulse" style={{ width: "100%" }}></div>
                                        </div>
                                      )}
                                    </div>

                                    {/* Complete Endgame Quest Button */}
                                    {!abyssiQuestCompleted && abyssStepVoidTerror === "completed" && abyssStepVoidFracture === "completed" && (
                                      <button
                                        onClick={() => {
                                          setAbyssiQuestCompleted(true);
                                          addSimLogs([
                                            { type: "info", text: "🏆 [Quest] Jornada 'abyssal_endgame_questline' concluída com prestígio de lenda!" },
                                            { type: "info", text: "🏆 [Reward] Desbloqueio: Rota marítima comercial e navio abissal liberados." }
                                          ]);
                                        }}
                                        className="w-full mt-1 py-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-slate-100 font-extrabold text-[10px] tracking-widest uppercase rounded shadow transition duration-300 animate-pulse"
                                      >
                                        Concluir Missão de Fim de Jogo
                                      </button>
                                    )}

                                    {/* Permission Flag Button */}
                                    {abyssiQuestCompleted && !abyssiPermissionFlag && (
                                      <div className="flex flex-col gap-1.5 bg-slate-950 p-2 rounded border border-slate-800/60 mt-1">
                                        <span className="text-[10px] font-bold text-slate-300">3. Visto Militar do Grande Marechal:</span>
                                        <p className="text-[9px] text-slate-400 leading-normal">Seu navio necessita da credencial militar oficial de acesso de trânsito (Abyssi Access Permission Flag) emitida pela Resistência Humana.</p>
                                        <button
                                          onClick={() => {
                                            setAbyssiPermissionFlag(true);
                                            addSimLogs([
                                              { type: "info", text: "🎖️ [System] Credencial militar concedida: Flag 'ABYSSI_ACCESS_PERMISSION' adicionada com sucesso!" },
                                              { type: "info", text: "🎖️ [Travel] Autorização portuária para Last Bastion ativa. Rota liberada para viagem!" }
                                            ]);
                                          }}
                                          className="w-full py-1.5 bg-slate-800 hover:bg-slate-700 text-violet-400 font-bold border border-violet-500/20 text-[9px] rounded transition uppercase"
                                        >
                                          Assinar Pacto de Resistência e Obter Flag
                                        </button>
                                      </div>
                                    )}
                                  </div>
                                )}
                              </div>
                            )}

                            {/* Region Detail Card & Travel Button */}
                            {activeSet && (
                              <div className="bg-slate-900 border border-slate-800 rounded-lg p-3 flex flex-col gap-2">
                                <div className="flex justify-between items-start">
                                  <div>
                                    <div className="flex items-center gap-1.5">
                                      <span className="text-xs font-extrabold text-slate-200">{activeSet.name}</span>
                                      <span className="text-[8px] bg-slate-800 text-slate-400 px-1 py-0.2 rounded font-medium border border-slate-700/30">
                                        {activeSet.type}
                                      </span>
                                    </div>
                                    <span className="text-[9px] text-slate-500 font-medium">Líder: <strong className="text-slate-400">{activeSet.leader}</strong></span>
                                  </div>
                                  <span className="text-[9px] text-amber-400 font-mono bg-amber-950/40 border border-amber-900/30 px-1.5 py-0.2 rounded font-bold">
                                    [{activeSet.x}, {activeSet.y}]
                                  </span>
                                </div>

                                <p className="text-[10px] text-slate-400 leading-relaxed italic border-l-2 border-amber-500/30 pl-2">
                                  {activeSet.desc}
                                </p>

                                <div className="border-t border-slate-800/80 pt-2 mt-1">
                                  {activeSet.id === "heart_of_light" ? (
                                    <div className="bg-slate-950 border border-slate-800/60 rounded p-2 text-center text-[10px] text-slate-400 font-medium flex items-center justify-center gap-1.5">
                                      <Sparkles className="w-3.5 h-3.5 text-amber-400 animate-pulse" />
                                      <span>O Cristal Primordial é protegido. Zona de Adoração e Apenas Visualização.</span>
                                    </div>
                                  ) : (
                                    <button
                                      onClick={() => handleTravelSim(activeSet.name, activeSet.x, activeSet.y)}
                                      className={`w-full py-2 rounded text-xs font-extrabold tracking-wide transition flex items-center justify-center gap-1.5 shadow ${
                                        isLocked 
                                          ? "bg-slate-800/60 hover:bg-slate-800 text-slate-500 border border-slate-700/30 cursor-not-allowed" 
                                          : selectedContinent === "holy" && !holyQuestCompleted
                                          ? "bg-amber-600/20 hover:bg-amber-600/30 text-amber-500 border border-amber-500/30 cursor-not-allowed"
                                          : selectedContinent === "abyssia" && (!abyssiQuestCompleted || !abyssiPermissionFlag)
                                          ? "bg-violet-600/20 hover:bg-violet-600/30 text-violet-500 border border-violet-500/30 cursor-not-allowed"
                                          : "bg-amber-600 hover:bg-amber-500 text-slate-950 font-black hover:shadow-amber-950/30 border border-amber-500/20"
                                      }`}
                                    >
                                      <MapPin className="w-3.5 h-3.5" />
                                      {isLocked 
                                        ? `Acesso Bloqueado (Lvl < ${activeCont.minLevel})` 
                                        : selectedContinent === "holy" && !holyQuestCompleted
                                        ? "Requer Provação Espiritual"
                                        : selectedContinent === "abyssia" && !abyssiQuestCompleted
                                        ? "Requer Missão de Fim de Jogo"
                                        : selectedContinent === "abyssia" && !abyssiPermissionFlag
                                        ? "Requer Permissão de Trânsito"
                                        : `Viajar para ${activeSet.name} (CS_MOVE_REQUEST)`}
                                    </button>
                                  )}
                                </div>
                              </div>
                            )}
                          </div>
                        );
                      })()}

                      {/* Map status */}
                      <div className="bg-slate-900/40 border border-slate-800/60 p-2.5 rounded-lg text-center text-[10px] font-mono">
                        {progX === 100 && progY === 100 ? (
                          <span className="text-amber-400 font-semibold uppercase">Zona Inicial Segura Ativa (Safe Zone)</span>
                        ) : progX >= 3800 && progX <= 4000 && progY >= 3800 && progY <= 4000 ? (
                          <span className="text-purple-400 font-semibold uppercase">Explorando o Continente das Sombras</span>
                        ) : progX >= 4800 && progX <= 5000 && progY >= 4800 && progY <= 5000 ? (
                          <span className="text-amber-400 font-semibold uppercase">Solenidade no Continente Sagrado</span>
                        ) : progX >= 5800 && progX <= 6000 && progY >= 5800 && progY <= 6000 ? (
                          <span className="text-sky-400 font-semibold uppercase">Sobrevivendo ao Continente de Gelo</span>
                        ) : (
                          <span className="text-emerald-400 font-semibold uppercase">Navegando em Região Autorizada</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Visualizador & Simulador Interativo de Curva de XP e Recompensas (CANONICAL EXPERIENCE CURVE PATCH) */}
                <div className="border-t border-slate-800/80 pt-6 mt-4 flex flex-col gap-4">
                  <div className="flex items-center gap-2.5 pb-2">
                    <div className="p-1.5 bg-amber-500/10 border border-amber-500/20 rounded-lg animate-pulse">
                      <Compass className="w-5 h-5 text-amber-400" />
                    </div>
                    <div>
                      <h4 className="font-bold text-xs text-slate-100 uppercase tracking-wider">Simulador da Curva de Experiência Canônica & Balanceamento</h4>
                      <p className="text-[10px] text-slate-400">Verifique como o teto teórico de Nível 9999 mitiga a inflação e distribui as fontes de poder de combate.</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-12 gap-5">
                    {/* Left Column: Phase Progress Indicator & XP Calculator */}
                    <div className="md:col-span-7 bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 flex flex-col gap-4">
                      {(() => {
                        const phase = getPhaseInfo(progLevel);
                        const requiredXp = getRequiredXpForLevel(progLevel);
                        const monsterXp = getMonsterXpForLevel(progLevel);
                        const questXp = getQuestXpForLevel(progLevel);
                        const dungeonPct = getDungeonXpPercentForLevel(progLevel);

                        // Estimates
                        const killsNeeded = Math.ceil(requiredXp / monsterXp);
                        const questsNeeded = Math.ceil(requiredXp / questXp);

                        return (
                          <>
                            {/* Phase Timeline Tracker */}
                            <div className="flex flex-col gap-2">
                              <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Fase do Progresso (Canônica):</span>
                              <div className="grid grid-cols-4 gap-1.5 text-center">
                                {[
                                  { id: 1, label: "1-100 (FAST)", color: "border-emerald-500/30 text-emerald-400" },
                                  { id: 2, label: "101-200 (MODERATE)", color: "border-blue-500/30 text-blue-400" },
                                  { id: 3, label: "201-400 (SLOW)", color: "border-amber-500/30 text-amber-400" },
                                  { id: 4, label: "401+ (LEGEND)", color: "border-purple-500/30 text-purple-400" }
                                ].map(p => (
                                  <button
                                    key={p.id}
                                    onClick={() => {
                                      const defaultLvlMap = [50, 150, 300, 500];
                                      setProgLevel(p.id === 1 ? 50 : p.id === 2 ? 150 : p.id === 3 ? 300 : 500);
                                    }}
                                    className={`py-1 rounded text-[9px] font-mono border transition-all duration-300 hover:border-amber-500/40 ${
                                      phase.phase === p.id 
                                        ? "bg-slate-900 border-amber-500 text-amber-300 font-bold shadow shadow-amber-500/10 scale-[1.02]"
                                        : "bg-slate-950/40 border-slate-900/80 text-slate-500"
                                    }`}
                                  >
                                    {p.label}
                                  </button>
                                ))}
                              </div>
                            </div>

                            {/* Phase Detail Card */}
                            <div className={`p-3 border rounded-lg ${phase.color} flex flex-col gap-1 transition-all duration-300`}>
                              <div className="flex justify-between items-center">
                                <span className="font-bold text-xs uppercase tracking-wider">{phase.title}</span>
                                <span className="text-[9px] font-mono font-black border border-current px-2 py-0.5 rounded uppercase">
                                  {phase.badge}
                                </span>
                              </div>
                              <p className="text-[10px] text-slate-300 leading-normal">
                                {phase.desc}
                              </p>
                            </div>

                            {/* Main Math Details Grid */}
                            <div className="grid grid-cols-2 gap-4 mt-1">
                              <div className="bg-slate-900/40 p-3 rounded-lg border border-slate-900/60 flex flex-col justify-between">
                                <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider block">XP Necessário p/ Próximo Nível:</span>
                                <div className="mt-1">
                                  <span className="text-sm font-black font-mono text-amber-400">{formatXpNumber(requiredXp)}</span>
                                  <span className="text-[10px] text-slate-400 block mt-0.5 font-mono">Totais acumulados na fase</span>
                                </div>
                              </div>

                              <div className="bg-slate-900/40 p-3 rounded-lg border border-slate-900/60 flex flex-col justify-between">
                                <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider block">Massa de Kills Equivalente:</span>
                                <div className="mt-1">
                                  <span className="text-sm font-black font-mono text-rose-400">
                                    {progLevel === 9999 ? "Nenhum (Teto Máximo)" : `${killsNeeded.toLocaleString("pt-BR")} Monstros`}
                                  </span>
                                  <span className="text-[10px] text-slate-400 block mt-0.5 font-mono">Kills de mobs de mesmo nível</span>
                                </div>
                              </div>
                            </div>

                            {/* Systemic Balance Warning (9999 Level Interpretation) */}
                            <div className="bg-slate-900/50 border border-slate-800/40 p-3 rounded-lg text-[10.5px] leading-relaxed text-slate-400">
                              <div className="flex items-center gap-1.5 text-slate-300 font-bold mb-1">
                                <Info className="w-3.5 h-3.5 text-blue-400" />
                                <span>Filosofia de Level Cap de Light and Shadow</span>
                              </div>
                              Nível 9999 é uma meta teórica infinita para retenção de longo prazo. O jogo <strong className="text-slate-200">não é balanceado</strong> para atingir este número, eliminando a inflação de base de dados e tornando cada nível acima de 400 um troféu monumental de puro prestígio!
                            </div>
                          </>
                        );
                      })()}
                    </div>

                    {/* Right Column: Combat Power Balancing & Reward Scaling */}
                    <div className="md:col-span-5 bg-slate-950/80 border border-slate-800/80 rounded-xl p-5 flex flex-col gap-4">
                      {(() => {
                        const power = getCombatPowerBreakdown(progLevel);
                        const monsterXp = getMonsterXpForLevel(progLevel);
                        const questXp = getQuestXpForLevel(progLevel);
                        const dungeonPct = getDungeonXpPercentForLevel(progLevel);

                        return (
                          <>
                            {/* Sources of Combat Power */}
                            <div className="flex flex-col gap-3">
                              <div className="flex justify-between items-center">
                                <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Origem do Poder de Combate:</span>
                                <span className="text-[9px] text-slate-400 font-mono">Nível {progLevel}</span>
                              </div>

                              <div className="space-y-2.5">
                                {/* Level base stats */}
                                <div>
                                  <div className="flex justify-between text-[10px] mb-1 font-mono text-slate-400">
                                    <span>Atributos do Nível Puro</span>
                                    <span className="font-bold text-slate-300">{power.level}%</span>
                                  </div>
                                  <div className="w-full bg-slate-900 rounded-full h-1.5 overflow-hidden">
                                    <div className="bg-blue-500 h-full transition-all duration-500" style={{ width: `${power.level}%` }}></div>
                                  </div>
                                </div>

                                {/* Gear and Element Compatibility */}
                                <div>
                                  <div className="flex justify-between text-[10px] mb-1 font-mono text-slate-400">
                                    <span>Equipamentos e Afinidades (Gear)</span>
                                    <span className="font-bold text-amber-400">{power.gear}%</span>
                                  </div>
                                  <div className="w-full bg-slate-900 rounded-full h-1.5 overflow-hidden">
                                    <div className="bg-amber-500 h-full transition-all duration-500" style={{ width: `${power.gear}%` }}></div>
                                  </div>
                                </div>

                                {/* Combat Execution and rotation */}
                                <div>
                                  <div className="flex justify-between text-[10px] mb-1 font-mono text-slate-400">
                                    <span>Maestria e Execução em Combate</span>
                                    <span className="font-bold text-emerald-400">{power.skill}%</span>
                                  </div>
                                  <div className="w-full bg-slate-900 rounded-full h-1.5 overflow-hidden">
                                    <div className="bg-emerald-500 h-full transition-all duration-500" style={{ width: `${power.skill}%` }}></div>
                                  </div>
                                </div>
                              </div>

                              <p className="text-[10px] text-slate-500 leading-normal italic mt-1 font-mono">
                                *Nota: Após o Midgame, o ganho de atributos por nível é mitigado. O combate exige equipamentos corretos e rotações limpas.
                              </p>
                            </div>

                            {/* Reward Scaling Table / Sincronia de Recompensas */}
                            <div className="flex flex-col gap-2 mt-1 border-t border-slate-800/60 pt-3">
                              <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Escalonamento de Recompensas do Mundo:</span>
                              
                              <div className="space-y-1.5 text-[10.5px] font-mono">
                                <div className="flex justify-between py-1 border-b border-slate-900/80">
                                  <span className="text-slate-400">Recompensa Quest Padrão:</span>
                                  <span className="text-amber-300 font-bold">+{formatXpNumber(questXp)} XP</span>
                                </div>
                                <div className="flex justify-between py-1 border-b border-slate-900/80">
                                  <span className="text-slate-400">XP por Monstro Abatido:</span>
                                  <span className="text-amber-300 font-bold">+{formatXpNumber(monsterXp)} XP</span>
                                </div>
                                <div className="flex justify-between py-1">
                                  <span className="text-slate-400">Mitigação de Masmorra:</span>
                                  <span className="text-violet-300 font-bold">{dungeonPct} do Nível</span>
                                </div>
                              </div>
                            </div>
                          </>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              </div>

            </motion.div>
          )}

          {/* Tab 1.5: Bestiário Canônico */}
          {activeTab === "bestiary" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="flex flex-col gap-6 animate-fade-in"
            >
              {/* Header Card */}
              <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 backdrop-blur-md shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 bg-amber-500/10 rounded-xl border border-amber-500/20">
                    <Sparkles className="w-7 h-7 text-amber-400" />
                  </div>
                  <div>
                    <h2 className="text-lg font-bold text-slate-100 tracking-tight">
                      Bestiário Canônico & Leis de Ecossistema (Monster AI Patch)
                    </h2>
                    <p className="text-xs text-slate-400 leading-relaxed">
                      Sistemas de classificação de 20 Famílias, Spawn Híbrido Dinâmico, Hierarquia de Demônios e Leis Territoriais de Dragões de acordo com o Monster AI Bible.
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[10px] bg-amber-500/10 text-amber-400 border border-amber-500/20 px-3 py-1 rounded-full font-mono font-bold tracking-wider uppercase">
                    Authoritative Bible Rules
                  </span>
                </div>
              </div>

              {/* Three-Column Architecture */}
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                
                {/* Left Column: 20 Bestiary Families */}
                <div className="lg:col-span-4 flex flex-col gap-4">
                  <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                      <BookOpen className="w-4 h-4 text-amber-400" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Famílias Oficiais (20)</h4>
                    </div>
                    
                    <p className="text-[10.5px] text-slate-400 leading-normal">
                      Hostis classificados estritamente por Famílias e Subfamílias. Rótulos artificiais como <span className="text-rose-400 font-bold">Comum/Épico/Lendário</span> são estritamente de balanceamento interno e não fazem parte da classificação canônica.
                    </p>

                    <div className="bg-rose-500/10 border border-rose-500/20 rounded-lg p-2.5 flex flex-col gap-1">
                      <div className="flex items-center gap-1.5 text-xs text-rose-300 font-bold">
                        <AlertTriangle className="w-3.5 h-3.5 text-rose-400" />
                        <span>Regra: Family ≠ Intelligence Tier</span>
                      </div>
                      <p className="text-[9.5px] text-slate-400 leading-normal">
                        Famílias definem apenas biologia, ontologia, papel ecológico e distribuição. Inteligência e sofisticação tática são de perfil exclusivo do Monster AI.
                      </p>
                    </div>

                    {/* Compact Family Scroller/Grid */}
                    <div className="grid grid-cols-2 gap-1.5 max-h-[300px] overflow-y-auto pr-1">
                      {[
                        { id: "humanoids", name: "Humanoids", examples: "Goblins, Orcs, Bandits, Pirates, Cultists" },
                        { id: "beasts", name: "Beasts", examples: "Wolves, Bears, Boars" },
                        { id: "predators", name: "Predators", examples: "Dire Wolves, Panthers, Sabertooths" },
                        { id: "insects_vermin", name: "Insects / Vermin", examples: "Giant Ants, Spiders, Scorpions, Beetles" },
                        { id: "reptilians", name: "Reptilians", examples: "Lizardmen, Basilisks, Serpents, Nagas" },
                        { id: "avians", name: "Avians", examples: "Harpies, Vultures, Ravens, Giant Eagles" },
                        { id: "aquatics", name: "Aquatics", examples: "Sharks, Sirens, Krakens, Sea Serpents" },
                        { id: "amphibians", name: "Amphibians", examples: "Giant Frogs, Toads, Salamanders" },
                        { id: "undead", name: "Undead", examples: "Skeletons, Zombies, Liches, Wraiths" },
                        { id: "spirits", name: "Spirits", examples: "Ghosts, Wisps, Phantoms" },
                        { id: "elementals", name: "Elementals", examples: "Fire, Ice, Storm Elementals" },
                        { id: "flora_creatures", name: "Flora Creatures", examples: "Treants, Carnivorous Plants, Vine Horrors" },
                        { id: "fungi", name: "Fungi", examples: "Mushroom Walkers, Spores, Mycelial" },
                        { id: "constructs", name: "Constructs", examples: "Golems, Animated Armor, Arcane Sentinels" },
                        { id: "cursed_beings", name: "Cursed Beings", examples: "Werewolves, Cursed Knights, Corrupted" },
                        { id: "aberrations", name: "Aberrations", examples: "Void Horrors, Flesh Mutants, Anomalies" },
                        { id: "demons", name: "Demons", examples: "Special Family (Tied to Primordials)" },
                        { id: "dragons", name: "Dragons", examples: "Special Family (Independent)" },
                        { id: "titans_colossals", name: "Titans / Colossals", examples: "Giants, Cyclops, Colossi" },
                        { id: "celestials", name: "Celestials", examples: "Seraphs, Sacred Guardians, Divine Beasts" }
                      ].map(fam => {
                        const isSel = selectedBestiaryFamily === fam.id;
                        return (
                          <button
                            key={fam.id}
                            onClick={() => setSelectedBestiaryFamily(fam.id)}
                            className={`p-2 rounded-lg border text-left transition-all duration-200 ${
                              isSel 
                                ? "bg-amber-500/10 border-amber-500 text-amber-300 font-bold shadow-md shadow-amber-950/20" 
                                : "bg-slate-900 border-slate-800 text-slate-400 hover:text-slate-200 hover:border-slate-700"
                            }`}
                          >
                            <span className="text-[11px] block truncate">{fam.name}</span>
                            <span className="text-[8px] text-slate-500 block truncate">{fam.examples}</span>
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  {/* Family Detail Card */}
                  {(() => {
                    const data: Record<string, { desc: string; sub: string[]; ex: string; rule: string; regionRule: string }> = {
                      humanoids: {
                        desc: "Ontologia humanoide abrangendo diversas tribos, sociedades e subculturas.",
                        sub: ["Goblinoids", "Orcs", "Outlaws"],
                        ex: "Goblins, Orcs, Bandidos, Piratas, Cultistas",
                        rule: "Classificação biológica para espécies humanoides sapientes com estruturas sociais complexas.",
                        regionRule: "Dispersos em todas as florestas e estepes litorâneas."
                      },
                      beasts: {
                        desc: "Fauna selvagem nativa operando por mero instinto territorial básico.",
                        sub: ["Canines", "Ursines", "Suidae"],
                        ex: "Lobos, Ursos, Javalis",
                        rule: "Fauna selvagem nativa que atua como peça fundamental na cadeia alimentar natural.",
                        regionRule: "Habitam florestas e cadeias montanhosas comuns."
                      },
                      predators: {
                        desc: "Predadores carnívoros nativos de grande porte no topo da cadeia ecológica.",
                        sub: ["Apex Canines", "Felines"],
                        ex: "Lobos Dire, Panteras, Dentes-de-Sabre",
                        rule: "Predadores carnívoros nativos de grande porte no topo da cadeia ecológica.",
                        regionRule: "Zonas de selva profunda e ruínas abandonadas."
                      },
                      insects_vermin: {
                        desc: "Invertebrados colossais e insetóides adaptados a ambientes subterrâneos e úmidos.",
                        sub: ["Formicidae", "Arachnids", "Scorpionidae", "Coleoptera"],
                        ex: "Formigas Gigantes, Aranhas, Escorpiões, Besouros",
                        rule: "Invertebrados colossais e insetóides adaptados a ambientes subterrâneos e úmidos.",
                        regionRule: "Túneis subterrâneos, desertos e pântanos profundos."
                      },
                      reptilians: {
                        desc: "Espécies de sangue frio que veneram elementos e guardam passagens antigas.",
                        sub: ["Saurians", "Basilisks", "Serpents", "Nagas"],
                        ex: "Lizardmen, Basiliscos, Serpentes, Nagas",
                        rule: "Econatureza de sangue frio adaptada a biomas de alta umidade, pântanos e rios.",
                        regionRule: "Zonas costeiras tropicais, pântanos e masmorras úmidas."
                      },
                      avians: {
                        desc: "Criaturas aladas capazes de rasgar os céus com extrema velocidade e mobilidade vertical.",
                        sub: ["Harpies", "Scavengers", "Raptors"],
                        ex: "Hárpias, Abutres, Corvos, Águias Gigantes",
                        rule: "Criaturas voadoras adaptadas para controle ecológico e caça aérea.",
                        regionRule: "Cumes montanhosos elevados e desfiladeiros rochosos."
                      },
                      aquatics: {
                        desc: "Habitantes das profundezas abissais insondáveis dos mares e rios profundos.",
                        sub: ["Selachimorpha", "Sirens", "Leviathans"],
                        ex: "Tubarões, Sereias, Krakens, Serpentes Marinhas",
                        rule: "Fauna marinha adaptada exclusivamente a águas profundas e recifes costeiros.",
                        regionRule: "Oceanos abertos, templos submersos e golfos isolados."
                      },
                      amphibians: {
                        desc: "Seres anfíbios que se movem com maestria entre a terra firme e as águas turvas.",
                        sub: ["Anura", "Caudata"],
                        ex: "Sapos Gigantes, Rãs, Salamandras",
                        rule: "Criaturas de transição terra-água com biologia altamente adaptável.",
                        regionRule: "Zonas pantanosas, manguezais e cavernas molhadas."
                      },
                      undead: {
                        desc: "Cadáveres reanimados e guerreiros caídos escravizados por corrupção necrótica eterna.",
                        sub: ["Skeletal", "Corpses", "Liches", "Wraiths"],
                        ex: "Esqueletos, Zumbis, Liches, Aparições",
                        rule: "Entidades ontológicas reanimadas por magia necrótica e corrupção espiritual.",
                        regionRule: "Cemitérios antigos, criptas escuras e pântanos de névoa."
                      },
                      spirits: {
                        desc: "Resquícios incorpóreos de almas presas no plano mortal por tragédias ou rituais.",
                        sub: ["Apparitions", "Wisps"],
                        ex: "Fantasmas, Fogos Fátuos, Phantoms",
                        rule: "Manifestações incorpóreas de energia e resíduos de consciência.",
                        regionRule: "Locais de batalhas históricas ou florestas místicas à noite."
                      },
                      elementals: {
                        desc: "Manifestações brutas de energias primordiais da natureza encarnadas fisicamente.",
                        sub: ["Pyroclastics", "Cryogenics", "Tempests"],
                        ex: "Elementais de Fogo, Gelo, Tempestade",
                        rule: "Constructos de energia elemental bruta moldados pelas forças fundamentais da natureza.",
                        regionRule: "Fendas elementais activas, cumes vulcânicos e glaciares."
                      },
                      flora_creatures: {
                        desc: "Vegetação e árvores despertas por correntes de magia antiga ou fungos.",
                        sub: ["Arboreal", "Carnivorous Flora", "Vines"],
                        ex: "Treants, Plantas Carnívoras, Horrores de Vinha",
                        rule: "Formas de vida vegetal senciente que servem de âncora para ecossistemas de florestas densas.",
                        regionRule: "Florestas ancestrais e estufas esquecidas."
                      },
                      fungi: {
                        desc: "Ecosporas sencientes capazes de espalhar mofo de decomposição em criaturas vivas.",
                        sub: ["Mycelials", "Spores"],
                        ex: "Caminhantes Cogumelo, Esporos da Praga, Horrores Myceliais",
                        rule: "Organismos decompositores que auxiliam na ciclagem de nutrientes e esporas.",
                        regionRule: "Pântanos sombrios, cavernas úmidas e o Submundo."
                      },
                      constructs: {
                        desc: "Maquinários arcanos e runas de engenharia programadas para defesa eterna.",
                        sub: ["Golems", "Animated", "Arcane Sentinels"],
                        ex: "Golems de Pedra, Armaduras Animadas, Sentinelas Arcanas",
                        rule: "Entidades artificiais inanimadas e programadas por runas arcanas.",
                        regionRule: "Masmorras arcanas, laboratórios de alquimia e templos antigos."
                      },
                      cursed_beings: {
                        desc: "Mortais afligidos por aflições de sangue, transformados em bestas ou demônios parciais.",
                        sub: ["Lycanthropes", "Corrupted Mortals"],
                        ex: "Lobisomens, Cavaleiros Amaldiçoados, Monges Corrompidos",
                        rule: "Antigos mortais afligidos por anomalias mágicas e corrupção de sangue.",
                        regionRule: "Castelos em ruínas, abadias profanadas e florestas malditas."
                      },
                      aberrations: {
                        desc: "Anomalias desfiguradas oriundas do Vazio que corrompem a própria física do mundo.",
                        sub: ["Void Horrors", "Flesh Mutants"],
                        ex: "Horrores do Vazio, Mutantes de Carne, Anomalias Cósmicas",
                        rule: "Anomalias do Vazio fora das leis da biologia convencional do continente.",
                        regionRule: "Fendas cósmicas, crateras de meteoros e poços abissais."
                      },
                      demons: {
                        desc: "Família canônica especial originária de reinos profanos e vinculada diretamente aos Primordiais.",
                        sub: ["Common Hierarchy", "Minor Hierarchy", "Commanders", "Archegenerals"],
                        ex: "Imps, Cães do Inferno, Pit Lords, Arquigenerais",
                        rule: "Entidades originárias de fendas e planos profanos vinculadas diretamente aos Primordiais.",
                        regionRule: "Zonas de invasão ativa, fendas infernais e continentes Primordiais."
                      },
                      dragons: {
                        desc: "Família canônica especial independente dos Primordiais, governada por territorialismo e ouro.",
                        sub: ["Lesser Drakes", "Juvenile Dragons", "Adult Dragons", "Sovereign Dragons"],
                        ex: "Wyverns, Dracos, Dragões Anciãos, Void Dragon Lord",
                        rule: "Soberanos antigos e independentes da criação primordial, de alta longevidade.",
                        regionRule: "Cumes vulcânicos isolados, cavernas de gelo e covis protegidos."
                      },
                      titans_colossals: {
                        desc: "Monólitos vivos da era dos gigantes com força física telúrica incomensurável.",
                        sub: ["Giants", "Cyclopes", "Colossi"],
                        ex: "Gigantes, Ciclope, Colosso de Ruínas",
                        rule: "Seres colossais de eras antigas que atuam como monólitos históricos vivos.",
                        regionRule: "Planaltos áridos, ruínas desérticas e passagens do norte."
                      },
                      celestials: {
                        desc: "Protetores solares radiantes que descem do plano de luz para punir transgressões.",
                        sub: ["Seraphs", "Sacred Guardians", "Divine Beasts"],
                        ex: "Serafins, Guardiões Sagrados, Bestas Divinas",
                        rule: "Seres de pura luz e ordem vinculados a santuários e planos superiores.",
                        regionRule: "Templos das nuvens, santuários de luz e altares sagrados."
                      }
                    };

                    const selectedData = data[selectedBestiaryFamily] || data.humanoids;
                    return (
                      <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                        <div className="flex justify-between items-center border-b border-slate-800/40 pb-2">
                          <span className="text-xs font-bold text-amber-400 uppercase tracking-wider font-mono">
                            Ficha Técnica: {selectedBestiaryFamily.toUpperCase()}
                          </span>
                          <span className="text-[9px] bg-slate-950/60 border border-slate-800 text-slate-400 px-2 py-0.5 rounded font-bold font-mono">
                            Hierarquia Ativa
                          </span>
                        </div>

                        <p className="text-xs text-slate-200 leading-normal">
                          {selectedData.desc}
                        </p>

                        <div className="grid grid-cols-2 gap-3 text-[11px] pt-1">
                          <div>
                            <span className="text-slate-500 font-bold block uppercase text-[9px] tracking-wider">Subfamílias Canônicas:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                              {selectedData.sub.map(s => (
                                <span key={s} className="bg-slate-950 border border-slate-800 text-slate-300 px-1.5 py-0.5 rounded text-[10px] font-mono">
                                  {s}
                                </span>
                              ))}
                            </div>
                          </div>
                          <div>
                            <span className="text-slate-500 font-bold block uppercase text-[9px] tracking-wider">Exemplos Práticos:</span>
                            <span className="text-slate-300 block mt-1 font-mono text-[10px] leading-relaxed">
                              {selectedData.ex}
                            </span>
                          </div>
                        </div>

                        <div className="border-t border-slate-800/40 pt-2.5 text-[11px] flex flex-col gap-1.5 leading-relaxed">
                          <p>
                            <strong className="text-amber-400/90 font-bold">Função Ecológica & Biológica:</strong> {selectedData.rule}
                          </p>
                          <p>
                            <strong className="text-amber-400/90 font-bold">Distribuição Ecológica:</strong> {selectedData.regionRule}
                          </p>
                        </div>
                      </div>
                    );
                  })()}
                </div>

                {/* Middle Column: Hybrid Spawn & Demon Spawn Laws */}
                <div className="lg:col-span-4 flex flex-col gap-4">
                  {/* Spawn Simulator Card */}
                  <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                      <Compass className="w-4 h-4 text-emerald-400" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Simulador de Spawn Híbrido</h4>
                    </div>

                    <p className="text-[10.5px] text-slate-400 leading-normal">
                      Os tempos e regras de renascimento são calculados com base na categoria e pressão dos jogadores no servidor local.
                    </p>

                    <div className="space-y-3 pt-1">
                      <div className="space-y-1">
                        <label className="text-[10px] text-slate-400 font-bold uppercase">Categoria de Spawn:</label>
                        <select
                          value={spawnCategory}
                          onChange={(e) => setSpawnCategory(e.target.value as any)}
                          className="w-full bg-slate-900 border border-slate-800 rounded-lg p-2 text-xs text-slate-200 outline-none focus:border-emerald-500/40"
                        >
                          <option value="Standard Spawn">Standard Spawn (Fixo: 30s a 5m)</option>
                          <option value="Dynamic Ecosystem Spawn">Dynamic Ecosystem (Ecológico - Adaptativo)</option>
                          <option value="Rare Spawn">Rare Spawn (Janela: 30m a 24h+)</option>
                          <option value="Legendary Spawn">Legendary Spawn (Baseado em Gatilhos)</option>
                        </select>
                      </div>

                      {spawnCategory === "Dynamic Ecosystem Spawn" && (
                        <div className="bg-slate-900/60 border border-slate-800/60 rounded-lg p-3 space-y-3">
                          <div className="space-y-1">
                            <div className="flex justify-between items-center text-[10px] font-semibold text-slate-400">
                              <span>Densidade de Jogadores (Área):</span>
                              <span className="text-emerald-400 font-mono">{playerDensity} players</span>
                            </div>
                            <input
                              type="range"
                              min="0"
                              max="50"
                              value={playerDensity}
                              onChange={(e) => setPlayerDensity(parseInt(e.target.value))}
                              className="w-full accent-emerald-500"
                            />
                          </div>

                          <div className="space-y-1">
                            <label className="text-[10px] text-slate-400 font-semibold block">Pressão de Caça do Clã:</label>
                            <div className="grid grid-cols-3 gap-1.5">
                              {["low", "medium", "high"].map(lvl => (
                                <button
                                  key={lvl}
                                  onClick={() => setHuntingPressure(lvl as any)}
                                  className={`py-1 rounded text-[9px] font-bold uppercase border transition ${
                                    huntingPressure === lvl
                                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                                      : "bg-slate-950 text-slate-500 border-slate-800 hover:border-slate-700"
                                  }`}
                                >
                                  {lvl === "low" ? "Baixa" : lvl === "medium" ? "Média" : "Alta"}
                                </button>
                              ))}
                            </div>
                          </div>
                        </div>
                      )}

                      <button
                        onClick={handleSimulateSpawn}
                        className="w-full py-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5 animate-pulse"
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                        Simular Evento de Spawn
                      </button>

                      {spawnSimResult && (
                        <div className="p-2.5 bg-slate-900 rounded-lg border border-slate-800 text-[10.5px] font-mono leading-relaxed text-slate-300">
                          {spawnSimResult}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Demon Spawn Law Card */}
                  <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                      <Flame className="w-4 h-4 text-red-500 animate-pulse" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Mecânica de Spawn Demônio</h4>
                    </div>

                    <p className="text-[10.5px] text-slate-400 leading-normal">
                      A lei demôniaca rege o nascimento em continentes primordiais e proíbe spawns de demônios naturais no Continente Principal.
                    </p>

                    <div className="space-y-3 pt-1">
                      <div className="grid grid-cols-2 gap-2">
                        <div className="space-y-1">
                          <label className="text-[9px] text-slate-400 font-bold uppercase block">Invocador:</label>
                          <select
                            value={demonSummoner}
                            onChange={(e) => setDemonSummoner(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2 py-1 text-[11px] text-slate-200 outline-none"
                          >
                            <option value="Common Demons">Common Demons</option>
                            <option value="Minor Demons">Minor Demons</option>
                            <option value="Commanders">Commanders</option>
                            <option value="Archegenerals">Archegenerals</option>
                          </select>
                        </div>
                        <div className="space-y-1">
                          <label className="text-[9px] text-slate-400 font-bold uppercase block">Alvo Invocado:</label>
                          <select
                            value={demonSummonTarget}
                            onChange={(e) => setDemonSummonTarget(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2 py-1 text-[11px] text-slate-200 outline-none"
                          >
                            <option value="Common Demons">Common Demons</option>
                            <option value="Minor Demons">Minor Demons</option>
                            <option value="Commanders">Commanders</option>
                          </select>
                        </div>
                      </div>

                      <div className="space-y-1">
                        <label className="text-[9px] text-slate-400 font-bold uppercase block">Região Geográfica:</label>
                        <select
                          value={demonRegion}
                          onChange={(e) => setDemonRegion(e.target.value as any)}
                          className="w-full bg-slate-900 border border-slate-800 rounded px-2 py-1 text-[11px] text-slate-200 outline-none focus:border-red-500/40"
                        >
                          <option value="Main Continent">Continente Principal (No Demon Spawning)</option>
                          <option value="Volcanic Demon Zone">Zona de Demônios Vulcânicos (Primordial Area)</option>
                          <option value="Deep Shadow Territory">Território das Sombras Profundas</option>
                        </select>
                      </div>

                      <button
                        onClick={handleSimulateDemonSummon}
                        className="w-full py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5"
                      >
                        <Sword className="w-3.5 h-3.5" />
                        Validar Regras de Invocação
                      </button>

                      <div className="p-2.5 bg-slate-900 rounded-lg border border-slate-800 text-[10.5px] font-mono leading-relaxed text-slate-300 text-center">
                        {demonSummonLog}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Right Column: Dragon Spawn Laws & Lair Hoarding */}
                <div className="lg:col-span-4 flex flex-col gap-4">
                  {/* Dragon Spawn and Territorial Law Card */}
                  <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                      <Shield className="w-4 h-4 text-rose-400" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Spawn e Territorialidade de Dragões</h4>
                    </div>

                    <p className="text-[10.5px] text-slate-400 leading-normal">
                      "Two apex dragons rarely tolerate the same territory." Dragões juvenis ou menores podem atuar como subordinados territoriais.
                    </p>

                    <div className="space-y-3 pt-1">
                      <div className="grid grid-cols-2 gap-2">
                        <div className="space-y-1">
                          <label className="text-[9px] text-slate-400 font-bold uppercase block">Dragão Primário:</label>
                          <select
                            value={dragonOne}
                            onChange={(e) => setDragonOne(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1 text-[11px] text-slate-200 outline-none"
                          >
                            <option value="void_dragon_lord">Void Dragon Lord (Soberano)</option>
                            <option value="volcanic_dragon">Volcanic Dragon (Adulto Apex)</option>
                            <option value="ancient_glacier_dragon">Ancient Glacier (Adulto Apex)</option>
                            <option value="wyvern">Wyvern (Dragão Menor)</option>
                          </select>
                        </div>
                        <div className="space-y-1">
                          <label className="text-[9px] text-slate-400 font-bold uppercase block">Dragão Secundário:</label>
                          <select
                            value={dragonTwo}
                            onChange={(e) => setDragonTwo(e.target.value)}
                            className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1 text-[11px] text-slate-200 outline-none"
                          >
                            <option value="void_dragon_lord">Void Dragon Lord (Soberano)</option>
                            <option value="volcanic_dragon">Volcanic Dragon (Adulto Apex)</option>
                            <option value="ancient_glacier_dragon">Ancient Glacier (Adulto Apex)</option>
                            <option value="wyvern">Wyvern (Dragão Menor)</option>
                          </select>
                        </div>
                      </div>

                      <div className="space-y-1">
                        <label className="text-[9px] text-slate-400 font-bold uppercase block">Território de Disputa:</label>
                        <select
                          value={dragonTerritory}
                          onChange={(e) => setDragonTerritory(e.target.value)}
                          className="w-full bg-slate-900 border border-slate-800 rounded px-2.5 py-1 text-[11px] text-slate-200 outline-none focus:border-rose-500/40"
                        >
                          <option value="Ancient Volcano Peak">Ancient Volcano Peak</option>
                          <option value="Shadow Ruin Caverns">Shadow Ruin Caverns</option>
                          <option value="Frozen Glacier Crevice">Frozen Glacier Crevice</option>
                        </select>
                      </div>

                      <button
                        onClick={handleSimulateDragonTerritory}
                        className="w-full py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/30 rounded-lg text-xs font-bold transition flex items-center justify-center gap-1.5"
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                        Verificar Tolerância Territorial
                      </button>

                      <div className="p-2.5 bg-slate-900 rounded-lg border border-slate-800 text-[10.5px] font-mono leading-relaxed text-slate-300 text-center">
                        {dragonCombatLog}
                      </div>
                    </div>
                  </div>

                  {/* Hoarding Law / Treasure Pile Card */}
                  <div className="bg-slate-950/80 border border-slate-800/80 rounded-xl p-4 flex flex-col gap-3">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-2">
                      <Gift className="w-4 h-4 text-amber-400 animate-pulse" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Lei de Acumulação de Tesouro (Hoarding)</h4>
                    </div>

                    <div className="bg-amber-500/5 border border-amber-500/10 rounded-lg p-3 space-y-2.5">
                      <p className="text-xs italic text-amber-300 font-medium">
                        “Dragons accumulate valuable objects within their lairs.”
                      </p>
                      <p className="text-[10.5px] text-slate-400 leading-relaxed">
                        Esta regra atua estritamente no âmbito do <strong>design de lore</strong> neste momento. Sistemas de geração de saques, tabelas de drop e fórmulas de valor de mercado não estão ativos.
                      </p>
                    </div>

                    <div className="space-y-1.5">
                      <span className="text-[9px] text-slate-500 font-bold uppercase block tracking-wider">Elementos de Covis Permitidos (Apenas Lore):</span>
                      <div className="grid grid-cols-2 gap-1.5">
                        {["Ouro", "Artefatos", "Relíquias", "Objetos Antigos", "Recursos Raros"].map((item) => (
                          <div key={item} className="bg-slate-900 border border-slate-800/60 px-2 py-1.5 rounded text-[10px] font-mono text-slate-300 flex items-center gap-1">
                            <span className="text-amber-500 font-bold">▪</span>
                            {item}
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>

              </div>
            </motion.div>
          )}

          {/* Tab: Loot & Penalidades (Canonical Loot System & Death Penalty Bibles) */}
          {activeTab === "loot_death" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 lg:grid-cols-12 gap-6"
            >
              {/* Left Column: Player Status & Loot Filter */}
              <div className="lg:col-span-4 space-y-6">
                
                {/* Player Status Card */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-xl space-y-4">
                  <div className="flex items-center gap-2 border-b border-slate-800/60 pb-3">
                    <UserCheck className="w-5 h-5 text-amber-400" />
                    <div>
                      <h3 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Status do Herói</h3>
                      <p className="text-[10px] text-slate-400 font-mono">Gabriela_Paladin • Level 60</p>
                    </div>
                  </div>

                  {/* Equipped Gear */}
                  <div className="space-y-2">
                    <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">Equipamentos Equipados (Item Loss Law):</span>
                    <div className="grid grid-cols-2 gap-2">
                      {["Arma", "Armadura", "Acessório", "Escudo"].map(slot => {
                        const equipped = playerEquipped.find(e => e.slot === slot);
                        return (
                          <div key={slot} className="bg-slate-950/80 border border-slate-850 p-2.5 rounded-lg flex flex-col justify-between min-h-[56px]">
                            <span className="text-[8px] text-slate-500 uppercase font-bold">{slot}</span>
                            {equipped ? (
                              <span className="text-[10.5px] font-semibold text-amber-300 truncate">{equipped.name}</span>
                            ) : (
                              <span className="text-[10.5px] text-rose-500 font-medium italic">Nenhum (Perdido)</span>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Backpack Inventory */}
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">Inventário de Mochila:</span>
                      <span className="text-[10px] font-mono text-amber-400 font-bold">{playerInventory.length} / 12 slots</span>
                    </div>
                    <div className="bg-slate-950/85 border border-slate-850 rounded-xl p-3 max-h-[180px] overflow-y-auto space-y-1.5 scrollbar-thin scrollbar-thumb-slate-800">
                      {playerInventory.length === 0 ? (
                        <div className="text-center py-6 text-xs text-slate-500 italic">Mochila vazia. Abata monstros para obter loot!</div>
                      ) : (
                        playerInventory.map(item => (
                          <div key={item.id} className="flex justify-between items-center bg-slate-900/60 border border-slate-800/40 px-2.5 py-1.5 rounded-lg">
                            <div className="flex flex-col">
                              <span className="text-[11px] font-medium text-slate-200">{item.name}</span>
                              <span className="text-[9px] text-slate-500 font-mono">Tipo: {item.type} • Val: {item.value} cada</span>
                            </div>
                            <span className="text-xs font-mono font-bold text-amber-400">x{item.qty}</span>
                          </div>
                        ))
                      )}
                    </div>
                  </div>

                  {/* Blessings & Death Trigger */}
                  <div className="border-t border-slate-800/60 pt-3 space-y-3">
                    <div className="bg-slate-950/50 p-2.5 rounded-lg border border-slate-850 space-y-2">
                      <div className="flex justify-between items-center">
                        <div className="flex flex-col">
                          <span className="text-[10.5px] font-bold text-slate-300 flex items-center gap-1">
                            <Shield className="w-3.5 h-3.5 text-amber-400" />
                            Bênçãos Ativas ({activeBlessingsCount}/7)
                          </span>
                          <span className="text-[9px] text-slate-500">Proteção linear contra perda na morte PvE</span>
                        </div>
                        <span className="text-xs font-mono font-bold text-amber-400">
                          {((activeBlessingsCount / 7) * 100).toFixed(1)}%
                        </span>
                      </div>

                      {/* Horizontal Selector 0 to 7 */}
                      <div className="flex items-center justify-between gap-1 bg-slate-900/60 p-1 rounded-md border border-slate-800/40">
                        {[0, 1, 2, 3, 4, 5, 6, 7].map((num) => (
                          <button
                            key={num}
                            onClick={() => setActiveBlessingsCount(num)}
                            className={`w-7 h-6 flex items-center justify-center text-[10px] font-mono font-bold rounded transition-all duration-150 ${
                              activeBlessingsCount === num
                                ? "bg-amber-500 text-slate-950 shadow-md shadow-amber-500/20 scale-105"
                                : "text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                            }`}
                          >
                            {num}
                          </button>
                        ))}
                      </div>

                      <div className="text-[8.5px] text-slate-400 text-center font-mono leading-tight">
                        {activeBlessingsCount === 7 
                          ? "Proteção Total (100%): Sem perda de itens equipados ou inventário"
                          : activeBlessingsCount === 0
                          ? "Sem Proteção (0%): Risco de perda total na morte"
                          : `Proteção de ${((activeBlessingsCount / 7) * 100).toFixed(3)}% contra perda`}
                      </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      <button
                        onClick={() => handleSimulatePlayerDeath(false)}
                        className="py-2.5 bg-rose-650 hover:bg-rose-700 text-slate-950 font-bold rounded-xl text-[10.5px] transition duration-200 flex items-center justify-center gap-1.5 uppercase tracking-wider"
                      >
                        <AlertTriangle className="w-3.5 h-3.5 text-slate-950" />
                        Morte PvE
                      </button>
                      <button
                        onClick={() => handleSimulatePlayerDeath(true)}
                        className="py-2.5 bg-purple-600 hover:bg-purple-700 text-slate-100 font-bold rounded-xl text-[10.5px] transition duration-200 flex items-center justify-center gap-1.5 uppercase tracking-wider shadow-md shadow-purple-900/30"
                      >
                        <Sword className="w-3.5 h-3.5 text-slate-100" />
                        Morte PvP
                      </button>
                    </div>
                  </div>
                </div>

                {/* Custom Loot Filter Card */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-xl space-y-4">
                  <div className="flex items-center gap-2 border-b border-slate-800/60 pb-3">
                    <Settings className="w-5 h-5 text-amber-400" />
                    <div>
                      <h3 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Filtro de Loot</h3>
                      <p className="text-[10px] text-slate-400">Direct Inventory Auto-Loot Filter Pipeline</p>
                    </div>
                  </div>

                  <div className="space-y-3 text-xs">
                    <div className="space-y-1">
                      <label className="text-[10px] text-slate-400 uppercase tracking-wider font-bold">Tipo de Item Permitido:</label>
                      <select
                        value={lootFilterType}
                        onChange={e => setLootFilterType(e.target.value)}
                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-300 font-medium font-sans"
                      >
                        <option value="all">Qualquer Tipo (Auto-Loot Tudo)</option>
                        <option value="material">Apenas Materiais de Criação</option>
                        <option value="consumable">Apenas Consumíveis</option>
                        <option value="currency">Apenas Moedas (Currency)</option>
                      </select>
                    </div>

                    <div className="space-y-1">
                      <label className="text-[10px] text-slate-400 uppercase tracking-wider font-bold">Nome do Item Contém:</label>
                      <input
                        type="text"
                        value={lootFilterName}
                        onChange={e => setLootFilterName(e.target.value)}
                        placeholder="Ex: Demônio, Fragmento, Ouro..."
                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-300 text-xs"
                      />
                    </div>

                    <div className="space-y-1">
                      <label className="text-[10px] text-slate-400 uppercase tracking-wider font-bold">Valor Mínimo (Moedas):</label>
                      <input
                        type="number"
                        value={lootFilterMinVal}
                        onChange={e => setLootFilterMinVal(parseInt(e.target.value) || 0)}
                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-300 text-xs font-mono"
                      />
                    </div>
                  </div>
                </div>

              </div>

              {/* Right Column: Loot Spawners, Corpses, and Protected Chest */}
              <div className="lg:col-span-8 space-y-6">
                
                {/* Simulated Spawner Card */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl space-y-4">
                  <div>
                    <h3 className="font-bold text-sm text-slate-200">Gerador de Loot por Abate (Drop Generator)</h3>
                    <p className="text-xs text-slate-400">Teste o drop rate canônico de 50-60% e a diferenciação de criaturas e hierarquia demoníaca.</p>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    {[
                      { id: "common_demon", name: "Demônio Comum", type: "demonic", desc: "Cinza, Sangue, Essência inferior" },
                      { id: "minor_demon", name: "Demônio Menor (Elite)", type: "demonic", desc: "Núcleos inferiores, fragmentos" },
                      { id: "commander_demon", name: "Comandante Demoníaco", type: "demonic", desc: "Núcleo superior, estilhaço primordial" },
                      { id: "archegeneral_demon", name: "Arcegeneral Demoníaco", type: "demonic", desc: "Relíquias primordiais, artefatos abissais" },
                      { id: "civilized_cultist", name: "Cultista Civilizado", type: "currency", desc: "Portador frequente de moedas de Prata/Bronze" },
                      { id: "biological_beast", name: "Fera Selvagem", type: "beast", desc: "Raramente carrega moedas diretamente. Couro/Garras." }
                    ].map(monster => (
                      <button
                        key={monster.id}
                        onClick={() => handleSimulateLootDrop(monster.id)}
                        className="bg-slate-950/80 hover:bg-slate-900/80 border border-slate-800 hover:border-amber-500/40 p-4 rounded-xl text-left transition flex flex-col justify-between gap-2.5 text-xs group"
                      >
                        <div className="flex justify-between items-start w-full">
                          <span className="font-bold text-slate-200 group-hover:text-amber-300 transition">{monster.name}</span>
                          <span className={`text-[8px] uppercase px-1.5 py-0.5 rounded font-mono ${
                            monster.type === "demonic" ? "bg-purple-900/40 text-purple-300 border border-purple-800/50" :
                            monster.type === "currency" ? "bg-amber-900/40 text-amber-300 border border-amber-800/50" :
                            "bg-emerald-900/40 text-emerald-300 border border-emerald-800/50"
                          }`}>{monster.type}</span>
                        </div>
                        <p className="text-[10px] text-slate-400 leading-relaxed">{monster.desc}</p>
                        <div className="w-full flex items-center justify-end text-[10px] font-bold text-amber-500/80 group-hover:text-amber-400 gap-1 mt-1">
                          Abater Monstro
                          <ArrowRight className="w-3 h-3 transition group-hover:translate-x-1" />
                        </div>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  
                  {/* Monster Corpse (Inventory Overflow storage) */}
                  <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-xl space-y-4">
                    <div className="flex justify-between items-center border-b border-slate-800/60 pb-3">
                      <div className="flex items-center gap-2">
                        <Clock className="w-4 h-4 text-purple-400" />
                        <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Cadáver do Monstro</h4>
                      </div>
                      {monsterCorpseTimer > 0 && (
                        <span className="text-[10px] font-mono bg-purple-950 text-purple-300 px-2 py-0.5 rounded border border-purple-800">
                          Persiste por: {monsterCorpseTimer}s
                        </span>
                      )}
                    </div>

                    <p className="text-[10.5px] text-slate-400 leading-relaxed">
                      <strong>Lei de Transbordamento (Overflow):</strong> Se o seu inventário estiver totalmente cheio, o excedente permanece retido no corpo por até 12 minutos.
                    </p>

                    <div className="bg-slate-950 border border-slate-850 rounded-xl p-3 min-h-[96px] max-h-[160px] overflow-y-auto space-y-1.5">
                      {monsterCorpseLoot.length === 0 ? (
                        <div className="text-center py-6 text-[10.5px] text-slate-500 italic">Nenhum corpo com overflow pendente de saque no chão.</div>
                      ) : (
                        monsterCorpseLoot.map((item, idx) => (
                          <div key={idx} className="flex justify-between items-center bg-purple-950/20 border border-purple-900/30 px-2.5 py-1.5 rounded-lg text-xs">
                            <span className="font-medium text-purple-200">{item.name}</span>
                            <span className="font-mono font-bold text-purple-400">x{item.qty}</span>
                          </div>
                        ))
                      )}
                    </div>

                    {monsterCorpseLoot.length > 0 && (
                      <button
                        onClick={handleLootMonsterCorpse}
                        className="w-full py-2 bg-purple-700 hover:bg-purple-850 text-slate-950 font-bold rounded-xl text-xs transition"
                      >
                        Saquear Corpo (Manual Loot)
                      </button>
                    )}
                  </div>

                  {/* Player Corpse (Open Loot emergent gameplay) */}
                  <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-xl space-y-4">
                    <div className="flex items-center gap-2 border-b border-slate-800/60 pb-3">
                      <AlertTriangle className="w-4 h-4 text-rose-500 animate-pulse" />
                      <h4 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Seu Corpo no Chão (Player Corpse)</h4>
                    </div>

                    <p className="text-[10.5px] text-slate-400 leading-relaxed">
                      <strong>Open Loot Law:</strong> Seus pertences derrubados na morte PvE tornam-se saqueáveis por qualquer aventureiro no mapa. Corra para recuperar ou arrisque perder!
                    </p>

                    <div className="bg-slate-950 border border-slate-850 rounded-xl p-3 min-h-[96px] max-h-[160px] overflow-y-auto space-y-1.5">
                      {!playerCorpseExists ? (
                        <div className="text-center py-6 text-[10.5px] text-slate-500 italic">Você não possui corpos de morte ativos no mundo.</div>
                      ) : (
                        <div className="space-y-1.5">
                          {playerCorpseEquipped.map((item, idx) => (
                            <div key={`eq-${idx}`} className="flex justify-between items-center bg-rose-950/30 border border-rose-900/40 px-2.5 py-1 rounded-lg text-[10.5px]">
                              <span className="font-medium text-rose-300 font-mono">[{item.slot}] {item.name}</span>
                              <span className="text-[9px] bg-rose-900/50 text-rose-300 px-1.5 rounded">Equipamento</span>
                            </div>
                          ))}
                          {playerCorpseLoot.map((item, idx) => (
                            <div key={`inv-${idx}`} className="flex justify-between items-center bg-rose-950/30 border border-rose-900/40 px-2.5 py-1 rounded-lg text-[10.5px]">
                              <span className="font-medium text-rose-200">{item.name}</span>
                              <span className="font-mono font-bold text-rose-400">x{item.qty}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>

                    {playerCorpseExists && (
                      <div className="grid grid-cols-2 gap-2">
                        <button
                          onClick={handleRetrievePlayerCorpse}
                          className="py-2 bg-slate-850 hover:bg-slate-750 text-amber-300 border border-slate-700/60 rounded-xl text-xs font-bold transition"
                        >
                          Recuperar Corpo
                        </button>
                        <button
                          onClick={handleOpportunisticLootPlayerCorpse}
                          className="py-2 bg-rose-950/50 hover:bg-rose-900/50 text-rose-300 border border-rose-800/40 rounded-xl text-xs font-bold transition"
                        >
                          AI Saquear (Ladrão)
                        </button>
                      </div>
                    )}
                  </div>

                </div>

                {/* Raid Boss Protected Chest Law Card */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl space-y-5">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-bold text-xs text-slate-200 uppercase tracking-wider">Baú Protegido de Boss (Protected Boss Chest Law)</h3>
                      <p className="text-[10.5px] text-slate-400 leading-relaxed">
                        Recompensas pessoais de chefes de raides nunca caem na zona de combate hostil, eliminando roubos e luto. Elas são enviadas a um baú seguro e inviolável.
                      </p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-950/60 border border-slate-850 rounded-xl p-4">
                    <div className="space-y-1.5">
                      <label className="text-[9px] text-slate-500 font-bold uppercase block tracking-wider">Tempo de Combate Ativo:</label>
                      <input
                        type="number"
                        value={bossEligibleSecs}
                        onChange={e => setBossEligibleSecs(parseInt(e.target.value) || 0)}
                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-200 font-mono text-xs"
                      />
                      <span className="text-[8.5px] text-slate-500 font-medium block">Requisito: Mínimo 30s</span>
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-[9px] text-slate-500 font-bold uppercase block tracking-wider">Contribuição Ativa (%):</label>
                      <input
                        type="number"
                        step="0.1"
                        value={bossEligibleContr}
                        onChange={e => setBossEligibleContr(parseFloat(e.target.value) || 0)}
                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-200 font-mono text-xs"
                      />
                      <span className="text-[8.5px] text-slate-500 font-medium block">Requisito: Mínimo 1.0%</span>
                    </div>

                    <div className="flex items-end">
                      <button
                        onClick={handleSimulateBossProtectedChest}
                        className="w-full py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold rounded-lg text-xs transition"
                      >
                        Avaliar & Gerar Loot
                      </button>
                    </div>
                  </div>

                  {/* Safe Zone Chest Render */}
                  {safeZoneChestSpawned && (
                    <motion.div
                      initial={{ scale: 0.95, opacity: 0 }}
                      animate={{ scale: 1, opacity: 1 }}
                      className="border border-amber-500/20 bg-amber-500/5 rounded-xl p-4 flex flex-col md:flex-row justify-between items-center gap-4 animate-in fade-in"
                    >
                      <div className="flex items-center gap-3">
                        <Lock className="w-8 h-8 text-amber-400 animate-pulse" />
                        <div>
                          <h4 className="text-xs font-bold text-amber-300 uppercase tracking-wider">Seu Baú de Recompensa (Safe-Zone)</h4>
                          <p className="text-[10px] text-slate-400">Inviolável • Propriedade pessoal exclusiva de Gabriela_Paladin</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <p className="text-[11px] font-mono text-amber-200">
                            {safeZoneChestLoot.map(i => `${i.qty}x ${i.name}`).join(", ")}
                          </p>
                        </div>
                        <button
                          onClick={handleClaimSafeZoneChest}
                          className="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold rounded-lg text-xs transition uppercase tracking-wider"
                        >
                          Resgatar Tudo
                        </button>
                      </div>
                    </motion.div>
                  )}
                </div>

                {/* Console System Log */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-xl space-y-3">
                  <div className="flex justify-between items-center border-b border-slate-800/60 pb-2">
                    <span className="text-[10px] font-bold text-slate-300 uppercase tracking-wider font-sans">Console de Logs de Auditoria do Loot & Morte</span>
                    <button
                      onClick={() => setBibleLogs([])}
                      className="text-[9px] text-slate-500 hover:text-slate-400 font-bold uppercase"
                    >
                      Limpar Logs
                    </button>
                  </div>

                  <div className="bg-slate-950/80 rounded-xl p-3.5 border border-slate-850 h-[140px] overflow-y-auto font-mono text-[10.5px] space-y-1.5 scrollbar-thin scrollbar-thumb-slate-800">
                    {bibleLogs.length === 0 ? (
                      <div className="text-slate-500 italic text-center py-6">Pronto para simulação...</div>
                    ) : (
                      bibleLogs.map((log, idx) => (
                        <div
                          key={idx}
                          className={`leading-relaxed ${
                            log.type === "success" ? "text-emerald-400" :
                            log.type === "error" ? "text-rose-400" :
                            log.type === "warning" ? "text-amber-400" :
                            "text-slate-400"
                          }`}
                        >
                          <span className="text-[9px] text-slate-600 mr-2">[{new Date().toTimeString().split(" ")[0]}]</span>
                          {log.text}
                        </div>
                      ))
                    )}
                  </div>
                </div>

                {/* ================================================================== */}
                {/* TRADE & MARKET CANONICAL SIMULATOR */}
                {/* ================================================================== */}
                <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 space-y-6">
                  <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                    <div>
                      <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                        Sub-sistema de Economia & Escambo (v1)
                      </span>
                      <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                        <ArrowLeftRight className="w-5 h-5 text-amber-500 animate-pulse" />
                        Painel de Controle e Simulação de Comércio & Mercado Global
                      </h3>
                    </div>
                    <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                      <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Foco Canônico</span>
                      <span className="text-xs font-mono font-bold text-amber-400">Player-Driven Economy</span>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* Left Panel: Safe Banking & Ground Trade */}
                    <div className="space-y-6">
                      {/* Secure Banking Card */}
                      <div className="bg-slate-950/70 border border-slate-850 p-5 rounded-xl space-y-4">
                        <div className="flex items-center gap-2 text-xs font-bold text-slate-200 uppercase tracking-wider">
                          <Landmark className="w-4 h-4 text-amber-400" />
                          <span>Banco Seguro de Ravenshire (Safe Banking)</span>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div className="bg-slate-900/60 border border-slate-800/50 p-3 rounded-lg">
                            <span className="text-[9px] text-slate-500 uppercase block font-bold font-mono">Bolsa (Carregado)</span>
                            <span className="text-sm font-bold font-mono text-amber-400">{carriedCoins} Ouro</span>
                            <span className="text-[8px] text-rose-400 block mt-1">Suj. a perda na morte</span>
                          </div>
                          <div className="bg-slate-900/60 border border-slate-800/50 p-3 rounded-lg">
                            <span className="text-[9px] text-slate-500 uppercase block font-bold font-mono">Cofre do Banco</span>
                            <span className="text-sm font-bold font-mono text-emerald-400">{bankedCoins} Ouro</span>
                            <span className="text-[8px] text-emerald-400 block mt-1">100% Protegido</span>
                          </div>
                        </div>

                        {/* Deposit/Withdraw Actions */}
                        <div className="grid grid-cols-2 gap-2 pt-1">
                          <button
                            onClick={() => {
                              if (carriedCoins <= 0) {
                                addBibleLog("error", "[Banco] Você não possui moedas na bolsa para depositar.");
                                return;
                              }
                              const amount = carriedCoins;
                              setBankedCoins(prev => prev + amount);
                              setCarriedCoins(0);
                              addBibleLog("success", `[Banco] Depositado: ${amount} Moedas de Ouro transferidas com segurança da Bolsa para o Cofre.`);
                            }}
                            className="py-1.5 bg-slate-900 hover:bg-slate-800 text-[10.5px] text-emerald-400 font-bold border border-slate-800 rounded-lg transition cursor-pointer"
                          >
                            Depositar Tudo
                          </button>
                          <button
                            onClick={() => {
                              if (bankedCoins <= 0) {
                                addBibleLog("error", "[Banco] Você não possui moedas no cofre para sacar.");
                                return;
                              }
                              const amount = bankedCoins;
                              setCarriedCoins(prev => prev + amount);
                              setBankedCoins(0);
                              addBibleLog("warning", `[Banco] Sacado: ${amount} Moedas de Ouro retiradas do Cofre Seguro. ATENÇÃO: Seu ouro agora está exposto a perda na morte (carried-currency-risk-rule)!`);
                            }}
                            className="py-1.5 bg-slate-900 hover:bg-slate-800 text-[10.5px] text-amber-400 font-bold border border-slate-800 rounded-lg transition cursor-pointer"
                          >
                            Sacar Tudo
                          </button>
                        </div>
                      </div>

                      {/* Secure Trade UI (Two-Phase Window) */}
                      <div className="bg-slate-950/70 border border-slate-850 p-5 rounded-xl space-y-4">
                        <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                          <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                            <Scale className="w-4 h-4 text-sky-400" />
                            Escambo Seguro Direto (Direct Trade)
                          </span>
                          <span className={`text-[8px] uppercase px-1.5 py-0.5 rounded font-bold font-mono ${
                            secureTradeStatus === "idle" ? "bg-slate-800 text-slate-400" :
                            secureTradeStatus === "initiated" ? "bg-amber-900/40 text-amber-300" :
                            secureTradeStatus === "completed" ? "bg-emerald-950 text-emerald-400 border border-emerald-800" :
                            "bg-sky-950 text-sky-400 border border-sky-850"
                          }`}>{secureTradeStatus}</span>
                        </div>

                        {secureTradeStatus === "idle" ? (
                          <div className="text-center py-6">
                            <p className="text-xs text-slate-500 mb-3">Abra uma janela de comércio direto segura com outro jogador.</p>
                            <button
                              onClick={() => {
                                setSecureTradeStatus("initiated");
                                setSelfOffer("Lâmina Curta [T1]");
                                setSelfOfferCurrency(20);
                                addBibleLog("info", "[Trade] Solicitando comércio direto seguro com Eldrin_Mage (secure-trade-rule).");
                              }}
                              className="px-4 py-2 bg-sky-600 hover:bg-sky-700 text-slate-950 font-bold rounded-lg text-xs transition cursor-pointer"
                            >
                              Convidar Eldrin_Mage para Negócio
                            </button>
                          </div>
                        ) : (
                          <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-3 text-xs">
                              {/* Left: Self Offer */}
                              <div className="bg-slate-900 p-3 rounded-lg border border-slate-800 space-y-2">
                                <span className="text-[9px] font-bold text-amber-400 uppercase block">Sua Oferta (Gabriela)</span>
                                
                                <div className="space-y-1">
                                  <label className="text-[8px] text-slate-500 uppercase">Item Oferecido:</label>
                                  <input
                                    type="text"
                                    value={selfOffer}
                                    onChange={(e) => {
                                      setSelfOffer(e.target.value);
                                      if (secureTradeStatus !== "initiated") {
                                        setSecureTradeStatus("initiated");
                                        addBibleLog("warning", "[Trade] Modificação detectada! Confirmações anteriores pautadas foram resetadas de forma segura (secure-trade-rule).");
                                      }
                                    }}
                                    className="w-full bg-slate-950 border border-slate-850 rounded px-1.5 py-1 text-[11px] text-slate-200 focus:outline-none focus:border-amber-500/50"
                                  />
                                </div>

                                <div className="space-y-1">
                                  <label className="text-[8px] text-slate-500 uppercase">Moedas Oferecidas:</label>
                                  <input
                                    type="number"
                                    value={selfOfferCurrency}
                                    onChange={(e) => {
                                      const val = Math.min(carriedCoins, parseInt(e.target.value) || 0);
                                      setSelfOfferCurrency(val);
                                      if (secureTradeStatus !== "initiated") {
                                        setSecureTradeStatus("initiated");
                                        addBibleLog("warning", "[Trade] Modificação de moedas detectada! Confirmações resetadas com sucesso (secure-trade-rule).");
                                      }
                                    }}
                                    className="w-full bg-slate-950 border border-slate-850 rounded px-1.5 py-1 text-[11px] font-mono text-amber-400 focus:outline-none focus:border-amber-500/50"
                                  />
                                </div>
                              </div>

                              {/* Right: Other Offer */}
                              <div className="bg-slate-900 p-3 rounded-lg border border-slate-800 space-y-2">
                                <span className="text-[9px] font-bold text-purple-400 uppercase block">Oferta de Eldrin_Mage</span>
                                <div className="text-[11px] py-1.5 text-slate-300 border-b border-slate-850 font-medium">
                                  {otherOffer}
                                </div>
                                <div className="text-xs font-mono font-bold text-amber-400 py-1 flex items-center gap-1">
                                  <Coins className="w-3.5 h-3.5" />
                                  {otherOfferCurrency} Ouro
                                </div>
                              </div>
                            </div>

                            {/* Trade State Buttons */}
                            {secureTradeStatus !== "completed" ? (
                              <div className="space-y-2">
                                <div className="flex gap-2">
                                  <button
                                    onClick={() => {
                                      if (secureTradeStatus === "initiated") {
                                        setSecureTradeStatus("confirmed_self");
                                        addBibleLog("info", "[Trade] Você confirmou a proposta. Aguardando confirmação final de Eldrin_Mage.");
                                      } else if (secureTradeStatus === "confirmed_self") {
                                        // Simulate opponent confirming
                                        setSecureTradeStatus("confirmed_both");
                                        addBibleLog("success", "[Trade] Eldrin_Mage aceitou a proposta! Ambas as partes confirmaram o comércio direto.");
                                      } else if (secureTradeStatus === "confirmed_both") {
                                        // Execute Trade
                                        setCarriedCoins(prev => prev - selfOfferCurrency + otherOfferCurrency);
                                        setSecureTradeStatus("completed");
                                        addBibleLog("success", `[Trade] Negociação Concluída! Você deu [${selfOffer}] e ${selfOfferCurrency} Ouro, e recebeu [${otherOffer}] e ${otherOfferCurrency} Ouro (secure-trade-rule).`);
                                      }
                                    }}
                                    className="flex-1 py-2 bg-emerald-600 hover:bg-emerald-700 text-slate-950 font-bold rounded-lg text-xs transition cursor-pointer"
                                  >
                                    {secureTradeStatus === "initiated" ? "Confirmar Proposta" :
                                     secureTradeStatus === "confirmed_self" ? "Simular Aceite do Oponente" :
                                     "Concluir e Efetuar Escambo"}
                                  </button>
                                  <button
                                    onClick={() => {
                                      setSecureTradeStatus("idle");
                                      addBibleLog("error", "[Trade] Janela de Comércio Direto Cancelada. Nenhum item ou moeda foi trocado.");
                                    }}
                                    className="py-2 px-3 bg-rose-950/40 border border-rose-900/50 text-rose-400 font-bold rounded-lg text-xs hover:bg-rose-900/30 transition cursor-pointer"
                                  >
                                    Cancelar Comércio
                                  </button>
                                </div>
                                <span className="text-[9px] text-slate-500 font-mono block text-center">
                                  *Qualquer alteração de valores por sua parte cancela o aceite e reinicia o fluxo.
                                </span>
                              </div>
                            ) : (
                              <button
                                onClick={() => {
                                  setSecureTradeStatus("idle");
                                }}
                                className="w-full py-2 bg-slate-800 text-slate-300 font-bold rounded-lg text-xs hover:bg-slate-750 transition cursor-pointer"
                              >
                                Nova Negociação Direta
                              </button>
                            )}
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Right Panel: Global Connected Market */}
                    <div className="space-y-6">
                      {/* Market Listings */}
                      <div className="bg-slate-950/70 border border-slate-850 p-5 rounded-xl flex flex-col justify-between min-h-[445px]">
                        <div className="space-y-4">
                          <div className="flex justify-between items-center border-b border-slate-900 pb-2">
                            <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                              <Scale className="w-4 h-4 text-amber-500" />
                              Quadro de Anúncios Global (Global Market)
                            </span>
                            <span className="text-[8px] font-mono text-slate-500">Unificado entre cidades (global-market-rule)</span>
                          </div>

                          <div className="space-y-2">
                            <span className="text-[9px] text-slate-500 uppercase font-bold tracking-wider">Anúncios Ativos no Mundo:</span>
                            <div className="space-y-1.5 max-h-[160px] overflow-y-auto scrollbar-thin scrollbar-thumb-slate-800">
                              {activeMarketListings.length === 0 ? (
                                <div className="text-center py-6 text-xs text-slate-500 italic">O mercado global está vazio no momento.</div>
                              ) : (
                                activeMarketListings.map((listing) => (
                                  <div key={listing.id} className="flex justify-between items-center bg-slate-900/60 border border-slate-850 p-2.5 rounded-lg">
                                    <div className="flex flex-col">
                                      <span className="text-xs font-semibold text-slate-200">{listing.name}</span>
                                      <span className="text-[9px] text-slate-500 font-mono">Vendedor: {listing.seller}</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                      <span className="text-xs font-bold font-mono text-amber-400">{listing.price} Ouro</span>
                                      {listing.seller === "Gabriela_Paladin" ? (
                                        <button
                                          onClick={() => {
                                            setActiveMarketListings(prev => prev.filter(l => l.id !== listing.id));
                                            addBibleLog("warning", `[Market] Anúncio cancelado! O item [${listing.name}] foi removido e retornado ao seu depot (market-cancellation-rule). A taxa de anúncio não é reembolsada.`);
                                          }}
                                          className="text-[9px] bg-rose-950/50 hover:bg-rose-900/40 text-rose-300 font-bold px-2 py-1 rounded border border-rose-900/30 transition cursor-pointer"
                                        >
                                          Cancelar
                                        </button>
                                      ) : (
                                        <button
                                          onClick={() => {
                                            const totalCost = listing.price;
                                            if (carriedCoins + bankedCoins < totalCost) {
                                              addBibleLog("error", `[Market] Ouro Insuficiente! Você precisa de ${totalCost} Ouro para comprar [${listing.name}].`);
                                              return;
                                            }
                                            // Deduct from carried first, then bank
                                            if (carriedCoins >= totalCost) {
                                              setCarriedCoins(prev => prev - totalCost);
                                            } else {
                                              const remainder = totalCost - carriedCoins;
                                              setCarriedCoins(0);
                                              setBankedCoins(prev => prev - remainder);
                                            }
                                            setActiveMarketListings(prev => prev.filter(l => l.id !== listing.id));
                                            addBibleLog("success", `[Market] Compra efetuada com sucesso! Você adquiriu [${listing.name}] por ${listing.price} Ouro. Dinheiro deduzido com segurança. Ouro transferido diretamente para o vendedor (market-payment-delivery-rule).`);
                                          }}
                                          className="text-[9px] bg-emerald-950 hover:bg-emerald-900 text-emerald-400 font-bold px-2.5 py-1 rounded border border-emerald-900/40 transition cursor-pointer"
                                        >
                                          Comprar
                                        </button>
                                      )}
                                    </div>
                                  </div>
                                ))
                              )}
                            </div>
                          </div>
                        </div>

                        {/* Advertise / Place Listing Form */}
                        <div className="bg-slate-900/50 border border-slate-850 rounded-xl p-4 mt-4 space-y-3">
                          <span className="text-[10px] text-slate-300 font-bold uppercase tracking-wider block">Anunciar um Item no Mercado (market-deposit-rule)</span>
                          
                          <div className="grid grid-cols-2 gap-3 text-xs">
                            <div className="space-y-1">
                              <label className="text-[8px] text-slate-500 uppercase font-bold block">Nome do Item:</label>
                              <select
                                id="market_item_name_select"
                                className="w-full bg-slate-950 border border-slate-800 rounded px-2 py-1.5 text-slate-300 font-medium font-sans focus:outline-none"
                              >
                                <option value="weapon_t1">Espada Básica [T1]</option>
                                <option value="armor_leather">Armadura de Couro [T1]</option>
                                <option value="ring_crit">Anel Crítico [T1]</option>
                                <option value="shield_wooden">Escudo de Madeira [T1]</option>
                              </select>
                            </div>

                            <div className="space-y-1">
                              <label className="text-[8px] text-slate-500 uppercase font-bold block">Preço de Venda (Ouro):</label>
                              <input
                                id="market_item_price_input"
                                type="number"
                                defaultValue={80}
                                className="w-full bg-slate-950 border border-slate-800 rounded px-2 py-1 text-xs font-mono text-amber-400 focus:outline-none"
                              />
                            </div>
                          </div>

                          <button
                            onClick={() => {
                              const selectEl = document.getElementById("market_item_name_select") as HTMLSelectElement;
                              const priceEl = document.getElementById("market_item_price_input") as HTMLInputElement;
                              if (!selectEl || !priceEl) return;
                              
                              const itemText = selectEl.options[selectEl.selectedIndex].text;
                              const priceVal = parseInt(priceEl.value) || 0;

                              if (priceVal <= 0) {
                                addBibleLog("error", "[Market] Defina um preço de venda válido maior que 0.");
                                return;
                              }

                              const listingFee = Math.max(1, Math.round(priceVal * 0.02)); // 2% listing fee upfront gold sink
                              if (carriedCoins < listingFee) {
                                addBibleLog("error", `[Market] Ouro insuficiente na Bolsa para a taxa de anúncio! Necessário: ${listingFee} Ouro (2% do valor de anúncio).`);
                                return;
                              }

                              setCarriedCoins(prev => prev - listingFee);
                              setActiveMarketListings(prev => [
                                ...prev,
                                {
                                  id: `custom_${Date.now()}`,
                                  name: itemText,
                                  price: priceVal,
                                  seller: "Gabriela_Paladin"
                                }
                              ]);

                              addBibleLog("success", `[Market] Sucesso! Item [${itemText}] anunciado com preço fixo de ${priceVal} Ouro (market-pricing-rule). Item depositado na custódia do mercado (market-deposit-rule). Taxa de anúncio de ${listingFee} Ouro cobrada antecipadamente (market-listing-fee-rule).`);
                            }}
                            className="w-full py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 font-extrabold rounded-lg text-xs transition uppercase tracking-wider cursor-pointer font-sans"
                          >
                            Anunciar no Mercado (Taxa 2% de Ouro Upfront)
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Manual Ground Drop Trust Simulator */}
                  <div className="bg-slate-950/40 border border-slate-850 p-4 rounded-xl flex flex-col sm:flex-row justify-between items-center gap-4">
                    <div className="space-y-1">
                      <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                        <Ban className="w-3.5 h-3.5 text-rose-500" />
                        Descarte de Itens no Chão (Manual Trade / Betrayal Risk)
                      </h4>
                      <p className="text-[10.5px] text-slate-500 font-sans">
                        Simule o descarte manual sem segurança do sistema. Qualquer outro jogador pode pegar seu item, ideal para interações sandbox puras ou traições em áreas inseguras (manual-trade-rule).
                      </p>
                    </div>
                    <button
                      onClick={() => {
                        addBibleLog("warning", "[ManualTrade] Você dropou 15 Moedas de Ouro e [Essencia Inferior de Demonio] no chão em zona neutra (manual-trade-rule).");
                        addBibleLog("error", "[Emergência] Jogador oportunista 'Sneaky_Thief_44' correu, pegou suas moedas jogadas e fugiu de volta para o portal! Risco total de manual_trade materializado.");
                      }}
                      className="px-4 py-2 bg-rose-950/40 border border-rose-900/50 hover:bg-rose-900/30 text-rose-400 font-bold rounded-lg text-xs transition shrink-0 uppercase tracking-wider cursor-pointer"
                    >
                      Simular Drop Manual no Solo
                    </button>
                  </div>
                </div>

              </div>
            </motion.div>
          )}

          {/* Tab 2: Godot Autoload Configuration */}
          {activeTab === "config" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 lg:grid-cols-12 gap-6"
            >
              <div className="lg:col-span-4 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex flex-col justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-4 text-amber-400">
                    <Settings className="w-5 h-5" />
                    <h3 className="font-bold text-lg tracking-tight text-slate-100">
                      O que é um Autoload?
                    </h3>
                  </div>

                  <p className="text-xs text-slate-300 leading-relaxed mb-4">
                    Na engine Godot, um <strong>Autoload</strong> (anteriormente chamado de Singleton) é um nó de controle global instanciado automaticamente logo que a engine inicializa, antes de carregar qualquer cena local. Ele permanece persistente ao longo de todo o ciclo de vida do jogo.
                  </p>

                  <div className="bg-slate-950/80 p-4 border border-slate-800/80 rounded-xl space-y-3">
                    <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider">
                      Vantagens desta arquitetura:
                    </h4>
                    <ul className="text-xs text-slate-400 space-y-2 list-disc list-inside">
                      <li><strong className="text-slate-300">Persistência:</strong> GameManager não é destruído ao transicionar cenas visuais do mundo ou menu.</li>
                      <li><strong className="text-slate-300">Desacoplamento:</strong> Outros scripts acessam o barramento usando <code className="text-amber-400 font-mono">GameManager.Instance</code>.</li>
                      <li><strong className="text-slate-300">Ponto de Entrada Único:</strong> Centraliza a máquina de estados global do cliente.</li>
                    </ul>
                  </div>
                </div>

                <div className="bg-slate-800/30 border border-slate-700/30 rounded-xl p-4 mt-6">
                  <div className="flex items-start gap-2.5">
                    <Info className="w-4 h-4 text-violet-400 mt-0.5 shrink-0" />
                    <p className="text-[11px] text-slate-400 leading-normal">
                      Certifique-se de configurar a propriedade <code className="text-violet-300 font-mono">ProcessMode = ProcessModeEnum.Always</code> no GameManager para garantir que a rede e os logs não congelem caso o jogo seja pausado.
                    </p>
                  </div>
                </div>
              </div>

              {/* Autoload Setup Steps */}
              <div className="lg:col-span-8 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex flex-col gap-6">
                <div>
                  <h3 className="font-bold text-lg tracking-tight text-slate-100 mb-2">
                    Configuração Passo a Passo do Autoload C#
                  </h3>
                  <p className="text-xs text-slate-400">
                    Siga estas instruções precisas dentro do Editor da Godot 4 para habilitar a inicialização automática do core:
                  </p>
                </div>

                <div className="space-y-4">
                  {[
                    {
                      step: "01",
                      title: "Criar o arquivo C# do GameManager",
                      desc: "Certifique-se de que o arquivo GameManager.cs já existe na sua pasta de projeto no diretório res://src/Core/GameManager.cs. Ele deve herdar da classe Godot.Node."
                    },
                    {
                      step: "02",
                      title: "Abrir as Configurações do Projeto",
                      desc: "Abra a Godot Engine. No menu superior esquerdo, clique em Projeto (Project) -> Configurações do Projeto (Project Settings)."
                    },
                    {
                      step: "03",
                      title: "Navegar até a aba Autoload",
                      desc: "Dentro da janela de configurações, clique na aba 'Autoload' (localizada na parte superior, entre as abas 'Geral' e 'Plugins')."
                    },
                    {
                      step: "04",
                      title: "Vincular o Script",
                      desc: "No campo 'Caminho' (Path), clique no ícone de pasta e selecione res://src/Core/GameManager.cs. No campo 'Nome' (Node Name), digite exatamente GameManager. Clique em 'Adicionar' (Add)."
                    },
                    {
                      step: "05",
                      title: "Verificar Habilitação",
                      desc: "Certifique-se de que a caixa de verificação 'Habilitar' (Enable) está marcada na coluna correspondente ao GameManager adicionado. Salve e feche."
                    }
                  ].map((s, idx) => (
                    <div key={idx} className="flex gap-4 items-start bg-slate-950/40 p-4 border border-slate-800/60 rounded-xl hover:border-slate-700/50 transition">
                      <div className="text-lg font-bold font-mono text-amber-500 bg-amber-500/10 h-10 w-10 rounded-lg flex items-center justify-center shrink-0 border border-amber-500/20">
                        {s.step}
                      </div>
                      <div>
                        <h4 className="text-xs font-bold text-slate-200 mb-1">{s.title}</h4>
                        <p className="text-xs text-slate-400 leading-relaxed">{s.desc}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </motion.div>
          )}

          {/* Tab 3: Compilation & Execution Guide */}
          {activeTab === "exec" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 lg:grid-cols-12 gap-6"
            >
              <div className="lg:col-span-5 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex flex-col justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-4 text-emerald-400">
                    <Shield className="w-5 h-5" />
                    <h3 className="font-bold text-lg tracking-tight text-slate-100">
                      Pré-requisitos do Ambiente
                    </h3>
                  </div>

                  <p className="text-xs text-slate-300 leading-relaxed mb-4">
                    Para compilar e depurar os scripts C# do MMORPG Light and Shadow com sucesso, verifique se sua máquina possui o ambiente configurado abaixo:
                  </p>

                  <div className="space-y-3">
                    <div className="bg-slate-950/80 p-3.5 border border-slate-800 rounded-xl">
                      <h4 className="text-xs font-bold text-slate-200 mb-1 flex items-center gap-1">
                        <span className="w-1.5 h-1.5 bg-amber-500 rounded-full"></span>
                        Godot Engine 4+ (.NET Edition)
                      </h4>
                      <p className="text-[11px] text-slate-400">
                        Atenção: A versão padrão da Godot não possui suporte a C#. Você DEVE fazer o download especificamente da versão &quot;Godot Engine - .NET&quot;.
                      </p>
                    </div>

                    <div className="bg-slate-950/80 p-3.5 border border-slate-800 rounded-xl">
                      <h4 className="text-xs font-bold text-slate-200 mb-1 flex items-center gap-1">
                        <span className="w-1.5 h-1.5 bg-violet-500 rounded-full"></span>
                        .NET SDK 8.0
                      </h4>
                      <p className="text-[11px] text-slate-400">
                        Instale a versão estável LTS mais recente do .NET 8.0 SDK da Microsoft para habilitar a compilação do MSBuild de código C#.
                      </p>
                    </div>

                    <div className="bg-slate-950/80 p-3.5 border border-slate-800 rounded-xl">
                      <h4 className="text-xs font-bold text-slate-200 mb-1 flex items-center gap-1">
                        <span className="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>
                        IDE Recomendada
                      </h4>
                      <p className="text-[11px] text-slate-400">
                        Rider (JetBrains), VS Code com a extensão oficial C# Dev Kit ou Visual Studio 2022 para preenchimento de código inteligível.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-3 text-[10px] text-slate-400 font-mono leading-relaxed mt-6">
                  💡 A Godot compila automaticamente o código C# ao pressionar F5 (Play) no editor. Ela gera um assembly compartilhado gerenciado (.dll) na pasta oculta .godot/.
                </div>
              </div>

              {/* Command line execution */}
              <div className="lg:col-span-7 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex flex-col gap-6">
                <div>
                  <h3 className="font-bold text-lg tracking-tight text-slate-100 mb-2">
                    Comandos de Compilação & Execução Direta (CLI)
                  </h3>
                  <p className="text-xs text-slate-400">
                    Utilize os comandos a seguir para testar, compilar e rodar o projeto fora do editor:
                  </p>
                </div>

                <div className="space-y-4">
                  <div>
                    <h4 className="text-xs font-bold text-slate-300 mb-1.5 uppercase tracking-wider">
                      1. Restaurar dependências NuGet & Compilar Assemblies
                    </h4>
                    <div className="bg-slate-950/90 rounded-lg p-3.5 border border-slate-800 relative group">
                      <code className="text-xs text-emerald-400 font-mono block">
                        dotnet restore<br />
                        dotnet build LightAndShadow.csproj -c Debug
                      </code>
                    </div>
                    <p className="text-[11px] text-slate-500 mt-1 leading-normal">
                      Restaura as bibliotecas base compiladas do Godot (.NET Sdk) e gera os arquivos binários temporários do jogo.
                    </p>
                  </div>

                  <div>
                    <h4 className="text-xs font-bold text-slate-300 mb-1.5 uppercase tracking-wider">
                      2. Executar o Bootstrap diretamente via Terminal
                    </h4>
                    <div className="bg-slate-950/90 rounded-lg p-3.5 border border-slate-800">
                      <code className="text-xs text-amber-300 font-mono block">
                        # Se estiver no diretório root do seu Godot (.NET)<br />
                        godot --path . --scene res://scenes/bootstrap.tscn
                      </code>
                    </div>
                    <p className="text-[11px] text-slate-500 mt-1 leading-normal">
                      Força o motor Godot a rodar iniciando explicitamente na cena do Bootstrap para testar o fluxo de estados do cliente.
                    </p>
                  </div>

                  <div>
                    <h4 className="text-xs font-bold text-slate-300 mb-1.5 uppercase tracking-wider">
                      3. Executar Testes Locais de Soquete do NetworkManager
                    </h4>
                    <p className="text-xs text-slate-300 leading-relaxed mb-2">
                      Para simular e debugar o comportamento do <code className="text-amber-400 font-mono">NetworkManager.cs</code>, você pode subir um servidor TCP eco simples no terminal para ver o loop de escrita/leitura e o Heartbeat do cliente funcionando em loop na porta <code className="text-violet-400 font-mono">8080</code>:
                    </p>
                    <div className="bg-slate-950/90 rounded-lg p-3.5 border border-slate-800">
                      <code className="text-xs text-slate-400 font-mono block">
                        # Comando terminal para rodar um servidor de teste TCP básico escutando na porta 8080<br />
                        ncat -lk 127.0.0.1 8080 -v
                      </code>
                    </div>
                  </div>
                </div>
              </div>
            </motion.div>
          )}

          {/* Tab: World Activity Matrix Sandbox */}
          {activeTab === "activity" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="space-y-6"
            >
              {/* Header section with description */}
              <div className="bg-gradient-to-r from-slate-900 via-slate-950 to-slate-900 border border-slate-800/80 rounded-2xl p-6 shadow-xl relative overflow-hidden">
                <div className="absolute right-0 top-0 h-full w-1/3 bg-radial-gradient(circle at right, rgba(56,189,248,0.03) 0%, transparent 70%) pointer-events-none" />
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                  <div className="space-y-1">
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      DIRETRIZ CANÔNICA SEXTA-FEIRA • SPRINT 6
                    </span>
                    <h3 className="font-extrabold text-xl tracking-tight text-slate-100 flex items-center gap-2">
                      <Compass className="w-5 h-5 text-sky-400 animate-spin-slow" />
                      Matriz de Atividades do Mundo (Sandbox Model)
                    </h3>
                    <p className="text-xs text-slate-400 leading-relaxed max-w-3xl">
                      Estrutura de sandbox de nós de atividade independentes, governados por riscos, recompensas e configurações de encontros. Sem progressão linear obrigatória de mapas ou travas artificiais de nível de personagem.
                    </p>
                  </div>
                  <div className="bg-slate-900/80 border border-slate-800/80 p-2.5 rounded-xl text-center shrink-0">
                    <span className="text-[9px] font-bold font-mono text-slate-500 uppercase block">Modelo Econômico</span>
                    <span className="text-xs font-mono font-bold text-amber-400">B+ Scarcity Law</span>
                  </div>
                </div>
              </div>

              {/* Canonical Principles Cards with ID anchors */}
              <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
                {[
                  {
                    id: "activity-freedom-principle",
                    title: "1. Liberdade de Atividades",
                    rule: "Players may engage in any available activity at any time, with no enforced progression sequence.",
                    color: "border-sky-500/20 text-sky-400"
                  },
                  {
                    id: "risk-reward-binding",
                    title: "2. Vinculação Risco-Retorno",
                    rule: "Higher risk activities must proportionally increase reward potential through loot quality, currency yield, or crafting material density.",
                    color: "border-amber-500/20 text-amber-400"
                  },
                  {
                    id: "death-activity-link",
                    title: "3. Consequência de Morte",
                    rule: "Activities must explicitly define how death alters reward loss, loot exposure, and economic consequence.",
                    color: "border-rose-500/20 text-rose-400"
                  },
                  {
                    id: "blessing-risk-modifier",
                    title: "4. Modificador de Bênção",
                    rule: "Blessings function as a global risk mitigation modifier applied across all activity types.",
                    color: "border-violet-500/20 text-violet-400"
                  },
                  {
                    id: "sandbox-activity-model",
                    title: "5. Modelo de Sandbox",
                    rule: "The world is a sandbox of independent activity nodes governed by risk, reward, and encounter configuration rather than player level pathing.",
                    color: "border-emerald-500/20 text-emerald-400"
                  }
                ].map((principle) => (
                  <div 
                    key={principle.id} 
                    id={principle.id}
                    className="bg-slate-950/80 border border-slate-850 p-3.5 rounded-xl flex flex-col justify-between space-y-2 hover:border-slate-800 transition-all duration-150"
                  >
                    <span className={`text-[10px] font-bold font-mono uppercase ${principle.color.split(" ")[1]}`}>
                      {principle.title}
                    </span>
                    <p className="text-[9.5px] text-slate-300 leading-relaxed font-mono">
                      "{principle.rule}"
                    </p>
                    <span className="text-[8px] text-slate-600 font-mono">ID: {principle.id}</span>
                  </div>
                ))}
              </div>

              {/* Main Interactive Matrix Panel */}
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Left Side: Activity Types Selector */}
                <div className="lg:col-span-5 space-y-4">
                  <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider">Categorias de Atividade Canônicas</h4>
                  
                  <div className="space-y-3">
                    {[
                      {
                        key: "HUNT",
                        name: "Hunt Activities",
                        purpose: "Aquisição de recursos de combate, moldes e gemas.",
                        icon: Sword,
                        risk: "Alta",
                        color: "from-rose-500/10 to-transparent border-rose-500/30 text-rose-400",
                        activeColor: "ring-2 ring-rose-500/50 border-rose-400"
                      },
                      {
                        key: "EXPLORATION",
                        name: "Exploration Activities",
                        purpose: "Descoberta, coleta de recursos do ambiente e codices de lore.",
                        icon: Compass,
                        risk: "Baixa",
                        color: "from-sky-500/10 to-transparent border-sky-500/30 text-sky-400",
                        activeColor: "ring-2 ring-sky-500/50 border-sky-400"
                      },
                      {
                        key: "QUEST",
                        name: "Quest / Contract Activities",
                        purpose: "Progressão de objetivos estruturados e bônus de facção.",
                        icon: FileText,
                        risk: "Variável",
                        color: "from-amber-500/10 to-transparent border-amber-500/30 text-amber-400",
                        activeColor: "ring-2 ring-amber-500/50 border-amber-400"
                      },
                      {
                        key: "BOSS",
                        name: "Boss / World Event Activities",
                        purpose: "Encontros de alto risco, mecânicas de fases e baú protegido.",
                        icon: Sparkles,
                        risk: "Extrema",
                        color: "from-violet-500/10 to-transparent border-violet-500/30 text-violet-400",
                        activeColor: "ring-2 ring-violet-500/50 border-violet-400"
                      }
                    ].map((act) => {
                      const Icon = act.icon;
                      const isSelected = selectedActivity === act.key;
                      return (
                        <button
                          key={act.key}
                          onClick={() => setSelectedActivity(act.key as any)}
                          className={`w-full text-left bg-gradient-to-r p-4 rounded-xl border transition-all duration-150 flex items-center justify-between group ${act.color} ${
                            isSelected ? act.activeColor : "hover:border-slate-700/80 hover:translate-x-1"
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            <div className="p-2 bg-slate-950/60 rounded-lg group-hover:scale-105 transition-transform">
                              <Icon className="w-5 h-5" />
                            </div>
                            <div>
                              <span className="text-[12px] font-bold text-slate-100 block">{act.name}</span>
                              <span className="text-[10px] text-slate-400 leading-tight block">{act.purpose}</span>
                            </div>
                          </div>
                          <div className="text-right">
                            <span className="text-[9px] font-bold uppercase tracking-wider text-slate-500 block">Risco</span>
                            <span className={`text-xs font-bold ${act.risk === "Baixa" ? "text-emerald-400" : act.risk === "Variável" ? "text-amber-400" : "text-rose-400"}`}>{act.risk}</span>
                          </div>
                        </button>
                      );
                    })}
                  </div>

                  {/* Blessing Slider Widget (Syncs with the global blessing tracker!) */}
                  <div className="bg-slate-950/50 border border-slate-850 p-4 rounded-xl space-y-3">
                    <div className="flex justify-between items-center">
                      <div className="flex items-center gap-1.5">
                        <Shield className="w-4 h-4 text-amber-400" />
                        <div>
                          <span className="text-xs font-bold text-slate-200 block">Sincronizador de Bênçãos</span>
                          <span className="text-[9px] text-slate-500">Mudar aqui altera as bênçãos ativas do herói</span>
                        </div>
                      </div>
                      <span className="text-xs font-mono font-bold text-amber-400">{activeBlessingsCount}/7</span>
                    </div>

                    <div className="flex items-center justify-between gap-1 bg-slate-900/60 p-1 rounded-md border border-slate-800/40">
                      {[0, 1, 2, 3, 4, 5, 6, 7].map((num) => (
                        <button
                          key={num}
                          onClick={() => setActiveBlessingsCount(num)}
                          className={`flex-1 h-7 flex items-center justify-center text-[11px] font-mono font-bold rounded transition-all duration-150 ${
                            activeBlessingsCount === num
                              ? "bg-amber-500 text-slate-950 shadow-md shadow-amber-500/20 scale-105"
                              : "text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                          }`}
                        >
                          {num}
                        </button>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Right Side: Detailed Simulator */}
                <div className="lg:col-span-7 bg-slate-900/40 border border-slate-850 rounded-2xl p-5 space-y-5">
                  <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
                    <span className="text-xs font-extrabold text-slate-300 uppercase tracking-widest">
                      Simulação do Nó de Atividade: {selectedActivity}
                    </span>
                    <span className="text-[10px] font-mono text-slate-500">
                      Fórmula Linear Aplicada
                    </span>
                  </div>

                  {/* Selected Activity Properties Card */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                    <div className="bg-slate-950/70 p-3 rounded-xl border border-slate-850">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold">Risco de Morte Base</span>
                      <span className="text-sm font-bold text-slate-200">
                        {selectedActivity === "HUNT" ? "High" : selectedActivity === "EXPLORATION" ? "Minimal" : selectedActivity === "QUEST" ? "Moderate" : "Extreme"}
                      </span>
                    </div>
                    <div className="bg-slate-950/70 p-3 rounded-xl border border-slate-850">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold">Loot Exposure Base</span>
                      <span className="text-sm font-bold text-slate-200">
                        {selectedActivity === "HUNT" ? "High" : selectedActivity === "EXPLORATION" ? "Minimal" : selectedActivity === "QUEST" ? "Moderate" : "Minimal"}
                      </span>
                    </div>
                    <div className="bg-slate-950/70 p-3 rounded-xl border border-slate-850">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold">Densidade Spawn</span>
                      <span className="text-sm font-bold text-slate-200">
                        {selectedActivity === "HUNT" ? "Baixa a Média" : selectedActivity === "EXPLORATION" ? "Extremamente Baixa" : selectedActivity === "QUEST" ? "Variável" : "Raid / Único"}
                      </span>
                    </div>
                    <div className="bg-slate-950/70 p-3 rounded-xl border border-slate-850">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold">Expectativa TTK</span>
                      <span className="text-sm font-bold text-slate-200">
                        {selectedActivity === "HUNT" ? "Alto (Lento)" : selectedActivity === "EXPLORATION" ? "Nulo" : selectedActivity === "QUEST" ? "Médio" : "Extremo (Fases)"}
                      </span>
                    </div>
                  </div>

                  {/* Calculated Exposure Metrics with Blessings applied */}
                  <div className="bg-slate-950/60 border border-slate-850 p-4 rounded-xl space-y-4">
                    <h5 className="text-[10px] font-bold text-slate-300 uppercase tracking-wider">Mitigação de Perda Atual</h5>
                    
                    {/* Linear Protection Formula Visualizer */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Death Risk Mitigation */}
                      <div className="space-y-1.5">
                        <div className="flex justify-between items-center text-[10px] font-mono">
                          <span className="text-slate-400">Risco de Morte Efetivo:</span>
                          <span className="font-bold text-rose-400">
                            {(() => {
                              if (activeBlessingsCount === 7) return "Completamente Mitigado (Seguro)";
                              if (activeBlessingsCount === 0) {
                                return selectedActivity === "HUNT" ? "High" : selectedActivity === "EXPLORATION" ? "Minimal" : selectedActivity === "QUEST" ? "Moderate" : "Extreme";
                              }
                              const baseLvl = selectedActivity === "HUNT" ? 4 : selectedActivity === "EXPLORATION" ? 1 : selectedActivity === "QUEST" ? 3 : 5;
                              const mitigation = Math.floor((activeBlessingsCount / 7) * baseLvl);
                              const finalLvl = Math.max(1, baseLvl - mitigation);
                              const tiers = ["", "Minimal", "Low", "Moderate", "High", "Extreme"];
                              return `Parcialmente Mitigado (${tiers[finalLvl]})`;
                            })()}
                          </span>
                        </div>
                        <div className="h-2 w-full bg-slate-900 rounded-full overflow-hidden border border-slate-800">
                          <div 
                            className="h-full bg-rose-500 transition-all duration-300"
                            style={{ 
                              width: `${(1 - activeBlessingsCount / 7) * 100}%` 
                            }}
                          />
                        </div>
                      </div>

                      {/* Loot Exposure Mitigation */}
                      <div className="space-y-1.5">
                        <div className="flex justify-between items-center text-[10px] font-mono">
                          <span className="text-slate-400">Derrube de Itens Efetivo:</span>
                          <span className="font-bold text-amber-400">
                            {(() => {
                              if (activeBlessingsCount === 7) return "Totalmente Protegido";
                              if (selectedActivity === "BOSS") return "Nulo (Protected Boss Chest)";
                              if (activeBlessingsCount === 0) {
                                return selectedActivity === "HUNT" ? "High" : selectedActivity === "EXPLORATION" ? "Minimal" : selectedActivity === "QUEST" ? "Moderate" : "Extreme";
                              }
                              const baseLvl = selectedActivity === "HUNT" ? 4 : selectedActivity === "EXPLORATION" ? 1 : selectedActivity === "QUEST" ? 3 : 5;
                              const mitigation = Math.floor((activeBlessingsCount / 7) * baseLvl);
                              const finalLvl = Math.max(1, baseLvl - mitigation);
                              const tiers = ["", "Minimal", "Low", "Moderate", "High", "Extreme"];
                              return `Parcialmente Mitigado (${tiers[finalLvl]})`;
                            })()}
                          </span>
                        </div>
                        <div className="h-2 w-full bg-slate-900 rounded-full overflow-hidden border border-slate-800">
                          <div 
                            className="h-full bg-amber-500 transition-all duration-300"
                            style={{ 
                              width: `${selectedActivity === "BOSS" ? 0 : (1 - activeBlessingsCount / 7) * 100}%` 
                            }}
                          />
                        </div>
                      </div>
                    </div>

                    {/* Protection Status Banner */}
                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/40 text-[10px] text-slate-400 leading-relaxed font-mono flex items-start gap-2.5">
                      {activeBlessingsCount === 7 ? (
                        <>
                          <Shield className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                          <div>
                            <span className="text-emerald-400 font-bold block">PROTEÇÃO TOTAL ATIVA:</span>
                            Nenhum item do seu equipamento ou inventário será perdido na morte PvE. No entanto, lembre-se da <span className="text-amber-400 font-bold">Blessing Consumption Law</span>: uma única morte consumirá todas as suas 7 bênçãos simultaneamente.
                          </div>
                        </>
                      ) : activeBlessingsCount === 0 ? (
                        <>
                          <AlertTriangle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                          <div>
                            <span className="text-rose-400 font-bold block">SEM PROTEÇÃO DE BÊNÇÃOS:</span>
                            Risco total de queda de itens! Ao morrer, seus pertences ficarão em um corpo público com <span className="text-rose-400 font-bold">Open Loot</span> ativado na área de combate. Qualquer jogador poderá saqueá-lo!
                          </div>
                        </>
                      ) : (
                        <>
                          <Info className="w-4 h-4 text-sky-400 shrink-0 mt-0.5" />
                          <div>
                            <span className="text-sky-400 font-bold block">PROTEÇÃO PARCIAL (MILITARIZADA):</span>
                            Você possui <span className="text-amber-400 font-bold">{activeBlessingsCount} bênçãos</span>, garantindo proteção parcial de proporção linear. O risco residual ainda expõe seu inventário a quedas de itens. Morte purgará todas as bênçãos ativas!
                          </div>
                        </>
                      )}
                    </div>
                  </div>

                  {/* Interactive Loot Roll Simulation OR Quest & Contract Board */}
                  {selectedActivity === "QUEST" ? (
                    <div className="bg-slate-950/60 border border-slate-850 p-5 rounded-xl space-y-4">
                      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center border-b border-slate-800/60 pb-3 gap-2">
                        <div>
                          <h5 className="text-[11px] font-extrabold text-slate-100 uppercase tracking-wider flex items-center gap-1.5">
                            <BookOpen className="w-4 h-4 text-amber-400" />
                            Quadro de Missões e Contratos (Mural de Facção)
                          </h5>
                          <span className="text-[9px] text-slate-500 font-mono">Total Sandbox Choice & Zero Mandatory Linear Paths</span>
                        </div>
                        <div className="flex items-center gap-1 bg-slate-900 p-1 rounded-lg border border-slate-800">
                          <button
                            onClick={() => {
                              setQuestTab("QUESTS");
                              setActiveQuestId("main_1");
                            }}
                            className={`px-2.5 py-1 text-[9.5px] font-bold rounded transition-all duration-150 ${
                              questTab === "QUESTS"
                                ? "bg-amber-500 text-slate-950 font-extrabold"
                                : "text-slate-400 hover:text-slate-200"
                            }`}
                          >
                            Campanha (Quests)
                          </button>
                          <button
                            onClick={() => {
                              setQuestTab("CONTRACTS");
                              setActiveQuestId("contract_hunt");
                            }}
                            className={`px-2.5 py-1 text-[9.5px] font-bold rounded transition-all duration-150 ${
                              questTab === "CONTRACTS"
                                ? "bg-amber-500 text-slate-950 font-extrabold"
                                : "text-slate-400 hover:text-slate-200"
                            }`}
                          >
                            Repetíveis (Contratos)
                          </button>
                        </div>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-12 gap-4">
                        {/* List Selector Column */}
                        <div className="md:col-span-5 space-y-2 max-h-[220px] overflow-y-auto pr-1">
                          {questTab === "QUESTS" ? (
                            [
                              { id: "main_1", title: "Caminho do Herói", cat: "MAIN", color: "border-sky-500/30 text-sky-400" },
                              { id: "story_1", title: "Origens Primordiais", cat: "STORY", color: "border-purple-500/30 text-purple-400" },
                              { id: "unlock_1", title: "Rito de Ignis", cat: "UNLOCK", color: "border-amber-500/30 text-amber-400" },
                              { id: "side_1", title: "Problema no Moinho", cat: "SIDE", color: "border-emerald-500/30 text-emerald-400" }
                            ].map(q => (
                              <button
                                key={q.id}
                                onClick={() => setActiveQuestId(q.id)}
                                className={`w-full text-left p-2 rounded-lg border text-[11px] transition-all duration-150 block ${
                                  activeQuestId === q.id
                                    ? "bg-slate-800 border-amber-500/60 ring-1 ring-amber-500/30"
                                    : "bg-slate-900/40 border-slate-800/60 hover:bg-slate-800/40"
                                }`}
                              >
                                <div className="flex justify-between items-center">
                                  <span className="font-bold text-slate-200 truncate">{q.title}</span>
                                  <span className={`text-[8px] font-mono font-extrabold px-1 rounded bg-slate-950 ${q.color}`}>{q.cat}</span>
                                </div>
                              </button>
                            ))
                          ) : (
                            [
                              { id: "contract_hunt", title: "Caça Lobos", type: "HUNT", color: "text-rose-400" },
                              { id: "contract_bounty", title: "Bounty: Gorila", type: "BOUNTY", color: "text-amber-400" },
                              { id: "contract_escort", title: "Escolta da Caravana", type: "ESCORT", color: "text-violet-400" },
                              { id: "contract_delivery", title: "Entrega de Tratados", type: "DELIVERY", color: "text-sky-400" }
                            ].map(c => (
                              <button
                                key={c.id}
                                onClick={() => setActiveQuestId(c.id)}
                                className={`w-full text-left p-2 rounded-lg border text-[11px] transition-all duration-150 block ${
                                  activeQuestId === c.id
                                    ? "bg-slate-800 border-amber-500/60 ring-1 ring-amber-500/30"
                                    : "bg-slate-900/40 border-slate-800/60 hover:bg-slate-800/40"
                                }`}
                              >
                                <div className="flex justify-between items-center">
                                  <span className="font-bold text-slate-200 truncate">{c.title}</span>
                                  <span className={`text-[8px] font-mono font-extrabold px-1 rounded bg-slate-950 ${c.color}`}>{c.type}</span>
                                </div>
                              </button>
                            ))
                          )}
                        </div>

                        {/* Details View Column */}
                        <div className="md:col-span-7 bg-slate-950/40 border border-slate-850/60 rounded-xl p-3.5 space-y-3 flex flex-col justify-between min-h-[220px]">
                          {(() => {
                            const selectedData = questTab === "QUESTS" 
                              ? [
                                  {
                                    id: "main_1",
                                    title: "Caminho do Herói (Main Quest)",
                                    faction: "Mage Orders",
                                    desc: "Essencial para a macro-progressão geral do mundo e desbloqueio de novos continentes. No entanto, o sandbox permite que você a ignore livremente.",
                                    rewardsText: "+50 XP Secundário, 1x Molde de Espada T1, 50 Moedas de Prata",
                                    rewardItem: { id: "weapon_t1", name: "Weapon Template T1", qty: 1, value: 50, type: "weapon" },
                                    laws: ["Macro-progresso", "Sem falha permanente", "NPCs Imortais"]
                                  },
                                  {
                                    id: "story_1",
                                    title: "Origens Primordiais (Story Quest)",
                                    faction: "Religious Orders",
                                    desc: "Expande a profundidade narrativa sobre o conflito ancestral dos Primordiais de Luz e Sombra. 100% opcional, focada puramente em imersão.",
                                    rewardsText: "+30 XP Secundário, 1x Essência Primordial, 250 Moedas de Bronze",
                                    rewardItem: { id: "soft_leather", name: "Couro Macio", qty: 2, value: 10, type: "material" },
                                    laws: ["100% Opcional", "Lore dos Primordiais", "Imersão de Fundo"]
                                  },
                                  {
                                    id: "unlock_1",
                                    title: "Rito Elemental de Ignis (Unlock Quest)",
                                    faction: "Mage Orders",
                                    desc: "Complete os mistérios do fogo para ganhar acesso a receitas de forja raras de Tier 2 sem amarras lineares.",
                                    rewardsText: "+40 XP Secundário, 1x Molde de Armadura T2, 100 Moedas de Bronze",
                                    rewardItem: { id: "armor_t2", name: "Armor Template T2", qty: 1, value: 120, type: "armor" },
                                    laws: ["Desbloqueio de Receitas", "Gates de Alinhamento", "Sandbox Choice"]
                                  },
                                  {
                                    id: "side_1",
                                    title: "Problema no Moinho (Side Quest)",
                                    faction: "Mercenary Guilds",
                                    desc: "Ajude o fazendeiro a consertar os moinhos locais danificados por hordas de monstros. Simples e abundante na região.",
                                    rewardsText: "+20 XP Secundário, 2x Poções de Cura",
                                    rewardItem: { id: "potion_heal", name: "Poção de Cura", qty: 2, value: 15, type: "consumable" },
                                    laws: ["Abundante regionalmente", "Apoio a NPCs locais", "Sem dependência linear"]
                                  }
                                ].find(x => x.id === activeQuestId)
                              : [
                                  {
                                    id: "contract_hunt",
                                    title: "Contrato: Caça Lobos (Hunt Contract)",
                                    faction: "Hunter Lodges",
                                    desc: "Abata os lobos cinzentos que ameaçam as ovelhas. Loops repetíveis infinitamente para sustentar a economia regional.",
                                    rewardsText: "+15 XP de Caçada, 3x Couro Macio",
                                    rewardItem: { id: "soft_leather", name: "Couro Macio", qty: 3, value: 10, type: "material" },
                                    laws: ["Infinitamente repetível", "Sustentação do Loop", "Foco em Atividade de Mundo"]
                                  },
                                  {
                                    id: "contract_bounty",
                                    title: "Bounty: O Gorila de Presas Rubras (Bounty Contract)",
                                    faction: "Hunter Lodges",
                                    desc: "Rastreie e derrote o alfa colossal. Perigo moderado em área aberta de conflito com spawn dinâmico.",
                                    rewardsText: "+30 XP de Caçada, 1x Combat Material T1",
                                    rewardItem: { id: "combat_material_t1", name: "Combat Material T1", qty: 1, value: 30, type: "material" },
                                    laws: ["Elite Alvo Único", "Spawn dinâmico no mapa", "B+ Scarcity"]
                                  },
                                  {
                                    id: "contract_escort",
                                    title: "Escolta da Caravana de Prata (Escort Contract)",
                                    faction: "Mercenary Guilds",
                                    desc: "Proteja a caravana comercial que viaja por zonas de PvP aberto. Alto risco de intervenção de outros jogadores!",
                                    rewardsText: "+40 XP de Guilda, 2x Minério de Ferro",
                                    rewardItem: { id: "iron_ore", name: "Minério de Ferro", qty: 2, value: 5, type: "material" },
                                    laws: ["Risco em PvP Aberto", "Escolta de Cargas", "Emergent Sandbox Conflict"]
                                  },
                                  {
                                    id: "contract_delivery",
                                    title: "Entrega de Tratados Diplomáticos (Delivery Contract)",
                                    faction: "Mage Orders",
                                    desc: "Entregue a correspondência diplomática de alta importância para o assentamento de pesquisa sem iniciar combates desnecessários.",
                                    rewardsText: "+10 XP de Guilda, 100 Moedas de Bronze",
                                    rewardItem: { id: "gold_coin", name: "Moeda de Ouro", qty: 1, value: 10000, type: "currency" },
                                    laws: ["Transporte pacífico", "Foco em exploração", "Sustentabilidade de Rotas"]
                                  }
                                ].find(x => x.id === activeQuestId);

                            if (!selectedData) return <div className="text-[10px] text-slate-500 font-mono">Selecione uma atividade para simular.</div>;

                            return (
                              <>
                                <div className="space-y-1.5">
                                  <div className="flex justify-between items-start">
                                    <h6 className="text-[11.5px] font-bold text-slate-200">{selectedData.title}</h6>
                                    <span className="text-[8.5px] font-mono text-amber-400 bg-slate-900 border border-slate-800 px-1.5 py-0.5 rounded">
                                      {selectedData.faction}
                                    </span>
                                  </div>
                                  <p className="text-[10px] text-slate-400 leading-relaxed font-mono">{selectedData.desc}</p>
                                  <div className="text-[10px] font-mono bg-slate-900/60 p-2 rounded border border-slate-850/40">
                                    <span className="text-amber-400 font-bold block mb-0.5">Recompensas Determinísticas:</span>
                                    <span className="text-slate-300">{selectedData.rewardsText}</span>
                                  </div>
                                </div>

                                <div className="space-y-2 pt-2 border-t border-slate-850">
                                  <div className="flex flex-wrap gap-1">
                                    {selectedData.laws.map((law, index) => (
                                      <span key={index} className="text-[8px] font-mono bg-slate-900 text-slate-400 px-1 rounded border border-slate-800">
                                        ▪ {law}
                                      </span>
                                    ))}
                                  </div>
                                  <button
                                    onClick={() => {
                                      // Simulate rewards integration
                                      const rewardItem = selectedData.rewardItem;
                                      
                                      // Push to player's simulated inventory
                                      const existingIdx = playerInventory.findIndex(item => item.id === rewardItem.id);
                                      if (existingIdx > -1) {
                                        const updatedInv = [...playerInventory];
                                        updatedInv[existingIdx] = {
                                          ...updatedInv[existingIdx],
                                          qty: updatedInv[existingIdx].qty + rewardItem.qty
                                        };
                                        setPlayerInventory(updatedInv);
                                      } else {
                                        setPlayerInventory(prev => [...prev, { ...rewardItem }]);
                                      }

                                      addBibleLog("success", `[QuestEngine] Concluído: ${selectedData.title}!`);
                                      addBibleLog("info", `[QuestEngine] Recompensas coletadas: ${selectedData.rewardsText}.`);
                                      addBibleLog("info", `[Canônico] Hunts continuam sendo a fonte primária de XP; Quests proveem recompensas secundárias de progressão.`);
                                    }}
                                    className="w-full py-1.5 bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold rounded-lg text-[10px] uppercase tracking-wider transition-colors duration-150 flex items-center justify-center gap-1"
                                  >
                                    <CheckCircle className="w-3.5 h-3.5" />
                                    Aceitar e Simular Conclusão
                                  </button>
                                </div>
                              </>
                            );
                          })()}
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="bg-slate-950/40 border border-slate-850 p-4 rounded-xl space-y-3.5">
                      <div className="flex justify-between items-center">
                        <h5 className="text-[10px] font-bold text-slate-300 uppercase tracking-wider flex items-center gap-1">
                          <Coins className="w-3.5 h-3.5 text-amber-400" />
                          Simulador de Saque (B+ Economics)
                        </h5>
                        <span className="text-[9px] text-slate-500 font-mono">60% Crafting Dominance</span>
                      </div>

                      <p className="text-[10px] text-slate-400 leading-relaxed">
                        Efetue um teste de loot simulado para esta atividade sob a influência das leis canônicas do ecossistema de recursos.
                      </p>

                      {/* Interactive Button */}
                      <div className="flex items-center gap-3">
                        <button
                          onClick={() => {
                            const activityLootRolls = {
                              HUNT: [
                                { text: "Encontrado: 340 Moedas de Bronze, 1x Molde de Espada (Comum)", quality: "comum" },
                                { text: "Encontrado: 560 Moedas de Bronze, 3x Essência Primordial (Raro)", quality: "raro" },
                                { text: "Encontrado: 120 Moedas de Prata, 1x Joia Sombria Perfeita (B+ Scarcity Roll!)", quality: "lendario" },
                                { text: "Encontrado: 150 Moedas de Bronze", quality: "comum" }
                              ],
                              EXPLORATION: [
                                { text: "Exploration rewards are abstract and undefined pending future canonical Crafting Bible definition.", quality: "info" }
                              ],
                              QUEST: [], // unused but kept for type safety
                              BOSS: [
                                { text: "Entregue no Baú Protegido (Safe Zone): 1x Moeda de Ouro, 1x Molde de Armadura Dracônica (Lendário!)", quality: "lendario" },
                                { text: "Entregue no Baú Protegido (Safe Zone): 450 Moedas de Prata", quality: "raro" },
                                { text: "Entregue no Baú Protegido (Safe Zone): 1x Moeda de Diamante (Consistência Absoluta!)", quality: "lendario" }
                              ]
                            };

                            const rolls = activityLootRolls[selectedActivity as "HUNT" | "EXPLORATION" | "BOSS"];
                            if (!rolls || rolls.length === 0) return;
                            const selectedRoll = rolls[Math.floor(Math.random() * rolls.length)];
                            
                            addBibleLog(
                              selectedRoll.quality === "lendario" ? "success" : selectedRoll.quality === "raro" ? "info" : "default" as any,
                              `[Loot Simulator] ${selectedRoll.text}`
                            );
                          }}
                          className="flex-1 bg-slate-800 hover:bg-slate-700 active:bg-slate-900 border border-slate-700 text-xs text-slate-100 font-bold py-2 px-4 rounded-xl transition-all duration-150 flex items-center justify-center gap-1.5"
                        >
                          <RefreshCw className="w-3.5 h-3.5 animate-spin-slow text-amber-400" />
                          Rolar Recompensa de Saque (Simulação)
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* ================================================================== */}
              {/* CRAFTING & ITEM TIER CANONICAL SIMULATOR (v2) */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      Subsistema de Forja Universal & Tiers de Itens v2
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <Hammer className="w-5 h-5 text-amber-500 animate-bounce-slow" />
                      Simulador de Forja Determinista e Scarcity de Tiers
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Taxa de Sucesso</span>
                    <span className="text-xs font-mono font-bold text-emerald-400">100% Determinista (No RNG)</span>
                  </div>
                </div>

                {/* Description and Principles Banner */}
                <div className="bg-slate-950/60 p-4 rounded-xl border border-slate-850 space-y-3">
                  <h5 className="text-[10.5px] font-extrabold text-slate-300 uppercase tracking-wider">Regras Canônicas Ativas (Lei de Identidade Fixa):</h5>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/40 text-[10px] font-mono text-slate-400 leading-relaxed">
                      <span className="text-amber-400 font-bold block mb-1">▪ Universalidade (Sem Profissões)</span>
                      Qualquer herói pode forjar qualquer item elegível. Não há classes de crafting, níveis de ferreiro ou árvores de especialização.
                    </div>
                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/40 text-[10px] font-mono text-slate-400 leading-relaxed">
                      <span className="text-amber-400 font-bold block mb-1">▪ Limite Estrito de Tiers (T4/T5 Drop-Only)</span>
                      Equipamentos de Tier 4 e Tier 5 são dropados apenas em atividades de alto risco e nunca podem ser craftados.
                    </div>
                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/40 text-[10px] font-mono text-slate-400 leading-relaxed">
                      <span className="text-amber-400 font-bold block mb-1">▪ Identidade Fixa (No Stat RNG)</span>
                      O crafting produz apenas o template determinista exato, sem upgrades, sem refinamentos e sem alteração de atributos.
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Selector */}
                  <div className="lg:col-span-5 space-y-4">
                    <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider">Selecione o Equipamento do Catálogo</h4>
                    
                    <div className="grid grid-cols-1 gap-2.5 max-h-[380px] overflow-y-auto pr-1">
                      {[
                        { id: "weapon_t1", name: "Weapon Template T1", tier: 1, type: "Weapon", craftable: true },
                        { id: "armor_t1", name: "Armor Template T1", tier: 1, type: "Armor", craftable: true },
                        { id: "accessory_t1", name: "Accessory Template T1", tier: 1, type: "Accessory", craftable: true },
                        { id: "weapon_t2", name: "Weapon Template T2", tier: 2, type: "Weapon", craftable: true },
                        { id: "armor_t2", name: "Armor Template T2", tier: 2, type: "Armor", craftable: true },
                        { id: "accessory_t2", name: "Accessory Template T2", tier: 2, type: "Accessory", craftable: true },
                        { id: "weapon_t3", name: "Weapon Template T3", tier: 3, type: "Weapon", craftable: true, hybrid: true },
                        { id: "armor_t3", name: "Armor Template T3", tier: 3, type: "Armor", craftable: true, hybrid: true },
                        { id: "accessory_t3", name: "Accessory Template T3", tier: 3, type: "Accessory", craftable: true, hybrid: true },
                        { id: "weapon_t4", name: "Weapon Template T4", tier: 4, type: "Weapon", craftable: false },
                        { id: "armor_t4", name: "Armor Template T4", tier: 4, type: "Armor", craftable: false },
                        { id: "weapon_t5", name: "Weapon Template T5", tier: 5, type: "Weapon", craftable: false },
                        { id: "armor_t5", name: "Armor Template T5", tier: 5, type: "Armor", craftable: false }
                      ].map((item) => (
                        <button
                          key={item.id}
                          onClick={() => setSelectedCraftItem(item.id)}
                          className={`w-full text-left p-3.5 rounded-xl border transition-all duration-150 flex items-center justify-between group ${
                            selectedCraftItem === item.id
                              ? "bg-amber-500/10 border-amber-500/50 text-amber-300"
                              : "bg-slate-950/40 border-slate-850 text-slate-400 hover:border-slate-800 hover:text-slate-200"
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            <span className={`text-[9px] font-extrabold uppercase px-1.5 py-0.5 rounded font-mono ${
                              item.tier === 1 ? "bg-slate-800 text-slate-300" :
                              item.tier === 2 ? "bg-sky-950 text-sky-400" :
                              item.tier === 3 ? "bg-amber-950 text-amber-400" :
                              item.tier === 4 ? "bg-rose-950 text-rose-400" : "bg-violet-950 text-violet-400 animate-pulse"
                            }`}>
                              T{item.tier}
                            </span>
                            <div>
                              <span className="text-xs font-bold block group-hover:text-slate-100 transition-colors">{item.name}</span>
                              <span className="text-[9px] text-slate-500 font-mono uppercase">{item.type}</span>
                            </div>
                          </div>

                          <div className="text-right font-mono">
                            <span className="text-[8px] text-slate-500 block uppercase font-bold">Obtenção</span>
                            <span className={`text-[10px] font-bold ${
                              item.craftable 
                                ? item.hybrid ? "text-amber-400" : "text-emerald-400" 
                                : "text-rose-400"
                            }`}>
                              {item.craftable ? item.hybrid ? "Forja + Drops" : "Apenas Forja" : "Apenas Drops (T4/T5)"}
                            </span>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Right Column: Detailed Craft Simulator */}
                  <div className="lg:col-span-7 bg-slate-950/60 border border-slate-850 rounded-2xl p-5 flex flex-col justify-between min-h-[380px]">
                    {(() => {
                      const itemsMap: Record<string, {
                        name: string;
                        tier: number;
                        type: string;
                        stats: string;
                        craftable: boolean;
                        hybrid?: boolean;
                        materials?: { name: string; qty: number }[];
                        quest?: string;
                        desc: string;
                      }> = {
                        weapon_t1: {
                          name: "Weapon Template T1",
                          tier: 1,
                          type: "weapon",
                          stats: "Dano: 15.0",
                          craftable: true,
                          materials: [{ name: "Combat Material T1", qty: 5 }],
                          quest: "Quest: Initial Combat Training",
                          desc: "Early game starter weapon template produced deterministically under Universal Crafting."
                        },
                        armor_t1: {
                          name: "Armor Template T1",
                          tier: 1,
                          type: "armor",
                          stats: "Defesa: 15.0",
                          craftable: true,
                          materials: [{ name: "Combat Material T1", qty: 8 }],
                          quest: "Quest: Arming the Outpost",
                          desc: "Early game starter armor template produced deterministically under Universal Crafting."
                        },
                        accessory_t1: {
                          name: "Accessory Template T1",
                          tier: 1,
                          type: "accessory",
                          stats: "Resistência: 5.0 | Crítico: 1%",
                          craftable: true,
                          materials: [{ name: "Combat Material T1", qty: 4 }],
                          quest: "Quest: Adornments of the Outpost",
                          desc: "Early game starter accessory template produced deterministically under Universal Crafting."
                        },
                        weapon_t2: {
                          name: "Weapon Template T2",
                          tier: 2,
                          type: "weapon",
                          stats: "Dano: 35.0 | Crítico: 2%",
                          craftable: true,
                          materials: [{ name: "Combat Material T1", qty: 10 }, { name: "Combat Material T2", qty: 3 }],
                          quest: "Quest: Blade of structured narration",
                          desc: "Low-mid game weapon template. Recipe unlocked through structured narrative events."
                        },
                        armor_t2: {
                          name: "Armor Template T2",
                          tier: 2,
                          type: "armor",
                          stats: "Defesa: 35.0",
                          craftable: true,
                          materials: [{ name: "Combat Material T1", qty: 15 }, { name: "Combat Material T2", qty: 5 }],
                          quest: "Quest: Armor of narrative milestones",
                          desc: "Low-mid game armor template. Recipe unlocked through structured narrative events."
                        },
                        accessory_t2: {
                          name: "Accessory Template T2",
                          tier: 2,
                          type: "accessory",
                          stats: "Resistência: 10.0 | Crítico: 3%",
                          craftable: true,
                          materials: [{ name: "Combat Material T2", qty: 6 }],
                          quest: "Quest: Adornments of the Deep",
                          desc: "Low-mid game accessory template. Recipe unlocked through structured narrative events."
                        },
                        weapon_t3: {
                          name: "Weapon Template T3",
                          tier: 3,
                          type: "weapon",
                          stats: "Dano: 65.0 | Crítico: 4%",
                          craftable: true,
                          hybrid: true,
                          materials: [{ name: "Combat Material T2", qty: 12 }, { name: "Combat Material T3", qty: 4 }],
                          quest: "Quest: Pinnacle of the Forge",
                          desc: "Mid game hybrid transition weapon template. Craftable or rarely dropped by Elites."
                        },
                        armor_t3: {
                          name: "Armor Template T3",
                          tier: 3,
                          type: "armor",
                          stats: "Defesa: 65.0 | Resistência: 10.0",
                          craftable: true,
                          hybrid: true,
                          materials: [{ name: "Combat Material T2", qty: 20 }, { name: "Combat Material T3", qty: 6 }],
                          quest: "Quest: Iron Citadel Trial",
                          desc: "Mid game hybrid transition armor template. Craftable or rarely dropped by Elites."
                        },
                        accessory_t3: {
                          name: "Accessory Template T3",
                          tier: 3,
                          type: "accessory",
                          stats: "Resistência: 20.0 | Crítico: 5%",
                          craftable: true,
                          hybrid: true,
                          materials: [{ name: "Combat Material T2", qty: 15 }, { name: "Combat Material T3", qty: 5 }],
                          quest: "Quest: Adornments of the Citadel",
                          desc: "Mid game hybrid transition accessory template. Craftable or rarely dropped by Elites."
                        },
                        weapon_t4: {
                          name: "Weapon Template T4",
                          tier: 4,
                          type: "weapon",
                          stats: "Dano: 105.0 | Resistência: 10.0 | Crítico: 6%",
                          craftable: false,
                          desc: "Late game master equipment template. Fully drop-only, exclusively obtained from high-tier content (Hunts, Bosses, Elite encounters)."
                        },
                        armor_t4: {
                          name: "Armor Template T4",
                          tier: 4,
                          type: "armor",
                          stats: "Defesa: 110.0 | Resistência: 25.0",
                          craftable: false,
                          desc: "Late game master armor template. Fully drop-only, exclusively obtained from high-tier content."
                        },
                        weapon_t5: {
                          name: "Weapon Template T5",
                          tier: 5,
                          type: "weapon",
                          stats: "Dano: 150.0 | Resistência: 20.0 | Crítico: 10%",
                          craftable: false,
                          desc: "Endgame apex weapon template. Drop-only under extreme rarity rules from apex-tier content (Bosses, sovereign dragons)."
                        },
                        armor_t5: {
                          name: "Armor Template T5",
                          tier: 5,
                          type: "armor",
                          stats: "Defesa: 160.0 | Resistência: 50.0",
                          craftable: false,
                          desc: "Endgame apex armor template. Drop-only under extreme rarity rules from apex-tier content."
                        }
                      };
                      const itemDetails = itemsMap[selectedCraftItem];

                      if (!itemDetails) return null;

                      return (
                        <div className="space-y-4 flex flex-col justify-between h-full">
                          <div className="space-y-3.5">
                            <div className="flex justify-between items-start">
                              <div>
                                <h4 className="font-extrabold text-sm text-slate-100 flex items-center gap-1.5">
                                  {itemDetails.name}
                                  {itemDetails.craftable ? (
                                    <span className="text-[9px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-1.5 py-0.5 rounded font-mono">
                                      CRAFTABLE
                                    </span>
                                  ) : (
                                    <span className="text-[9px] bg-rose-500/10 text-rose-400 border border-rose-500/20 px-1.5 py-0.5 rounded font-mono">
                                      DROP-ONLY (NON-CRAFTABLE)
                                    </span>
                                  )}
                                </h4>
                                <p className="text-[10.5px] text-slate-400 leading-relaxed mt-1">
                                  {itemDetails.desc}
                                </p>
                              </div>
                              <div className="text-right">
                                <span className="text-[9px] text-slate-500 block uppercase tracking-wider font-bold font-mono">Estrutura de Tier</span>
                                <span className="text-xs font-mono font-bold text-amber-400">Tier {itemDetails.tier}</span>
                              </div>
                            </div>

                            {/* Stats Card */}
                            <div className="bg-slate-900/60 p-3 rounded-xl border border-slate-800/40">
                              <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Atributos Deterministas (Template Fixo)</span>
                              <span className="text-xs font-mono font-bold text-slate-200">{itemDetails.stats || "Atributos de ponta do endgame"}</span>
                            </div>

                            {/* Crafting Requirements or Drop Rules */}
                            {itemDetails.craftable ? (
                              <div className="space-y-2">
                                <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Requisitos do Projeto de Forja</span>
                                <div className="grid grid-cols-2 gap-2">
                                  {itemDetails.materials.map((mat, i) => (
                                    <div key={i} className="bg-slate-900 border border-slate-800/50 px-2.5 py-1.5 rounded-lg flex justify-between text-[11px] font-mono">
                                      <span className="text-slate-400">{mat.name}</span>
                                      <span className="text-amber-400 font-bold">x{mat.qty}</span>
                                    </div>
                                  ))}
                                </div>
                                <div className="bg-slate-900/50 p-2.5 rounded-lg border border-slate-800/20 text-[10px] text-slate-400 font-mono flex items-center gap-2">
                                  <BookOpen className="w-3.5 h-3.5 text-sky-400 shrink-0" />
                                  <span>Unloque de Receita: <strong className="text-sky-400">{itemDetails.quest}</strong></span>
                                </div>
                              </div>
                            ) : (
                              <div className="bg-rose-500/5 border border-rose-500/10 p-4 rounded-xl text-center space-y-2">
                                <Lock className="w-6 h-6 text-rose-500 mx-auto" />
                                <div>
                                  <span className="text-rose-400 font-bold block text-xs">BLOQUEIO CANÔNICO DE TIER {itemDetails.tier}</span>
                                  <p className="text-[10px] text-slate-400 max-w-md mx-auto leading-relaxed mt-1">
                                    Em total conformidade com a <strong className="text-slate-300">Lei de Scarcity do Item Tier v2</strong>, equipamentos de Tier 4 e Tier 5 nunca podem ser produzidos artificialmente. Eles devem ser conquistados organicamente através de saques em eventos de mundo aberto de alta intensidade.
                                  </p>
                                </div>
                              </div>
                            )}
                          </div>

                          {/* Interactive Forge Actions */}
                          <div className="pt-4 border-t border-slate-900/80">
                            <button
                              onClick={() => {
                                if (!itemDetails.craftable) {
                                  addBibleLog(
                                    "error",
                                    `[Crafting Bible] ERRO CANÔNICO: Itens de Tier ${itemDetails.tier} (${itemDetails.name}) não podem ser craftados sob nenhuma circunstância!`
                                  );
                                  return;
                                }
                                
                                // Deterministic Craft Success
                                addBibleLog(
                                  "success",
                                  `[Crafting Bible] Sucesso! Item de Tier ${itemDetails.tier} [${itemDetails.name}] forjado determinísticamente sob a Forja Universal. Sem RNG de Atributos, 100% Taxa de Sucesso.`
                                );
                              }}
                              className={`w-full py-2.5 px-4 rounded-xl text-xs font-bold transition-all duration-150 flex items-center justify-center gap-1.5 ${
                                itemDetails.craftable
                                  ? "bg-amber-500 hover:bg-amber-400 active:bg-amber-600 text-slate-950 shadow-lg shadow-amber-500/10"
                                  : "bg-slate-800 text-slate-500 cursor-not-allowed border border-slate-700/50"
                              }`}
                            >
                              <Hammer className="w-4 h-4" />
                              {itemDetails.craftable 
                                ? `Forjar ${itemDetails.name} (Universal & Determinista)` 
                                : "Não-Craftável (Exclusivo de Drop de Elite/Boss)"}
                            </button>
                          </div>
                        </div>
                      );
                    })()}
                  </div>
                </div>
              </div>

              {/* ================================================================== */}
              {/* GUILD & SOCIAL BIBLE CANONICAL SIMULATOR */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      Subsistema de Coordenação Social & Guildas (v1)
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <Users className="w-5 h-5 text-amber-500 animate-bounce-slow" />
                      Painel de Gestão e Simulação de Guildas
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Foco Canônico</span>
                    <span className="text-xs font-mono font-bold text-amber-400">Lightweight Social Group</span>
                  </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Guild Info & Rules */}
                  <div className="lg:col-span-5 space-y-4">
                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-4 space-y-3">
                      <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                        <Users className="w-4 h-4 text-sky-400" />
                        Status da Guilda Simulado
                      </h4>

                      <div className="space-y-2 text-xs">
                        <div className="flex justify-between border-b border-slate-900 pb-1.5">
                          <span className="text-slate-400">Nome:</span>
                          <span className="font-bold text-slate-200">{simulatedGuildName}</span>
                        </div>
                        <div className="flex justify-between border-b border-slate-900 pb-1.5">
                          <span className="text-slate-400">Membros:</span>
                          <span className="font-mono font-bold text-slate-200">{guildMembersCount} / 30</span>
                        </div>
                        <div className="flex justify-between border-b border-slate-900 pb-1.5">
                          <span className="text-slate-400">Sede (Guild House):</span>
                          <span className={`font-bold ${hasGuildHouse ? "text-emerald-400" : "text-rose-400"}`}>
                            {hasGuildHouse ? "Adquirida (Gold-Only)" : "Nenhuma Sede"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-slate-400">Seu Ouro Simulado:</span>
                          <span className="font-mono font-bold text-amber-400">{guildGoldBalance} Ouro</span>
                        </div>
                      </div>
                    </div>

                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-4 space-y-3">
                      <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                        <Shield className="w-4 h-4 text-amber-400" />
                        Diretrizes e Regras do Sistema
                      </h4>
                      <ul className="space-y-2 text-[10px] text-slate-400 font-mono">
                        <li className="flex items-start gap-1.5">
                          <span className="text-amber-400">•</span>
                          <span><strong>guild-size-limit:</strong> Limite máximo estrito de 30 membros para preservar coesão.</span>
                        </li>
                        <li className="flex items-start gap-1.5">
                          <span className="text-amber-400">•</span>
                          <span><strong>guild-rank-rule:</strong> Três patentes fixas: Líder, Oficial, Membro.</span>
                        </li>
                        <li className="flex items-start gap-1.5">
                          <span className="text-amber-400">•</span>
                          <span><strong>guild-storage-rule:</strong> Não existe baú ou banco compartilhado (economia privada).</span>
                        </li>
                        <li className="flex items-start gap-1.5">
                          <span className="text-amber-400">•</span>
                          <span><strong>guild-alliance-rule:</strong> Sem sistemas formais de aliança (antizerg/emergente).</span>
                        </li>
                        <li className="flex items-start gap-1.5">
                          <span className="text-amber-400">•</span>
                          <span><strong>guild-pvp-rule:</strong> PvP continua ativo entre membros fora de zonas seguras (traição permitida).</span>
                        </li>
                      </ul>
                    </div>
                  </div>

                  {/* Right Column: Actions & Log */}
                  <div className="lg:col-span-7 bg-slate-950/60 border border-slate-850 rounded-xl p-5 flex flex-col justify-between min-h-[350px]">
                    <div className="space-y-4">
                      <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                        <Activity className="w-4 h-4 text-emerald-400" />
                        Simulador de Ações Sociais
                      </h4>

                      <div className="grid grid-cols-2 gap-3">
                        {/* 1. Free Creation */}
                        <button
                          onClick={() => {
                            setSimulatedGuildName("Guardiões do Abismo");
                            setGuildMembersCount(1);
                            addBibleLog("success", `[GuildEngine] Sucesso! Guilda "Guardiões do Abismo" criada livremente. Custo: 0 Moedas (Conforme guild-creation-rule).`);
                          }}
                          className="bg-slate-900 border border-slate-800 hover:border-amber-500/40 p-3 rounded-lg text-left transition flex flex-col justify-between gap-1 text-[11px]"
                        >
                          <span className="font-bold text-slate-200">Criar Guilda (Livre)</span>
                          <span className="text-[9px] text-slate-500">Custo zero de moedas</span>
                        </button>

                        {/* 2. Simulate Member Invites */}
                        <button
                          onClick={() => {
                            if (guildMembersCount >= 30) {
                              addBibleLog("error", `[GuildEngine] ERRO CANÔNICO: Limite estrito de 30 membros atingido! Não é possível convidar mais membros (guild-size-limit).`);
                            } else {
                              const newCount = guildMembersCount + 1;
                              setGuildMembersCount(newCount);
                              addBibleLog("info", `[GuildEngine] Novo membro aceito na guilda. Total: ${newCount}/30 membros.`);
                            }
                          }}
                          className="bg-slate-900 border border-slate-800 hover:border-amber-500/40 p-3 rounded-lg text-left transition flex flex-col justify-between gap-1 text-[11px]"
                        >
                          <span className="font-bold text-slate-200 flex items-center gap-1.5">
                            <UserPlus className="w-3.5 h-3.5 text-sky-400" />
                            Convidar Membro (+1)
                          </span>
                          <span className="text-[9px] text-slate-500">Cap estrito de 30 membros</span>
                        </button>

                        {/* 3. Buy Guild House */}
                        <button
                          onClick={() => {
                            if (hasGuildHouse) {
                              addBibleLog("warning", "[GuildEngine] Sua guilda já possui uma Sede ativa.");
                              return;
                            }
                            if (guildGoldBalance < 100) {
                              addBibleLog("error", `[GuildEngine] Ouro Insuficiente! A Sede custa exatamente 100 Moedas de Ouro (Apenas Gold-Only, sem fiat ou moedas premium).`);
                            } else {
                              setGuildGoldBalance(prev => prev - 100);
                              setHasGuildHouse(true);
                              addBibleLog("success", `[GuildEngine] Sede de Guilda Adquirida! 100 Ouro deduzidos (guild-house-acquisition-rule). Sede habilitada como Hub de Utilidade + Social.`);
                            }
                          }}
                          className={`p-3 rounded-lg text-left transition flex flex-col justify-between gap-1 text-[11px] ${
                            hasGuildHouse 
                              ? "bg-slate-900 border border-emerald-900/40 opacity-70 cursor-not-allowed" 
                              : "bg-slate-900 border border-slate-800 hover:border-amber-500/40"
                          }`}
                        >
                          <span className="font-bold text-slate-200 flex items-center gap-1.5">
                            <Home className="w-3.5 h-3.5 text-amber-400" />
                            Comprar Sede (100 Ouro)
                          </span>
                          <span className="text-[9px] text-slate-500">Exclusivamente via Ouro</span>
                        </button>

                        {/* 4. Open-World friendly fire test */}
                        <button
                          onClick={() => {
                            addBibleLog("warning", `[GuildEngine] PvP Embate! Membro "Kaelen" e "Eldrin" duelaram fora da zona de segurança. Sem imunidade de guilda sob a diretriz guild-pvp-rule!`);
                          }}
                          className="bg-slate-900 border border-slate-800 hover:border-amber-500/40 p-3 rounded-lg text-left transition flex flex-col justify-between gap-1 text-[11px]"
                        >
                          <span className="font-bold text-rose-400">Ativar Combate Interno</span>
                          <span className="text-[9px] text-slate-500">Testar PvP fora de Safe Zones</span>
                        </button>
                      </div>
                    </div>

                    {/* Sede Utilities status when acquired */}
                    <div className="bg-slate-900/40 border border-slate-850 p-4 rounded-xl mt-4">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Utilidades Ativas da Sede</span>
                      {hasGuildHouse ? (
                        <div className="grid grid-cols-2 gap-2 text-[10px] text-slate-300 font-mono mt-1.5">
                          <div className="flex items-center gap-1">
                            <CheckCircle className="w-3 h-3 text-emerald-400" /> Espaço de Reunião
                          </div>
                          <div className="flex items-center gap-1">
                            <CheckCircle className="w-3 h-3 text-emerald-400" /> Estação de Forja
                          </div>
                          <div className="flex items-center gap-1">
                            <CheckCircle className="w-3 h-3 text-emerald-400" /> NPC de Suprimentos
                          </div>
                          <div className="flex items-center gap-1">
                            <CheckCircle className="w-3 h-3 text-emerald-400" /> Mural de Contratos
                          </div>
                        </div>
                      ) : (
                        <span className="text-[10px] text-slate-500 font-mono block mt-1">Sede Inativa. Adquira a Sede por 100 Moedas de Ouro para ativar as facilidades de crafting e murais de contrato privado.</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {/* ================================================================== */}
              {/* RACE & CHARACTER CREATION BIBLE CANONICAL SIMULATOR */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      Subsistema de Criação de Personagens (v1)
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <UserPlus className="w-5 h-5 text-amber-500 animate-pulse" />
                      Painel de Criação de Herói & Simulação de Paridade de Raças
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Diretriz Canônica</span>
                    <span className="text-xs font-mono font-bold text-amber-400">Identity First, Meta Free</span>
                  </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Interactive Form */}
                  <div className="lg:col-span-5 space-y-4">
                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                      <div className="flex items-center gap-1.5 text-xs font-bold text-slate-200 uppercase tracking-wider">
                        <Sparkles className="w-4 h-4 text-amber-400" />
                        <span>Novo Herói (Character Creator)</span>
                      </div>

                      {/* Name input */}
                      <div className="space-y-1">
                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Nome do Herói</label>
                        <input
                          type="text"
                          value={createdCharName}
                          onChange={(e) => setCreatedCharName(e.target.value)}
                          placeholder="Ex: Valen_Sunblade"
                          className="w-full bg-slate-900/80 border border-slate-800 rounded-lg px-3 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-amber-500/50"
                        />
                      </div>

                      {/* Race Selection */}
                      <div className="space-y-1">
                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Raça do Personagem (playable-races-rule)</label>
                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-1.5">
                          {(["Forest Elf", "Human", "Dwarf", "Ice Elf", "Green Orc"] as const).map((r) => (
                            <button
                              key={r}
                              type="button"
                              onClick={() => {
                                setCreatedCharRace(r);
                                setCharacterCreationLog(`Selecionado: ${r}. Todas as estatísticas de combate permanecem idênticas.`);
                              }}
                              className={`py-1.5 px-2 rounded-lg text-[10.5px] font-medium transition cursor-pointer text-center ${
                                createdCharRace === r
                                  ? "bg-amber-500 text-slate-950 font-bold border border-amber-400"
                                  : "bg-slate-900 hover:bg-slate-850 text-slate-400 border border-slate-800"
                              }`}
                            >
                              {r}
                            </button>
                          ))}
                        </div>
                      </div>

                      {/* Class Selection */}
                      <div className="space-y-1">
                        <div className="flex justify-between items-center">
                          <label className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Classe (class-race-freedom-rule)</label>
                          <span className="text-[8px] text-amber-400 font-mono uppercase font-bold">Livre de restrições</span>
                        </div>
                        <div className="grid grid-cols-4 gap-1.5">
                          {["Knight", "Mage", "Archer", "Cleric"].map((c) => (
                            <button
                              key={c}
                              type="button"
                              onClick={() => setCreatedCharClass(c)}
                              className={`py-1.5 rounded-lg text-[10.5px] font-medium transition cursor-pointer text-center ${
                                createdCharClass === c
                                  ? "bg-sky-600 text-slate-950 font-bold border border-sky-400"
                                  : "bg-slate-900 hover:bg-slate-850 text-slate-400 border border-slate-800"
                              }`}
                            >
                              {c}
                            </button>
                          ))}
                        </div>
                      </div>

                      {/* Gender Selection */}
                      <div className="space-y-1">
                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Gênero (gender-selection-rule)</label>
                        <div className="grid grid-cols-2 gap-2">
                          {(["Male", "Female"] as const).map((g) => (
                            <button
                              key={g}
                              type="button"
                              onClick={() => setCreatedCharGender(g)}
                              className={`py-1.5 rounded-lg text-[10.5px] font-medium transition cursor-pointer text-center ${
                                createdCharGender === g
                                  ? "bg-indigo-600 text-slate-100 font-bold border border-indigo-400"
                                  : "bg-slate-900 hover:bg-slate-850 text-slate-400 border border-slate-800"
                              }`}
                            >
                              {g === "Male" ? "Masculino" : "Feminino"}
                            </button>
                          ))}
                        </div>
                      </div>

                      {/* Action Trigger */}
                      <button
                        onClick={() => {
                          if (!createdCharName.trim()) {
                            addBibleLog("error", "[Criação] Insira um nome válido para seu herói.");
                            return;
                          }
                          addBibleLog("success", `[Hero Creator] Personagem Criado com Sucesso! Bem-vindo, ${createdCharName} (${createdCharRace} ${createdCharClass} - ${createdCharGender}).`);
                          addBibleLog("info", `[Onboarding] Posição inicial estabelecida em Ironhold Bastion (ironhold-starting-city-rule) no mesmo mapa inicial unificado (starting-zone-rule).`);
                          addBibleLog("success", `[Neutrality] Paridade Estatística: Nenhuma vantagem ou bônus de combate aplicado (race-neutrality-rule). Hitbox padronizado para todos (racial-hitbox-rule).`);
                          setCharacterCreationLog(`Criado com sucesso! Iniciou em Ironhold Bastion.`);
                        }}
                        className="w-full py-2 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-slate-950 font-extrabold rounded-lg text-xs transition uppercase tracking-widest cursor-pointer"
                      >
                        Materializar Novo Personagem
                      </button>
                    </div>

                    {/* Creation Console Log */}
                    <div className="bg-slate-950/40 border border-slate-850 p-4 rounded-xl">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Console do Servidor de Criação</span>
                      <div className="mt-1 bg-black/60 font-mono text-[10.5px] p-2.5 rounded border border-slate-900 text-slate-300 leading-normal">
                        <span className="text-amber-500 font-bold">&gt;</span> {characterCreationLog}
                      </div>
                    </div>
                  </div>

                  {/* Right Column: Canonical Rules & Parity Verification */}
                  <div className="lg:col-span-7 space-y-4">
                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                      <div className="flex items-center justify-between border-b border-slate-900 pb-2.5">
                        <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                          <Shield className="w-4 h-4 text-emerald-400" />
                          Mapeamento de Regras & Paridade Técnica (Strict-Parity)
                        </span>
                        <span className="text-[9px] font-mono font-bold text-emerald-400 bg-emerald-950/50 border border-emerald-900 px-2 py-0.5 rounded">
                          Verificado
                        </span>
                      </div>

                      {/* Technical specifications proving race-neutrality */}
                      <div className="bg-slate-900/60 border border-slate-800 p-4 rounded-lg space-y-3">
                        <span className="text-[10px] text-amber-400 font-mono font-bold uppercase tracking-wider block">Matriz de Paridade de Atributos Globais</span>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-xs">
                          <div className="bg-slate-950/60 border border-slate-850 p-2 rounded">
                            <span className="text-[8px] text-slate-500 uppercase block font-mono">Bônus Dano</span>
                            <span className="font-bold font-mono text-slate-300">+0.0%</span>
                          </div>
                          <div className="bg-slate-950/60 border border-slate-850 p-2 rounded">
                            <span className="text-[8px] text-slate-500 uppercase block font-mono">Bônus Defesa</span>
                            <span className="font-bold font-mono text-slate-300">+0.0%</span>
                          </div>
                          <div className="bg-slate-950/60 border border-slate-850 p-2 rounded">
                            <span className="text-[8px] text-slate-500 uppercase block font-mono">Velocidade</span>
                            <span className="font-bold font-mono text-slate-300">100% (Ident.)</span>
                          </div>
                          <div className="bg-slate-950/60 border border-slate-850 p-2 rounded">
                            <span className="text-[8px] text-slate-500 uppercase block font-mono">Resistências</span>
                            <span className="font-bold font-mono text-slate-300">+0.0%</span>
                          </div>
                        </div>

                        {/* Standardized Hitbox verification */}
                        <div className="bg-slate-950/40 border border-slate-850/60 p-3 rounded-lg text-[11px] text-slate-400 flex items-center justify-between font-mono">
                          <span className="text-slate-300 font-bold">Standardized Hitbox Cylinder:</span>
                          <span className="text-emerald-400">Raio: 0.45m | Altura: 1.80m (racial-hitbox-rule)</span>
                        </div>
                      </div>

                      {/* Playable Races Details & Aesthetics */}
                      <div className="space-y-2">
                        <span className="text-[10px] text-slate-400 uppercase font-bold tracking-wider block">Verificação de Raça Permanente & Sem Evoluções (racial-permanence-rule):</span>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px]">
                          <div className="bg-slate-900/40 border border-slate-850/80 p-3 rounded-lg">
                            <span className="font-bold text-slate-300 block">Dwarf & Green Orc</span>
                            <span className="text-slate-500 text-[10px] block mt-0.5">Stout e musculoso, focado em barbas trançadas ou runas vulcânicas. Hitbox idêntico de 1.80m garante justiça nas frestas de PvP!</span>
                          </div>
                          <div className="bg-slate-900/40 border border-slate-850/80 p-3 rounded-lg">
                            <span className="font-bold text-slate-300 block">Forest Elf & Ice Elf</span>
                            <span className="text-slate-500 text-[10px] block mt-0.5">Estilos de folhas orgânicas ou adornos de cristal. Paridade estatística total impede metas de atributos ou velocidade racial.</span>
                          </div>
                        </div>
                      </div>

                      {/* Canon rules tags check list */}
                      <div className="bg-slate-900/30 border border-slate-850 p-3.5 rounded-lg space-y-2.5">
                        <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider block">Verificadores de Diretrizes Ativas (Bíblia de Raças):</span>
                        
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[10.5px]">
                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">race-core-principle</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Puramente estética e focada em identidade.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">playable-races-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Exatamente 5 raças oficiais.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">class-race-freedom-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Toda classe é utilizável por qualquer raça.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">ironhold-starting-city-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Inicia em Ironhold Bastion com Depot e Banco.</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* ================================================================== */}
              {/* CLASS & VOCATION BIBLE CANONICAL SIMULATOR */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-sky-400 font-mono tracking-widest uppercase block">
                      Subsistema de Vocações & Progressão (v1)
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <Shield className="w-5 h-5 text-sky-400 animate-pulse" />
                      Painel de Controle de Classes, Estatísticas & Restrições de Itens
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Foco Teórico</span>
                    <span className="text-xs font-mono font-bold text-sky-400">Delayed Specialization Model</span>
                  </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  {/* Left Column: Slider and Selection */}
                  <div className="lg:col-span-5 space-y-4">
                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                      <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                        <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                          <Activity className="w-4 h-4 text-sky-400" />
                          Progresso de Nível & Vínculo
                        </span>
                        <span className="text-[10.5px] font-mono font-bold text-amber-400">
                          Nível {vocationLevel}
                        </span>
                      </div>

                      {/* Slider Input */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-[10px] text-slate-500 font-bold uppercase tracking-wider">
                          <span>Aprendiz (Nvl 1-9)</span>
                          <span className="text-amber-500">Nível 10 (Selecione)</span>
                          <span>Nível Máximo (100)</span>
                        </div>
                        <input
                          type="range"
                          min="1"
                          max="100"
                          value={vocationLevel}
                          onChange={(e) => {
                            const newLevel = parseInt(e.target.value);
                            setVocationLevel(newLevel);
                            
                            if (newLevel < 10) {
                              if (selectedVocationClass !== "Novice") {
                                setSelectedVocationClass("Novice");
                                setVocationEquippedWeapon("Espada Básica [T1]");
                                setVocationConsoleLog(`Nível reduzido para ${newLevel}. Retornou ao estado classless Aprendiz (Novice) (novice-phase-rule).`);
                                addBibleLog("warning", `[Vocação] Nível reduzido para ${newLevel}. Revertido temporariamente ao estado classless Novice.`);
                              } else {
                                setVocationConsoleLog(`Nível ajustado para ${newLevel}. O personagem é um Aprendiz (Novice) neutro.`);
                              }
                            } else {
                              if (selectedVocationClass === "Novice") {
                                setVocationConsoleLog(`Nível ${newLevel} alcançado! Seleção de Classe desbloqueada (class-selection-rule). Selecione uma classe definitiva abaixo.`);
                              } else {
                                setVocationConsoleLog(`Nível ajustado para ${newLevel}. Atributos da classe ${selectedVocationClass} recalculados com sucesso.`);
                              }
                            }
                          }}
                          className="w-full accent-amber-500 bg-slate-900 h-1.5 rounded-lg appearance-none cursor-pointer"
                        />
                      </div>

                      {/* Class Selection Box */}
                      <div className="space-y-2.5">
                        <div className="flex justify-between items-center">
                          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Selecione uma Classe (nível 10+)</span>
                          {vocationLevel < 10 && (
                            <span className="text-[9px] text-rose-400 font-mono font-bold uppercase animate-pulse">Bloqueado até Nvl 10</span>
                          )}
                        </div>

                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                          {(["Knight", "Mage", "Archer", "Assassin", "Cleric"] as const).map((cls) => {
                            const isSelected = selectedVocationClass === cls;
                            const isDisabled = vocationLevel < 10;
                            return (
                              <button
                                key={cls}
                                type="button"
                                disabled={isDisabled}
                                onClick={() => {
                                  setSelectedVocationClass(cls);
                                  setVocationConsoleLog(`Especialização Concluída! Classe [${cls}] selecionada permanentemente de forma irreversível (class-permanence-rule).`);
                                  addBibleLog("success", `[Vocação] Nova Especialização: Você agora é oficialmente um ${cls}! Seus atributos base e habilidades exclusivas foram destravados.`);
                                }}
                                className={`py-2 px-2.5 rounded-lg text-xs font-bold transition-all text-center ${
                                  isSelected
                                    ? "bg-sky-600 text-slate-950 border border-sky-400"
                                    : isDisabled
                                    ? "bg-slate-950 text-slate-600 border border-slate-900 cursor-not-allowed opacity-50"
                                    : "bg-slate-900 hover:bg-slate-850 text-slate-300 border border-slate-800 cursor-pointer"
                                }`}
                              >
                                {cls}
                              </button>
                            );
                          })}
                        </div>
                      </div>

                      {/* Equipment restrictions preview */}
                      <div className="border-t border-slate-900 pt-3 space-y-2">
                        <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block">Equipamento e Compatibilidade (item-class-requirement-rule)</span>
                        <div className="space-y-1.5">
                          <label className="text-[9px] text-slate-400 uppercase font-mono block">Escolha uma Arma Avançada para Equipar:</label>
                          <div className="grid grid-cols-2 gap-1.5 text-[10.5px]">
                            {[
                              { name: "Espada Básica [T1]", req: "Qualquer" },
                              { name: "Espada de Batalha [T3]", req: "Knight" },
                              { name: "Cajado de Gelo [T3]", req: "Mage" },
                              { name: "Arco do Patrulheiro [T2]", req: "Archer" },
                              { name: "Adaga Sombria [T3]", req: "Assassin" },
                              { name: "Tomo Sagrado [T3]", req: "Cleric" }
                            ].map((wpn) => {
                              const isEquipped = vocationEquippedWeapon === wpn.name;
                              return (
                                <button
                                  key={wpn.name}
                                  onClick={() => {
                                    // Rule verification logic
                                    if (wpn.req === "Qualquer") {
                                      setVocationEquippedWeapon(wpn.name);
                                      setVocationConsoleLog(`Item [${wpn.name}] equipado com sucesso. Arma livre de restrições (novice-weapon-rule).`);
                                      addBibleLog("success", `[Equipar] Sucesso! Equipou [${wpn.name}] (arma comum desprovida de restrições).`);
                                    } else {
                                      if (vocationLevel < 10) {
                                        // Novice rules
                                        setVocationConsoleLog(`Erro: Novices só equipam armas iniciais básicas livremente em Ironhold. Armas avançadas de tier superior requerem nível 10 e classe compatível.`);
                                        addBibleLog("error", `[Equipar] Bloqueado! Como Novice (nível 1-9), você só pode equipar armas iniciais livres de restrições.`);
                                      } else if (selectedVocationClass === wpn.req) {
                                        setVocationEquippedWeapon(wpn.name);
                                        setVocationConsoleLog(`Item [${wpn.name}] equipado com sucesso! Requisito atendido: [${wpn.req}] (item-class-requirement-rule).`);
                                        addBibleLog("success", `[Equipar] Sucesso! Equipou [${wpn.name}] correspondente à sua classe ${selectedVocationClass}.`);
                                      } else {
                                        setVocationConsoleLog(`Erro: Falha no item-template! Requer classe [${wpn.req}]. Sua classe atual é [${selectedVocationClass}] (item-class-requirement-rule).`);
                                        addBibleLog("error", `[Equipar] Rejeitado! O item [${wpn.name}] requer a classe [${wpn.req}]. Sua classe atual é ${selectedVocationClass}.`);
                                      }
                                    }
                                  }}
                                  className={`p-1.5 rounded border text-left flex justify-between items-center transition cursor-pointer ${
                                    isEquipped
                                      ? "bg-amber-950/40 border-amber-500 text-amber-300"
                                      : "bg-slate-900 hover:bg-slate-850 border-slate-800 text-slate-400"
                                  }`}
                                >
                                  <span className="truncate">{wpn.name}</span>
                                  <span className="text-[8px] px-1 bg-slate-950 rounded text-slate-500 shrink-0 ml-1">
                                    {wpn.req}
                                  </span>
                                </button>
                              );
                            })}
                          </div>
                        </div>
                      </div>

                      {/* Sandbox Reset */}
                      <button
                        onClick={() => {
                          setVocationLevel(1);
                          setSelectedVocationClass("Novice");
                          setVocationEquippedWeapon("Espada Básica [T1]");
                          setVocationConsoleLog("Sandbox Reiniciado! Personagem retornou ao Nível 1 como Aprendiz (Novice).");
                          addBibleLog("info", "[Sandbox] Reset: Nível e Vocação restaurados para Novice.");
                        }}
                        className="w-full py-1.5 bg-slate-900 hover:bg-slate-800 text-[10.5px] text-slate-400 border border-slate-800 rounded-lg transition cursor-pointer font-bold font-mono"
                      >
                        Reiniciar Simulação do Sandbox
                      </button>
                    </div>

                    {/* Vocation Console Log */}
                    <div className="bg-slate-950/40 border border-slate-850 p-4 rounded-xl">
                      <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Console do Servidor de Vocações</span>
                      <div className="mt-1 bg-black/60 font-mono text-[10.5px] p-2.5 rounded border border-slate-900 text-slate-300 leading-normal">
                        <span className="text-sky-500 font-bold">&gt;</span> {vocationConsoleLog}
                      </div>
                    </div>
                  </div>

                  {/* Right Column: Stat Calculator & Canonical Rules */}
                  <div className="lg:col-span-7 space-y-4">
                    <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                      <div className="flex items-center justify-between border-b border-slate-900 pb-2.5">
                        <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                          <Database className="w-4 h-4 text-sky-400" />
                          Simulação de Atributos & Escalonamento de Nível
                        </span>
                        <span className="text-[9px] font-mono font-bold text-sky-400 bg-sky-950/50 border border-sky-900 px-2 py-0.5 rounded">
                          Ativo (v1.0.0)
                        </span>
                      </div>

                      {/* Class Attributes calculation logic */}
                      {(() => {
                        // Class attribute definitions
                        const classProfiles = {
                          Novice: { hpBase: 90, hpGrowth: 9, manaBase: 50, manaGrowth: 4, defBase: 8, defGrowth: 1.0, speed: 100, crit: 0.05, spells: ["Ataque Básico Corporal", "Arremesso de Pedra"] },
                          Knight: { hpBase: 150, hpGrowth: 15, manaBase: 40, manaGrowth: 2, defBase: 25, defGrowth: 2.0, speed: 95, crit: 0.05, spells: ["Investida Destruidora", "Bastião Provocador", "Fúria de Escudos"] },
                          Mage: { hpBase: 80, hpGrowth: 8, manaBase: 180, manaGrowth: 12, defBase: 5, defGrowth: 0.5, speed: 100, crit: 0.08, spells: ["Relâmpago da Tempestade", "Esfera de Fogo Celestial", "Nevasca Gélida"] },
                          Archer: { hpBase: 110, hpGrowth: 11, manaBase: 70, manaGrowth: 4, defBase: 12, defGrowth: 1.0, speed: 110, crit: 0.12, spells: ["Chuva de Flechas de Aço", "Tiro de Precisão Penetrante", "Disparo de Concussão"] },
                          Assassin: { hpBase: 100, hpGrowth: 10, manaBase: 60, manaGrowth: 3, defBase: 10, defGrowth: 1.0, speed: 120, crit: 0.20, spells: ["Golpe de Adaga Vital", "Passo de Sombras Invisível", "Envenenamento Ágil"] },
                          Cleric: { hpBase: 120, hpGrowth: 12, manaBase: 130, manaGrowth: 8, defBase: 15, defGrowth: 1.5, speed: 98, crit: 0.05, spells: ["Prece de Restauração Divina", "Barreira de Proteção Sagrada", "Explosão Purificadora de Luz"] }
                        };

                        const profile = classProfiles[selectedVocationClass];
                        const totalHP = profile.hpBase + (vocationLevel - 1) * profile.hpGrowth;
                        const totalMana = profile.manaBase + (vocationLevel - 1) * profile.manaGrowth;
                        const totalDef = Math.round(profile.defBase + (vocationLevel - 1) * profile.defGrowth);

                        return (
                          <div className="space-y-4">
                            {/* Visual Stats display */}
                            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                              {/* HP card */}
                              <div className="bg-slate-900/60 border border-slate-800/80 p-3 rounded-lg text-center">
                                <span className="text-[9px] text-slate-500 font-mono font-bold block uppercase">Pontos de Vida (HP)</span>
                                <span className="text-base font-bold font-mono text-emerald-400 mt-1 block">{totalHP} HP</span>
                                <span className="text-[8px] text-slate-500 block font-mono">Base: {profile.hpBase} (+{profile.hpGrowth}/nvl)</span>
                              </div>

                              {/* Mana card */}
                              <div className="bg-slate-900/60 border border-slate-800/80 p-3 rounded-lg text-center">
                                <span className="text-[9px] text-slate-500 font-mono font-bold block uppercase">Mana Max</span>
                                <span className="text-base font-bold font-mono text-sky-400 mt-1 block">{totalMana} MP</span>
                                <span className="text-[8px] text-slate-500 block font-mono">Base: {profile.manaBase} (+{profile.manaGrowth}/nvl)</span>
                              </div>

                              {/* Defense card */}
                              <div className="bg-slate-900/60 border border-slate-800/80 p-3 rounded-lg text-center col-span-2 sm:col-span-1">
                                <span className="text-[9px] text-slate-500 font-mono font-bold block uppercase">Física & Mag Def</span>
                                <span className="text-base font-bold font-mono text-slate-300 mt-1 block">{totalDef} DEF</span>
                                <span className="text-[8px] text-slate-500 block font-mono">Base: {profile.defBase} (+{profile.defGrowth}/nvl)</span>
                              </div>
                            </div>

                            {/* Additional attributes */}
                            <div className="grid grid-cols-2 gap-3 text-xs bg-slate-900/40 p-3 rounded-lg border border-slate-850 font-mono">
                              <div className="flex justify-between">
                                <span className="text-slate-500">Velocidade Base:</span>
                                <span className="text-amber-400 font-bold">{profile.speed} (unif.)</span>
                              </div>
                              <div className="flex justify-between">
                                <span className="text-slate-500">Chanc. Crítico:</span>
                                <span className="text-purple-400 font-bold">{(profile.crit * 100).toFixed(0)}%</span>
                              </div>
                            </div>

                            {/* Spells/Skills exclusivity preview */}
                            <div className="space-y-2">
                              <span className="text-[10px] text-slate-400 uppercase font-bold tracking-wider block">Habilidades Ativas Exclusivas Destravadas (class-skill-exclusivity-rule):</span>
                              <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                                {profile.spells.map((sp) => (
                                  <div key={sp} className="bg-slate-900 border border-slate-850 p-2.5 rounded-lg flex items-center gap-2">
                                    <div className="w-5 h-5 rounded bg-sky-950 border border-sky-800 text-sky-400 flex items-center justify-center font-mono text-[9px] font-bold">★</div>
                                    <span className="text-[11px] text-slate-300 font-medium truncate">{sp}</span>
                                  </div>
                                ))}
                              </div>
                            </div>
                          </div>
                        );
                      })()}

                      {/* Canon rules tags check list */}
                      <div className="bg-slate-900/30 border border-slate-850 p-3.5 rounded-lg space-y-2.5">
                        <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider block">Verificadores de Diretrizes Ativas (Bíblia de Classes):</span>
                        
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[10.5px]">
                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">class-core-principle</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Início neutro e especialização tardia.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">novice-phase-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Níveis 1-9 são Novices sem classe.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">class-selection-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Escolha habilitada no Nível 10.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">class-permanence-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Escolha irreversível e permanente.</p>
                            </div>
                          </div>

                          <div className="flex items-start gap-1.5 col-span-1 sm:col-span-2">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                            <div>
                              <span className="font-semibold text-slate-300 font-mono">item-class-requirement-rule</span>
                              <p className="text-[9.5px] text-slate-500 leading-none mt-0.5">Template de item dita restrições por classe (ex: Cajado de Gelo requer Mage).</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* ================================================================== */}
              {/* SPELL & SKILL BIBLE CANONICAL SIMULATOR */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      Subsistema de Habilidades, Magias & Recursos (v2)
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <Sparkles className="w-5 h-5 text-amber-400 animate-pulse" />
                      Painel de Controle do Conjurador & Monitor de Mana em Combate
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Foco Mecânico</span>
                    <span className="text-xs font-mono font-bold text-amber-400">Combat-Driven Regen</span>
                  </div>
                </div>

                {(() => {
                  // Profile calculations in local scope
                  const classProfiles = {
                    Novice: { hpBase: 90, hpGrowth: 9, manaBase: 50, manaGrowth: 4, classCoef: 0.3 },
                    Knight: { hpBase: 150, hpGrowth: 15, manaBase: 40, manaGrowth: 2, classCoef: 0.2 },
                    Mage: { hpBase: 80, hpGrowth: 8, manaBase: 180, manaGrowth: 12, classCoef: 1.0 },
                    Archer: { hpBase: 110, hpGrowth: 11, manaBase: 70, manaGrowth: 4, classCoef: 0.5 },
                    Assassin: { hpBase: 100, hpGrowth: 10, manaBase: 60, manaGrowth: 3, classCoef: 0.4 },
                    Cleric: { hpBase: 120, hpGrowth: 12, manaBase: 130, manaGrowth: 8, classCoef: 0.7 }
                  };
                  const activeProfile = classProfiles[selectedVocationClass];
                  const maxHpCalculated = activeProfile.hpBase + (vocationLevel - 1) * activeProfile.hpGrowth;
                  const maxManaCalculated = activeProfile.manaBase + (vocationLevel - 1) * activeProfile.manaGrowth;

                  // Define spell sets
                  const spellSets: Record<string, Array<{ id: string; name: string; manaCost: number; cooldown: number; basePower: number; elemental: string; desc: string }>> = {
                    Novice: [
                      { id: "basic_punch", name: "Golpe de Punho", manaCost: 0, cooldown: 1.5, basePower: 12, elemental: "Physical", desc: "Golpe simples de curto alcance livre de mana." },
                      { id: "rock_throw", name: "Lançar Pedra", manaCost: 5, cooldown: 2.0, basePower: 18, elemental: "Physical", desc: "Arremessa um projétil rústico do cenário." }
                    ],
                    Knight: [
                      { id: "slash_strike", name: "Slash Strike", manaCost: 10, cooldown: 2.0, basePower: 25, elemental: "Physical", desc: "Golpe pesado instantâneo com arma cortante." },
                      { id: "shield_bash", name: "Shield Bash", manaCost: 15, cooldown: 4.0, basePower: 35, elemental: "Physical", desc: "Atordoa o alvo usando o escudo, gerando impacto físico." }
                    ],
                    Mage: [
                      { id: "fireball", name: "Fireball", manaCost: 25, cooldown: 3.0, basePower: 60, elemental: "Fire", desc: "Dispara uma esfera incandescente causando alto dano elemental de fogo." },
                      { id: "ice_barrage", name: "Ice Barrage", manaCost: 20, cooldown: 2.5, basePower: 45, elemental: "Ice", desc: "Projeta fragmentos de gelo puro contra a barreira do oponente." }
                    ],
                    Archer: [
                      { id: "piercing_shot", name: "Piercing Shot", manaCost: 12, cooldown: 2.5, basePower: 35, elemental: "Physical", desc: "Disparo rápido perfurante que ignora parte da defesa física." },
                      { id: "rain_of_arrows", name: "Rain of Arrows", manaCost: 25, cooldown: 6.0, basePower: 55, elemental: "Physical", desc: "Chuva implacável de projéteis de aço em área." }
                    ],
                    Assassin: [
                      { id: "shadow_stab", name: "Shadow Stab", manaCost: 15, cooldown: 2.0, basePower: 40, elemental: "Shadow", desc: "Estocada rápida pelas sombras, causando dano crítico sombrio." },
                      { id: "poison_dart", name: "Poison Dart", manaCost: 18, cooldown: 4.0, basePower: 25, elemental: "Nature", desc: "Atira um dardo tóxico que aplica dano contínuo de natureza." }
                    ],
                    Cleric: [
                      { id: "holy_heal", name: "Holy Heal", manaCost: 20, cooldown: 3.0, basePower: 50, elemental: "Holy", desc: "Conjura luz restauradora curando os ferimentos do herói." },
                      { id: "divine_burst", name: "Divine Burst", manaCost: 15, cooldown: 2.5, basePower: 30, elemental: "Holy", desc: "Uma explosão celestial que bane as forças da escuridão." }
                    ]
                  };

                  const currentSpells = spellSets[selectedVocationClass];

                  // Active cooldown decrementor timer hook simulated locally via state trigger in other actions,
                  // and we also have the global interval initialized above.
                  const handleCastSpell = (spell: { id: string; name: string; manaCost: number; cooldown: number; basePower: number; elemental: string }) => {
                    // Cooldown check
                    if (spellCooldowns[spell.id] && spellCooldowns[spell.id] > 0) {
                      setSpellSkillConsoleLog(`Erro de Recarga: [${spell.name}] está em cooldown por mais ${spellCooldowns[spell.id]}s.`);
                      addBibleLog("error", `[Conjurador] Falha! Habilidade [${spell.name}] em tempo de recarga.`);
                      return;
                    }

                    // Mana check
                    if (spellMana < spell.manaCost) {
                      setSpellSkillConsoleLog(`Erro de Mana: Mana insuficiente para [${spell.name}]. Requer ${spell.manaCost} MP (atual: ${spellMana} MP).`);
                      addBibleLog("error", `[Conjurador] Mana insuficiente para conjurar [${spell.name}]. Ative o Combate Automático para recuperar mana.`);
                      return;
                    }

                    // Spend mana and trigger cooldown (instant-cast-rule)
                    setSpellMana(prev => Math.max(0, prev - spell.manaCost));
                    setSpellCooldowns(prev => ({ ...prev, [spell.id]: spell.cooldown }));

                    // Calculate proficiency factor
                    const isMagic = ["Fire", "Ice", "Shadow", "Nature"].includes(spell.elemental);
                    const isHoly = spell.elemental === "Holy" || spell.name.includes("Heal");
                    
                    const proficiency = isHoly
                      ? skillHealingProficiency
                      : isMagic
                      ? skillMagicProficiency
                      : skillSwordProficiency;

                    const classCoef = activeProfile.classCoef;
                    const finalPower = Math.round((spell.basePower + proficiency * 2.5) * classCoef * scalingEquipMultiplier * scalingElementalMod);

                    // Skill progression through usage (skill-progression-rule)
                    if (isHoly) {
                      setSkillHealingProficiency(prev => Math.min(100, prev + 1));
                    } else if (isMagic) {
                      setSkillMagicProficiency(prev => Math.min(100, prev + 1));
                    } else {
                      setSkillSwordProficiency(prev => Math.min(100, prev + 1));
                    }

                    if (spell.name.includes("Heal")) {
                      setSpellHp(prev => Math.min(maxHpCalculated, prev + finalPower));
                      setSpellSkillConsoleLog(`[Instant Cast] Sucesso! Curou +${finalPower} HP usando [${spell.name}]. Custo: ${spell.manaCost} MP. Cooldown: ${spell.cooldown}s (instant-cast-rule).`);
                      addBibleLog("success", `[Conjurador] Sucesso! ${spell.name} restaurou ${finalPower} HP. Proficiência em Cura aumentada.`);
                    } else {
                      // Attack triggers hit mana regeneration (mana-regeneration-rule)
                      setSpellMana(prev => Math.min(maxManaCalculated, prev + 6));
                      setSpellSkillConsoleLog(`[Instant Cast] Sucesso! Desferiu [${spell.name}] causando ${finalPower} de dano elemental (${spell.elemental}). Custo: ${spell.manaCost} MP. (+6 MP regenerados no impacto)`);
                      addBibleLog("success", `[Conjurador] Sucesso! ${spell.name} causou ${finalPower} de dano (${spell.elemental}). Proficiência de habilidade aumentada.`);
                    }
                  };

                  const triggerBasicAutoAttack = () => {
                    // Basic attack is automatic, doesn't consume mana, and generates mana
                    const basicPower = 10 + Math.round(skillSwordProficiency * 1.5);
                    setSpellMana(prev => Math.min(maxManaCalculated, prev + 10)); // Generates 10 mana (mana-regeneration-rule)
                    setSkillSwordProficiency(prev => Math.min(100, prev + 1)); // usage increase
                    setSpellSkillConsoleLog(`[Auto Attack] Desferido golpe físico automático de curto alcance: causou ${basicPower} de dano. Regenerou +10 MP devido à atividade física de combate (auto-attack-rule & mana-regeneration-rule).`);
                    addBibleLog("info", `[Combate Ativo] Auto Attack desferido com sucesso. +10 de Mana regenerado de forma combativa.`);
                  };

                  return (
                    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                      {/* Left Column: Combat Resource & Tracker */}
                      <div className="lg:col-span-5 space-y-4">
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                          <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                            <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                              <Activity className="w-4 h-4 text-rose-400" />
                              Monitor de Recursos do Conjurador
                            </span>
                            <span className="text-[10px] font-mono text-slate-500 uppercase font-bold">Combat Active</span>
                          </div>

                          {/* HP Bar */}
                          <div className="space-y-1">
                            <div className="flex justify-between items-center text-xs">
                              <span className="text-slate-400 font-bold flex items-center gap-1">
                                <span className="w-2 h-2 rounded-full bg-emerald-500" />
                                Pontos de Vida (HP)
                              </span>
                              <span className="font-mono text-emerald-400 font-bold">{Math.min(spellHp, maxHpCalculated)} / {maxHpCalculated}</span>
                            </div>
                            <div className="w-full bg-slate-900 h-2.5 rounded-full overflow-hidden border border-slate-800">
                              <div
                                className="bg-gradient-to-r from-emerald-500 to-teal-400 h-full transition-all duration-300"
                                style={{ width: `${Math.min(100, (spellHp / maxHpCalculated) * 100)}%` }}
                              />
                            </div>
                          </div>

                          {/* Mana Bar */}
                          <div className="space-y-1">
                            <div className="flex justify-between items-center text-xs">
                              <span className="text-slate-400 font-bold flex items-center gap-1">
                                <span className="w-2 h-2 rounded-full bg-sky-500 animate-pulse" />
                                Mana de Combate (MP)
                              </span>
                              <span className="font-mono text-sky-400 font-bold">{Math.min(spellMana, maxManaCalculated)} / {maxManaCalculated}</span>
                            </div>
                            <div className="w-full bg-slate-900 h-2.5 rounded-full overflow-hidden border border-slate-800">
                              <div
                                className="bg-gradient-to-r from-sky-500 via-blue-500 to-indigo-500 h-full transition-all duration-300"
                                style={{ width: `${Math.min(100, (spellMana / maxManaCalculated) * 100)}%` }}
                              />
                            </div>
                            <span className="text-[9px] text-slate-500 font-mono italic block">
                              Regeneração passiva desabilitada. Ataque ou acerte magias para gerar Mana! (mana-regeneration-rule)
                            </span>
                          </div>

                          {/* Interactive Action Buttons */}
                          <div className="grid grid-cols-2 gap-2.5 pt-2">
                            <button
                              onClick={triggerBasicAutoAttack}
                              className="py-2 px-3 bg-gradient-to-r from-rose-950/40 to-rose-900/30 hover:from-rose-900/50 hover:to-rose-850/40 text-rose-300 border border-rose-800/60 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 cursor-pointer"
                            >
                              <Sword className="w-3.5 h-3.5 text-rose-400" />
                              Ataque Automático
                            </button>

                            <button
                              onClick={() => {
                                setSpellHp(maxHpCalculated);
                                setSpellMana(maxManaCalculated);
                                setSpellSkillConsoleLog("Recursos restaurados! HP e Mana regenerados de forma segura na Zona Neutra de Ironhold.");
                                addBibleLog("info", "[Conjurador] Recursos de vida e mana restaurados para o patamar máximo.");
                              }}
                              className="py-2 px-3 bg-slate-900 hover:bg-slate-850 text-slate-300 border border-slate-800 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 cursor-pointer"
                            >
                              <RefreshCw className="w-3.5 h-3.5 text-slate-400" />
                              Zerar Recursos
                            </button>
                          </div>
                        </div>

                        {/* Skill Proficiency Progression Tracker */}
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                          <div className="flex items-center justify-between border-b border-slate-900 pb-2">
                            <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                              <Database className="w-4 h-4 text-purple-400" />
                              Proficiências e Maestria de Combate
                            </span>
                            <span className="text-[10px] font-mono font-bold text-purple-400 bg-purple-950/40 border border-purple-900 px-1.5 rounded">
                              Progression-by-Use
                            </span>
                          </div>

                          <div className="space-y-3">
                            {/* Sword Usage */}
                            <div className="space-y-1">
                              <div className="flex justify-between items-center text-[11px]">
                                <span className="text-slate-400 font-medium">Uso de Armas Físicas (Sword, Bow, Axe)</span>
                                <span className="font-mono font-bold text-amber-400">{skillSwordProficiency} / 100</span>
                              </div>
                              <div className="w-full bg-slate-900 h-2 rounded-full overflow-hidden">
                                <div
                                  className="bg-amber-400 h-full transition-all"
                                  style={{ width: `${skillSwordProficiency}%` }}
                                />
                              </div>
                            </div>

                            {/* Magic Power */}
                            <div className="space-y-1">
                              <div className="flex justify-between items-center text-[11px]">
                                <span className="text-slate-400 font-medium">Uso de Magia de Ataque (Fire, Ice, Shadow)</span>
                                <span className="font-mono font-bold text-sky-400">{skillMagicProficiency} / 100</span>
                              </div>
                              <div className="w-full bg-slate-900 h-2 rounded-full overflow-hidden">
                                <div
                                  className="bg-sky-400 h-full transition-all"
                                  style={{ width: `${skillMagicProficiency}%` }}
                                />
                              </div>
                            </div>

                            {/* Holy / Healing */}
                            <div className="space-y-1">
                              <div className="flex justify-between items-center text-[11px]">
                                <span className="text-slate-400 font-medium">Domínio Sagrado e Curas (Holy Magic)</span>
                                <span className="font-mono font-bold text-emerald-400">{skillHealingProficiency} / 100</span>
                              </div>
                              <div className="w-full bg-slate-900 h-2 rounded-full overflow-hidden">
                                <div
                                  className="bg-emerald-400 h-full transition-all"
                                  style={{ width: `${skillHealingProficiency}%` }}
                                />
                              </div>
                            </div>
                          </div>

                          <div className="grid grid-cols-3 gap-1.5 pt-1">
                            <button
                              onClick={() => {
                                setSkillSwordProficiency(prev => Math.min(100, prev + 5));
                                setSpellSkillConsoleLog("Treinou com o espantalho de ferro: proficiência com armas físicas aumentada em +5.");
                                addBibleLog("success", "[Treinador] Você exercitou seus golpes físicos com sucesso!");
                              }}
                              className="py-1 px-1 bg-slate-900 hover:bg-slate-800 text-[10px] text-slate-400 rounded border border-slate-800 font-medium text-center cursor-pointer"
                            >
                              +5 Física
                            </button>
                            <button
                              onClick={() => {
                                setSkillMagicProficiency(prev => Math.min(100, prev + 5));
                                setSpellSkillConsoleLog("Canalizou correntes elementais puras: proficiência em magias ofensivas aumentada em +5.");
                                addBibleLog("success", "[Treinador] Suas canalizações elementais aumentaram a afinidade mágica.");
                              }}
                              className="py-1 px-1 bg-slate-900 hover:bg-slate-800 text-[10px] text-slate-400 rounded border border-slate-800 font-medium text-center cursor-pointer"
                            >
                              +5 Magia
                            </button>
                            <button
                              onClick={() => {
                                setSkillHealingProficiency(prev => Math.min(100, prev + 5));
                                setSpellSkillConsoleLog("Entoou cânticos de restauração no santuário: proficiência de cura aumentada em +5.");
                                addBibleLog("success", "[Treinador] Suas preces de reabilitação ficaram mais próximas da perfeição.");
                              }}
                              className="py-1 px-1 bg-slate-900 hover:bg-slate-800 text-[10px] text-slate-400 rounded border border-slate-800 font-medium text-center cursor-pointer"
                            >
                              +5 Cura
                            </button>
                          </div>
                        </div>
                      </div>

                      {/* Right Column: Spell Deck & Live Scaling Calculator */}
                      <div className="lg:col-span-7 space-y-4">
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                          <div className="flex items-center justify-between border-b border-slate-900 pb-2.5">
                            <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center gap-1.5">
                              <Sparkles className="w-4 h-4 text-amber-400" />
                              Conjuração de Magias & Cooldowns de Classe
                            </span>
                            <span className="text-[10px] font-mono text-emerald-400 bg-emerald-950/40 border border-emerald-900 px-1.5 rounded uppercase font-bold">
                              {selectedVocationClass} Deck
                            </span>
                          </div>

                          {/* Interactive Spell Buttons */}
                          <div className="space-y-3">
                            <span className="text-[10px] text-slate-500 font-mono uppercase font-bold tracking-wider block">
                              Habilidades Disponíveis (instant-cast-rule & spell-cooldown-rule):
                            </span>
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                              {currentSpells.map((spell) => {
                                const cdRemaining = spellCooldowns[spell.id] || 0;
                                const isCdActive = cdRemaining > 0;
                                return (
                                  <button
                                    key={spell.id}
                                    onClick={() => handleCastSpell(spell)}
                                    className={`p-3.5 rounded-xl border text-left flex flex-col justify-between transition-all relative overflow-hidden cursor-pointer ${
                                      isCdActive
                                        ? "bg-slate-950 border-rose-950/80 cursor-not-allowed text-slate-500"
                                        : "bg-slate-900/80 hover:bg-slate-850/90 border-slate-800 text-slate-200 hover:border-slate-700"
                                    }`}
                                  >
                                    {isCdActive && (
                                      <div
                                        className="absolute bottom-0 left-0 bg-rose-500/10 h-1 transition-all"
                                        style={{ width: `${(cdRemaining / spell.cooldown) * 100}%` }}
                                      />
                                    )}
                                    
                                    <div className="flex justify-between items-start w-full">
                                      <div>
                                        <span className="font-extrabold text-[12.5px] tracking-tight block">
                                          {spell.name}
                                        </span>
                                        <p className="text-[9.5px] text-slate-400 leading-tight mt-1 line-clamp-2">
                                          {spell.desc}
                                        </p>
                                      </div>
                                      
                                      <div className="text-right shrink-0 ml-1">
                                        <span className="text-[9px] font-mono bg-slate-950 border border-slate-800 text-sky-400 px-1.5 py-0.5 rounded font-bold block">
                                          {spell.manaCost} MP
                                        </span>
                                        <span className="text-[8px] text-slate-500 font-mono block mt-1">
                                          CD: {spell.cooldown}s
                                        </span>
                                      </div>
                                    </div>

                                    {/* Action Footnote / CD */}
                                    <div className="flex items-center justify-between mt-2.5 pt-2 border-t border-slate-850 w-full text-[9.5px]">
                                      <span className="font-mono text-slate-500 uppercase font-bold flex items-center gap-1">
                                        {spell.elemental === "Fire" && <Flame className="w-3 h-3 text-rose-500" />}
                                        {spell.elemental === "Ice" && <Snowflake className="w-3 h-3 text-sky-400" />}
                                        {spell.elemental === "Shadow" && <Moon className="w-3 h-3 text-purple-400" />}
                                        {spell.elemental === "Nature" && <Leaf className="w-3 h-3 text-emerald-400" />}
                                        {spell.elemental === "Holy" && <Sparkles className="w-3 h-3 text-yellow-400" />}
                                        {spell.elemental === "Physical" && <Sword className="w-3 h-3 text-amber-500" />}
                                        {spell.elemental}
                                      </span>

                                      {isCdActive ? (
                                        <span className="font-mono text-rose-400 font-extrabold">Recarga: {cdRemaining}s</span>
                                      ) : (
                                        <span className="text-emerald-400 font-bold uppercase tracking-wider text-[8px] bg-emerald-950/40 border border-emerald-900 px-1.5 rounded">Pronto</span>
                                      )}
                                    </div>
                                  </button>
                                );
                              })}
                            </div>
                          </div>

                          {/* Multi-layered Scaling Visualizer Calculator */}
                          <div className="border-t border-slate-900 pt-4 space-y-3">
                            <span className="text-[10px] text-slate-400 uppercase font-bold tracking-wider block">
                              Mecanismo de Escalonamento de Dano Dinâmico (skill-scaling-rule):
                            </span>

                            <div className="grid grid-cols-2 gap-3.5 bg-slate-900/60 p-4 rounded-xl border border-slate-850 font-mono text-[11px] text-slate-300">
                              <div className="space-y-2">
                                <label className="text-[9px] text-slate-500 uppercase block font-bold">Multiplicador de Equipamento:</label>
                                <select
                                  value={scalingEquipMultiplier}
                                  onChange={(e) => setScalingEquipMultiplier(parseFloat(e.target.value))}
                                  className="w-full bg-slate-950 border border-slate-800 text-amber-400 text-xs rounded p-1.5 focus:outline-none"
                                >
                                  <option value="1.0">Arma Comum [1.0x]</option>
                                  <option value="1.15">Espada/Arco de Aço [1.15x]</option>
                                  <option value="1.3">Cajado do Aprendiz [1.3x]</option>
                                  <option value="1.6">Cajado Divino Lendário [1.6x]</option>
                                </select>
                              </div>

                              <div className="space-y-2">
                                <label className="text-[9px] text-slate-500 uppercase block font-bold">Afinidade Elemental do Alvo:</label>
                                <select
                                  value={scalingElementalMod}
                                  onChange={(e) => setScalingElementalMod(parseFloat(e.target.value))}
                                  className="w-full bg-slate-950 border border-slate-800 text-amber-400 text-xs rounded p-1.5 focus:outline-none"
                                >
                                  <option value="1.0">Neutro [1.0x]</option>
                                  <option value="1.5">Vulnerável (+50%) [1.5x]</option>
                                  <option value="0.5">Resistente (-50%) [0.5x]</option>
                                </select>
                              </div>

                              {/* Show live calculations for selected class spells */}
                              <div className="col-span-2 bg-slate-950/80 p-3 rounded-lg border border-slate-900 space-y-1 text-xs">
                                <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider block border-b border-slate-900 pb-1 mb-1">
                                  Visualizador de Fórmulas ao Vivo:
                                </span>
                                <div className="space-y-1">
                                  {currentSpells.map((sp) => {
                                    const isMagic = ["Fire", "Ice", "Shadow", "Nature"].includes(sp.elemental);
                                    const isHoly = sp.elemental === "Holy" || sp.name.includes("Heal");
                                    const prof = isHoly ? skillHealingProficiency : isMagic ? skillMagicProficiency : skillSwordProficiency;
                                    
                                    const classCoef = activeProfile.classCoef;
                                    const finalResult = Math.round((sp.basePower + prof * 2.5) * classCoef * scalingEquipMultiplier * scalingElementalMod);

                                    return (
                                      <div key={sp.id} className="flex justify-between items-center text-[10.5px]">
                                        <span className="text-slate-400 font-medium">{sp.name}:</span>
                                        <span className="text-slate-300">
                                          ({sp.basePower} base + {prof} prof * 2.5) * {classCoef} class * {scalingEquipMultiplier} gear * {scalingElementalMod} ele = <strong className="text-amber-400">{finalResult}</strong>
                                        </span>
                                      </div>
                                    );
                                  })}
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>

                        {/* Spell Simulator Logs */}
                        <div className="bg-slate-950/40 border border-slate-850 p-4 rounded-xl">
                          <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Console de Magias e Habilidades do Servidor</span>
                          <div className="mt-1 bg-black/60 font-mono text-[10.5px] p-2.5 rounded border border-slate-900 text-slate-300 leading-normal">
                            <span className="text-amber-500 font-bold">&gt;</span> {spellSkillConsoleLog}
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })()}
              </div>

              {/* ================================================================== */}
              {/* WORLD CONTENT GENERATION & EXPANSION SIMULATOR */}
              {/* ================================================================== */}
              <div className="bg-slate-900/40 border border-slate-850 rounded-2xl p-6 space-y-6 mt-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between border-b border-slate-800/80 pb-4 gap-2">
                  <div>
                    <span className="text-[10px] font-bold text-amber-400 font-mono tracking-widest uppercase block">
                      Content Generation & World Expansion Framework (Canonical)
                    </span>
                    <h3 className="font-extrabold text-base tracking-tight text-slate-100 flex items-center gap-2">
                      <Globe className="w-5 h-5 text-emerald-400 animate-pulse" />
                      Simulador de Geração de Conteúdo e Expansão do Mundo
                    </h3>
                  </div>
                  <div className="bg-slate-950/60 border border-slate-800/40 px-3 py-1.5 rounded-lg text-right">
                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Foco Canônico</span>
                    <span className="text-xs font-mono font-bold text-emerald-400">Activity & Risk Driven</span>
                  </div>
                </div>

                {(() => {
                  // Biomes definition
                  const biomes = [
                    { id: "barren_desert", name: "Deserto Estéril", density: 1.2, bias: "FireProtectionGear", element: "Fire", desc: "Clima extremo de alta densidade." },
                    { id: "frozen_peaks", name: "Picos Congelados", density: 0.8, bias: "IceProtectionGear", element: "Ice", desc: "Clima frio de baixa densidade." },
                    { id: "whispering_swamp", name: "Pântano Sussurrante", density: 1.5, bias: "PoisonResistGear", element: "Nature", desc: "Ambiente tóxico superpovoado." },
                    { id: "ruined_keep", name: "Fortaleza Arruinada", density: 1.0, bias: "WeaponMaterials", element: "Shadow", desc: "Ruínas assombradas com materiais raros." }
                  ];

                  const riskTiers = {
                    Minimal: { hp: 0.8, dmg: 0.8, drop: 0.8, color: "text-emerald-400 border-emerald-950" },
                    Moderate: { hp: 1.0, dmg: 1.0, drop: 1.0, color: "text-sky-400 border-sky-950" },
                    Severe: { hp: 1.5, dmg: 1.4, drop: 1.5, color: "text-amber-400 border-amber-950" },
                    Extreme: { hp: 2.5, dmg: 2.2, drop: 3.0, color: "text-rose-400 border-rose-950" }
                  };

                  const currentBiome = biomes.find(b => b.id === selectedBiome) || biomes[0];
                  const currentRisk = riskTiers[activeRiskTier];

                  // Calculated Dynamic Density based on slider * biome factor
                  const finalCalculatedDensity = Math.round(encounterDensity * currentBiome.density);

                  const handleSimulateSpawn = () => {
                    const families = ["Demon", "Dragon", "Beast", "Undead"];
                    const selectedFamily = families[Math.floor(Math.random() * families.length)];

                    const baseStats: Record<string, { hp: number; dmg: number; name: string }> = {
                      Demon: { hp: 150, dmg: 25, name: "Gorgon Spawn" },
                      Dragon: { hp: 450, dmg: 45, name: "Wyvern Cub" },
                      Beast: { hp: 100, dmg: 15, name: "Ravenshire Wolf" },
                      Undead: { hp: 200, dmg: 20, name: "Crypt Ghoul" }
                    };

                    const b = baseStats[selectedFamily];
                    const finalHP = Math.round(b.hp * currentRisk.hp);
                    const isFireAmp = currentBiome.element === "Fire" && selectedFamily === "Demon";
                    const finalDMG = Math.round(b.dmg * currentRisk.dmg * (isFireAmp ? 1.2 : 1.0));

                    const logMsg = `[SPAWN DINÂMICO] Spawnou ${b.name} (${selectedFamily}) em ${currentBiome.name}. HP Final: ${finalHP} (Base: ${b.hp} * Risco: ${currentRisk.hp}x). Dano Final: ${finalDMG} (Base: ${b.dmg} * Risco: ${currentRisk.dmg}x${isFireAmp ? " * Fire Biome Amp 1.2x" : ""}). Loot bias ativo: [${currentBiome.bias}] (Drop x${currentRisk.drop}).`;

                    setWorldContentLogs(prev => [logMsg, ...prev.slice(0, 19)]);
                    addBibleLog("success", `[Spawn Canônico] Gerado ${b.name} (${selectedFamily}) com ${finalHP} HP.`);
                  };

                  const handleSimulateEvent = () => {
                    const biomeEvents: Record<string, { boss: string; family: string; condition: string }> = {
                      barren_desert: { boss: "Gorgoroth, o Incinerador", family: "Demon", condition: "Ruptura Infernal" },
                      frozen_peaks: { boss: "Vermidrax, o Sopro Gélido", family: "Dragon", condition: "Tempestade Ártica" },
                      whispering_swamp: { boss: "Miasma Pestilento", family: "Beast", condition: "Emanação Tóxica" },
                      ruined_keep: { boss: "Lorde Malakor", family: "Undead", condition: "Eclipse Profano" }
                    };

                    const ev = biomeEvents[selectedBiome];
                    const logMsg = `[EVENTO MUNDIAL] Chefe Global [${ev.boss}] (${ev.family}) manifestou-se devido à trigger [${ev.condition}] no bioma [${currentBiome.name}]! Recomenda-se Risco: Severe/Extreme. Recompensas de loot em área com multiplicador global x${currentRisk.drop}.`;

                    setWorldContentLogs(prev => [logMsg, ...prev.slice(0, 19)]);
                    addBibleLog("warning", `[Evento Canônico] Boss ${ev.boss} apareceu devido a ${ev.condition}!`);
                  };

                  return (
                    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                      {/* Left Column: Biome Selector & Density Sliders */}
                      <div className="lg:col-span-5 space-y-4">
                        {/* Biome Selector */}
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-3.5">
                          <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider block flex items-center gap-1">
                            <Compass className="w-4 h-4 text-emerald-400" />
                            1. Escolha o Bioma Ativo
                          </span>

                          <div className="grid grid-cols-1 gap-2">
                            {biomes.map((b) => {
                              const isSelected = selectedBiome === b.id;
                              return (
                                <button
                                  key={b.id}
                                  onClick={() => {
                                    setSelectedBiome(b.id);
                                    setWorldContentLogs(prev => [`[Bioma Ativo] Alterado para ${b.name}. Densidade base modificada para ${b.density}x. Afinidade elemental: [${b.element}].`, ...prev.slice(0, 19)]);
                                  }}
                                  className={`p-3 rounded-lg border text-left transition flex items-center justify-between cursor-pointer ${
                                    isSelected
                                      ? "bg-emerald-950/40 border-emerald-500 text-emerald-300 ring-1 ring-emerald-500/20"
                                      : "bg-slate-900/60 border-slate-850 text-slate-400 hover:bg-slate-850 hover:text-slate-300"
                                  }`}
                                >
                                  <div>
                                    <span className="font-bold text-xs block">{b.name}</span>
                                    <span className="text-[9px] text-slate-500 block leading-tight">{b.desc}</span>
                                  </div>
                                  <div className="text-right">
                                    <span className="text-[8px] font-mono text-slate-500 uppercase block font-bold">Elemento</span>
                                    <span className="text-[10.5px] font-mono font-bold text-emerald-400">{b.element}</span>
                                  </div>
                                </button>
                              );
                            })}
                          </div>
                        </div>

                        {/* Density Slider */}
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-3">
                          <div className="flex justify-between items-center">
                            <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider block flex items-center gap-1">
                              <Sliders className="w-4 h-4 text-emerald-400" />
                              2. Densidade de Encontros
                            </span>
                            <span className="text-xs font-mono font-bold text-emerald-400">{encounterDensity}%</span>
                          </div>

                          <input
                            type="range"
                            min="10"
                            max="150"
                            value={encounterDensity}
                            onChange={(e) => setEncounterDensity(parseInt(e.target.value))}
                            className="w-full h-1.5 bg-slate-900 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                          />

                          <div className="bg-slate-900/40 p-2.5 rounded border border-slate-850 flex justify-between items-center text-[10.5px] font-mono text-slate-300">
                            <span>Densidade Efetiva:</span>
                            <span className="text-amber-400 font-extrabold">
                              {encounterDensity}% * {currentBiome.density}x (Bioma) = {finalCalculatedDensity}%
                            </span>
                          </div>
                        </div>
                      </div>

                      {/* Right Column: Risk Scaling & Interactive Spawn Action */}
                      <div className="lg:col-span-7 space-y-4">
                        {/* Risk Tier Selection & Overlay */}
                        <div className="bg-slate-950/60 border border-slate-850 rounded-xl p-5 space-y-4">
                          <span className="text-[10px] text-slate-400 font-bold uppercase tracking-wider block flex items-center gap-1">
                            <Activity className="w-4 h-4 text-rose-400" />
                            3. Defina o Tier de Risco do Nó
                          </span>

                          <div className="grid grid-cols-4 gap-2">
                            {(Object.keys(riskTiers) as Array<keyof typeof riskTiers>).map((tier) => {
                              const isSelected = activeRiskTier === tier;
                              const rData = riskTiers[tier];
                              return (
                                <button
                                  key={tier}
                                  onClick={() => {
                                    setActiveRiskTier(tier);
                                    setWorldContentLogs(prev => [`[Risco de Nó] Alterado para ${tier}. HP Mod: ${rData.hp}x, Dano Mod: ${rData.dmg}x.`, ...prev.slice(0, 19)]);
                                  }}
                                  className={`py-2 px-1 rounded-lg text-center font-bold text-[10.5px] transition cursor-pointer border ${
                                    isSelected
                                      ? "bg-slate-900 text-slate-100 border-amber-500 ring-1 ring-amber-500/20"
                                      : "bg-slate-950 text-slate-500 border-slate-850 hover:bg-slate-900 hover:text-slate-400"
                                  }`}
                                >
                                  {tier}
                                </button>
                              );
                            })}
                          </div>

                          {/* Risk Modifiers Grid Overlay */}
                          <div className="grid grid-cols-3 gap-2 bg-slate-900/50 p-3 rounded-lg border border-slate-850 text-center font-mono">
                            <div className="p-1">
                              <span className="text-[8px] text-slate-500 uppercase block font-bold">Multiplicador HP</span>
                              <span className="text-xs font-bold text-slate-200">x{currentRisk.hp}</span>
                            </div>
                            <div className="p-1 border-l border-slate-800">
                              <span className="text-[8px] text-slate-500 uppercase block font-bold">Multiplicador Dano</span>
                              <span className="text-xs font-bold text-slate-200">x{currentRisk.dmg}</span>
                            </div>
                            <div className="p-1 border-l border-slate-800">
                              <span className="text-[8px] text-slate-500 uppercase block font-bold">Loot Drop Rate</span>
                              <span className="text-xs font-bold text-amber-400">x{currentRisk.drop}</span>
                            </div>
                          </div>

                          {/* Trigger Buttons */}
                          <div className="grid grid-cols-2 gap-3 pt-1">
                            <button
                              onClick={handleSimulateSpawn}
                              className="py-2.5 px-4 bg-gradient-to-r from-emerald-950 to-emerald-900 hover:from-emerald-900 hover:to-emerald-850 text-emerald-100 border border-emerald-850 rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 cursor-pointer shadow-md"
                            >
                              <Sword className="w-4 h-4 text-emerald-400" />
                              Simular Spawn Dinâmico
                            </button>

                            <button
                              onClick={handleSimulateEvent}
                              className="py-2.5 px-4 bg-gradient-to-r from-amber-950 to-amber-900 hover:from-amber-900 hover:to-amber-850 text-amber-100 border border-amber-850 rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 cursor-pointer shadow-md"
                            >
                              <Sparkles className="w-4 h-4 text-amber-400" />
                              Desencadear Evento Global
                            </button>
                          </div>
                        </div>

                        {/* Interactive Logs */}
                        <div className="bg-slate-950/60 border border-slate-850 p-4 rounded-xl space-y-2">
                          <div className="flex justify-between items-center border-b border-slate-900 pb-1.5">
                            <span className="text-[9px] text-slate-500 uppercase tracking-wider block font-bold font-mono">Log do Simulador de Expansão de Mundo</span>
                            <button
                              onClick={() => setWorldContentLogs([])}
                              className="text-[9px] font-mono text-rose-400 hover:text-rose-300 transition"
                            >
                              Limpar Console
                            </button>
                          </div>
                          
                          <div className="bg-black/40 border border-slate-900 p-3 rounded font-mono text-[10.5px] text-slate-300 leading-normal max-h-[140px] overflow-y-auto space-y-1.5 scrollbar-thin scrollbar-thumb-slate-800 scrollbar-track-transparent">
                            {worldContentLogs.length === 0 ? (
                              <span className="text-slate-600 block italic">Sem logs registrados. Clique nos botões acima para simular spawns e eventos.</span>
                            ) : (
                              worldContentLogs.map((log, i) => (
                                <div key={i} className="border-b border-slate-950/40 pb-1 last:border-0">
                                  <span className="text-emerald-500 font-extrabold mr-1">&gt;</span>
                                  {log}
                                </div>
                              ))
                            )}
                          </div>
                        </div>

                        {/* World Scale Bible Travel Simulator */}
                        <div className="bg-slate-900/40 border border-slate-850 p-4 rounded-xl space-y-4">
                          <div className="flex items-center gap-2 pb-1.5 border-b border-slate-800">
                            <Compass className="w-4 h-4 text-amber-400 animate-spin-slow" />
                            <span className="text-xs font-bold text-slate-200 uppercase tracking-wider font-mono">
                              Simulador de Escala do Mundo & Viagem (World Scale Bible)
                            </span>
                          </div>

                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                            <div className="space-y-2">
                              <label className="text-[10px] uppercase font-bold text-slate-400 block font-mono">
                                Escolha o Continente:
                              </label>
                              <select
                                value={selectedScaleContinent}
                                onChange={(e) => setSelectedScaleContinent(e.target.value)}
                                className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2 text-slate-300 font-mono focus:outline-none focus:border-amber-500/50"
                              >
                                <option value="main_continent">Main Continent (Balanced Temperate)</option>
                                <option value="fire_continent">Fire Continent (Southern - Hot)</option>
                                <option value="ice_continent">Ice Continent (Northern - Frozen)</option>
                                <option value="holy_continent">Holy Continent (Sacred Plains)</option>
                                <option value="shadow_continent">Shadow Continent (Dark/Dual Capitals)</option>
                                <option value="nature_continent">Nature Continent (Forest Spires)</option>
                                <option value="abyssia">Abyssia (Void/Endgame Island)</option>
                              </select>
                            </div>

                            <div className="space-y-2">
                              <label className="text-[10px] uppercase font-bold text-slate-400 block font-mono">
                                Tipo de Locomoção / Montaria:
                              </label>
                              <select
                                value={selectedMountMultiplier}
                                onChange={(e) => setSelectedMountMultiplier(parseFloat(e.target.value))}
                                className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2 text-slate-300 font-mono focus:outline-none focus:border-amber-500/50"
                              >
                                <option value="1.0">A Pé (Velocidade Base: 100%)</option>
                                <option value="1.25">Montaria Terrestre Básica (Velocidade: +125%)</option>
                                <option value="1.40">Montaria Terrestre Épica (Velocidade: +140%)</option>
                              </select>
                            </div>
                          </div>

                          {/* Calculation Output Card */}
                          {(() => {
                            const isAbyssia = selectedScaleContinent === "abyssia";
                            const widthTiles = isAbyssia ? 16000 : 12000;
                            const heightTiles = isAbyssia ? 16000 : 12000;
                            const totalArea = isAbyssia ? 256 : 144;
                            const explorableArea = isAbyssia ? 230.4 : 129.6;
                            const inaccessibleArea = isAbyssia ? 25.6 : 14.4;
                            
                            // Base crossing time on foot:
                            // Standard: ~52.5 mins (middle of 45-60 mins range)
                            // Abyssia: ~80.0 mins (middle of 70-90 mins range)
                            const baseCrossingMinutes = isAbyssia ? 80.0 : 52.5;
                            const currentTravelTimeMinutes = baseCrossingMinutes / selectedMountMultiplier;
                            
                            const getClimateVibe = (id: string) => {
                              if (id === "ice_continent") return "❄️ Extremo Norte: Inverno eterno e primordial. Neve natural.";
                              if (id === "fire_continent") return "🔥 Extremo Sul: Rios de magma e calor do vulcão primordial.";
                              if (id === "shadow_continent") return "💀 Leste Sombrio: Império aristocrático com Duas Capitais (Noctharyn & Necrathis).";
                              if (id === "holy_continent") return "✨ Sagrado: Planícies douradas e o Cristal de Luz primordial.";
                              return "🌍 Distribuição Livre: Clima local característico e sandbox aberto.";
                            };

                            const getSettlementList = (id: string) => {
                              switch (id) {
                                case "main_continent":
                                  return [
                                    { name: "Ironhold Bastion", scale: "Small Settlement", size: "400-700 tiles", time: "2-5 min" },
                                    { name: "Stone Tirith", scale: "Large City", size: "1600-2500 tiles", time: "10-18 min" },
                                    { name: "Blackwater Bay", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Ravenshire", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Thornwall", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" }
                                  ];
                                case "fire_continent":
                                  return [
                                    { name: "Pyra Magnus", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Crimson Hollow", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Molten Anvil", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Primordial Volcano", scale: "NOT YET CANONIZED", size: "-", time: "-" }
                                  ];
                                case "ice_continent":
                                  return [
                                    { name: "Elarisheim", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Frosthaven", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Khaz Tirith", scale: "NOT YET CANONIZED", size: "-", time: "-" },
                                    { name: "Ymirr's Hidden Cavern", scale: "Large City", size: "1600-2500 tiles", time: "10-18 min" }
                                  ];
                                case "holy_continent":
                                  return [
                                    { name: "Luminaar", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Lunareth", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Sunwall", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" }
                                  ];
                                case "shadow_continent":
                                  return [
                                    { name: "Noctharyn", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Grimharbor", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Kar'goth", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Necrathis", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Vel'Sharum", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" }
                                  ];
                                case "nature_continent":
                                  return [
                                    { name: "Elarin", scale: "Capital City", size: "2500-4000 tiles", time: "15-30 min" },
                                    { name: "Sylvaris", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Oakenspire", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" },
                                    { name: "Grunhold", scale: "Medium City", size: "800-1500 tiles", time: "5-10 min" }
                                  ];
                                case "abyssia":
                                  return [
                                    { name: "Last Bastion", scale: "Small Settlement", size: "400-700 tiles", time: "2-5 min" }
                                  ];
                                default:
                                  return [];
                              }
                            };

                            return (
                              <div className="bg-slate-950/80 p-3 rounded-lg border border-slate-850 space-y-3">
                                {/* Grid of Key Metrics */}
                                <div className="grid grid-cols-2 gap-2 text-center font-mono">
                                  <div className="p-2 bg-slate-900/60 rounded border border-slate-900">
                                    <span className="text-[8px] text-slate-500 uppercase block font-bold">Travessia Total Estimada</span>
                                    <span className="text-sm font-bold text-amber-400">
                                      {currentTravelTimeMinutes.toFixed(1)} min
                                    </span>
                                  </div>
                                  <div className="p-2 bg-slate-900/60 rounded border border-slate-900">
                                    <span className="text-[8px] text-slate-500 uppercase block font-bold">Bônus de Viagem</span>
                                    <span className="text-sm font-bold text-violet-400">
                                      {((selectedMountMultiplier - 1.0) * 100).toFixed(0)}% extra
                                    </span>
                                  </div>
                                </div>

                                {/* Physical Dimensions Card */}
                                <div className="p-2 bg-slate-900/30 rounded border border-slate-900 space-y-1.5 text-[10px] font-mono">
                                  <div className="flex justify-between border-b border-slate-800/60 pb-1 text-slate-400">
                                    <span>Dimensões Físicas:</span>
                                    <span className="text-amber-500 font-bold">{widthTiles.toLocaleString()} x {heightTiles.toLocaleString()} tiles ({totalArea} km²)</span>
                                  </div>
                                  <div className="flex justify-between text-slate-400">
                                    <span>Área Explorável (90%):</span>
                                    <span className="text-emerald-400">{explorableArea.toFixed(1)} km²</span>
                                  </div>
                                  <div className="flex justify-between text-slate-400">
                                    <span>Barreiras / Obstáculos (10%):</span>
                                    <span className="text-rose-400">{inaccessibleArea.toFixed(1)} km²</span>
                                  </div>
                                </div>

                                {/* Settlements Scale Audit */}
                                <div className="space-y-1.5">
                                  <span className="text-[8.5px] uppercase font-bold text-slate-500 block font-mono">Escala de Assentamentos Ativos:</span>
                                  <div className="max-h-[140px] overflow-y-auto space-y-1 pr-1 scrollbar-thin scrollbar-thumb-slate-800 scrollbar-track-transparent">
                                    {getSettlementList(selectedScaleContinent).map((settlement, i) => (
                                      <div key={i} className="flex flex-col sm:flex-row justify-between bg-slate-900/40 border border-slate-900 p-1.5 rounded text-[10px] font-mono gap-1">
                                        <div className="flex items-center gap-1">
                                          <span className="text-amber-500">🏰</span>
                                          <span className="text-slate-300 font-bold">{settlement.name}</span>
                                        </div>
                                        <div className="flex items-center gap-1 text-[9px] justify-between sm:justify-start">
                                          <span className={`px-1 py-0.5 rounded text-[8px] font-extrabold uppercase ${
                                            settlement.scale === "Capital City" ? "bg-amber-500/20 text-amber-300 border border-amber-500/30" :
                                            settlement.scale === "Large City" ? "bg-violet-500/20 text-violet-300 border border-violet-500/30" :
                                            settlement.scale === "Medium City" ? "bg-blue-500/20 text-blue-300 border border-blue-500/30" :
                                            settlement.scale === "Small Settlement" ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30" :
                                            "bg-slate-500/20 text-slate-400 border border-slate-500/30"
                                          }`}>
                                            {settlement.scale}
                                          </span>
                                          {settlement.size !== "-" && (
                                            <span className="text-slate-400 text-[9px]">
                                              ({settlement.size} | ⏱️ {settlement.time})
                                            </span>
                                          )}
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                </div>

                                <div className="space-y-1 text-[10.5px] font-mono leading-relaxed text-slate-400 pt-1.5 border-t border-slate-900">
                                  <div className="flex items-start gap-1">
                                    <span className="text-amber-500">📍</span>
                                    <span>
                                      <strong>Diretriz Climática:</strong> {getClimateVibe(selectedScaleContinent)}
                                    </span>
                                  </div>
                                  <div className="flex items-start gap-1">
                                    <span className="text-violet-500">🏇</span>
                                    <span>
                                      <strong>Propriedade de Montaria:</strong> Fornece bônus exclusivo de viagem. Zero atributos de combate, resistências ou habilidades concedidas (<code className="text-slate-300">mount-mobility-rule</code>).
                                    </span>
                                  </div>
                                  <div className="flex items-start gap-1">
                                    <span className="text-blue-500">📐</span>
                                    <span>
                                      <strong>Coordenadas do Sistema:</strong> As coordenadas representam âncoras lógicas de posicionamento e streaming, não a dimensão física literal (<code className="text-slate-300">logical-coordinate-rule</code>).
                                    </span>
                                  </div>
                                </div>
                              </div>
                            );
                          })()}
                        </div>

                        {/* Canonical Rules Checks */}
                        <div className="bg-slate-900/30 border border-slate-850 p-4 rounded-xl space-y-2">
                          <span className="text-[9px] text-slate-400 font-bold uppercase tracking-wider block">Verificadores de Diretrizes de Conteúdo do Mundo:</span>
                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[10px] font-mono">
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">content-core-principle</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">encounter-generation-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">bestiary-content-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">spawn-system-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">boss-event-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">biome-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">content-scaling-rule & dungeon-framework-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">mount-mobility-rule (125-140%)</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">continent-traversal-rule (45-60 min)</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">logical-coordinate-rule</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <CheckCircle className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              <span className="text-slate-300">world-layout-rule (Ice N / Fire S)</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })()}
              </div>


            </motion.div>
          )}
          
          {/* Tab: World Foundation Simulator */}
          {activeTab === "world" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="w-full text-slate-100"
            >
              <Suspense fallback={<div className="bg-slate-900 border border-slate-800/80 p-6 rounded-2xl text-center text-xs text-slate-400 font-mono animate-pulse">Carregando World Simulator...</div>}>
                <WorldSimulator />
              </Suspense>
            </motion.div>
          )}

          {/* Tab: Progression Bible & Combat Simulator */}
          {activeTab === "progression" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100"
            >
              <ProgressionSimulator />
            </motion.div>
          )}

          {/* Tab: Vocation / Class Bible Simulator */}
          {activeTab === "vocation" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100"
            >
              <Suspense fallback={<div className="bg-slate-900 border border-slate-800/80 p-6 rounded-2xl text-center text-xs text-slate-400 font-mono animate-pulse">Carregando Vocation Simulator...</div>}>
                <VocationSimulator />
              </Suspense>
            </motion.div>
          )}

          {/* Tab: Spells & Skills Bible Simulator */}
          {activeTab === "spells" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100"
            >
              <Suspense fallback={<div className="bg-slate-900 border border-slate-800/80 p-6 rounded-2xl text-center text-xs text-slate-400 font-mono animate-pulse">Carregando Spell Simulator...</div>}>
                <SpellSkillSimulator />
              </Suspense>
            </motion.div>
          )}

          {/* Tab: Itemization Bible & Build Simulator */}
          {activeTab === "itemization" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100"
            >
              <ItemizationSimulator />
            </motion.div>
          )}

          {/* Tab 4: Clean Architecture Explanation */}
          {activeTab === "arch" && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="grid grid-cols-1 lg:grid-cols-12 gap-6"
            >
              <div className="lg:col-span-12 bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl space-y-6">
                <div>
                  <h3 className="font-bold text-lg tracking-tight bg-gradient-to-r from-amber-400 to-violet-400 bg-clip-text text-transparent mb-2">
                    Clean Architecture & State Patterns no MMORPG Light and Shadow
                  </h3>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    A arquitetura foi minuciosamente desenhada para suportar a complexidade inerente de um MMORPG moderno, focado em alta escalabilidade, desacoplamento estrito e facilidade de manutenção.
                  </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="bg-slate-950/80 p-5 border border-slate-800/80 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-amber-400 font-bold text-xs uppercase tracking-wider">
                      <Layers className="w-4 h-4" />
                      1. Desacoplamento Estrito
                    </div>
                    <p className="text-xs text-slate-400 leading-relaxed">
                      Cenas de UI (como o Menu Principal) nunca alteram estados do jogo ou se comunicam diretamente com a rede. Em vez disso, elas publicam eventos genéricos e tipados no <code className="text-amber-300 font-mono">EventBus.cs</code> (ex: <code className="text-slate-200 font-mono">OnLoginAttempted</code>). Os controladores de estado interceptam essa mensagem e comutam a lógica com segurança.
                    </p>
                  </div>

                  <div className="bg-slate-950/80 p-5 border border-slate-800/80 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-violet-400 font-bold text-xs uppercase tracking-wider">
                      <Activity className="w-4 h-4" />
                      2. State Machine Baseada em Interfaces
                    </div>
                    <p className="text-xs text-slate-400 leading-relaxed">
                      Cada um dos 8 estados do ciclo de vida global é um objeto isolado herdando da interface <code className="text-violet-300 font-mono">IAppState.cs</code>. Os métodos <code className="text-slate-200 font-mono">EnterAsync()</code> e <code className="text-slate-200 font-mono">ExitAsync()</code> garantem carregamentos assíncronos não bloqueantes e limpezas adequadas de buffers e listeners da rede.
                    </p>
                  </div>

                  <div className="bg-slate-950/80 p-5 border border-slate-800/80 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-emerald-400 font-bold text-xs uppercase tracking-wider">
                      <Database className="w-4 h-4" />
                      3. Tratamento de Rede Orientado a Eventos
                    </div>
                    <p className="text-xs text-slate-400 leading-relaxed">
                      O <code className="text-emerald-300 font-mono">NetworkManager.cs</code> opera em threads assíncronas em background. Ele escuta pacotes brutos do soquete, decodifica o cabeçalho de comprimento, cria uma instância unificada de <code className="text-slate-200 font-mono">GamePacket.cs</code> e delega o processamento ao EventBus de forma segura, garantindo que o loop visual do Godot não sofra lag ou stuttering.
                    </p>
                  </div>
                </div>

                <div className="bg-slate-950/60 border border-slate-800/80 rounded-xl p-5">
                  <h4 className="text-xs font-bold text-slate-300 mb-2 flex items-center gap-1.5 uppercase">
                    <Info className="w-4 h-4 text-amber-500" />
                    Fluxo Sequencial das Transições de Estado Globais:
                  </h4>
                  <div className="flex flex-col md:flex-row justify-between items-stretch md:items-center gap-2 pt-2">
                    {[
                      { step: "Boot", desc: "Instancia Managers" },
                      { step: "Loading", desc: "Carrega Bootstrap UI" },
                      { step: "Menu", desc: "Painel de Login" },
                      { step: "Connecting", desc: "Abertura TCP Socket" },
                      { step: "CharSelection", desc: "Sincroniza Herói" },
                      { step: "InGame", desc: "Loop Gameplay" },
                      { step: "Disconnected", desc: "Aviso de Queda" },
                      { step: "Shutdown", desc: "Fecha Soquetes / Salva" }
                    ].map((st, i, arr) => (
                      <React.Fragment key={st.step}>
                        <div className="flex-1 bg-slate-900 border border-slate-800/60 p-3 rounded-lg text-center shadow-md">
                          <span className="text-[10px] font-bold font-mono text-amber-400">
                            {st.step}
                          </span>
                          <p className="text-[9px] text-slate-500 mt-1 leading-none">{st.desc}</p>
                        </div>
                        {i < arr.length - 1 && (
                          <ChevronRight className="w-4 h-4 text-slate-700 hidden md:block shrink-0" />
                        )}
                      </React.Fragment>
                    ))}
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

      </main>

      {/* Elegant minimalist footer */}
      <footer className="border-t border-slate-800/80 bg-slate-950 py-4 px-6 flex flex-col sm:flex-row justify-between items-center text-[11px] text-slate-500 gap-2">
        <p className="font-semibold uppercase tracking-wider text-slate-600">
          Light & Shadow MMORPG Client Bootstrap Framework • 2026
        </p>
        <p className="text-slate-600 font-mono">
          Godot 4.2+ • MSBuild • C# 12 • .NET 8.0 LTS
        </p>
      </footer>

    </div>
  );
}
