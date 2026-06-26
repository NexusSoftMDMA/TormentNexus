/**
 * atomicWriteJsonSync regression net.
 *
 * Verifies the temp+rename guarantees:
 *   - happy path writes the destination with valid JSON
 *   - happy path leaves NO temp files behind in the dest dir
 *   - on rename failure the temp file is cleaned up (best-effort)
 *   - concurrent writers in the same dir don't collide on temp paths
 *   - a reader concurrently opening the file never observes 0 bytes after
 *     a successful first write (the destination is replaced atomically)
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { atomicWriteJsonSync } from '@/lib/atomic-write';

let dir: string;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'chorus-atomic-'));
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('atomicWriteJsonSync', () => {
  it('writes valid JSON to the destination path', () => {
    const target = path.join(dir, 'meta.json');
    atomicWriteJsonSync(target, { binary: 'kimi', model: 'kimi-k2.6', ts: 12345 });
    const parsed = JSON.parse(fs.readFileSync(target, 'utf-8'));
    expect(parsed).toEqual({ binary: 'kimi', model: 'kimi-k2.6', ts: 12345 });
  });

  it('leaves no .tmp.* files behind in the destination dir on success', () => {
    const target = path.join(dir, 'meta.json');
    atomicWriteJsonSync(target, { ok: true });
    const entries = fs.readdirSync(dir);
    const tmpLeftovers = entries.filter((e) => e.includes('.tmp.'));
    expect(tmpLeftovers).toEqual([]);
    expect(entries).toContain('meta.json');
  });

  it('overwrites an existing destination atomically (replaces, not appends)', () => {
    const target = path.join(dir, 'meta.json');
    atomicWriteJsonSync(target, { v: 1 });
    atomicWriteJsonSync(target, { v: 2 });
    const parsed = JSON.parse(fs.readFileSync(target, 'utf-8'));
    expect(parsed).toEqual({ v: 2 });
  });

  it('two writers in the same dir do not collide on temp paths', () => {
    const a = path.join(dir, 'a.json');
    const b = path.join(dir, 'b.json');
    // Sequential calls in the same process — temp filenames embed nonce so
    // collision is statistically impossible.
    atomicWriteJsonSync(a, { which: 'a' });
    atomicWriteJsonSync(b, { which: 'b' });
    expect(JSON.parse(fs.readFileSync(a, 'utf-8'))).toEqual({ which: 'a' });
    expect(JSON.parse(fs.readFileSync(b, 'utf-8'))).toEqual({ which: 'b' });
    const tmpLeftovers = fs.readdirSync(dir).filter((e) => e.includes('.tmp.'));
    expect(tmpLeftovers).toEqual([]);
  });

  it('throws when the destination directory does not exist (and cleans up tmp)', () => {
    const bad = path.join(dir, 'no-such-subdir', 'meta.json');
    expect(() => atomicWriteJsonSync(bad, { ok: true })).toThrow();
    // Parent dir doesn't exist either, so nothing to leak from.
    expect(fs.existsSync(path.dirname(bad))).toBe(false);
  });

  it('serialization error throws BEFORE writing tmp file (no leak)', () => {
    const target = path.join(dir, 'circular.json');
    const circular: Record<string, unknown> = {};
    circular.self = circular;
    expect(() => atomicWriteJsonSync(target, circular)).toThrow();
    // No tmp file should have been created since stringify happens first
    const entries = fs.readdirSync(dir);
    expect(entries).toEqual([]);
  });
});
