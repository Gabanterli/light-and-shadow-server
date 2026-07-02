import React, { useState, useEffect, useRef } from "react";
import {
  User,
  Key,
  RefreshCw,
  Play,
  Save,
  Plus,
  Trash,
  LogOut,
  Database,
  Activity,
  Wifi,
  WifiOff,
  Coins,
  Sword,
  Shield,
  Heart,
  Sparkles,
  BookOpen,
  Terminal,
  Download,
  Upload,
  ArrowRight,
  MapPin,
  Compass,
  Gift,
  CheckCircle,
  AlertTriangle,
  Flame,
  Info,
  ShieldAlert,
  Sliders,
  Wand2
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { GameScreen } from "../core/components/GameScreen";
import {
  CharacterProfile,
  SessionState,
  ServerApiLog,
  InventoryItem,
  EquippedItem,
  QuestState
} from "../types";

// Starter Quests configuration
const TUTORIAL_QUEST: QuestState = {
  quest_id: "quest_tutorial_sewer_rats",
  title: "A Infestação nos Esgotos de Ironhold",
  description: "Limpe os esgotos da cidade derrotando 10 Ratos de Esgoto sob as ordens do Capitão Kenneth.",
  objectives: [
    {
      type: "KillMonster",
      target_id: "sewer_rat",
      required_qty: 10,
      current_qty: 0
    }
  ],
  rewards: {
    experience: 110,
    gold: 15,
    items: [
      { item_id: "potion_heal", quantity: 3 }
    ]
  }
};

// Initial default character set if local storage is empty
const INITIAL_CHARACTERS_MOCK: CharacterProfile[] = [
  {
    id: "char-1e24-4a21-98bc-e76bcf129990",
    name: "Gabriela_Paladin",
    level: 1,
    xp: 0,
    vocation_state: "Novice",
    position: { x: 100.0, y: 100.0, z: 0.0, region: "Ironhold Bastion" },
    gold: 250,
    inventory: [
      { id: "bronze_coin", name: "Moeda de Bronze", qty: 250, value: 1, type: "currency" },
      { id: "potion_heal", name: "Poção de Cura", qty: 5, value: 15, type: "consumable" },
      { id: "iron_ore", name: "Minério de Ferro", qty: 3, value: 5, type: "material" }
    ],
    equipment: [
      { slot: "Arma", id: "sword_basic", name: "Espada Básica de Treinamento", value: 100 },
      { slot: "Escudo", id: "shield_wooden", name: "Escudo de Madeira", value: 150 }
    ],
    quest_state: {
      activeQuests: [JSON.parse(JSON.stringify(TUTORIAL_QUEST))],
      completedQuestIds: []
    },
    created_at: new Date(Date.now() - 172800000).toISOString(), // 2 days ago
    last_login: new Date().toISOString()
  }
];

interface GameEntrySimulatorProps {
  // Sync handlers for updating App.tsx states
  syncInventory?: (items: any[]) => void;
  syncEquipment?: (items: any[]) => void;
  syncLevel?: (lvl: number) => void;
  syncClass?: (cls: string) => void;
  syncCoins?: (coins: number) => void;
  syncCharName?: (name: string) => void;
}

export function GameEntrySimulator({
  syncInventory,
  syncEquipment,
  syncLevel,
  syncClass,
  syncCoins,
  syncCharName
}: GameEntrySimulatorProps) {
  // Persistence Core
  const [characters, setCharacters] = useState<CharacterProfile[]>([]);
  const [activeCharId, setActiveCharId] = useState<string>("");
  const [sessionState, setSessionState] = useState<SessionState>("LOGIN");
  const [usernameInput, setUsernameInput] = useState("gabriella_dev");
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  
  // Loading Simulation States
  const [loadingPercent, setLoadingPercent] = useState(0);
  const [loadingTip, setLoadingTip] = useState("");
  
  // New Character Creation Form States
  const [newCharName, setNewCharName] = useState("");
  const [newCharPath, setNewCharPath] = useState<"melee" | "ranged" | "magic">("melee");
  const [creationError, setCreationError] = useState("");

  // Gameplay Simulation States
  const [combatLogs, setCombatLogs] = useState<string[]>(["[Session] Bem-vindo a Ironhold Bastion! Fale com o Capitão Kenneth para iniciar o treinamento."]);
  const [isFightingRat, setIsFightingRat] = useState(false);
  const [activeRatHp, setActiveRatHp] = useState(320);
  const [playerHp, setPlayerHp] = useState(120);
  const [playerMana, setPlayerMana] = useState(40);
  const [autosavePulse, setAutosavePulse] = useState(false);

  // Authoritative Server Logger
  const [apiLogs, setApiLogs] = useState<ServerApiLog[]>([]);
  const consoleEndRef = useRef<HTMLDivElement>(null);

  // Initialize and load characters from LocalStorage (Authoritative persistence layer)
  useEffect(() => {
    const stored = localStorage.getItem("light_shadow_characters");
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        if (parsed && Array.isArray(parsed) && parsed.length > 0) {
          setCharacters(parsed);
        } else {
          setCharacters(INITIAL_CHARACTERS_MOCK);
          localStorage.setItem("light_shadow_characters", JSON.stringify(INITIAL_CHARACTERS_MOCK));
        }
      } catch (e) {
        setCharacters(INITIAL_CHARACTERS_MOCK);
      }
    } else {
      setCharacters(INITIAL_CHARACTERS_MOCK);
      localStorage.setItem("light_shadow_characters", JSON.stringify(INITIAL_CHARACTERS_MOCK));
    }

    addApiLog("GET", "/api/characters", 200, 15, "Sincronizando tabelas locais");
  }, []);

  // Periodic autosave trigger (every 30 seconds)
  useEffect(() => {
    if (sessionState === "IN_WORLD" || sessionState === "TUTORIAL_ACTIVE") {
      const interval = setInterval(() => {
        handleAutosave();
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [sessionState, activeCharId, characters]);

  // Scroll to bottom of API console
  useEffect(() => {
    if (consoleEndRef.current) {
      consoleEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [apiLogs]);

  const activeChar = characters.find(c => c.id === activeCharId);

  // Helper to add API transaction logs
  const addApiLog = (
    method: "GET" | "POST" | "PUT" | "DELETE" | "WEBSOCKET",
    path: string,
    status: number,
    latencyMs: number,
    payload: string
  ) => {
    const newLog: ServerApiLog = {
      id: `log-${Math.random().toString(36).substr(2, 9)}`,
      timestamp: new Date().toLocaleTimeString(),
      method,
      path,
      status,
      latencyMs,
      payload
    };
    setApiLogs(prev => [...prev, newLog]);
  };

  // Autoload/Save Core Logic
  const saveToBackend = (updatedList: CharacterProfile[]) => {
    setCharacters(updatedList);
    localStorage.setItem("light_shadow_characters", JSON.stringify(updatedList));
  };

  // Save current player state explicitly
  const handleAutosave = () => {
    if (!activeCharId) return;
    setAutosavePulse(true);
    setTimeout(() => setAutosavePulse(false), 1000);

    const updated = characters.map(c => {
      if (c.id === activeCharId) {
        return {
          ...c,
          last_login: new Date().toISOString()
        };
      }
      return c;
    });

    saveToBackend(updated);
    addApiLog("PUT", `/api/characters/${activeCharId}/autosave`, 200, 24, JSON.stringify({
      id: activeCharId,
      timestamp: new Date().toISOString(),
      action: "AUTOSAVE_HEARTBEAT"
    }));
  };

  // Login connection handshake simulation
  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    if (!usernameInput.trim()) return;

    setIsConnecting(true);
    addApiLog("POST", "/api/auth/login", 101, 150, JSON.stringify({ username: usernameInput }));

    setTimeout(() => {
      setIsConnecting(false);
      setIsConnected(true);
      setSessionState("CHARACTER_SELECT");
      addApiLog("GET", "/api/auth/session", 200, 42, JSON.stringify({ status: "AUTHENTICATED", username: usernameInput, socket: "CONNECTED_TCP" }));
      
      // Auto-select first character if exists
      if (characters.length > 0) {
        setActiveCharId(characters[0].id);
      }
    }, 1200);
  };

  // Character selection confirmation
  const handleSelectCharacter = (id: string) => {
    setActiveCharId(id);
    addApiLog("GET", `/api/characters/${id}`, 200, 28, `Sincronizando perfil: ${characters.find(c => c.id === id)?.name}`);
  };

  // Character creation
  const handleCreateCharacter = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCharName.trim()) {
      setCreationError("Nome do herói é obrigatório.");
      return;
    }
    if (newCharName.length < 3 || newCharName.length > 20) {
      setCreationError("Nome deve ter entre 3 e 20 caracteres.");
      return;
    }
    if (characters.some(c => c.name.toLowerCase() === newCharName.toLowerCase())) {
      setCreationError("Este nome de personagem já existe no reino.");
      return;
    }

    setCreationError("");
    
    // Choose start kit depending on the vocation path
    let startInventory: InventoryItem[] = [
      { id: "bronze_coin", name: "Moeda de Bronze", qty: 250, value: 1, type: "currency" },
      { id: "potion_heal", name: "Poção de Cura", qty: 5, value: 15, type: "consumable" }
    ];
    let startEquipment: EquippedItem[] = [];

    if (newCharPath === "melee") {
      startEquipment = [
        { slot: "Arma", id: "sword_basic", name: "Espada Básica de Treinamento", value: 100 },
        { slot: "Escudo", id: "shield_wooden", name: "Escudo de Madeira", value: 150 }
      ];
    } else if (newCharPath === "ranged") {
      startEquipment = [
        { slot: "Arma", id: "bow_short", name: "Arco Curto de Treinamento", value: 100 }
      ];
      startInventory.push({ id: "arrow_bundle", name: "Aljava Básica de Treinamento", qty: 100, value: 1, type: "ammo" });
    } else {
      startEquipment = [
        { slot: "Arma", id: "sacred_scepter", name: "Cetro Sagrado de Treinamento", value: 120 }
      ];
      startInventory.push({ id: "spellbook_basic", name: "Livro de Feitiços Básico", qty: 1, value: 100, type: "book" });
    }

    const newChar: CharacterProfile = {
      id: `char-${Math.random().toString(36).substr(2, 9)}-${Math.random().toString(36).substr(2, 9)}`,
      name: newCharName.trim().replace(/\s+/g, "_"),
      level: 1,
      xp: 0,
      vocation_state: "Novice",
      position: { x: 100.0, y: 100.0, z: 0.0, region: "Ironhold Bastion" },
      gold: 250,
      inventory: startInventory,
      equipment: startEquipment,
      quest_state: {
        activeQuests: [JSON.parse(JSON.stringify(TUTORIAL_QUEST))],
        completedQuestIds: []
      },
      created_at: new Date().toISOString(),
      last_login: new Date().toISOString()
    };

    const updatedList = [...characters, newChar];
    saveToBackend(updatedList);
    setActiveCharId(newChar.id);
    setNewCharName("");
    setSessionState("CHARACTER_SELECT");
    
    addApiLog("POST", "/api/characters", 201, 85, JSON.stringify(newChar));
  };

  // Delete character
  const handleDeleteCharacter = (id: string, name: string) => {
    if (!window.confirm(`Tem certeza que deseja apagar permanentemente o herói ${name}?`)) return;

    const filtered = characters.filter(c => c.id !== id);
    saveToBackend(filtered);
    if (activeCharId === id) {
      setActiveCharId(filtered.length > 0 ? filtered[0].id : "");
    }
    addApiLog("DELETE", `/api/characters/${id}`, 200, 36, `Personagem deletado do banco: ${name}`);
  };

  // Enter world sequence (Loading simulator)
  const handleEnterWorld = () => {
    if (!activeChar) return;

    setSessionState("LOADING");
    setLoadingPercent(0);

    const tips = [
      "Mitigação Ativa: Pressione escudo para reduzir dano recebido em 60%.",
      "Combate de alto TTK: Cada monstro possui HP substancial, evite pulls de mobs extras.",
      "Regeneração Natural: HP e Mana regeneram fora de combate e ao descançar.",
      "As masmorras de Ironhold ao sul requerem nível 2 e poções de vida prontas."
    ];
    setLoadingTip(tips[Math.floor(Math.random() * tips.length)]);
    addApiLog("POST", `/api/session/enter-world`, 101, 50, JSON.stringify({ char_id: activeCharId, level: activeChar.level }));

    const interval = setInterval(() => {
      setLoadingPercent(prev => {
        if (prev >= 100) {
          clearInterval(interval);
          
          // Determine if player starts in tutorial or full sandbox
          const isTutorial = activeChar.level === 1 && activeChar.quest_state.activeQuests.some(q => q.quest_id === "quest_tutorial_sewer_rats");
          setSessionState(isTutorial ? "TUTORIAL_ACTIVE" : "IN_WORLD");
          
          // Trigger synchronizations to global app states
          if (syncInventory) syncInventory(activeChar.inventory);
          if (syncEquipment) syncEquipment(activeChar.equipment);
          if (syncLevel) syncLevel(activeChar.level);
          if (syncClass) syncClass(activeChar.vocation_state);
          if (syncCoins) syncCoins(activeChar.gold);
          if (syncCharName) syncCharName(activeChar.name);

          addApiLog("POST", `/api/session/spawn`, 200, 110, JSON.stringify({
            char_id: activeChar.id,
            status: "SPAWNED",
            coordinates: activeChar.position,
            tutorial_active: isTutorial
          }));

          return 100;
        }
        return prev + 20;
      });
    }, 300);
  };

  // Exit world
  const handleExitWorld = () => {
    if (!activeChar) return;

    handleAutosave();
    setSessionState("CHARACTER_SELECT");
    addApiLog("POST", `/api/session/exit`, 200, 75, JSON.stringify({ char_id: activeCharId, final_coordinates: activeChar.position }));
  };

  // Simulation: Fight 1 Sewer Rat (High TTK combat simulator)
  const simulateRatCombat = () => {
    if (!activeChar || isFightingRat) return;

    setIsFightingRat(true);
    setActiveRatHp(320);
    setPlayerHp(120);
    setPlayerMana(40);
    setCombatLogs(prev => [...prev, "⚔️ Um Sewer Rat raivoso surge das águas escuras! Iniciando combate tático..."]);

    let currentRatHp = 320;
    let currentPlayerHp = 120;
    let turn = 1;

    const interval = setInterval(() => {
      if (currentRatHp <= 0 || currentPlayerHp <= 0) {
        clearInterval(interval);
        setIsFightingRat(false);

        if (currentRatHp <= 0) {
          // Success! Update active quest objectives
          setCombatLogs(prev => [...prev, "🏆 Vitória! O Sewer Rat foi abatido. Você coletou Cauda de Rato e moedas de bronze."]);

          // Update characters state (authoritative update, trigger save)
          const updated = characters.map(c => {
            if (c.id === activeCharId) {
              const activeQuests = c.quest_state.activeQuests.map(q => {
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

              // Add loot items to inventory
              let currentInv = [...c.inventory];
              const bronzeIdx = currentInv.findIndex(i => i.id === "bronze_coin");
              const randomBronze = Math.floor(Math.random() * 3) + 1;
              if (bronzeIdx !== -1) {
                currentInv[bronzeIdx] = { ...currentInv[bronzeIdx], qty: currentInv[bronzeIdx].qty + randomBronze };
              }
              
              // 60% chance to get rat tail
              if (Math.random() < 0.6) {
                const tailIdx = currentInv.findIndex(i => i.id === "rat_tail");
                if (tailIdx !== -1) {
                  currentInv[tailIdx] = { ...currentInv[tailIdx], qty: currentInv[tailIdx].qty + 1 };
                } else {
                  currentInv.push({ id: "rat_tail", name: "Cauda de Rato", qty: 1, value: 2, type: "material" });
                }
              }

              const newXp = c.xp + 10;

              return {
                ...c,
                xp: newXp,
                inventory: currentInv,
                gold: c.gold + randomBronze,
                quest_state: {
                  ...c.quest_state,
                  activeQuests
                }
              };
            }
            return c;
          });

          saveToBackend(updated);
          
          // Sync with main app
          const newlyUpdatedChar = updated.find(c => c.id === activeCharId);
          if (newlyUpdatedChar) {
            if (syncInventory) syncInventory(newlyUpdatedChar.inventory);
            if (syncCoins) syncCoins(newlyUpdatedChar.gold);
          }

          addApiLog("PUT", `/api/player/combat-event`, 200, 32, JSON.stringify({
            char_id: activeCharId,
            killed: "sewer_rat",
            xp_gained: 10,
            loot_added: ["bronze_coin", "rat_tail"]
          }));
        } else {
          setCombatLogs(prev => [...prev, "💀 Você foi derrotado! Você acordou no centro de cura espiritual de Ironhold Bastion."]);
        }
        return;
      }

      // Alternate attack and defense turns (high TTK feeling)
      if (turn % 2 === 1) {
        // Player attacks
        const dmg = Math.floor(Math.random() * 15) + 18; // ~2.7 DPS representation
        currentRatHp = Math.max(0, currentRatHp - dmg);
        setActiveRatHp(currentRatHp);
        setCombatLogs(prev => [...prev, `[Turno ${Math.ceil(turn/2)}] Você desferiu um golpe preciso! Cauda no rato sofreu -${dmg} de dano físico. (HP Rato: ${currentRatHp}/320)`]);
      } else {
        // Rat attacks, player blocks
        const ratDmgRaw = 5;
        const blockSuccess = Math.random() < 0.7;
        const netDmg = blockSuccess ? 2 : ratDmgRaw;
        currentPlayerHp = Math.max(0, currentPlayerHp - netDmg);
        setPlayerHp(currentPlayerHp);
        setCombatLogs(prev => [
          ...prev,
          blockSuccess 
            ? `🛡️ [Active Defense] Você levantou seu escudo de treinamento a tempo! Absorveu 60% do dano. Sofreu apenas -${netDmg} de dano.` 
            : `💥 O Sewer Rat mordeu sua bota! Sofreu -${netDmg} de dano físico.`
        ]);
      }

      turn++;
    }, 1000);
  };

  // Simulation: Heal using poção
  const handleUseHealPotion = () => {
    if (!activeChar || isFightingRat) return;

    const potIdx = activeChar.inventory.findIndex(i => i.id === "potion_heal" && i.qty > 0);
    if (potIdx === -1) {
      alert("Você não possui poções de cura em seu inventário!");
      return;
    }

    const updated = characters.map(c => {
      if (c.id === activeCharId) {
        const inv = [...c.inventory];
        if (inv[potIdx].qty > 1) {
          inv[potIdx] = { ...inv[potIdx], qty: inv[potIdx].qty - 1 };
        } else {
          inv.splice(potIdx, 1);
        }
        return { ...c, inventory: inv };
      }
      return c;
    });

    saveToBackend(updated);
    if (syncInventory) syncInventory(updated.find(c => c.id === activeCharId)?.inventory || []);

    setCombatLogs(prev => [...prev, "🧪 Você consumiu uma Poção de Cura. Regenerou +50 de HP instantaneamente!"]);
    addApiLog("POST", `/api/player/use-item`, 200, 18, JSON.stringify({ char_id: activeCharId, item_id: "potion_heal" }));
  };

  // Complete Quest Tutorial and reach Level 2!
  const handleCompleteTutorialQuest = () => {
    if (!activeChar) return;

    const tutorialQuest = activeChar.quest_state.activeQuests.find(q => q.quest_id === "quest_tutorial_sewer_rats");
    if (!tutorialQuest) return;

    const ratObjective = tutorialQuest.objectives.find(obj => obj.type === "KillMonster" && obj.target_id === "sewer_rat");
    if (!ratObjective || ratObjective.current_qty < ratObjective.required_qty) {
      alert("Você deve derrotar 10 Ratos de Esgoto primeiro!");
      return;
    }

    // Complete quest! Reaching Level 2!
    const updated = characters.map(c => {
      if (c.id === activeCharId) {
        const activeQuests = c.quest_state.activeQuests.filter(q => q.quest_id !== "quest_tutorial_sewer_rats");
        const completedQuestIds = [...c.quest_state.completedQuestIds, "quest_tutorial_sewer_rats"];
        
        // Add level rewards
        let inv = [...c.inventory];
        const goldIdx = inv.findIndex(i => i.id === "bronze_coin");
        if (goldIdx !== -1) {
          inv[goldIdx] = { ...inv[goldIdx], qty: inv[goldIdx].qty + 15 };
        }
        
        // Add 3 more potions
        const potIdx = inv.findIndex(i => i.id === "potion_heal");
        if (potIdx !== -1) {
          inv[potIdx] = { ...inv[potIdx], qty: inv[potIdx].qty + 3 };
        } else {
          inv.push({ id: "potion_heal", name: "Poção de Cura", qty: 3, value: 15, type: "consumable" });
        }

        return {
          ...c,
          level: 2,
          xp: 110, // hit level 2
          gold: c.gold + 15,
          inventory: inv,
          quest_state: {
            activeQuests,
            completedQuestIds
          }
        };
      }
      return c;
    });

    saveToBackend(updated);
    setSessionState("IN_WORLD");

    const newlyUpdatedChar = updated.find(c => c.id === activeCharId);
    if (newlyUpdatedChar) {
      if (syncInventory) syncInventory(newlyUpdatedChar.inventory);
      if (syncLevel) syncLevel(newlyUpdatedChar.level);
      if (syncCoins) syncCoins(newlyUpdatedChar.gold);
    }

    setCombatLogs(prev => [
      ...prev,
      "🎉 [Quest Concluída] Você entregou a missão ao Capitão Kenneth!",
      "🎖️ Recompensas: +110 XP, +15 Moedas, +3 Poções de Cura.",
      "🆙 Você subiu para o Nível 2! Suas habilidades canônicas de combate foram expandidas!"
    ]);

    addApiLog("POST", `/api/quests/complete`, 200, 95, JSON.stringify({
      char_id: activeCharId,
      quest_id: "quest_tutorial_sewer_rats",
      level_reached: 2,
      xp_awarded: 110
    }));
  };

  // Cheat code / manual cheat to instantly level up
  const handleLevelCheat = () => {
    if (!activeChar) return;

    const nextLvl = activeChar.level + 1;
    const updated = characters.map(c => {
      if (c.id === activeCharId) {
        return {
          ...c,
          level: nextLvl,
          xp: 0
        };
      }
      return c;
    });

    saveToBackend(updated);
    if (syncLevel) syncLevel(nextLvl);

    setCombatLogs(prev => [...prev, `⚡ [Cheat] Nível alterado manualmente para ${nextLvl}!`]);
    addApiLog("PUT", `/api/characters/${activeCharId}/level-up`, 200, 15, JSON.stringify({ char_id: activeCharId, new_level: nextLvl }));
  };

  if ((sessionState === "IN_WORLD" || sessionState === "TUTORIAL_ACTIVE") && activeChar) {
    return (
      <GameScreen
        activeChar={activeChar}
        characters={characters}
        onUpdateCharacter={(updatedChar) => {
          const updated = characters.map(c => c.id === updatedChar.id ? updatedChar : c);
          setCharacters(updated);
          localStorage.setItem("light_shadow_characters", JSON.stringify(updated));
        }}
        onExitWorld={() => {
          setSessionState("CHARACTER_SELECT");
          addApiLog("POST", "/api/session/logout", 200, 15, JSON.stringify({ char_id: activeCharId }));
        }}
        addApiLog={addApiLog}
        initialPlayerHp={playerHp}
        initialPlayerMana={playerMana}
        syncInventory={syncInventory}
        syncCoins={syncCoins}
        syncLevel={syncLevel}
      />
    );
  }

  return (
    <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100">
      
      {/* LEFT: Game Entry Terminal Screen (Col 7) */}
      <div className="xl:col-span-7 flex flex-col gap-6">
        
        {/* Connection status rail indicator */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 flex items-center justify-between backdrop-blur-sm shadow-md">
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-xl ${isConnected ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" : "bg-rose-500/10 text-rose-400 border border-rose-500/20"}`}>
              {isConnected ? <Wifi className="w-5 h-5 animate-pulse" /> : <WifiOff className="w-5 h-5" />}
            </div>
            <div>
              <span className="text-[10px] text-slate-500 uppercase font-mono tracking-widest block">Servidor de Handshake</span>
              <span className="text-sm font-bold tracking-tight">
                {isConnected ? "Sessão Ativa: Autenticado" : "Desconectado • Handshake pendente"}
              </span>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono px-3 py-1 bg-slate-950/80 border border-slate-800 rounded-lg text-slate-400">
              Ping: {isConnected ? "42ms" : "--"}
            </span>
            {autosavePulse && (
              <span className="text-[10px] font-mono text-emerald-400 animate-pulse font-bold bg-emerald-500/10 px-2.5 py-1 border border-emerald-500/20 rounded-lg">
                💾 AUTOSAVE
              </span>
            )}
          </div>
        </div>

        {/* SCREEN STATE MACHINE VIEWER CONTAINER */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-6 backdrop-blur-sm shadow-xl flex-1 flex flex-col min-h-[500px]">
          
          <AnimatePresence mode="wait">
            
            {/* STATE 1: LOGIN */}
            {sessionState === "LOGIN" && (
              <motion.div
                key="login"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="flex-1 flex flex-col justify-center items-center py-10"
              >
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-500 to-violet-600 flex items-center justify-center shadow-lg shadow-violet-500/10 mb-6">
                  <Play className="w-8 h-8 text-slate-950" />
                </div>
                
                <h2 className="text-2xl font-bold tracking-tight text-center bg-gradient-to-r from-amber-200 to-violet-200 bg-clip-text text-transparent">
                  Light and Shadow Gateway
                </h2>
                <p className="text-xs text-slate-400 text-center max-w-sm mt-1 mb-8">
                  Insira suas credenciais de desenvolvimento para abrir soquetes TCP e efetuar handshake com a infraestrutura autoritativa do servidor.
                </p>

                <form onSubmit={handleLogin} className="w-full max-w-md space-y-4">
                  <div className="space-y-2">
                    <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block">ID de Conta / Nome de Usuário</label>
                    <div className="relative">
                      <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                      <input
                        type="text"
                        value={usernameInput}
                        onChange={(e) => setUsernameInput(e.target.value)}
                        className="w-full pl-11 pr-4 py-3 bg-slate-950/80 border border-slate-800 rounded-xl text-sm font-semibold text-slate-200 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition-all font-mono"
                        placeholder="Nome de Usuário"
                        disabled={isConnecting}
                      />
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={isConnecting}
                    className="w-full py-3.5 px-4 bg-gradient-to-r from-amber-500 to-violet-600 hover:from-amber-600 hover:to-violet-700 disabled:from-slate-800 disabled:to-slate-800 text-slate-950 disabled:text-slate-500 font-bold text-xs uppercase tracking-wider rounded-xl transition-all shadow-lg flex items-center justify-center gap-2"
                  >
                    {isConnecting ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" />
                        Estabelecendo TCP Socket...
                      </>
                    ) : (
                      <>
                        Entrar no Servidor
                        <ArrowRight className="w-4 h-4" />
                      </>
                    )}
                  </button>
                </form>

                <div className="mt-8 flex gap-4 text-[10px] font-mono text-slate-500">
                  <span className="flex items-center gap-1"><Database className="w-3 h-3" /> PostgreSQL Activo</span>
                  <span className="flex items-center gap-1"><Activity className="w-3 h-3" /> Auth: Off-chain</span>
                </div>
              </motion.div>
            )}

            {/* STATE 2: CHARACTER SELECT / CREATE CHOICE */}
            {sessionState === "CHARACTER_SELECT" && (
              <motion.div
                key="char_select"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="flex-1 flex flex-col h-full"
              >
                <div className="flex justify-between items-center mb-6">
                  <div>
                    <h3 className="font-bold text-lg text-slate-200">Selecionar Herói</h3>
                    <p className="text-xs text-slate-500">Escolha uma alma ativa ou desperte uma nova vocação.</p>
                  </div>
                  <button
                    onClick={() => setSessionState("CHARACTER_CREATION")}
                    className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-amber-400 text-[11px] font-bold uppercase tracking-wider rounded-lg flex items-center gap-1.5 transition-all"
                  >
                    <Plus className="w-3.5 h-3.5" /> Criar Herói
                  </button>
                </div>

                {characters.length === 0 ? (
                  <div className="flex-1 flex flex-col items-center justify-center p-10 border border-dashed border-slate-800 rounded-xl bg-slate-950/20">
                    <User className="w-12 h-12 text-slate-600 mb-3" />
                    <p className="text-sm font-semibold text-slate-400">Nenhum herói encontrado no banco de dados.</p>
                    <p className="text-xs text-slate-600 text-center max-w-xs mt-1 mb-4">A base PostgreSQL está vazia para esta conta. Crie o primeiro herói para começar.</p>
                    <button
                      onClick={() => setSessionState("CHARACTER_CREATION")}
                      className="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 text-xs font-bold uppercase tracking-wider rounded-lg shadow-md transition-all"
                    >
                      Criar Novo Herói
                    </button>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 flex-1 overflow-y-auto pr-1">
                    {characters.map(char => {
                      const isSelected = char.id === activeCharId;
                      return (
                        <div
                          key={char.id}
                          onClick={() => handleSelectCharacter(char.id)}
                          className={`p-4 rounded-xl border transition-all cursor-pointer flex flex-col justify-between ${
                            isSelected
                              ? "bg-gradient-to-br from-slate-900 to-slate-950 border-amber-500/80 shadow-md shadow-amber-500/5"
                              : "bg-slate-950/50 border-slate-800 hover:border-slate-700"
                          }`}
                        >
                          <div>
                            <div className="flex justify-between items-start">
                              <div>
                                <h4 className="font-bold text-slate-200 font-mono text-sm">{char.name}</h4>
                                <span className="text-[10px] text-slate-500 font-mono">ID: {char.id.substring(0, 13)}...</span>
                              </div>
                              <span className={`text-[10px] px-2 py-0.5 rounded-lg border font-mono font-semibold ${
                                char.level > 1 ? "bg-amber-500/10 text-amber-400 border-amber-500/20" : "bg-slate-800 text-slate-400 border-slate-700"
                              }`}>
                                Nvl {char.level}
                              </span>
                            </div>

                            <div className="grid grid-cols-2 gap-2 mt-3 text-[10px] font-mono text-slate-400">
                              <span className="flex items-center gap-1"><Sword className="w-3 h-3 text-violet-400" /> Voc: {char.vocation_state}</span>
                              <span className="flex items-center gap-1"><Coins className="w-3 h-3 text-amber-500" /> Ouro: {char.gold}g</span>
                              <span className="flex items-center gap-1"><MapPin className="w-3 h-3 text-emerald-400" /> Pos: {char.position.region}</span>
                              <span className="flex items-center gap-1"><Gift className="w-3 h-3 text-pink-400" /> Quest: {char.quest_state.activeQuests.length} ativa</span>
                            </div>
                          </div>

                          <div className="mt-4 pt-3 border-t border-slate-900 flex justify-between items-center">
                            <span className="text-[9px] text-slate-500 font-mono">Último save: {new Date(char.last_login).toLocaleTimeString()}</span>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteCharacter(char.id, char.name);
                              }}
                              className="p-1 hover:bg-rose-500/10 text-slate-600 hover:text-rose-400 rounded-lg transition-all"
                              title="Excluir Personagem"
                            >
                              <Trash className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}

                {characters.length > 0 && (
                  <div className="mt-6 pt-4 border-t border-slate-800/60 flex justify-end gap-3">
                    <button
                      onClick={() => setSessionState("LOGIN")}
                      className="px-4 py-2 text-slate-400 hover:text-slate-200 text-xs font-bold uppercase tracking-wider rounded-lg transition-all flex items-center gap-1.5"
                    >
                      <LogOut className="w-4 h-4" /> Desconectar
                    </button>
                    <button
                      onClick={handleEnterWorld}
                      disabled={!activeCharId}
                      className="px-6 py-2.5 bg-gradient-to-r from-amber-500 to-yellow-600 hover:from-amber-600 hover:to-yellow-700 disabled:from-slate-800 disabled:to-slate-800 text-slate-950 disabled:text-slate-500 font-bold text-xs uppercase tracking-wider rounded-lg shadow-lg flex items-center gap-1.5 transition-all"
                    >
                      <Play className="w-4 h-4 fill-slate-950" /> Entrar no Mundo (Ironhold)
                    </button>
                  </div>
                )}
              </motion.div>
            )}

            {/* STATE 3: CHARACTER CREATION */}
            {sessionState === "CHARACTER_CREATION" && (
              <motion.div
                key="char_creation"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="flex-1 flex flex-col h-full"
              >
                <div className="mb-6">
                  <h3 className="font-bold text-lg text-slate-200">Criar Novo Personagem</h3>
                  <p className="text-xs text-slate-500">Desperte um herói e prepare seu pacote inicial de treinamento tático.</p>
                </div>

                <form onSubmit={handleCreateCharacter} className="space-y-4 flex-1">
                  <div className="space-y-2">
                    <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block">Nome do Herói (Sem Espaços)</label>
                    <input
                      type="text"
                      value={newCharName}
                      onChange={(e) => setNewCharName(e.target.value.replace(/\s+/g, "_"))}
                      className="w-full px-4 py-2.5 bg-slate-950/80 border border-slate-800 rounded-xl text-sm font-semibold font-mono focus:outline-none focus:border-amber-500"
                      placeholder="Ex: Valen_Sunblade"
                    />
                    {creationError && <p className="text-xs text-rose-400 font-mono mt-1">{creationError}</p>}
                  </div>

                  <div className="space-y-2">
                    <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block">Caminho Vocacional Inicial</label>
                    <div className="grid grid-cols-3 gap-3">
                      {[
                        { id: "melee", name: "Corpo a Corpo (Melee)", desc: "Espada Básica e Escudo de Madeira. Foco em bloqueio.", icon: Shield },
                        { id: "ranged", name: "À Distância (Ranged)", desc: "Arco Curto e Aljava. Foco em precisão.", icon: Sword },
                        { id: "magic", name: "Mana primordial (Magic)", desc: "Cetro e Livro de Feitiços. Foco em recursos.", icon: Wand2 }
                      ].map(path => {
                        const isSel = newCharPath === path.id;
                        const Icon = path.icon;
                        return (
                          <div
                            key={path.id}
                            onClick={() => setNewCharPath(path.id as any)}
                            className={`p-3 rounded-xl border cursor-pointer transition-all flex flex-col gap-2 ${
                              isSel 
                                ? "bg-amber-500/5 border-amber-500/80 shadow-md shadow-amber-500/5" 
                                : "bg-slate-950/40 border-slate-800 hover:border-slate-700"
                            }`}
                          >
                            <div className="flex items-center gap-1.5 text-xs font-bold">
                              <Icon className={`w-4 h-4 ${isSel ? "text-amber-400 animate-pulse" : "text-slate-400"}`} />
                              <span className={isSel ? "text-amber-300" : "text-slate-300"}>{path.name}</span>
                            </div>
                            <p className="text-[10px] text-slate-500 leading-normal">{path.desc}</p>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  <div className="bg-slate-950/40 border border-slate-800/80 rounded-xl p-4 text-xs text-slate-400 leading-relaxed">
                    <p className="font-bold text-slate-300 mb-1 flex items-center gap-1 font-mono uppercase text-[10px]"><Info className="w-3.5 h-3.5 text-amber-500" /> Nota de Arquitetura do Onboarding:</p>
                    A vocação é mantida em <code className="text-amber-300 font-mono font-bold bg-slate-900/60 px-1 py-0.5 rounded">Novice</code> até o nível 10 para evitar builds corrompidas no início. Seus equipamentos iniciais definem como você enfrentará a primeira missão nos esgotos de Ironhold Bastion.
                  </div>

                  <div className="pt-4 border-t border-slate-800/60 flex justify-end gap-3">
                    <button
                      type="button"
                      onClick={() => setSessionState("CHARACTER_SELECT")}
                      className="px-4 py-2 text-slate-400 hover:text-slate-200 text-xs font-bold uppercase tracking-wider rounded-lg transition-all"
                    >
                      Voltar
                    </button>
                    <button
                      type="submit"
                      className="px-6 py-2.5 bg-gradient-to-r from-amber-500 to-violet-600 hover:from-amber-600 hover:to-violet-700 text-slate-950 font-bold text-xs uppercase tracking-wider rounded-lg shadow-lg transition-all flex items-center gap-1.5"
                    >
                      Concluir Criação <CheckCircle className="w-4 h-4" />
                    </button>
                  </div>
                </form>
              </motion.div>
            )}

            {/* STATE 4: LOADING SCREEN */}
            {sessionState === "LOADING" && (
              <motion.div
                key="loading"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex-1 flex flex-col justify-center items-center py-10"
              >
                <div className="relative flex items-center justify-center mb-6">
                  <RefreshCw className="w-16 h-16 text-amber-500 animate-spin" />
                  <span className="absolute text-xs font-mono font-bold text-amber-400">{loadingPercent}%</span>
                </div>

                <h3 className="font-bold text-lg text-slate-200 uppercase tracking-widest font-mono">Carregando Instância...</h3>
                <p className="text-[10px] text-slate-500 font-mono uppercase mt-1 tracking-widest animate-pulse">Autenticando Handshake TCP • Sincronizando PostgreSQL</p>
                
                <div className="w-full max-w-sm bg-slate-950 border border-slate-800/60 rounded-full h-2 mt-6 overflow-hidden">
                  <div className="bg-amber-500 h-full transition-all duration-300" style={{ width: `${loadingPercent}%` }} />
                </div>

                <div className="max-w-md mt-10 p-4 border border-slate-800/80 rounded-xl bg-slate-950/40 text-center">
                  <span className="text-[9px] font-mono text-amber-400 uppercase tracking-widest font-bold block mb-1">Dica de Sobrevivência</span>
                  <p className="text-xs text-slate-400 leading-relaxed italic">"{loadingTip}"</p>
                </div>
              </motion.div>
            )}

            {/* STATE 5 & 6: IN WORLD & TUTORIAL ACTIVE (Simulated game view) */}
            {(sessionState === "IN_WORLD" || sessionState === "TUTORIAL_ACTIVE") && activeChar && (
              <motion.div
                key="in_world"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex-1 flex flex-col h-full justify-between"
              >
                {/* Active Player Status HUD */}
                <div>
                  <div className="flex justify-between items-start pb-4 border-b border-slate-800/80 mb-4">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-xl bg-slate-950 border border-slate-800 flex items-center justify-center font-bold text-amber-400 font-mono shadow-md">
                        {activeChar.level}
                      </div>
                      <div>
                        <span className="text-[10px] text-slate-500 font-mono uppercase tracking-widest block">Herói Ativo</span>
                        <h4 className="font-bold text-slate-200 font-mono text-sm">{activeChar.name}</h4>
                      </div>
                    </div>

                    <div className="text-right">
                      <span className="text-[10px] text-slate-500 font-mono uppercase tracking-widest block">Fração de Mundo</span>
                      <span className="text-xs font-bold text-emerald-400 flex items-center gap-1 justify-end font-mono">
                        <MapPin className="w-3.5 h-3.5" /> {activeChar.position.region} (100, 100, 0)
                      </span>
                    </div>
                  </div>

                  {/* Character Stats & Buff Bars */}
                  <div className="grid grid-cols-3 gap-3 mb-4">
                    <div className="bg-slate-950/80 border border-slate-900 p-3 rounded-xl flex items-center gap-2.5">
                      <div className="p-1.5 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg">
                        <Heart className="w-4 h-4 fill-red-500/10" />
                      </div>
                      <div className="flex-1">
                        <span className="text-[9px] text-slate-500 font-mono uppercase block">Vida (HP)</span>
                        <div className="flex justify-between items-center text-[11px] font-mono font-bold mb-0.5">
                          <span>{playerHp}/120</span>
                        </div>
                        <div className="w-full bg-slate-900 h-1 rounded-full overflow-hidden">
                          <div className="bg-red-500 h-full" style={{ width: `${(playerHp/120)*100}%` }} />
                        </div>
                      </div>
                    </div>

                    <div className="bg-slate-950/80 border border-slate-900 p-3 rounded-xl flex items-center gap-2.5">
                      <div className="p-1.5 bg-sky-500/10 border border-sky-500/20 text-sky-400 rounded-lg">
                        <Wand2 className="w-4 h-4" />
                      </div>
                      <div className="flex-1">
                        <span className="text-[9px] text-slate-500 font-mono uppercase block">Mana (MP)</span>
                        <div className="flex justify-between items-center text-[11px] font-mono font-bold mb-0.5">
                          <span>{playerMana}/40</span>
                        </div>
                        <div className="w-full bg-slate-900 h-1 rounded-full overflow-hidden">
                          <div className="bg-sky-500 h-full" style={{ width: `${(playerMana/40)*100}%` }} />
                        </div>
                      </div>
                    </div>

                    <div className="bg-slate-950/80 border border-slate-900 p-3 rounded-xl flex items-center gap-2.5">
                      <div className="p-1.5 bg-amber-500/10 border border-amber-500/20 text-amber-400 rounded-lg">
                        <Coins className="w-4 h-4" />
                      </div>
                      <div>
                        <span className="text-[9px] text-slate-500 font-mono uppercase block">Ouro Carregado</span>
                        <span className="text-sm font-mono font-bold text-amber-300">{activeChar.gold} gold</span>
                      </div>
                    </div>
                  </div>

                  {/* ACTIVE TUTORIAL QUEST CARD */}
                  {sessionState === "TUTORIAL_ACTIVE" && (
                    <div className="border border-amber-500/30 bg-amber-500/5 rounded-xl p-4 mb-4">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="flex items-center gap-1.5">
                            <Sparkles className="w-4 h-4 text-amber-400 animate-pulse" />
                            <span className="text-[10px] text-amber-400 font-bold uppercase tracking-wider font-mono">Missão Ativa: Tutorial do Capitão</span>
                          </div>
                          <h4 className="font-bold text-sm text-slate-200 mt-1">A Infestação nos Esgotos de Ironhold</h4>
                          <p className="text-[11px] text-slate-400 mt-0.5 leading-relaxed">
                            Limpe os esgotos da cidade derrotando 10 Ratos de Esgoto sob as ordens do Capitão Kenneth.
                          </p>
                        </div>

                        {/* Deliver button if objective met */}
                        {activeChar.quest_state.activeQuests.some(q => 
                          q.quest_id === "quest_tutorial_sewer_rats" && 
                          q.objectives.some(obj => obj.type === "KillMonster" && obj.current_qty >= obj.required_qty)
                        ) && (
                          <button
                            onClick={handleCompleteTutorialQuest}
                            className="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-slate-950 text-[10px] font-bold uppercase tracking-wider rounded-lg shadow-lg animate-bounce flex items-center gap-1 transition-all shrink-0"
                          >
                            <Gift className="w-3.5 h-3.5" /> Entregar Quest
                          </button>
                        )}
                      </div>

                      {/* Quest Objectives Status */}
                      {activeChar.quest_state.activeQuests.find(q => q.quest_id === "quest_tutorial_sewer_rats")?.objectives.map((obj, i) => (
                        <div key={i} className="mt-3 pt-3 border-t border-slate-800 flex justify-between items-center text-xs font-mono">
                          <span className="text-slate-300 flex items-center gap-1">
                            <Sword className="w-3.5 h-3.5 text-slate-500" /> Abater Ratos de Esgoto (Sewer Rats)
                          </span>
                          <span className={`font-bold ${obj.current_qty >= obj.required_qty ? "text-emerald-400" : "text-amber-400"}`}>
                            {obj.current_qty} / {obj.required_qty} {obj.current_qty >= obj.required_qty ? "✓ COMPLETO" : ""}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* ACTIVE STANDARD PLAY STATE CARD */}
                  {sessionState === "IN_WORLD" && (
                    <div className="border border-slate-800 bg-slate-950/40 rounded-xl p-4 mb-4 flex justify-between items-center">
                      <div>
                        <div className="flex items-center gap-1.5 text-emerald-400">
                          <CheckCircle className="w-4 h-4" />
                          <span className="text-[10px] font-bold uppercase tracking-wider font-mono">Formação Completa</span>
                        </div>
                        <h4 className="font-bold text-sm text-slate-200 mt-1">Livre pelo Mundo Aberto</h4>
                        <p className="text-[11px] text-slate-400 mt-0.5">
                          Você limpou os esgotos e graduou para o nível {activeChar.level}. O continente está desbloqueado!
                        </p>
                      </div>

                      <div className="text-right">
                        <span className="text-[9px] font-mono text-slate-500 uppercase block">Total de Quests Concluídas</span>
                        <span className="text-base font-mono font-bold text-emerald-400">
                          {activeChar.quest_state.completedQuestIds.length} Missões
                        </span>
                      </div>
                    </div>
                  )}

                  {/* GAME ACTIONS CONSOLE */}
                  <div className="bg-slate-950/80 border border-slate-900 rounded-xl p-4">
                    <span className="text-[10px] text-slate-500 font-mono uppercase tracking-widest block mb-2 flex items-center gap-1"><Terminal className="w-3.5 h-3.5" /> Console de Combate & Logs de Turno</span>
                    <div className="h-36 overflow-y-auto space-y-1 pr-1 font-mono text-[11px] text-slate-400 bg-slate-950 p-2.5 border border-slate-900 rounded-lg">
                      {combatLogs.slice(-15).map((log, idx) => {
                        let color = "text-slate-400";
                        if (log.includes("🏆") || log.includes("🎉") || log.includes("🆙")) color = "text-emerald-400 font-bold";
                        else if (log.includes("💥") || log.includes("💀")) color = "text-rose-400";
                        else if (log.includes("🛡️") || log.includes("active")) color = "text-amber-400";
                        else if (log.includes("⚡") || log.includes("Cheat")) color = "text-fuchsia-400";
                        return (
                          <div key={idx} className={`${color} leading-relaxed`}>
                            {log}
                          </div>
                        );
                      })}
                      {isFightingRat && (
                        <div className="text-[10px] font-bold text-amber-500 flex items-center gap-1.5 animate-pulse pt-1">
                          <RefreshCw className="w-3.5 h-3.5 animate-spin" /> Processando combate tático de alto TTK...
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                {/* GAME ACTIONS FOOTER BAR */}
                <div className="mt-6 pt-4 border-t border-slate-800/80 flex flex-wrap justify-between items-center gap-3">
                  <div className="flex gap-2">
                    {sessionState === "TUTORIAL_ACTIVE" && (
                      <button
                        onClick={simulateRatCombat}
                        disabled={isFightingRat}
                        className="px-4 py-2 bg-gradient-to-r from-red-500 to-amber-600 hover:from-red-600 hover:to-amber-700 disabled:opacity-50 text-slate-950 font-bold text-xs uppercase tracking-wider rounded-lg shadow-lg flex items-center gap-1.5 transition-all"
                      >
                        <Sword className="w-4 h-4 fill-slate-950" /> Batalhar Rat (10s)
                      </button>
                    )}
                    
                    <button
                      onClick={handleUseHealPotion}
                      disabled={isFightingRat}
                      className="px-4 py-2 bg-slate-800 hover:bg-slate-700 disabled:opacity-50 text-slate-200 font-bold text-xs uppercase tracking-wider rounded-lg border border-slate-700 flex items-center gap-1.5 transition-all"
                    >
                      🧪 Beber Poção
                    </button>

                    <button
                      onClick={handleLevelCheat}
                      disabled={isFightingRat}
                      className="px-3.5 py-2 bg-slate-900 hover:bg-slate-850 text-fuchsia-400 border border-fuchsia-950 hover:border-fuchsia-800 text-xs font-mono font-bold rounded-lg transition-all"
                      title="Cheats de Desenvolvimento para testar Level Up"
                    >
                      ⚡ Level Up Cheat
                    </button>
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={handleAutosave}
                      className="px-3 py-2 bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-slate-200 rounded-lg border border-slate-800 text-xs font-bold uppercase tracking-wider transition-all flex items-center gap-1"
                    >
                      <Save className="w-3.5 h-3.5" /> Forçar Salve
                    </button>
                    <button
                      onClick={handleExitWorld}
                      disabled={isFightingRat}
                      className="px-4 py-2 bg-slate-950 border border-slate-800 hover:border-rose-900/50 hover:text-rose-400 text-slate-400 text-xs font-bold uppercase tracking-wider rounded-lg transition-all flex items-center gap-1.5"
                    >
                      <LogOut className="w-4 h-4" /> Deslogar do Mundo
                    </button>
                  </div>
                </div>

              </motion.div>
            )}

          </AnimatePresence>

        </div>

      </div>

      {/* RIGHT: Backend Authoritative Core Inspector (Col 5) */}
      <div className="xl:col-span-5 flex flex-col gap-6">
        
        {/* Save/Load Lifecycle Diagram Card */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-md">
          <h3 className="font-bold text-sm tracking-tight bg-gradient-to-r from-amber-400 to-violet-400 bg-clip-text text-transparent mb-1 flex items-center gap-1.5 uppercase font-mono">
            <Database className="w-4 h-4 text-violet-400" /> Save/Load Core Lifecycle
          </h3>
          <p className="text-[10px] text-slate-400 leading-normal mb-4">
            Visualização da arquitetura autoritativa do servidor. O cliente comunica-se via TCP/WebSockets e o PersistenceManager executa transações atômicas com Optimistic Locking.
          </p>

          <div className="bg-slate-950/80 border border-slate-900 rounded-xl p-4 font-mono text-[9px] text-slate-400 leading-tight space-y-3">
            <div className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse shrink-0"></span>
              <strong className="text-slate-300">Cliente (Godot):</strong> Dispara <code className="text-amber-300">SaveCharacter</code> em eventos canônicos ou periodicamente.
            </div>
            
            <div className="flex justify-center">
              <span className="text-slate-600 font-bold font-mono">↓↓ [TCP Handshake / SSL Sockets] ↓↓</span>
            </div>

            <div className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-violet-400 shrink-0"></span>
              <strong className="text-slate-300">PersistenceManager (Go):</strong> Abre transação PostgreSQL isolada em <code className="text-violet-300">RepeatableRead</code>.
            </div>

            <div className="flex justify-center animate-pulse">
              <span className="text-slate-600 font-bold font-mono">↓↓ [Validar Versão / Optimistic Locking] ↓↓</span>
            </div>

            <div className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0"></span>
              <strong className="text-slate-300">Database (PostgreSQL):</strong> Atualiza se <code className="text-emerald-300">version == snapshot_version</code>. Se conflitante, faz rollback!
            </div>

            <div className="border-t border-slate-900 pt-3 flex gap-2 justify-between">
              <span className="text-[9px] bg-slate-900 px-2 py-1 border border-slate-800 rounded text-slate-500">Atomicidade: Estrita</span>
              <span className="text-[9px] bg-slate-900 px-2 py-1 border border-slate-800 rounded text-emerald-500 font-bold">Locks: Sem Freada</span>
            </div>
          </div>
        </div>

        {/* Database Raw Inspector (Durable Cloud DB Schema Simulation) */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-md flex-1 flex flex-col justify-between">
          <div>
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-bold text-sm tracking-tight bg-gradient-to-r from-violet-400 to-pink-400 bg-clip-text text-transparent flex items-center gap-1.5 uppercase font-mono">
                <Database className="w-4 h-4 text-pink-400" /> PostgreSQL Raw Table Record
              </h3>
              <span className="text-[9px] font-mono font-bold text-slate-500 px-2 py-0.5 bg-slate-950/80 border border-slate-800 rounded">Table: characters</span>
            </div>
            <p className="text-[10px] text-slate-400 leading-normal mb-4">
              Registro real serializado no LocalStorage simulando a base PostgreSQL do servidor de desenvolvimento.
            </p>

            <div className="font-mono text-[10px] text-slate-400 bg-slate-950 border border-slate-900 p-3 rounded-xl h-56 overflow-y-auto">
              {activeChar ? (
                <pre>{JSON.stringify(activeChar, null, 2)}</pre>
              ) : (
                <div className="flex h-full items-center justify-center text-slate-600 italic">
                  [Nenhum herói ativo ou selecionado]
                </div>
              )}
            </div>
          </div>

          <div className="mt-4 pt-4 border-t border-slate-900 flex justify-between gap-3 text-xs">
            <button
              onClick={() => {
                if (!activeChar) return;
                const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(activeChar, null, 2));
                const dlAnchorElem = document.createElement("a");
                dlAnchorElem.setAttribute("href", dataStr);
                dlAnchorElem.setAttribute("download", `${activeChar.name}_profile.json`);
                dlAnchorElem.click();
                addApiLog("GET", `/api/characters/${activeCharId}/export`, 200, 10, "Exportação concluída com sucesso");
              }}
              disabled={!activeChar}
              className="flex-1 py-2 bg-slate-850 hover:bg-slate-800 text-slate-300 font-semibold rounded-lg border border-slate-800 flex items-center justify-center gap-1.5 transition-all disabled:opacity-50"
            >
              <Download className="w-3.5 h-3.5" /> Exportar Perfil
            </button>
            <label className="flex-1 py-2 bg-slate-850 hover:bg-slate-800 text-slate-300 font-semibold rounded-lg border border-slate-800 flex items-center justify-center gap-1.5 transition-all cursor-pointer text-center">
              <Upload className="w-3.5 h-3.5" /> Importar Perfil
              <input
                type="file"
                accept=".json"
                className="hidden"
                onChange={(e) => {
                  const fileReader = new FileReader();
                  if (e.target.files && e.target.files[0]) {
                    fileReader.readAsText(e.target.files[0], "UTF-8");
                    fileReader.onload = (event) => {
                      try {
                        const parsed = JSON.parse(event.target?.result as string);
                        if (parsed && parsed.id && parsed.name) {
                          const existsIdx = characters.findIndex(c => c.id === parsed.id);
                          let updatedList = [...characters];
                          if (existsIdx !== -1) {
                            updatedList[existsIdx] = parsed;
                          } else {
                            updatedList.push(parsed);
                          }
                          saveToBackend(updatedList);
                          setActiveCharId(parsed.id);
                          addApiLog("POST", "/api/characters/import", 200, 30, `Perfil importado: ${parsed.name}`);
                          alert(`Herói ${parsed.name} importado com sucesso!`);
                        } else {
                          alert("Estrutura JSON inválida.");
                        }
                      } catch (error) {
                        alert("Erro ao ler arquivo JSON.");
                      }
                    };
                  }
                }}
              />
            </label>
          </div>
        </div>

        {/* Server Transaction Terminal (REST / API log console) */}
        <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-md">
          <div className="flex justify-between items-center mb-1">
            <h3 className="font-bold text-sm tracking-tight bg-gradient-to-r from-sky-400 to-indigo-400 bg-clip-text text-transparent flex items-center gap-1.5 uppercase font-mono">
              <Terminal className="w-4 h-4 text-sky-400" /> Authoritative API Server Logs
            </h3>
            <button
              onClick={() => setApiLogs([])}
              className="text-[9px] font-mono text-slate-600 hover:text-slate-400 transition-all uppercase"
            >
              Limpar Console
            </button>
          </div>
          <p className="text-[10px] text-slate-400 leading-normal mb-4">
            Histórico das requisições REST recebidas pelo PersistenceManager. Sincroniza metadados do jogador à prova de falhas.
          </p>

          <div className="bg-slate-950 border border-slate-900 p-3 rounded-xl h-56 overflow-y-auto space-y-1.5 font-mono text-[9px]">
            {apiLogs.length === 0 ? (
              <div className="text-slate-700 italic h-full flex items-center justify-center">[Nenhum log de conexão recebido]</div>
            ) : (
              apiLogs.map(log => {
                let methodColor = "text-amber-500";
                if (log.method === "POST") methodColor = "text-violet-400";
                else if (log.method === "DELETE") methodColor = "text-rose-400 font-bold";
                else if (log.method === "WEBSOCKET") methodColor = "text-emerald-400 font-bold";

                const isSuccess = log.status >= 200 && log.status < 300;

                return (
                  <div key={log.id} className="border-b border-slate-900/60 pb-1 flex justify-between items-start gap-2 leading-relaxed">
                    <div className="flex-1">
                      <span className="text-slate-600">[{log.timestamp}]</span>{" "}
                      <span className={`${methodColor} font-bold`}>{log.method}</span>{" "}
                      <span className="text-slate-300">{log.path}</span>
                      <p className="text-[8px] text-slate-500 leading-tight mt-0.5 max-w-sm overflow-hidden truncate">{log.payload}</p>
                    </div>
                    <div className="text-right shrink-0">
                      <span className={isSuccess ? "text-emerald-400" : "text-amber-400"}>{log.status}</span>
                      <span className="text-slate-600 block text-[8px]">{log.latencyMs}ms</span>
                    </div>
                  </div>
                );
              })
            )}
            <div ref={consoleEndRef} />
          </div>
        </div>

      </div>

    </div>
  );
}
