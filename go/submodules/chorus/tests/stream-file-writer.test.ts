import { describe, it, expect, afterEach, beforeEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { StreamFileWriter } from '../src/daemon/runner';

let dir: string;
let target: string;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'chorus-sfw-'));
  target = path.join(dir, 'answer.md');
  fs.writeFileSync(target, '');
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('StreamFileWriter', () => {
  it('does not flush small writes immediately', () => {
    const w = new StreamFileWriter(target, 4096, 750);
    w.write('hello');
    expect(fs.readFileSync(target, 'utf-8')).toBe('');
    w.flushNow();
  });

  it('flushes synchronously when buffer crosses byte threshold', () => {
    const w = new StreamFileWriter(target, 16, 5000);
    w.write('x'.repeat(16));
    expect(fs.readFileSync(target, 'utf-8').length).toBe(16);
    w.flushNow();
  });

  it('flushNow drains any pending buffer', () => {
    const w = new StreamFileWriter(target, 4096, 5000);
    w.write('partial');
    w.flushNow();
    expect(fs.readFileSync(target, 'utf-8')).toBe('partial');
  });

  it('quiet timer flushes pending data', async () => {
    const w = new StreamFileWriter(target, 4096, 50);
    w.write('timer');
    await new Promise((r) => setTimeout(r, 120));
    expect(fs.readFileSync(target, 'utf-8')).toBe('timer');
  });

  it('write("") is a no-op', () => {
    const w = new StreamFileWriter(target, 4096, 5000);
    w.write('');
    w.flushNow();
    expect(fs.readFileSync(target, 'utf-8')).toBe('');
  });

  it('multiple writes append in order after flush', () => {
    const w = new StreamFileWriter(target, 4096, 5000);
    w.write('one ');
    w.write('two ');
    w.write('three');
    w.flushNow();
    expect(fs.readFileSync(target, 'utf-8')).toBe('one two three');
  });

  it('flips to dead state on flush failure and surfaces the error', () => {
    // Point the writer at an unwritable path (target dir was deleted).
    const deadPath = path.join(dir, 'subdir', 'no-such-parent.md');
    const w = new StreamFileWriter(deadPath, 16, 5000);
    expect(w.isDead()).toBe(false);
    // 16 bytes triggers flush, which fails because the parent dir doesn't
    // exist. The first write enqueues, the second crosses the threshold
    // and triggers an internal flushNow that fails.
    w.write('x'.repeat(20));
    expect(w.isDead()).toBe(true);
    expect(w.lastError()).toBeInstanceOf(Error);
    expect(w.lastError()?.message ?? '').toMatch(/ENOENT|no such/i);
  });

  it('subsequent writes after death are dropped silently (no further errors)', () => {
    const deadPath = path.join(dir, 'subdir', 'still-dead.md');
    const w = new StreamFileWriter(deadPath, 16, 5000);
    w.write('x'.repeat(20)); // triggers death
    expect(w.isDead()).toBe(true);
    // These should not throw, regardless of size
    expect(() => w.write('more data after death')).not.toThrow();
    expect(() => w.flushNow()).not.toThrow();
    expect(w.isDead()).toBe(true);
  });
});
