"use client";
import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { FEATURES } from "@/data/features";
import { Check, Activity } from "lucide-react";

export default function FeatureShowcase() {
  const [active, setActive] = useState(FEATURES[0].id);
  const feature = FEATURES.find((f) => f.id === active)!;

  return (
    <section id="features" className="max-w-[1320px] mx-auto px-5 sm:px-8 py-16 sm:py-24">
      <div className="text-center mb-8 sm:mb-10">
        <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--tt-fg-dim)] mb-3">
          What you&apos;ll see
        </p>
        <h2 className="text-[28px] sm:text-[44px] leading-[1.1] tracking-[-0.02em] font-semibold text-[var(--tt-fg)] max-w-2xl mx-auto">
          One dashboard. <span className="text-[var(--tt-brand)]">Every agent.</span>
        </h2>
      </div>

      {/* Tabs — horizontally scrollable on mobile, wraps on larger screens */}
      <div className="-mx-5 sm:mx-auto mb-8 sm:mb-10 sm:w-fit overflow-x-auto sm:overflow-visible">
        <div className="flex sm:flex-wrap items-center sm:justify-center gap-1 px-5 sm:px-1 py-1 mx-auto sm:rounded-[var(--tt-radius-lg)] sm:border sm:border-[var(--tt-border)] sm:bg-[var(--tt-panel)] w-max sm:w-auto">
        {FEATURES.map((f) => (
          <button
            key={f.id}
            onClick={() => setActive(f.id)}
            className={`relative h-9 px-4 rounded-[var(--tt-radius)] text-[12.5px] font-medium tracking-tight transition-colors whitespace-nowrap ${
              active === f.id
                ? "tt-tint-2 text-[var(--tt-fg)]"
                : "text-[var(--tt-fg-muted)] hover:text-[var(--tt-fg)]"
            }`}
          >
            {f.label}
            {active === f.id && (
              <span aria-hidden className="absolute inset-x-3 -bottom-px h-px bg-[var(--tt-brand)]" />
            )}
          </button>
        ))}
        </div>
      </div>

      <AnimatePresence mode="wait">
        <motion.div
          key={feature.id}
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -12 }}
          transition={{ duration: 0.25 }}
        >
          <h3 className="text-center text-[20px] sm:text-[32px] leading-[1.2] tracking-[-0.02em] font-semibold text-[var(--tt-fg)] max-w-3xl mx-auto mb-8 sm:mb-10">
            {feature.headline}
          </h3>

          {/* Screenshot */}
          <div className="relative mb-8 sm:mb-10">
            <div aria-hidden className="absolute -inset-x-6 -inset-y-6 pointer-events-none bg-gradient-to-tr from-[color:var(--tt-brand-glow)] via-transparent to-emerald-500/5 blur-3xl" />
            <div className="relative rounded-[var(--tt-radius-lg)] sm:rounded-[var(--tt-radius-xl)] overflow-hidden border border-[var(--tt-border-strong)] bg-[var(--tt-panel)] shadow-[0_30px_120px_-30px_rgba(96,165,250,0.30)]">
              <div className="flex items-center gap-1.5 px-3 sm:px-4 h-9 bg-[var(--tt-raised)] border-b border-[var(--tt-border)]">
                <span className="w-2.5 h-2.5 rounded-full bg-rose-400/50" />
                <span className="w-2.5 h-2.5 rounded-full bg-amber-400/50" />
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-400/50" />
                <span className="ml-3 inline-flex items-center gap-1.5 text-[11px] font-mono text-[var(--tt-fg-dim)] truncate">
                  <Activity size={11} className="text-[var(--tt-brand)]" />
                  tokentelemetry · {feature.label.toLowerCase()}
                </span>
              </div>
              <img
                src={feature.screenshot}
                alt={feature.label}
                className="block w-full h-auto"
                loading="lazy"
                decoding="async"
              />
            </div>
          </div>

          {/* Bullets */}
          <div className="grid md:grid-cols-3 gap-3">
            {feature.bullets.map((b, i) => (
              <div
                key={i}
                className="flex gap-3 p-4 rounded-[var(--tt-radius-lg)] border border-[var(--tt-border)] bg-[var(--tt-panel)] hover:border-[var(--tt-border-strong)] transition-colors"
              >
                <span className="mt-0.5 h-5 w-5 rounded-md bg-[color:var(--tt-brand-glow)] border border-[color:var(--tt-brand)]/25 grid place-items-center shrink-0">
                  <Check size={11} className="text-[var(--tt-brand)]" strokeWidth={3} />
                </span>
                <p className="text-[13.5px] text-[var(--tt-fg-muted)] leading-relaxed">{b}</p>
              </div>
            ))}
          </div>
        </motion.div>
      </AnimatePresence>
    </section>
  );
}
