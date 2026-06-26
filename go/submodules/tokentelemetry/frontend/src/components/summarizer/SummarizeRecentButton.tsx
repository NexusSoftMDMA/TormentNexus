"use client";

import { useState } from "react";
import { Sparkles, Loader2, Check } from "lucide-react";
import { Button } from "@/components/ui";
import { summarizeRecent, type RecentTally } from "@/lib/summarizer";

/**
 * Dashboard action: batch-summarize the N most recent sessions and show the
 * returned tally (requested / summarized / skipped / failed) inline.
 */
export default function SummarizeRecentButton({ limit = 10 }: { limit?: number }) {
  const [running, setRunning] = useState(false);
  const [tally, setTally] = useState<RecentTally | null>(null);
  const [error, setError] = useState<string | null>(null);

  const run = async () => {
    setRunning(true);
    setError(null);
    try {
      setTally(await summarizeRecent(limit));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to summarize.");
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {tally && !running && (
        <span className="hidden md:flex items-center gap-1.5 text-[11px] text-[var(--tt-fg-muted)]">
          <Check size={12} className="text-[var(--tt-success-fg)]" />
          {tally.summarized} summarized · {tally.skipped} cached · {tally.failed} failed
        </span>
      )}
      {error && <span className="hidden md:inline text-[11px] text-[var(--tt-danger-fg)]">{error}</span>}
      <Button variant="secondary" size="md" onClick={run} disabled={running} title={`Summarize the ${limit} most recent sessions`}>
        {running ? <><Loader2 size={14} className="animate-spin" /> Summarizing…</> : <><Sparkles size={14} /> Summarize recent</>}
      </Button>
    </div>
  );
}
