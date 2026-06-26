import Link from "next/link";
import { ArrowRight } from "lucide-react";

// Hermes caduceus — same line-art as the dashboard's brand mark.
function Caduceus({ size = 12 }: { size?: number }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size} height={size} viewBox="0 0 24 24"
      fill="none" stroke="currentColor" strokeWidth="2"
      strokeLinecap="round" strokeLinejoin="round"
      aria-hidden
    >
      <line x1="12" y1="3" x2="12" y2="21" />
      <path d="M8 5L4.5 2.5L6.5 6" />
      <path d="M8 5L5.5 6.5L7.5 8" />
      <path d="M16 5L19.5 2.5L17.5 6" />
      <path d="M16 5L18.5 6.5L16.5 8" />
      <path d="M9 9C7.5 11 7.5 13 9 15" />
      <path d="M9 15C10 17 11 17 12 16" />
      <path d="M15 9C16.5 11 16.5 13 15 15" />
      <path d="M15 15C14 17 13 17 12 16" />
    </svg>
  );
}

// One "beat" of marquee content. The track renders this *twice* side-by-side
// so the CSS `translateX(-50%)` produces a seamless loop. Three short beats
// is the most a person can read sideways at a glance — no more.
function MarqueeBeat() {
  const sep = (
    <span aria-hidden className="inline-flex items-center text-[#eab308]/40 select-none mx-12">
      <span className="text-[10px]">◆</span>
    </span>
  );
  return (
    <div className="flex items-center whitespace-nowrap text-[13px] font-medium tracking-[-0.005em] text-[var(--tt-fg)]">
      <span className="inline-flex items-center gap-2.5">
        <span className="inline-flex items-center gap-1.5 text-[#eab308]">
          <Caduceus size={13} />
          <span className="text-[10px] font-bold uppercase tracking-[0.22em]">New</span>
        </span>
        <span>
          <strong className="text-[var(--tt-fg)]">Hermes Agent</strong>
          <span className="text-[var(--tt-fg-muted)]"> observability is live</span>
        </span>
      </span>
      {sep}
      <span>
        See what your <strong className="text-[#eab308]">autonomous agents</strong> actually do
      </span>
      {sep}
      <span className="inline-flex items-center gap-1.5 text-[#eab308] font-semibold">
        Open the /hermes dashboard <ArrowRight size={12} />
      </span>
      {sep}{/* trailing separator so beat-1 → beat-2 has the same rhythm as
              the in-beat joins; without it the loop boundary is visible */}
    </div>
  );
}

export default function MarqueeBanner() {
  return (
    <Link
      href="#hermes"
      aria-label="Jump to the Hermes Agent section"
      className="block relative overflow-hidden border-b border-[var(--tt-border)] bg-[linear-gradient(90deg,rgba(234,179,8,0.06),rgba(234,179,8,0.02),rgba(234,179,8,0.06))] group"
    >
      {/* edge fades so text doesn't hit the viewport edges hard */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-y-0 left-0 w-16 z-10"
        style={{ background: "linear-gradient(90deg, var(--tt-canvas), transparent)" }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-y-0 right-0 w-16 z-10"
        style={{ background: "linear-gradient(270deg, var(--tt-canvas), transparent)" }}
      />
      <div className="h-10 flex items-center">
        <div className="tt-marquee-track flex shrink-0">
          {/* duplicate so the loop is seamless */}
          <MarqueeBeat />
          <MarqueeBeat />
        </div>
      </div>
    </Link>
  );
}
