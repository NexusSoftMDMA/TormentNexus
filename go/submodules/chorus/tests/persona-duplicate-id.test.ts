/**
 * Unit tests for nextDuplicateId.
 *
 * Covers the slot-walk happy path, the sanity-cap fallback, and the
 * deterministic-suffix injection used to satisfy the React Compiler
 * purity rule. The helper is invoked from the personas page's
 * "duplicate row" button handler.
 */
import { describe, expect, it } from 'vitest';
import { nextDuplicateId } from '../src/lib/persona-duplicate-id.js';

describe('nextDuplicateId', () => {
  it('returns `<id>-copy` when the bare `-copy` slot is free', () => {
    const id = nextDuplicateId({
      sourceId: 'security-auditor',
      taken: new Set(['security-auditor', 'pessimist']),
    });
    expect(id).toBe('security-auditor-copy');
  });

  it('walks to `-copy-2`, `-copy-3` … when prior copies exist', () => {
    const taken = new Set([
      'security-auditor',
      'security-auditor-copy',
      'security-auditor-copy-2',
      'security-auditor-copy-3',
    ]);
    expect(nextDuplicateId({ sourceId: 'security-auditor', taken })).toBe(
      'security-auditor-copy-4',
    );
  });

  it('returns the first free slot, not the one after the highest seen', () => {
    // `copy-3` is missing — picker should fill the gap rather than
    // walk past to `copy-5`.
    const taken = new Set([
      'security-auditor',
      'security-auditor-copy',
      'security-auditor-copy-2',
      'security-auditor-copy-4',
    ]);
    expect(nextDuplicateId({ sourceId: 'security-auditor', taken })).toBe(
      'security-auditor-copy-3',
    );
  });

  it('falls through to the unique-suffix when the sanity cap is hit', () => {
    // Pre-populate every slot in [base, base-2 .. base-4] then cap the
    // walk at 5. The picker should reach the cap and call the suffix.
    const taken = new Set<string>();
    for (let i = 0; i < 5; i++) {
      taken.add(i === 0 ? 'p-copy' : `p-copy-${i + 1}`);
    }
    const id = nextDuplicateId({
      sourceId: 'p',
      taken,
      maxAttempts: 5,
      uniqueSuffix: () => 999,
    });
    expect(id).toBe('p-copy-999');
  });

  it('uses the injected suffix source instead of Date.now', () => {
    // Build a `taken` set that exhausts up to maxAttempts so the suffix
    // path actually fires. Default cap is 100 — populate copy through
    // copy-99.
    const taken = new Set<string>();
    taken.add('persona-copy');
    for (let i = 2; i < 100; i++) taken.add(`persona-copy-${i}`);

    let calls = 0;
    const id = nextDuplicateId({
      sourceId: 'persona',
      taken,
      uniqueSuffix: () => {
        calls += 1;
        return 'TEST-SUFFIX';
      },
    });
    expect(id).toBe('persona-copy-TEST-SUFFIX');
    expect(calls).toBe(1);
  });

  it('does not reuse an ID already in `taken`', () => {
    // Defensive: even when `-copy` is free we pick it; but if it's
    // taken we never return it.
    for (let attempt = 0; attempt < 50; attempt++) {
      const taken = new Set(['p', 'p-copy']);
      const result = nextDuplicateId({ sourceId: 'p', taken });
      expect(taken.has(result)).toBe(false);
    }
  });
});
