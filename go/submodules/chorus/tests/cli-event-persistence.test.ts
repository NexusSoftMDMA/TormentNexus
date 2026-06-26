/**
 * Tests the daemon's cli_error vs cli_warning persistence contract:
 *   - cli_error  → state='errored', output starts '[cli_error] ...' or '[<error.kind>] ...'
 *   - cli_warning → state='warning', output starts '[cli_warning] ...' or '[<error.kind>] ...'
 *
 * Pre-fix every cli_warning was stored as state='errored' / [cli_error]
 * prefix. The model-fallback path (PR #36) emits cli_warning events whose
 * payload describes a *successful* fallback transition, so the audit row
 * MUST NOT look like a reviewer crash. Caught while smoke-testing the
 * fallback chain end-to-end against opencode-go/this-model-does-not-exist.
 *
 * The daemon's translation lives inline in src/daemon/index.ts (~line 305).
 * Re-implementing it here keeps the contract checkable without bringing
 * up Fastify; the source comment points at this test as the lock.
 */
import { describe, it, expect } from 'vitest';

interface RunnerEventLike {
  type: 'cli_error' | 'cli_warning';
  payload: Record<string, unknown>;
}

// Mirrors src/daemon/index.ts L305+. Update both together.
function translate(event: RunnerEventLike): { state: 'errored' | 'warning'; output: string } {
  const payload = event.payload;
  const errorObj = (payload.error as Record<string, unknown> | undefined) ?? {};
  const message =
    (errorObj.message as string | undefined) ??
    (payload.message as string | undefined) ??
    'unknown error';
  const isWarning = event.type === 'cli_warning';
  const persistedState: 'errored' | 'warning' = isWarning ? 'warning' : 'errored';
  const tag =
    (errorObj.kind as string | undefined) ?? (isWarning ? 'cli_warning' : 'cli_error');
  return { state: persistedState, output: `[${tag}] ${message}` };
}

describe('cli_error vs cli_warning persistence translation', () => {
  it('persists cli_error as state=errored with [cli_error] prefix', () => {
    const r = translate({
      type: 'cli_error',
      payload: { message: 'codex died', role: 'reviewer' },
    });
    expect(r.state).toBe('errored');
    expect(r.output).toBe('[cli_error] codex died');
  });

  it('persists cli_error with explicit error.kind keeping the kind in the prefix', () => {
    const r = translate({
      type: 'cli_error',
      payload: {
        role: 'reviewer',
        error: { kind: 'quota_exhausted', message: 'codex out of credits' },
      },
    });
    expect(r.state).toBe('errored');
    expect(r.output).toBe('[quota_exhausted] codex out of credits');
  });

  it('persists cli_warning as state=warning with [cli_warning] prefix (not errored)', () => {
    const r = translate({
      type: 'cli_warning',
      payload: {
        role: 'reviewer',
        reason: 'model_fallback',
        message:
          'Reviewer model "opencode-go/typo" produced no answer; retrying with "opencode-go/deepseek-v4-pro".',
      },
    });
    expect(r.state).toBe('warning');
    expect(r.output).toMatch(/^\[cli_warning\] Reviewer model "opencode-go\/typo"/);
  });

  it('honours an explicit error.kind on a cli_warning over the cli_warning default tag', () => {
    const r = translate({
      type: 'cli_warning',
      payload: {
        role: 'doer',
        error: { kind: 'permission_prompt_recovered', message: 'navigated dialog via Right Enter' },
      },
    });
    expect(r.state).toBe('warning');
    expect(r.output).toBe('[permission_prompt_recovered] navigated dialog via Right Enter');
  });
});
