import { Type, Square, Circle, Image, Plus, Star, Sparkles, AlignLeft, Italic, Underline, Undo, Download } from "lucide-react";
import type { CanvasElement } from "./types";

const ELEMENT_TILES: { label: string; icon: typeof Type }[] = [
  { label: "Text", icon: Type },
  { label: "Shape", icon: Square },
  { label: "Circle", icon: Circle },
  { label: "Image", icon: Image },
  { label: "Line", icon: Plus },
  { label: "Star", icon: Star },
];

const BRAND_SWATCHES = [
  "#6b7689",
  "#2c3142",
  "#cdd3df",
  "#3d8f7a",
  "#c98b5b",
  "#ffffff",
];

const COLOR_SWATCHES = [
  { hex: "#ffffff", active: true },
  { hex: "#cdd3df", active: false },
  { hex: "#6b7689", active: false },
  { hex: "#3d8f7a", active: false },
  { hex: "#c98b5b", active: false },
];

const MAGIC_CHIPS = ["Make it bolder", "Match brand", "Rewrite copy"];

export interface DesignViewProps {
  canvasElements?: CanvasElement[];
}

export function DesignView({ canvasElements = [] }: DesignViewProps) {
  return (
    <>
      {/* view header */}
      <div
        className="flex flex-none items-center gap-3 border-b border-shell-border px-[22px]"
        style={{ height: "54px" }}
      >
        <h2 className="text-[17px] font-bold tracking-[-0.02em]">Untitled poster</h2>
        <span className="text-[12px] text-shell-text-tertiary">Poster · 1080 x 1350</span>
        <div className="ml-auto flex gap-1.5">
          <button
            type="button"
            className="flex h-[30px] items-center gap-1.5 rounded-[9px] border border-shell-border bg-shell-surface px-3 text-[11.5px] font-semibold text-shell-text-secondary"
          >
            <Undo size={14} />
            Undo
          </button>
          <button
            type="button"
            className="flex h-[30px] items-center gap-1.5 rounded-[9px] px-3 text-[11.5px] font-semibold text-white"
            style={{ background: "linear-gradient(135deg,#a9b0c2,#8b92a3)", border: "none" }}
          >
            <Download size={14} />
            Export
          </button>
        </div>
      </div>

      {/* three-column body */}
      <div className="flex min-h-0 flex-1">
        {/* left elements panel */}
        <div className="w-[210px] flex-none overflow-auto border-r border-shell-border bg-shell-bg-deep px-3 py-3.5">
          <div className="mb-2 ml-0.5 text-[10.5px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
            Add
          </div>
          <div className="mb-4 grid grid-cols-2 gap-2">
            {ELEMENT_TILES.map(({ label, icon: Icon }) => (
              <button
                key={label}
                type="button"
                aria-label={label}
                className="flex aspect-square cursor-pointer flex-col items-center justify-center rounded-[11px] border border-shell-border bg-shell-surface text-shell-text-secondary transition-all hover:-translate-y-0.5 hover:bg-shell-surface-active hover:text-shell-text"
              >
                <Icon size={22} />
              </button>
            ))}
          </div>
          <div className="mb-2 ml-0.5 text-[10.5px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
            Brand colors
          </div>
          <div className="flex flex-wrap gap-[7px]">
            {BRAND_SWATCHES.map((hex) => (
              <button
                key={hex}
                type="button"
                aria-label={hex}
                className="h-6 w-6 cursor-pointer rounded-[7px]"
                style={{
                  background: hex,
                  border: "1px solid rgba(255,255,255,0.12)",
                }}
              />
            ))}
          </div>
        </div>

        {/* center canvas stage */}
        <div
          className="flex min-w-0 flex-1 flex-col"
          style={{
            background:
              "repeating-conic-gradient(rgba(255,255,255,0.018) 0% 25%, transparent 0% 50%) 0 0/22px 22px, var(--color-shell-bg, #1d1d1f)",
          }}
        >
          {/* stage bar */}
          <div className="flex h-10 flex-none items-center gap-2.5 border-b border-shell-border bg-shell-bg-deep px-[18px]">
            <div className="flex items-center gap-1.5 rounded-[8px] border border-shell-border bg-shell-surface px-2.5 py-[5px] text-[11px] font-semibold text-shell-text-secondary">
              <Type size={13} />
              Heading
            </div>
            <div className="flex items-center gap-1.5 rounded-[8px] border border-shell-border bg-shell-surface px-2.5 py-[5px] text-[11px] font-semibold text-shell-text-secondary">
              Sohne · 40
            </div>
            <span className="ml-auto text-[11.5px] font-mono text-shell-text-tertiary">68%</span>
          </div>

          {/* artboard */}
          <div className="flex flex-1 items-center justify-center overflow-hidden p-[26px]">
            <div
              className="relative overflow-hidden rounded-[8px]"
              style={{
                width: "360px",
                height: "450px",
                boxShadow: "0 24px 60px -16px rgba(0,0,0,0.6)",
                background:
                  "radial-gradient(130% 120% at 20% 12%, #5a6b86, transparent 55%), linear-gradient(150deg, #2c3142, #171a24)",
              }}
            >
              {/* decorative blob */}
              <div
                className="pointer-events-none absolute"
                style={{
                  bottom: "-40px",
                  right: "-40px",
                  width: "200px",
                  height: "200px",
                  borderRadius: "50%",
                  background: "radial-gradient(circle, #8b92a3, transparent 68%)",
                  opacity: 0.5,
                }}
              />
              {/* poster text content */}
              <div
                className="absolute text-[11px] font-bold uppercase tracking-[2px]"
                style={{ top: "42px", left: "32px", color: "rgba(255,255,255,0.7)" }}
              >
                taOS Studios
              </div>
              <div
                className="absolute font-extrabold leading-[1.02] text-white"
                style={{ top: "64px", left: "30px", right: "30px", fontSize: "40px", letterSpacing: "-1px" }}
              >
                Build it<br />your way.
              </div>
              <div
                className="absolute text-[14px] leading-[1.5]"
                style={{ top: "230px", left: "32px", right: "40px", color: "rgba(255,255,255,0.82)" }}
              >
                Dedicated studios for every project, running on hardware you already own.
              </div>
              <div
                className="absolute rounded-full bg-white text-[13px] font-bold"
                style={{ bottom: "36px", left: "32px", color: "#1d1d1f", padding: "9px 20px" }}
              >
                Get started
              </div>

              {canvasElements.map((el) =>
                el.type === "image" ? (
                  <img
                    key={el.id}
                    src={el.url}
                    alt={el.prompt}
                    className="absolute object-cover"
                    style={{
                      left: el.x,
                      top: el.y,
                      width: el.width,
                      height: el.height,
                      borderRadius: "6px",
                    }}
                  />
                ) : null,
              )}

              {/* selection box */}
              <div
                className="pointer-events-none absolute rounded-[4px]"
                style={{ left: "26px", top: "58px", right: "26px", height: "92px", border: "1.5px solid #a9b0c2" }}
              />
              {/* corner handles */}
              {[
                { left: "26px", top: "58px" },
                { left: "calc(100% - 26px)", top: "58px" },
                { left: "26px", top: "150px" },
                { left: "calc(100% - 26px)", top: "150px" },
              ].map((pos, i) => (
                <div
                  key={i}
                  className="pointer-events-none absolute h-[11px] w-[11px] rounded-full border-[1.5px] bg-white"
                  style={{ borderColor: "#a9b0c2", transform: "translate(-50%, -50%)", ...pos }}
                />
              ))}
            </div>
          </div>
        </div>

        {/* right properties panel */}
        <div className="flex w-[248px] flex-none flex-col gap-4 overflow-auto border-l border-shell-border p-4">
          {/* text */}
          <div>
            <label className="mb-2 block text-[11px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
              Text
            </label>
            <div className="flex gap-2">
              <div className="flex flex-1 items-center justify-between rounded-[10px] border border-shell-border bg-shell-surface px-[11px] py-[9px] text-[12.5px] font-semibold">
                Sohne <span className="text-shell-text-tertiary">&#9662;</span>
              </div>
              <div className="w-[62px] rounded-[10px] border border-shell-border bg-shell-surface py-[9px] text-center text-[12.5px] font-semibold font-mono">
                40
              </div>
            </div>
          </div>

          {/* style */}
          <div>
            <label className="mb-2 block text-[11px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
              Style
            </label>
            <div className="flex gap-[7px]">
              <button
                type="button"
                className="flex h-[34px] flex-1 items-center justify-center rounded-[9px] border border-shell-border bg-shell-surface-active text-[13px] font-extrabold text-shell-text"
              >
                B
              </button>
              <button
                type="button"
                className="flex h-[34px] flex-1 items-center justify-center rounded-[9px] border border-shell-border bg-shell-surface text-[13px] italic text-shell-text-secondary"
              >
                <Italic size={15} />
              </button>
              <button
                type="button"
                className="flex h-[34px] flex-1 items-center justify-center rounded-[9px] border border-shell-border bg-shell-surface text-[13px] text-shell-text-secondary underline"
              >
                <Underline size={15} />
              </button>
              <button
                type="button"
                className="flex h-[34px] flex-1 items-center justify-center rounded-[9px] border border-shell-border bg-shell-surface text-shell-text-secondary"
              >
                <AlignLeft size={15} />
              </button>
            </div>
          </div>

          {/* color */}
          <div>
            <label className="mb-2 block text-[11px] font-bold uppercase tracking-[0.06em] text-shell-text-tertiary">
              Color
            </label>
            <div className="flex flex-wrap gap-[7px]">
              {COLOR_SWATCHES.map(({ hex, active }) => (
                <button
                  key={hex}
                  type="button"
                  aria-label={hex}
                  className="h-6 w-6 cursor-pointer rounded-[7px]"
                  style={{
                    background: hex,
                    border: "1px solid rgba(255,255,255,0.12)",
                    outline: active ? "2px solid #a9b0c2" : "none",
                    outlineOffset: "2px",
                  }}
                />
              ))}
            </div>
          </div>

          {/* magic edits box */}
          <div
            className="rounded-[14px] p-[13px]"
            style={{
              border: "1px solid rgba(139,146,163,0.35)",
              background:
                "radial-gradient(120% 130% at 12% 10%, rgba(139,146,163,0.35), transparent 60%), var(--color-shell-surface, rgba(255,255,255,0.045))",
            }}
          >
            <div className="flex items-center gap-[7px] text-[12.5px] font-bold">
              <Sparkles size={15} className="text-shell-text-secondary" />
              Magic edits
            </div>
            <p className="mt-1.5 text-[11.5px] leading-[1.45] text-shell-text-secondary">
              Ask for a change in words. taOS restyles the selected layer, on your cluster.
            </p>
            <div className="mt-2.5 flex flex-wrap gap-1.5">
              {MAGIC_CHIPS.map((chip) => (
                <button
                  key={chip}
                  type="button"
                  className="rounded-full bg-shell-surface-active px-[10px] py-[5px] text-[11px] font-semibold text-shell-text-secondary"
                >
                  {chip}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
