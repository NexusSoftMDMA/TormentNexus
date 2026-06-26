/**
 * Structured logging substrate (round-2-deferred §3).
 *
 * Every log line is one JSON object on stdout. The wire shape mirrors
 * pino's defaults so a future swap to pino is transparent for any
 * downstream consumer (`chorus logs --chat <id>`, log aggregators, etc):
 *
 *   {"level":30,"time":1714654200000,"pid":1234,"hostname":"box","chatId":"abc","msg":"phase done"}
 *
 * - `level` is the numeric pino value (debug=20, info=30, warn=40, error=50)
 * - `time` is ms since epoch (not `ts`)
 * - `pid` + `hostname` are emitted once per line, captured at logger-init
 *
 * Child loggers carry their context fields — call sites in the runner do
 *   `const log = chatLogger(chatId, { phase: 'review', role: 'reviewer' })`
 *   log.info({lineage:'openai'}, 'reviewer started')
 * and the chatId / phase / role get baked into every line that logger
 * emits without each call site repeating them.
 *
 * No runtime dependency. The default is `console.log(JSON.stringify(...))` —
 * if a future load profile justifies it, swap to pino without changing
 * call sites OR the wire shape.
 *
 * `CHORUS_LOG_LEVEL` env var gates the floor (debug | info | warn | error).
 * Default is info. Setting it to debug surfaces verbose internals; setting
 * it to error silences phase-progress noise during dogfood streaming.
 *
 * Round-1 dogfood (PR #7) caught:
 *   - ts/level were clobberable by user fields → core fields now win
 *   - JSON.stringify could throw on circular refs / BigInt → safe wrapper
 *     emits a degraded fallback line instead of crashing the daemon
 *   - Error objects serialized to `{}` → expanded to {message,name,stack}
 *
 * Test seam: pass `_writer` when constructing — tests inject a function
 * that pushes lines into an array and assert exact JSON shape. Production
 * call sites use the default `console.log` chain.
 */

import { hostname } from 'os';

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/** Pino-compatible numeric levels. Wire shape matches pino's default. */
const LEVEL_RANK: Record<LogLevel, number> = {
  debug: 20,
  info: 30,
  warn: 40,
  error: 50,
};

export interface LogFields {
  /** Primary correlation key — every chat-scoped log line carries this. */
  chatId?: string;
  /** Phase id within the template (review, plan, ...). */
  phase?: string;
  /** doer | reviewer — which role this line is about. */
  role?: 'doer' | 'reviewer';
  /** Friendly agent name (claude-code, codex-cli, gemini-cli, ...). */
  agent?: string;
  /** Provider lineage (anthropic | openai | google | opencode | moonshot). */
  lineage?: string;
  /** HTTP-request correlation key — set on routes / cockpit hits. */
  requestId?: string;
  /** Free-form structured tags. */
  [extra: string]: unknown;
}

export interface Logger {
  debug(fields: LogFields | string, msg?: string): void;
  info(fields: LogFields | string, msg?: string): void;
  warn(fields: LogFields | string, msg?: string): void;
  error(fields: LogFields | string, msg?: string): void;
  /** Returns a new logger that bakes `bound` into every emitted line. */
  child(bound: LogFields): Logger;
}

type Writer = (line: string) => void;

const defaultWriter: Writer = (line) => {
  // process.stdout.write is sync + line-buffered when stdout is a pipe (the
  // typical daemon shape). console.log adds a trailing newline that we
  // want; it's also what existing chorus call sites use, so output ordering
  // stays the same when sites migrate.
  console.log(line);
};

/** Captured once at module load. Daemon doesn't move between machines. */
const HOST = (() => {
  try { return hostname(); } catch { return 'unknown'; }
})();
const PID = process.pid;

/**
 * Replacer for JSON.stringify that:
 *   - Expands `Error` instances to {message, name, stack, cause?}
 *   - Defangs BigInt to its string form (JSON.stringify throws on raw BigInt)
 *   - Detects circular references and returns the literal string
 *     '[Circular]' rather than throwing.
 *
 * Falls through to the default stringify behaviour for everything else.
 */
function buildSafeReplacer(): (key: string, value: unknown) => unknown {
  const seen = new WeakSet<object>();
  return function safe(_key: string, value: unknown): unknown {
    if (typeof value === 'bigint') return value.toString();
    if (value instanceof Error) {
      const out: Record<string, unknown> = {
        message: value.message,
        name: value.name,
        stack: value.stack,
      };
      const cause = (value as { cause?: unknown }).cause;
      if (cause !== undefined) out.cause = cause;
      return out;
    }
    if (value !== null && typeof value === 'object') {
      if (seen.has(value as object)) return '[Circular]';
      seen.add(value as object);
    }
    return value;
  };
}

/**
 * Stringify a log line without ever throwing. The replacer handles
 * Error / BigInt / circular refs; if it still fails (toJSON that
 * throws, exotic proxies, etc.) we emit a degraded fallback line
 * referencing the failure rather than crashing the daemon's logger.
 */
function safeStringify(line: Record<string, unknown>): string {
  try {
    return JSON.stringify(line, buildSafeReplacer());
  } catch (err) {
    return JSON.stringify({
      level: line.level,
      time: line.time,
      pid: line.pid,
      hostname: line.hostname,
      msg: 'log_serialize_failed',
      err: err instanceof Error ? err.message : String(err),
    });
  }
}

class LoggerImpl implements Logger {
  constructor(
    private readonly bound: LogFields,
    private readonly threshold: number,
    private readonly writer: Writer,
  ) {}

  private emit(level: LogLevel, fieldsOrMsg: LogFields | string, msg?: string): void {
    if (LEVEL_RANK[level] < this.threshold) return;

    const isFields = typeof fieldsOrMsg === 'object' && fieldsOrMsg !== null;
    const fields = isFields ? fieldsOrMsg : {};
    const message = isFields ? msg : fieldsOrMsg;

    // Spread order: user-provided spreads FIRST, then core wire fields
    // (level, time, pid, hostname). User fields cannot clobber the core
    // shape — round-1 dogfood caught a `log.info({level:'fake'})` foot-gun.
    const line: Record<string, unknown> = {
      ...this.bound,
      ...fields,
      ...(message !== undefined ? { msg: message } : {}),
      level: LEVEL_RANK[level],
      time: Date.now(),
      pid: PID,
      hostname: HOST,
    };

    this.writer(safeStringify(line));
  }

  debug(fields: LogFields | string, msg?: string): void { this.emit('debug', fields, msg); }
  info(fields: LogFields | string, msg?: string): void { this.emit('info', fields, msg); }
  warn(fields: LogFields | string, msg?: string): void { this.emit('warn', fields, msg); }
  error(fields: LogFields | string, msg?: string): void { this.emit('error', fields, msg); }

  child(bound: LogFields): Logger {
    return new LoggerImpl({ ...this.bound, ...bound }, this.threshold, this.writer);
  }
}

function resolveThreshold(): number {
  const raw = (process.env.CHORUS_LOG_LEVEL ?? '').toLowerCase();
  if (raw === 'debug' || raw === 'info' || raw === 'warn' || raw === 'error') {
    return LEVEL_RANK[raw];
  }
  return LEVEL_RANK.info;
}

/**
 * Build the root logger. The daemon constructs ONE root and passes child
 * loggers down to call sites. Tests construct their own with an in-memory
 * writer.
 */
export function createLogger(opts?: {
  /** Test seam: capture lines instead of writing to stdout. */
  _writer?: Writer;
  /** Override the env-resolved threshold. */
  _level?: LogLevel;
}): Logger {
  const writer = opts?._writer ?? defaultWriter;
  const threshold = opts?._level ? LEVEL_RANK[opts._level] : resolveThreshold();
  return new LoggerImpl({}, threshold, writer);
}

/**
 * Module-singleton root logger for general-purpose call sites that don't
 * have a chat / request context. Prefer `chatLogger(chatId)` when the
 * chat correlation key is known.
 */
export const logger: Logger = createLogger();

/** Convenience: child logger pre-bound to a chat correlation key. */
export function chatLogger(chatId: string, extra?: LogFields): Logger {
  return logger.child({ chatId, ...extra });
}
