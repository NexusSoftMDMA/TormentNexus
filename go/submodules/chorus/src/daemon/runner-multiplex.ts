/**
 * Multi-subscriber wrapper around runChat.
 *
 * Singleton runner registry — exactly one runChat per chatId, ever. SSE
 * re-attachers (browser refresh, tab open, polling, MCP wait_for_chat)
 * all subscribe to the same in-memory event bus instead of re-firing the
 * runner. Without this, every refresh of the run page used to spawn a
 * fresh doer + 2 reviewers, hammering the LLM CLIs and thrashing memory.
 *
 * onEvent fans out to every subscribed SSE and persists side effects
 * exactly once, regardless of subscriber count.
 */

import { chats, phaseEvents } from '../lib/db/index.js';
import { chatLogger } from '../lib/logger.js';
import type { TemplateSchema } from '../lib/template-schema.js';
import { ErrorDetector } from './error-detector.js';
import * as participantAborts from './participant-aborts.js';
import { runChat } from './runner.js';
import type { TmuxManager } from './tmux-types.js';

export interface Subscriber {
  /** Returns true if buffer available, false if full. */
  write: (line: string) => boolean;
  paused: boolean;
  queue: string[];
  close: () => void;
}

export interface ActiveRun {
  promise: Promise<void>;
  subscribers: Set<Subscriber>;
  abortController: AbortController;
}

const activeRuns = new Map<string, ActiveRun>();

export function getActiveRun(chatId: string): ActiveRun | undefined {
  return activeRuns.get(chatId);
}

export function abortActiveRun(chatId: string): boolean {
  const active = activeRuns.get(chatId);
  if (!active) return false;
  active.abortController.abort();
  return true;
}

/**
 * Reconstruct a RunnerEvent from a persisted phase_events row. Used to
 * replay past events to a freshly-attached SSE so the run page renders
 * the history without waiting for the next live event. Returns null for
 * rows we can't faithfully reconstruct.
 */
export function phaseEventToRunnerEvent(
  chatId: string,
  ev: Awaited<ReturnType<typeof phaseEvents.list>>[number],
): Record<string, unknown> | null {
  // State → event-type mapping for replay. `warning` was added in
  // chorus-085 for per-slot fallback success (and surfaces failed
  // reviewer attempts that nonetheless got past the threshold). Pre-
  // fix this state was unmapped → returned null → replay dropped the
  // event → on page reload the failed reviewer's card showed QUEUED
  // forever (the failure was lost). Map to phase_done so the card at
  // least surfaces as completed; the failure summary written to
  // answer.md drives the actual error display.
  const baseType =
    ev.state === 'drafting'
      ? 'phase_start'
      : ev.state === 'submitted' || ev.state === 'warning' || ev.state === 'errored'
        ? 'phase_done'
        : ev.state === 'blocked'
          ? 'phase_failed'
          : null;
  if (!baseType) {
    console.warn(
      `[chorus] phase event replay: unmapped state "${ev.state}" for chat ${chatId}`,
    );
    return null;
  }
  return {
    chatId,
    type: baseType,
    payload: {
      phaseIdx: ev.phase_idx,
      kind: ev.phase_kind,
      role: ev.role,
      agent: ev.agent_id ?? undefined,
      output: ev.output ?? undefined,
      replay: true,
    },
    ts: ev.started_at,
  };
}

interface RunWithMultiplexArgs {
  chatId: string;
  template: ReturnType<typeof TemplateSchema.parse>;
  chat: NonNullable<Awaited<ReturnType<typeof chats.getById>>>;
  tmuxMgr: TmuxManager;
  errorDetector: ErrorDetector;
}

const VALID_PHASE_KINDS = [
  'plan',
  'spec',
  'tests',
  'implement',
  'review',
  'verify',
  'divergence',
  'review_only',
] as const;
type PhaseKind = (typeof VALID_PHASE_KINDS)[number];

const VALID_CHAT_STATUSES = [
  'drafting',
  'reviewing',
  'approved',
  'merged',
  'blocked',
  'cancelled',
  'failed',
  'no_review',
] as const;
type ChatStatus = (typeof VALID_CHAT_STATUSES)[number];

function parseAttachedFiles(raw: string | null | undefined): string[] | undefined {
  if (!raw) return undefined;
  try {
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed) && parsed.every((p) => typeof p === 'string')) {
      return parsed;
    }
  } catch {
    /* ignore */
  }
  return undefined;
}

export function runWithMultiplex(args: RunWithMultiplexArgs): ActiveRun {
  const { chatId, template, chat, tmuxMgr, errorDetector } = args;

  // Explicit cancellation goes through POST /chats/:id/cancel which calls
  // entry.abortController.abort(). Closing an SSE does NOT abort.
  const abortController = new AbortController();
  const subscribers = new Set<Subscriber>();

  // Pending DB writes from onEvent. Fire-and-forget here would race
  // against the activeRuns.delete in `.finally()` below — a reattaching
  // SSE could see activeRuns empty (slot released) but read the stale
  // chats row (status='reviewing') and start a duplicate run. Drain
  // this set before releasing the slot.
  const pendingWrites = new Set<Promise<unknown>>();
  const trackWrite = <T,>(p: Promise<T>): Promise<T> => {
    pendingWrites.add(p);
    p.finally(() => pendingWrites.delete(p));
    return p;
  };

  const onEvent: Parameters<typeof runChat>[0]['onEvent'] = (event) => {
    const line = `data: ${JSON.stringify(event)}\n\n`;
    const toRemove: Subscriber[] = [];
    for (const sub of Array.from(subscribers)) {
      try {
        if (sub.paused) {
          sub.queue.push(line);
          if (sub.queue.length > 1000) {
            // Queue cap exceeded; drop subscriber to prevent unbounded
            // memory.
            toRemove.push(sub);
            sub.close();
          }
        } else {
          const canContinue = sub.write(line);
          if (!canContinue) {
            // Buffer full; pause. Drain listener (set up by the SSE
            // handler) flushes the queue when the kernel buffer recovers.
            sub.paused = true;
          }
        }
      } catch {
        /* dead subscriber */
        toRemove.push(sub);
      }
    }
    for (const sub of toRemove) {
      subscribers.delete(sub);
    }

    if (
      event.type === 'phase_start' ||
      event.type === 'phase_done' ||
      event.type === 'phase_failed'
    ) {
      const payload = event.payload as Record<string, unknown>;
      const kind = payload.kind as string;
      const phaseKind: PhaseKind = (VALID_PHASE_KINDS as readonly string[]).includes(kind)
        ? (kind as PhaseKind)
        : 'plan';
      // Fire-and-forget — onEvent is typed `(e) => void` and is called
      // synchronously from the runner; awaiting here would block the
      // entire fan-out chain. SQLite serializes writes via WAL anyway.
      // Tracked in pendingWrites so the .finally drain ensures DB state
      // is consistent before activeRuns.delete fires.
      void trackWrite(
        phaseEvents
          .create({
            chat_id: chatId,
            phase_idx: (payload.phaseIdx as number) ?? 0,
            phase_kind: phaseKind,
            role: (payload.role as 'doer' | 'reviewer') ?? 'doer',
            agent_id: (payload.agent as string) ?? null,
            state:
              event.type === 'phase_start'
                ? 'drafting'
                : event.type === 'phase_done'
                  ? 'submitted'
                  : 'blocked',
            output: (payload.output as string) ?? null,
            cost_usd: 0,
            tokens_in: 0,
            tokens_out: 0,
            started_at: event.ts,
            finished_at:
              event.type === 'phase_done' || event.type === 'phase_failed'
                ? Date.now()
                : null,
          })
          .catch((err: unknown) => {
            chatLogger(chatId).error(
              { err: err instanceof Error ? err.message : String(err) },
              'phaseEvents.create failed',
            );
          }),
      );
    }

    // Persist per-CLI failure events (cli_error / cli_warning) so the
    // cockpit can surface "this reviewer failed with <reason>" on the
    // per-card UI AND post-mortem inspection (sqlite, /chats/:id) shows
    // the failure even after chat-done has fired. Without this, transient
    // subprocess crashes (opencode lock contention, codex quota, gemini
    // rate-limit-with-empty-stdout) wrote 0-byte answer.md files and
    // disappeared from the DB.
    //
    // cli_error ⇒ state='errored', cli_warning ⇒ state='warning'. The
    // replay path (phaseEventToRunnerEvent) ignores both states the same
    // way it always has, so live subscribers are unaffected. Pre-fix every
    // cli_warning landed as state='errored', which made a successful
    // per-slot model fallback look like a reviewer crash in the audit
    // trail.
    if (event.type === 'cli_error' || event.type === 'cli_warning') {
      const payload = event.payload as Record<string, unknown>;
      const kind = payload.phaseKind as string | undefined;
      const phaseKind: PhaseKind =
        kind && (VALID_PHASE_KINDS as readonly string[]).includes(kind)
          ? (kind as PhaseKind)
          : 'review';
      const errorObj = (payload.error as Record<string, unknown> | undefined) ?? {};
      const message =
        (errorObj.message as string | undefined) ??
        (payload.message as string | undefined) ??
        'unknown error';
      const isWarning = event.type === 'cli_warning';
      const persistedState: 'errored' | 'warning' = isWarning ? 'warning' : 'errored';
      const tag =
        (errorObj.kind as string | undefined) ??
        (isWarning ? 'cli_warning' : 'cli_error');
      void trackWrite(
        phaseEvents
          .create({
            chat_id: chatId,
            phase_idx: (payload.phaseIdx as number) ?? 0,
            phase_kind: phaseKind,
            role: (payload.role as 'doer' | 'reviewer') ?? 'reviewer',
            agent_id: (payload.agent as string) ?? null,
            state: persistedState,
            // Pack the failure / warning context into output so the
            // cockpit's existing event-list rendering surfaces the
            // message without a schema change.
            output: `[${tag}] ${message}`,
            cost_usd: 0,
            tokens_in: 0,
            tokens_out: 0,
            started_at: event.ts,
            finished_at: event.ts,
          })
          .catch((err: unknown) => {
            chatLogger(chatId).error(
              { err: err instanceof Error ? err.message : String(err) },
              `phaseEvents.create (${persistedState}) failed`,
            );
          }),
      );
    }

    // Update chats.status on terminal event. Runner emits status='completed'
    // for the happy path; we map to 'approved' to fit the chats.status
    // enum. Tracked so .finally drains before releasing the activeRuns
    // slot — otherwise a reattaching SSE could see no active run + stale
    // 'reviewing' status and start a dup run.
    if (event.type === 'chat_done') {
      const payload = event.payload as Record<string, unknown>;
      const status = (payload.status as string) ?? 'completed';
      // verdict is the reviewer-level outcome (separate from system-level
      // status). Always persist when present so review-only chats with
      // verdict='request_changes' are distinguishable from standard chats
      // whose status='approved' implicitly means verdict='approved'. Cap
      // defensively at 32 chars — verdicts are enum-shaped strings;
      // anything longer is bogus.
      const rawVerdict = payload.verdict;
      const verdict =
        typeof rawVerdict === 'string' && rawVerdict.length > 0 && rawVerdict.length <= 32
          ? rawVerdict
          : null;
      void trackWrite(
        chats
          .update(chatId, {
            status: (status === 'completed' ? 'approved' : status) as ChatStatus,
            ...(verdict !== null ? { verdict } : {}),
            ...(typeof payload.prUrl === 'string' && payload.prUrl.length > 0
              ? { pr_url: payload.prUrl }
              : {}),
            ...(typeof payload.shipError === 'string' && payload.shipError.length > 0
              ? { ship_error: payload.shipError }
              : {}),
            finished_at: Date.now(),
          })
          .catch((err: unknown) => {
            chatLogger(chatId).error(
              { err: err instanceof Error ? err.message : String(err) },
              'chats.update on chat_done failed',
            );
          }),
      );
    }
  };

  const promise = runChat({
    chatId,
    template,
    work: chat.work,
    artifact: chat.artifact ?? undefined,
    repoPath: chat.repo_path ?? undefined,
    attachedFiles: parseAttachedFiles(chat.attached_files),
    abortSignal: abortController.signal,
    tmuxMgr,
    errorDetector,
    onEvent,
  }).finally(async () => {
    // Drain pending DB writes BEFORE releasing the slot. Without this,
    // the chat_done chats.update can still be in flight when a
    // reattaching SSE sees activeRuns empty + reads stale chats row →
    // starts a duplicate run, burns subscription quota, writes duplicate
    // phase events. allSettled so a failed write doesn't leak unhandled
    // rejections — individual .catch handlers above already log.
    if (pendingWrites.size > 0) {
      await Promise.allSettled(Array.from(pendingWrites));
    }
    activeRuns.delete(chatId);
    // Server-initiated subscriber close. The cockpit closes its
    // EventSource on chat_done already, but a misbehaving / disconnected
    // client can leave the subscriber object pinned in the set. Without
    // this sweep the underlying hijacked socket lingers (held open by
    // Fastify's raw.write reference) until the OS TCP keepalive reaps
    // it. close() swallows its own errors.
    for (const sub of Array.from(subscribers)) {
      try {
        sub.close();
      } catch {
        /* dead socket — already closed by the client */
      }
    }
    subscribers.clear();
    // Drop any per-participant abort controllers left over by aborted /
    // crashed runners. They should already have released themselves via
    // their `finally` blocks, but a stale entry would leak across chats
    // if a runner exited abnormally.
    participantAborts.cleanupChat(chatId);
  });

  const entry: ActiveRun = { promise, subscribers, abortController };
  activeRuns.set(chatId, entry);
  return entry;
}

/**
 * Snapshot of all active runs — for graceful shutdown only. Don't use
 * for steady-state route handling; getActiveRun(chatId) is the right
 * accessor for the chat-keyed lookup.
 */
export function activeRunsSnapshot(): ActiveRun[] {
  return Array.from(activeRuns.values());
}

export function activeRunsCount(): number {
  return activeRuns.size;
}
