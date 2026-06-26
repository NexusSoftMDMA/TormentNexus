/**
 * Voices DB seam tests — CRUD, filters, immutable-ID upsert semantics for
 * single-model CLIs (model_id rewrites without ID change), DELETE allowed
 * for cli-sourced rows.
 *
 * Mirrors the per-test isolation pattern from tests/db.test.ts.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import os from 'os';
import path from 'path';
import fs from 'fs';
import { randomUUID } from 'crypto';

import { _resetDbForTests, getDb, voices } from '@/lib/db';

let dbPath: string;

beforeEach(async () => {
  dbPath = path.join(os.tmpdir(), `chorus-voices-${randomUUID()}.db`);
  process.env.CHORUS_DB_PATH = dbPath;
  await _resetDbForTests();
  await getDb();
});

afterEach(async () => {
  await _resetDbForTests();
  for (const suffix of ['', '-shm', '-wal']) {
    try { fs.unlinkSync(dbPath + suffix); } catch { /* best-effort */ }
  }
  delete process.env.CHORUS_DB_PATH;
});

describe('voices schema', () => {
  it('is created on fresh DB init', async () => {
    const db = await getDb();
    const result = await db.execute(
      `SELECT name FROM sqlite_master WHERE type='table' AND name='voices'`,
    );
    expect(result.rows).toHaveLength(1);
  });
});

describe('voices.upsert', () => {
  it('inserts a new row with default enabled=true', async () => {
    const v = await voices.upsert({
      id: 'claude-code',
      label: 'Claude (claude-opus-4-7)',
      source: 'cli',
      provider: 'claude-code',
      model_id: 'claude-opus-4-7',
      lineage: 'anthropic',
    });
    expect(v.id).toBe('claude-code');
    expect(v.enabled).toBe(true);
    expect(v.created_at).toBeGreaterThan(0);
    expect(v.vendor_family).toBeNull();
  });

  it('respects explicit enabled=false on insert', async () => {
    const v = await voices.upsert({
      id: 'codex-cli',
      label: 'Codex',
      source: 'cli',
      provider: 'codex-cli',
      model_id: 'gpt-5.5',
      lineage: 'openai',
      enabled: false,
    });
    expect(v.enabled).toBe(false);
  });

  // BEHAVIORAL CONTRACT (per gem-2 round 1 BLOCKER 3):
  // Single-model CLIs use immutable IDs. On version bump, model_id +
  // label rewrite WITHOUT changing the row's primary key. created_at is
  // preserved; enabled is preserved. Templates referencing `claude-code`
  // automatically pick up the new model on next dispatch.
  it('preserves id + created_at + enabled when re-upserting with new model_id (immutable single-model CLI)', async () => {
    const first = await voices.upsert({
      id: 'claude-code',
      label: 'Claude (claude-opus-4-6)',
      source: 'cli',
      provider: 'claude-code',
      model_id: 'claude-opus-4-6',
      lineage: 'anthropic',
    });
    // User disables the voice
    await voices.update('claude-code', { enabled: false });

    await new Promise((r) => setTimeout(r, 5));

    // CLI ships a new bundled model — seed re-upserts with new model_id + label
    const second = await voices.upsert({
      id: 'claude-code',
      label: 'Claude (claude-opus-4-7)',
      source: 'cli',
      provider: 'claude-code',
      model_id: 'claude-opus-4-7',
      lineage: 'anthropic',
    });

    expect(second.id).toBe('claude-code');
    expect(second.created_at).toBe(first.created_at); // preserved
    expect(second.model_id).toBe('claude-opus-4-7'); // updated
    expect(second.label).toContain('claude-opus-4-7'); // updated
    expect(second.enabled).toBe(false); // preserved across seed
    expect(second.updated_at).toBeGreaterThanOrEqual(first.updated_at);
  });

  it('preserves user-set enabled=false on re-upsert (seed never trumps user toggle)', async () => {
    await voices.upsert({
      id: 'opencode-cli:opencode-go/kimi-k2.6',
      label: 'Kimi K2.6 (via OpenCode Go)',
      source: 'cli',
      provider: 'opencode-cli',
      model_id: 'opencode-go/kimi-k2.6',
      lineage: 'moonshot',
    });
    await voices.update('opencode-cli:opencode-go/kimi-k2.6', { enabled: false });

    // Seed loop runs again
    const v = await voices.upsert({
      id: 'opencode-cli:opencode-go/kimi-k2.6',
      label: 'Kimi K2.6 (via OpenCode Go)',
      source: 'cli',
      provider: 'opencode-cli',
      model_id: 'opencode-go/kimi-k2.6',
      lineage: 'moonshot',
    });
    expect(v.enabled).toBe(false);
  });

  // REGRESSION: pre-fix `upsert` always preserved existing.enabled, silently
  // dropping any explicit input.enabled override. That made the seed loop's
  // re-detect path a no-op — once a row was auto-disabled it could never be
  // re-enabled by the seeder, requiring a manual cockpit toggle. The fix
  // makes input.enabled win when explicitly provided.
  it('upsert respects explicit input.enabled=true on an existing disabled row', async () => {
    await voices.upsert({
      id: 'gemini-cli',
      label: 'Gemini',
      source: 'cli',
      provider: 'gemini-cli',
      model_id: 'gemini-3.1-pro-preview',
      lineage: 'google',
    });
    await voices.update('gemini-cli', {
      enabled: false,
      disabled_reason: 'auto_missing',
    });

    // Seed loop re-detects the CLI and explicitly re-enables.
    const restored = await voices.upsert({
      id: 'gemini-cli',
      label: 'Gemini',
      source: 'cli',
      provider: 'gemini-cli',
      model_id: 'gemini-3.1-pro-preview',
      lineage: 'google',
      enabled: true,
      disabled_reason: null,
    });
    expect(restored.enabled).toBe(true);
    expect(restored.disabled_reason).toBeNull();
  });

  // The other half of the contract: when input omits `enabled`, the seed
  // must NOT reset a user-toggled disable. Tested in the existing
  // "preserves user-set enabled=false" case above; this assertion just
  // pins the disabled_reason side too.
  it('voices.update flipping enabled→false stamps disabled_reason="user" when caller does not supply one', async () => {
    await voices.upsert({
      id: 'codex-cli',
      label: 'Codex',
      source: 'cli',
      provider: 'codex-cli',
      model_id: 'gpt-5.5',
      lineage: 'openai',
    });
    const updated = await voices.update('codex-cli', { enabled: false });
    expect(updated.enabled).toBe(false);
    expect(updated.disabled_reason).toBe('user');
  });

  it('voices.update flipping enabled→true clears disabled_reason', async () => {
    await voices.upsert({
      id: 'codex-cli',
      label: 'Codex',
      source: 'cli',
      provider: 'codex-cli',
      model_id: 'gpt-5.5',
      lineage: 'openai',
    });
    await voices.update('codex-cli', {
      enabled: false,
      disabled_reason: 'auto_missing',
    });
    const restored = await voices.update('codex-cli', { enabled: true });
    expect(restored.enabled).toBe(true);
    expect(restored.disabled_reason).toBeNull();
  });

  it('stores vendor_family when provided', async () => {
    const v = await voices.upsert({
      id: 'opencode-cli:opencode-go/deepseek-v4-pro',
      label: 'DeepSeek V4 Pro',
      source: 'cli',
      provider: 'opencode-cli',
      model_id: 'opencode-go/deepseek-v4-pro',
      lineage: 'opencode',
      vendor_family: 'deepseek',
    });
    expect(v.vendor_family).toBe('deepseek');
  });
});

describe('voices.list', () => {
  beforeEach(async () => {
    // Seed a varied set
    await voices.upsert({ id: 'claude-code', label: 'Claude', source: 'cli', provider: 'claude-code', model_id: 'claude-opus-4-7', lineage: 'anthropic' });
    await voices.upsert({ id: 'codex-cli', label: 'Codex', source: 'cli', provider: 'codex-cli', model_id: 'gpt-5.5', lineage: 'openai', enabled: false });
    await voices.upsert({ id: 'opencode-cli:opencode-go/kimi-k2.6', label: 'Kimi K2.6', source: 'cli', provider: 'opencode-cli', model_id: 'opencode-go/kimi-k2.6', lineage: 'moonshot' });
    await voices.upsert({ id: 'opencode-cli:opencode-go/deepseek-v4-pro', label: 'DeepSeek V4 Pro', source: 'cli', provider: 'opencode-cli', model_id: 'opencode-go/deepseek-v4-pro', lineage: 'opencode', vendor_family: 'deepseek' });
    await voices.upsert({ id: 'openrouter:moonshotai/kimi-k2', label: 'Kimi K2 (OR)', source: 'api', provider: 'openrouter', model_id: 'moonshotai/kimi-k2', lineage: 'moonshot' });
  });

  // BEHAVIORAL CONTRACT (per cdx-1 round 1 BLOCKER 1):
  // Default returns ALL voices including disabled. Fleet/management
  // surfaces need disabled rows for the re-enable workflow.
  it('default (no filter) returns all rows including disabled', async () => {
    const all = await voices.list();
    expect(all).toHaveLength(5);
    expect(all.some((v) => !v.enabled)).toBe(true);
  });

  it('filter by lineage', async () => {
    const result = await voices.list({ lineage: 'moonshot' });
    expect(result).toHaveLength(2); // opencode-cli:kimi + openrouter:kimi
    expect(result.every((v) => v.lineage === 'moonshot')).toBe(true);
  });

  it('filter by source', async () => {
    expect(await voices.list({ source: 'cli' })).toHaveLength(4);
    expect(await voices.list({ source: 'api' })).toHaveLength(1);
  });

  it('filter by provider (groups OpenCode multi-vendor in one fleet card)', async () => {
    const result = await voices.list({ provider: 'opencode-cli' });
    expect(result).toHaveLength(2);
    // Different lineages but same provider (per gem-2 BLOCKER 5)
    expect(result.map((v) => v.lineage).sort()).toEqual(['moonshot', 'opencode']);
  });

  it('filter by enabled=true (template-dropdown context)', async () => {
    const result = await voices.list({ enabled: true });
    expect(result).toHaveLength(4);
    expect(result.every((v) => v.enabled)).toBe(true);
  });

  it('filter by enabled=false', async () => {
    const result = await voices.list({ enabled: false });
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('codex-cli');
  });

  it('combines filters (provider + enabled)', async () => {
    const result = await voices.list({ provider: 'opencode-cli', enabled: true });
    expect(result).toHaveLength(2);
  });

  it('sorts by provider then label', async () => {
    const all = await voices.list();
    // claude-code, codex-cli, opencode-cli (DeepSeek before Kimi alphabetically), openrouter
    expect(all[0].provider).toBe('claude-code');
    expect(all[1].provider).toBe('codex-cli');
    expect(all[2].provider).toBe('opencode-cli');
    expect(all[3].provider).toBe('opencode-cli');
    expect(all[4].provider).toBe('openrouter');
    // Within opencode-cli, label ASC
    expect(all[2].label.localeCompare(all[3].label)).toBeLessThanOrEqual(0);
  });
});

describe('voices.getById', () => {
  it('returns a voice by id', async () => {
    await voices.upsert({ id: 'claude-code', label: 'Claude', source: 'cli', provider: 'claude-code', model_id: 'claude-opus-4-7', lineage: 'anthropic' });
    const v = await voices.getById('claude-code');
    expect(v).not.toBeNull();
    expect(v!.id).toBe('claude-code');
  });

  it('returns null for unknown id', async () => {
    expect(await voices.getById('does-not-exist')).toBeNull();
  });
});

describe('voices.update', () => {
  beforeEach(async () => {
    await voices.upsert({ id: 'claude-code', label: 'Claude', source: 'cli', provider: 'claude-code', model_id: 'claude-opus-4-7', lineage: 'anthropic' });
  });

  it('toggles enabled', async () => {
    const v = await voices.update('claude-code', { enabled: false });
    expect(v.enabled).toBe(false);
  });

  it('updates label', async () => {
    const v = await voices.update('claude-code', { label: 'Claude (custom)' });
    expect(v.label).toBe('Claude (custom)');
  });

  it('updates costs', async () => {
    const v = await voices.update('claude-code', {
      input_cost_per_mtok: 15,
      output_cost_per_mtok: 75,
    });
    expect(v.input_cost_per_mtok).toBe(15);
    expect(v.output_cost_per_mtok).toBe(75);
  });

  it('partial update preserves other fields', async () => {
    await voices.update('claude-code', { label: 'X' });
    const v = await voices.update('claude-code', { enabled: false });
    expect(v.label).toBe('X');
    expect(v.enabled).toBe(false);
  });

  it('throws for unknown id', async () => {
    await expect(voices.update('nope', { enabled: false })).rejects.toThrow();
  });

  it('bumps updated_at', async () => {
    const before = (await voices.getById('claude-code'))!;
    await new Promise((r) => setTimeout(r, 5));
    const after = await voices.update('claude-code', { label: 'New' });
    expect(after.updated_at).toBeGreaterThan(before.updated_at);
  });
});

// BEHAVIORAL CONTRACT (per gem-2 round 1 MED):
// DELETE is allowed for both cli and api source. cli rows auto-heal on
// next seed if the model is still detected by the gateway. This unblocks
// users cleaning up deprecated OpenCode models.
describe('voices.delete', () => {
  it('removes a cli-sourced row', async () => {
    await voices.upsert({ id: 'opencode-cli:opencode-go/old-model', label: 'Old', source: 'cli', provider: 'opencode-cli', model_id: 'opencode-go/old-model', lineage: 'opencode' });
    await voices.delete('opencode-cli:opencode-go/old-model');
    expect(await voices.getById('opencode-cli:opencode-go/old-model')).toBeNull();
  });

  it('removes an api-sourced row', async () => {
    await voices.upsert({ id: 'openrouter:foo/bar', label: 'Bar', source: 'api', provider: 'openrouter', model_id: 'foo/bar', lineage: 'opencode' });
    await voices.delete('openrouter:foo/bar');
    expect(await voices.getById('openrouter:foo/bar')).toBeNull();
  });

  it('is a no-op for unknown id', async () => {
    await expect(voices.delete('nope')).resolves.toBeUndefined();
  });
});
