/**
 * DB seam regression net for the libsql migration (planning/libsql-migration.md).
 *
 * Strategy: this suite must be GREEN against the current better-sqlite3
 * implementation FIRST, then again after the @libsql/client swap. Same
 * assertions, new transport — that's how we catch async ripple bugs.
 *
 * Each test gets a fresh temp DB via CHORUS_DB_PATH + _resetDbForTests().
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import os from 'os';
import path from 'path';
import fs from 'fs';
import { randomUUID } from 'crypto';

import {
  _resetDbForTests,
  chats,
  getDb,
  personas,
  phaseEvents,
  secrets,
  settings,
  templates,
} from '@/lib/db';

let dbPath: string;

beforeEach(async () => {
  dbPath = path.join(os.tmpdir(), `chorus-test-${randomUUID()}.db`);
  process.env.CHORUS_DB_PATH = dbPath;
  await _resetDbForTests();
  // Ensure the schema is loaded by triggering init now (rather than on
  // first .execute() call). This also catches any init-path regressions
  // before the per-test assertions start.
  await getDb();
});

afterEach(async () => {
  await _resetDbForTests();
  for (const suffix of ['', '-shm', '-wal']) {
    try { fs.unlinkSync(dbPath + suffix); } catch { /* best-effort */ }
  }
  delete process.env.CHORUS_DB_PATH;
});

describe('getDb() init', () => {
  it('creates schema on first open of a missing DB', async () => {
    const db = await getDb();
    const result = await db.execute(
      `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`,
    );
    const names = result.rows.map((r) => r.name as string);
    expect(names).toContain('chats');
    expect(names).toContain('phase_events');
    expect(names).toContain('templates');
    expect(names).toContain('settings');
    expect(names).toContain('secrets');
    expect(names).toContain('personas');
    expect(names).toContain('voices');
  });

  it('idempotent ALTER TABLE — re-init on existing DB does not error', async () => {
    await getDb();
    await _resetDbForTests();
    // Re-open the same file. ALTER TABLE statements should be skipped via
    // PRAGMA table_info gating; the personas CREATE IF NOT EXISTS is also
    // idempotent.
    await expect(getDb()).resolves.toBeDefined();
  });

  it('CHORUS_DB_PATH env is honoured (not the home-dir default)', async () => {
    await getDb();
    expect(fs.existsSync(dbPath)).toBe(true);
  });

  // Init-failure recovery (rejected-promise cleanup) is enforced by the
  // `.catch(() => { dbInitPromise = null; throw err; })` handler in
  // getDb(). It's intentionally not unit-tested — triggering a controlled
  // libsql init failure in-process is awkward (libsql can hang rather
  // than throw on filesystem errors), and the handler itself is two
  // lines of trivial control flow. The risk it guards against (daemon
  // locked forever after one transient init error) is real; the source
  // pattern is the safety. See planning/libsql-migration.md §6.
});

describe('chats', () => {
  it('create + getById round-trip', async () => {
    const created = await chats.create({ work: 'fix the bug', template_id: 'code-review' });
    expect(created.id).toBeTruthy();
    expect(created.work).toBe('fix the bug');
    expect(created.template_id).toBe('code-review');
    expect(created.status).toBe('drafting');
    expect(created.current_phase_idx).toBe(0);
    expect(created.yolo).toBe(false);
    expect(created.attached_files).toBeNull();
    expect(created.repo_path).toBeNull();
    expect(created.pr_url).toBeNull();
    expect(created.ship_error).toBeNull();
    expect(created.artifact).toBeNull();
    expect(created.verdict).toBeNull();
    expect(created.created_at).toBeGreaterThan(0);

    const fetched = await chats.getById(created.id);
    expect(fetched).not.toBeNull();
    expect(fetched!.id).toBe(created.id);
    expect(fetched!.work).toBe('fix the bug');
  });

  it('getById returns null for unknown id', async () => {
    expect(await chats.getById('nope')).toBeNull();
  });

  it('artifact column round-trips for review-only chats', async () => {
    const artifact = '--- a/foo\n+++ b/foo\n@@ -1 +1 @@\n-old\n+new\n';
    const created = await chats.create({
      work: 'review the diff',
      template_id: 'review-only',
      artifact,
    });
    expect(created.artifact).toBe(artifact);

    const fetched = await chats.getById(created.id);
    expect(fetched?.artifact).toBe(artifact);
  });

  it('verdict column starts null and persists update', async () => {
    const created = await chats.create({ work: 'w', template_id: 't' });
    expect(created.verdict).toBeNull();

    const updated = await chats.update(created.id, { verdict: 'request_changes' });
    expect(updated.verdict).toBe('request_changes');

    const refetched = await chats.getById(created.id);
    expect(refetched?.verdict).toBe('request_changes');
  });

  it('template_snapshot starts null and round-trips after setTemplateSnapshot', async () => {
    const created = await chats.create({ work: 'w', template_id: 'code-review' });
    expect(created.template_snapshot).toBeNull();

    const snapshot = JSON.stringify({ id: 'code-review', phases: [{ x: 1 }] });
    await chats.setTemplateSnapshot(created.id, snapshot);

    const fetched = await chats.getById(created.id);
    expect(fetched?.template_snapshot).toBe(snapshot);
  });

  it('setTemplateSnapshot is write-once — second call does not overwrite', async () => {
    const created = await chats.create({ work: 'w', template_id: 'code-review' });
    const first = JSON.stringify({ id: 'code-review', version: 1 });
    const second = JSON.stringify({ id: 'code-review', version: 2 });

    await chats.setTemplateSnapshot(created.id, first);
    await chats.setTemplateSnapshot(created.id, second); // should be a no-op

    const fetched = await chats.getById(created.id);
    expect(fetched?.template_snapshot).toBe(first);
  });

  it('chats.update does not clobber template_snapshot', async () => {
    // Regression guard: the UPDATE statement must NOT reset
    // template_snapshot to NULL. Editing the template in the cockpit (which
    // doesn't go through chats.update at all) was the original bug; this
    // test pins the contract that `update()` is also snapshot-safe.
    const created = await chats.create({ work: 'w', template_id: 'code-review' });
    const snapshot = JSON.stringify({ id: 'code-review', phases: [] });
    await chats.setTemplateSnapshot(created.id, snapshot);

    await chats.update(created.id, { status: 'reviewing' });

    const fetched = await chats.getById(created.id);
    expect(fetched?.template_snapshot).toBe(snapshot);
    expect(fetched?.status).toBe('reviewing');
  });

  it('setTemplateSnapshot does not bump updated_at', async () => {
    // The snapshot write is internal runner bookkeeping, not a user-
    // visible mutation. Bumping updated_at would re-shuffle the chat to
    // the top of the recent-list every time a chat fired, which is
    // disorienting (user didn't touch anything). The helper omits
    // `updated_at = ?` from its UPDATE — pin that contract here.
    const created = await chats.create({ work: 'w', template_id: 'code-review' });
    const before = created.updated_at;
    // Sleep one tick so any naïve `updated_at = Date.now()` would
    // produce a different value, otherwise the test would pass by
    // coincidence on a fast machine.
    await new Promise((r) => setTimeout(r, 5));

    await chats.setTemplateSnapshot(created.id, JSON.stringify({ id: 't' }));

    const fetched = await chats.getById(created.id);
    expect(fetched?.updated_at).toBe(before);
  });

  it('setTemplateSnapshot tolerates rows that already have a snapshot — no error', async () => {
    // Daemon-restart-resume scenario: runChat re-enters and calls the
    // helper again. The IS-NULL guard makes the second UPDATE affect
    // zero rows. Must not throw or surface a constraint error.
    const created = await chats.create({ work: 'w', template_id: 'code-review' });
    await chats.setTemplateSnapshot(created.id, JSON.stringify({ id: 'a' }));
    await expect(
      chats.setTemplateSnapshot(created.id, JSON.stringify({ id: 'b' })),
    ).resolves.toBeUndefined();
  });

  it('list filters by status + orders by updated_at DESC', async () => {
    const a = await chats.create({ work: 'a', template_id: 't' });
    const b = await chats.create({ work: 'b', template_id: 't' });
    await chats.update(a.id, { status: 'reviewing' });
    const all = await chats.list();
    expect(all.length).toBe(2);
    // updated_at DESC — `a` was just touched, so it comes first.
    expect(all[0].id).toBe(a.id);

    const reviewing = await chats.list({ status: 'reviewing' });
    expect(reviewing.length).toBe(1);
    expect(reviewing[0].id).toBe(a.id);

    const drafting = await chats.list({ status: 'drafting' });
    expect(drafting.length).toBe(1);
    expect(drafting[0].id).toBe(b.id);
  });

  it('list respects limit + offset', async () => {
    for (let i = 0; i < 5; i++) {
      await chats.create({ work: `c${i}`, template_id: 't' });
    }
    expect(await chats.list({ limit: 2 })).toHaveLength(2);
    expect(await chats.list({ limit: 2, offset: 4 })).toHaveLength(1);
  });

  it('update merges partial + bumps updated_at', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    const updatedAt0 = c.updated_at;
    // Sleep enough for the updated_at clock to tick.
    const start = Date.now();
    while (Date.now() === start) { /* spin */ }
    const updated = await chats.update(c.id, { status: 'merged', pr_url: 'https://example/pr/1' });
    expect(updated.status).toBe('merged');
    expect(updated.pr_url).toBe('https://example/pr/1');
    expect(updated.work).toBe('x'); // unchanged
    expect(updated.updated_at).toBeGreaterThan(updatedAt0);
    expect(updated.created_at).toBe(c.created_at); // immutable
  });

  it('cancel sets status + finished_at', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    const cancelled = await chats.cancel(c.id);
    expect(cancelled.status).toBe('cancelled');
    expect(cancelled.finished_at).toBeGreaterThan(0);
  });

  it('delete removes chat AND cascades to phase_events atomically', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'gem-1',
      state: 'submitted',
      output: 'plan body',
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    expect(await phaseEvents.list(c.id)).toHaveLength(1);

    await chats.delete(c.id);
    expect(await chats.getById(c.id)).toBeNull();
    expect(await phaseEvents.list(c.id)).toHaveLength(0);
  });

  it('attached_files passes through unchanged', async () => {
    const c = await chats.create({
      work: 'x',
      template_id: 't',
      attached_files: 'src/foo.ts,src/bar.ts',
    });
    expect(c.attached_files).toBe('src/foo.ts,src/bar.ts');
  });
});

describe('phaseEvents', () => {
  it('create returns row with auto-incremented id', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    const ev = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'review',
      role: 'reviewer',
      agent_id: 'cdx-1',
      state: 'submitted',
      output: 'looks fine',
      cost_usd: 0.01,
      tokens_in: 100,
      tokens_out: 50,
      started_at: Date.now(),
      finished_at: Date.now(),
    });
    expect(ev.id).toBeGreaterThan(0);
    expect(ev.chat_id).toBe(c.id);
    expect(ev.cost_usd).toBe(0.01);
    expect(ev.tokens_in).toBe(100);
  });

  it('list orders by phase_idx, id', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    const baseEvent = {
      chat_id: c.id,
      role: 'doer' as const,
      agent_id: null,
      state: 'submitted' as const,
      output: null,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    };
    await phaseEvents.create({ ...baseEvent, phase_idx: 1, phase_kind: 'review' });
    await phaseEvents.create({ ...baseEvent, phase_idx: 0, phase_kind: 'plan' });
    await phaseEvents.create({ ...baseEvent, phase_idx: 0, phase_kind: 'plan' });

    const list = await phaseEvents.list(c.id);
    expect(list).toHaveLength(3);
    expect(list[0].phase_idx).toBe(0);
    expect(list[1].phase_idx).toBe(0);
    expect(list[2].phase_idx).toBe(1);
    // Same phase_idx → stable id order.
    expect(list[0].id).toBeLessThan(list[1].id);
  });

  it('caps oversized output to MAX bytes with truncation marker (head + tail preserved)', async () => {
    const c = await chats.create({ work: 'big', template_id: 't' });
    const head = 'HEAD_MARKER_'.repeat(8);
    const tail = 'TAIL_MARKER_'.repeat(8);
    const filler = 'x'.repeat(512 * 1024); // 512 KB filler
    const oversized = head + filler + tail;
    const event = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'claude-code',
      state: 'submitted',
      output: oversized,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    const stored = event.output ?? '';
    expect(stored).toContain('truncated');
    expect(stored).toContain('HEAD_MARKER_');
    expect(stored).toContain('TAIL_MARKER_');
    expect(Buffer.byteLength(stored, 'utf-8')).toBeLessThanOrEqual(260 * 1024);
  });

  it('truncation marker contains the actual chat id (not literal placeholder)', async () => {
    const c = await chats.create({ work: 'cap', template_id: 't' });
    const oversized = 'x'.repeat(500 * 1024);
    const event = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'claude-code',
      state: 'submitted',
      output: oversized,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    const stored = event.output ?? '';
    expect(stored).toContain(c.id);
    expect(stored).not.toContain('<chatId>');
  });

  it('update with explicit output:null clears the stored output (regression: null was silently preserved)', async () => {
    const c = await chats.create({ work: 'clear', template_id: 't' });
    const event = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'claude-code',
      state: 'submitted',
      output: 'stored body',
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    expect(event.output).toBe('stored body');
    const cleared = await phaseEvents.update(event.id, { output: null });
    expect(cleared.output).toBeNull();
  });

  it('update without output key preserves existing output (no re-cap)', async () => {
    const c = await chats.create({ work: 'preserve', template_id: 't' });
    const original = 'preserved body';
    const event = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'claude-code',
      state: 'submitted',
      output: original,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    const after = await phaseEvents.update(event.id, { state: 'reviewing' });
    expect(after.output).toBe(original);
    expect(after.state).toBe('reviewing');
  });

  it('passes through outputs at or below the cap unchanged', async () => {
    const c = await chats.create({ work: 'small', template_id: 't' });
    const small = 'a'.repeat(1000);
    const event = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: 'claude-code',
      state: 'submitted',
      output: small,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: Date.now(),
      finished_at: null,
    });
    expect(event.output).toBe(small);
  });

  it('update merges partial without resetting started_at', async () => {
    const c = await chats.create({ work: 'x', template_id: 't' });
    const ev = await phaseEvents.create({
      chat_id: c.id,
      phase_idx: 0,
      phase_kind: 'plan',
      role: 'doer',
      agent_id: null,
      state: 'drafting',
      output: null,
      cost_usd: 0,
      tokens_in: 0,
      tokens_out: 0,
      started_at: 12345,
      finished_at: null,
    });
    const updated = await phaseEvents.update(ev.id, { state: 'submitted', output: 'done' });
    expect(updated.state).toBe('submitted');
    expect(updated.output).toBe('done');
    expect(updated.started_at).toBe(12345); // immutable
  });
});

describe('templates', () => {
  it('create + getById + list', async () => {
    const t = await templates.create('hello', 'name: hello\nphases: []\n', 'user');
    expect(t.id).toBe('hello');
    expect(t.source).toBe('user');
    expect(t.yaml).toContain('name: hello');

    expect(await templates.getById('hello')).not.toBeNull();
    expect(await templates.list()).toHaveLength(1);
  });

  // BEHAVIORAL CONTRACT (per cdx-1 review): templates uses
  // `INSERT OR REPLACE INTO templates (..., created_at, updated_at)
  //  VALUES (..., now, now)` — so re-creating wipes created_at to now.
  // The libsql migration is a pure transport swap, so this assertion
  // must remain GREEN after the swap. Personas use a different (read-then-
  // upsert) pattern that DOES preserve created_at — see persona test below.
  it('INSERT OR REPLACE wipes created_at on re-create (current behavior)', async () => {
    await templates.create('hello', 'first', 'user');
    const first = (await templates.getById('hello'))!;
    // Sleep until the clock advances (Date.now() resolution is 1ms).
    await new Promise((r) => setTimeout(r, 5));
    await templates.create('hello', 'second', 'user');
    const second = (await templates.getById('hello'))!;
    expect(second.created_at).toBeGreaterThan(first.created_at);
    expect(second.yaml).toBe('second');
  });

  it('coerces BLOB yaml to string via coerceTemplateYaml (ArrayBuffer for libsql, Buffer for better-sqlite3)', async () => {
    const db = await getDb();
    // libsql accepts Uint8Array as a BLOB arg — readback comes as ArrayBuffer.
    await db.execute({
      sql: `INSERT INTO templates (id, source, yaml, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
      args: ['blob-tmpl', 'user', new TextEncoder().encode('name: from-blob\n'), Date.now(), Date.now()],
    });
    const t = await templates.getById('blob-tmpl');
    expect(t).not.toBeNull();
    expect(typeof t!.yaml).toBe('string');
    expect(t!.yaml).toBe('name: from-blob\n');
  });
});

describe('settings', () => {
  it('JSON-string round-trip', async () => {
    await settings.set('opencode.enabled_models', ['a', 'b']);
    expect(await settings.get('opencode.enabled_models')).toEqual(['a', 'b']);
  });

  it('boolean round-trip', async () => {
    await settings.set('yolo', true);
    expect(await settings.get('yolo')).toBe(true);
  });

  it('plain string passes through (non-JSON fallback)', async () => {
    // settings.set stores raw string when value is a string; settings.get
    // tries JSON.parse first, falls back to raw.
    await settings.set('plain', 'just-a-string');
    expect(await settings.get('plain')).toBe('just-a-string');
  });

  it('get returns null for unknown key', async () => {
    expect(await settings.get('does-not-exist')).toBeNull();
  });

  it('getAll returns all parsed values', async () => {
    await settings.set('k1', 'v1');
    await settings.set('k2', { nested: 1 });
    const all = await settings.getAll();
    expect(all.k1).toBe('v1');
    expect(all.k2).toEqual({ nested: 1 });
  });

  it('set overwrites existing key', async () => {
    await settings.set('foo', 1);
    await settings.set('foo', 2);
    expect(await settings.get('foo')).toBe(2);
  });
});

describe('secrets', () => {
  it('set + get round-trip with meta', async () => {
    await secrets.set('openrouter', 'api_key', 'sk-or-test', { hint: 'Vivek personal' });
    const got = await secrets.get('openrouter');
    expect(got).not.toBeNull();
    expect(got!.kind).toBe('api_key');
    expect(got!.value).toBe('sk-or-test');
    expect(got!.meta).toBe(JSON.stringify({ hint: 'Vivek personal' }));
  });

  it('set without meta stores null', async () => {
    await secrets.set('claude-code', 'cli_subscription', 'session-token');
    expect((await secrets.get('claude-code'))!.meta).toBeNull();
  });

  it('list omits value', async () => {
    await secrets.set('openrouter', 'api_key', 'sk-or-test');
    const list = await secrets.list();
    expect(list).toHaveLength(1);
    expect((list[0] as Record<string, unknown>).value).toBeUndefined();
  });

  it('overwrites on re-set (PRIMARY KEY collision)', async () => {
    await secrets.set('openrouter', 'api_key', 'sk-1');
    await secrets.set('openrouter', 'api_key', 'sk-2');
    expect((await secrets.get('openrouter'))!.value).toBe('sk-2');
  });
});

describe('personas', () => {
  // BEHAVIORAL CONTRACT: personas.upsert reads the existing row's
  // created_at and writes it back on UPDATE — preserving the original
  // creation time. This is the OPPOSITE of templates (which wipe). The
  // libsql migration must preserve this distinction.
  it('upsert PRESERVES created_at on re-upsert (per cdx-1 review)', async () => {
    await personas.upsert({
      id: 'sentinel',
      label: 'Sentinel',
      one_liner: 'security',
      system_prompt: 'v1',
      builtin: true,
    });
    const first = (await personas.getById('sentinel'))!;
    await new Promise((r) => setTimeout(r, 5));
    await personas.upsert({
      id: 'sentinel',
      label: 'Sentinel',
      one_liner: 'security',
      system_prompt: 'v2',
      builtin: true,
    });
    const second = (await personas.getById('sentinel'))!;
    expect(second.created_at).toBe(first.created_at);
    expect(second.updated_at).toBeGreaterThanOrEqual(first.updated_at);
    expect(second.system_prompt).toBe('v2');
  });

  it('builtin coerces from 0/1 integer', async () => {
    await personas.upsert({
      id: 'builtin-row',
      label: 'X',
      one_liner: 'x',
      system_prompt: 'x',
      builtin: true,
    });
    await personas.upsert({
      id: 'user-row',
      label: 'Y',
      one_liner: 'y',
      system_prompt: 'y',
      builtin: false,
    });
    expect((await personas.getById('builtin-row'))!.builtin).toBe(true);
    expect((await personas.getById('user-row'))!.builtin).toBe(false);
  });

  it('list orders by label ASC', async () => {
    await personas.upsert({ id: 'b', label: 'Beta', one_liner: '', system_prompt: '' });
    await personas.upsert({ id: 'a', label: 'Alpha', one_liner: '', system_prompt: '' });
    const list = await personas.list();
    expect(list.map((p) => p.label)).toEqual(['Alpha', 'Beta']);
  });

  it('delete removes row', async () => {
    await personas.upsert({ id: 'tmp', label: 'T', one_liner: '', system_prompt: '' });
    expect(await personas.getById('tmp')).not.toBeNull();
    await personas.delete('tmp');
    expect(await personas.getById('tmp')).toBeNull();
  });

  it('upsert with recommended_lineage + forked_from', async () => {
    await personas.upsert({
      id: 'fork',
      label: 'Fork',
      one_liner: 'forked',
      system_prompt: 'p',
      recommended_lineage: 'anthropic',
      builtin: false,
      forked_from: 'sentinel',
    });
    const got = (await personas.getById('fork'))!;
    expect(got.recommended_lineage).toBe('anthropic');
    expect(got.forked_from).toBe('sentinel');
  });
});
