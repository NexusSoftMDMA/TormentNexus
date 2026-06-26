"use client";
import { track } from "@/lib/track";
import { FAQ_ITEMS } from "./faq-items";

export default function FAQ() {
  return (
    <section id="faq" className="border-t border-[var(--tt-border)]">
      <div className="max-w-3xl mx-auto px-5 sm:px-8 py-14 sm:py-28">
        <div className="text-center mb-8 sm:mb-10">
          <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--tt-fg-dim)] mb-3">FAQ</p>
          <h2 className="text-[26px] sm:text-[38px] leading-[1.1] tracking-[-0.02em] font-semibold text-[var(--tt-fg)]">
            Common questions
          </h2>
        </div>
        <div className="space-y-2">
          {FAQ_ITEMS.map((item, i) => (
            <details
              key={i}
              onToggle={(e) => {
                if ((e.currentTarget as HTMLDetailsElement).open)
                  track("faq_open", { question: item.q });
              }}
              className="group rounded-[var(--tt-radius-lg)] border border-[var(--tt-border)] bg-[var(--tt-panel)] open:bg-[var(--tt-raised)] transition-colors"
            >
              <summary className="flex items-center justify-between cursor-pointer list-none gap-4 px-5 py-4">
                <h3 className="text-[var(--tt-fg)] font-medium text-[15px] sm:text-[16px] tracking-[-0.005em]">
                  {item.q}
                </h3>
                <span className="text-[var(--tt-fg-dim)] text-xl leading-none transition-transform group-open:rotate-45 select-none">
                  +
                </span>
              </summary>
              <p className="px-5 pb-5 -mt-1 text-[13.5px] text-[var(--tt-fg-muted)] leading-relaxed">
                {item.a}
              </p>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
