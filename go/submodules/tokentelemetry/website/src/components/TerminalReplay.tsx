"use client";
import { useEffect, useState } from "react";
import { Activity } from "lucide-react";
import { TERMINAL_SCRIPT, SCRIPT_DURATION_MS, type ScriptLine } from "@/data/terminal-script";

const COLOR: Record<ScriptLine["kind"], string> = {
  header:    "text-[var(--tt-fg-faint)]",
  user:      "text-[#60a5fa]",
  reasoning: "text-[var(--tt-warn-fg)] italic",
  tool:      "text-[var(--tt-info-fg)]",
  result:    "text-[var(--tt-fg-dim)]",
  cost:      "text-[var(--tt-success-fg)]",
};

export default function TerminalReplay() {
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    const start = Date.now();
    const id = setInterval(() => {
      const t = (Date.now() - start) % (SCRIPT_DURATION_MS + 1500);
      setElapsed(t);
    }, 60);
    return () => clearInterval(id);
  }, []);

  const visible = TERMINAL_SCRIPT.filter((l) => l.delay <= elapsed);

  return (
    <div className="rounded-[var(--tt-radius-xl)] border border-[var(--tt-border-strong)] bg-[var(--tt-sunken)] overflow-hidden shadow-[0_30px_120px_-30px_rgba(96,165,250,0.35)]">
      <div className="flex items-center gap-1.5 px-4 h-9 bg-[var(--tt-raised)] border-b border-[var(--tt-border)]">
        <span className="w-2.5 h-2.5 rounded-full bg-rose-400/50" />
        <span className="w-2.5 h-2.5 rounded-full bg-amber-400/50" />
        <span className="w-2.5 h-2.5 rounded-full bg-emerald-400/50" />
        <span className="ml-3 inline-flex items-center gap-1.5 text-[11px] font-mono text-[var(--tt-fg-dim)] truncate">
          <Activity size={11} className="text-[var(--tt-brand)]" /> tokentelemetry · session-trace
        </span>
      </div>
      <div className="p-5 sm:p-6 font-mono text-[12px] sm:text-[13px] leading-relaxed min-h-[320px] sm:min-h-[360px] space-y-1.5">
        {visible.map((line, i) => (
          <div key={i} className={`${COLOR[line.kind]} transition-opacity`}>
            {line.text}
          </div>
        ))}
        <span className="inline-block w-2 h-4 bg-[var(--tt-success-fg)] align-middle animate-pulse" />
      </div>
    </div>
  );
}
