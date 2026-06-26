"use client";

import { useState } from "react";
import { chatDisplayTitle } from "@/lib/chat-title";

/**
 * Renders the chat brief at the top of the run page.
 *
 * The title is one CSS-truncated line that uses the full row width and
 * ellipsizes only when the viewport actually runs out of space. Click
 * anywhere on the title to reveal the full `work` blob (persona prompt
 * + user request + inlined files) in a scrollable <pre>; click again
 * to collapse. No separate expander button — the entire title row IS
 * the toggle.
 */
// Anything longer than this gets a click-to-expand affordance even when
// it fits on one line — at narrow viewports it'll truncate, and even at
// wide widths the user often wants to read it wrapped instead of
// scanning horizontally.
const BRIEF_EXPANDABLE_THRESHOLD = 100;

export function BriefHeading({ work }: { work: string }) {
  const [expanded, setExpanded] = useState(false);
  const displayTitle = chatDisplayTitle(work);
  const hasExpandableBody =
    work.trim() !== displayTitle.trim() ||
    work.length > BRIEF_EXPANDABLE_THRESHOLD;

  return (
    <div className="min-w-0">
      <button
        type="button"
        onClick={hasExpandableBody ? () => setExpanded((e) => !e) : undefined}
        title={hasExpandableBody ? (expanded ? "Hide full brief" : "Click to show full brief") : displayTitle}
        disabled={!hasExpandableBody}
        className="block w-full min-w-0 text-left disabled:cursor-default"
      >
        <h1 className="truncate text-sm font-medium tracking-tight">
          {displayTitle}
        </h1>
      </button>
      {hasExpandableBody && expanded && (
        <pre className="mt-2 max-h-72 overflow-auto whitespace-pre-wrap break-words rounded-md border border-border bg-background px-5 py-4 font-mono text-[12px] leading-relaxed text-foreground/90">
          {work}
        </pre>
      )}
    </div>
  );
}
