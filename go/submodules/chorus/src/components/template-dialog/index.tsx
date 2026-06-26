"use client";

/**
 * TemplateDialog — single dialog for both NEW and EDIT modes.
 *
 * Modes:
 *   new  — opens with a default skeleton; id derives from name on save.
 *   edit — opens populated from a Template; preserves the id. YAML tab
 *          default for review_only / ship templates (form mode would drop
 *          fields the form doesn't model). Otherwise form tab default.
 *
 * Validation runs live (`validateTemplateYaml`); save is gated on
 * `valid: true`. Server re-validates and returns the same shape, so a
 * successful client validate followed by a server reject is a real bug.
 *
 * Save calls `saveTemplate({id, yaml})`. Success → `onSaved(template)`.
 * Failure → keeps the dialog open, shows server error inline.
 */

import { useEffect, useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  CheckCircle2,
  FileCode2,
  Loader2,
  Pencil,
  Plus,
  Sparkles,
  X,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { saveTemplate, DaemonError } from "@/lib/api";
import type { Template } from "@/lib/cockpit-types";
import { selectLiveYaml } from "@/lib/template-live-yaml";
import {
  validateTemplateYaml,
  type TemplateValidationIssue,
} from "@/lib/template-validation";
import { cn } from "@/lib/utils";
import { DEFAULT_FORM } from "./constants";
import { buildYamlFromForm, slugify } from "./emit";
import { FormPanel } from "./form-panel";
import { parseYamlToForm } from "./parse";
import { TabButton } from "./primitives";
import type { FormState } from "./types";
import { YamlPanel } from "./yaml-panel";

export interface TemplateDialogProps {
  /** When set, dialog opens in EDIT mode, populated from this template. */
  editing?: Template;
  /** Trigger element. New mode wraps with "+ New template" by default. */
  trigger?: React.ReactNode;
  /** Called after a successful save. Parent should refresh its list. */
  onSaved?: (savedId: string) => void;
}

export function TemplateDialog({
  editing,
  trigger,
  onSaved,
}: TemplateDialogProps) {
  const [open, setOpen] = useState(false);
  const isEdit = Boolean(editing);

  // Depend on the *content* (id + yaml) rather than object reference so a
  // parent re-render with a new Template object pointing at the same
  // template doesn't wipe in-flight dialog edits.
  const initial = useMemo(() => {
    if (!editing) {
      return {
        form: DEFAULT_FORM,
        yaml: buildYamlFromForm(DEFAULT_FORM),
        lossy: false,
        reasons: [] as string[],
      };
    }
    const parsed = parseYamlToForm(editing.yaml, editing.id);
    return {
      form: parsed.form,
      yaml: editing.yaml,
      lossy: parsed.formLossy,
      reasons: parsed.lossyReasons,
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editing?.id, editing?.yaml]);

  const [form, setForm] = useState<FormState>(initial.form);
  const [yamlText, setYamlText] = useState<string>(initial.yaml);
  const [yamlDirty, setYamlDirty] = useState(false);
  // Without this, opening a builtin template in Form mode immediately
  // re-stringifies via buildYamlFromForm — which strips comments and
  // reformats. The server's byte-equality check then fails on save and
  // the row gets promoted from 'builtin' to 'user', losing all comments
  // + breaking the builtin-resync loop. Save uses the original YAML
  // when neither dirty flag is set.
  const [formDirty, setFormDirty] = useState(false);
  const [tab, setTab] = useState<"form" | "yaml">(
    initial.lossy ? "yaml" : "form",
  );
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [serverIssues, setServerIssues] = useState<TemplateValidationIssue[]>([]);

  // Reset internal form state when the parent passes a different
  // template. Same trade-off as persona-dialog: this component owns
  // its own dirty/tab/error state which would be discarded by a
  // key-based remount mid-edit. Migration to lifted state is v0.8.
  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect */
    setForm(initial.form);
    setYamlText(initial.yaml);
    setYamlDirty(false);
    setFormDirty(false);
    setTab(initial.lossy ? "yaml" : "form");
    setSaveError(null);
    setServerIssues([]);
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [initial]);

  // Hold the latest `initial` in a ref so the setTimeout-based reset on
  // dialog close always reads the freshest value. Without this, save →
  // close → 200ms-later reset captured the OLD initial via closure, then
  // overwrote the new state already set by the [initial] useEffect when
  // the parent re-rendered with the saved template. Assignment moved
  // into useEffect to satisfy the React Compiler refs-during-render
  // rule; the only reader is `reset()` invoked 200ms after dialog close,
  // long after this effect has committed.
  const initialRef = useRef(initial);
  useEffect(() => {
    initialRef.current = initial;
  });

  // Source-of-truth precedence — see selectLiveYaml docstring. The
  // "return original yaml verbatim" step is what prevents builtin
  // templates from being silently promoted to 'user' source via
  // stringify-induced byte drift.
  const liveYaml = selectLiveYaml({
    yamlDirty,
    tab,
    formDirty,
    yamlText,
    formYaml: buildYamlFromForm(form),
    initialYaml: initial.yaml,
  });
  const validation = useMemo(() => validateTemplateYaml(liveYaml), [liveYaml]);

  function setFormField<K extends keyof FormState>(k: K, v: FormState[K]) {
    const next = { ...form, [k]: v };
    setForm(next);
    setFormDirty(true);
    if (!yamlDirty) setYamlText(buildYamlFromForm(next));
  }

  /**
   * Switching from a dirty YAML pane back to Form would drop the YAML
   * edits silently — buildYamlFromForm(form) would overwrite them on the
   * next setFormField. Re-parse the YAML into the form first so the form
   * reflects what the user just typed, then clear yamlDirty.
   *
   * If the YAML can't be parsed back (lossy fields, syntax errors), keep
   * the user on the YAML tab — refusing to switch beats silently nuking
   * their edits.
   */
  function handleTabChange(next: "form" | "yaml") {
    if (next === "form" && yamlDirty) {
      const reparsed = parseYamlToForm(yamlText, form.id);
      if (reparsed.formLossy) {
        return;
      }
      setForm(reparsed.form);
      setYamlDirty(false);
      // Promoting dirty YAML into the form counts as form-mode edits.
      // Without this, switching back to Form would drop the comments the
      // user just edited away from, then liveYaml would fall back to
      // initial.yaml on save — losing the user's intent.
      setFormDirty(true);
    }
    setTab(next);
  }

  function reset() {
    const i = initialRef.current;
    setForm(i.form);
    setYamlText(i.yaml);
    setYamlDirty(false);
    setFormDirty(false);
    setTab(i.lossy ? "yaml" : "form");
    setSaving(false);
    setSaveError(null);
    setServerIssues([]);
  }

  function handleClose(o: boolean) {
    setOpen(o);
    if (!o) setTimeout(reset, 200);
  }

  async function handleSave() {
    if (!validation.valid || saving) return;
    const idToSave = isEdit
      ? editing!.id
      : form.id || slugify(form.name) || "untitled";
    setSaving(true);
    setSaveError(null);
    setServerIssues([]);
    try {
      const saved = await saveTemplate({ id: idToSave, yaml: liveYaml });
      onSaved?.(saved.id);
      handleClose(false);
    } catch (err) {
      if (err instanceof DaemonError) {
        setSaveError(err.message);
        const detailIssues =
          (err.details as { issues?: TemplateValidationIssue[] } | undefined)
            ?.issues ?? [];
        setServerIssues(detailIssues);
      } else {
        setSaveError(
          err instanceof Error ? err.message : "Save failed (unknown error)",
        );
      }
    } finally {
      setSaving(false);
    }
  }

  // validation.valid already enforces non-empty `name` via TemplateSchema.
  // Gating on form.name additionally would block raw-YAML paste in YAML
  // mode (form.name only updates from form-tab edits).
  const canSave = validation.valid && !saving;

  const triggerEl = trigger ?? (
    <button
      type="button"
      aria-label={isEdit ? "Edit template" : "New template"}
      className="inline-flex shrink-0 items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground shadow-sm transition hover:bg-primary/90"
    >
      {isEdit ? <Pencil className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
      <span className="hidden sm:inline">
        {isEdit ? "Edit" : "New template"}
      </span>
    </button>
  );

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogTrigger asChild>{triggerEl}</DialogTrigger>

      <DialogContent
        showCloseButton={false}
        className="flex max-h-[85vh] w-[min(96vw,720px)] flex-col gap-0 overflow-hidden border-border bg-card p-0 shadow-2xl sm:max-w-none"
      >
        <div className="flex items-center gap-3 border-b border-border px-6 py-4">
          <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-primary/15 text-primary ring-1 ring-primary/20">
            <FileCode2 className="h-4 w-4" />
          </div>
          <div className="flex-1">
            <DialogTitle className="text-[15px] font-semibold leading-tight tracking-tight">
              {isEdit ? `Edit template — ${editing!.name}` : "New template"}
            </DialogTitle>
            <DialogDescription className="mt-0.5 text-xs text-muted-foreground">
              {isEdit
                ? "Edit the YAML directly or use the form. Validation runs live."
                : "A reusable review workflow. Edit visually below or paste raw YAML."}
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

        <div className="border-b border-border bg-card/40 px-6">
          <div className="-mb-px flex items-center gap-1">
            <TabButton
              active={tab === "form"}
              onClick={() => !initial.lossy && handleTabChange("form")}
              disabled={initial.lossy}
            >
              <Sparkles className="h-3.5 w-3.5" />
              Form
              {initial.lossy && (
                <span className="ml-1 rounded bg-muted px-1 text-[9px] font-medium text-muted-foreground">
                  YAML only
                </span>
              )}
            </TabButton>
            <TabButton
              active={tab === "yaml"}
              onClick={() => handleTabChange("yaml")}
            >
              <FileCode2 className="h-3.5 w-3.5" />
              YAML
              {!validation.valid && tab === "form" && (
                <span className="ml-1 h-1.5 w-1.5 rounded-full bg-destructive" />
              )}
            </TabButton>
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {initial.lossy && tab === "yaml" && (
            <div className="border-b border-amber-500/30 bg-amber-500/5 px-6 py-2.5 text-[11px] text-amber-200">
              <div className="flex items-start gap-2">
                <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                <div className="flex-1">
                  <div className="font-medium">
                    Form editing disabled for this template
                  </div>
                  <ul className="mt-1 space-y-0.5 leading-snug">
                    {initial.reasons.map((r, i) => (
                      <li key={i}>· {r}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          )}

          {tab === "form" ? (
            <FormPanel form={form} setField={setFormField} />
          ) : (
            <YamlPanel
              yaml={yamlText}
              filename={`${form.id || slugify(form.name) || "untitled"}.yaml`}
              issues={validation.issues}
              onChange={(v) => {
                setYamlText(v);
                setYamlDirty(true);
              }}
            />
          )}
        </div>

        <div className="flex flex-col gap-2 border-t border-border bg-card/40 px-6 py-3">
          {saveError && (
            <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-[11px] text-destructive">
              <div className="flex items-start gap-1.5">
                <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0" />
                <div className="flex-1">
                  <div className="font-medium">Save failed</div>
                  <div className="mt-0.5 leading-snug">{saveError}</div>
                  {serverIssues.length > 0 && (
                    <ul className="mt-1 space-y-0.5">
                      {serverIssues.slice(0, 5).map((i, idx) => (
                        <li key={idx}>
                          <span className="font-mono">{i.path}</span>:{" "}
                          {i.message}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </div>
          )}
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              {validation.valid ? (
                <>
                  <CheckCircle2 className="h-3 w-3 text-emerald-400" />
                  <span>Ready to save</span>
                </>
              ) : (
                <>
                  <AlertTriangle className="h-3 w-3 text-destructive/80" />
                  <span>
                    {validation.issues.length}{" "}
                    {validation.issues.length === 1 ? "issue" : "issues"} to
                    resolve
                  </span>
                </>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => handleClose(false)}
                className="h-9 rounded-md border border-border bg-card px-4 text-sm font-medium text-muted-foreground transition hover:border-muted-foreground/30 hover:text-foreground"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={!canSave}
                onClick={handleSave}
                className={cn(
                  "inline-flex h-9 items-center gap-1.5 rounded-md px-4 text-sm font-medium shadow-sm transition",
                  canSave
                    ? "bg-primary text-primary-foreground hover:bg-primary/90"
                    : "cursor-not-allowed bg-muted text-muted-foreground/60 shadow-none",
                )}
              >
                {saving && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                {saving
                  ? "Saving…"
                  : isEdit
                    ? "Save changes"
                    : "Save template"}
              </button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

