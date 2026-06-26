/**
 * Logger substrate tests — round-2-deferred §3.
 *
 * Pins:
 *   - Each emit produces exactly one JSON line on the writer
 *   - Bound fields (chatId / phase / role / lineage) propagate to children
 *   - `info(msg)` and `info(fields, msg)` overloads both work
 *   - Level threshold gates output (debug suppressed at info threshold)
 *   - JSON output is parseable + carries the level + ts fields
 */
import { describe, it, expect } from 'vitest';
import { createLogger, chatLogger, type LogFields } from '../src/lib/logger';

function captureLogger(level?: 'debug' | 'info' | 'warn' | 'error') {
  const lines: string[] = [];
  const log = createLogger({
    _writer: (line) => lines.push(line),
    _level: level,
  });
  return { log, lines };
}

describe('createLogger', () => {
  it('emits one JSON line per call with pino-shaped time + level + msg', () => {
    const { log, lines } = captureLogger('info');
    log.info('hello world');
    expect(lines).toHaveLength(1);
    const parsed = JSON.parse(lines[0]);
    // pino wire shape: numeric level, `time` (not `ts`), pid + hostname.
    expect(parsed.level).toBe(30);
    expect(parsed.msg).toBe('hello world');
    expect(typeof parsed.time).toBe('number');
    expect(parsed.time).toBeGreaterThan(Date.now() - 5_000);
    expect(typeof parsed.pid).toBe('number');
    expect(typeof parsed.hostname).toBe('string');
  });

  it('accepts fields-only and fields-plus-msg overloads', () => {
    const { log, lines } = captureLogger('info');
    log.info({ chatId: 'c1' }, 'phase started');
    log.info({ chatId: 'c2', custom: 42 });
    const a = JSON.parse(lines[0]);
    const b = JSON.parse(lines[1]);
    expect(a.chatId).toBe('c1');
    expect(a.msg).toBe('phase started');
    expect(b.chatId).toBe('c2');
    expect(b.custom).toBe(42);
    expect(b.msg).toBeUndefined();
  });

  it('threshold suppresses lower-level lines (debug=20, info=30, warn=40, error=50)', () => {
    const { log, lines } = captureLogger('warn');
    log.debug('debug line');
    log.info('info line');
    log.warn('warn line');
    log.error('error line');
    expect(lines).toHaveLength(2);
    expect(JSON.parse(lines[0]).level).toBe(40);
    expect(JSON.parse(lines[1]).level).toBe(50);
  });

  it('threshold can be lowered to debug', () => {
    const { log, lines } = captureLogger('debug');
    log.debug({ probe: 1 });
    log.info({ probe: 2 });
    expect(lines).toHaveLength(2);
  });
});

describe('round-1 dogfood fixes', () => {
  it('user-provided fields cannot clobber level / time / pid / hostname', () => {
    // Round-1 reviewers (claude+gemini) flagged: log.info({level:'fake', ts:0})
    // could overwrite the core wire shape. Core fields now win.
    const { log, lines } = captureLogger('info');
    log.info(
      // deliberately collide every core key
      { level: 999, time: 0, pid: 999_999, hostname: 'fake.example.com' },
      'sneaky',
    );
    const parsed = JSON.parse(lines[0]);
    expect(parsed.level).toBe(30); // info, not 999
    expect(parsed.time).toBeGreaterThan(Date.now() - 5_000);
    expect(parsed.pid).toBe(process.pid);
    expect(parsed.hostname).not.toBe('fake.example.com');
    expect(parsed.msg).toBe('sneaky');
  });

  it('Error instances expand to {message, name, stack}', () => {
    const { log, lines } = captureLogger('info');
    const err = new Error('boom');
    log.error({ err }, 'something failed');
    const parsed = JSON.parse(lines[0]);
    expect(parsed.err).toMatchObject({
      message: 'boom',
      name: 'Error',
    });
    expect(typeof parsed.err.stack).toBe('string');
  });

  it('Error.cause is preserved when present', () => {
    const { log, lines } = captureLogger('info');
    const root = new Error('root cause');
    const wrapped = new Error('wrapped', { cause: root });
    log.error({ err: wrapped }, 'wrapped');
    const parsed = JSON.parse(lines[0]);
    expect(parsed.err.cause.message).toBe('root cause');
  });

  it('does not throw on a circular-reference field — emits literal "[Circular]"', () => {
    const { log, lines } = captureLogger('info');
    type Cycle = { name: string; self?: unknown };
    const cycle: Cycle = { name: 'a' };
    cycle.self = cycle;
    expect(() => log.info({ payload: cycle }, 'cycle')).not.toThrow();
    const parsed = JSON.parse(lines[0]);
    // The payload object is preserved; the inner self ref is collapsed.
    expect(parsed.payload.name).toBe('a');
    expect(parsed.payload.self).toBe('[Circular]');
  });

  it('serializes BigInt as a string instead of throwing', () => {
    const { log, lines } = captureLogger('info');
    const big = BigInt('9007199254740993');
    expect(() => log.info({ count: big }, 'big')).not.toThrow();
    const parsed = JSON.parse(lines[0]);
    expect(parsed.count).toBe('9007199254740993');
  });

  it('emits a fallback line if even the safe replacer cannot stringify', () => {
    // Trigger a hostile toJSON. The replacer doesn't intercept .toJSON,
    // so this exercises the outer try/catch in safeStringify.
    const { log, lines } = captureLogger('info');
    const hostile = {
      toJSON() {
        throw new Error('toJSON threw');
      },
    };
    expect(() => log.info({ payload: hostile }, 'should not crash')).not.toThrow();
    const parsed = JSON.parse(lines[0]);
    expect(parsed.msg).toBe('log_serialize_failed');
    expect(parsed.err).toContain('toJSON threw');
    expect(parsed.level).toBe(30);
  });
});

describe('child logger', () => {
  it('bakes bound fields into every emitted line', () => {
    const { log, lines } = captureLogger('info');
    const c = log.child({ chatId: 'abc', phase: 'review' });
    c.info('phase started');
    c.warn({ kind: 'quota_exhausted' }, 'reviewer hit limit');
    const a = JSON.parse(lines[0]);
    const b = JSON.parse(lines[1]);
    expect(a.chatId).toBe('abc');
    expect(a.phase).toBe('review');
    expect(b.chatId).toBe('abc');
    expect(b.phase).toBe('review');
    expect(b.kind).toBe('quota_exhausted');
  });

  it('grandchild merges bindings — chatId from parent, lineage from child', () => {
    const { log, lines } = captureLogger('info');
    const chat = log.child({ chatId: 'abc' });
    const reviewer = chat.child({ role: 'reviewer', lineage: 'openai' });
    reviewer.info({ agent: 'codex-cli-0' }, 'streaming');
    const parsed = JSON.parse(lines[0]);
    expect(parsed.chatId).toBe('abc');
    expect(parsed.role).toBe('reviewer');
    expect(parsed.lineage).toBe('openai');
    expect(parsed.agent).toBe('codex-cli-0');
  });

  it('per-call fields override bound fields when keys collide', () => {
    const { log, lines } = captureLogger('info');
    const c = log.child({ phase: 'review' });
    c.info({ phase: 'plan' }, 'override-test');
    expect(JSON.parse(lines[0]).phase).toBe('plan');
  });

  it('chatLogger convenience builder bakes chatId', () => {
    // Uses the module-singleton root, so we can't assert lines without
    // hijacking stdout — but we can confirm the returned logger emits a
    // child shape with the right field name.
    const c = chatLogger('xyz');
    expect(typeof c.info).toBe('function');
    expect(typeof c.child).toBe('function');
    // Child should still carry chatId baked in.
    const grand = c.child({ phase: 'review' });
    expect(typeof grand.info).toBe('function');
  });
});

describe('JSON shape contract', () => {
  it('produces strictly-valid JSON for objects with mixed types', () => {
    const { log, lines } = captureLogger('info');
    const fields: LogFields = {
      chatId: 'abc',
      requestId: 'req-123',
      role: 'reviewer',
      lineage: 'anthropic',
      retries: 3,
      ok: true,
      tags: ['a', 'b'],
    };
    log.info(fields, 'mixed types');
    const parsed = JSON.parse(lines[0]);
    expect(parsed.chatId).toBe('abc');
    expect(parsed.requestId).toBe('req-123');
    expect(parsed.role).toBe('reviewer');
    expect(parsed.lineage).toBe('anthropic');
    expect(parsed.retries).toBe(3);
    expect(parsed.ok).toBe(true);
    expect(parsed.tags).toEqual(['a', 'b']);
  });
});
