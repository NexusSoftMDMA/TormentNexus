import { z } from 'zod';
import { getDb } from './connection.js';

const PhaseEventSchema = z.object({
  id: z.number().int(),
  chat_id: z.string(),
  phase_idx: z.number().int(),
  phase_kind: z.enum([
    'plan',
    'spec',
    'tests',
    'implement',
    'review',
    'verify',
    'divergence',
    'review_only',
  ]),
  role: z.enum(['doer', 'reviewer']),
  agent_id: z.string().nullable(),
  // 'warning' is the persisted state for cli_warning events that aren't
  // terminal failures — model_fallback transitions in particular. Pre-fix
  // every cli_warning was stored as 'errored', which made a successful
  // per-slot fallback look like a reviewer crash in audit logs. Replay
  // (phaseEventToRunnerEvent) ignores this state the same way it ignores
  // 'errored' / 'reviewing' / 'approved' / 'revising', so live SSE
  // traffic is unaffected.
  state: z.enum([
    'drafting',
    'submitted',
    'reviewing',
    'approved',
    'revising',
    'blocked',
    'errored',
    'warning',
  ]),
  output: z.string().nullable(),
  cost_usd: z.number().default(0),
  tokens_in: z.number().int().default(0),
  tokens_out: z.number().int().default(0),
  started_at: z.number().int(),
  finished_at: z.number().int().nullable(),
});

export type PhaseEvent = z.infer<typeof PhaseEventSchema>;

/**
 * Hard cap on phase_events.output length. SQLite handles big TEXT cells
 * fine in isolation, but libsql wraps each row's columns in a result
 * payload that gets reshipped on every read. A 4 MB reviewer transcript
 * stored inline turns every `phaseEvents.list(chatId)` call into a 4 MB
 * fetch — and the run page calls list() on every SSE re-attach. Cap at
 * 256 KB; if a real output is bigger we keep the head + tail and emit a
 * truncation marker pointing at the on-disk artifact dir. Long-form
 * artifacts already live under ~/.chorus/chats/<id>/.
 */
const MAX_PHASE_OUTPUT_BYTES = 256 * 1024;

function buildTruncationMarker(chatId: string): string {
  return `\n\n... [truncated — see ~/.chorus/chats/${chatId}/ for full transcript] ...\n\n`;
}

function capOutput(output: string | null, chatId: string): string | null {
  if (output === null) return null;
  // Byte length, not char count — SQLite stores UTF-8 and a 4 MB string
  // of multi-byte chars would still bloat the row.
  const byteLen = Buffer.byteLength(output, 'utf-8');
  if (byteLen <= MAX_PHASE_OUTPUT_BYTES) return output;
  // Keep head (192 KB) + tail (32 KB). Bytes-based slicing on the
  // underlying Buffer to stay under the cap even with multi-byte runes.
  const buf = Buffer.from(output, 'utf-8');
  const headBytes = 192 * 1024;
  const tailBytes = 32 * 1024;
  const head = buf.subarray(0, headBytes).toString('utf-8');
  const tail = buf.subarray(buf.length - tailBytes).toString('utf-8');
  return head + buildTruncationMarker(chatId) + tail;
}

export const phaseEvents = {
  async create(event: Omit<PhaseEvent, 'id'>): Promise<PhaseEvent> {
    const db = await getDb();
    const validated = PhaseEventSchema.omit({ id: true }).parse(event);
    const output = capOutput(validated.output, validated.chat_id);

    const result = await db.execute({
      sql: `
        INSERT INTO phase_events (chat_id, phase_idx, phase_kind, role, agent_id, state, output, cost_usd, tokens_in, tokens_out, started_at, finished_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `,
      args: [
        validated.chat_id,
        validated.phase_idx,
        validated.phase_kind,
        validated.role,
        validated.agent_id,
        validated.state,
        output,
        validated.cost_usd,
        validated.tokens_in,
        validated.tokens_out,
        validated.started_at,
        validated.finished_at,
      ],
    });

    // libsql returns lastInsertRowid as bigint; cast to number for the
    // existing API. Chorus row counts stay well under 2^53.
    const id = Number(result.lastInsertRowid);
    const row = await phaseEvents.getById(id);
    if (!row) throw new Error(`phaseEvents.create: row vanished: ${id}`);
    return row;
  },

  async list(chatId: string): Promise<PhaseEvent[]> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT * FROM phase_events WHERE chat_id = ? ORDER BY phase_idx, id',
      args: [chatId],
    });
    return result.rows.map((row) => PhaseEventSchema.parse(row));
  },

  async getById(id: number): Promise<PhaseEvent | null> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT * FROM phase_events WHERE id = ?',
      args: [id],
    });
    if (result.rows.length === 0) return null;
    return PhaseEventSchema.parse(result.rows[0]);
  },

  async update(
    id: number,
    partial: Partial<Omit<PhaseEvent, 'id' | 'started_at'>>,
  ): Promise<PhaseEvent> {
    const db = await getDb();
    const event = await phaseEvents.getById(id);
    if (!event) {
      throw new Error(`Phase event ${id} not found`);
    }

    // CRITICAL: distinguish "output omitted from partial" (preserve
    // existing) from "output explicitly null" (caller wants to clear).
    // The naive `partial.output ?? event.output` collapses both into
    // "preserve" because null ?? x → x, which would silently drop
    // intentional clears. Detect via the `in` operator on the typed key.
    // Already-capped outputs in event.output pass through unchanged
    // (cap is idempotent); only newly-supplied outputs need re-capping.
    const nextOutput =
      'output' in partial
        ? capOutput(partial.output ?? null, event.chat_id)
        : event.output;
    const updated = { ...event, ...partial, output: nextOutput };

    await db.execute({
      sql: `
        UPDATE phase_events
        SET chat_id = ?, phase_idx = ?, phase_kind = ?, role = ?, agent_id = ?, state = ?, output = ?, cost_usd = ?, tokens_in = ?, tokens_out = ?, finished_at = ?
        WHERE id = ?
      `,
      args: [
        updated.chat_id,
        updated.phase_idx,
        updated.phase_kind,
        updated.role,
        updated.agent_id,
        updated.state,
        updated.output,
        updated.cost_usd,
        updated.tokens_in,
        updated.tokens_out,
        updated.finished_at,
        id,
      ],
    });

    const row = await phaseEvents.getById(id);
    if (!row) throw new Error(`phaseEvents.update: row vanished: ${id}`);
    return row;
  },
};
