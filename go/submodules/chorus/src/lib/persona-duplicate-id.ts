/**
 * Pick the next free ID for a duplicated persona row.
 *
 * Walks `<sourceId>-copy`, `<sourceId>-copy-2`, ... up to a sanity cap;
 * if that's exhausted (the user is duplicating an absurd number of
 * times) falls back to a unique-suffixed ID.
 *
 * Extracted to a pure helper so we can unit-test the slot-walking
 * logic AND so the React Compiler purity rule stops flagging the inline
 * `Date.now()` use as an in-render impurity. The unique-suffix source
 * is a parameter — caller passes `Date.now` in production, tests pass
 * a deterministic counter.
 */
export interface NextDuplicateIdInput {
  sourceId: string;
  taken: Set<string>;
  /** Sanity cap on `-copy-N` walks. Default 100 — user almost certainly
   *  isn't intentionally duplicating the same row 100+ times. */
  maxAttempts?: number;
  /** Injectable time source. Production passes `Date.now`; tests pass
   *  a deterministic stub. */
  uniqueSuffix?: () => number | string;
}

export function nextDuplicateId({
  sourceId,
  taken,
  maxAttempts = 100,
  uniqueSuffix = Date.now,
}: NextDuplicateIdInput): string {
  const base = `${sourceId}-copy`;
  if (!taken.has(base)) return base;
  for (let i = 2; i < maxAttempts; i++) {
    const candidate = `${base}-${i}`;
    if (!taken.has(candidate)) return candidate;
  }
  return `${base}-${uniqueSuffix()}`;
}
