import React from "react";
import { ShieldAlert, User, Flame } from "lucide-react";

interface TibiaBattleListProps {
  playerPos: { x: number; y: number };
  monsters: any[];
  targetId: string | null;
  onSelectTarget: (id: string | null) => void;
  onInteractKenneth: () => void;
  playerHp: number;
}

export function TibiaBattleList({
  playerPos,
  monsters,
  targetId,
  onSelectTarget,
  onInteractKenneth,
  playerHp
}: TibiaBattleListProps) {
  // Filter active nearby creatures within 7 tiles range
  const nearbyMonsters = monsters.filter(
    (m) => !m.isDead && Math.abs(m.x - playerPos.x) + Math.abs(m.y - playerPos.y) <= 7
  );

  const isKennethNearby = Math.abs(playerPos.x - 10) + Math.abs(playerPos.y - 3) <= 7;

  return (
    <div className="bg-slate-900/80 border border-slate-800 p-3 rounded-xl flex flex-col shadow-md">
      <div className="flex justify-between items-center mb-2 font-mono text-[10px] text-slate-400 font-bold uppercase tracking-widest">
        <span className="flex items-center gap-1.5 text-rose-400">
          <Flame className="w-3.5 h-3.5 text-rose-500 animate-pulse" />
          BATTLE LIST (BATALHA)
        </span>
        <span className="text-[9px] text-slate-500">
          {nearbyMonsters.length + (isKennethNearby ? 1 : 0) + 1} ativos
        </span>
      </div>

      <div className="bg-slate-950 border border-slate-900 rounded p-1.5 max-h-[140px] overflow-y-auto space-y-1">
        {/* Render Player itself */}
        <div className="p-1.5 bg-slate-900/40 border border-slate-800/40 rounded flex items-center justify-between text-[11px] font-mono text-slate-300">
          <div className="flex items-center gap-1.5">
            <span className="text-amber-400">🛡️</span>
            <span className="font-semibold text-slate-200">Você (Novice)</span>
          </div>
          <div className="w-12 bg-slate-950 h-1.5 border border-slate-900 rounded-full overflow-hidden shrink-0">
            <div 
              className="bg-emerald-500 h-full" 
              style={{ width: `${Math.min(100, (playerHp / 120) * 100)}%` }} 
            />
          </div>
        </div>

        {/* Render Captain Kenneth NPC */}
        {isKennethNearby && (
          <div 
            onClick={onInteractKenneth}
            className="p-1.5 bg-blue-950/20 border border-blue-900/30 hover:bg-blue-950/40 rounded flex items-center justify-between text-[11px] font-mono text-blue-300 cursor-pointer"
            title="Clique para conversar com Kenneth"
          >
            <div className="flex items-center gap-1.5">
              <span>👤</span>
              <span className="font-semibold text-blue-200">Capitão Kenneth</span>
            </div>
            <span className="text-[8px] bg-blue-500/20 px-1 border border-blue-400/30 text-blue-400 rounded">
              NPC
            </span>
          </div>
        )}

        {/* Render nearby Monsters */}
        {nearbyMonsters.length === 0 ? (
          <div className="py-4 text-center text-slate-600 text-[10px] italic font-mono">
            Nenhuma criatura hostil próxima
          </div>
        ) : (
          nearbyMonsters.map((m) => {
            const isTargeted = targetId === m.id;
            return (
              <div
                key={m.id}
                onClick={() => onSelectTarget(isTargeted ? null : m.id)}
                className={`p-1.5 rounded flex items-center justify-between text-[11px] font-mono cursor-pointer transition-all ${
                  isTargeted
                    ? "bg-red-500/10 border border-red-500 text-red-300"
                    : "bg-slate-900/60 border border-slate-850 hover:border-slate-750 text-slate-400"
                }`}
                title={isTargeted ? "Atacando! Clique para cancelar mira." : "Clique para iniciar ataque canônico."}
              >
                <div className="flex items-center gap-1.5">
                  <span className={isTargeted ? "text-red-500 animate-pulse font-bold" : "text-slate-500"}>
                    {isTargeted ? "▶" : "☠"}
                  </span>
                  <span className={isTargeted ? "font-bold text-red-200" : ""}>{m.name}</span>
                </div>
                
                <div className="flex items-center gap-2 shrink-0">
                  <div className="w-12 bg-slate-950 h-1.5 border border-slate-900 rounded-full overflow-hidden">
                    <div 
                      className="bg-red-500 h-full transition-all duration-300" 
                      style={{ width: `${(m.hp / m.maxHp) * 100}%` }} 
                    />
                  </div>
                  {isTargeted && (
                    <span className="text-[7.5px] bg-red-500 text-slate-950 font-bold px-1 rounded">
                      ALVO
                    </span>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
