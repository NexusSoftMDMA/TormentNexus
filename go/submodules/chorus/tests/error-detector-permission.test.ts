import { describe, it, expect } from 'vitest';
import { ErrorDetector } from '@/daemon/error-detector';
import { opencodeShim } from '@/daemon/agents/opencode';
import { kimiShim } from '@/daemon/agents/kimi';

/**
 * Regression tests for the OpenCode 2-step permission dialog handling.
 *
 * Background — see opencode.ts:recoverKeys for the full rationale. Live opencode
 * 1.14.x produces a 3-button row first, then a nested 2-button confirm when the
 * user picks "Allow always". The earlier `[Right, Enter]` sequence cleared
 * step 1 only; the test below pins both:
 *   1) The shim sends 3 keys (clears both dialogs in one shot).
 *   2) The detector regex matches step 1 ONLY, so a second poll can't re-fire
 *      the recovery against step 2 and accidentally Reject.
 */

const STEP_ONE = `
  ┃  △ Permission required
  ┃    # Show OS release info
  ┃
  ┃  $ cat /etc/os-release
  ┃
  ┃   Allow once   Allow always   Reject
`;

const STEP_TWO_NESTED_CONFIRM = `
  ┃  △ Always allow
  ┃
  ┃  This will allow the following patterns until OpenCode is restarted
  ┃
  ┃  - cat *
  ┃
  ┃   Confirm   Cancel
`;

describe('opencode permission dialog recovery', () => {
  it('shim emits exactly [Right, Enter, Enter] for permission_prompt', () => {
    expect(opencodeShim.recoverKeys?.permission_prompt).toEqual([
      'Right',
      'Enter',
      'Enter',
    ]);
  });

  it('kimi shim mirrors opencode (same upstream UI)', () => {
    expect(kimiShim.recoverKeys?.permission_prompt).toEqual([
      'Right',
      'Enter',
      'Enter',
    ]);
  });

  it('detector matches the step-1 three-button row', () => {
    const det = new ErrorDetector();
    const err = det.inspect('test-session', 'opencode', STEP_ONE);
    expect(err).not.toBeNull();
    expect(err?.kind).toBe('permission_prompt');
    expect(err?.lineage).toBe('opencode');
  });

  it('detector does NOT match the nested confirm dialog (would re-fire keys destructively)', () => {
    const det = new ErrorDetector();
    // The nested dialog contains "Always allow" in its heading, but no
    // "Allow once / Allow always / Reject" button row. The earlier broad
    // regex matched it; the new tight regex must not.
    const err = det.inspect('test-session-2', 'opencode', STEP_TWO_NESTED_CONFIRM);
    expect(err).toBeNull();
  });

  it('detector treats moonshot (kimi standalone) the same way — same upstream UI', () => {
    const det = new ErrorDetector();
    expect(det.inspect('s1', 'moonshot', STEP_ONE)?.kind).toBe('permission_prompt');
    expect(det.inspect('s2', 'moonshot', STEP_TWO_NESTED_CONFIRM)).toBeNull();
  });

  it('detector still matches Claude/Codex/Gemini broad phrasings (regression guard)', () => {
    const det = new ErrorDetector();
    // Claude — "Approve and run"
    const claudePane = '\n  Approve and run this command?\n';
    expect(det.inspect('a', 'anthropic', claudePane)?.kind).toBe('permission_prompt');
    // Codex — "Approve this call"
    const codexPane = '\n  Approve this call to write_file?\n';
    expect(det.inspect('b', 'openai', codexPane)?.kind).toBe('permission_prompt');
    // Gemini — "Allow this tool" (canonical broad phrasing)
    const geminiPane = '\n  Allow this tool to run?\n';
    expect(det.inspect('c', 'google', geminiPane)?.kind).toBe('permission_prompt');
  });

  it('detector ignores opencode pane text containing "approve" in chat content', () => {
    const det = new ErrorDetector();
    // User asked: "approve this PR" — this is chat content, not a button row.
    const pane = '\n  user: please review and approve this PR when you have time\n';
    expect(det.inspect('s', 'opencode', pane)).toBeNull();
  });

  it('detector dedup suppresses repeated permission_prompt for opencode step-1 (no re-fire spam)', () => {
    const det = new ErrorDetector();
    expect(det.inspect('s', 'opencode', STEP_ONE)?.kind).toBe('permission_prompt');
    // Same dialog, second poll — dedup kicks in
    expect(det.inspect('s', 'opencode', STEP_ONE)).toBeNull();
  });
});
