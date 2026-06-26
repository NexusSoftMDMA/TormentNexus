import { z } from 'zod';
import { getDb } from './connection.js';

const TemplateSchema = z.object({
  id: z.string(),
  source: z.enum(['builtin', 'user']),
  yaml: z.string(),
  // SQLite stores 0/1 ints; coerce to boolean for callers. Default 1
  // (true) on legacy rows and on user-authored templates that haven't
  // gone through the adapter.
  is_complete: z.preprocess((v) => {
    if (typeof v === 'boolean') return v;
    if (typeof v === 'number') return v !== 0;
    // libsql may surface INTEGER columns as bigint on some platforms.
    // Coerce via Number() so we don't need BigInt literals (which the
    // ES2017 target doesn't allow).
    if (typeof v === 'bigint') return Number(v) !== 0;
    return true;
  }, z.boolean()),
  created_at: z.number().int(),
  updated_at: z.number().int(),
});

export type Template = z.infer<typeof TemplateSchema>;

/**
 * SQLite stores text columns as TEXT, but `INSERT ... readfile(...)` and
 * some admin tools write BLOBs. Transport-specific note (verified during
 * plan review):
 *   - better-sqlite3 surfaces BLOBs as Node Buffer (extends Uint8Array).
 *   - @libsql/client surfaces BLOBs as ArrayBuffer (NOT instanceof Uint8Array).
 * Check ArrayBuffer first; fall through to Uint8Array as a defensive
 * belt in case a Buffer-typed value sneaks in via direct stmt.run paths
 * or future migrations.
 */
function coerceTemplateYaml(row: unknown): unknown {
  if (!row || typeof row !== 'object') return row;
  const r = row as Record<string, unknown>;
  if (r.yaml instanceof ArrayBuffer) {
    return { ...r, yaml: new TextDecoder().decode(new Uint8Array(r.yaml)) };
  }
  if (r.yaml instanceof Uint8Array) {
    return { ...r, yaml: new TextDecoder().decode(r.yaml) };
  }
  return r;
}

export const templates = {
  async create(
    id: string,
    yaml: string,
    source: 'builtin' | 'user' = 'user',
    isComplete: boolean = true,
  ): Promise<Template> {
    const db = await getDb();
    const now = Date.now();

    await db.execute({
      sql: `
        INSERT OR REPLACE INTO templates (id, source, yaml, is_complete, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
      `,
      args: [id, source, yaml, isComplete ? 1 : 0, now, now],
    });

    const row = await templates.getById(id);
    if (!row) throw new Error(`templates.create: row vanished: ${id}`);
    return row;
  },

  async list(): Promise<Template[]> {
    const db = await getDb();
    const result = await db.execute('SELECT * FROM templates ORDER BY created_at DESC');
    return result.rows.map((row) => TemplateSchema.parse(coerceTemplateYaml(row)));
  },

  async getById(id: string): Promise<Template | null> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT * FROM templates WHERE id = ?',
      args: [id],
    });
    if (result.rows.length === 0) return null;
    return TemplateSchema.parse(coerceTemplateYaml(result.rows[0]));
  },

  /** Hard-delete by id. Caller is responsible for refusing built-in rows
   *  before reaching here — the boot seed in src/daemon/index.ts re-creates
   *  built-in templates from `templates/*.yaml`, so a built-in delete would
   *  come back on next start. */
  async delete(id: string): Promise<void> {
    const db = await getDb();
    await db.execute({ sql: 'DELETE FROM templates WHERE id = ?', args: [id] });
  },
};
