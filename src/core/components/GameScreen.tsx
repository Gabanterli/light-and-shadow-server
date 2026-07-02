import React, { useState, useEffect, useRef } from "react";
import {
  User,
  Shield,
  Sword,
  Coins,
  Heart,
  Wand2,
  MapPin,
  Gift,
  CheckCircle,
  Play,
  LogOut,
  Save,
  Terminal,
  Flame,
  Sparkles,
  Zap,
  Info,
  Package,
  Compass,
  ShieldAlert,
  ArrowDown
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { CharacterProfile, QuestState, InventoryItem, EquippedItem } from "../../types";

// Import modular Tibia controls
import { TibiaMinimap } from "./TibiaMinimap";
import { TibiaEquipment } from "./TibiaEquipment";
import { TibiaBattleList } from "./TibiaBattleList";
import { TibiaChat } from "./TibiaChat";

// Import progression pipeline and engine integrations
import { progressionEventBus } from "../progression_event_bus";
import { progressionEngine } from "../progression_engine";
import { TibiaBattleSystem } from "../TibiaBattleSystem";
import { CombatEngine } from "../CombatEngine";
import { SpellSystem } from "../SpellSystem";
import { QuestSystem } from "../QuestSystem";

// Map dimensions
const MAP_WIDTH = 20;
const MAP_HEIGHT = 20;

// Tile types
type TileType = "FLOOR_STONE" | "FLOOR_SEWER" | "WALL" | "LADDER_UP" | "LADDER_DOWN" | "WATER" | "LOCKED_GATE";

interface MapTile {
  x: number;
  y: number;
  type: TileType;
  name: string;
}

interface Monster {
  id: string;
  name: string;
  type: "sewer_rat" | "giant_rat" | "sewer_slime";
  x: number;
  y: number;
  hp: number;
  maxHp: number;
  damage: number;
  expReward: number;
  isDead: boolean;
  respawnTime: number; // in turns/ticks
}

interface MapCorpse {
  id: string;
  x: number;
  y: number;
  name: string;
  items: InventoryItem[];
  looted: boolean;
}

interface FloatingText {
  id: string;
  x: number;
  y: number;
  text: string;
  color: string;
}

interface GameScreenProps {
  activeChar: CharacterProfile;
  characters: CharacterProfile[];
  onUpdateCharacter: (updatedChar: CharacterProfile) => void;
  onExitWorld: () => void;
  addApiLog: (
    method: "GET" | "POST" | "PUT" | "DELETE" | "WEBSOCKET",
    path: string,
    status: number,
    latencyMs: number,
    payload: string
  ) => void;
  initialPlayerHp: number;
  initialPlayerMana: number;
  syncInventory?: (items: any[]) => void;
  syncCoins?: (coins: number) => void;
  syncLevel?: (lvl: number) => void;
}

export function GameScreen({
  activeChar,
  characters,
  onUpdateCharacter,
  onExitWorld,
  addApiLog,
  initialPlayerHp,
  initialPlayerMana,
  syncInventory,
  syncCoins,
  syncLevel
}: GameScreenProps) {
  // --- CORE GAME STATE ---
  const [playerPos, setPlayerPos] = useState({ x: activeChar.position?.x ?? 10, y: activeChar.position?.y ?? 5 });
  const [playerHp, setPlayerHp] = useState(initialPlayerHp);
  const [playerMana, setPlayerMana] = useState(initialPlayerMana);
  const [playerLevel, setPlayerLevel] = useState(activeChar.level);
  const [playerGold, setPlayerGold] = useState(activeChar.gold);
  const [playerXp, setPlayerXp] = useState(activeChar.xp);
  const [playerInventory, setPlayerInventory] = useState<InventoryItem[]>(activeChar.inventory || []);
  const [playerEquipment, setPlayerEquipment] = useState<EquippedItem[]>(activeChar.equipment || []);
  const [questState, setQuestState] = useState<CharacterProfile["quest_state"]>(activeChar.quest_state);

  // Monsters & Corpses
  const [monsters, setMonsters] = useState<Monster[]>([]);
  const [corpses, setCorpses] = useState<MapCorpse[]>([]);
  const [openContainer, setOpenContainer] = useState<{
    name: string;
    items: InventoryItem[];
    corpseId: string;
  } | null>(null);

  const [targetId, setTargetId] = useState<string | null>(null);

  // Fog of War (Explored coordinates)
  const [exploredTiles, setExploredTiles] = useState<Record<string, boolean>>({});

  // Floating text messages for combat/spell feedback
  const [floatingTexts, setFloatingTexts] = useState<FloatingText[]>([]);

  // Logs terminal
  const [gameLogs, setGameLogs] = useState<string[]>([
    `[Sessão] Conectado como ${activeChar.name}. Bem-vindo a Ironhold Bastion!`,
    `[Mouse Semantics] Botão Esquerdo: Inspeciona itens/npcs/terreno. Botão Direito: Ataca, conversa, caminha ou abre recipientes.`,
    `[Dica] Diga "exura" ou "exori flam" no chat para conjurar magias canônicas!`
  ]);

  // Dialogue with Captain Kenneth
  const [npcDialogue, setNpcDialogue] = useState<string | null>(null);

  // --- INTEGRATE AUTHORITATIVE BACKEND ENGINE ---
  useEffect(() => {
    progressionEngine.initialize(activeChar, (updatedChar, event, details) => {
      // Keep UI state variables synchronized with server-authoritative engine
      setPlayerLevel(updatedChar.level);
      setPlayerXp(updatedChar.xp);
      setPlayerGold(updatedChar.gold);
      setPlayerInventory(updatedChar.inventory || []);
      setQuestState(updatedChar.quest_state);
      setPlayerPos(updatedChar.position);

      // Invoke the parent update handler to store the canonical profile
      onUpdateCharacter(updatedChar);

      if (syncCoins) syncCoins(updatedChar.gold);
      if (syncInventory) syncInventory(updatedChar.inventory || []);
      if (syncLevel) syncLevel(updatedChar.level);

      // Log the authoritative event to the REST API debugger
      if (event) {
        addApiLog(
          "POST",
          `/api/progression/event/${event.toLowerCase()}`,
          200,
          8,
          JSON.stringify({ details, level: updatedChar.level, xp: updatedChar.xp, skills: updatedChar.skills })
        );
      }
    });
  }, [activeChar]);
  
  // Spell CD states
  const [exuraCD, setExuraCD] = useState(false);
  const [strikeCD, setStrikeCD] = useState(false);

  // Auto-attack tick timer
  const attackTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Generate Map tiles once
  const [mapTiles, setMapTiles] = useState<MapTile[][]>([]);

  // Sound/VFX toggle
  const [visualEffects, setVisualEffects] = useState(true);

  // Is player in sewer map section?
  const isInSewers = playerPos.y >= 12;

  // Track the rat kills for the tutorial quest
  const activeTutorialQuest = questState.activeQuests.find(q => q.quest_id === "quest_tutorial_sewer_rats");
  const ratKilledCount = activeTutorialQuest?.objectives.find(o => o.target_id === "sewer_rat")?.current_qty || 0;

  // --- INITIALIZATION ---
  useEffect(() => {
    // Generate static 20x20 tile layout
    const tiles: MapTile[][] = [];
    for (let y = 0; y < MAP_HEIGHT; y++) {
      const row: MapTile[] = [];
      for (let x = 0; x < MAP_WIDTH; x++) {
        let type: TileType = "FLOOR_STONE";
        let name = "Piso de Pedra";

        // Boundary walls
        if (x === 0 || x === MAP_WIDTH - 1 || y === 0 || y === MAP_HEIGHT - 1) {
          type = "WALL";
          name = "Muralha de Ironhold";
        }
        // Sewer separation wall (with a ladder access)
        else if (y === 11) {
          if (x === 10) {
            type = "LADDER_DOWN";
            name = "Escada para os Esgotos";
          } else {
            type = "WALL";
            name = "Parede Estrutural";
          }
        }
        // Sewer area
        else if (y >= 12) {
          if (x === 10 && y === 12) {
            type = "LADDER_UP";
            name = "Escada de Retorno";
          } else {
            type = "FLOOR_SEWER";
            name = "Lodo do Esgoto";
            
            // Random sewer obstacles/water streams
            if ((x === 4 || x === 5 || x === 14 || x === 15) && y !== 14) {
              type = "WATER";
              name = "Água Contaminada";
            }
            if ((x === 9 || x === 11) && y === 16) {
              type = "WALL";
              name = "Pilastra do Esgoto";
            }
            // Boss gate
            if (x === 18 && y === 15) {
              type = "LOCKED_GATE";
              name = "Portão dos Esgotos Profundos";
            }
          }
        }
        // Town obstacles
        else {
          if (y === 4 && (x === 4 || x === 5 || x === 14 || x === 15)) {
            type = "WALL";
            name = "Estátua dos Heróis Antigos";
          }
        }

        row.push({ x, y, type, name });
      }
      tiles.push(row);
    }
    setMapTiles(tiles);

    // Spawn initial monsters in the sewers
    const initialMonsters: Monster[] = [
      { id: "rat-1", name: "Sewer Rat", type: "sewer_rat", x: 3, y: 14, hp: 320, maxHp: 320, damage: 5, expReward: 12, isDead: false, respawnTime: 0 },
      { id: "rat-2", name: "Sewer Rat", type: "sewer_rat", x: 7, y: 15, hp: 320, maxHp: 320, damage: 5, expReward: 12, isDead: false, respawnTime: 0 },
      { id: "rat-3", name: "Sewer Rat", type: "sewer_rat", x: 12, y: 17, hp: 320, maxHp: 320, damage: 6, expReward: 12, isDead: false, respawnTime: 0 },
      { id: "rat-4", name: "Sewer Rat", type: "sewer_rat", x: 16, y: 14, hp: 320, maxHp: 320, damage: 5, expReward: 12, isDead: false, respawnTime: 0 },
      { id: "rat-5", name: "Sewer Rat", type: "sewer_rat", x: 6, y: 18, hp: 320, maxHp: 320, damage: 5, expReward: 12, isDead: false, respawnTime: 0 },
      { id: "rat-6", name: "Sewer Rat", type: "sewer_rat", x: 14, y: 18, hp: 320, maxHp: 320, damage: 6, expReward: 12, isDead: false, respawnTime: 0 },
      // Elite Slime deeper down
      { id: "slime-1", name: "Sewer Slime", type: "sewer_slime", x: 18, y: 18, hp: 450, maxHp: 450, damage: 9, expReward: 25, isDead: false, respawnTime: 0 }
    ];
    setMonsters(initialMonsters);

    // Sync state values
    setPlayerHp(initialPlayerHp);
    setPlayerMana(initialPlayerMana);
    setPlayerLevel(activeChar.level);
    setPlayerGold(activeChar.gold);
    setPlayerXp(activeChar.xp);
    setPlayerInventory(activeChar.inventory || []);
    setPlayerEquipment(activeChar.equipment || []);
    setQuestState(activeChar.quest_state);

    // Initial fog of war explore around spawn coordinates
    const initialExploreRadius = 5;
    const initialExplored: Record<string, boolean> = {};
    for (let dy = -initialExploreRadius; dy <= initialExploreRadius; dy++) {
      for (let dx = -initialExploreRadius; dx <= initialExploreRadius; dx++) {
        const tx = 10 + dx;
        const ty = 5 + dy;
        if (tx >= 0 && tx < MAP_WIDTH && ty >= 0 && ty < MAP_HEIGHT) {
          initialExplored[`${tx},${ty}`] = true;
        }
      }
    }
    setExploredTiles(initialExplored);

    // Log load
    addApiLog("GET", "/api/world/map-data", 200, 14, "Carregando coordenadas espaciais 20x20");
  }, [activeChar]);

  // --- AUTOMATIC REGENERATION & MONSTER MOVEMENT & SPAWNING ---
  useEffect(() => {
    const gameLoop = setInterval(() => {
      // 1. Natural HP & MP regeneration (out of combat)
      setPlayerHp(prev => Math.min(120, prev + (targetId ? 0 : 2)));
      setPlayerMana(prev => Math.min(40, prev + (targetId ? 1 : 2)));

      // 2. Monster passive movement & respawn ticks
      setMonsters(prev =>
        prev.map(mon => {
          if (mon.isDead) {
            if (mon.respawnTime <= 1) {
              // Respawn rat
              addLog(`[Mundo] Um novo ${mon.name} surgiu nas sombras.`);
              return { ...mon, isDead: false, hp: mon.maxHp, respawnTime: 0 };
            }
            return { ...mon, respawnTime: mon.respawnTime - 1 };
          }

          // Random move if not dead (30% chance per tick to maintain stability)
          if (Math.random() < 0.3) {
            const directions = [
              { dx: 0, dy: 1 },
              { dx: 0, dy: -1 },
              { dx: 1, dy: 0 },
              { dx: -1, dy: 0 }
            ];
            const dir = directions[Math.floor(Math.random() * directions.length)];
            const nx = mon.x + dir.dx;
            const ny = mon.y + dir.dy;

            // Check boundaries & walkability in sewers (only sewers)
            if (ny >= 12 && ny < MAP_HEIGHT - 1 && nx > 0 && nx < MAP_WIDTH - 1) {
              const destType = mapTiles[ny]?.[nx]?.type;
              if (destType && destType !== "WALL" && destType !== "WATER" && destType !== "LOCKED_GATE") {
                return { ...mon, x: nx, y: ny };
              }
            }
          }
          return mon;
        })
      );
    }, 4000);

    return () => clearInterval(gameLoop);
  }, [targetId, mapTiles]);

  // --- SPATIAL ATTACK CORE (AUTO ATTACK CYCLE) ---
  useEffect(() => {
    if (targetId) {
      const monster = monsters.find(m => m.id === targetId);
      if (!monster || monster.isDead) {
        setTargetId(null);
        return;
      }

      // Auto-attack triggers every 1.8 seconds (Tibia's typical combat clock)
      attackTimerRef.current = setInterval(() => {
        handlePlayerAttack();
      }, 1800);
    } else {
      if (attackTimerRef.current) {
        clearInterval(attackTimerRef.current);
        attackTimerRef.current = null;
      }
    }

    return () => {
      if (attackTimerRef.current) {
        clearInterval(attackTimerRef.current);
      }
    };
  }, [targetId, monsters, playerPos, playerEquipment, playerHp]);

  // --- FLOATING TEXT CLEANUP ---
  useEffect(() => {
    if (floatingTexts.length > 0) {
      const timer = setTimeout(() => {
        setFloatingTexts(prev => prev.slice(1));
      }, 1500);
      return () => clearTimeout(timer);
    }
  }, [floatingTexts]);

  // --- FOG OF WAR EXPANSION ---
  useEffect(() => {
    const exploreRadius = 5;
    setExploredTiles(prev => {
      const next = { ...prev };
      for (let dy = -exploreRadius; dy <= exploreRadius; dy++) {
        for (let dx = -exploreRadius; dx <= exploreRadius; dx++) {
          const tx = playerPos.x + dx;
          const ty = playerPos.y + dy;
          if (tx >= 0 && tx < MAP_WIDTH && ty >= 0 && ty < MAP_HEIGHT) {
            next[`${tx},${ty}`] = true;
          }
        }
      }
      return next;
    });
  }, [playerPos]);

  // --- HELPERS ---
  const addLog = (msg: string) => {
    setGameLogs(prev => [...prev, msg].slice(-40));
  };

  const addFloatingText = (tileX: number, tileY: number, text: string, color: string) => {
    const id = `float-${Math.random().toString(36).substring(2, 11)}`;
    setFloatingTexts(prev => [...prev, { id, x: tileX, y: tileY, text, color }]);
  };

  // --- MOVEMENT CONTROLS ---
  const attemptMove = (dx: number, dy: number) => {
    const nx = playerPos.x + dx;
    const ny = playerPos.y + dy;

    if (nx < 0 || nx >= MAP_WIDTH || ny < 0 || ny >= MAP_HEIGHT) return;

    const destTile = mapTiles[ny]?.[nx];
    if (!destTile) return;

    if (destTile.type === "WALL" || destTile.type === "WATER") {
      addLog(`[Sistema] Bloqueado: você não pode caminhar sobre ${destTile.name}.`);
      return;
    }

    if (destTile.type === "LOCKED_GATE") {
      if (playerLevel < 2) {
        addLog(`[Sistema] Portão Trancado: Requer Nível 2 e autorização do Capitão.`);
        return;
      }
      addLog(`[Mundo] Você destrancou o portão dos esgotos profundos!`);
    }

    if (destTile.type === "LADDER_DOWN" || destTile.type === "LADDER_UP") {
      progressionEventBus.emit("ON_LADDER_INTERACTION", {
        playerId: activeChar.id,
        timestamp: Date.now(),
        ladderType: destTile.type,
        fromX: nx,
        fromY: ny
      });
      addLog(
        destTile.type === "LADDER_DOWN"
          ? "⚔️ [Mundo] Você desceu para a Escuridão dos Esgotos de Ironhold."
          : "🛡️ [Mundo] Você subiu de volta para o Bastião Seguro de Ironhold."
      );
      addApiLog(
        "POST",
        "/api/session/teleport",
        200,
        18,
        JSON.stringify({
          to: destTile.type === "LADDER_DOWN" ? "Sewers" : "Town",
          coords: { x: 10, y: destTile.type === "LADDER_DOWN" ? 13 : 10 }
        })
      );
      return;
    }

    setPlayerPos({ x: nx, y: ny });

    // Close dialog if leaving Kenneth range
    if (Math.abs(nx - 10) > 1 || Math.abs(ny - 3) > 1) {
      setNpcDialogue(null);
    }
  };

  // Capture Keyboard Movement + Spell Shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (document.activeElement?.tagName === "INPUT" || document.activeElement?.tagName === "TEXTAREA") {
        return;
      }

      switch (e.key.toLowerCase()) {
        case "w":
        case "arrowup":
          e.preventDefault();
          attemptMove(0, -1);
          break;
        case "s":
        case "arrowdown":
          e.preventDefault();
          attemptMove(0, 1);
          break;
        case "a":
        case "arrowleft":
          e.preventDefault();
          attemptMove(-1, 0);
          break;
        case "d":
        case "arrowright":
          e.preventDefault();
          attemptMove(1, 0);
          break;
        case "1":
          castExura();
          break;
        case "2":
          castStrike();
          break;
        case "escape":
          e.preventDefault();
          if (targetId) {
            setTargetId(null);
            addLog("[Mira] Ataque cancelado.");
          }
          break;
        default:
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [playerPos, mapTiles, targetId, monsters, playerMana, playerHp, playerLevel]);

  // --- MOUSE SEMANTIC EVENT IMPLEMENTATIONS ---
  
  // 1. LEFT CLICK (SHIFT+LOOK / INSPECTION LOG)
  const handleMapTileLeftClick = (x: number, y: number) => {
    const mon = monsters.find(m => m.x === x && m.y === y && !m.isDead);
    if (mon) {
      addLog(`👁️ Você vê ${mon.name}. (HP: ${mon.hp}/${mon.maxHp}).`);
      return;
    }

    if (x === 10 && y === 3) {
      addLog(`👁️ Você vê Capitão Kenneth. O honorável guardião de Ironhold.`);
      return;
    }

    const corpse = corpses.find(c => c.x === x && c.y === y && !c.looted);
    if (corpse) {
      addLog(`👁️ Você vê o ${corpse.name}.`);
      return;
    }

    const tile = mapTiles[y]?.[x];
    if (tile) {
      addLog(`👁️ Você vê ${tile.name}.`);
    }
  };

  // 2. RIGHT CLICK (PRIMARY CONTEXT-SENSITIVE INTERACTION)
  const handleMapTileRightClick = (e: React.MouseEvent, x: number, y: number) => {
    e.preventDefault(); // crucial to prevent native context menus

    // A. Is there a monster on this tile?
    const monster = monsters.find(m => m.x === x && m.y === y && !m.isDead);
    if (monster) {
      setTargetId(monster.id);
      addLog(`⚔️ [Target] Alvo travado: ${monster.name}. Iniciou auto-ataque.`);
      return;
    }

    // B. Is there Captain Kenneth?
    if (x === 10 && y === 3) {
      interactWithKenneth();
      return;
    }

    // C. Is there an unlooted corpse?
    const corpse = corpses.find(c => c.x === x && c.y === y && !c.looted);
    if (corpse) {
      setOpenContainer({
        name: corpse.name,
        items: corpse.items,
        corpseId: corpse.id
      });
      addLog(`🎒 [Loot] Abrindo recipiente de saque: ${corpse.name}.`);
      return;
    }

    // D. Check for ladders, doors, or openable locks
    const tile = mapTiles[y]?.[x];
    if (tile) {
      if (tile.type === "LADDER_DOWN" || tile.type === "LADDER_UP") {
        // Walk towards ladder cell
        const dx = Math.sign(x - playerPos.x);
        const dy = Math.sign(y - playerPos.y);
        attemptMove(dx, dy);
        return;
      }
      
      if (tile.type === "LOCKED_GATE") {
        if (playerLevel < 2) {
          addLog(`[Mundo] O portão trancado impede sua passagem.`);
        } else {
          addLog(`[Mundo] O portão se abre suavemente.`);
        }
        return;
      }
    }

    // E. Default empty ground -> walk to tile (1 step)
    const dx = Math.sign(x - playerPos.x);
    const dy = Math.sign(y - playerPos.y);
    attemptMove(dx, dy);
  };

  // --- CAPTAIN KENNETH NPC DIALOGUE SYSTEM ---
  const interactWithKenneth = () => {
    const dist = Math.abs(playerPos.x - 10) + Math.abs(playerPos.y - 3);
    if (dist > 1.5) {
      addLog("[NPC] Capitão Kenneth: 'Aproxime-se, guerreiro! Não fale comigo de tão longe.'");
      return;
    }

    if (playerLevel >= 2) {
      setNpcDialogue(
        "Capitão Kenneth: 'Excelente trabalho nos esgotos, recruta! Você agora é livre para explorar as ruínas e expandir suas vocações. O continente precisa de heróis habilidosos como você!'"
      );
      addLog("[NPC] Conversando com Capitão Kenneth.");
      return;
    }

    if (ratKilledCount >= 10) {
      setNpcDialogue(
        "Capitão Kenneth: 'Estupendo! Você varreu aqueles ratos imundos das galerias de Ironhold. Aqui está sua recompensa canônica de treinamento militar. Você provou ser forte!'"
      );
      addLog("[NPC] Conversando com Capitão Kenneth. Pronto para entregar a Missão!");
    } else {
      setNpcDialogue(
        `Capitão Kenneth: 'A guarnição está sitiada por Sewer Rats imundos nos bueiros ao sul! Desça pelas escadas e elimine pelo menos 10 Ratos de Esgoto. Abates atuais: ${ratKilledCount}/10'`
      );
      addLog("[NPC] Conversando com Capitão Kenneth. Missão ativa.");
    }
  };

  const handleDeliverQuest = () => {
    if (ratKilledCount < 10) return;

    // Complete the quest authoritatively in the backend Progression Engine
    QuestSystem.completeQuest(activeChar.id, "quest_tutorial_sewer_rats", 110, 15);

    // Award the 3 potions of healing via the item pickup event
    progressionEventBus.emit("ON_ITEM_PICKUP", {
      playerId: activeChar.id,
      timestamp: Date.now(),
      itemId: "potion_heal",
      itemName: "Poção de Cura",
      qty: 3,
      value: 15,
      type: "consumable"
    });

    setNpcDialogue("Capitão Kenneth: 'Parabéns pela formatura! Você atingiu o Nível 2 e obteve o título de Guardião Oficial de Ironhold. Vá com a luz!'");
    addLog("🎉 [Quest Concluída] Entregou a missão ao Capitão Kenneth!");
    addLog("🎖️ Recompensas: +110 XP, +15 Moedas de Bronze, +3 Poções de Cura.");
    addLog("🆙 Você subiu para o Nível 2! Suas habilidades canônicas de combate foram expandidas!");

    addApiLog("POST", "/api/quests/complete", 200, 48, JSON.stringify({
      char_id: activeChar.id,
      quest_id: "quest_tutorial_sewer_rats",
      level_reached: 2
    }));
  };

  // --- COMBAT INTERACTION CORE (AUTOATTACK STEP) ---
  const handlePlayerAttack = () => {
    if (!targetId) return;

    let activeMonster: Monster | null = null;
    let targetIdx = -1;
    for (let i = 0; i < monsters.length; i++) {
      if (monsters[i].id === targetId) {
        activeMonster = monsters[i];
        targetIdx = i;
        break;
      }
    }

    if (!activeMonster || activeMonster.isDead) {
      setTargetId(null);
      return;
    }

    // Range checks
    const dist = Math.abs(playerPos.x - activeMonster.x) + Math.abs(playerPos.y - activeMonster.y);
    let maxRange = 1; // Default Novice melee fists
    let weaponName = "Ataque Desarmado";
    const hasSword = playerEquipment.some(eq => eq.id === "sword_basic");

    if (hasSword) {
      maxRange = 1;
      weaponName = "Espada Básica de Treinamento";
    }

    if (dist > maxRange) {
      addLog(`[Combate] Fora de alcance! Aproxime-se para bater com ${weaponName}. Distância: ${dist} blocos.`);
      return;
    }

    // Combat Damage
    let baseDmg = Math.floor(Math.random() * 10) + 16; // 16-25 physical damage
    if (hasSword) baseDmg += 6; // sword gets bonus damage!

    const finalMonsterHp = Math.max(0, activeMonster.hp - baseDmg);

    addFloatingText(activeMonster.x, activeMonster.y, `-${baseDmg}`, "text-red-500 font-extrabold");
    addLog(`⚔️ [Auto-ataque] Desferiu ${baseDmg} de dano no ${activeMonster.name}. (HP: ${finalMonsterHp}/${activeMonster.maxHp})`);

    // Process weapon strike and skill usage in backend CombatEngine
    CombatEngine.processWeaponStrike(activeChar.id, baseDmg, hasSword ? "sword" : "melee", "Physical");

    // Monster retaliation inside adjacent reach
    if (dist === 1) {
      const blockSuccess = Math.random() < 0.65 && playerEquipment.some(eq => eq.id === "shield_wooden");
      const netDmg = blockSuccess ? 2 : activeMonster.damage;
      const finalPlayerHp = Math.max(0, playerHp - netDmg);

      setPlayerHp(finalPlayerHp);
      addFloatingText(playerPos.x, playerPos.y, `-${netDmg}`, blockSuccess ? "text-amber-400 font-bold" : "text-orange-500");
      addLog(
         blockSuccess
           ? `🛡️ [Active Defense] Bloqueio tático! Absorveu impacto com escudo de madeira. Sofreu -${netDmg} de dano.`
           : `💥 O ${activeMonster.name} mordeu você! Sofreu -${netDmg} de dano físico.`
      );

      // Process shielding block in backend CombatEngine
      if (blockSuccess) {
        CombatEngine.processShieldBlock(activeChar.id, activeMonster.damage - 2);
      }

      if (finalPlayerHp <= 0) {
        // Spiritual resurrection in Safe Town center
        setPlayerHp(80);
        setPlayerMana(20);
        setPlayerPos({ x: 10, y: 5 });
        setTargetId(null);
        setOpenContainer(null);
        addLog("💀 Você foi derrotado! Você acordou no centro de cura espiritual de Ironhold Bastion.");
        addApiLog("POST", "/api/player/respawn", 200, 31, JSON.stringify({ char_id: activeChar.id, death: "sewer_rat_combat" }));
        return;
      }
    }

    // Update monster or spawn its corpse container
    const updatedMonsters = [...monsters];
    if (finalMonsterHp <= 0) {
      updatedMonsters[targetIdx] = {
        ...activeMonster,
        hp: 0,
        isDead: true,
        respawnTime: 4 // Stay dead for 4 loops (16s)
      };

      setTargetId(null);
      addLog(`🏆 Vitória! O ${activeMonster.name} foi derrotado. Obteve +${activeMonster.expReward} EXP.`);

      // Trigger authoritative monster kill on TibiaBattleSystem / ProgressionEngine
      TibiaBattleSystem.triggerMonsterKill(
        activeChar.id,
        activeMonster.id,
        activeMonster.type,
        activeMonster.name,
        activeMonster.expReward
      );

      // SPAWN CORPSE CONTAINER WITH EXTRA EXTRACTABLE ITEMS
      const corpseId = `corpse-${Date.now()}-${Math.random().toString(36).substring(2, 6)}`;
      const lootItems: InventoryItem[] = [
        { id: "bronze_coin", name: "Moeda de Bronze", qty: Math.floor(Math.random() * 4) + 4, value: 1, type: "currency" }
      ];
      if (Math.random() < 0.7) {
        lootItems.push({ id: "rat_tail", name: "Cauda de Rato", qty: 1, value: 2, type: "material" });
      }
      if (Math.random() < 0.2) {
        lootItems.push({ id: "potion_heal", name: "Poção de Cura", qty: 1, value: 15, type: "consumable" });
      }

      const newCorpse: MapCorpse = {
        id: corpseId,
        x: activeMonster.x,
        y: activeMonster.y,
        name: `Corpo de ${activeMonster.name}`,
        items: lootItems,
        looted: false
      };
      setCorpses(prev => [...prev, newCorpse]);
      addLog(`🧪 O corpo de ${activeMonster.name} jaz no chão. Clique com o botão direito para abri-lo.`);

      addApiLog("POST", "/api/player/combat-event", 200, 15, JSON.stringify({
        killed: activeMonster.name,
        xp_awarded: activeMonster.expReward
      }));
    } else {
      updatedMonsters[targetIdx] = {
        ...activeMonster,
        hp: finalMonsterHp
      };
    }

    setMonsters(updatedMonsters);
  };

  // --- CAST SPELLS ---
  const castExura = () => {
    if (exuraCD) return;
    if (playerMana < 10) {
      addLog("❌ Mana insuficiente! Exura consome 10 de MP.");
      return;
    }

    setExuraCD(true);
    setPlayerMana(prev => prev - 10);
    setPlayerHp(prev => Math.min(120, prev + 55));
    addFloatingText(playerPos.x, playerPos.y, `+55`, "text-teal-400 font-extrabold");
    addLog("✨ Spell: [Exura (Cura)] - Regenerou +55 de HP.");

    // Authoritative magic level progression
    SpellSystem.castSpell(activeChar.id, "exura", 10, "holy");

    setTimeout(() => setExuraCD(false), 2000); // 2s lock
    addApiLog("POST", "/api/player/cast-spell", 200, 10, JSON.stringify({ spell: "exura", cost: 10 }));
  };

  const castStrike = () => {
    if (strikeCD) return;
    if (!targetId) {
      addLog("❌ Sem alvo! Use o botão direito para travar em uma criatura antes de lançar Exori Flam.");
      return;
    }

    const activeMonster = monsters.find(m => m.id === targetId);
    if (!activeMonster || activeMonster.isDead) return;

    if (playerMana < 15) {
      addLog("❌ Mana insuficiente! Exori Flam consome 15 de MP.");
      return;
    }

    setStrikeCD(true);
    setPlayerMana(prev => prev - 15);

    const dmg = Math.floor(Math.random() * 20) + 38; // 38-57 fire damage
    const finalMonsterHp = Math.max(0, activeMonster.hp - dmg);

    addFloatingText(activeMonster.x, activeMonster.y, `-${dmg} 🔥`, "text-amber-500 font-extrabold");
    addLog(`🔥 Spell: [Exori Flam] - Lançou chamas ardentes no ${activeMonster.name}! Sofreu -${dmg} de dano.`);

    // Authoritative magic level and fire elemental affinity progression
    SpellSystem.castSpell(activeChar.id, "exori_flam", 15, "fire");
    SpellSystem.applyElementalSpellDamage(activeChar.id, dmg, "fire");

    // Update HP directly inside state loop
    setMonsters(prev =>
      prev.map(mon => {
        if (mon.id === targetId) {
          return { ...mon, hp: finalMonsterHp };
        }
        return mon;
      })
    );

    // If killed by strike, trigger autoattack cleanup
    if (finalMonsterHp <= 0) {
      setTimeout(() => handlePlayerAttack(), 50);
    }

    setTimeout(() => setStrikeCD(false), 3000); // 3s lock
    addApiLog("POST", "/api/player/cast-spell", 200, 12, JSON.stringify({ spell: "exori_flam", cost: 15 }));
  };

  // --- LOOT SELECTION FROM OPEN CONTAINER ---
  const lootItemFromCorpse = (itemId: string) => {
    if (!openContainer) return;
    const corpse = corpses.find(c => c.id === openContainer.corpseId);
    if (!corpse) return;

    const itemIdx = corpse.items.findIndex(i => i.id === itemId);
    if (itemIdx === -1) return;

    const item = corpse.items[itemIdx];

    // Transfer loot
    let updatedInv = [...playerInventory];
    let nextGold = playerGold;
    if (item.type === "currency" && item.id === "bronze_coin") {
      nextGold += item.qty;
      setPlayerGold(nextGold);
      if (syncCoins) syncCoins(nextGold);
    } else {
      const existingIdx = updatedInv.findIndex(i => i.id === item.id);
      if (existingIdx !== -1) {
        updatedInv[existingIdx] = { ...updatedInv[existingIdx], qty: updatedInv[existingIdx].qty + item.qty };
      } else {
        updatedInv.push({ ...item });
      }
      setPlayerInventory(updatedInv);
      if (syncInventory) syncInventory(updatedInv);
    }

    // Remove from container list
    const updatedCorpseItems = corpse.items.filter(i => i.id !== itemId);
    corpse.items = updatedCorpseItems;

    if (updatedCorpseItems.length === 0) {
      corpse.looted = true;
      setOpenContainer(null);
      addLog(`🎒 [Loot] Corpo completamente saqueado.`);
    } else {
      setOpenContainer({
        ...openContainer,
        items: updatedCorpseItems
      });
    }

    addLog(`[Loot] Recolheu: ${item.qty}x ${item.name}.`);

    triggerCharacterUpdate({
      ...activeChar,
      gold: nextGold,
      inventory: updatedInv
    });
  };

  // --- EQUIP/UNEQUIP UTILITIES ---
  const handleEquipItem = (item: InventoryItem) => {
    let slot = "";
    if (item.id === "sword_basic") {
      slot = "Weapon";
    } else if (item.id === "shield_wooden") {
      slot = "Offhand";
    } else {
      addLog(`[Equipamento] ${item.name} não pode ser colocado nos slots ativos.`);
      return;
    }

    let updatedEquipment = [...playerEquipment];
    let updatedInv = [...playerInventory];

    // Check if slot holds gear
    const existing = updatedEquipment.find(eq => eq.slot === slot);
    if (existing) {
      // Return to inventory
      const idx = updatedInv.findIndex(i => i.id === existing.id);
      if (idx !== -1) {
        updatedInv[idx].qty += 1;
      } else {
        updatedInv.push({ id: existing.id, name: existing.name, qty: 1, value: existing.value, type: "equipment" });
      }
      updatedEquipment = updatedEquipment.filter(eq => eq.slot !== slot);
    }

    // Deduct new item from inventory
    const newIdx = updatedInv.findIndex(i => i.id === item.id);
    if (newIdx !== -1) {
      if (updatedInv[newIdx].qty > 1) {
        updatedInv[newIdx].qty -= 1;
      } else {
        updatedInv.splice(newIdx, 1);
      }
    }

    updatedEquipment.push({ slot, id: item.id, name: item.name, value: item.value });

    setPlayerEquipment(updatedEquipment);
    setPlayerInventory(updatedInv);
    if (syncInventory) syncInventory(updatedInv);

    addLog(`[Equipamento] Equipou ${item.name} no slot ${slot}.`);

    triggerCharacterUpdate({
      ...activeChar,
      equipment: updatedEquipment,
      inventory: updatedInv
    });
  };

  const handleUnequipItem = (slot: string) => {
    let updatedEquipment = [...playerEquipment];
    const item = updatedEquipment.find(eq => eq.slot === slot);
    if (!item) return;

    let updatedInv = [...playerInventory];
    const existingIdx = updatedInv.findIndex(i => i.id === item.id);
    if (existingIdx !== -1) {
      updatedInv[existingIdx].qty += 1;
    } else {
      updatedInv.push({ id: item.id, name: item.name, qty: 1, value: item.value, type: "equipment" });
    }

    updatedEquipment = updatedEquipment.filter(eq => eq.slot !== slot);

    setPlayerEquipment(updatedEquipment);
    setPlayerInventory(updatedInv);
    if (syncInventory) syncInventory(updatedInv);

    addLog(`[Equipamento] Removeu ${item.name} do slot ${slot}.`);

    triggerCharacterUpdate({
      ...activeChar,
      equipment: updatedEquipment,
      inventory: updatedInv
    });
  };

  const handleUseHealPotion = () => {
    const potIdx = playerInventory.findIndex(i => i.id === "potion_heal" && i.qty > 0);
    if (potIdx === -1) {
      alert("Você não possui poções de cura em sua mochila!");
      return;
    }

    let inv = [...playerInventory];
    if (inv[potIdx].qty > 1) {
      inv[potIdx] = { ...inv[potIdx], qty: inv[potIdx].qty - 1 };
    } else {
      inv.splice(potIdx, 1);
    }

    setPlayerInventory(inv);
    setPlayerHp(prev => Math.min(120, prev + 50));
    addFloatingText(playerPos.x, playerPos.y, `+50`, "text-teal-400 font-extrabold");
    addLog("🧪 Consumiu uma Poção de Cura. Regenerou +50 de HP.");

    if (syncInventory) syncInventory(inv);

    triggerCharacterUpdate({
      ...activeChar,
      inventory: inv
    });

    addApiLog("POST", "/api/player/use-item", 200, 16, JSON.stringify({ item_id: "potion_heal" }));
  };

  const handleDropItem = (itemId: string) => {
    let inv = [...playerInventory];
    const idx = inv.findIndex(i => i.id === itemId);
    if (idx === -1) return;

    const item = inv[idx];
    if (item.qty > 1) {
      inv[idx] = { ...inv[idx], qty: item.qty - 1 };
    } else {
      inv.splice(idx, 1);
    }

    setPlayerInventory(inv);
    addLog(`[Mochila] Descartou 1x ${item.name} no chão.`);
    if (syncInventory) syncInventory(inv);

    triggerCharacterUpdate({
      ...activeChar,
      inventory: inv
    });
  };

  const triggerCharacterUpdate = (updatedChar: CharacterProfile) => {
    onUpdateCharacter(updatedChar);
  };

  // --- RENDER GAME CELL STYLING ---
  const renderMapGrid = () => {
    if (mapTiles.length === 0) return null;

    // Viewport camera is exactly 11 columns by 13 rows centered on playerPos
    const viewRows = 11;
    const viewCols = 13;
    const halfY = Math.floor(viewRows / 2);
    const halfX = Math.floor(viewCols / 2);

    const viewportTiles = [];
    for (let r = -halfY; r <= halfY; r++) {
      const tileY = playerPos.y + r;
      for (let c = -halfX; c <= halfX; c++) {
        const tileX = playerPos.x + c;

        if (tileX < 0 || tileX >= MAP_WIDTH || tileY < 0 || tileY >= MAP_HEIGHT) {
          viewportTiles.push({ x: tileX, y: tileY, isOOB: true, tile: null });
        } else {
          viewportTiles.push({ x: tileX, y: tileY, isOOB: false, tile: mapTiles[tileY][tileX] });
        }
      }
    }

    return (
      <div
        className="grid bg-slate-950 border-4 border-slate-900 rounded-xl overflow-hidden shadow-2xl relative select-none"
        onContextMenu={(e) => e.preventDefault()} // Block browser right-clicks
        style={{
          gridTemplateColumns: `repeat(${viewCols}, minmax(0, 1fr))`,
          aspectRatio: `${viewCols}/${viewRows}`
        }}
      >
        {viewportTiles.map(({ x, y, isOOB, tile }, index) => {
          if (isOOB || !tile) {
            return <div key={index} className="bg-black border border-black" />;
          }

          // Spatial light calculations (Player holds a dynamic torch)
          const distFromPlayer = Math.sqrt(Math.pow(x - playerPos.x, 2) + Math.pow(y - playerPos.y, 2));
          let lightFactor = 1.0;
          if (isInSewers) {
            lightFactor = Math.max(0.12, 1.0 - (distFromPlayer * 0.18));
          } else {
            lightFactor = Math.max(0.35, 1.0 - (distFromPlayer * 0.08));
          }

          let tileBg = "bg-slate-900";
          let tileColor = "text-slate-500";

          if (tile.type === "WALL") {
            tileBg = isInSewers ? "bg-stone-900 border-slate-950" : "bg-slate-800 border-slate-700";
            tileColor = isInSewers ? "text-stone-700" : "text-slate-600";
          } else if (tile.type === "FLOOR_STONE") {
            tileBg = "bg-slate-950 border-slate-900/60";
            tileColor = "text-slate-700";
          } else if (tile.type === "FLOOR_SEWER") {
            tileBg = "bg-zinc-950 border-emerald-950/10";
            tileColor = "text-emerald-950/20";
          } else if (tile.type === "WATER") {
            tileBg = "bg-cyan-950 border-cyan-900/10 animate-pulse";
            tileColor = "text-cyan-900/30";
          } else if (tile.type === "LADDER_DOWN") {
            tileBg = "bg-amber-950/40 border-amber-900/10";
            tileColor = "text-amber-500 animate-pulse";
          } else if (tile.type === "LADDER_UP") {
            tileBg = "bg-violet-950/40 border-violet-900/10";
            tileColor = "text-violet-400 animate-pulse";
          } else if (tile.type === "LOCKED_GATE") {
            tileBg = "bg-amber-950/80 border-amber-900";
            tileColor = "text-amber-300";
          }

          const isPlayerHere = playerPos.x === x && playerPos.y === y;
          const isKennethHere = x === 10 && y === 3;
          const activeMonster = monsters.find(m => m.x === x && m.y === y && !m.isDead);
          const activeCorpse = corpses.find(c => c.x === x && c.y === y && !c.looted);
          const isTargeted = activeMonster && targetId === activeMonster.id;

          const shadowStyle = {
            filter: `brightness(${lightFactor * 100}%) contrast(${95 + (lightFactor * 10)}%)`
          };

          return (
            <div
              key={index}
              onClick={() => handleMapTileLeftClick(x, y)}
              onContextMenu={(e) => handleMapTileRightClick(e, x, y)}
              style={shadowStyle}
              className={`relative border-[1px] border-slate-950/30 flex items-center justify-center transition-all duration-150 cursor-pointer group ${tileBg}`}
            >
              {/* Tile coordinate labels on mouse-over */}
              <span className="absolute bottom-0 right-0 text-[6.5px] font-mono text-slate-800 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
                {x},{y}
              </span>

              {/* Standard floor background character */}
              {!isPlayerHere && !isKennethHere && !activeMonster && !activeCorpse && (
                <div className={`${tileColor} font-bold text-xs pointer-events-none`}>
                  {tile.type === "WALL" ? "█" : tile.type === "WATER" ? "≈" : tile.type === "LADDER_DOWN" ? "▼" : tile.type === "LADDER_UP" ? "▲" : "."}
                </div>
              )}

              {/* Loot Corpse Rendering */}
              {activeCorpse && !activeMonster && !isPlayerHere && (
                <div className="relative flex flex-col items-center justify-center">
                  <span className="text-sm select-none animate-pulse">🦴</span>
                  <span className="absolute -bottom-4 text-[7px] font-bold text-amber-500 bg-slate-950/90 px-1 border border-amber-950/20 rounded font-mono whitespace-nowrap z-10 pointer-events-none">
                    Corpo
                  </span>
                </div>
              )}

              {/* NPC Kenneth */}
              {isKennethHere && (
                <div className="relative flex flex-col items-center justify-center">
                  <div className="w-7 h-7 rounded-full bg-blue-600 border-2 border-blue-400 flex items-center justify-center shadow animate-pulse">
                    <User className="w-4 h-4 text-white" />
                  </div>
                  {ratKilledCount < 10 ? (
                    <span className="absolute -top-3 text-[10px] bg-amber-500 text-slate-950 px-1 border border-slate-950 rounded-full font-bold animate-bounce scale-90">
                      !
                    </span>
                  ) : (
                    <span className="absolute -top-3 text-[10px] bg-emerald-500 text-white px-1 border border-slate-950 rounded-full font-bold animate-bounce">
                      ✓
                    </span>
                  )}
                  <span className="absolute -bottom-4 text-[7px] font-bold text-blue-400 whitespace-nowrap bg-slate-950/80 px-1 border border-slate-900 rounded font-mono">
                    Kenneth
                  </span>
                </div>
              )}

              {/* Monsters */}
              {activeMonster && (
                <div className="relative flex flex-col items-center justify-center">
                  <div
                    className={`w-7 h-7 rounded-xl flex items-center justify-center border-2 shadow transition-all ${
                      isTargeted
                        ? "bg-red-500/30 border-red-500 scale-105 ring-2 ring-red-500/30"
                        : "bg-red-950/70 border-red-850 hover:border-red-650"
                    }`}
                  >
                    {activeMonster.type === "sewer_rat" ? (
                      <span className="text-xs select-none">🐀</span>
                    ) : (
                      <span className="text-xs select-none animate-pulse">🦠</span>
                    )}
                  </div>

                  {isTargeted && (
                    <div className="absolute inset-0 border border-dashed border-red-500 rounded animate-ping pointer-events-none" />
                  )}

                  {/* Monster HP Bar */}
                  <div className="absolute -top-3.5 w-8 bg-slate-950 h-1 border border-slate-900 rounded-full overflow-hidden">
                    <div
                      className="bg-red-500 h-full transition-all duration-300"
                      style={{ width: `${(activeMonster.hp / activeMonster.maxHp) * 100}%` }}
                    />
                  </div>

                  <span className="absolute -bottom-4 text-[7px] font-bold text-red-400 whitespace-nowrap bg-slate-950/80 px-1 border border-slate-900 rounded font-mono">
                    {activeMonster.name}
                  </span>
                </div>
              )}

              {/* Player Character Sprite */}
              {isPlayerHere && (
                <div className="relative flex flex-col items-center justify-center z-10">
                  <motion.div
                    layoutId="player_character_sprite"
                    className="w-7 h-7 rounded-full bg-gradient-to-tr from-amber-500 to-amber-300 border-2 border-slate-950 flex items-center justify-center shadow-md ring-2 ring-amber-400/20"
                  >
                    <User className="w-4 h-4 text-slate-950" />
                  </motion.div>
                  <div className="absolute -inset-2 bg-amber-500/10 rounded-full filter blur-sm -z-10 animate-pulse pointer-events-none" />
                  <span className="absolute -bottom-4 text-[7px] font-bold text-amber-400 whitespace-nowrap bg-slate-950/90 px-1.5 border border-amber-500/30 rounded font-mono">
                    Novice
                  </span>
                </div>
              )}

              {/* Floating text messages */}
              <AnimatePresence>
                {floatingTexts
                  .filter(ft => ft.x === x && ft.y === y)
                  .map(ft => (
                    <motion.div
                      key={ft.id}
                      initial={{ opacity: 0, y: 15, scale: 0.8 }}
                      animate={{ opacity: 1, y: -20, scale: 1.1 }}
                      exit={{ opacity: 0, y: -35 }}
                      className={`absolute font-mono text-xs tracking-wider font-extrabold select-none pointer-events-none drop-shadow-[0_2px_4px_rgba(0,0,0,0.95)] z-50 ${ft.color}`}
                    >
                      {ft.text}
                    </motion.div>
                  ))}
              </AnimatePresence>
            </div>
          );
        })}
      </div>
    );
  };

  // Status panel calculated parameters
  const nextLvlXp = playerLevel === 1 ? 110 : 400;
  const xpPercent = Math.min(100, Math.floor((playerXp / nextLvlXp) * 100));

  return (
    <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 text-slate-100">
      
      {/* CENTER PANEL: Main Game Viewport, NPC Dialogue, and logs */}
      <div className="xl:col-span-8 flex flex-col gap-4">
        
        {/* Viewport stats banner */}
        <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-2xl flex items-center justify-between backdrop-blur shadow-md">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 animate-spin-slow">
              <Compass className="w-5 h-5" />
            </div>
            <div>
              <span className="text-[10px] text-slate-500 uppercase font-mono tracking-widest block">TELA DO JOGO (MAIN VIEWPORT)</span>
              <span className="text-sm font-bold tracking-tight text-amber-300">
                {isInSewers ? "Ironhold Sewers — Esgotos de Ironhold" : "Ironhold Bastion — Zona Segura de Ironhold"}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[10px] font-mono text-slate-400 bg-slate-950 border border-slate-800 rounded px-2.5 py-1">
              Coordenadas: ({playerPos.x}, {playerPos.y})
            </span>
            <span className="text-[10px] font-mono bg-amber-500/10 text-amber-400 px-2.5 py-1 border border-amber-500/20 rounded">
              LIVRE DE ATALHOS MODULADOS
            </span>
          </div>
        </div>

        {/* NPC DIALOGUE DRAWER */}
        {npcDialogue && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-blue-950/50 border-2 border-blue-500/30 rounded-2xl p-4 flex gap-4 items-center justify-between backdrop-blur-sm shadow"
          >
            <div className="flex items-center gap-3 flex-1">
              <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
                <User className="w-5 h-5" />
              </div>
              <p className="text-xs text-blue-200 leading-relaxed font-semibold">{npcDialogue}</p>
            </div>
            
            {ratKilledCount >= 10 && playerLevel === 1 ? (
              <button
                onClick={handleDeliverQuest}
                className="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-slate-950 text-xs font-bold uppercase tracking-wider rounded-lg shadow-lg flex items-center gap-1 shrink-0 animate-bounce"
              >
                <Gift className="w-4 h-4" /> Entregar Quest
              </button>
            ) : (
              <button
                onClick={() => setNpcDialogue(null)}
                className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-xs text-slate-300 font-semibold rounded-lg"
              >
                Fechar
              </button>
            )}
          </motion.div>
        )}

        {/* MAIN GAME 2D GRID SCREEN */}
        {renderMapGrid()}

        {/* COMPACT FOOTER LOG SUMMARY */}
        <div className="flex justify-between items-center bg-slate-950/60 p-3 border border-slate-850 rounded-xl gap-4">
          <div className="text-[10.5px] text-slate-500 font-mono flex items-center gap-2">
            <Info className="w-4 h-4 text-amber-500 shrink-0" />
            <span>Mova-se com <strong>WASD / Setas</strong>. Clique esquerdo inspeciona. Clique direito interage.</span>
          </div>
          <button
            onClick={onExitWorld}
            className="px-4 py-2 bg-slate-900 hover:bg-rose-950 hover:text-rose-400 text-slate-400 border border-slate-800 text-xs font-bold uppercase tracking-wider rounded-lg transition-all flex items-center gap-1 shrink-0"
          >
            <LogOut className="w-4 h-4" /> Desconectar
          </button>
        </div>

      </div>

      {/* RIGHT COLUMN: CLASSIC TIBIA HUD SIDEBAR PANEL */}
      <div className="xl:col-span-4 flex flex-col gap-4 max-h-[850px] overflow-y-auto pr-1">
        
        {/* 1. MINI MAP */}
        <TibiaMinimap
          playerPos={playerPos}
          exploredTiles={exploredTiles}
          mapTiles={mapTiles}
          MAP_WIDTH={MAP_WIDTH}
          MAP_HEIGHT={MAP_HEIGHT}
          isInSewers={isInSewers}
        />

        {/* 2. PLAYER STATUS PANEL */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4 shadow-md">
          <div className="flex justify-between items-center mb-3">
            <span className="text-[10px] text-slate-400 font-mono font-bold uppercase tracking-widest flex items-center gap-1">
              <User className="w-3.5 h-3.5 text-amber-400" /> STATUS DO JOGADOR
            </span>
            <span className="text-xs font-mono font-extrabold text-amber-400">Level {playerLevel}</span>
          </div>

          <div className="space-y-3">
            {/* Health */}
            <div>
              <div className="flex justify-between text-[10px] font-mono font-bold text-slate-400 mb-1">
                <span>Hp (Vida)</span>
                <span className="text-red-400">{playerHp} / 120</span>
              </div>
              <div className="bg-slate-950 border border-slate-850 h-2 rounded-full overflow-hidden">
                <div 
                  className="bg-emerald-500 h-full transition-all duration-300"
                  style={{ width: `${(playerHp / 120) * 100}%` }}
                />
              </div>
            </div>

            {/* Mana */}
            <div>
              <div className="flex justify-between text-[10px] font-mono font-bold text-slate-400 mb-1">
                <span>Mp (Mana)</span>
                <span className="text-sky-400">{playerMana} / 40</span>
              </div>
              <div className="bg-slate-950 border border-slate-850 h-2 rounded-full overflow-hidden">
                <div 
                  className="bg-sky-500 h-full transition-all duration-300"
                  style={{ width: `${(playerMana / 40) * 100}%` }}
                />
              </div>
            </div>

            {/* Experience percent details */}
            <div>
              <div className="flex justify-between text-[10px] font-mono font-bold text-slate-500 mb-1">
                <span>XP: {playerXp} / {nextLvlXp}</span>
                <span>{xpPercent}%</span>
              </div>
              <div className="bg-slate-950 border border-slate-850 h-1.5 rounded-full overflow-hidden">
                <div 
                  className="bg-violet-500 h-full transition-all duration-300"
                  style={{ width: `${xpPercent}%` }}
                />
              </div>
            </div>

            {/* Gold */}
            <div className="bg-slate-950 p-2.5 rounded border border-slate-900 flex justify-between items-center text-xs font-mono">
              <span className="text-slate-400 flex items-center gap-1">
                <Coins className="w-3.5 h-3.5 text-amber-500" /> Gold gp:
              </span>
              <span className="text-amber-300 font-extrabold">{playerGold} gp</span>
            </div>
          </div>
        </div>

        {/* 3. RETRO EQUIPMENT */}
        <TibiaEquipment
          equipment={playerEquipment}
          onUnequip={handleUnequipItem}
          race={activeChar.race}
        />

        {/* 4. INVENTORY CONTAINERS */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4 shadow-md flex flex-col gap-3">
          
          {/* Main Backpack UI */}
          <div>
            <div className="flex justify-between items-center mb-2 font-mono text-[10px] text-slate-400 font-bold uppercase tracking-widest">
              <span className="flex items-center gap-1">
                <Package className="w-3.5 h-3.5 text-teal-400" /> MOCHILA PRINCIPAL
              </span>
              <span>{playerInventory.length} itens</span>
            </div>

            <div className="bg-slate-950 border border-slate-900 rounded p-2 max-h-[140px] overflow-y-auto space-y-1.5">
              {playerInventory.length === 0 ? (
                <div className="text-center text-slate-600 italic text-[10px] py-4 font-mono">
                  Mochila Vazia
                </div>
              ) : (
                playerInventory.map((item, index) => {
                  const isEquipable = item.id === "sword_basic" || item.id === "shield_wooden";
                  const isHealPot = item.id === "potion_heal";

                  return (
                    <div key={index} className="flex justify-between items-center p-1.5 bg-slate-900/50 border border-slate-850 rounded text-[11px] font-mono">
                      <div>
                        <span className="font-bold text-slate-200">{item.name}</span>
                        <span className="text-[8px] text-slate-500 block">Qtd: {item.qty} • gp: {item.value}</span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        {isEquipable && (
                          <button
                            onClick={() => handleEquipItem(item)}
                            className="px-1.5 py-0.5 bg-amber-500 hover:bg-amber-600 text-slate-950 font-extrabold text-[8px] uppercase rounded transition-colors"
                          >
                            Equipar
                          </button>
                        )}
                        {isHealPot && (
                          <button
                            onClick={handleUseHealPotion}
                            className="px-1.5 py-0.5 bg-emerald-500 hover:bg-emerald-600 text-slate-950 font-extrabold text-[8px] uppercase rounded transition-colors"
                          >
                            Usar
                          </button>
                        )}
                        <button
                          onClick={() => handleDropItem(item.id)}
                          className="px-1.5 py-0.5 bg-slate-800 hover:bg-rose-950 hover:text-rose-400 text-slate-400 text-[8px] uppercase rounded transition-colors"
                          title="Largar item no chão"
                        >
                          Largar
                        </button>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* DUAL CONTAINER: CORPSE CONTAINER Loot Box */}
          {openContainer && (
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="border border-amber-500/30 bg-amber-500/5 rounded-lg p-2.5"
            >
              <div className="flex justify-between items-center mb-1.5 text-[10px] font-mono text-amber-400 font-bold uppercase tracking-wider">
                <span>📥 RECIPINETE: {openContainer.name}</span>
                <button
                  onClick={() => setOpenContainer(null)}
                  className="text-slate-400 hover:text-slate-200 text-[9px]"
                >
                  X
                </button>
              </div>

              <div className="bg-slate-950 border border-slate-900 rounded p-1.5 max-h-[110px] overflow-y-auto space-y-1">
                {openContainer.items.length === 0 ? (
                  <div className="text-center text-slate-600 text-[9px] py-2 italic font-mono">
                    Vazio
                  </div>
                ) : (
                  openContainer.items.map((item, idx) => (
                    <div 
                      key={idx}
                      onClick={() => lootItemFromCorpse(item.id)}
                      className="p-1 bg-slate-900/80 hover:bg-amber-950/20 border border-slate-850 hover:border-amber-500/20 rounded flex justify-between items-center text-[10px] font-mono cursor-pointer transition-all"
                      title="Clique para saquear item para sua mochila"
                    >
                      <span>{item.name}</span>
                      <span className="text-amber-400 font-bold">x{item.qty} (Loot)</span>
                    </div>
                  ))
                )}
              </div>
            </motion.div>
          )}

        </div>

        {/* 5. BATTLE LIST */}
        <TibiaBattleList
          playerPos={playerPos}
          monsters={monsters}
          targetId={targetId}
          onSelectTarget={setTargetId}
          onInteractKenneth={interactWithKenneth}
          playerHp={playerHp}
        />

        {/* 6. TABBED CHAT LOG AND CONSOLE SPELLS */}
        <TibiaChat
          gameLogs={gameLogs}
          onSendMessage={(msg) => {
            addLog(`${activeChar.name} fala: "${msg}"`);
          }}
          onCastExura={castExura}
          onCastStrike={castStrike}
        />

      </div>

    </div>
  );
}
