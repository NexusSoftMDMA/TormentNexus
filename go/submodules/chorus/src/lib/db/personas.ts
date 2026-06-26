import { z } from 'zod';
import { getDb } from './connection.js';

const PersonaRowSchema = z.object({
  id: z.string(),
  label: z.string(),
  one_liner: z.string(),
  system_prompt: z.string(),
  recommended_lineage: z.string().nullable(),
  builtin: z.coerce.boolean(),
  forked_from: z.string().nullable(),
  created_at: z.number().int(),
  updated_at: z.number().int(),
});

export type PersonaRow = z.infer<typeof PersonaRowSchema>;

export interface PersonaUpsertInput {
  id: string;
  label: string;
  one_liner: string;
  system_prompt: string;
  recommended_lineage?: string | null;
  builtin?: boolean;
  forked_from?: string | null;
}

export const personas = {
  async upsert(input: PersonaUpsertInput): Promise<PersonaRow> {
    const db = await getDb();
    const now = Date.now();
    const existing = await personas.getById(input.id);

    await db.execute({
      sql: `
        INSERT OR REPLACE INTO personas
          (id, label, one_liner, system_prompt, recommended_lineage, builtin, forked_from, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
      `,
      args: [
        input.id,
        input.label,
        input.one_liner,
        input.system_prompt,
        input.recommended_lineage ?? null,
        input.builtin ? 1 : 0,
        input.forked_from ?? null,
        existing?.created_at ?? now,
        now,
      ],
    });

    const row = await personas.getById(input.id);
    if (!row) throw new Error(`personas.upsert: row vanished: ${input.id}`);
    return row;
  },

  async list(): Promise<PersonaRow[]> {
    const db = await getDb();
    const result = await db.execute('SELECT * FROM personas ORDER BY label ASC');
    return result.rows.map((row) => PersonaRowSchema.parse(row));
  },

  async getById(id: string): Promise<PersonaRow | null> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT * FROM personas WHERE id = ?',
      args: [id],
    });
    if (result.rows.length === 0) return null;
    return PersonaRowSchema.parse(result.rows[0]);
  },

  async delete(id: string): Promise<void> {
    const db = await getDb();
    await db.execute({ sql: 'DELETE FROM personas WHERE id = ?', args: [id] });
  },
};
