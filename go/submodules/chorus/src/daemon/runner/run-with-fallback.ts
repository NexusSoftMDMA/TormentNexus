/**
 * Per-slot model fallback chain.
 *
 * A reviewer or doer slot can list multiple models in `models[]`. When the
 * primary returns null (subprocess produced no answer — quota, rate-limit,
 * crash, empty-stream), this helper falls through to the next model in the
 * list and retries. Auth and lineage selection are stable across the
 * fallback because all models in a slot share the same lineage; only the
 * `--model X` argv changes.
 *
 * Returning `null` from `attempt` is the agreed-upon "this model didn't
 * produce an answer, please try the next one" signal. Throwing is
 * propagated unchanged — a throw means something went wrong outside the
 * model run (e.g. the chat dir disappeared) and falling through wouldn't
 * help. Callers that want exceptions to engage the fallback should catch
 * them inside `attempt` and return null instead.
 *
 * `onFallback(fromModel, toModel, fromIdx)` fires once per transition so
 * the runner can emit a `cli_warning` event with `reason: 'model_fallback'`
 * and the cockpit can show "claude-opus-4-7 → claude-sonnet-4-6" on the
 * card. The index tells the cockpit how deep into the chain we are.
 *
 * If `models` is empty or undefined, exactly one attempt is made with
 * `undefined` as the model — this matches the existing "no model = lineage
 * default" semantics in the shim layer.
 */
export async function runWithModelFallback<T>(
  models: string[] | undefined,
  attempt: (model: string | undefined) => Promise<T | null>,
  onFallback: (
    fromModel: string | undefined,
    toModel: string | undefined,
    fromIdx: number,
  ) => void,
): Promise<T | null> {
  const list: (string | undefined)[] =
    models && models.length > 0 ? [...models] : [undefined];

  for (let i = 0; i < list.length; i++) {
    const result = await attempt(list[i]);
    if (result !== null) return result;
    if (i < list.length - 1) {
      onFallback(list[i], list[i + 1], i);
    }
  }
  return null;
}

/**
 * v0.8 cross-lineage variant. Takes (lineage, model) tuples; each attempt
 * gets the entry so the caller can pick a different shim when lineage
 * changes. Same null-fallthrough semantics as runWithModelFallback.
 *
 * The chain is expected to be non-empty (at minimum one entry with
 * `model: undefined` for the lineage default). buildSlotFallbackChain
 * guarantees this.
 */
export interface ChainEntry {
  lineage: string;
  model: string | undefined;
}

export async function runWithChainFallback<T>(
  chain: readonly ChainEntry[],
  attempt: (entry: ChainEntry) => Promise<T | null>,
  onFallback: (from: ChainEntry, to: ChainEntry, fromIdx: number) => void,
): Promise<T | null> {
  if (chain.length === 0) return null;
  for (let i = 0; i < chain.length; i++) {
    const result = await attempt(chain[i]);
    if (result !== null) return result;
    if (i < chain.length - 1) {
      onFallback(chain[i], chain[i + 1], i);
    }
  }
  return null;
}
