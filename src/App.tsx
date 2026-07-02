import React, { useState } from "react";
import { Sparkles } from "lucide-react";
import { GameBootstrapRoot } from "./core/components/GameBootstrapRoot";

export default function App() {
  // Sincronização de estados do jogador
  const [playerInventory, setPlayerInventory] = useState<{ id: string; name: string; qty: number; value: number; type: string }[]>([]);
  const [playerEquipped, setPlayerEquipped] = useState<any[]>([]);
  const [progLevel, setProgLevel] = useState<number>(1);
  const [progClass, setProgClass] = useState<string>("Knight");
  const [carriedCoins, setCarriedCoins] = useState<number>(100);
  const [createdCharName, setCreatedCharName] = useState<string>("Hero");

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
              Official Production Client
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
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
