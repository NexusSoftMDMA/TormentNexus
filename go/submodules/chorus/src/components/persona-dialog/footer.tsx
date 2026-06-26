"use client";

import { AlertTriangle, CheckCircle2, Loader2, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ValidationIssue } from "./helpers";

interface FooterProps {
  saveError: string | null;
  issues: ValidationIssue[];
  valid: boolean;
  saving: boolean;
  deleting: boolean;
  confirmingDelete: boolean;
  canSave: boolean;
  isEdit: boolean;
  isBuiltin: boolean;
  onCancel: () => void;
  onSave: () => void;
  onDelete: () => void;
}

export function Footer({
  saveError,
  issues,
  valid,
  saving,
  deleting,
  confirmingDelete,
  canSave,
  isEdit,
  isBuiltin,
  onCancel,
  onSave,
  onDelete,
}: FooterProps) {
  return (
    <div className="flex flex-col gap-2 border-t border-border bg-card/40 px-6 py-3">
      {saveError && (
        <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-[11px] text-destructive">
          <div className="flex items-start gap-1.5">
            <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0" />
            <div className="flex-1 leading-snug">{saveError}</div>
          </div>
        </div>
      )}
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
          {valid ? (
            <>
              <CheckCircle2 className="h-3 w-3 text-emerald-400" />
              <span>Ready to save</span>
            </>
          ) : (
            <>
              <AlertTriangle className="h-3 w-3 text-destructive/80" />
              <span>
                {issues.length} {issues.length === 1 ? "issue" : "issues"} to
                resolve
              </span>
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          {isEdit && !isBuiltin && (
            <button
              type="button"
              onClick={onDelete}
              disabled={deleting || saving}
              className={cn(
                "inline-flex h-9 items-center gap-1.5 rounded-md border px-3 text-sm font-medium transition",
                confirmingDelete
                  ? "border-destructive bg-destructive/15 text-destructive hover:bg-destructive/25"
                  : "border-border bg-card text-muted-foreground hover:border-destructive/50 hover:text-destructive",
              )}
            >
              {deleting ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Trash2 className="h-3.5 w-3.5" />
              )}
              {deleting
                ? "Deleting…"
                : confirmingDelete
                  ? "Click again to confirm"
                  : "Delete"}
            </button>
          )}
          <button
            type="button"
            onClick={onCancel}
            className="h-9 rounded-md border border-border bg-card px-4 text-sm font-medium text-muted-foreground transition hover:border-muted-foreground/30 hover:text-foreground"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={!canSave}
            onClick={onSave}
            className={cn(
              "inline-flex h-9 items-center gap-1.5 rounded-md px-4 text-sm font-medium shadow-sm transition",
              canSave
                ? "bg-primary text-primary-foreground hover:bg-primary/90"
                : "cursor-not-allowed bg-muted text-muted-foreground/60 shadow-none",
            )}
          >
            {saving && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            {saving ? "Saving…" : isEdit ? "Save changes" : "Save persona"}
          </button>
        </div>
      </div>
    </div>
  );
}
