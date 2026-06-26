/**
 * Atomic JSON write — writes to a temp file in the same dir, then renames
 * over the destination. fs.renameSync is atomic on POSIX (single inode swap),
 * so a reader concurrently opening the destination either sees the previous
 * complete file or the new complete file — never a half-written one.
 *
 * Why this matters: the cockpit polls _meta.json sidecars and chat-level
 * meta.json to render run cards. A daemon crash mid-write (or even a slow
 * write under load) could leave a 0-byte or truncated file that JSON.parse
 * rejects, and the cockpit silently swallows the error and the card never
 * renders. Atomic-rename eliminates that window.
 *
 * The temp file uses a `.tmp.<pid>.<rand>` suffix so concurrent writers in
 * the same dir (multiple participants in a phase) don't collide on the
 * temp path. Rename is the synchronisation point; the OS guarantees only
 * one of N concurrent renames "wins" the destination at any instant.
 *
 * Caller note: we deliberately use sync APIs because the existing call
 * sites are sync (writeTransportMeta in kimi.ts, runner setup). An async
 * variant can be added later if a hot path needs it.
 */
import fs from 'node:fs';
import path from 'node:path';

export function atomicWriteJsonSync(targetPath: string, value: unknown): void {
  const dir = path.dirname(targetPath);
  const base = path.basename(targetPath);
  // pid + 8 hex chars from Date.now() is sufficient; same-process collision
  // is impossible (sequential calls), cross-process needs the pid + nonce.
  const nonce = Math.floor(Math.random() * 0xffffffff).toString(16).padStart(8, '0');
  const tmpPath = path.join(dir, `.${base}.tmp.${process.pid}.${nonce}`);

  // Stringify FIRST so a JSON serialization error doesn't leave an empty
  // tmp file on disk (otherwise the catch-and-cleanup below has more to do).
  const body = JSON.stringify(value, null, 2);

  try {
    fs.writeFileSync(tmpPath, body, 'utf-8');
    fs.renameSync(tmpPath, targetPath);
  } catch (err) {
    // Best-effort cleanup of the temp file if rename failed (e.g. EXDEV
    // on a cross-device tmpdir, EACCES on perms). Re-throw so callers can
    // decide whether to swallow (informational sidecars) or escalate.
    try { fs.unlinkSync(tmpPath); } catch { /* tmp may not exist */ }
    throw err;
  }
}
