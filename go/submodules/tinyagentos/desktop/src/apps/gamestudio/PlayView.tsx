import { useCallback, useEffect, useRef, useState } from "react";
import {
  Play,
  Pause,
  RotateCcw,
  Maximize2,
  Monitor,
  Smartphone,
  Glasses,
  LogOut,
  ListTree,
  Loader2,
  Check,
  Clock,
} from "lucide-react";
import { GAME_CREATED_EVENT, takeCreatedGame } from "./game-state";
import { useGameScene } from "./useGameScene";
import { BUILD_LOG } from "./templates";
import type { BuildStep, DevicePreview, Template } from "./types";

/* ------------------------------------------------------------------ */
/*  PlayView — the real three.js preview stage                         */
/*                                                                     */
/*  HARD REQUIREMENT (Jay): whenever the stage is fullscreen, an        */
/*  always-visible "Exit to taOS" pill is layered on top, and Escape    */
/*  exits fullscreen on desktop. The user is never trapped.             */
/*                                                                      */
/*  Real: WebGLRenderer scene, play/pause, WASD + drag controls, a      */
/*  live FPS overlay, the Fullscreen API, device resize (Desktop /      */
/*  Mobile). Stubbed-but-honest: XR is a labelled affordance (real      */
/*  WebXR is a later phase); the build log is an illustrative preview   */
/*  of the future agent trace.                                          */
/* ------------------------------------------------------------------ */

const DEVICE_SIZES: Record<DevicePreview, { maxW: number | null; aspect: string }> = {
  desktop: { maxW: null, aspect: "16 / 10" },
  mobile: { maxW: 300, aspect: "9 / 16" },
  xr: { maxW: 560, aspect: "16 / 8" },
};

export interface PlayViewProps {
  windowId: string;
  template: Template;
}

export function PlayView({ windowId, template }: PlayViewProps) {
  const scene = useGameScene(template.scene);
  const { hostRef, playing, togglePlay, setPlaying, fps, reset, supported } = scene;

  const stageRef = useRef<HTMLDivElement | null>(null);
  const [device, setDevice] = useState<DevicePreview>("desktop");
  const [isFs, setIsFs] = useState(false);
  const [buildLog, setBuildLog] = useState<BuildStep[]>(BUILD_LOG);
  const [sourcePrompt, setSourcePrompt] = useState<string | null>(null);

  useEffect(() => {
    const applySnapshot = () => {
      const { steps, prompt } = takeCreatedGame(windowId);
      if (steps?.length) setBuildLog(steps);
      if (prompt) setSourcePrompt(prompt);
    };
    const onCreated = (e: Event) => {
      const detail = (e as CustomEvent<{ windowId?: string }>).detail;
      if (detail?.windowId !== windowId) return;
      applySnapshot();
    };
    applySnapshot();
    window.addEventListener(GAME_CREATED_EVENT, onCreated);
    return () => window.removeEventListener(GAME_CREATED_EVENT, onCreated);
  }, [template.id, windowId]);

  // Track native fullscreen state. The Exit pill + Esc both call exitFs.
  useEffect(() => {
    const onChange = () => setIsFs(document.fullscreenElement === stageRef.current);
    document.addEventListener("fullscreenchange", onChange);
    return () => document.removeEventListener("fullscreenchange", onChange);
  }, []);

  const enterFs = useCallback(() => {
    const el = stageRef.current;
    el?.requestFullscreen?.().catch(() => {});
  }, []);

  const exitFs = useCallback(() => {
    if (document.fullscreenElement) document.exitFullscreen?.().catch(() => {});
  }, []);

  const toggleFs = useCallback(() => {
    if (document.fullscreenElement) exitFs();
    else enterFs();
  }, [enterFs, exitFs]);

  // Escape exits fullscreen on desktop. The browser fires its own Esc for
  // native fullscreen too; this guarantees the affordance and keeps the
  // app's own state in sync, so the user is never trapped.
  useEffect(() => {
    if (!isFs) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") exitFs();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [isFs, exitFs]);

  const size = DEVICE_SIZES[device];
  const isXr = device === "xr";

  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
      <div className="flex flex-col gap-3.5 p-[22px]">
        {/* heading */}
        <div className="flex flex-wrap items-baseline gap-2.5">
          <h2 className="text-[17px] font-bold tracking-[-0.02em]">{template.title}</h2>
          <span className="text-[11px] text-shell-text-tertiary">
            {template.genre} · live preview · auto-saved
          </span>
        </div>
        {sourcePrompt && (
          <p className="max-w-3xl text-[12px] leading-relaxed text-shell-text-secondary">
            <span className="font-semibold text-shell-text">From your prompt:</span> {sourcePrompt}
          </p>
        )}

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-[1fr_312px]">
          {/* ---- stage + transport ---- */}
          <div className="flex flex-col gap-3">
            <div
              className="relative mx-auto w-full overflow-hidden rounded-2xl border border-shell-border-strong bg-[#0a0d14] shadow-card"
              style={{
                aspectRatio: isFs ? undefined : size.aspect,
                maxWidth: isFs ? undefined : (size.maxW ?? undefined),
                height: isFs ? "100%" : undefined,
              }}
              ref={stageRef}
            >
              {/* three.js canvas host */}
              <div ref={hostRef} className="absolute inset-0 h-full w-full" />

              {/* honest fallback if WebGL is unavailable */}
              {!supported && (
                <div className="absolute inset-0 flex items-center justify-center px-6 text-center text-[12.5px] text-white/70">
                  This device could not start WebGL, so the 3D preview is unavailable here.
                </div>
              )}

              {/* XR is a labelled affordance in phase 1 (real WebXR later) */}
              {isXr && (
                <div className="pointer-events-none absolute inset-x-0 bottom-0 z-10 bg-gradient-to-t from-black/65 to-transparent px-3 pb-3 pt-8">
                  <span className="rounded-full bg-black/55 px-2.5 py-1 text-[10.5px] font-semibold text-white/85 backdrop-blur-sm">
                    XR preview framing · real WebXR headset support arrives in a later phase
                  </span>
                </div>
              )}

              {/* PERSISTENT Exit to taOS — always visible while fullscreen.
                  Slate glass pill, pinned top-left, safe-area aware. */}
              {isFs && (
                <button
                  type="button"
                  onClick={exitFs}
                  aria-label="Exit to taOS"
                  className="absolute z-50 inline-flex items-center gap-1.5 rounded-full border border-accent/45 bg-[rgba(16,18,24,0.62)] px-3 py-1.5 text-[12px] font-bold text-white shadow-lg backdrop-blur-md transition-all hover:-translate-y-0.5 hover:bg-[rgba(28,32,42,0.82)]"
                  style={{
                    top: "max(12px, env(safe-area-inset-top, 12px))",
                    left: "max(12px, env(safe-area-inset-left, 12px))",
                  }}
                >
                  <LogOut size={15} className="-scale-x-100" />
                  Exit to taOS
                  <span className="ml-0.5 rounded border border-white/25 px-1.5 py-px text-[9.5px] font-bold tracking-wide text-white/60">
                    ESC
                  </span>
                </button>
              )}

              {/* live FPS / debug HUD (real, from the rAF loop) */}
              <div
                className="absolute z-40 flex items-center gap-1.5 rounded-lg border border-white/10 bg-black/45 px-2.5 py-1.5 font-mono text-[10.5px] text-emerald-200/90 backdrop-blur-sm"
                style={{ top: isFs ? "62px" : "10px", left: "10px" }}
              >
                <span>
                  <b className="font-bold text-white">{fps}</b> fps
                </span>
                <span className="text-white/25">|</span>
                <span>
                  <b className="font-bold text-white">{playing ? "running" : "paused"}</b>
                </span>
              </div>

              {/* viewport controls (top-right) */}
              <div className="absolute right-2.5 top-2.5 z-40 flex gap-1.5">
                <button
                  type="button"
                  onClick={reset}
                  aria-label="Reset scene"
                  className="grid h-[30px] w-[30px] place-items-center rounded-lg border border-white/10 bg-black/45 text-white backdrop-blur-sm transition-colors hover:bg-black/65"
                >
                  <RotateCcw size={15} />
                </button>
                <button
                  type="button"
                  onClick={toggleFs}
                  aria-label="Toggle fullscreen"
                  className="grid h-[30px] w-[30px] place-items-center rounded-lg border border-white/10 bg-black/45 text-white backdrop-blur-sm transition-colors hover:bg-black/65"
                >
                  <Maximize2 size={15} />
                </button>
              </div>

              {/* play overlay before first Play */}
              {!playing && (
                <button
                  type="button"
                  onClick={() => setPlaying(true)}
                  aria-label="Play scene"
                  className="absolute inset-0 z-30 grid place-items-center"
                >
                  <span className="grid h-[66px] w-[66px] place-items-center rounded-full border border-white/30 bg-white/15 shadow-xl backdrop-blur-sm transition-transform hover:scale-105">
                    <Play size={26} className="translate-x-0.5 text-white" />
                  </span>
                </button>
              )}

              <div
                className="pointer-events-none absolute bottom-3 left-3 z-20 text-[12px] font-semibold text-white/90"
                style={{ textShadow: "0 1px 6px rgba(0,0,0,0.6)" }}
              >
                {template.title}
                <span className="block text-[10.5px] font-medium text-white/60">
                  drag to orbit · WASD to move
                </span>
              </div>
            </div>

            {/* transport + device preview */}
            <div className="flex flex-wrap items-center gap-2.5">
              <div className="flex gap-1.5">
                <button
                  type="button"
                  onClick={reset}
                  aria-label="Restart"
                  className="grid h-9 w-9 place-items-center rounded-xl border border-shell-border bg-shell-surface-active text-shell-text transition-colors hover:bg-white/10"
                >
                  <RotateCcw size={15} />
                </button>
                <button
                  type="button"
                  onClick={togglePlay}
                  className="flex h-9 items-center gap-1.5 rounded-full border border-shell-border bg-shell-surface-active px-4 text-[12.5px] font-bold text-shell-text transition-colors hover:bg-white/10"
                >
                  {playing ? <Pause size={15} /> : <Play size={15} />}
                  {playing ? "Pause" : "Play"}
                </button>
              </div>

              <div
                className="ml-auto inline-flex gap-0.5 rounded-full border border-shell-border bg-shell-surface p-[3px]"
                role="group"
                aria-label="Device preview"
              >
                {(
                  [
                    { id: "desktop", label: "Desktop", Icon: Monitor },
                    { id: "mobile", label: "Mobile", Icon: Smartphone },
                    { id: "xr", label: "XR", Icon: Glasses },
                  ] as const
                ).map(({ id, label, Icon }) => {
                  const on = device === id;
                  return (
                    <button
                      key={id}
                      type="button"
                      aria-pressed={on}
                      onClick={() => setDevice(id)}
                      className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-[12px] font-semibold transition-colors ${
                        on
                          ? "bg-shell-surface-active text-shell-text"
                          : "text-shell-text-secondary hover:text-shell-text"
                      }`}
                    >
                      <Icon size={13} />
                      {label}
                    </button>
                  );
                })}
              </div>
            </div>

            {isXr && (
              <p className="text-[11.5px] leading-relaxed text-shell-text-secondary">
                <span className="font-semibold text-shell-text">XR / VR.</span> Phase 1 shows the
                headset framing only. Real WebXR play, with a world-pinned Exit to taOS panel in
                view, lands in a later phase.
              </p>
            )}
          </div>

          {/* ---- build log / skills panel (illustrative) ---- */}
          <aside className="overflow-hidden rounded-2xl border border-shell-border bg-shell-surface shadow-card">
            <div className="flex items-center gap-2 border-b border-shell-border px-3.5 py-3">
              <ListTree size={15} className="text-accent" />
              <h3 className="text-[13px] font-bold">Build log</h3>
              <span className="ml-auto text-[11px] text-shell-text-tertiary">
                {sourcePrompt ? "live" : "preview"}
              </span>
            </div>
            <div className="max-h-[470px] overflow-auto p-1.5">
              {buildLog.map((step, i) => (
                <div
                  key={i}
                  className={`grid grid-cols-[26px_1fr] gap-2 rounded-xl px-2.5 py-2 hover:bg-shell-surface ${
                    step.director ? "border-l-2 border-accent/30" : ""
                  }`}
                >
                  <span
                    className={`mt-0.5 grid h-[22px] w-[22px] place-items-center rounded-lg ${
                      step.state === "done"
                        ? "bg-emerald-500/15 text-emerald-400"
                        : step.state === "run"
                          ? "bg-accent-soft text-accent"
                          : "bg-shell-surface text-shell-text-tertiary"
                    }`}
                  >
                    {step.state === "done" ? (
                      <Check size={13} />
                    ) : step.state === "run" ? (
                      <Loader2 size={13} className="animate-spin motion-reduce:animate-none" />
                    ) : (
                      <Clock size={13} />
                    )}
                  </span>
                  <div>
                    <div className="text-[12px] font-bold">{step.who}</div>
                    <div className="mt-0.5 text-[11.5px] leading-snug text-shell-text-secondary">
                      {step.what}
                    </div>
                    <span className="mt-1 inline-block rounded-full bg-shell-surface-active px-1.5 py-0.5 text-[9.5px] font-bold uppercase tracking-wide text-shell-text-secondary">
                      {step.tag}
                    </span>
                  </div>
                </div>
              ))}
              <p className="px-2.5 pb-1 pt-2 text-[11px] leading-relaxed text-shell-text-tertiary">
                {sourcePrompt
                  ? "Steps from your latest create run. Full offline model generation arrives in a later phase."
                  : "An illustrative trace of the agent build pipeline. Generate from Create or pick a template to load a live scene."}
              </p>
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}
