import type { ReactNode } from "react";
import { Code2 } from "lucide-react";

/**
 * Shared monospace code/prose viewer.
 *
 * Used for: persona system prompts, template YAML, run-page brief expansion,
 * settings YAML editor preview. Previously each callsite hand-rolled its
 * own pre-block with subtly different font sizes, line heights, and text
 * colours — visible side-by-side as inconsistency. One primitive locks the
 * rhythm.
 *
 *   header chip:  Code2 icon + filename (mono 11px) + optional char-count
 *   body:         pre-formatted text at text-[12px] / leading-relaxed /
 *                 text-foreground/90 with whitespace-pre-wrap break-words
 *   footer:       optional dim text-[11px] strip (used by personas to show
 *                 the source-file hint and invoke command)
 *
 * The body slot can be a plain string, a <pre>, or arbitrary children — the
 * outer chrome stays identical.
 */
export function CodeBlock({
  filename,
  charCount,
  footer,
  maxHeightClassName = "max-h-[70vh]",
  children,
}: {
  filename: string;
  charCount?: number;
  footer?: ReactNode;
  maxHeightClassName?: string;
  children: ReactNode;
}) {
  // Flex column so callers can pass h-full / h-[calc(...)] on a parent
  // wrapper and the pre body will grow to fill the remaining height
  // (header + footer stay fixed-height). Without flex-col the pre's
  // max-height capped the layout regardless of the parent's height.
  return (
    <div className="flex h-full min-w-0 flex-col overflow-hidden rounded-lg border border-border bg-card">
      <div className="flex shrink-0 items-center justify-between gap-2 border-b border-border bg-card/60 px-4 py-2.5">
        <div className="flex min-w-0 items-center gap-2 font-mono text-[11px] text-muted-foreground">
          <Code2 className="h-3 w-3 shrink-0" />
          <span className="truncate">{filename}</span>
        </div>
        {typeof charCount === "number" && (
          <span className="font-mono text-[10px] text-muted-foreground/70">
            {charCount.toLocaleString()} chars
          </span>
        )}
      </div>
      <pre
        className={`${maxHeightClassName} min-h-0 flex-1 overflow-auto whitespace-pre-wrap break-words bg-background px-5 py-4 font-mono text-[12px] leading-relaxed text-foreground/90`}
      >
        {children}
      </pre>
      {footer && (
        <div className="flex shrink-0 items-center justify-between gap-2 border-t border-border bg-card/60 px-4 py-2.5 text-[11px] text-muted-foreground">
          {footer}
        </div>
      )}
    </div>
  );
}
