import React from "react";
import { Compass } from "lucide-react";

interface TibiaMinimapProps {
  playerPos: { x: number; y: number };
  exploredTiles: Record<string, boolean>;
  mapTiles: any[][];
  MAP_WIDTH: number;
  MAP_HEIGHT: number;
  isInSewers: boolean;
}

export function TibiaMinimap({
  playerPos,
  exploredTiles,
  mapTiles,
  MAP_WIDTH,
  MAP_HEIGHT,
  isInSewers
}: TibiaMinimapProps) {
  return (
    <div className="bg-slate-950 border border-slate-800 p-2.5 rounded-xl shadow-inner flex flex-col items-center">
      <div className="flex items-center gap-1.5 mb-2 font-mono text-[10px] text-slate-400 font-bold uppercase tracking-widest w-full">
        <Compass className="w-3.5 h-3.5 text-amber-500" />
        <span>MINI MAPA</span>
      </div>

      <div 
        className="grid border border-slate-900 bg-black rounded shadow"
        style={{
          gridTemplateColumns: `repeat(${MAP_WIDTH}, minmax(0, 1fr))`,
          width: "140px",
          height: "140px"
        }}
      >
        {Array.from({ length: MAP_HEIGHT }).map((_, y) => {
          return Array.from({ length: MAP_WIDTH }).map((_, x) => {
            const coordKey = `${x},${y}`;
            const isExplored = exploredTiles[coordKey];
            const isPlayer = playerPos.x === x && playerPos.y === y;

            let color = "bg-black";
            if (isExplored) {
              const tile = mapTiles[y]?.[x];
              if (tile) {
                if (tile.type === "WALL") {
                  color = "bg-slate-700";
                } else if (tile.type === "WATER") {
                  color = "bg-blue-600";
                } else if (tile.type === "LADDER_DOWN" || tile.type === "LADDER_UP") {
                  color = "bg-amber-400";
                } else if (tile.type === "LOCKED_GATE") {
                  color = "bg-rose-600";
                } else if (y >= 12) {
                  color = "bg-stone-800"; // sewer ground
                } else {
                  color = "bg-emerald-800"; // town ground
                }
              }
            }

            return (
              <div
                key={`${x}-${y}`}
                className={`relative aspect-square ${isPlayer ? "bg-white" : color}`}
                style={{ fontSize: "1px" }}
              >
                {isPlayer && (
                  <div className="absolute inset-0 bg-yellow-400 rounded-full animate-ping" />
                )}
              </div>
            );
          });
        })}
      </div>

      <div className="flex justify-between w-full mt-2 font-mono text-[9px] text-slate-500">
        <span>X: {playerPos.x} Y: {playerPos.y}</span>
        <span>{isInSewers ? "SEWERS" : "TOWN"}</span>
      </div>
    </div>
  );
}
