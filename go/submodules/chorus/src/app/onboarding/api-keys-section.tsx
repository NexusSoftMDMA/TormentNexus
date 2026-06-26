"use client";

import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  AlertTriangle,
  Check,
  ExternalLink,
  Loader2,
} from "lucide-react";
import {
  addOpenRouterVoices,
  listOpenRouterModels,
  saveOpenRouterKey,
  type OpenRouterModel,
} from "@/lib/api/openrouter";
import { cn } from "@/lib/utils";

/**
 * API-key entry during onboarding. v0.7 ships only the OpenRouter slot —
 * other providers' direct API paths aren't wired yet, so showing their
 * inputs would persist a key that does nothing.
 *
 * Flow when the user saves an OpenRouter key:
 *   1. POST /openrouter/save-key — daemon validates against the
 *      catalog endpoint, returns { valid, error? }.
 *   2. On valid → fetch the catalog (GET /openrouter/models) and reveal
 *      a small picker so the user can add a couple of voices before
 *      finishing onboarding. Saves a second-trip to /connect just to
 *      enable a model.
 *   3. On invalid → inline error, key is NOT persisted.
 */

// Curated starter set — surfaced first in the picker so a user who
// just wants "any 2 voices to try chorus" doesn't have to scroll.
// Prefixed `openrouter:` form is what voices.id stores; the picker
// reconciles against the bare model_id.
const STARTER_PICKS = [
  "anthropic/claude-sonnet-4.5",
  "openai/gpt-5",
  "google/gemini-2.5-pro",
  "deepseek/deepseek-chat-v3.1",
  "moonshotai/kimi-k2",
];

type SaveState =
  | { kind: "idle" }
  | { kind: "saving" }
  | { kind: "invalid"; message: string }
  // Key validated + persisted, but the catalog fetch failed. User can
  // retry just the catalog without re-saving the key (which would be a
  // wasted round-trip).
  | { kind: "catalog_failed"; message: string }
  | { kind: "loading_catalog" }
  | { kind: "valid"; modelCount: number };

interface ApiKeysSectionProps {
  apiKeys: Record<string, string>;
  updateApiKey: (provider: string, value: string) => void;
}

export function ApiKeysSection({ apiKeys, updateApiKey }: ApiKeysSectionProps) {
  const apiKey = apiKeys["openrouter"] ?? "";
  const [saveState, setSaveState] = useState<SaveState>({ kind: "idle" });
  const [catalog, setCatalog] = useState<OpenRouterModel[] | null>(null);
  const [picked, setPicked] = useState<Set<string>>(new Set(STARTER_PICKS));
  const [adding, setAdding] = useState(false);
  const [addedMsg, setAddedMsg] = useState<string | null>(null);

  async function onSaveKey() {
    if (apiKey.trim().length === 0 || saveState.kind === "saving") return;
    setSaveState({ kind: "saving" });
    try {
      const v = await saveOpenRouterKey(apiKey.trim());
      if (!v.valid) {
        setSaveState({
          kind: "invalid",
          message: v.error ?? "OpenRouter rejected this key.",
        });
        return;
      }
      // Key is now persisted server-side. From this point a catalog
      // failure transitions to `catalog_failed` (recoverable via the
      // Reload button), not back to `invalid` (which would imply the
      // key itself was bad).
      await loadCatalog();
    } catch (err) {
      setSaveState({
        kind: "invalid",
        message: err instanceof Error ? err.message : String(err),
      });
    }
  }

  async function loadCatalog() {
    setSaveState({ kind: "loading_catalog" });
    try {
      const { models } = await listOpenRouterModels();
      models.sort((a, b) => a.id.localeCompare(b.id));
      setCatalog(models);
      setSaveState({ kind: "valid", modelCount: models.length });
    } catch (err) {
      setSaveState({
        kind: "catalog_failed",
        message: err instanceof Error ? err.message : String(err),
      });
    }
  }

  async function onAddVoices() {
    if (picked.size === 0 || adding) return;
    setAdding(true);
    setAddedMsg(null);
    try {
      const r = await addOpenRouterVoices(Array.from(picked));
      const added = r.added.length;
      const skipped = r.skipped.length;
      setAddedMsg(
        `Added ${added} voice${added === 1 ? "" : "s"}` +
          (skipped > 0 ? ` · skipped ${skipped} (unknown id)` : ""),
      );
    } catch (err) {
      setAddedMsg(
        `Couldn't add voices: ${err instanceof Error ? err.message : String(err)}`,
      );
    } finally {
      setAdding(false);
    }
  }

  return (
    <section className="mb-8">
      <h2 className="mb-3 flex flex-wrap items-baseline gap-x-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        <span>API access</span>
        <span className="font-normal normal-case text-muted-foreground/70">
          optional — gives you any model on OpenRouter as a chorus voice
        </span>
      </h2>
      <Card className="space-y-3 bg-card p-4">
        <div className="flex items-center justify-between">
          <label
            htmlFor="apikey-openrouter"
            className="text-sm font-medium"
          >
            OpenRouter API key
          </label>
          <a
            href="https://openrouter.ai/keys"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
          >
            Get a key <ExternalLink className="h-2.5 w-2.5" />
          </a>
        </div>

        <div className="flex items-center gap-2">
          <Input
            id="apikey-openrouter"
            type="password"
            autoComplete="off"
            spellCheck={false}
            placeholder="sk-or-v1-…"
            value={apiKey}
            onChange={(e) => {
              updateApiKey("openrouter", e.target.value);
              if (saveState.kind !== "idle") setSaveState({ kind: "idle" });
              setCatalog(null);
            }}
            className="flex-1 font-mono text-xs"
            disabled={
              saveState.kind === "saving" || saveState.kind === "loading_catalog"
            }
          />
          <button
            type="button"
            onClick={onSaveKey}
            disabled={
              apiKey.trim().length === 0 ||
              saveState.kind === "saving" ||
              saveState.kind === "loading_catalog" ||
              saveState.kind === "valid"
            }
            className={cn(
              "inline-flex h-9 items-center gap-1.5 rounded-md border px-3 text-[11px] font-medium transition disabled:cursor-not-allowed disabled:opacity-50",
              saveState.kind === "valid"
                ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-300"
                : "border-primary/40 bg-primary/10 text-primary hover:bg-primary/20",
            )}
          >
            {(saveState.kind === "saving" ||
              saveState.kind === "loading_catalog") && (
              <Loader2 className="h-3 w-3 animate-spin" />
            )}
            {saveState.kind === "valid" && <Check className="h-3 w-3" />}
            {saveState.kind === "saving"
              ? "Validating…"
              : saveState.kind === "loading_catalog"
                ? "Loading models…"
                : saveState.kind === "valid"
                  ? "Validated"
                  : "Validate & save"}
          </button>
        </div>

        {saveState.kind === "invalid" && (
          <p className="flex items-center gap-1.5 text-xs text-destructive">
            <AlertTriangle className="h-3 w-3" />
            {saveState.message}
          </p>
        )}

        {saveState.kind === "catalog_failed" && (
          <div className="flex items-start justify-between gap-3 rounded-md border border-amber-500/40 bg-amber-500/5 px-3 py-2">
            <p className="flex items-start gap-1.5 text-xs text-amber-100">
              <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0" />
              <span>
                Key saved, but couldn&apos;t load the model catalog:{" "}
                {saveState.message}
              </span>
            </p>
            <button
              type="button"
              onClick={loadCatalog}
              className="shrink-0 rounded-md border border-amber-500/40 bg-amber-500/10 px-2.5 py-1 text-[11px] font-medium text-amber-100 hover:bg-amber-500/20"
            >
              Reload models
            </button>
          </div>
        )}

        {saveState.kind === "valid" && catalog && (
          <div className="space-y-2 rounded-md border border-emerald-500/20 bg-emerald-500/5 p-3">
            <p className="text-[11px] text-emerald-300">
              ✓ Key valid · {saveState.modelCount} models available. Pick a
              starter set now or browse the full catalog on the Connect page
              later.
            </p>
            <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
              {STARTER_PICKS.filter((id) => catalog.some((m) => m.id === id)).map(
                (id) => {
                  const sel = picked.has(id);
                  const m = catalog.find((c) => c.id === id);
                  return (
                    <label
                      key={id}
                      className={cn(
                        "flex cursor-pointer items-center gap-2 rounded border px-2 py-1.5 text-[11px] transition",
                        sel
                          ? "border-primary/50 bg-primary/10 text-foreground"
                          : "border-border bg-card text-muted-foreground hover:border-muted-foreground/30",
                      )}
                    >
                      <input
                        type="checkbox"
                        checked={sel}
                        onChange={(e) => {
                          setPicked((prev) => {
                            const next = new Set(prev);
                            if (e.target.checked) next.add(id);
                            else next.delete(id);
                            return next;
                          });
                        }}
                        className="h-3 w-3 shrink-0"
                      />
                      <span className="truncate font-mono">{m?.id ?? id}</span>
                    </label>
                  );
                },
              )}
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={onAddVoices}
                disabled={picked.size === 0 || adding}
                className="inline-flex h-7 items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-2.5 text-[11px] font-medium text-primary transition hover:bg-primary/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {adding ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                Add {picked.size > 0 ? picked.size : ""} voice
                {picked.size === 1 ? "" : "s"}
              </button>
              {addedMsg && (
                <p className="text-[11px] text-emerald-300">{addedMsg}</p>
              )}
            </div>
          </div>
        )}

        <p className="text-[11px] leading-relaxed text-muted-foreground/80">
          Stored locally in <code>~/.chorus/chorus.db</code>. Used only to
          call OpenRouter on your behalf. Direct API support for other
          providers (Anthropic, OpenAI, Google, xAI) is on the roadmap —
          today, route those models via OpenRouter.
        </p>
      </Card>
    </section>
  );
}
