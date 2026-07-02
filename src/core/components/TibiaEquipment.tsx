import React from "react";
import { Shield, Sword, Package, Heart, Zap } from "lucide-react";
import { EquippedItem } from "../../types";

interface TibiaEquipmentProps {
  equipment: EquippedItem[];
  onUnequip: (slot: string) => void;
  race?: string;
}

export function TibiaEquipment({ equipment, onUnequip, race }: TibiaEquipmentProps) {
  // Map out our slots to the classic 3x3 Tibia layout:
  // Row 1: Necklace, Head, Backpack
  // Row 2: Weapon, Armor, Offhand
  // Row 3: Ring, Legs, Boots

  const getSlotItem = (slotName: string) => {
    return equipment.find(
      (eq) => eq.slot.toLowerCase() === slotName.toLowerCase()
    );
  };

  const renderSlot = (slotName: string, label: string, defaultIcon: React.ReactNode) => {
    const item = getSlotItem(slotName);
    return (
      <div
        onClick={() => item && onUnequip(slotName)}
        className={`w-12 h-12 rounded border flex flex-col items-center justify-center cursor-pointer transition-all ${
          item
            ? "bg-amber-950/40 border-amber-500/60 hover:bg-amber-900/30 text-amber-300"
            : "bg-slate-950/80 border-slate-900 text-slate-700 hover:border-slate-800"
        }`}
        title={item ? `Equipado: ${item.name} (${slotName}). Clique para desequipar.` : `Slot vazio: ${label}`}
      >
        {item ? (
          <div className="flex flex-col items-center text-center">
            <span className="text-[14px] leading-tight select-none">
              {slotName === "Weapon" ? "🗡️" : slotName === "Offhand" ? "🛡️" : slotName === "Backpack" ? "🎒" : "📦"}
            </span>
            <span className="text-[7px] font-mono text-amber-400 font-bold tracking-tight block truncate max-w-[44px]">
              {item.name.split(" ")[0]}
            </span>
          </div>
        ) : (
          <div className="flex flex-col items-center text-center opacity-40">
            {defaultIcon}
            <span className="text-[6.5px] font-mono font-bold mt-0.5">{label}</span>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="bg-slate-900/80 border border-slate-800 p-3 rounded-xl flex flex-col items-center shadow-md">
      <div className="flex items-center gap-1.5 mb-3 font-mono text-[10px] text-slate-400 font-bold uppercase tracking-widest w-full">
        <Shield className="w-3.5 h-3.5 text-pink-400" />
        <span>EQUIPAMENTOS</span>
      </div>

      <div className="grid grid-cols-3 gap-2 bg-slate-950 p-2 border border-slate-900 rounded">
        {/* Row 1 */}
        {renderSlot("Necklace", "NECK", <span className="text-xs font-serif">📿</span>)}
        {renderSlot("Head", "HEAD", <span className="text-xs font-serif">🪖</span>)}
        {renderSlot("Backpack", "BAG", <span className="text-xs font-serif">🎒</span>)}

        {/* Row 2 */}
        {renderSlot("Weapon", "WEAP", <Sword className="w-3.5 h-3.5" />)}
        {renderSlot("Armor", "ARM", <span className="text-xs font-serif">👕</span>)}
        {renderSlot("Offhand", "SHLD", <Shield className="w-3.5 h-3.5" />)}

        {/* Row 3 */}
        {renderSlot("Ring", "RING", <span className="text-xs font-serif">💍</span>)}
        {renderSlot("Legs", "LEGS", <span className="text-xs font-serif">👖</span>)}
        {renderSlot("Boots", "BOOT", <span className="text-xs font-serif">🥾</span>)}
      </div>

      {race && (
        <div className="text-[8px] font-mono text-slate-500 uppercase tracking-wide mt-2 text-center">
          Veste: <span className="text-slate-300 font-semibold">{race} Outfit</span>
        </div>
      )}
    </div>
  );
}
