import { useState } from "react";
import { Share2, Check, Lock, Globe, Info } from "lucide-react";
import type { Template, Visibility } from "./types";

/* ------------------------------------------------------------------ */
/*  ShareView — publish card + test checklist                          */
/*                                                                     */
/*  Honesty: real store-publish is a later phase (ties to the optional- */
/*  install / Store work). "Share to Store" does not fake success; it   */
/*  surfaces a clear "publishing flows through the Store, coming" note. */
/* ------------------------------------------------------------------ */

const CHECKS = [
  "Runs on desktop",
  "Runs on mobile",
  "Fullscreen works, with an Exit to taOS control",
  "Holds frame rate on this device",
] as const;

export interface ShareViewProps {
  template: Template;
}

export function ShareView({ template }: ShareViewProps) {
  const [title, setTitle] = useState(template.title);
  const [description, setDescription] = useState(template.desc);
  const [visibility, setVisibility] = useState<Visibility>("community");
  const [showComing, setShowComing] = useState(false);

  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
      <div
        className="flex flex-col gap-3.5 p-[22px]"
        style={{ paddingBottom: "calc(22px + env(safe-area-inset-bottom, 0px))" }}
      >
        <header>
          <h2 className="text-[17px] font-bold tracking-[-0.02em]">Publish to the Store</h2>
          <p className="mt-1 text-[12.5px] text-shell-text-secondary">
            Share "{title}" with your taOS Store. Community games are reviewed before they go public.
          </p>
        </header>

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-[1.1fr_0.9fr]">
          {/* publish form */}
          <div className="flex flex-col gap-4 rounded-2xl border border-shell-border bg-shell-surface p-4 shadow-card">
            <div
              className="relative h-[150px] overflow-hidden rounded-2xl border border-shell-border-strong"
              style={{ background: template.cover }}
            >
              <span className="absolute left-2.5 top-2.5 rounded-full bg-black/40 px-2 py-1 text-[10px] font-bold uppercase tracking-wide text-white/85 backdrop-blur-sm">
                {template.genre}
              </span>
              <span
                className="absolute bottom-2.5 left-3 text-[17px] font-extrabold text-white"
                style={{ textShadow: "0 2px 10px rgba(0,0,0,0.6)" }}
              >
                {title || "Untitled"}
              </span>
            </div>

            <Field label="Game title">
              <input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full rounded-xl border border-shell-border bg-shell-bg-deep px-3.5 py-2.5 text-[13px] text-shell-text focus-visible:border-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/20"
              />
            </Field>

            <Field label="Description">
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full resize-none rounded-xl border border-shell-border bg-shell-bg-deep px-3.5 py-2.5 text-[13px] text-shell-text focus-visible:border-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/20"
              />
            </Field>

            <Field label="Visibility">
              <div className="flex gap-2">
                <VisibilityOption
                  on={visibility === "private"}
                  onClick={() => setVisibility("private")}
                  Icon={Lock}
                  title="Private"
                  desc="Only you and people you invite can play it."
                />
                <VisibilityOption
                  on={visibility === "community"}
                  onClick={() => setVisibility("community")}
                  Icon={Globe}
                  title="Community"
                  desc="Submit to the Store for everyone to find and install."
                />
              </div>
            </Field>

            <div className="flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={() => setShowComing(true)}
                className="flex h-[44px] items-center gap-2 rounded-full bg-gradient-to-br from-accent to-accent/70 px-5 text-[13px] font-bold text-white shadow-lg shadow-accent/20 transition-all hover:-translate-y-0.5 hover:brightness-105"
              >
                <Share2 size={17} />
                Share to Store
              </button>
              <span className="text-[11px] text-shell-text-tertiary">
                A reviewer checks community games before they go public.
              </span>
            </div>

            {showComing && (
              <div
                role="status"
                className="flex items-start gap-2.5 rounded-xl border border-accent/30 bg-accent-soft px-3.5 py-3"
              >
                <Info size={16} className="mt-0.5 flex-none text-accent" />
                <div className="text-[12.5px] leading-relaxed text-shell-text-secondary">
                  <span className="font-semibold text-shell-text">
                    Publishing flows through the Store, coming soon.
                  </span>{" "}
                  A later phase packages the game and submits it to your taOS Store for review and
                  install. Nothing is published yet.
                </div>
              </div>
            )}
          </div>

          {/* test checklist + meta */}
          <div className="flex flex-col gap-4">
            <div className="rounded-2xl border border-shell-border bg-shell-surface shadow-card">
              <div className="flex items-center gap-2 px-4 pt-3.5">
                <Check size={15} className="text-accent" />
                <h3 className="text-[13px] font-bold">Tested in Play</h3>
              </div>
              <div className="p-4">
                <ul className="flex flex-col gap-2">
                  {CHECKS.map((c) => (
                    <li key={c} className="flex items-center gap-2.5 text-[12.5px]">
                      <span className="grid h-5 w-5 flex-none place-items-center rounded-md bg-emerald-500/15 text-emerald-400">
                        <Check size={12} />
                      </span>
                      {c}
                    </li>
                  ))}
                </ul>
                <div className="my-4 h-px bg-shell-border" />
                <p className="text-[11px] leading-relaxed text-shell-text-tertiary">
                  Run the game in Play &amp; Test any time before you publish. These are the checks a
                  later phase will gate publishing on.
                </p>
              </div>
            </div>

            <div className="rounded-2xl border border-shell-border bg-shell-surface p-4 shadow-card">
              <dl className="flex flex-col gap-2.5 text-[12.5px]">
                <MetaRow label="Engine" value="taOS Game Runtime" />
                <MetaRow label="Preview" value="three.js (WebGL)" />
                <MetaRow label="Plays on" value="Desktop · Mobile · XR" />
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <span className="mb-1.5 block text-[12px] font-semibold text-shell-text-secondary">
        {label}
      </span>
      {children}
    </div>
  );
}

function VisibilityOption({
  on,
  onClick,
  Icon,
  title,
  desc,
}: {
  on: boolean;
  onClick: () => void;
  Icon: typeof Lock;
  title: string;
  desc: string;
}) {
  return (
    <button
      type="button"
      aria-pressed={on}
      onClick={onClick}
      className={`flex-1 rounded-xl border p-3 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/40 ${
        on
          ? "border-accent/40 bg-accent-soft"
          : "border-shell-border bg-shell-surface hover:bg-white/10"
      }`}
    >
      <div className="flex items-center gap-1.5 text-[12.5px] font-bold">
        <Icon size={14} className="text-accent" />
        {title}
      </div>
      <div className="mt-1 text-[11px] leading-snug text-shell-text-secondary">{desc}</div>
    </button>
  );
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <dt className="text-shell-text-secondary">{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}
