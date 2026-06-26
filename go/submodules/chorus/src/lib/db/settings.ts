import { getDb } from './connection.js';

export type SettingRow = {
  key: string;
  value: string;
};

export const settings = {
  async get(key: string): Promise<unknown | null> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT value FROM settings WHERE key = ?',
      args: [key],
    });
    if (result.rows.length === 0) return null;
    const value = result.rows[0].value as string;
    try {
      return JSON.parse(value);
    } catch {
      return value;
    }
  },

  async set(key: string, value: unknown): Promise<void> {
    const db = await getDb();
    const stringValue = typeof value === 'string' ? value : JSON.stringify(value);
    await db.execute({
      sql: 'INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)',
      args: [key, stringValue],
    });
  },

  async getAll(): Promise<Record<string, unknown>> {
    const db = await getDb();
    const result = await db.execute('SELECT key, value FROM settings');
    const out: Record<string, unknown> = {};
    for (const row of result.rows) {
      const k = row.key as string;
      const v = row.value as string;
      try {
        out[k] = JSON.parse(v);
      } catch {
        out[k] = v;
      }
    }
    return out;
  },
};
