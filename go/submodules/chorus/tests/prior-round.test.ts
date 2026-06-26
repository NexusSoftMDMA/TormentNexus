import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { readPriorRoundFeedback } from '../src/daemon/runner/prior-round';

let tmpDir: string;

beforeEach(() => {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'chorus-prior-round-'));
});

afterEach(() => {
  fs.rmSync(tmpDir, { recursive: true, force: true });
});

function writeReviewerAnswer(
  chatDir: string,
  round: number,
  agent: string,
  idx: number,
  body: string,
): void {
  const dir = path.join(chatDir, `round-${round}`, `reviewer-${agent}-${idx}`);
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(path.join(dir, 'answer.md'), body, 'utf-8');
}

describe('readPriorRoundFeedback', () => {
  it('returns empty string for round 1 (no prior round)', () => {
    expect(readPriorRoundFeedback(tmpDir, 1)).toBe('');
  });

  it('returns empty string when prior round dir is missing', () => {
    expect(readPriorRoundFeedback(tmpDir, 5)).toBe('');
  });

  it('returns empty string when prior round has no reviewer dirs', () => {
    fs.mkdirSync(path.join(tmpDir, 'round-1'), { recursive: true });
    fs.mkdirSync(path.join(tmpDir, 'round-1', 'doer-claude-code'), { recursive: true });
    expect(readPriorRoundFeedback(tmpDir, 2)).toBe('');
  });

  it('inlines reviewer answers from round N-1, sorted by dir name', () => {
    writeReviewerAnswer(tmpDir, 1, 'codex-cli', 0, 'codex says request changes\n## DONE');
    writeReviewerAnswer(tmpDir, 1, 'gemini-cli', 1, 'gemini says approve\n## DONE');

    const out = readPriorRoundFeedback(tmpDir, 2);

    expect(out).toContain('## Prior round feedback');
    expect(out).toContain('did not reach reviewer consensus');
    expect(out).toContain('### Reviewer: codex-cli (#0)');
    expect(out).toContain('codex says request changes');
    expect(out).toContain('### Reviewer: gemini-cli (#1)');
    expect(out).toContain('gemini says approve');
    // Stable order: codex (0) before gemini (1)
    expect(out.indexOf('codex-cli')).toBeLessThan(out.indexOf('gemini-cli'));
  });

  it('skips reviewer dirs without an answer.md', () => {
    fs.mkdirSync(path.join(tmpDir, 'round-1', 'reviewer-claude-code-0'), { recursive: true });
    writeReviewerAnswer(tmpDir, 1, 'codex-cli', 1, 'real answer\n## DONE');

    const out = readPriorRoundFeedback(tmpDir, 2);

    expect(out).toContain('### Reviewer: codex-cli (#1)');
    expect(out).not.toContain('claude-code (#0)');
  });

  it('skips empty / whitespace-only answer files', () => {
    writeReviewerAnswer(tmpDir, 1, 'codex-cli', 0, '   \n\n   \n');
    writeReviewerAnswer(tmpDir, 1, 'gemini-cli', 1, 'real content\n## DONE');

    const out = readPriorRoundFeedback(tmpDir, 2);

    expect(out).toContain('gemini-cli (#1)');
    expect(out).not.toContain('codex-cli (#0)');
  });

  it('truncates a single oversized reviewer answer to 16KB with a marker', () => {
    const big = 'x'.repeat(20 * 1024);
    writeReviewerAnswer(tmpDir, 1, 'codex-cli', 0, big);

    const out = readPriorRoundFeedback(tmpDir, 2);

    expect(out).toContain('### Reviewer: codex-cli (#0)');
    expect(out).toContain('truncated — full answer was');
    // Ensure we didn't drop the entire answer body
    expect(out).toMatch(/x{1000,}/);
  });

  it('caps the total feedback block at 64KB and marks omitted reviewers', () => {
    // Five reviewers, each 16KB → 80KB unbounded; cap kicks in after 4.
    for (let i = 0; i < 5; i++) {
      writeReviewerAnswer(tmpDir, 1, `agent-${i}`, i, 'x'.repeat(16 * 1024 - 200));
    }

    const out = readPriorRoundFeedback(tmpDir, 2);

    expect(out).toContain('reviewer');
    expect(out).toContain('omitted');
    expect(out).toContain('64-byte cap'.replace('64', String(64 * 1024)));
  });

  it('returns empty string when round <= 0 (defensive)', () => {
    expect(readPriorRoundFeedback(tmpDir, 0)).toBe('');
    expect(readPriorRoundFeedback(tmpDir, -1)).toBe('');
  });
});
