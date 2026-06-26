import { CheckCircle2, AlertTriangle, Loader2 } from "lucide-react";
import type { ParticipantState } from "./types";

/**
 * Small status chip rendered in the top-right of every ParticipantCard.
 * Mirrors the card-level border colour but with crisp text + icon, so the
 * grid is scannable at a glance even when the user is staring at the
 * stream area in the middle.
 */
export function StateBadge({ state }: { state: ParticipantState }) {
  switch (state) {
    case "done":
      return (
        <span className="flex items-center gap-1 rounded-md bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-medium text-emerald-400">
          <CheckCircle2 className="h-3 w-3" /> DONE
        </span>
      );
    case "errored":
      return (
        <span className="flex items-center gap-1 rounded-md bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" /> ERRORED
        </span>
      );
    case "working":
      return (
        <span className="flex items-center gap-1 rounded-md bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
          <Loader2 className="h-3 w-3 animate-spin" /> WORKING
        </span>
      );
    case "pending":
      return (
        <span className="flex items-center gap-1 rounded-md bg-muted/40 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground/80">
          QUEUED
        </span>
      );
    default:
      return (
        <span className="rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
          IDLE
        </span>
      );
  }
}
