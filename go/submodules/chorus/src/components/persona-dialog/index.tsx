"use client";

/**
 * PersonaDialog — single dialog for both NEW and EDIT.
 *
 * Personas are flat (no YAML/Form duality like templates), so this is a
 * single form panel. The system prompt is the bulk of the content — a
 * Markdown body that gets prepended to the doer/reviewer's ask at run
 * time, so the textarea stretches to fill the dialog.
 *
 * Built-in source semantics mirror templates:
 *   - Save on a builtin row → server demotes it to user-owned so the
 *     boot-time seed leaves it alone.
 *   - Delete is hidden for builtins because the seed would recreate them.
 *     Users can clone-and-rename instead via the "Fork" trigger.
 */

import { Pencil, Plus, Sparkles, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { DaemonError, deletePersona, savePersona } from "@/lib/api";
import type { Persona } from "@/lib/api/personas";
import { FormBody } from "./form-body";
import { Footer } from "./footer";
import { DEFAULT_FORM, slugify, validate, type FormState } from "./helpers";

export interface PersonaDialogProps {
  /** When set, dialog opens in EDIT mode populated from this persona.
   *  Built-ins are edited in place — the server demotes them to
   *  user-owned on save and the boot seed skips user rows. The
   *  Duplicate per-row affordance covers "I want a renamed copy". */
  editing?: Persona;
  /** Renders the trigger. New mode wraps with a "+ New persona" button by default. */
  trigger?: React.ReactNode;
  /** Called after a successful save. Parent should refresh its list. */
  onSaved?: (savedId: string) => void;
  /** Called after a successful delete. Parent should refresh its list. */
  onDeleted?: (deletedId: string) => void;
  /** When provided, opens the dialog automatically on mount. Used by the
   *  page after a Duplicate so the user lands directly in the new row's
   *  edit form. The parent should clear this once consumed (typically
   *  via remounting the dialog with a different React key). */
  defaultOpen?: boolean;
}

export function PersonaDialog({
  editing,
  trigger,
  onSaved,
  onDeleted,
  defaultOpen,
}: PersonaDialogProps) {
  const [open, setOpen] = useState(Boolean(defaultOpen));
  const isEdit = Boolean(editing);
  const isBuiltin = Boolean(editing?.builtin);

  // Depend on the *content* (id + updated_at) rather than the object
  // reference so a parent re-render with a fresh Persona pointing at the
  // same row doesn't wipe in-flight edits.
  const initial = useMemo<FormState>(() => {
    if (!editing) return DEFAULT_FORM;
    return {
      id: editing.id,
      label: editing.label,
      one_liner: editing.one_liner,
      recommended_lineage: editing.recommended_lineage ?? "",
      system_prompt: editing.system_prompt ?? "",
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editing?.id, editing?.updated_at]);

  const [form, setForm] = useState<FormState>(initial);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  // Tracks whether the user manually edited the id field; once they do,
  // typing in the label stops auto-syncing the slug.
  const [idDirty, setIdDirty] = useState(isEdit);

  // Reset internal form state when the parent passes a different
  // persona (id+updated_at change). The "right" pattern per React
  // Compiler is to remount via parent's `key` prop, but this dialog
  // also owns its own open/close state — a key change mid-edit would
  // close the dialog and discard pending edits. Keeping the
  // controlled-reset pattern; migration to lifted open state is v0.8.
  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect */
    setForm(initial);
    setSaveError(null);
    setConfirmingDelete(false);
    setIdDirty(isEdit);
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [initial, isEdit]);

  // Hold the latest `initial` in a ref so the setTimeout-based reset
  // (`reset()` 200ms after dialog close) always reads the freshest value
  // — without it, save → close → 200ms-later reset captured the OLD
  // initial via closure. Assignment moved into useEffect so the React
  // Compiler purity check passes; the timing change is safe because the
  // only reader is `reset()`, which runs from setTimeout 200ms after
  // the close, well after this effect has committed.
  const initialRef = useRef(initial);
  useEffect(() => {
    initialRef.current = initial;
  });

  const issues = useMemo(() => validate(form, isEdit), [form, isEdit]);
  const valid = issues.length === 0;

  function setField<K extends keyof FormState>(k: K, v: FormState[K]) {
    setForm((prev) => {
      const next = { ...prev, [k]: v };
      if (k === "label" && !idDirty && !isEdit) {
        next.id = slugify(String(v));
      }
      return next;
    });
  }

  function reset() {
    setForm(initialRef.current);
    setSaving(false);
    setDeleting(false);
    setSaveError(null);
    setConfirmingDelete(false);
    setIdDirty(isEdit);
  }

  function handleClose(o: boolean) {
    setOpen(o);
    if (!o) setTimeout(reset, 200);
  }

  async function handleSave() {
    if (!valid || saving) return;
    const idToSave = isEdit ? editing!.id : form.id || slugify(form.label);
    setSaving(true);
    setSaveError(null);
    try {
      const saved = await savePersona({
        id: idToSave,
        label: form.label.trim(),
        one_liner: form.one_liner.trim(),
        system_prompt: form.system_prompt,
        recommended_lineage: form.recommended_lineage || null,
      });
      onSaved?.(saved.id);
      handleClose(false);
    } catch (err) {
      setSaveError(
        err instanceof DaemonError
          ? err.message
          : err instanceof Error
            ? err.message
            : "Save failed (unknown error)",
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!editing || isBuiltin || deleting) return;
    if (!confirmingDelete) {
      setConfirmingDelete(true);
      // Auto-disarm after 8s so an abandoned arming doesn't fire on a
      // much later stray click. 4s was too tight in real-user testing —
      // anyone double-checking the row label blew the window.
      setTimeout(() => setConfirmingDelete((cur) => (cur ? false : cur)), 8000);
      return;
    }
    setDeleting(true);
    setSaveError(null);
    try {
      await deletePersona(editing.id);
      onDeleted?.(editing.id);
      handleClose(false);
    } catch (err) {
      setSaveError(
        err instanceof DaemonError
          ? err.message
          : err instanceof Error
            ? err.message
            : "Delete failed (unknown error)",
      );
      setConfirmingDelete(false);
    } finally {
      setDeleting(false);
    }
  }

  const canSave = valid && !saving && !deleting;

  const triggerEl =
    trigger ??
    (isEdit ? (
      <button
        type="button"
        aria-label={`Edit persona ${editing!.label}`}
        className="grid h-7 w-7 place-items-center rounded-md border border-border bg-card text-muted-foreground transition hover:border-muted-foreground/40 hover:bg-accent/60 hover:text-foreground"
      >
        <Pencil className="h-3.5 w-3.5" />
      </button>
    ) : (
      <button
        type="button"
        aria-label="New persona"
        className="inline-flex shrink-0 items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground shadow-sm transition hover:bg-primary/90"
      >
        <Plus className="h-4 w-4" />
        <span className="hidden sm:inline">New persona</span>
      </button>
    ));

  const titleText = isEdit ? `Edit persona — ${editing!.label}` : "New persona";
  const descText = isEdit
    ? isBuiltin
      ? "Saving promotes this built-in to a user-owned row at the same id so your edits survive daemon restarts."
      : "Edit the worldview, recommended lineage, and system prompt."
    : "Personas are reusable worldviews that prepend to a doer or reviewer's prompt.";

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogTrigger asChild>{triggerEl}</DialogTrigger>

      <DialogContent
        showCloseButton={false}
        className="flex max-h-[85vh] w-[min(96vw,720px)] flex-col gap-0 overflow-hidden border-border bg-card p-0 shadow-2xl sm:max-w-none"
      >
        <div className="flex items-center gap-3 border-b border-border px-6 py-4">
          <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-primary/15 text-primary ring-1 ring-primary/20">
            <Sparkles className="h-4 w-4" />
          </div>
          <div className="flex-1 min-w-0">
            <DialogTitle className="text-[15px] font-semibold leading-tight tracking-tight">
              {titleText}
            </DialogTitle>
            <DialogDescription className="mt-0.5 text-xs text-muted-foreground">
              {descText}
            </DialogDescription>
          </div>
          <button
            type="button"
            onClick={() => handleClose(false)}
            className="grid h-7 w-7 place-items-center rounded-md text-muted-foreground transition hover:bg-accent hover:text-foreground"
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <FormBody
          form={form}
          setField={setField}
          setIdDirty={setIdDirty}
          isEdit={isEdit}
          isBuiltin={isBuiltin}
          issues={issues}
        />

        <Footer
          saveError={saveError}
          issues={issues}
          valid={valid}
          saving={saving}
          deleting={deleting}
          confirmingDelete={confirmingDelete}
          canSave={canSave}
          isEdit={isEdit}
          isBuiltin={isBuiltin}
          onCancel={() => handleClose(false)}
          onSave={handleSave}
          onDelete={handleDelete}
        />
      </DialogContent>
    </Dialog>
  );
}
