/**
 * Buffered append-writer for live LLM streaming.
 *
 * Why this exists: the previous implementation called fs.appendFileSync on
 * every text_delta event from the CLI shim. A typical Opus run emits ~5
 * deltas/sec for ~5 minutes = ~1500 sync disk writes per doer, each one
 * blocking the daemon event loop. Worse, the cockpit polls answer.md every
 * 8s — it doesn't benefit from per-delta granularity.
 *
 * Now we flush whenever the buffer crosses 4KB OR a 750ms quiet timer
 * fires, whichever comes first. Cockpit poll cadence is unchanged but
 * the daemon does ~50 writes per run instead of ~1500.
 *
 * Caller MUST call flushNow() before any direct read of `path` or before
 * fs.writeFileSync overwrites it — the buffer is otherwise lost.
 *
 * Failure semantics: if the underlying appendFileSync ever throws (FS
 * ENOSPC, EACCES, etc.) the writer flips to a "dead" state and surfaces
 * the cause via lastError(). Subsequent write() calls become no-ops; the
 * buffered chunk that failed is dropped (re-trying would just fail again
 * synchronously and leak more memory). The runner inspects isDead() /
 * lastError() in its finally block to surface a cli_error event so the
 * user knows their answer.md is partial — silent loss was the previous
 * behavior and round-2 review flagged it.
 */
import * as fs from 'fs';

export class StreamFileWriter {
  private buf = '';
  private flushTimer: NodeJS.Timeout | null = null;
  private dead = false;
  private lastErr: Error | null = null;

  constructor(
    private readonly filePath: string,
    private readonly flushBytes = 4096,
    private readonly flushMs = 750,
  ) {}

  write(chunk: string): void {
    if (!chunk || this.dead) return;
    this.buf += chunk;
    if (this.buf.length >= this.flushBytes) {
      this.flushNow();
    } else if (this.flushTimer === null) {
      this.flushTimer = setTimeout(() => this.flushNow(), this.flushMs);
    }
  }

  flushNow(): void {
    if (this.flushTimer !== null) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }
    if (this.buf.length === 0 || this.dead) return;
    try {
      fs.appendFileSync(this.filePath, this.buf);
      this.buf = '';
    } catch (err) {
      // Permanent failure — flip dead, drop the in-flight buffer to free
      // memory, surface via lastError() for the runner to forward as a
      // cli_error so the user sees something instead of a quietly stale
      // answer.md.
      this.dead = true;
      this.lastErr = err instanceof Error ? err : new Error(String(err));
      this.buf = '';
    }
  }

  isDead(): boolean {
    return this.dead;
  }

  lastError(): Error | null {
    return this.lastErr;
  }
}
