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
  Wand2,
  Lock,
  Compass as CompassIcon,
  HardDrive,
  Cpu,
  Monitor,
  Eye,
  EyeOff
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { GameScreen } from "./GameScreen";
import {
  CharacterProfile,
  SessionState,
  ServerApiLog,
  InventoryItem,
  EquippedItem,
  QuestState
} from "../../types";

// State Machine for the entry pipeline
export type BootstrapState =
  | "SPLASH_SCREEN"
  | "AUTHENTICATION"
  | "CHARACTER_SELECT"
  | "CHARACTER_CREATION"
  | "LOADING_WORLD"
  | "WORLD_SPAWN";

// Default Starter Quest
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

const INITIAL_CHARACTERS_MOCK: CharacterProfile[] = [
  {
    id: "char-1e24-4a21-98bc-e76bcf129990",
    name: "Gabriela_Paladin",
    race: "Human",
    level: 1,
    xp: 0,
    vocation_state: "Novice",
    position: { x: 10.0, y: 5.0, z: 0.0, region: "Ironhold Bastion" },
    gold: 250,
    inventory: [
      { id: "bronze_coin", name: "Moeda de Bronze", qty: 250, value: 1, type: "currency" },
      { id: "potion_heal", name: "Poção de Cura", qty: 5, value: 15, type: "consumable" },
      { id: "iron_ore", name: "Minério de Ferro", qty: 3, value: 5, type: "material" }
    ],
    equipment: [
      { slot: "Weapon", id: "sword_basic", name: "Espada Básica de Treinamento", value: 100 },
      { slot: "Shield", id: "shield_wooden", name: "Escudo de Madeira", value: 150 }
    ],
    quest_state: {
      activeQuests: [JSON.parse(JSON.stringify(TUTORIAL_QUEST))],
      completedQuestIds: []
    },
    created_at: new Date(Date.now() - 172800000).toISOString(),
    last_login: new Date().toISOString()
  }
];

// RPG lore/immersive survival tips for loading screens
const IMMERSIVE_TIPS = [
  "Ratos de esgoto atacam em bando se você se aproximar demais do lodo profundo.",
  "O escudo wooden wood reduz em até 60% o dano de investidas básicas quando você está mirando.",
  "Mantenha sua Mana ativa. A magia Exura drena apenas 10 MP mas restaura 50 pontos de vida.",
  "Capitão Kenneth costuma treinar novos recrutas no pátio norte de Ironhold.",
  "Você pode usar poções de cura rapidamente clicando nelas em seu inventário lateral.",
  "Os esgotos segregados guardam segredos selados há mais de duzentos anos.",
  "Sempre guarde as caudas de rato coletadas: alquimistas pagam uma boa quantia de bronze por elas."
];

interface GameBootstrapRootProps {
  syncInventory?: (items: any[]) => void;
  syncEquipment?: (items: any[]) => void;
  syncLevel?: (lvl: number) => void;
  syncClass?: (cls: string) => void;
  syncCoins?: (coins: number) => void;
  syncCharName?: (name: string) => void;
}

export function GameBootstrapRoot({
  syncInventory,
  syncEquipment,
  syncLevel,
  syncClass,
  syncCoins,
  syncCharName
}: GameBootstrapRootProps) {
  // Authoritative Persistence States
  const [characters, setCharacters] = useState<CharacterProfile[]>([]);
  const [activeCharId, setActiveCharId] = useState<string>("");
  
  // Strict Entry State Flow
  const [bootstrapState, setBootstrapState] = useState<BootstrapState>("SPLASH_SCREEN");
  
  // Credentials
  const [usernameInput, setUsernameInput] = useState("gabriella_dev");
  const [passwordInput, setPasswordInput] = useState("••••••••");
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  
  // Loading Layer States
  const [loadingPercent, setLoadingPercent] = useState(0);
  const [loadingLogIndex, setLoadingLogIndex] = useState(0);
  const [loadingTip, setLoadingTip] = useState(IMMERSIVE_TIPS[0]);
  const [loadingLogs, setLoadingLogs] = useState<string[]>([]);
  
  // Character Creator States
  const [newCharName, setNewCharName] = useState("");
  const [newCharRace, setNewCharRace] = useState<"Human" | "Forest Elf" | "Ice Elf" | "Dwarf" | "Green Orc">("Human");
  const [creationError, setCreationError] = useState("");

  // Server Transaction / Debug Console
  const [apiLogs, setApiLogs] = useState<ServerApiLog[]>([]);
  const [devModeActive, setDevModeActive] = useState(true);
  const consoleEndRef = useRef<HTMLDivElement>(null);

  // Active stats copy for GameScreen hydration
  const [playerHp, setPlayerHp] = useState(120);
  const [playerMana, setPlayerMana] = useState(40);

  // --- INITIALIZE ACCOUNTS ---
  useEffect(() => {
    const stored = localStorage.getItem("light_shadow_characters");
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        if (parsed && Array.isArray(parsed)) {
          setCharacters(parsed);
        } else {
          setCharacters(INITIAL_CHARACTERS_MOCK);
          localStorage.setItem("light_shadow_characters", JSON.stringify(INITIAL_CHARACTERS_MOCK));
        }
      } catch (e) {
        setCharacters(INITIAL_CHARACTERS_MOCK);
      }
    } else {
      // Don't pre-populate mock, let it be empty to force character creation flow
      // unless they really want a default. Let's make it empty to show the robust Creation flow!
      setCharacters([]);
    }
    
    addApiLog("GET", "/api/client/system-check", 200, 12, "Client engine v2.6-stable OK. Hardware graphics hydrated.");
  }, []);

  // --- AUTOSCROLL DEV WINDOW ---
  useEffect(() => {
    if (consoleEndRef.current) {
      consoleEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [apiLogs]);

  const activeChar = characters.find(c => c.id === activeCharId);

  // Helper to add REST/WS logs
  const addApiLog = (
    method: "GET" | "POST" | "PUT" | "DELETE" | "WEBSOCKET",
    path: string,
    status: number,
    latencyMs: number,
    payload: string
  ) => {
    const newLog: ServerApiLog = {
      id: `log-${Math.random().toString(36).substring(2, 11)}`,
      timestamp: new Date().toLocaleTimeString(),
      method,
      path,
      status,
      latencyMs,
      payload
    };
    setApiLogs(prev => [...prev, newLog].slice(-30));
  };

  const saveToBackend = (updatedList: CharacterProfile[]) => {
    setCharacters(updatedList);
    localStorage.setItem("light_shadow_characters", JSON.stringify(updatedList));
  };

  // --- 1. CONNECT CLIENT FROM SPLASH ---
  const handleClientBoot = () => {
    addApiLog("GET", "/api/auth/handshake", 101, 80, "Initiating secure socket handshake protocol.");
    setBootstrapState("AUTHENTICATION");
  };

  // --- 2. AUTHENTICATION / LOGIN HANDSHAKE ---
  const handleLoginSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!usernameInput.trim()) return;

    setIsConnecting(true);
    addApiLog("POST", "/api/auth/login", 101, 120, JSON.stringify({ username: usernameInput, mode: "development_auth" }));

    setTimeout(() => {
      setIsConnecting(false);
      setIsConnected(true);
      addApiLog("GET", "/api/auth/session", 200, 30, JSON.stringify({ status: "AUTHENTICATED", socket: "CONNECTED_TCP", pool: "PROD_USA_EAST" }));
      
      // SESSION RESTORE LOGIC
      // If characters exist -> go to selection
      // If no characters exist -> force character creation
      if (characters.length > 0) {
        setBootstrapState("CHARACTER_SELECT");
        setActiveCharId(characters[0].id);
      } else {
        addApiLog("GET", "/api/characters", 404, 15, "Nenhum personagem registrado nesta conta.");
        setBootstrapState("CHARACTER_CREATION");
      }
    }, 1000);
  };

  // --- 3. CHARACTER SELECT ACTIONS ---
  const handleSelectCharacter = (id: string) => {
    setActiveCharId(id);
    addApiLog("GET", `/api/characters/${id}`, 200, 20, `Carregando metadados de: ${characters.find(c => c.id === id)?.name}`);
  };

  const handleDeleteCharacter = (id: string, name: string) => {
    const confirmation = window.confirm(`Tem certeza que deseja apagar o herói ${name}? Esta ação é permanente no banco de dados.`);
    if (!confirmation) return;

    const filtered = characters.filter(c => c.id !== id);
    saveToBackend(filtered);
    addApiLog("DELETE", `/api/characters/${id}`, 200, 35, `Registro ${name} removido da tabela.`);

    if (activeCharId === id) {
      if (filtered.length > 0) {
        setActiveCharId(filtered[0].id);
      } else {
        setActiveCharId("");
        setBootstrapState("CHARACTER_CREATION"); // force creation if last deleted
      }
    }
  };

  // --- 4. CHARACTER CREATION ---
  const handleCreateCharacterSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCharName.trim()) {
      setCreationError("O nome do herói é obrigatório.");
      return;
    }
    if (newCharName.length < 3 || newCharName.length > 20) {
      setCreationError("O nome do herói deve conter entre 3 e 20 caracteres.");
      return;
    }
    // Prevent spacebar or weird tags
    if (/[^a-zA-Z0-9_]/.test(newCharName)) {
      setCreationError("Use apenas letras, números e underlines (_). Sem espaços ou caracteres especiais.");
      return;
    }
    if (characters.some(c => c.name.toLowerCase() === newCharName.toLowerCase())) {
      setCreationError("Este nome de personagem já existe no reino.");
      return;
    }

    setCreationError("");
    addApiLog("POST", "/api/characters", 201, 45, `Serializando novo herói: ${newCharName} (${newCharRace})`);

    // starting stats & inventory for Novices
    const startInventory: InventoryItem[] = [
      { id: "bronze_coin", name: "Moeda de Bronze", qty: 250, value: 1, type: "currency" },
      { id: "potion_heal", name: "Poção de Cura", qty: 5, value: 15, type: "consumable" }
    ];
    
    // Novices start with a standard training sword and shield
    const startEquipment: EquippedItem[] = [
      { slot: "Weapon", id: "sword_basic", name: "Espada Básica de Treinamento", value: 100 },
      { slot: "Shield", id: "shield_wooden", name: "Escudo de Madeira", value: 150 }
    ];

    const newChar: CharacterProfile = {
      id: `char-${Math.random().toString(36).substring(2, 10)}-4a21-${Math.random().toString(36).substring(2, 6)}-e76bcf129990`,
      name: newCharName,
      race: newCharRace,
      level: 1,
      xp: 0,
      vocation_state: "Novice",
      position: { x: 10, y: 5, z: 0, region: "Ironhold Bastion" },
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

    const updated = [...characters, newChar];
    saveToBackend(updated);
    setActiveCharId(newChar.id);
    
    // Clear form
    setNewCharName("");
    setBootstrapState("CHARACTER_SELECT");
  };

  // --- 5. ENTER WORLD IMMERSION & LOADING SIMULATOR ---
  const handleEnterWorld = () => {
    if (!activeCharId) return;
    
    setBootstrapState("LOADING_WORLD");
    setLoadingPercent(0);
    setLoadingLogIndex(0);
    setLoadingTip(IMMERSIVE_TIPS[Math.floor(Math.random() * IMMERSIVE_TIPS.length)]);
    setLoadingLogs([]);

    addApiLog("POST", `/api/session/enter-world`, 101, 50, JSON.stringify({ char_id: activeCharId }));
  };

  // Simulated Load Timeline
  useEffect(() => {
    if (bootstrapState !== "LOADING_WORLD") return;

    const interval = setInterval(() => {
      setLoadingPercent(prev => {
        const next = prev + 5;
        if (next >= 100) {
          clearInterval(interval);
          setTimeout(() => {
            // Complete transition
            setBootstrapState("WORLD_SPAWN");
            addApiLog("POST", "/api/session/spawn", 200, 25, JSON.stringify({ status: "SUCCESS", biome: "Ironhold Bastion", coordinates: { x: 10, y: 5 } }));
            
            // Sync to parent hooks
            if (activeChar) {
              if (syncInventory) syncInventory(activeChar.inventory);
              if (syncEquipment) syncEquipment(activeChar.equipment);
              if (syncLevel) syncLevel(activeChar.level);
              if (syncClass) syncClass(activeChar.vocation_state);
              if (syncCoins) syncCoins(activeChar.gold);
              if (syncCharName) syncCharName(activeChar.name);
            }
          }, 600);
          return 100;
        }
        return next;
      });
    }, 150);

    return () => clearInterval(interval);
  }, [bootstrapState, activeCharId]);

  // Handle step-by-step loading log insertions
  useEffect(() => {
    if (bootstrapState !== "LOADING_WORLD") return;

    const percentageToLogs = [
      { p: 5, log: "🔌 Estabelecendo handshake com servidor central..." },
      { p: 25, log: "🛡️ Autenticando token do herói e sincronizando PostgreSQL local..." },
      { p: 45, log: "🌍 Carregando dados espaciais de Ironhold Bastion (20x20)..." },
      { p: 65, log: "🕷️ Inicializando ecossistema de esgotos e respawn de monstros..." },
      { p: 85, log: "🕯️ Ativando iluminação dinâmica de tocha e fumaça local..." },
      { p: 100, log: "✅ Spawn de personagem pronto para imersão!" }
    ];

    percentageToLogs.forEach(step => {
      if (loadingPercent >= step.p && !loadingLogs.includes(step.log)) {
        setLoadingLogs(prev => [...prev, step.log]);
        addApiLog("WEBSOCKET", "/stream/world-load", 200, 10, step.log);
      }
    });
  }, [loadingPercent, bootstrapState]);

  return (
    <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100">
      
      {/* LEFT OR MAIN: Launcher / State Frame (Col 7 or Full depending on dev mode) */}
      <div className={`${devModeActive ? "xl:col-span-8" : "xl:col-span-12"} flex flex-col gap-6`}>
        
        {/* Launcher Header Status Area */}
        <div className="bg-slate-900/80 border border-slate-800/80 rounded-2xl p-4 flex items-center justify-between backdrop-blur-md shadow-lg">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <CompassIcon className="w-5 h-5 animate-spin-slow" />
            </div>
            <div>
              <span className="text-[10px] text-slate-500 uppercase font-mono tracking-widest block">CLIENT BOOTSTRAP CONTROL</span>
              <span className="text-sm font-bold text-slate-200">
                Light and Shadow Client Launcher <span className="text-xs text-amber-500 font-mono">v2.6-stable</span>
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setDevModeActive(!devModeActive)}
              className="px-3 py-1.5 bg-slate-950 border border-slate-800 hover:border-slate-700 rounded-lg text-xs font-mono text-slate-400 flex items-center gap-1.5 transition-all"
            >
              {devModeActive ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
              {devModeActive ? "Ocultar Console Dev" : "Mostrar Console Dev"}
            </button>
            <span className="text-xs font-mono px-2.5 py-1 bg-slate-950/80 border border-slate-800 rounded-lg text-emerald-400 font-semibold flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              SERVER ONLINE
            </span>
          </div>
        </div>

        {/* AUTHORITATIVE STATE MACHINE CONTAINER */}
        <div className="bg-slate-900/50 border border-slate-800/80 rounded-3xl p-6 backdrop-blur-sm shadow-2xl flex-1 flex flex-col min-h-[580px] relative overflow-hidden justify-center">
          
          <AnimatePresence mode="wait">
            
            {/* 1. SPLASH SCREEN */}
            {bootstrapState === "SPLASH_SCREEN" && (
              <motion.div
                key="splash"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                transition={{ duration: 0.4 }}
                className="flex-1 flex flex-col justify-center items-center py-12 text-center"
              >
                <div className="relative mb-6">
                  <div className="absolute -inset-4 bg-gradient-to-tr from-amber-500 to-violet-600 rounded-full blur-2xl opacity-20 animate-pulse"></div>
                  <div className="relative w-24 h-24 rounded-3xl bg-gradient-to-tr from-amber-500 to-violet-600 flex items-center justify-center shadow-2xl border border-white/10">
                    <Sparkles className="w-12 h-12 text-slate-950 animate-bounce" />
                  </div>
                </div>

                <h1 className="text-4xl font-extrabold tracking-tight bg-gradient-to-r from-amber-200 via-orange-300 to-violet-300 bg-clip-text text-transparent drop-shadow">
                  LIGHT AND SHADOW
                </h1>
                <p className="text-xs uppercase tracking-widest font-mono text-amber-500/80 font-bold mt-1.5">
                  The Ironhold Crusades • MMORPG Client
                </p>
                <p className="text-xs text-slate-400 max-w-md mt-4 leading-relaxed">
                  Desça às entranhas do Bastião Seguro de Ironhold. Teste suas vocações primordiais, derrote pragas no subterrâneo e conquiste itens lendários com combate tático e mecânicas canônicas de grid.
                </p>

                <div className="mt-10 w-full max-w-sm space-y-3">
                  <button
                    onClick={handleClientBoot}
                    className="w-full py-4 bg-gradient-to-r from-amber-500 via-orange-500 to-violet-600 hover:from-amber-600 hover:to-violet-700 text-slate-950 font-bold text-xs uppercase tracking-widest rounded-xl transition-all shadow-xl shadow-amber-500/10 flex items-center justify-center gap-2 group border-t border-white/20"
                  >
                    Estreitar Conexão de Jogo
                    <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                  </button>
                  <div className="flex justify-center gap-6 text-[10px] font-mono text-slate-500 pt-3">
                    <span className="flex items-center gap-1"><HardDrive className="w-3.5 h-3.5" /> Client v2.6.4</span>
                    <span className="flex items-center gap-1"><Cpu className="w-3.5 h-3.5" /> Direct3D 12</span>
                    <span className="flex items-center gap-1"><Monitor className="w-3.5 h-3.5" /> 60 FPS Target</span>
                  </div>
                </div>
              </motion.div>
            )}

            {/* 2. AUTHENTICATION */}
            {bootstrapState === "AUTHENTICATION" && (
              <motion.div
                key="auth"
                initial={{ opacity: 0, y: 15 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -15 }}
                transition={{ duration: 0.3 }}
                className="flex-1 flex flex-col justify-center items-center py-8"
              >
                <div className="w-16 h-16 rounded-2xl bg-slate-950 border border-slate-800 flex items-center justify-center shadow-lg mb-4">
                  <Lock className="w-6 h-6 text-amber-400" />
                </div>
                
                <h2 className="text-2xl font-bold tracking-tight text-center text-slate-200">
                  Autenticação do Servidor
                </h2>
                <p className="text-xs text-slate-400 text-center max-w-sm mt-1 mb-6 leading-relaxed">
                  Insira o ID da sua Conta para carregar os registros de heróis no banco PostgreSQL persistente.
                </p>

                <form onSubmit={handleLoginSubmit} className="w-full max-w-md space-y-4 bg-slate-950/40 p-6 border border-slate-800/60 rounded-2xl">
                  <div className="space-y-2">
                    <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block">ID de Conta / Nome de Usuário</label>
                    <div className="relative">
                      <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                      <input
                        type="text"
                        value={usernameInput}
                        onChange={(e) => setUsernameInput(e.target.value)}
                        className="w-full pl-11 pr-4 py-3 bg-slate-950/80 border border-slate-850 rounded-xl text-sm font-semibold text-slate-200 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition-all font-mono"
                        placeholder="Nome de Usuário"
                        disabled={isConnecting}
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block">Token de Acesso / Senha</label>
                    <div className="relative">
                      <Key className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-600" />
                      <input
                        type="password"
                        value={passwordInput}
                        onChange={(e) => setPasswordInput(e.target.value)}
                        className="w-full pl-11 pr-4 py-3 bg-slate-950/80 border border-slate-850 rounded-xl text-sm font-semibold text-slate-500 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition-all font-mono"
                        placeholder="Senha"
                        disabled
                      />
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={isConnecting}
                    className="w-full py-3.5 px-4 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 disabled:from-slate-800 disabled:to-slate-800 text-slate-950 disabled:text-slate-500 font-bold text-xs uppercase tracking-wider rounded-xl transition-all shadow-lg flex items-center justify-center gap-2 border-t border-white/10"
                  >
                    {isConnecting ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" />
                        Abrindo TCP Socket...
                      </>
                    ) : (
                      <>
                        Fazer Login Seguro
                        <ArrowRight className="w-4 h-4" />
                      </>
                    )}
                  </button>
                </form>

                <button
                  onClick={() => setBootstrapState("SPLASH_SCREEN")}
                  className="mt-6 text-xs text-slate-500 hover:text-slate-300 font-mono"
                >
                  ← Voltar para tela inicial
                </button>
              </motion.div>
            )}

            {/* 3. CHARACTER SELECT */}
            {bootstrapState === "CHARACTER_SELECT" && (
              <motion.div
                key="char_select"
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.98 }}
                transition={{ duration: 0.3 }}
                className="flex-1 flex flex-col h-full justify-between"
              >
                <div>
                  <div className="flex justify-between items-center mb-6 border-b border-slate-800/60 pb-4">
                    <div>
                      <h3 className="font-bold text-xl text-slate-200">Selecione seu Herói</h3>
                      <p className="text-xs text-slate-500">Escolha uma alma ativa ou crie um novo personagem para os esgotos.</p>
                    </div>
                    <button
                      onClick={() => setBootstrapState("CHARACTER_CREATION")}
                      className="px-3.5 py-2 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/30 text-amber-400 text-xs font-bold uppercase tracking-wider rounded-lg flex items-center gap-1.5 transition-all"
                    >
                      <Plus className="w-4 h-4" /> Criar Novo Herói
                    </button>
                  </div>

                  {characters.length === 0 ? (
                    <div className="py-16 flex flex-col items-center justify-center border border-dashed border-slate-800 rounded-2xl bg-slate-950/20">
                      <User className="w-12 h-12 text-slate-700 mb-3" />
                      <p className="text-sm font-semibold text-slate-400">Nenhum herói ativo encontrado.</p>
                      <p className="text-xs text-slate-600 text-center max-w-xs mt-1 mb-4">Crie o primeiro guerreiro para iniciar sua jornada em Ironhold Bastion.</p>
                      <button
                        onClick={() => setBootstrapState("CHARACTER_CREATION")}
                        className="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 text-xs font-bold uppercase tracking-wider rounded-lg shadow-md transition-all font-mono"
                      >
                        Criar Novo Herói
                      </button>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-h-[380px] overflow-y-auto pr-1">
                      {characters.map(char => {
                        const isSelected = char.id === activeCharId;
                        return (
                          <div
                            key={char.id}
                            onClick={() => handleSelectCharacter(char.id)}
                            className={`p-5 rounded-2xl border text-left transition-all cursor-pointer flex flex-col justify-between ${
                              isSelected
                                ? "bg-gradient-to-br from-slate-900 to-slate-950 border-amber-500/80 shadow-md shadow-amber-500/5 ring-1 ring-amber-500/20"
                                : "bg-slate-950/50 border-slate-800 hover:border-slate-700 hover:bg-slate-900/10"
                            }`}
                          >
                            <div>
                              <div className="flex justify-between items-start">
                                <div className="flex items-center gap-3">
                                  <div className={`p-2 rounded-xl ${isSelected ? "bg-amber-500/10 text-amber-400" : "bg-slate-900 text-slate-500"}`}>
                                    <User className="w-5 h-5" />
                                  </div>
                                  <div>
                                    <h4 className="font-bold text-slate-200 font-mono text-sm tracking-tight">{char.name}</h4>
                                    <span className="text-[10px] text-slate-500 font-mono block">ID: {char.id.substring(0, 13)}...</span>
                                  </div>
                                </div>
                                <span className={`text-[10px] px-2.5 py-1 rounded-lg border font-mono font-extrabold ${
                                  char.level > 1 ? "bg-amber-500/10 text-amber-400 border-amber-500/20" : "bg-slate-800 text-slate-400 border-slate-700"
                                }`}>
                                  LVL {char.level}
                                </span>
                              </div>

                              <div className="grid grid-cols-2 gap-y-2 gap-x-4 mt-4 text-[11px] font-mono text-slate-400 border-t border-slate-900/60 pt-3">
                                <span className="flex items-center gap-1.5"><Sword className="w-3.5 h-3.5 text-violet-400" /> Vocação: {char.vocation_state}</span>
                                <span className="flex items-center gap-1.5"><Coins className="w-3.5 h-3.5 text-amber-500" /> Gold: {char.gold}gp</span>
                                <span className="flex items-center gap-1.5"><MapPin className="w-3.5 h-3.5 text-emerald-400" /> Região: {char.position.region}</span>
                                <span className="flex items-center gap-1.5"><Gift className="w-3.5 h-3.5 text-pink-400" /> Quests: {char.quest_state.activeQuests.length} ativa</span>
                              </div>
                            </div>

                            <div className="mt-4 pt-3 border-t border-slate-900 flex justify-between items-center">
                              <span className="text-[9px] text-slate-500 font-mono block">Último Login: {new Date(char.last_login).toLocaleTimeString()}</span>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleDeleteCharacter(char.id, char.name);
                                }}
                                className="p-1.5 hover:bg-rose-500/10 text-slate-600 hover:text-rose-400 rounded-lg transition-all"
                                title="Excluir herói permanentemente"
                              >
                                <Trash className="w-4 h-4" />
                              </button>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>

                <div className="mt-6 pt-4 border-t border-slate-800/60 flex justify-end gap-3">
                  <button
                    onClick={() => setBootstrapState("AUTHENTICATION")}
                    className="px-4 py-2 text-slate-400 hover:text-slate-200 text-xs font-bold uppercase tracking-wider rounded-lg transition-all flex items-center gap-1.5"
                  >
                    <LogOut className="w-4 h-4" /> Desconectar Conta
                  </button>
                  <button
                    onClick={handleEnterWorld}
                    disabled={!activeCharId}
                    className="px-6 py-3 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 disabled:from-slate-800 disabled:to-slate-800 text-slate-950 disabled:text-slate-500 font-bold text-xs uppercase tracking-wider rounded-lg shadow-lg flex items-center gap-1.5 transition-all border-t border-white/15"
                  >
                    <Play className="w-4 h-4 fill-slate-950" /> Entrar no Mundo (Ironhold)
                  </button>
                </div>
              </motion.div>
            )}

            {/* 4. CHARACTER CREATION */}
            {bootstrapState === "CHARACTER_CREATION" && (
              <motion.div
                key="char_create"
                initial={{ opacity: 0, y: 15 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -15 }}
                transition={{ duration: 0.3 }}
                className="flex-1 flex flex-col h-full justify-between"
              >
                <div>
                  <div className="mb-6 border-b border-slate-800/60 pb-3">
                    <h3 className="font-bold text-xl text-slate-200">Desperte sua Linhagem Racial</h3>
                    <p className="text-xs text-slate-500">Selecione uma raça e forneça sua identificação canônica para registrar o herói.</p>
                  </div>

                  <form onSubmit={handleCreateCharacterSubmit} className="space-y-5">
                    <div className="space-y-2">
                      <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block font-bold">Identidade / Nome do Herói (Sem Espaços)</label>
                      <input
                        type="text"
                        value={newCharName}
                        onChange={(e) => {
                          const val = e.target.value.replace(/\s+/g, "_");
                          setNewCharName(val);
                        }}
                        className="w-full px-4 py-3 bg-slate-950/80 border border-slate-800 rounded-xl text-sm font-semibold font-mono text-slate-200 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition-all"
                        placeholder="Ex: Alistair_Blade"
                        maxLength={20}
                      />
                      {creationError && <p className="text-xs text-rose-400 font-mono">{creationError}</p>}
                    </div>

                    <div className="space-y-2">
                      <label className="text-[10px] font-mono text-slate-500 uppercase tracking-widest block font-bold">Selecione sua Raça (Aparência Canônica Fixa)</label>
                      <div className="grid grid-cols-2 md:grid-cols-5 gap-2">
                        {[
                          { id: "Human", name: "Humano", desc: "Equilibrado. Traje de Explorador de Algodão e Capuz do Bastião." },
                          { id: "Forest Elf", name: "Elfo da Floresta", desc: "Ágil. Manto de Folhas Tecidas e Bandagem Esmeralda." },
                          { id: "Ice Elf", name: "Elfo de Gelo", desc: "Místico. Manto de Lã Ártica Azul e Brincos de Cristal Glacial." },
                          { id: "Dwarf", name: "Anão", desc: "Robusto. Armadura de Couro Reforçado e Fivelas de Bronze." },
                          { id: "Green Orc", name: "Orc Verde", desc: "Feroz. Túnica Rústica de Pele de Lobo e Marcas de Argila." }
                        ].map(race => {
                          const isSel = newCharRace === race.id;
                          return (
                            <div
                              key={race.id}
                              onClick={() => setNewCharRace(race.id as any)}
                              className={`p-3 rounded-xl border cursor-pointer transition-all flex flex-col justify-between text-left h-36 ${
                                isSel 
                                  ? "bg-amber-500/5 border-amber-500/80 shadow-md shadow-amber-500/5 ring-1 ring-amber-500/10" 
                                  : "bg-slate-950/40 border-slate-800 hover:border-slate-700"
                              }`}
                            >
                              <div className="text-xs font-bold font-mono">
                                <span className={isSel ? "text-amber-300" : "text-slate-300"}>{race.name}</span>
                              </div>
                              <p className="text-[10px] text-slate-500 leading-snug mt-1">{race.desc}</p>
                              <span className="text-[8px] font-mono uppercase tracking-widest text-slate-600 block mt-auto">
                                {isSel ? "● SELECIONADO" : "SELECIONAR"}
                              </span>
                            </div>
                          );
                        })}
                      </div>
                    </div>

                    <div className="bg-slate-950/40 border border-slate-800/80 rounded-xl p-4 text-xs text-slate-400 leading-relaxed text-left flex gap-3">
                      <Info className="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
                      <div>
                        <p className="font-bold text-slate-300 font-mono uppercase text-[9px] mb-0.5">Nota de Arquitetura do Onboarding:</p>
                        Todos os aventureiros nascem sob a vocação primordial de <code className="text-amber-300 font-mono font-bold bg-slate-900/60 px-1 py-0.5 rounded">Novice</code> (Level 1-9). Ao alcançar o <code className="text-amber-300 font-mono">Level 10</code>, você poderá visitar a Câmara de Ascensão para consagrar uma vocação verdadeira.
                      </div>
                    </div>
                  </form>
                </div>

                <div className="mt-6 pt-4 border-t border-slate-800/60 flex justify-end gap-3">
                  {characters.length > 0 && (
                    <button
                      type="button"
                      onClick={() => setBootstrapState("CHARACTER_SELECT")}
                      className="px-4 py-2 text-slate-400 hover:text-slate-200 text-xs font-bold uppercase tracking-wider rounded-lg transition-all"
                    >
                      Voltar
                    </button>
                  )}
                  <button
                    onClick={handleCreateCharacterSubmit}
                    className="px-6 py-2.5 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-slate-950 font-bold text-xs uppercase tracking-wider rounded-lg shadow-lg transition-all flex items-center gap-1.5 border-t border-white/10"
                  >
                    Registrar e Despertar <CheckCircle className="w-4 h-4" />
                  </button>
                </div>
              </motion.div>
            )}

            {/* 5. LOADING WORLD */}
            {bootstrapState === "LOADING_WORLD" && (
              <motion.div
                key="loading"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex-1 flex flex-col justify-center items-center py-10"
              >
                <div className="relative flex items-center justify-center mb-6">
                  <div className="absolute inset-0 rounded-full border border-amber-500/20 animate-ping"></div>
                  <RefreshCw className="w-16 h-16 text-amber-500 animate-spin" />
                  <span className="absolute text-xs font-mono font-extrabold text-amber-400">{loadingPercent}%</span>
                </div>

                <h3 className="font-bold text-lg text-slate-200 uppercase tracking-widest font-mono">Conectando a Ironhold...</h3>
                <p className="text-[10px] text-slate-500 font-mono uppercase mt-1 tracking-widest animate-pulse">Sincronizando Metadados • Hydrating SQLite Shards</p>
                
                <div className="w-full max-w-md bg-slate-950 border border-slate-800/60 rounded-full h-2.5 mt-6 overflow-hidden p-[2px]">
                  <div className="bg-gradient-to-r from-amber-500 via-orange-500 to-violet-600 h-full rounded-full transition-all duration-300 shadow shadow-amber-500/30" style={{ width: `${loadingPercent}%` }} />
                </div>

                {/* Staggered progress terminal logs */}
                <div className="w-full max-w-md mt-6 bg-slate-950/80 border border-slate-900 rounded-xl p-3 h-28 overflow-y-auto text-left font-mono text-[9px] text-slate-500 space-y-1">
                  {loadingLogs.map((log, index) => (
                    <div key={index} className="text-slate-400 animate-fade-in flex items-center gap-1.5">
                      <span className="text-emerald-500">✔</span>
                      <span>{log}</span>
                    </div>
                  ))}
                </div>

                <div className="max-w-md mt-8 p-4 border border-slate-800/80 rounded-xl bg-slate-950/40 text-center">
                  <span className="text-[9px] font-mono text-amber-400 uppercase tracking-widest font-bold block mb-1">Dica de Sobrevivência</span>
                  <p className="text-xs text-slate-400 leading-relaxed italic">"{loadingTip}"</p>
                </div>
              </motion.div>
            )}

            {/* 6. WORLD SPAWN (Renders GameScreen) */}
            {bootstrapState === "WORLD_SPAWN" && activeChar && (
              <motion.div
                key="game_view"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.5 }}
                className="flex-1 flex flex-col h-full"
              >
                <GameScreen
                  activeChar={activeChar}
                  characters={characters}
                  onUpdateCharacter={(updatedChar) => {
                    const updated = characters.map(c => c.id === updatedChar.id ? updatedChar : c);
                    setCharacters(updated);
                    localStorage.setItem("light_shadow_characters", JSON.stringify(updated));
                  }}
                  onExitWorld={() => {
                    setBootstrapState("CHARACTER_SELECT");
                    addApiLog("POST", "/api/session/logout", 200, 15, JSON.stringify({ char_id: activeCharId }));
                  }}
                  addApiLog={addApiLog}
                  initialPlayerHp={playerHp}
                  initialPlayerMana={playerMana}
                  syncInventory={syncInventory}
                  syncCoins={syncCoins}
                  syncLevel={syncLevel}
                />
              </motion.div>
            )}

          </AnimatePresence>

        </div>

      </div>

      {/* RIGHT PANEL: Developer Transaction Console / REST API debugger (Collapsible) */}
      {devModeActive && (
        <div className="xl:col-span-4 flex flex-col gap-6">
          
          {/* Authoritative PostgreSQL Inspector */}
          <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-md flex flex-col">
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-bold text-xs tracking-tight bg-gradient-to-r from-violet-400 to-pink-400 bg-clip-text text-transparent flex items-center gap-1.5 uppercase font-mono">
                <Database className="w-4 h-4 text-pink-400" /> PostgreSQL Raw Records
              </h3>
              <span className="text-[9px] font-mono font-bold text-slate-500 px-2 py-0.5 bg-slate-950/80 border border-slate-800 rounded">Table: characters</span>
            </div>
            <p className="text-[10px] text-slate-400 leading-normal mb-4">
              Visualização das tabelas PostgreSQL simuladas em localStorage. Sincronização em tempo real.
            </p>

            <div className="font-mono text-[9px] text-slate-500 bg-slate-950 border border-slate-900 p-3 rounded-xl h-44 overflow-y-auto text-left">
              {activeChar ? (
                <pre className="text-slate-400">{JSON.stringify(activeChar, null, 2)}</pre>
              ) : (
                <div className="flex h-full items-center justify-center text-slate-600 italic">
                  [Aguardando seleção de herói]
                </div>
              )}
            </div>

            <div className="mt-4 pt-3 border-t border-slate-900 flex justify-between gap-3 text-[11px]">
              <button
                onClick={() => {
                  if (!activeChar) return;
                  const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(activeChar, null, 2));
                  const dlAnchorElem = document.createElement("a");
                  dlAnchorElem.setAttribute("href", dataStr);
                  dlAnchorElem.setAttribute("download", `${activeChar.name}_profile.json`);
                  dlAnchorElem.click();
                  addApiLog("GET", `/api/characters/${activeCharId}/export`, 200, 10, "Perfil exportado como JSON.");
                }}
                disabled={!activeChar}
                className="flex-1 py-1.5 bg-slate-850 hover:bg-slate-800 text-slate-300 font-semibold rounded-lg border border-slate-800 flex items-center justify-center gap-1.5 transition-all disabled:opacity-50"
              >
                <Download className="w-3.5 h-3.5" /> Exportar Perfil
              </button>
              <label className="flex-1 py-1.5 bg-slate-850 hover:bg-slate-800 text-slate-300 font-semibold rounded-lg border border-slate-800 flex items-center justify-center gap-1.5 transition-all cursor-pointer text-center">
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

          {/* Core Handshake Log Timeline */}
          <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-5 backdrop-blur-sm shadow-md">
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-bold text-xs tracking-tight bg-gradient-to-r from-sky-400 to-indigo-400 bg-clip-text text-transparent flex items-center gap-1.5 uppercase font-mono">
                <Terminal className="w-4 h-4 text-sky-400" /> Authoritative API Logs
              </h3>
              <button
                onClick={() => setApiLogs([])}
                className="text-[9px] font-mono text-slate-600 hover:text-slate-400 transition-all uppercase"
              >
                Limpar Console
              </button>
            </div>
            <p className="text-[10px] text-slate-400 leading-normal mb-4">
              Requisições REST recebidas pelo PersistenceManager. Controla isolamento de transações.
            </p>

            <div className="bg-slate-950 border border-slate-900 p-3 rounded-xl h-64 overflow-y-auto space-y-2 font-mono text-[9px] text-left">
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
                    <div key={log.id} className="border-b border-slate-900/60 pb-1.5 flex justify-between items-start gap-2 leading-relaxed">
                      <div className="flex-1 min-w-0">
                        <span className="text-slate-600">[{log.timestamp}]</span>{" "}
                        <span className={`${methodColor} font-bold`}>{log.method}</span>{" "}
                        <span className="text-slate-300 break-all">{log.path}</span>
                        <p className="text-[8px] text-slate-500 leading-tight mt-0.5 truncate">{log.payload}</p>
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

          {/* Architectonics info */}
          <div className="bg-slate-900/60 border border-slate-800/60 rounded-2xl p-4 text-[11px] text-slate-400 text-left space-y-2">
            <span className="font-bold text-amber-400 block font-mono text-[9px] uppercase tracking-wider">🔒 ESTADO DO BOOTSTRAP:</span>
            <div className="grid grid-cols-2 gap-1 font-mono text-[10px]">
              <span className="text-slate-500">Etapa Atual:</span>
              <span className="text-slate-200 font-bold">{bootstrapState}</span>
              <span className="text-slate-500">Sessão:</span>
              <span className={isConnected ? "text-emerald-400" : "text-rose-400"}>
                {isConnected ? "CONNECTED" : "OFFLINE"}
              </span>
              <span className="text-slate-500">Handshake:</span>
              <span className="text-slate-300">TCP Stream Socket</span>
            </div>
          </div>

        </div>
      )}

    </div>
  );
}
