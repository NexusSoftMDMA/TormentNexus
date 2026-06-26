/**
 * Vitest equivalents of the previous inline `runTests` self-checks in
 * src/daemon/error-detector.ts. Same fixtures, same assertions; lifted
 * into a proper test harness as part of the public-review cleanup pass.
 */
import { describe, expect, it } from 'vitest';
import { ErrorDetector } from '../src/daemon/error-detector.js';

interface OpenCodeState {
  errCount: number;
  lastErrAt: number;
  lastSuccessAt: number;
}

function getOpenCodeState(detector: ErrorDetector): Map<string, OpenCodeState> {
  return (detector as unknown as { openCodeState: Map<string, OpenCodeState> })
    .openCodeState;
}

describe('ErrorDetector.inspect — quota_exhausted (codex)', () => {
  it('parses Codex quota text and returns a quota_exhausted error', () => {
    const detector = new ErrorDetector();
    const paneText =
      "You've hit your usage limit. Upgrade to Plus to continue using Codex (https://chatgpt.com/explore/plus), or try again at Apr 30th, 2026 10:05 PM.";
    const error = detector.inspect('test-session-1', 'openai', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('quota_exhausted');
    expect(error!.lineage).toBe('openai');
    expect(error!.message).toContain('Resets');
    expect(error!.resetAt).toBeDefined();
    expect(Number.isFinite(error!.resetAt!)).toBe(true);
  });
});

describe('ErrorDetector.inspect — token_refresh_lost (codex)', () => {
  it('flags token-refresh failures with a re-authenticate CTA', () => {
    const detector = new ErrorDetector();
    const paneText =
      'Your access token could not be refreshed because your refresh token was already used. Please log out and sign in again.';
    const error = detector.inspect('test-session-2', 'openai', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('token_refresh_lost');
    expect(error!.lineage).toBe('openai');
    expect(error!.cta ?? '').toContain('Re-authenticate');
  });
});

describe('ErrorDetector.inspect — mcp_handshake_failed (codex)', () => {
  it('flags MCP handshake failures with a re-authenticate CTA', () => {
    const detector = new ErrorDetector();
    const paneText =
      'failed: handshaking with MCP server failed: Send message error Transport ... Your authentication token has been invalidated';
    const error = detector.inspect('test-session-3', 'openai', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('mcp_handshake_failed');
    expect(error!.lineage).toBe('openai');
    expect(error!.cta ?? '').toContain('Re-authenticate');
  });
});

describe('ErrorDetector.inspect — opencode_db_corrupt', () => {
  it('triggers after 3 sustained "Provider returned error" hits', () => {
    const detector = new ErrorDetector();
    const now = Date.now();
    const state = getOpenCodeState(detector);
    state.set('test-session-4a', {
      errCount: 0,
      lastErrAt: now - 70000,
      lastSuccessAt: now - 70000,
    });

    expect(
      detector.inspect('test-session-4a', 'opencode', 'Provider returned error'),
    ).toBeNull();
    expect(
      detector.inspect('test-session-4a', 'opencode', 'Provider returned error'),
    ).toBeNull();
    const third = detector.inspect(
      'test-session-4a',
      'opencode',
      'Provider returned error',
    );
    expect(third).not.toBeNull();
    expect(third!.kind).toBe('opencode_db_corrupt');
  });

  it('does NOT trigger when an interleaved success sentinel resets the counter', () => {
    const detector = new ErrorDetector();
    detector.inspect('test-session-4b', 'opencode', 'Provider returned error');
    detector.inspect('test-session-4b', 'opencode', 'Provider returned error');
    detector.inspect('test-session-4b', 'opencode', '## DONE');
    const err = detector.inspect(
      'test-session-4b',
      'opencode',
      'Provider returned error',
    );
    expect(err).toBeNull();
  });
});

describe('ErrorDetector.reset', () => {
  it('removes per-session state', () => {
    const detector = new ErrorDetector();
    const state = getOpenCodeState(detector);
    state.set('test-session-5', {
      errCount: 10,
      lastErrAt: Date.now(),
      lastSuccessAt: Date.now() - 100000,
    });
    detector.reset('test-session-5');
    expect(state.has('test-session-5')).toBe(false);
  });
});

describe('ErrorDetector.cleanup', () => {
  it('removes sessions idle longer than the cutoff and keeps fresh ones', () => {
    const detector = new ErrorDetector();
    const state = getOpenCodeState(detector);
    const now = Date.now();
    state.set('stale-session', {
      errCount: 1,
      lastErrAt: now - 600000,
      lastSuccessAt: now - 600000,
    });
    state.set('fresh-session', {
      errCount: 1,
      lastErrAt: now - 10000,
      lastSuccessAt: now - 10000,
    });
    detector.cleanup(300000);
    expect(state.has('stale-session')).toBe(false);
    expect(state.has('fresh-session')).toBe(true);
  });
});

describe('ErrorDetector.inspect — non-matching lineages', () => {
  it('returns null for lineages that have no detectors', () => {
    const detector = new ErrorDetector();
    const error = detector.inspect('test-session-7', 'anthropic', 'Some random output');
    expect(error).toBeNull();
  });
});

describe('ErrorDetector.inspect — quota_exhausted (anthropic)', () => {
  it('parses Claude usage-limit text and surfaces a CTA', () => {
    const detector = new ErrorDetector();
    const paneText =
      'Claude usage limit reached. Your limit will reset at Mar 5th, 2026 9:00 PM.';
    const error = detector.inspect('claude-1', 'anthropic', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('quota_exhausted');
    expect(error!.lineage).toBe('anthropic');
    expect(error!.cta).toContain('Switch to a different Claude account');
  });
});

describe('ErrorDetector.inspect — quota_exhausted (gemini)', () => {
  it('flags RESOURCE_EXHAUSTED as quota_exhausted', () => {
    const detector = new ErrorDetector();
    const paneText = '{"code":429,"message":"RESOURCE_EXHAUSTED: Quota exceeded"}';
    const error = detector.inspect('gem-1', 'google', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('quota_exhausted');
    expect(error!.lineage).toBe('google');
  });

  it('surfaces ModelNotFoundError as a Connect-page CTA', () => {
    const detector = new ErrorDetector();
    const paneText =
      'ModelNotFoundError: 404 Not Found: model gemini-3-flash does not exist';
    const error = detector.inspect('gem-2', 'google', paneText);
    expect(error).not.toBeNull();
    expect(error!.cta).toContain('Pick a different Gemini model');
  });
});

describe('ErrorDetector.inspect — quota_exhausted (opencode)', () => {
  it('detects OpenCode-Go subscription out of credits', () => {
    const detector = new ErrorDetector();
    const paneText = 'Error: subscription quota exceeded — please top up';
    const error = detector.inspect('oc-1', 'opencode', paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('quota_exhausted');
    expect(error!.cta).toContain('Top up at opencode.ai');
  });
});

describe('ErrorDetector.inspect — generic auth prompt across CLIs', () => {
  it.each([
    ['anthropic', 'Please run `claude login` to authenticate.'],
    ['openai', 'Run `codex login` to sign in to ChatGPT.'],
    ['google', 'Authentication required. Run gcloud auth.'],
    ['opencode', 'Error: not logged in — run opencode auth login'],
    ['moonshot', 'kimi: not logged in. Please log in to continue.'],
  ])('flags %s "please log in" prompts', (lineage, paneText) => {
    const detector = new ErrorDetector();
    const error = detector.inspect(`auth-${lineage}`, lineage, paneText);
    expect(error).not.toBeNull();
    expect(error!.kind).toBe('token_refresh_lost');
    expect(error!.cta).toContain('login');
  });
});

describe('Codex quota reset-time parsing (via inspect)', () => {
  it.each([
    'Apr 30th, 2026 10:05 PM',
    'May 1st, 2026 3:15 AM',
    'June 22nd, 2026 11:59 PM',
    'July 3rd, 2026 12:00 AM',
  ])('parses ordinal-date "%s" into a finite resetAt', (resetText) => {
    const detector = new ErrorDetector();
    const paneText = `You've hit your usage limit. Upgrade to Plus to continue using Codex (https://chatgpt.com/explore/plus), or try again at ${resetText}.`;
    const error = detector.inspect('rt-' + resetText, 'openai', paneText);
    expect(error).not.toBeNull();
    expect(error!.resetAt).toBeDefined();
    expect(Number.isFinite(error!.resetAt!)).toBe(true);
  });
});
