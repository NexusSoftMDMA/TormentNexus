import { z } from 'zod';
import { chatEventsBus } from '../chat-events-bus.js';
import { generateUlid, getDb } from './connection.js';

const ChatRowSchema = z.object({
  id: z.string(),
  /**
   * URL-friendly slug derived from `work` on chat creation. Nullable
   * for legacy rows; the daemon backfills these lazily on first
   * list-load. UNIQUE partial index resolves /runs/<slug> in O(1).
   */
  slug: z.string().nullable().default(null),
  work: z.string(),
  template_id: z.string(),
  status: z.enum([
    'drafting',
    'reviewing',
    'approved',
    'merged',
    'blocked',
    'cancelled',
    'failed',
    'no_review',
  ]),
  current_phase_idx: z.number().int(),
  yolo: z.coerce.boolean().default(false),
  attached_files: z.string().nullable(),
  repo_path: z.string().nullable().default(null),
  pr_url: z.string().nullable().default(null),
  ship_error: z.string().nullable().default(null),
  artifact: z.string().nullable().default(null),
  verdict: z.string().nullable().default(null),
  /**
   * Frozen template JSON captured at first run-fire. Stays NULL on chats
   * that pre-date the column. Stored as a JSON string in SQLite; readers
   * upstream parse it. The DB row keeps it as a raw string so we don't
   * pay the parse cost on list pages that never use it.
   */
  template_snapshot: z.string().nullable().default(null),
  created_at: z.number().int(),
  updated_at: z.number().int(),
  finished_at: z.number().int().nullable(),
});

export type ChatRow = z.infer<typeof ChatRowSchema>;

const CreateChatSchema = z.object({
  work: z.string(),
  template_id: z.string(),
  attached_files: z.string().optional(),
  /** Absolute path to user's repo for Ship phase. Optional. */
  repo_path: z.string().optional(),
  /** Artifact text for review-only templates. Optional at the DB layer; the
   *  chat-create endpoint enforces it when the template requires one. */
  artifact: z.string().optional(),
  /** Skip ask-user gates for this run. The runner only honours this on the
   *  ship phase today; safe to pass on any chat. */
  yolo: z.boolean().optional(),
});

export type CreateChatInput = z.infer<typeof CreateChatSchema>;

export const chats = {
  async create(input: CreateChatInput): Promise<ChatRow> {
    const db = await getDb();
    const validated = CreateChatSchema.parse(input);
    const ulid = generateUlid();
    const now = Date.now();

    // Generate a unique URL slug BEFORE insert so the returned row is
    // complete (no second SELECT/UPDATE round trip). The collision check
    // uses chats.slugExists which tolerates an in-flight partner row
    // (same slug requested concurrently) — the UNIQUE index is the
    // authoritative race-loser. Retry loop catches the constraint
    // violation, regenerates the slug, and INSERTs again. Capped at 3
    // attempts: if we can't get a unique slug after 3 collisions there's
    // something structurally wrong (clock skew, slug generator bug);
    // fail loud.
    const { generateChatSlug } = await import('../chat-slug.js');

    const MAX_SLUG_ATTEMPTS = 3;
    for (let attempt = 1; attempt <= MAX_SLUG_ATTEMPTS; attempt++) {
      const slug = await generateChatSlug({
        work: validated.work,
        templateId: validated.template_id,
        existsFn: chats.slugExists,
      });

      try {
        await db.execute({
          sql: `
            INSERT INTO chats (id, slug, work, template_id, status, current_phase_idx, yolo, attached_files, repo_path, artifact, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
          `,
          args: [
            ulid,
            slug,
            validated.work,
            validated.template_id,
            'drafting',
            0,
            validated.yolo ? 1 : 0,
            validated.attached_files || null,
            validated.repo_path || null,
            validated.artifact || null,
            now,
            now,
          ],
        });
        const row = await chats.getById(ulid);
        if (!row) throw new Error(`chats.create: row vanished after insert: ${ulid}`);
        chatEventsBus.emitChange(row.id, 'created');
        return row;
      } catch (err: unknown) {
        // libsql surfaces UNIQUE violations as Error with message
        // containing "UNIQUE constraint failed: ...idx_chats_slug" (the
        // partial-index name). Retry on this exact pattern; rethrow
        // anything else (FK violations, type errors, conn drops).
        const message = err instanceof Error ? err.message : String(err);
        const isSlugCollision = /UNIQUE constraint failed.*chats\.slug|idx_chats_slug/i.test(message);
        if (!isSlugCollision || attempt === MAX_SLUG_ATTEMPTS) throw err;
      }
    }
    // Unreachable — the loop above either returns or throws on the final attempt.
    throw new Error('chats.create: unique slug allocation failed after retries');
  },

  /** Used by generateChatSlug — does any chat already use this slug? */
  async slugExists(slug: string): Promise<boolean> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT 1 FROM chats WHERE slug = ? LIMIT 1',
      args: [slug],
    });
    return result.rows.length > 0;
  },

  async getBySlug(slug: string): Promise<ChatRow | null> {
    const db = await getDb();
    const result = await db.execute({
      sql: 'SELECT * FROM chats WHERE slug = ?',
      args: [slug],
    });
    if (result.rows.length === 0) return null;
    return ChatRowSchema.parse(result.rows[0]);
  },

  /**
   * Resolve by slug OR id. Falls back to id lookup when the slug query
   * misses, so legacy URLs (`/runs/<ULID>`) keep working forever.
   */
  async getBySlugOrId(slugOrId: string): Promise<ChatRow | null> {
    const { looksLikeSlug } = await import('../chat-slug.js');
    if (looksLikeSlug(slugOrId)) {
      const bySlug = await chats.getBySlug(slugOrId);
      if (bySlug) return bySlug;
    }
    return chats.getById(slugOrId);
  },

  async list(opts?: { status?: string; limit?: number; offset?: number }): Promise<ChatRow[]> {
    const db = await getDb();
    let sql = 'SELECT * FROM chats';
    const args: unknown[] = [];

    if (opts?.status) {
      sql += ' WHERE status = ?';
      args.push(opts.status);
    }

    sql += ' ORDER BY updated_at DESC';

    if (opts?.limit) {
      sql += ' LIMIT ?';
      args.push(opts.limit);
    }

    if (opts?.offset) {
      sql += ' OFFSET ?';
      args.push(opts.offset);
    }

    const result = await db.execute({ sql, args: args as never });
    return result.rows.map((row) => ChatRowSchema.parse(row));
  },

  async getById(id: string): Promise<ChatRow | null> {
    const db = await getDb();
    const result = await db.execute({ sql: 'SELECT * FROM chats WHERE id = ?', args: [id] });
    if (result.rows.length === 0) return null;
    return ChatRowSchema.parse(result.rows[0]);
  },

  async update(id: string, partial: Partial<Omit<ChatRow, 'id' | 'created_at'>>): Promise<ChatRow> {
    const db = await getDb();
    const chat = await chats.getById(id);
    if (!chat) {
      throw new Error(`Chat ${id} not found`);
    }

    const updated = {
      ...chat,
      ...partial,
      id: chat.id,
      created_at: chat.created_at,
      updated_at: Date.now(),
    };

    await db.execute({
      sql: `
        UPDATE chats
        SET work = ?, template_id = ?, status = ?, current_phase_idx = ?, yolo = ?, attached_files = ?, repo_path = ?, pr_url = ?, ship_error = ?, artifact = ?, verdict = ?, updated_at = ?, finished_at = ?
        WHERE id = ?
      `,
      args: [
        updated.work,
        updated.template_id,
        updated.status,
        updated.current_phase_idx,
        updated.yolo ? 1 : 0,
        updated.attached_files,
        updated.repo_path,
        updated.pr_url,
        updated.ship_error,
        updated.artifact,
        updated.verdict,
        updated.updated_at,
        updated.finished_at,
        id,
      ],
    });

    const row = await chats.getById(id);
    if (!row) throw new Error(`chats.update: row vanished: ${id}`);
    chatEventsBus.emitChange(row.id, 'updated');
    return row;
  },

  async cancel(id: string): Promise<ChatRow> {
    return chats.update(id, { status: 'cancelled', finished_at: Date.now() });
  },

  /**
   * Write-once template snapshot. The runner calls this on first fire so
   * the cockpit can render old runs against the template they actually
   * used, even if the live template gets edited later.
   *
   * Idempotent: the `IS NULL` predicate makes a second call (e.g. after
   * a daemon restart resumes a chat) a no-op. We never overwrite an
   * existing snapshot — by definition the chat already ran against it.
   *
   * `update_at` is intentionally NOT bumped here. This is internal
   * runner bookkeeping, not a user-visible mutation; bumping it would
   * reorder the chat in the cockpit's recent-list every time a chat
   * fired, even though nothing the user cares about changed.
   */
  async setTemplateSnapshot(id: string, snapshotJson: string): Promise<void> {
    const db = await getDb();
    await db.execute({
      sql: 'UPDATE chats SET template_snapshot = ? WHERE id = ? AND template_snapshot IS NULL',
      args: [snapshotJson, id],
    });
  },

  /**
   * Hard-delete a chat. Removes the row + cascades to phase_events.
   * Atomic via libsql transaction so a partial failure can't leave
   * orphaned phase_events. Caller is responsible for filesystem cleanup
   * (chat artifacts in ~/.chorus/chats/<id>) and for ensuring no active
   * session is running (cancel first if needed).
   */
  async delete(id: string): Promise<void> {
    const db = await getDb();
    const tx = await db.transaction('write');
    try {
      // Phase events first to avoid orphans (no FK enforcement, but the
      // chat semantically owns its events).
      await tx.execute({ sql: 'DELETE FROM phase_events WHERE chat_id = ?', args: [id] });
      await tx.execute({ sql: 'DELETE FROM chats WHERE id = ?', args: [id] });
      await tx.commit();
      chatEventsBus.emitChange(id, 'deleted');
    } catch (e) {
      await tx.rollback();
      throw e;
    }
  },
};
