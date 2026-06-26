"use client";

import { Repeat, RotateCw, Trash2, X } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import type { Template } from "@/lib/types";

interface HeaderActionsProps {
  chatId: string;
  status: string;
  isTerminal: boolean;
  template: Template | null;
  onCancel: () => void;
}

export function HeaderActions({
  chatId,
  status,
  isTerminal,
  template,
  onCancel,
}: HeaderActionsProps) {
  const router = useRouter();
  const [retrying, setRetrying] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const handleRetry = async () => {
    setRetrying(true);
    try {
      const res = await fetch(`/api/daemon/chats/${chatId}/rerun`, {
        method: "POST",
      });
      if (!res.ok) {
        window.alert(
          "Couldn't start a new run — Chorus didn't respond. Try restarting it from your terminal: chorus start",
        );
        setRetrying(false);
        return;
      }
      const body = (await res.json()) as {
        ok: boolean;
        data?: { slug?: string; id?: string };
        error?: { code?: string; message?: string };
      };
      // Daemon returns HTTP 200 with `{ok: false, error: {...}}` for
      // validation/conflict failures (e.g. rerun-while-active). Without
      // surfacing it, the button just stops spinning and the user has
      // no idea why nothing happened.
      if (!body.ok) {
        const msg = body.error?.message ?? "Unknown error from Chorus.";
        window.alert(`Couldn't start a new run: ${msg}`);
        setRetrying(false);
        return;
      }
      const target = body.data?.slug ?? body.data?.id;
      if (target) {
        router.push(`/runs/${target}`);
        router.refresh();
      } else {
        // ok:true with no slug/id is a daemon-side bug, but we still
        // need to unstick the button.
        window.alert(
          "Chorus accepted the retry but didn't return a chat id. Refresh and try again.",
        );
        setRetrying(false);
      }
    } catch {
      window.alert("Retry failed. Network error.");
      setRetrying(false);
    }
  };

  const handleDelete = async () => {
    if (
      !window.confirm(
        "Delete this chat permanently? This removes all reviewer output and history. You cannot undo this.",
      )
    ) {
      return;
    }
    setDeleting(true);
    try {
      const res = await fetch(`/api/daemon/chats/${chatId}`, {
        method: "DELETE",
      });
      if (res.ok) {
        router.push("/runs");
        router.refresh();
      } else {
        window.alert(
          "Couldn't delete this chat — Chorus didn't respond. Try restarting it from your terminal: chorus start",
        );
        setDeleting(false);
      }
    } catch {
      window.alert("Delete failed. Network error.");
      setDeleting(false);
    }
  };

  return (
    <div className="flex shrink-0 items-center gap-2">
      {/* Pause button removed — was decorative, no backend support for
          mid-run pause. Resume after pause requires session state we
          don't have today. Will reintroduce when the daemon gains real
          pause/resume in v0.8. */}
      {status === "cancelled" || status === "failed" ? (
        <button
          type="button"
          disabled={retrying}
          onClick={handleRetry}
          className="flex items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-3 py-1.5 text-xs font-medium text-primary transition hover:border-primary hover:bg-primary/20 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <RotateCw className={`h-3.5 w-3.5 ${retrying ? "animate-spin" : ""}`} />
          {retrying ? "Restarting…" : "Retry"}
        </button>
      ) : (
        <button
          type="button"
          disabled={isTerminal}
          onClick={async () => {
            await fetch(`/api/daemon/chats/${chatId}/cancel`, { method: "POST" });
            onCancel();
          }}
          className="flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-destructive/40 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-50"
        >
          <X className="h-3.5 w-3.5" />
          Cancel
        </button>
      )}
      {/* Once the chat is terminal, give the user a one-click re-entry
          to /new with the same template pre-selected. Without this
          button users had to find the template manually. */}
      {isTerminal && template && (
        <Link
          href={`/new?template=${encodeURIComponent(template.id)}`}
          className="flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-primary/40 hover:text-primary"
        >
          <Repeat className="h-3.5 w-3.5" />
          Run again
        </Link>
      )}
      <button
        type="button"
        disabled={deleting}
        onClick={handleDelete}
        className="flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:border-destructive/40 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Trash2 className="h-3.5 w-3.5" />
        {deleting ? "Deleting…" : "Delete"}
      </button>
    </div>
  );
}
