/**
 * OpenRouter inline integration — validate API key, list available models,
 * add models as voices.
 *
 * Why an inline flow: the v0.7 voices table can already store
 * `source='api'` voices, but until now there was no UX path to populate
 * them — users could only enable CLI-detected voices. OpenRouter is the
 * widest practical API gateway (Anthropic, OpenAI, Google, Meta, Kimi,
 * DeepSeek, Mistral, Grok, …) so a single key unlocks dozens of models
 * across every lineage we score for diversity.
 *
 * Architecture lifted from 9router's `/api/providers/validate` route
 * (which we studied before writing this) — same Bearer-auth contract,
 * same `GET /api/v1/models` validation. Differences:
 *   - We persist the key in chorus's `secrets` table, not a JSON blob.
 *   - We classify each model into a chorus lineage via the existing
 *     `classifyOpencodeModel` helper (the "anthropic/claude-..." prefix
 *     pattern matches OpenRouter's id format directly).
 *   - We do NOT proxy chat completions through this module yet — that
 *     comes in a follow-up HTTP-shim PR. This PR ships the validate +
 *     model-catalog + voices-insert flow only; selecting an OpenRouter
 *     voice in a template won't dispatch successfully until the shim
 *     lands. The voices appear in the picker as preview-only.
 */

import { secrets, voices } from '../lib/db/index.js';
import { recordHealth } from '../lib/cli-health.js';
import { classifyOpencodeModel } from '../lib/voices.js';

const OPENROUTER_BASE = 'https://openrouter.ai/api/v1';
const VALIDATE_TIMEOUT_MS = 8_000;
const MODELS_TIMEOUT_MS = 15_000;

export interface OpenRouterModel {
  id: string;
  name: string;
  contextLength?: number;
  /** USD per 1M input tokens. Some models report cost per request — those become null. */
  inputCostPerMtok?: number;
  outputCostPerMtok?: number;
}

export interface ValidateResult {
  valid: boolean;
  /** Surfaced to the cockpit alert when invalid. */
  error?: string;
}

/**
 * Live-probe an API key by hitting `/api/v1/models`. OpenRouter rejects
 * bad keys with HTTP 401, so a 200 means the key is real. Anything else
 * (network errors, 5xx, 403) we surface verbatim — the cockpit then
 * shows the user what to fix.
 */
export async function validateKey(apiKey: string): Promise<ValidateResult> {
  if (!apiKey || apiKey.trim().length === 0) {
    return { valid: false, error: 'API key is empty' };
  }
  try {
    // Hit /auth/key — the only catalog-side endpoint that actually
    // requires the bearer to be valid. /models is public and returns
    // 200 for ANY auth header (including junk), so the previous
    // implementation rubber-stamped invalid keys as valid. Verified
    // behaviour: 401 on invalid bearer, 200 with key metadata on valid.
    const res = await fetch(`${OPENROUTER_BASE}/auth/key`, {
      headers: { Authorization: `Bearer ${apiKey}` },
      signal: AbortSignal.timeout(VALIDATE_TIMEOUT_MS),
    });
    if (res.ok) return { valid: true };
    if (res.status === 401 || res.status === 403) {
      return { valid: false, error: 'Invalid API key' };
    }
    return { valid: false, error: `OpenRouter returned ${res.status}` };
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    return { valid: false, error: `Network error: ${message}` };
  }
}

/**
 * Fetch the full model catalog. Uses the stored OpenRouter key from
 * the secrets table — callers must save the key first via
 * `saveKey` (or pass an explicit key for one-off lookups).
 *
 * We trim the response to the four fields the cockpit picker actually
 * needs, dropping the long descriptions, capability flags, and per-modality
 * pricing. Keeps the wire payload manageable (200+ models * ~2 KB each
 * crosses 400 KB unfiltered).
 *
 * Defensive pagination: OpenRouter currently returns the full catalog
 * in a single response with no cursor. If they ever add pagination
 * (`next_cursor` or `cursor` field), this loop follows it up to
 * MAX_PAGES guard. Today's behaviour is identical to a single fetch.
 */
const MAX_PAGES = 20;

interface OpenRouterModelsBody {
  data?: Array<{
    id: string;
    name?: string;
    context_length?: number;
    pricing?: { prompt?: string; completion?: string };
  }>;
  next_cursor?: string;
  cursor?: string;
}

function mapModel(m: NonNullable<OpenRouterModelsBody['data']>[number]): OpenRouterModel {
  // OpenRouter pricing is reported as USD-per-token (e.g. "0.000003"
  // for $3/Mtok). Convert to per-Mtok USD to match chorus's voices
  // table convention. Strings parse to NaN if missing — guard that.
  const promptCost = m.pricing?.prompt ? parseFloat(m.pricing.prompt) : NaN;
  const completionCost = m.pricing?.completion ? parseFloat(m.pricing.completion) : NaN;
  const result: OpenRouterModel = { id: m.id, name: m.name ?? m.id };
  if (typeof m.context_length === 'number') {
    result.contextLength = m.context_length;
  }
  if (Number.isFinite(promptCost)) {
    result.inputCostPerMtok = promptCost * 1_000_000;
  }
  if (Number.isFinite(completionCost)) {
    result.outputCostPerMtok = completionCost * 1_000_000;
  }
  return result;
}

export async function listModels(apiKey?: string): Promise<OpenRouterModel[]> {
  const key = apiKey ?? (await secrets.get('openrouter'))?.value;
  if (!key) {
    throw new Error(
      'No OpenRouter API key saved. POST /openrouter/save-key with {apiKey} first.',
    );
  }

  const all: OpenRouterModel[] = [];
  let cursor: string | undefined;
  let pages = 0;
  do {
    const url = cursor
      ? `${OPENROUTER_BASE}/models?cursor=${encodeURIComponent(cursor)}`
      : `${OPENROUTER_BASE}/models`;
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${key}` },
      signal: AbortSignal.timeout(MODELS_TIMEOUT_MS),
    });
    if (!res.ok) {
      throw new Error(`OpenRouter /models returned ${res.status}`);
    }
    const body = (await res.json()) as OpenRouterModelsBody;
    if (Array.isArray(body.data)) {
      all.push(...body.data.map(mapModel));
    }
    // Future-compat: OpenRouter may add `next_cursor` (most common naming)
    // or `cursor`. We stop when neither is a non-empty string. Safety cap
    // prevents an infinite loop if a buggy server echoes the same cursor.
    const next = body.next_cursor ?? body.cursor;
    cursor = typeof next === 'string' && next.length > 0 ? next : undefined;
    pages += 1;
  } while (cursor && pages < MAX_PAGES);

  return all;
}

/**
 * Validate, then persist the key under provider='openrouter' in secrets.
 * Returns the validation result; on failure the key is NOT saved.
 *
 * On successful save, immediately flip cli-health to healthy. Without
 * this the home card stays "Auth broken — User not found" until the
 * next successful chat call records health on its own — confusing UX
 * because validateKey already confirmed the new key works against
 * OpenRouter's /models endpoint just now.
 */
export async function saveKey(apiKey: string): Promise<ValidateResult> {
  const result = await validateKey(apiKey);
  if (!result.valid) return result;
  await secrets.set('openrouter', 'api_key', apiKey);
  // Best-effort: a recordHealth failure shouldn't block the save.
  recordHealth({ lineage: 'openrouter', status: 'healthy' }).catch((err: unknown) => {
    console.error('[chorus] recordHealth failed for openrouter on saveKey:', err);
  });
  return result;
}

/**
 * Add the chosen models as voices. Idempotent — voices.upsert dedupes
 * by id. Returns the list of voice ids actually inserted/updated.
 *
 * Voice id format mirrors the kimi/opencode-go convention:
 *   `openrouter:<openrouter-model-id>`
 * so the picker can show "openrouter:anthropic/claude-3.5-sonnet" and
 * a future template can reference it unambiguously.
 */
export async function addModelsAsVoices(
  modelIds: string[],
  apiKey?: string,
): Promise<{ added: string[]; skipped: string[] }> {
  if (modelIds.length === 0) return { added: [], skipped: [] };
  // Pull the catalog ONCE so we can fill labels + pricing for each id
  // without re-hitting OpenRouter per-model. Optional `apiKey` plumbs
  // through to listModels so callers that already hold the key don't
  // race a concurrent save-key write to the secrets table.
  const catalog = await listModels(apiKey);
  const byId = new Map(catalog.map((m) => [m.id, m]));

  const added: string[] = [];
  const skipped: string[] = [];
  for (const modelId of modelIds) {
    const meta = byId.get(modelId);
    if (!meta) {
      skipped.push(modelId);
      continue;
    }
    const { lineage, vendor_family } = classifyOpencodeModel(modelId);
    const voiceId = `openrouter:${modelId}`;
    // Enabled by default — the HTTP dispatch shim now exists at
    // src/daemon/agents/openrouter.ts, so a template referencing this
    // voice will dispatch successfully via /api/v1/chat/completions.
    await voices.upsert({
      id: voiceId,
      label: meta.name,
      source: 'api',
      provider: 'openrouter',
      model_id: modelId,
      lineage,
      vendor_family: vendor_family ?? null,
      input_cost_per_mtok: meta.inputCostPerMtok ?? null,
      output_cost_per_mtok: meta.outputCostPerMtok ?? null,
      enabled: true,
    });
    added.push(voiceId);
  }
  return { added, skipped };
}
