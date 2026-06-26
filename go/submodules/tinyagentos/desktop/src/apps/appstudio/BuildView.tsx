import { useCallback, useEffect, useRef, useState } from "react";
import { CheckCircle2, Folder, Bell, Loader2, Sparkles } from "lucide-react";
import { PROMPT_SEEDED_EVENT, takePendingPrompt } from "./build-state";
import { streamTaosAgentChat } from "./stream-chat";

/* ------------------------------------------------------------------ */
/*  BuildView -- sandbox preview + build log + prompt bar              */
/* ------------------------------------------------------------------ */

const CHORES = [
  { who: "M", task: "Take out the bins", sub: "Mara · Mon", done: true },
  { who: "B", task: "Walk the dog", sub: "Ben · daily", pts: 5, done: false },
  { who: "I", task: "Load dishwasher", sub: "Ivo · today", pts: 3, done: false },
  { who: "M", task: "Water the plants", sub: "Mara · Wed", pts: 2, done: false },
];

const CAPS = [
  { icon: Folder, label: "Workspace files" },
  { icon: Bell, label: "Send notifications" },
];

const DEFAULT_PROMPT = "give each person a weekly points total and a little leaderboard";

export function BuildView() {
  const [prompt, setPrompt] = useState(DEFAULT_PROMPT);
  const [output, setOutput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [streaming, setStreaming] = useState(false);
  const [model, setModel] = useState<string | null>(null);
  const logEndRef = useRef<HTMLDivElement>(null);
  const outputChunksRef = useRef<string[]>([]);
  const abortRef = useRef<AbortController | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      abortRef.current?.abort();
    };
  }, []);

  useEffect(() => {
    const seeded = takePendingPrompt();
    if (seeded) setPrompt(seeded);

    const onSeeded = (e: Event) => {
      const detail = (e as CustomEvent<string>).detail;
      if (detail) setPrompt(detail);
    };
    window.addEventListener(PROMPT_SEEDED_EVENT, onSeeded);
    return () => window.removeEventListener(PROMPT_SEEDED_EVENT, onSeeded);
  }, []);

  useEffect(() => {
    fetch("/api/taos-agent/settings")
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data?.model !== undefined) setModel(data.model);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!streaming || error) {
      logEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [streaming, error]);

  const handleBuild = useCallback(async () => {
    if (streaming || !model) return;
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    const text = prompt.trim();
    if (!text) return;

    outputChunksRef.current = [];
    setOutput("");
    setError(null);
    setStreaming(true);

    const buildPrompt =
      `Help me build a taOS SDK app in App Studio. User request: ${text}`;

    try {
      await streamTaosAgentChat(
        [{ role: "user", content: buildPrompt }],
        (delta) => {
          if (!mountedRef.current) return;
          outputChunksRef.current.push(delta);
          setOutput(outputChunksRef.current.join(""));
        },
        (message) => {
          if (!mountedRef.current) return;
          setError(message);
        },
        { signal: controller.signal },
      );
    } catch (e) {
      if (!(e instanceof DOMException && e.name === "AbortError") && mountedRef.current) {
        setError(String(e));
      }
    } finally {
      if (abortRef.current === controller) abortRef.current = null;
      if (mountedRef.current) setStreaming(false);
    }
  }, [prompt, streaming, model]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        e.preventDefault();
        void handleBuild();
      }
    },
    [handleBuild],
  );

  const noModel = !model;
  const showIdleLog = !streaming && !output && !error;

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* view header */}
      <div className="flex h-[54px] flex-none items-center gap-3 border-b border-shell-border px-[22px]">
        <h2 className="text-[17px] font-bold tracking-[-0.02em]">Build</h2>
        <span className="text-[12px] text-shell-text-tertiary">
          Chore Quest &middot; taOS app &middot; sandboxed
        </span>
      </div>

      {/* build area */}
      <div className="flex min-h-0 flex-1">
        {/* sandbox panel */}
        <div
          className="flex min-w-0 flex-1 flex-col p-[22px]"
          style={{
            background:
              "repeating-conic-gradient(rgba(255,255,255,.016) 0% 25%, transparent 0% 50%) 0 0 / 22px 22px, var(--tw-color-shell-bg, transparent)",
          }}
        >
          {/* chip row */}
          <div className="mb-[14px] flex items-center gap-[9px]">
            <Chip icon={<CheckCircle2 size={13} className="text-[#5fbf78]" />} label="Live preview" />
            <Chip label="Sandboxed" />
            <span className="ml-auto text-[11px] text-shell-text-tertiary">
              taOS SDK 1.0 &middot; no network
            </span>
          </div>

          {/* app window in window */}
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[14px] border border-shell-border-strong shadow-[0_18px_40px_-16px_rgba(0,0,0,0.6)]">
            {/* mini titlebar */}
            <div
              className="flex h-[38px] flex-none items-center gap-[9px] px-[13px]"
              style={{ background: "linear-gradient(135deg,#6f7687,#525868)", color: "#fff" }}
            >
              <div className="flex gap-1.5">
                {[0, 1, 2].map((i) => (
                  <span
                    key={i}
                    className="inline-block h-[9px] w-[9px] rounded-full"
                    style={{ background: "rgba(255,255,255,0.5)" }}
                  />
                ))}
              </div>
              <span className="text-[12px] font-bold">Chore Quest</span>
            </div>

            {/* app body */}
            <div
              className="flex-1 overflow-auto p-[18px_20px]"
              style={{ background: "#202024", color: "#e8eaed" }}
            >
              <div className="mb-0.5 text-[18px] font-extrabold tracking-[-0.02em] text-white">
                This week
              </div>
              <div className="mb-4 text-[12px]" style={{ color: "#aab" }}>
                3 of 6 done &middot; Team Henderson
              </div>

              {CHORES.map((c) => (
                <div
                  key={c.task}
                  className="mb-[9px] flex items-center gap-3 rounded-[11px] p-[11px_12px]"
                  style={{ background: "rgba(255,255,255,0.05)" }}
                >
                  <div
                    className="flex h-[30px] w-[30px] flex-none items-center justify-center rounded-full text-[12px] font-bold text-white"
                    style={{ background: "linear-gradient(135deg,#7c8ba1,#aab4c9)" }}
                  >
                    {c.who}
                  </div>
                  <div>
                    <div className="text-[13.5px] font-semibold" style={{ color: "#f4f5f7" }}>
                      {c.task}
                    </div>
                    <div className="text-[11px]" style={{ color: "#99a" }}>
                      {c.sub}
                    </div>
                  </div>
                  {c.done ? (
                    <CheckCircle2 size={20} className="ml-auto" style={{ color: "#5fbf78" }} />
                  ) : (
                    <span
                      className="ml-auto rounded-full px-[10px] py-1 text-[11px] font-bold"
                      style={{ color: "#cdd3df", background: "rgba(255,255,255,0.08)" }}
                    >
                      +{c.pts}
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* build log panel */}
        <div className="flex w-[300px] flex-none flex-col border-l border-shell-border">
          <div className="flex items-center gap-2 px-[18px] pb-2 pt-4 text-[13px] font-bold">
            {streaming ? (
              <Loader2 size={16} className="animate-spin text-accent" />
            ) : (
              <CheckCircle2 size={16} className="text-accent" />
            )}
            Build log
          </div>

          <div className="min-h-0 flex-1 overflow-auto px-[14px] pb-2">
            {showIdleLog && (
              <p className="px-1 py-2 text-[12px] text-shell-text-tertiary">
                Describe your app below and press Build to stream the agent&apos;s plan and output here.
              </p>
            )}
            {error && (
              <div
                className="mb-2 rounded-[11px] border border-red-500/30 bg-red-500/10 px-3 py-2 text-[12px] text-red-300"
                role="alert"
              >
                {error}
              </div>
            )}
            {output && (
              <pre className="whitespace-pre-wrap break-words rounded-[11px] bg-shell-surface p-[10px] text-[12px] leading-relaxed text-shell-text-secondary">
                {output}
              </pre>
            )}
            {streaming && !output && !error && (
              <div className="flex items-center gap-2 px-1 py-2 text-[12px] text-shell-text-tertiary">
                <Loader2 size={14} className="animate-spin" />
                Building...
              </div>
            )}
            <div ref={logEndRef} />
          </div>

          {/* capabilities */}
          <div className="mx-[14px] border-t border-shell-border pt-3">
            <div className="mb-[9px] text-[10.5px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
              Capabilities requested
            </div>
            {CAPS.map(({ icon: Icon, label }) => (
              <div key={label} className="flex items-center gap-[9px] px-0.5 py-1.5 text-[12px] text-shell-text-secondary">
                <Icon size={14} className="text-accent" />
                {label}
                <span className="ml-auto text-[10.5px] font-bold" style={{ color: "#5fbf78" }}>
                  Granted
                </span>
              </div>
            ))}
          </div>

          {/* model pill */}
          <div className="mt-auto border-t border-shell-border p-[13px_14px]">
            <div className="flex items-center gap-[10px] rounded-[12px] border border-shell-border bg-shell-surface p-[9px_12px]">
              <div
                className="h-6 w-6 rounded-[7px]"
                style={{ background: "linear-gradient(135deg,#7c8ba1,#aab4c9)" }}
              />
              <div>
                <div className="text-[12px] font-semibold">
                  {model ?? "No model selected"}
                </div>
                <div className="text-[10px] text-shell-text-tertiary">
                  {noModel ? "choose a model in taOS Assistant settings" : "taOS agent"}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* prompt bar */}
      <div className="flex flex-none items-center gap-3 border-t border-shell-border bg-shell-bg-deep px-[22px] py-4">
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={2}
          disabled={streaming}
          placeholder="Describe what your app should do..."
          className="flex min-h-[50px] flex-1 resize-none items-center rounded-[15px] border border-shell-border bg-shell-surface px-4 py-3 text-[13.5px] text-shell-text-secondary placeholder:text-shell-text-tertiary focus-visible:border-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/20 disabled:opacity-60"
        />
        <button
          type="button"
          onClick={() => void handleBuild()}
          disabled={streaming || !prompt.trim() || noModel}
          className="flex h-[50px] flex-none items-center gap-[9px] rounded-[15px] border-none px-6 text-[14px] font-bold text-white disabled:cursor-not-allowed disabled:opacity-50"
          style={{ background: "linear-gradient(135deg,var(--color-accent),var(--color-accent))" }}
        >
          {streaming ? <Loader2 size={18} className="animate-spin" /> : <Sparkles size={18} />}
          {streaming ? "Building..." : "Build"}
        </button>
      </div>
    </div>
  );
}

function Chip({ icon, label }: { icon?: React.ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-[7px] rounded-[8px] border border-shell-border bg-shell-surface px-[11px] py-[5px] text-[11px] font-semibold text-shell-text-secondary">
      {icon}
      {label}
    </div>
  );
}