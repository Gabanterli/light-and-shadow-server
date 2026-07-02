import React, { useState, useEffect, useRef } from "react";
import { Terminal, Send, HelpCircle } from "lucide-react";

interface TibiaChatProps {
  gameLogs: string[];
  onSendMessage: (msg: string) => void;
  onCastExura: () => void;
  onCastStrike: () => void;
}

export function TibiaChat({
  gameLogs,
  onSendMessage,
  onCastExura,
  onCastStrike
}: TibiaChatProps) {
  const [activeTab, setActiveTab] = useState<"local" | "server">("local");
  const [chatInput, setChatInput] = useState("");
  const chatScrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll chat to bottom when new logs arrive or tab changes
  useEffect(() => {
    if (chatScrollRef.current) {
      chatScrollRef.current.scrollTop = chatScrollRef.current.scrollHeight;
    }
  }, [gameLogs, activeTab]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!chatInput.trim()) return;

    const lowerInput = chatInput.trim().toLowerCase();

    // Context Spell Casting from Chat Input
    if (lowerInput === "exura" || lowerInput === "1") {
      onCastExura();
    } else if (lowerInput === "exori flam" || lowerInput === "flam" || lowerInput === "2") {
      onCastStrike();
    } else {
      onSendMessage(chatInput);
    }

    setChatInput("");
  };

  // Filter logs for each channel
  const filteredLogs = gameLogs.filter((log) => {
    if (activeTab === "server") {
      // Server log contains system logs, quests, and rewards
      return (
        log.includes("[Mundo]") ||
        log.includes("[Sistema]") ||
        log.includes("[Quest") ||
        log.includes("[Sessão]") ||
        log.includes("🏆") ||
        log.includes("🎉") ||
        log.includes("🆙") ||
        log.includes("🎖️")
      );
    } else {
      // Local chat contains dialogs, spell casts, and user inputs
      return (
        !log.includes("[Sessão]") &&
        !log.includes("[Quest Concluída]") &&
        !log.includes("Sincronizando")
      );
    }
  });

  return (
    <div className="bg-slate-900/80 border border-slate-800 p-3 rounded-xl flex flex-col h-60 shadow-md">
      {/* Tabs */}
      <div className="flex justify-between items-center mb-2 border-b border-slate-850 pb-1.5 shrink-0">
        <div className="flex gap-2 font-mono text-[10px] font-bold">
          <button
            onClick={() => setActiveTab("local")}
            className={`px-2.5 py-1 rounded transition-all ${
              activeTab === "local"
                ? "bg-amber-500/10 border border-amber-500/20 text-amber-400"
                : "text-slate-500 hover:text-slate-300"
            }`}
          >
            LOCAL CHAT
          </button>
          <button
            onClick={() => setActiveTab("server")}
            className={`px-2.5 py-1 rounded transition-all ${
              activeTab === "server"
                ? "bg-violet-500/10 border border-violet-500/20 text-violet-400"
                : "text-slate-500 hover:text-slate-300"
            }`}
          >
            SERVER LOG
          </button>
        </div>

        <span className="text-[9px] font-mono text-slate-500 flex items-center gap-1">
          <Terminal className="w-3.5 h-3.5" /> CONSOLE CANÔNICO
        </span>
      </div>

      {/* Messages viewport */}
      <div
        ref={chatScrollRef}
        className="flex-1 bg-slate-950 border border-slate-900 rounded p-2.5 overflow-y-auto space-y-1 font-mono text-[10.5px] leading-relaxed text-slate-400 text-left mb-2"
      >
        {filteredLogs.length === 0 ? (
          <div className="text-slate-700 italic text-[10px] py-4 text-center">
            Nenhuma mensagem neste canal
          </div>
        ) : (
          filteredLogs.map((log, index) => {
            let logColor = "text-slate-300";
            if (log.includes("🏆") || log.includes("🎉") || log.includes("🆙") || log.includes("🎖️")) {
              logColor = "text-emerald-400 font-bold";
            } else if (log.includes("💥") || log.includes("💀")) {
              logColor = "text-rose-400";
            } else if (log.includes("🛡️") || log.includes("[Active Defense]")) {
              logColor = "text-amber-400";
            } else if (log.includes("Spell:") || log.includes("✨")) {
              logColor = "text-sky-400 font-semibold";
            } else if (log.includes("[Target]") || log.includes("[Mira]")) {
              logColor = "text-red-400";
            } else if (log.includes("fala:") || log.includes("Kenneth:")) {
              logColor = "text-blue-300 font-medium";
            }

            return (
              <div key={index} className={logColor}>
                {log}
              </div>
            );
          })
        )}
      </div>

      {/* Input box */}
      <form onSubmit={handleSubmit} className="flex gap-1.5 shrink-0">
        <input
          type="text"
          value={chatInput}
          onChange={(e) => setChatInput(e.target.value)}
          className="flex-1 px-3 py-1.5 bg-slate-950 border border-slate-900 rounded text-xs text-slate-200 font-mono focus:outline-none focus:border-amber-500/60 transition-all placeholder:text-slate-700"
          placeholder="Digite mensagens locais ou feitiços (ex: 'exura', 'exori flam')..."
        />
        <button
          type="submit"
          className="px-3 py-1.5 bg-slate-950 hover:bg-slate-850 border border-slate-900 hover:border-slate-800 text-slate-400 hover:text-slate-200 rounded flex items-center justify-center transition-all"
        >
          <Send className="w-3.5 h-3.5" />
        </button>
      </form>
    </div>
  );
}
