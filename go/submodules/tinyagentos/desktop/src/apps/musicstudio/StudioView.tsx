import { Square, Play, Circle, Check } from "lucide-react";

/* Muted track color palette -- content, not chrome */
const TRACK_COLORS = {
  drum: "var(--ms-tk-drum, #c98b6b)",
  bass: "var(--ms-tk-bass, #6f8aa8)",
  keys: "var(--ms-tk-keys, #7faa90)",
  pad: "var(--ms-tk-pad, #9a87b0)",
  lead: "var(--ms-tk-lead, #c0a86a)",
} as const;

type TrackColor = keyof typeof TRACK_COLORS;

interface Track {
  id: string;
  label: string;
  color: TrackColor;
  volume: number;
  muted: boolean;
}

const TRACKS: Track[] = [
  { id: "drum", label: "Drums", color: "drum", volume: 72, muted: false },
  { id: "bass", label: "Bass", color: "bass", volume: 60, muted: false },
  { id: "keys", label: "Keys", color: "keys", volume: 54, muted: false },
  { id: "pad", label: "Pad", color: "pad", volume: 48, muted: true },
  { id: "lead", label: "Lead", color: "lead", volume: 66, muted: false },
];

interface Clip {
  trackId: string;
  label: string;
  left: number;
  width: number;
  opacity?: number;
  bars: number[];
}

const CLIPS: Clip[] = [
  { trackId: "drum", label: "Drums", left: 0, width: 288, bars: [40, 90, 30, 70, 40, 95, 30, 60, 40, 88, 30, 70] },
  { trackId: "bass", label: "Bassline", left: 72, width: 216, bars: [60, 50, 80, 45, 70, 55, 85, 50] },
  { trackId: "keys", label: "Rhodes", left: 144, width: 288, bars: [50, 65, 40, 75, 55, 60] },
  { trackId: "pad", label: "Pad (muted)", left: 216, width: 360, opacity: 0.5, bars: [30, 35, 32, 38] },
  { trackId: "lead", label: "Lead", left: 288, width: 180, bars: [80, 45, 90, 55, 70] },
];

const RULER_BARS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

const PIANO_KEYS = [
  { label: "C4", black: false },
  { label: "B3", black: true },
  { label: "A3", black: false },
  { label: "G3", black: true },
  { label: "F3", black: false },
  { label: "E3", black: true },
  { label: "D3", black: false },
];

interface PianoNote {
  left: number;
  top: number;
  width: number;
}

const PIANO_NOTES: PianoNote[] = [
  { left: 8, top: 14, width: 32 },
  { left: 44, top: 46, width: 64 },
  { left: 112, top: 14, width: 32 },
  { left: 148, top: 78, width: 48 },
  { left: 200, top: 46, width: 64 },
  { left: 268, top: 30, width: 32 },
  { left: 304, top: 62, width: 80 },
];

const KNOBS = ["Swing", "Drive", "Tone", "Width"];
const FX_ROWS = [
  { label: "Reverb", on: true },
  { label: "Tape delay", on: false },
  { label: "Sidechain", on: true },
];

export function StudioView() {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* transport bar */}
      <div
        className="flex flex-none items-center gap-3.5 border-b border-shell-border bg-shell-bg-deep px-5"
        style={{ height: "58px" }}
      >
        <div className="flex items-center gap-2">
          {/* stop */}
          <button
            type="button"
            aria-label="Stop"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-shell-border bg-shell-surface text-shell-text"
          >
            <Square size={16} />
          </button>
          {/* play */}
          <button
            type="button"
            aria-label="Play"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] text-white"
            style={{ background: "linear-gradient(135deg, var(--color-accent-strong, #a9b0c2), var(--color-accent, #8b92a3))" }}
          >
            <Play size={16} fill="currentColor" />
          </button>
          {/* record */}
          <button
            type="button"
            aria-label="Record"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-shell-border bg-shell-surface"
          >
            <Circle size={16} fill="#e06464" className="text-[#e06464]" />
          </button>
        </div>

        <div className="ml-1.5 flex items-center gap-4">
          <span className="font-mono text-[19px] font-bold tracking-[-0.01em] tabular-nums">
            003 . 2 . 1
          </span>
          <div className="flex flex-col leading-[1.1]">
            <span className="text-[14px] font-bold tabular-nums">92</span>
            <span className="text-[9.5px] uppercase tracking-[0.06em] text-shell-text-tertiary">BPM</span>
          </div>
          <div className="flex flex-col leading-[1.1]">
            <span className="text-[14px] font-bold tabular-nums">4/4</span>
            <span className="text-[9.5px] uppercase tracking-[0.06em] text-shell-text-tertiary">Time</span>
          </div>
          <div className="flex flex-col leading-[1.1]">
            <span className="text-[14px] font-bold">A min</span>
            <span className="text-[9.5px] uppercase tracking-[0.06em] text-shell-text-tertiary">Key</span>
          </div>
        </div>

        <div className="ml-auto">
          <div className="flex items-center gap-1.5 rounded-[10px] border border-shell-border bg-shell-surface px-3 py-[7px]">
            <Check size={14} className="text-shell-text-secondary" />
            <span className="text-[11.5px] font-semibold text-shell-text-secondary">Saved to Workspace</span>
          </div>
        </div>
      </div>

      {/* arrange area */}
      <div className="flex min-h-0 flex-1">
        {/* track list column */}
        <div className="w-[188px] flex-none overflow-auto border-r border-shell-border bg-shell-bg-deep">
          {/* ruler label */}
          <div className="flex h-7 items-center border-b border-shell-border px-3">
            <span className="text-[10px] font-bold uppercase tracking-[0.05em] text-shell-text-tertiary">
              Tracks
            </span>
          </div>

          {TRACKS.map((track, i) => (
            <div
              key={track.id}
              className={`flex flex-col justify-center gap-1.5 border-b border-shell-border px-3 py-[9px] ${i === 0 ? "bg-shell-surface" : ""}`}
              style={{ height: "62px" }}
            >
              <div className="flex items-center gap-2">
                <span
                  className="h-[9px] w-[9px] flex-none rounded-[3px]"
                  style={{ background: TRACK_COLORS[track.color] }}
                />
                <span className="text-[12.5px] font-semibold">{track.label}</span>
                <div className="ml-auto flex gap-1">
                  <span className={`flex h-[18px] w-[18px] items-center justify-center rounded-[5px] text-[9.5px] font-extrabold ${track.muted ? "bg-accent text-white" : "bg-shell-surface-active text-shell-text-tertiary"}`}>
                    M
                  </span>
                  <span className="flex h-[18px] w-[18px] items-center justify-center rounded-[5px] bg-shell-surface-active text-[9.5px] font-extrabold text-shell-text-tertiary">
                    S
                  </span>
                </div>
              </div>
              <div className="relative h-1 rounded-full bg-shell-surface-active">
                <span
                  className="absolute inset-y-0 left-0 rounded-full bg-shell-text-tertiary"
                  style={{ width: `${track.volume}%` }}
                />
              </div>
            </div>
          ))}
        </div>

        {/* scrolling timeline */}
        <div
          className="relative min-w-0 flex-1 overflow-auto"
          style={{
            background:
              "linear-gradient(90deg, var(--color-shell-border, rgba(255,255,255,0.08)) 1px, transparent 1px) 0 0 / 72px 100%, var(--color-shell-bg, #1d1d1f)",
          }}
        >
          {/* bar ruler */}
          <div className="sticky top-0 z-10 flex h-7 border-b border-shell-border bg-shell-bg-deep">
            {RULER_BARS.map((b) => (
              <div
                key={b}
                className="flex w-[72px] flex-none items-center border-r border-shell-border pl-[7px] text-[10px] font-bold text-shell-text-tertiary"
              >
                {b}
              </div>
            ))}
          </div>

          {/* lanes */}
          {TRACKS.map((track) => {
            const clip = CLIPS.find((c) => c.trackId === track.id);
            return (
              <div
                key={track.id}
                className="relative border-b border-shell-border"
                style={{ height: "62px" }}
              >
                {clip && (
                  <div
                    className={`absolute top-[9px] flex items-end overflow-hidden rounded-[8px] border border-white/[0.18] px-[7px] pb-[5px] ${track.id === "drum" ? "outline outline-2 outline-offset-[-1px] outline-white" : ""}`}
                    style={{
                      left: `${clip.left}px`,
                      width: `${clip.width}px`,
                      height: "44px",
                      background: TRACK_COLORS[track.color],
                      opacity: clip.opacity ?? 1,
                      boxShadow: "0 3px 10px rgba(0,0,0,0.3)",
                    }}
                  >
                    <span className="absolute left-2 top-[5px] text-[9.5px] font-bold text-white/90">
                      {clip.label}
                    </span>
                    <div className="flex h-[18px] w-full items-end gap-[1.5px] opacity-65">
                      {clip.bars.map((h, idx) => (
                        <span
                          key={idx}
                          className="flex-1 rounded-[1px] bg-white/85"
                          style={{ height: `${h}%` }}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}

          {/* playhead */}
          <div
            className="pointer-events-none absolute inset-y-0 z-20 w-[2px]"
            style={{
              left: "230px",
              background: "var(--color-accent-strong, #a9b0c2)",
              boxShadow: "0 0 8px var(--color-accent-glow, rgba(139,146,163,0.35))",
            }}
          >
            <span
              className="absolute left-[-4px] top-0 block h-0 w-0 border-[5px] border-transparent"
              style={{ borderTopColor: "var(--color-accent-strong, #a9b0c2)" }}
            />
          </div>
        </div>

        {/* right inspector */}
        <div className="flex w-[236px] flex-none flex-col gap-4 overflow-auto border-l border-shell-border p-4">
          <div className="flex items-center gap-[9px] text-[14px] font-bold tracking-[-0.01em]">
            <span
              className="h-[11px] w-[11px] flex-none rounded-[4px]"
              style={{ background: TRACK_COLORS.drum }}
            />
            Drums
          </div>
          <p className="mt-[-10px] text-[11px] text-shell-text-tertiary">
            taOS Drum Kit - Boom Bap
          </p>

          <div className="grid grid-cols-2 gap-3.5">
            {KNOBS.map((label) => (
              <div key={label} className="flex flex-col items-center gap-[7px]">
                <div className="relative h-[54px] w-[54px] rounded-full border border-shell-border bg-shell-surface">
                  <span
                    className="absolute left-1/2 top-[7px] h-[18px] w-[3px] -translate-x-1/2 origin-bottom rounded-[2px] bg-accent"
                    style={{ transform: "translateX(-50%) rotate(38deg)", transformOrigin: "bottom center" }}
                  />
                </div>
                <span className="text-[10.5px] font-semibold uppercase tracking-[0.04em] text-shell-text-tertiary">
                  {label}
                </span>
              </div>
            ))}
          </div>

          <div className="flex flex-col gap-2">
            {FX_ROWS.map((fx) => (
              <div
                key={fx.label}
                className="flex items-center gap-2.5 rounded-[11px] border border-shell-border bg-shell-surface px-3 py-2.5"
              >
                <span className="text-[12.5px] font-semibold">{fx.label}</span>
                <div
                  className="relative ml-auto h-[19px] w-[34px] rounded-full"
                  style={{ background: fx.on ? "var(--color-accent, #8b92a3)" : "var(--color-shell-surface-active, rgba(255,255,255,0.10))" }}
                >
                  <span
                    className="absolute top-[2px] h-[15px] w-[15px] rounded-full bg-white transition-all"
                    style={{ left: fx.on ? "17px" : "2px" }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* piano roll strip */}
      <div className="flex h-[128px] flex-none border-t border-shell-border bg-shell-bg-deep">
        {/* key labels */}
        <div className="flex w-[54px] flex-none flex-col border-r border-shell-border font-mono">
          {PIANO_KEYS.map((k) => (
            <div
              key={k.label}
              className={`flex flex-1 items-center border-b border-shell-border px-[5px] py-[1px] text-[8px] text-shell-text-tertiary ${k.black ? "bg-black/[0.18]" : ""}`}
            >
              {k.label}
            </div>
          ))}
        </div>

        {/* note grid */}
        <div
          className="relative flex-1 overflow-hidden"
          style={{
            background:
              "linear-gradient(90deg, var(--color-shell-border, rgba(255,255,255,0.08)) 1px, transparent 1px) 0 0 / 36px 100%, linear-gradient(0deg, var(--color-shell-border, rgba(255,255,255,0.08)) 1px, transparent 1px) 0 0 / 100% 16px",
          }}
        >
          {PIANO_NOTES.map((note, i) => (
            <span
              key={i}
              className="absolute h-[13px] rounded-[3px]"
              style={{
                left: `${note.left}px`,
                top: `${note.top}px`,
                width: `${note.width}px`,
                background: TRACK_COLORS.bass,
                boxShadow: "0 1px 3px rgba(0,0,0,0.4)",
              }}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
