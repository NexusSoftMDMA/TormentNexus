/**
 * SSE stream handler — `/api/v1/chats/:id/stream`. Multiplexes onto an
 * active runChat via runner-multiplex; replays past phase_events from
 * DB so a late-attach run page sees history immediately.
 *
 * ─────────────────────────────────────────────────────────────────────
 *  SSE event vocabulary — frozen v0.7. Adding a new event TYPE is
 *  non-breaking (consumers ignore unknown types). Adding a new payload
 *  KEY to an existing type is non-breaking. RENAMING a type or REMOVING
 *  a required key is a wire break — bump /api/v2 first.
 *
 *  | type             | required payload                                 |
 *  |------------------|--------------------------------------------------|
 *  | phase_start      | chatId, phaseIdx, ts                             |
 *  | phase_progress   | chatId, phaseIdx, round, role, agent, elapsedMs  |
 *  | phase_done       | chatId, phaseIdx, ts                             |
 *  | participant_done | chatId, phaseIdx, round, role, agent             |
 *  | text_delta       | chatId, text, ts                                 |
 *  | tool_call        | chatId, tool, args, ts                           |
 *  | cli_error        | chatId, error: { code, message, lineage? }       |
 *  | error            | chatId, error: { code, message, details? }       |
 *  | chat_done        | chatId, status, verdict, replay?, ts             |
 *
 *  All `error.code` values come from the canonical enum in
 *  src/daemon/api-response.ts. `ts` is unix ms (grandfathered — we
 *  intentionally did NOT convert to ISO 8601 in the v0.7 freeze; new
 *  fields shipped from v0.8 onward should use ISO 8601).
 * ─────────────────────────────────────────────────────────────────────
 */

import type { FastifyInstance } from 'fastify';
import { chats, phaseEvents, templates } from '../../lib/db/index.js';
import { chatLogger } from '../../lib/logger.js';
import type { TemplateSchema } from '../../lib/template-schema.js';
import { errorResponse } from '../api-response.js';
import type { ErrorDetector } from '../error-detector.js';
import {
  getActiveRun,
  phaseEventToRunnerEvent,
  runWithMultiplex,
  type Subscriber,
} from '../runner-multiplex.js';
import { getParsedTemplate } from '../template-cache.js';
import type { TmuxManager } from '../tmux-types.js';
import { isValidChatId } from './chats-validation.js';

const TERMINAL_STATUSES = [
  'approved',
  'merged',
  'blocked',
  'cancelled',
  'failed',
  'no_review',
] as const;

interface RegisterStreamRouteArgs {
  tmuxMgr: TmuxManager;
  errorDetector: ErrorDetector;
}

export function registerChatStreamRoute(
  fastify: FastifyInstance,
  { tmuxMgr, errorDetector }: RegisterStreamRouteArgs,
): void {
  fastify.get<{ Params: { id: string } }>('/chats/:id/stream', async (request, reply) => {
    const param = request.params.id;
    if (!isValidChatId(param)) {
      reply.code(400);
      return errorResponse('validation', 'invalid chat id');
    }

    try {
      const chat = await chats.getBySlugOrId(param);
      if (!chat) {
        reply.code(404);
        return errorResponse('not_found', 'chat not found');
      }
      // From here on, `chatId` is the row's authoritative ULID — every
      // downstream key (activeRuns, subscribers, runWithMultiplex) uses
      // the ULID, never the slug.
      const chatId = chat.id;

      const tmplRow = await templates.getById(chat.template_id);
      if (!tmplRow) {
        reply.code(404);
        return errorResponse('not_found', 'template not found');
      }

      // Cached by templateId + updated_at so SSE re-attaches don't
      // re-parse on every browser refresh.
      let template: ReturnType<typeof TemplateSchema.parse>;
      try {
        template = getParsedTemplate(tmplRow.id, tmplRow.yaml, tmplRow.updated_at);
      } catch (parseError) {
        reply.code(400);
        return errorResponse(
          'validation',
          `Invalid template: ${parseError instanceof Error ? parseError.message : String(parseError)}`,
        );
      }

      // Take ownership of the underlying socket. Without `reply.hijack()`
      // Fastify would auto-end the response when this async handler
      // returns — the SSE would close immediately after the initial
      // replay even though we still want to keep streaming live events.
      reply.hijack();

      // Set SSE headers.
      //
      // Do NOT add Content-Encoding: gzip here, and do not stick a
      // buffering proxy in front of this route. SSE is line-delimited
      // (`data: ...\n\n`); gzip's compression window batches bytes
      // until flush, which collapses many small events into one frame
      // and breaks the client's per-event parser.
      reply.raw.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
      });

      const subscriber: Subscriber = {
        paused: false,
        queue: [],
        write: (line: string) => {
          try {
            return reply.raw.write(line);
          } catch {
            /* connection closed mid-write */
            return false;
          }
        },
        close: () => {
          reply.raw.end();
        },
      };

      // Replay past phase_events from DB so a late-attach run page sees
      // history immediately instead of a blank screen. Best-effort —
      // DB doesn't capture phase_progress or cli_error, so live tail is
      // still richer.
      //
      // Backpressure: when subscriber.write() returns false (kernel
      // buffer full), every subsequent reconstructed event must go to
      // subscriber.queue — NOT keep calling write() into a paused
      // socket. The drain handler below flushes the queue once the
      // kernel buffer recovers.
      const pastEvents = await phaseEvents.list(chatId);
      for (const ev of pastEvents) {
        const reconstructed = phaseEventToRunnerEvent(chatId, ev);
        if (!reconstructed) continue;
        const line = `data: ${JSON.stringify(reconstructed)}\n\n`;
        if (subscriber.paused) {
          subscriber.queue.push(line);
          continue;
        }
        if (!subscriber.write(line)) {
          subscriber.paused = true;
        }
      }

      // If chat is already terminal, replay is enough — close after
      // sending a synthetic chat_done so the client knows it's caught up.
      if ((TERMINAL_STATUSES as readonly string[]).includes(chat.status)) {
        const line = `data: ${JSON.stringify({
          chatId,
          type: 'chat_done',
          payload: {
            status: chat.status === 'approved' ? 'completed' : chat.status,
            verdict: chat.status === 'approved' ? 'approved' : chat.status,
            ...(chat.pr_url ? { prUrl: chat.pr_url } : {}),
            ...(chat.ship_error ? { shipError: chat.ship_error } : {}),
            replay: true,
          },
          ts: chat.finished_at ?? Date.now(),
        })}\n\n`;
        subscriber.write(line);
        reply.raw.end();
        return;
      }

      // CRITICAL: clear `paused` unconditionally on drain even if the
      // queue is empty. A write that returns false with no queued
      // follow-up at drain time would otherwise leave the subscriber
      // permanently paused — every later event would funnel into the
      // queue (because dispatch in onEvent only queues when paused),
      // and no further drain ever fires (the kernel buffer is already
      // empty). Order: unpause first, then flush whatever queued up.
      const onDrain = () => {
        if (!subscriber.paused) return;
        subscriber.paused = false;
        while (subscriber.queue.length > 0) {
          const queuedLine = subscriber.queue.shift() as string;
          const canContinue = subscriber.write(queuedLine);
          if (!canContinue) {
            subscriber.paused = true;
            break;
          }
        }
      };
      reply.raw.on('drain', onDrain);

      // Either attach to an in-flight runner or fire a fresh one. The
      // singleton invariant — exactly one runChat per chatId at any
      // time — is what fixes the load-spike bug.
      const existing = getActiveRun(chatId);
      if (existing) {
        existing.subscribers.add(subscriber);
        request.raw.on('close', () => {
          existing.subscribers.delete(subscriber);
          reply.raw.removeListener('drain', onDrain);
        });
        return;
      }

      // No active run — fire one and register. Persistence + status
      // updates are part of the multiplexed onEvent so they happen
      // exactly once even when multiple SSEs subscribe.
      const run = runWithMultiplex({ chatId, template, chat, tmuxMgr, errorDetector });
      // Chain `.catch` on the ActiveRun.promise so an async rejection
      // from runChat doesn't escape as an unhandled rejection. Node
      // >= 15 hard-exits the daemon on those — exactly the failure the
      // launch-eve gemini review flagged on this stream-attached path.
      run.promise.catch((err: unknown) => {
        chatLogger(chatId).error(
          { err: err instanceof Error ? err.message : String(err) },
          'stream-attached runner failed',
        );
      });
      run.subscribers.add(subscriber);
      request.raw.on('close', () => {
        run.subscribers.delete(subscriber);
        reply.raw.removeListener('drain', onDrain);
      });
      return;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      request.log.error(error);
      // Tell Fastify "I own this socket now" before writing on
      // reply.raw, even on the error path — without this, when the
      // throw happens BEFORE the success-path reply.hijack() (e.g. a
      // DB error in chats.getBySlugOrId, or the YAML parse path
      // failing), Fastify will try to auto-serialize the handler's
      // return value AFTER we've already written headers + an error
      // frame. That double-write throws ERR_HTTP_HEADERS_SENT. hijack()
      // here is idempotent if already called.
      try {
        reply.hijack();
      } catch {
        /* already hijacked */
      }
      if (!reply.raw.headersSent) {
        reply.raw.writeHead(200, {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          Connection: 'keep-alive',
        });
        // SSE error event uses the canonical envelope shape so clients
        // can rely on `error.code` + `error.message` regardless of
        // whether the failure surfaced over REST or SSE.
        reply.raw.write(
          `data: ${JSON.stringify({
            type: 'error',
            error: { code: 'internal', message },
          })}\n\n`,
        );
      }
      reply.raw.end();
    }
  });
}
