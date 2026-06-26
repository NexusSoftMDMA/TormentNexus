/**
 * Wire envelope shared by every daemon HTTP response.
 *
 * Why a uniform envelope: the cockpit's fetchFromDaemon helper inspects
 * `ok` and either returns `data` or throws a DaemonError carrying
 * `error.code` + `error.message`. Keeping these types in one place lets
 * the route modules and the chats-flow code in index.ts stay aligned
 * without each importing fastify or redefining the shape inline.
 */

import type { FastifyReply } from 'fastify';

/**
 * Canonical error codes emitted by daemon route handlers.
 *
 * Adding a code: bump this list AND make sure the cockpit's DaemonError
 * doesn't try to stringly-switch on it (it doesn't today). MCP consumers
 * read `error.message` not `error.code`, so adding codes is safe; renaming
 * an existing code is a wire break.
 *
 * Source-of-truth audit done 2026-05-05 (round-2 review of the API
 * shape-freeze plan): grep of `src/daemon/routes/*.ts` for
 * `errorResponse(...)` first arg confirmed these 11 are actually in use.
 * `unknown` lives in `src/lib/api/client.ts` as a CLIENT-side fallback
 * only and is intentionally not in this enum.
 */
export type ErrorCode =
  | 'validation' // request shape / required-field missing
  | 'not_found' // resource doesn't exist
  | 'conflict' // resource exists but state forbids the op (e.g. cancel a finished chat)
  | 'bad_request' // request was structurally fine but semantically rejected
  | 'parse_error' // body was not valid JSON / SSE line was malformed
  | 'connection_failed' // daemon could not reach an upstream (cli, OpenRouter HTTP)
  | 'cli_failed' // CLI exited non-zero (codex, gemini, opencode)
  | 'db_error' // SQLite / libsql write failed
  | 'stats_error' // /stats aggregation failed
  | 'openrouter_error' // OpenRouter HTTP shim returned a non-2xx
  | 'internal'; // unexpected daemon-side exception

export interface ErrorResponse {
  ok: false;
  error: {
    code: ErrorCode | string; // accept legacy/unknown for forward-compat
    message: string;
    /**
     * Optional structured payload. Used today for zod-issue lists from
     * /templates POST (`{issues: [{path, message}, ...]}`) so the cockpit
     * can pin each error to the field it references. Kept loose so future
     * routes can attach their own structured detail without changing the
     * envelope shape.
     */
    details?: Record<string, unknown>;
  };
}

export interface SuccessResponse<T> {
  ok: true;
  data: T;
}

export type ApiResponse<T> = SuccessResponse<T> | ErrorResponse;

export function errorResponse(
  code: ErrorCode | string,
  message: string,
  details?: Record<string, unknown>,
): ErrorResponse {
  return {
    ok: false,
    error: details ? { code, message, details } : { code, message },
  };
}

export function successResponse<T>(data: T): SuccessResponse<T> {
  return {
    ok: true,
    data,
  };
}

/**
 * Envelope for list endpoints. Locked in as `{items, total, hasMore}`
 * pre-launch so that adding limit/offset query params later is a
 * non-breaking change. `total = items.length` and `hasMore = false`
 * today — the envelope alone freezes the shape.
 *
 * Caller wraps this inside `successResponse(...)` so the wire shape is
 * `{ ok: true, data: { items, total, hasMore } }`.
 */
export interface ListEnvelope<T> {
  items: T[];
  total: number;
  hasMore: boolean;
}

export function listEnvelope<T>(items: T[], hasMore = false): ListEnvelope<T> {
  return { items, total: items.length, hasMore };
}

/**
 * Default HTTP status per ErrorCode. Routes can override but should
 * very rarely need to — these match standard REST semantics.
 */
const DEFAULT_STATUS: Record<string, number> = {
  validation: 400,
  bad_request: 400,
  parse_error: 400,
  not_found: 404,
  conflict: 409,
  cli_failed: 502,
  connection_failed: 502,
  openrouter_error: 502,
  db_error: 500,
  stats_error: 500,
  internal: 500,
};

/**
 * Send an error response with the right HTTP status code AND the
 * canonical envelope. Replaces the older pattern of returning the body
 * from `errorResponse(...)` without setting `reply.code`, which left
 * 4xx-shaped errors arriving as HTTP 200.
 *
 * Use this in route handlers from now on. Existing `errorResponse(...)`
 * call-sites that still bare-return the body keep working (HTTP 200 +
 * `ok: false`) until migrated; cockpit + MCP both read the envelope, not
 * the status. New code should use `sendError` so we can drop the legacy
 * pattern in v0.8.
 */
export function sendError(
  reply: FastifyReply,
  code: ErrorCode,
  message: string,
  details?: Record<string, unknown>,
  statusOverride?: number,
): ErrorResponse {
  const status = statusOverride ?? DEFAULT_STATUS[code] ?? 500;
  reply.code(status);
  return errorResponse(code, message, details);
}
