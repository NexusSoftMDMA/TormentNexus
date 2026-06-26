import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { spawn, type ChildProcess } from 'child_process';

import { resolveCliEntryPath, resolveMonorepoRoot } from './orchestratorPaths.js';

export interface BackgroundCoreBootstrapOptions {
    healthUrl?: string;
    host?: string;
    startupTimeoutMs?: number;
    pollIntervalMs?: number;
    waitForReady?: boolean;
    cliEntryPath?: string | null;
    log?: (message?: unknown, ...optionalParams: unknown[]) => void;
}

export interface BackgroundCoreBootstrapResult {
    status: 'already-running' | 'spawned' | 'warming' | 'launch-unavailable';
    pid?: number;
    cliEntryPath?: string;
}

type FetchLike = typeof globalThis.fetch;
type SpawnLike = (
    command: string,
    args: ReadonlyArray<string>,
    options: {
        detached?: boolean;
        stdio?: 'ignore';
        windowsHide?: boolean;
    },
) => Pick<ChildProcess, 'pid' | 'unref'>;

interface BackgroundCoreBootstrapDeps {
    fetchImpl?: FetchLike;
    spawnImpl?: SpawnLike;
    waitImpl?: (ms: number) => Promise<void>;
}

const DEFAULT_HEALTH_URL = 'http://127.0.0.1:3001/health';
const DEFAULT_STARTUP_TIMEOUT_MS = 15_000;
const DEFAULT_POLL_INTERVAL_MS = 500;

function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function isCoreBridgeHealthy(
    healthUrl: string = DEFAULT_HEALTH_URL,
    fetchImpl: FetchLike = globalThis.fetch,
): Promise<boolean> {
    try {
        const response = await fetchImpl(healthUrl, { method: 'GET' });
        return response.ok;
    } catch {
        return false;
    }
}

export async function waitForCoreBridge(
    options: {
        healthUrl?: string;
        timeoutMs?: number;
        pollIntervalMs?: number;
    } = {},
    deps: Pick<BackgroundCoreBootstrapDeps, 'fetchImpl' | 'waitImpl'> = {},
): Promise<boolean> {
    const healthUrl = options.healthUrl ?? DEFAULT_HEALTH_URL;
    const timeoutMs = options.timeoutMs ?? DEFAULT_STARTUP_TIMEOUT_MS;
    const pollIntervalMs = options.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS;
    const fetchImpl = deps.fetchImpl ?? globalThis.fetch;
    const waitImpl = deps.waitImpl ?? delay;
    const deadline = Date.now() + timeoutMs;

    do {
        if (await isCoreBridgeHealthy(healthUrl, fetchImpl)) {
            return true;
        }

        if (Date.now() >= deadline) {
            break;
        }

        await waitImpl(pollIntervalMs);
    } while (true);

    return false;
}

export async function ensureDashboardRunning(log: (msg?: unknown) => void = () => undefined): Promise<void> {
    const dashboardPort = 3000;
    const dashboardUrl = `http://127.0.0.1:${dashboardPort}/dashboard`;

    try {
        const res = await fetch(dashboardUrl, { method: 'GET', signal: AbortSignal.timeout(1000) });
        if (res.ok) {
            return; // Already healthy
        }
    } catch {}

    const root = resolveMonorepoRoot(process.cwd()) || resolveMonorepoRoot(path.dirname(fileURLToPath(import.meta.url)));
    if (!root) {
        log('[TormentNexus Dashboard] Monorepo root not found. Skipping auto-start.');
        return;
    }

    const webDir = path.join(root, 'apps', 'web');
    const startScript = path.join(webDir, 'scripts', 'start.mjs');
    const devScript = path.join(webDir, 'scripts', 'dev.mjs');
    const standaloneServer = path.join(webDir, '.next', 'standalone', 'apps', 'web', 'server.js');

    const hasStandalone = fs.existsSync(standaloneServer);
    const scriptToRun = hasStandalone ? startScript : devScript;

    if (!fs.existsSync(scriptToRun)) {
        log(`[TormentNexus Dashboard] Start script not found at ${scriptToRun}. Skipping auto-start.`);
        return;
    }

    log(`[TormentNexus Dashboard] Lazy spawning dashboard server via ${path.basename(scriptToRun)} on port ${dashboardPort}...`);
    try {
        const child = spawn(process.execPath, [scriptToRun, '--port', String(dashboardPort), '--host', '127.0.0.1'], {
            detached: true,
            stdio: 'ignore',
            windowsHide: true,
            cwd: webDir,
            env: {
                ...process.env,
                TORMENTNEXUS_TRPC_UPSTREAM: 'http://127.0.0.1:4100/trpc',
            }
        });
        child.unref?.();
        log(`[TormentNexus Dashboard] Dashboard background server spawned successfully (PID: ${child.pid}).`);
    } catch (e: any) {
        log(`[TormentNexus Dashboard] Failed to spawn dashboard server: ${e.message}`);
    }
}

export async function ensureBackgroundCoreRunning(
    options: BackgroundCoreBootstrapOptions = {},
    deps: BackgroundCoreBootstrapDeps = {},
): Promise<BackgroundCoreBootstrapResult> {
    const healthUrl = options.healthUrl ?? DEFAULT_HEALTH_URL;
    const log = options.log ?? (() => undefined);
    const fetchImpl = deps.fetchImpl ?? globalThis.fetch;
    const waitImpl = deps.waitImpl ?? delay;

    if (await isCoreBridgeHealthy(healthUrl, fetchImpl)) {
        void ensureDashboardRunning(log);
        return { status: 'already-running' };
    }

    const cliEntryPath = Object.prototype.hasOwnProperty.call(options, 'cliEntryPath')
        ? (options.cliEntryPath ?? null)
        : resolveCliEntryPath();
    if (!cliEntryPath) {
        log('[TormentNexus Core] Background core bootstrap skipped: CLI entrypoint not found.');
        return { status: 'launch-unavailable' };
    }

    const spawnImpl = deps.spawnImpl ?? spawn;
    const args = [cliEntryPath, 'start'];

    if (options.host) {
        args.push('--host', options.host);
    }

    const child = spawnImpl(process.execPath, args, {
        detached: true,
        stdio: 'ignore',
        windowsHide: true,
    });
    child.unref?.();

    // Trigger lazy dashboard boot in parallel
    void ensureDashboardRunning(log);

    if (options.waitForReady === false) {
        return {
            status: 'spawned',
            pid: child.pid,
            cliEntryPath,
        };
    }

    const ready = await waitForCoreBridge(
        {
            healthUrl,
            timeoutMs: options.startupTimeoutMs ?? DEFAULT_STARTUP_TIMEOUT_MS,
            pollIntervalMs: options.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS,
        },
        {
            fetchImpl,
            waitImpl,
        },
    );

    return {
        status: ready ? 'spawned' : 'warming',
        pid: child.pid,
        cliEntryPath,
    };
}