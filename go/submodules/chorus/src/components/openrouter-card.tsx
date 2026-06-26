"use client";

import { useState } from "react";
import {
  Check,
  Loader2,
  AlertTriangle,
  Search,
  ExternalLink,
} from "lucide-react";
import { cn } from "@/lib/utils";
import {
  saveOpenRouterKey,
  listOpenRouterModels,
  addOpenRouterVoices,
  type OpenRouterModel,
} from "@/lib/api/openrouter";
import { listVoices, updateVoice, type Voice } from "@/lib/api/voices";
import { UI_LINEAGE_BRAND } from "@/lib/lineage-maps";

/**
 * OpenRouter "connection" card on /connect. Header is always visible;
 * body shows inline. Two states:
 *   - No voices yet: header + key input + Save button.
 *   - Already-added voices: header + toggleable voice grid; "Add more
 *     models" button lazy-loads the catalog (300+ models — too heavy
 *     to fetch unconditionally on every page load).
 */

interface Props {
  /** Existing openrouter voices (provider=openrouter, source=api). */
  voices: Voice[];
}

export function OpenRouterCard({ voices: initialVoices }: Props) {
  const [voices, setVoices] = useState<Voice[]>(initialVoices);
  const [catalog, setCatalog] = useState<OpenRouterModel[] | null>(null);
  const [loadingCatalog, setLoadingCatalog] = useState(false);
  const [showCatalog, setShowCatalog] = useState(false);

  const [apiKey, setApiKey] = useState("");
  const [keyBusy, setKeyBusy] = useState(false);
  const [keyError, setKeyError] = useState<string | null>(null);
  const [showKeyInput, setShowKeyInput] = useState(false);

  const [filter, setFilter] = useState("");
  const [picked, setPicked] = useState<Set<string>>(new Set());
  const [adding, setAdding] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);
  const [addedMsg, setAddedMsg] = useState<string | null>(null);

  const [savingId, setSavingId] = useState<string | null>(null);
  const [toggleError, setToggleError] = useState<string | null>(null);

  const enabledCount = voices.filter((v) => v.enabled).length;
  const hasVoices = voices.length > 0;

  async function loadCatalog() {
    if (loadingCatalog) return;
    setLoadingCatalog(true);
    setKeyError(null);
    setShowCatalog(true);
    try {
      const { models } = await listOpenRouterModels();
      models.sort((a, b) => a.id.localeCompare(b.id));
      setCatalog(models);
      setShowKeyInput(false);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setShowKeyInput(true);
      if (!/no openrouter api key/i.test(message)) {
        setKeyError(message);
      }
    } finally {
      setLoadingCatalog(false);
    }
  }

  async function saveKey() {
    if (apiKey.trim().length === 0) return;
    setKeyBusy(true);
    setKeyError(null);
    try {
      const v = await saveOpenRouterKey(apiKey.trim());
      if (!v.valid) {
        setKeyError(v.error ?? "Validation failed");
        return;
      }
      setApiKey("");
      setShowKeyInput(false);
      const { models } = await listOpenRouterModels();
      models.sort((a, b) => a.id.localeCompare(b.id));
      setCatalog(models);
      setShowCatalog(true);
    } catch (err) {
      setKeyError(err instanceof Error ? err.message : String(err));
    } finally {
      setKeyBusy(false);
    }
  }

  async function addPicked() {
    if (picked.size === 0) return;
    setAdding(true);
    setAddError(null);
    setAddedMsg(null);
    try {
      const r = await addOpenRouterVoices(Array.from(picked));
      const added = r.added.length;
      const skipped = r.skipped.length;
      setAddedMsg(
        `Added ${added} voice${added === 1 ? "" : "s"}` +
          (skipped > 0 ? ` · skipped ${skipped} (unknown id)` : ""),
      );
      setPicked(new Set());
      try {
        const next = await listVoices({ source: "api", provider: "openrouter" });
        setVoices(next);
      } catch {
        /* best-effort reload — voices were inserted regardless */
      }
    } catch (err) {
      setAddError(err instanceof Error ? err.message : String(err));
    } finally {
      setAdding(false);
    }
  }

  async function toggleVoice(v: Voice) {
    setSavingId(v.id);
    setToggleError(null);
    try {
      const next = await updateVoice(v.id, { enabled: !v.enabled });
      setVoices((prev) => prev.map((p) => (p.id === next.id ? next : p)));
    } catch (err) {
      setToggleError(err instanceof Error ? err.message : String(err));
    } finally {
      setSavingId(null);
    }
  }

  const filtered = filter.trim()
    ? (catalog ?? []).filter(
        (m) =>
          m.id.toLowerCase().includes(filter.toLowerCase().trim()) ||
          m.name.toLowerCase().includes(filter.toLowerCase().trim()),
      )
    : catalog ?? [];

  return (
    <div
      className={cn(
        "flex h-72 flex-col rounded-lg border border-border",
        UI_LINEAGE_BRAND.openrouter.gradient,
      )}
    >
      <div className="flex shrink-0 items-center gap-3 px-4 py-3">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          <span
            className={cn(
              "h-2 w-2 shrink-0 rounded-full",
              UI_LINEAGE_BRAND.openrouter.dot,
            )}
          />
          <h3 className="whitespace-nowrap text-sm font-semibold">OpenRouter</h3>
          {hasVoices ? (
            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
              <Check className="h-3 w-3" /> {enabledCount} enabled
            </span>
          ) : (
            <span className="inline-flex items-center rounded-full border border-border bg-card px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
              Not configured
            </span>
          )}
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-3 overflow-y-auto border-t border-border bg-card/30 p-4">
        {!hasVoices && (
          <p className="text-[11px] leading-relaxed text-muted-foreground">
            Voices added here run via OpenRouter&apos;s chat-completions API.
            Costs come back per-call from the API and surface on the run page.
          </p>
        )}

        {hasVoices && (
          <div className="space-y-2">
            <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
              Models
            </p>
            <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
              {voices.map((v) => (
                <button
                  key={v.id}
                  type="button"
                  disabled={savingId === v.id}
                  onClick={() => toggleVoice(v)}
                  title={v.model_id}
                  className={cn(
                    "flex items-center gap-2 rounded border px-2 py-1.5 text-left text-[11px] transition disabled:opacity-60",
                    v.enabled
                      ? "border-primary/50 bg-primary/10 text-foreground"
                      : "border-border bg-card text-muted-foreground hover:border-muted-foreground/30",
                  )}
                >
                  <div
                    className={cn(
                      "grid h-3 w-3 shrink-0 place-items-center rounded-sm border transition",
                      v.enabled
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border",
                    )}
                  >
                    {v.enabled && <Check className="h-2 w-2" />}
                  </div>
                  <span className="truncate font-mono">{v.model_id}</span>
                </button>
              ))}
            </div>
            {toggleError && (
              <p className="text-[11px] text-destructive">{toggleError}</p>
            )}
          </div>
        )}

        {(showKeyInput || !hasVoices) && (
          <div className="space-y-2 rounded-md border border-border bg-background/40 p-3">
            <div className="flex items-center justify-between">
              <label className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                {hasVoices ? "Replace API key" : "OpenRouter API key"}
              </label>
              <a
                href="https://openrouter.ai/keys"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground"
              >
                Get a key <ExternalLink className="h-2.5 w-2.5" />
              </a>
            </div>
            <div className="flex items-center gap-2">
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-or-v1-…"
                disabled={keyBusy}
                className="h-8 flex-1 rounded-md border border-border bg-background px-2 font-mono text-xs focus:border-primary/60 focus:outline-none"
              />
              <button
                type="button"
                onClick={saveKey}
                disabled={keyBusy || apiKey.trim().length === 0}
                className="inline-flex h-8 items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-3 text-[11px] font-medium text-primary transition hover:bg-primary/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {keyBusy ? (
                  <Loader2 className="h-3 w-3 animate-spin" />
                ) : (
                  <Check className="h-3 w-3" />
                )}
                {keyBusy ? "Saving…" : "Save"}
              </button>
            </div>
            {keyError && (
              <p className="flex items-center gap-1 text-[11px] text-destructive">
                <AlertTriangle className="h-3 w-3" />
                {keyError}
              </p>
            )}
          </div>
        )}

        <div className="flex flex-wrap items-center gap-2">
          {hasVoices && !showCatalog && (
            <button
              type="button"
              onClick={loadCatalog}
              disabled={loadingCatalog}
              className="inline-flex h-8 items-center gap-1.5 rounded-md border border-border bg-card px-3 text-[11px] font-medium text-foreground transition hover:border-muted-foreground/30 disabled:opacity-50"
            >
              {loadingCatalog ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : null}
              Add more models
            </button>
          )}
          {hasVoices && !showKeyInput && (
            <button
              type="button"
              onClick={() => setShowKeyInput(true)}
              className="text-[11px] text-muted-foreground underline-offset-2 hover:underline"
            >
              Replace API key
            </button>
          )}
        </div>

        {showCatalog && (
          <>
            {loadingCatalog && (
              <p className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <Loader2 className="h-3 w-3 animate-spin" /> Loading catalog…
              </p>
            )}

            {catalog && catalog.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center justify-between gap-2">
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                    {hasVoices ? "Add more models" : "Pick models"}
                  </p>
                  <button
                    type="button"
                    onClick={addPicked}
                    disabled={picked.size === 0 || adding}
                    className="inline-flex h-7 items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-2.5 text-[11px] font-medium text-primary transition hover:bg-primary/20 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {adding ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                    Add {picked.size > 0 ? picked.size : ""}
                  </button>
                </div>

                <div className="flex items-center gap-2 rounded-md border border-border bg-background/40 px-2">
                  <Search className="h-3 w-3 text-muted-foreground" />
                  <input
                    type="text"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder="filter by id or name…"
                    className="h-7 flex-1 bg-transparent text-[11px] focus:outline-none"
                  />
                  <span className="text-[10px] text-muted-foreground">
                    {filtered.length}/{catalog.length}
                  </span>
                </div>

                {addedMsg && (
                  <p className="text-[11px] text-emerald-300">{addedMsg}</p>
                )}
                {addError && (
                  <p className="flex items-center gap-1 text-[11px] text-destructive">
                    <AlertTriangle className="h-3 w-3" />
                    {addError}
                  </p>
                )}

                <div className="max-h-72 overflow-y-auto rounded-md border border-border">
                  {filtered.map((m) => {
                    const checked = picked.has(m.id);
                    const already = voices.some((v) => v.model_id === m.id);
                    return (
                      <label
                        key={m.id}
                        className={cn(
                          "flex cursor-pointer items-center gap-2 border-b border-border px-2 py-1.5 text-[11px] last:border-b-0 hover:bg-card/40",
                          already && "opacity-50",
                        )}
                      >
                        <input
                          type="checkbox"
                          checked={checked}
                          disabled={already}
                          onChange={(e) => {
                            setPicked((prev) => {
                              const next = new Set(prev);
                              if (e.target.checked) next.add(m.id);
                              else next.delete(m.id);
                              return next;
                            });
                          }}
                          className="h-3 w-3 shrink-0"
                        />
                        <div className="min-w-0 flex-1">
                          <div className="truncate font-mono text-[11px]">
                            {m.id}
                            {already && (
                              <span className="ml-1 text-[10px] text-muted-foreground">
                                · added
                              </span>
                            )}
                          </div>
                          <div className="truncate text-[10px] text-muted-foreground">
                            {m.name}
                            {m.contextLength
                              ? ` · ${(m.contextLength / 1000).toFixed(0)}k ctx`
                              : ""}
                            {m.inputCostPerMtok !== undefined ||
                            m.outputCostPerMtok !== undefined
                              ? ` · ${formatPrice(m.inputCostPerMtok)}/${formatPrice(m.outputCostPerMtok)} per Mtok`
                              : ""}
                          </div>
                        </div>
                      </label>
                    );
                  })}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function formatPrice(usd?: number): string {
  if (usd === undefined) return "?";
  if (usd < 0.01) return `$${usd.toFixed(4)}`;
  return `$${usd.toFixed(2)}`;
}
