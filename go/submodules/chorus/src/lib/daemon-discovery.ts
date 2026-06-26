import fs from 'fs';
import os from 'os';
import path from 'path';
import { atomicWriteJsonSync } from './atomic-write';

/**
 * Runtime-port discovery for chorus daemon + cockpit.
 *
 * Pre-v0.8 every consumer hardcoded `http://127.0.0.1:7707` (daemon)
 * and `http://127.0.0.1:5050` (cockpit). When those ports were squatted
 * (VSCode tunnel pre-forwards in WSL, lingering processes, sudo'd
 * zombies), `chorus start` failed with a generic "port in use" error
 * and there was no safe way to shift — MCP clients would still hit the
 * squatter on the original port.
 *
 * v0.8 decouples runtime ports from literals via `~/.chorus/daemon.json`.
 * `chorus start` walks for free ports, writes the chosen pair, and
 * every consumer reads from this file at startup.
 *
 * Resolution order (in order of precedence):
 *   1. Live `daemon.json` whose recorded port answers /api/v1/health.
 *   2. `CHORUS_DAEMON_URL` env var (remote-daemon override).
 *   3. Hardcoded fallback `http://127.0.0.1:7707` (v0.7 compat).
 *
 * Same idea for the cockpit, with `CHORUS_COCKPIT_URL` as the env
 * override (rarely set; mostly useful for dev with custom ports).
 */

export const DEFAULT_DAEMON_PORT = 7707;
export const DEFAULT_COCKPIT_PORT = 5050;
export const DEFAULT_DAEMON_URL = `http://127.0.0.1:${DEFAULT_DAEMON_PORT}`;
export const DEFAULT_COCKPIT_URL = `http://127.0.0.1:${DEFAULT_COCKPIT_PORT}`;

export const DAEMON_PORT_RANGE = 14;
export const COCKPIT_PORT_RANGE = 14;

export interface DaemonInfo {
  schemaVersion: 1;
  daemonPort: number;
  cockpitPort: number;
  daemonPid: number;
  cockpitPid: number | null;
  startedAt: string;
  version: string;
}

export function daemonInfoPath(): string {
  return path.join(os.homedir(), '.chorus', 'daemon.json');
}

/**
 * Atomic write via the shared helper. Caller is responsible for ensuring
 * the directory exists; we mkdirSync defensively.
 */
export function writeDaemonInfo(info: DaemonInfo): void {
  const target = daemonInfoPath();
  fs.mkdirSync(path.dirname(target), { recursive: true });
  atomicWriteJsonSync(target, info);
}

/**
 * Best-effort delete. Called on `chorus stop` to avoid leaving a stale
 * file that the next `chorus start` would have to recover from.
 */
export function clearDaemonInfo(): void {
  try {
    fs.unlinkSync(daemonInfoPath());
  } catch {
    /* missing is fine */
  }
}

/**
 * Read + validate daemon.json. Returns null on missing file, parse
 * failure, or schema mismatch — caller treats null as "no recorded
 * daemon, use defaults".
 *
 * Does NOT do liveness checks here; that's the caller's job via
 * `isDaemonAlive`. Separating read-from-disk from probe-the-port keeps
 * the synchronous read path cheap.
 */
export function readDaemonInfo(): DaemonInfo | null {
  let raw: string;
  try {
    raw = fs.readFileSync(daemonInfoPath(), 'utf-8');
  } catch {
    return null;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!parsed || typeof parsed !== 'object') return null;
  const obj = parsed as Record<string, unknown>;
  if (obj.schemaVersion !== 1) return null;
  if (typeof obj.daemonPort !== 'number' || typeof obj.cockpitPort !== 'number') {
    return null;
  }
  if (typeof obj.daemonPid !== 'number') return null;
  return {
    schemaVersion: 1,
    daemonPort: obj.daemonPort,
    cockpitPort: obj.cockpitPort,
    daemonPid: obj.daemonPid,
    cockpitPid:
      typeof obj.cockpitPid === 'number' ? obj.cockpitPid : null,
    startedAt: typeof obj.startedAt === 'string' ? obj.startedAt : '',
    version: typeof obj.version === 'string' ? obj.version : '',
  };
}

/**
 * PID-alive check. Cross-platform via `process.kill(pid, 0)` — throws
 * on dead PIDs, returns true on EPERM (process exists, owned by another
 * uid). Cheap pre-filter before the slower health probe.
 */
export function isPidAlive(pid: number): boolean {
  if (!Number.isFinite(pid) || pid <= 0) return false;
  try {
    process.kill(pid, 0);
    return true;
  } catch (err) {
    // EPERM = exists but not signalable from us → still alive.
    return (err as NodeJS.ErrnoException).code === 'EPERM';
  }
}

/**
 * Probe `/api/v1/health` on the recorded daemon port. Used for stale-
 * file detection and to reject foreign processes that happen to bind
 * the same port. The chorus health envelope is `{ ok: true, data: { version } }` —
 * very specific shape so we can't be fooled by a random other server.
 */
export async function isDaemonHealthy(
  port: number,
  timeoutMs = 1000,
): Promise<boolean> {
  try {
    const ac = new AbortController();
    const timer = setTimeout(() => ac.abort(), timeoutMs);
    const res = await fetch(`http://127.0.0.1:${port}/api/v1/health`, {
      signal: ac.signal,
    });
    clearTimeout(timer);
    if (!res.ok) return false;
    const envelope = (await res.json()) as {
      ok?: boolean;
      data?: { version?: string };
    };
    return envelope.ok === true && typeof envelope.data?.version === 'string';
  } catch {
    return false;
  }
}

/**
 * PID-alive pre-check + health probe. Returns the recorded info iff
 * both pass; null otherwise. Used by every read-side consumer to make
 * "is the recorded daemon really running?" a single call.
 */
export async function readLiveDaemonInfo(
  options: { healthTimeoutMs?: number } = {},
): Promise<DaemonInfo | null> {
  const info = readDaemonInfo();
  if (!info) return null;
  if (!isPidAlive(info.daemonPid)) return null;
  const healthy = await isDaemonHealthy(
    info.daemonPort,
    options.healthTimeoutMs ?? 1000,
  );
  return healthy ? info : null;
}

/**
 * Public resolver — what URL should the next daemon HTTP call go to?
 *
 * Order:
 *   1. Live local daemon (daemon.json with PID alive + health pass).
 *      Note: env var is bypassed when a local daemon is alive, because
 *      old MCP configs may have CHORUS_DAEMON_URL=http://127.0.0.1:7707
 *      hardcoded; honoring that would defeat the entire port-shift fix.
 *   2. `CHORUS_DAEMON_URL` env var (for remote-daemon use cases — when
 *      no local daemon.json exists, the env var is the explicit override).
 *   3. Hardcoded default (v0.7 back-compat for first-ever installs).
 */
export async function resolveDaemonUrl(): Promise<string> {
  const live = await readLiveDaemonInfo();
  if (live) return `http://127.0.0.1:${live.daemonPort}`;
  if (process.env.CHORUS_DAEMON_URL) return process.env.CHORUS_DAEMON_URL;
  return DEFAULT_DAEMON_URL;
}

/**
 * Cockpit URL counterpart. The cockpit's port is recorded alongside
 * the daemon in daemon.json, so resolution is the same shape.
 */
export async function resolveCockpitUrl(): Promise<string> {
  const live = await readLiveDaemonInfo();
  if (live) return `http://127.0.0.1:${live.cockpitPort}`;
  if (process.env.CHORUS_COCKPIT_URL) return process.env.CHORUS_COCKPIT_URL;
  return DEFAULT_COCKPIT_URL;
}

/**
 * Walk a port range looking for one that is NOT in use AND not held by
 * a healthy chorus daemon. Returns the first usable port or null if the
 * full range is taken.
 *
 * `isInUse` is injected so callers can use the existing `isPortInUse`
 * helper from cli/port-utils without forming a circular import.
 */
export async function pickFreePort(
  preferredPort: number,
  range: number,
  isInUse: (port: number) => Promise<boolean>,
): Promise<number | null> {
  for (let i = 0; i < range; i++) {
    const port = preferredPort + i;
    if (!(await isInUse(port))) return port;
  }
  return null;
}
