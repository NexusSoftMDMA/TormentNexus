/**
 * DB seam — barrel re-exports per-table modules.
 *
 * Connection lifecycle (getDb, _resetDbForTests, resolveDbPath) lives in
 * connection.ts; each table has its own file with schema + ops.
 *
 * Rollback lever for the libsql migration: the v0.7 swap from
 * better-sqlite3 was a clean transport change. If a hot-path perf
 * regression turns up in production, the rollback is a clean revert
 * (NOT a swap to the sync `libsql` package — its API would require
 * unwinding every `await` in this layer and its callers).
 */

export { _resetDbForTests, getDb, resolveDbPath } from './connection.js';
export { chats } from './chats.js';
export { phaseEvents } from './phase-events.js';
export { templates } from './templates.js';
export { settings } from './settings.js';
export { secrets } from './secrets.js';
export { personas, type PersonaRow } from './personas.js';
export { voices, type VoiceUpsertInput } from './voices.js';
