/**
 * Model-pricing lookup. Backs the home page's "plan equivalent" spend
 * column — for CLI shims that don't self-report cost (gemini-cli emits
 * tokens only; codex `exec` emits nothing), we synthesize cost from
 * tokens × $/token using OpenRouter's public model catalog.
 *
 * OpenRouter exposes prices in $ per token (not per Mtok) at
 * `GET https://openrouter.ai/api/v1/models` — no auth required for the
 * catalog. We fetch once at first need, cache in memory, persist a 24h
 * disk copy at `~/.chorus/model-pricing.json` so a daemon restart reuses
 * the snapshot instead of refetching cold.
 *
 * Model-id matching is dash/dot-insensitive: chorus uses `claude-opus-4-7`
 * while OpenRouter uses `anthropic/claude-opus-4.7`. We normalize both
 * sides to "lowercased, dots → dashes, vendor prefix stripped" before
 * the lookup so the same logical model resolves regardless of cosmetic
 * id drift.
 */
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

interface ModelPrice {
  /** $/token (NOT per Mtok). Multiply by token count to get USD spend. */
  inputCostPerToken: number;
  /** $/token for output (a.k.a. completion). */
  outputCostPerToken: number;
}

const CACHE_PATH = path.join(os.homedir(), '.chorus', 'model-pricing.json');
const DISK_TTL_MS = 24 * 60 * 60 * 1000;
const MEMORY_TTL_MS = 24 * 60 * 60 * 1000;
const FETCH_TIMEOUT_MS = 8_000;
const FAILURE_COOLDOWN_MS = 60_000;
const OPENROUTER_MODELS_URL = 'https://openrouter.ai/api/v1/models';

interface PricingSnapshot {
  fetchedAt: number;
  /** Normalized model key → price. Both unprefixed (`gemini-2-5-pro`) and
   *  vendor-prefixed (`google/gemini-2-5-pro`) variants are stored so
   *  callers using either form match. */
  prices: Record<string, ModelPrice>;
}

let memoryCache: PricingSnapshot | null = null;
let inflightFetch: Promise<PricingSnapshot | null> | null = null;
// Negative cache: last failed-fetch timestamp. Suppresses re-fetch
// attempts for FAILURE_COOLDOWN_MS so an offline daemon doesn't pay an
// 8s timeout on every reviewer/doer that lacks costUsd. opencode-cli-2
// + cli-3 + openrouter all flagged this as a hot-path latency bug.
let lastFetchFailureAt: number | null = null;

/**
 * Lowercase + replace dots with dashes. Used on both lookup key and
 * catalog ids so `claude-opus-4-7` matches `anthropic/claude-opus-4.7`.
 */
function normalize(id: string): string {
  return id.toLowerCase().replace(/\./g, '-');
}

/**
 * Strip a leading `openrouter:` voice-id prefix the cockpit uses for
 * OpenRouter voices — the actual model id is everything after the colon.
 * Gateway prefixes (`opencode-go/`, etc.) are NOT stripped here; instead
 * `getModelPricing` falls back to a bare-id lookup when the full
 * normalized form misses. This handles `opencode-go/kimi-k2.6` →
 * bare `kimi-k2-6` → catalog hit on `moonshotai/kimi-k2.6`.
 */
function stripVoicePrefix(id: string): string {
  if (id.startsWith('openrouter:')) return id.slice('openrouter:'.length);
  return id;
}

interface OpenRouterModel {
  id?: unknown;
  pricing?: {
    prompt?: unknown;
    completion?: unknown;
  };
}

interface OpenRouterModelsResponse {
  data?: unknown;
}

async function fetchOpenRouterCatalog(): Promise<PricingSnapshot | null> {
  try {
    const ac = new AbortController();
    const timer = setTimeout(() => ac.abort(), FETCH_TIMEOUT_MS);
    const res = await fetch(OPENROUTER_MODELS_URL, { signal: ac.signal });
    clearTimeout(timer);
    if (!res.ok) return null;
    const body = (await res.json()) as OpenRouterModelsResponse;
    if (!Array.isArray(body.data)) return null;
    const prices: Record<string, ModelPrice> = {};
    for (const raw of body.data as OpenRouterModel[]) {
      if (typeof raw.id !== 'string' || !raw.id) continue;
      const promptStr = raw.pricing?.prompt;
      const completionStr = raw.pricing?.completion;
      const inputCost =
        typeof promptStr === 'string' ? Number.parseFloat(promptStr) : NaN;
      const outputCost =
        typeof completionStr === 'string'
          ? Number.parseFloat(completionStr)
          : NaN;
      if (!Number.isFinite(inputCost) || !Number.isFinite(outputCost)) continue;
      const fullId = raw.id; // e.g. "anthropic/claude-opus-4.7"
      const normalizedFull = normalize(fullId);
      const bareId = fullId.includes('/') ? fullId.split('/').pop() : fullId;
      const normalizedBare = bareId ? normalize(bareId) : null;
      const price: ModelPrice = {
        inputCostPerToken: inputCost,
        outputCostPerToken: outputCost,
      };
      // Index both forms. Bare form lets a chorus call with model id
      // `gemini-2.5-pro` (no vendor prefix) match `google/gemini-2.5-pro`;
      // prefixed form lets a call with `x-ai/grok-4` match exactly.
      // Bare collisions across vendors are possible (theoretically two
      // vendors could ship a model with the same suffix); first-wins is
      // fine for chorus's lineage-disambiguated calls.
      if (normalizedBare && !(normalizedBare in prices)) {
        prices[normalizedBare] = price;
      }
      prices[normalizedFull] = price;
    }
    return { fetchedAt: Date.now(), prices };
  } catch {
    return null;
  }
}

function loadDiskCache(): PricingSnapshot | null {
  try {
    if (!fs.existsSync(CACHE_PATH)) return null;
    const raw = fs.readFileSync(CACHE_PATH, 'utf-8');
    const parsed = JSON.parse(raw) as PricingSnapshot;
    if (
      typeof parsed.fetchedAt !== 'number' ||
      !parsed.prices ||
      typeof parsed.prices !== 'object'
    ) {
      return null;
    }
    if (Date.now() - parsed.fetchedAt > DISK_TTL_MS) return null;
    return parsed;
  } catch {
    return null;
  }
}

function persistDiskCache(snapshot: PricingSnapshot): void {
  try {
    const dir = path.dirname(CACHE_PATH);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    // Atomic write: render to a temp file then rename. POSIX rename is
    // atomic on the same filesystem, so a crash mid-write leaves either
    // the old file or the fully-written new file — never a truncated one.
    // Reviewers (codex, opencode-cli-2/3/4, openrouter) all flagged the
    // direct writeFileSync as a corruption risk on crash.
    const tmpPath = `${CACHE_PATH}.tmp.${process.pid}`;
    fs.writeFileSync(tmpPath, JSON.stringify(snapshot), { encoding: 'utf-8' });
    fs.renameSync(tmpPath, CACHE_PATH);
  } catch {
    /* best-effort — pricing cache is informational */
  }
}

/**
 * Ensure the in-memory cache is populated and fresh. Loads the disk
 * copy if it exists and is within DISK_TTL_MS, otherwise fetches from
 * OpenRouter. Concurrent callers share an inflight promise so a hot
 * daemon boot doesn't fan out N fetches. A failed fetch parks the
 * cooldown so the daemon doesn't re-issue 8s-timeout calls back-to-back.
 *
 * Memory TTL (MEMORY_TTL_MS) was added after opencode-cli-4 flagged
 * "memory cache never expires" — a long-running daemon would otherwise
 * serve week-old pricing despite the disk file expiring on schedule.
 */
async function ensureCache(): Promise<PricingSnapshot | null> {
  const now = Date.now();
  if (memoryCache && now - memoryCache.fetchedAt < MEMORY_TTL_MS) {
    return memoryCache;
  }
  // Memory expired — fall through to disk/fetch. Don't clear `memoryCache`
  // yet; a failed refresh below should still be able to serve a stale
  // value rather than nothing (graceful degradation when the network is
  // out and nothing newer is available).
  const disk = loadDiskCache();
  if (disk) {
    memoryCache = disk;
    return memoryCache;
  }
  // Negative cache: if a recent fetch failed, don't retry within the
  // cooldown window. Returns the stale memory cache (if any) so callers
  // get a best-effort answer instead of paying another 8s timeout.
  if (
    lastFetchFailureAt !== null &&
    now - lastFetchFailureAt < FAILURE_COOLDOWN_MS
  ) {
    return memoryCache;
  }
  if (inflightFetch) return inflightFetch;
  inflightFetch = (async () => {
    try {
      const fetched = await fetchOpenRouterCatalog();
      if (fetched) {
        memoryCache = fetched;
        lastFetchFailureAt = null;
        persistDiskCache(fetched);
      } else {
        lastFetchFailureAt = Date.now();
      }
      return memoryCache;
    } finally {
      inflightFetch = null;
    }
  })();
  return inflightFetch;
}

/**
 * Look up pricing for a given chorus model id. Returns null when the
 * model isn't in the OpenRouter catalog or the catalog isn't reachable
 * (caller falls back to "no shadow cost recorded" — degraded gracefully).
 *
 * Lookup strategy (each step skipped if previous matched):
 *   1. Full normalized id — `claude-opus-4-7` → catalog `claude-opus-4-7`,
 *      `x-ai/grok-4` → catalog `x-ai/grok-4`.
 *   2. Bare suffix after the last `/` — handles gateway-prefixed ids
 *      like `opencode-go/kimi-k2.6` whose prefix isn't a known vendor;
 *      the bare `kimi-k2-6` key matches OpenRouter's `moonshotai/kimi-k2.6`.
 *      First-wins on bare-id collisions across vendors is acknowledged
 *      in fetchOpenRouterCatalog.
 */
export async function getModelPricing(
  modelId: string,
): Promise<ModelPrice | null> {
  if (!modelId) return null;
  const cache = await ensureCache();
  if (!cache) return null;
  const normalized = normalize(stripVoicePrefix(modelId));
  const direct = cache.prices[normalized];
  if (direct) return direct;
  if (normalized.includes('/')) {
    const bare = normalized.slice(normalized.lastIndexOf('/') + 1);
    return cache.prices[bare] ?? null;
  }
  return null;
}

/**
 * Convert (inputTokens, outputTokens, cachedInputTokens, model) → USD
 * cost. Returns undefined when pricing is unavailable so callers can
 * leave costUsd unset rather than reporting a fake $0.
 *
 * Contract: `inputTokens` and `cachedInputTokens` are MUTUALLY EXCLUSIVE
 * partitions of total input — Claude Code emits them this way (the
 * `usage.input_tokens` field is non-cached only; cached lives in
 * `cache_read_input_tokens`). We sum both and price at the full input
 * rate. OpenRouter's catalog doesn't expose per-model cached pricing
 * reliably, so cached-as-full-price is intentional: it over-estimates
 * the plan-equivalent spend by 0–10% on cache-heavy reviews, which is
 * the right direction for "what your subscription saves you" framing
 * (never under-promise the saving). Reviewers caught this — earlier
 * draft accepted cachedInputTokens but never read it, leaving the
 * contract ambiguous.
 */
export async function synthesizeCostUsd(
  modelId: string | undefined,
  usage: {
    inputTokens?: number;
    outputTokens?: number;
    cachedInputTokens?: number;
  },
): Promise<number | undefined> {
  if (!modelId) return undefined;
  const inputTokens = usage.inputTokens ?? 0;
  const cachedInputTokens = usage.cachedInputTokens ?? 0;
  const outputTokens = usage.outputTokens ?? 0;
  const totalInput = inputTokens + cachedInputTokens;
  if (totalInput <= 0 && outputTokens <= 0) return undefined;
  const price = await getModelPricing(modelId);
  if (!price) return undefined;
  const cost =
    totalInput * price.inputCostPerToken +
    outputTokens * price.outputCostPerToken;
  return Number.isFinite(cost) && cost >= 0 ? cost : undefined;
}

// Test seam — exercised from tests/model-pricing.test.ts.
export const _testing = {
  normalize,
  stripVoicePrefix,
  setMemoryCache: (snapshot: PricingSnapshot | null): void => {
    memoryCache = snapshot;
    inflightFetch = null;
  },
  getMemoryCache: (): PricingSnapshot | null => memoryCache,
  CACHE_PATH,
  DISK_TTL_MS,
};
