/**
 * Unit tests for buildHeadlessArgs — the pure argv builder for
 * `codex exec` invocations from chorus reviewer/doer runs.
 *
 * Locks in the --ignore-user-config flag (issues #10, #16): codex
 * loaded with the user's MCP servers/plugins has hung mid-call
 * in two independent reproductions. This test guards against
 * a regression where someone removes the flag thinking it's noise.
 */

import { describe, expect, it } from 'vitest';
import { buildHeadlessArgs } from '@/daemon/agents/codex';
import type { HeadlessSpawnOptions } from '@/daemon/agents/types';

const baseOpts: HeadlessSpawnOptions = {
  accountId: 'test-account',
  cwd: '/tmp/chorus-test',
  promptText: 'review this',
  timeoutMs: 60_000,
};

describe('buildHeadlessArgs', () => {
  it('always includes --ignore-user-config to dodge user MCP/plugin/hook hangs', () => {
    const args = buildHeadlessArgs(baseOpts);
    expect(args).toContain('--ignore-user-config');
  });

  it('always includes --skip-git-repo-check (chorus dirs are not repos)', () => {
    const args = buildHeadlessArgs(baseOpts);
    expect(args).toContain('--skip-git-repo-check');
  });

  it('reads prompt from stdin via final `-` arg', () => {
    const args = buildHeadlessArgs(baseOpts);
    expect(args[args.length - 1]).toBe('-');
  });

  it('starts with "exec" subcommand', () => {
    expect(buildHeadlessArgs(baseOpts)[0]).toBe('exec');
  });

  it('passes --model when supplied', () => {
    const args = buildHeadlessArgs({ ...baseOpts, model: 'gpt-5.5' });
    const idx = args.indexOf('--model');
    expect(idx).toBeGreaterThan(-1);
    expect(args[idx + 1]).toBe('gpt-5.5');
  });

  it('omits --model when not supplied', () => {
    const args = buildHeadlessArgs(baseOpts);
    expect(args).not.toContain('--model');
  });

  it('full sandbox → --dangerously-bypass-approvals-and-sandbox', () => {
    const args = buildHeadlessArgs({ ...baseOpts, sandbox: 'full' });
    expect(args).toContain('--dangerously-bypass-approvals-and-sandbox');
  });

  it('strict sandbox → -c sandbox_mode="read-only"', () => {
    const args = buildHeadlessArgs({ ...baseOpts, sandbox: 'strict' });
    const idx = args.indexOf('-c');
    expect(idx).toBeGreaterThan(-1);
    expect(args[idx + 1]).toBe('sandbox_mode="read-only"');
  });

  it('networkAccess → -c network override', () => {
    const args = buildHeadlessArgs({ ...baseOpts, networkAccess: true });
    expect(args).toContain('sandbox_workspace_write.network_access=true');
  });

  it('combines all flags in expected order', () => {
    const args = buildHeadlessArgs({
      ...baseOpts,
      model: 'gpt-5.5',
      sandbox: 'strict',
      networkAccess: true,
    });
    expect(args).toEqual([
      'exec',
      '--skip-git-repo-check',
      '--ignore-user-config',
      '-c', 'sandbox_mode="read-only"',
      '-c', 'sandbox_workspace_write.network_access=true',
      '--model', 'gpt-5.5',
      '-',
    ]);
  });
});
