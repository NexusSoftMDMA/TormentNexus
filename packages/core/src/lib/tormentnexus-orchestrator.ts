import { existsSync, readFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

type TormentNexusLockRecord = {
    port?: number;
    host?: string;
};

type OrchestratorEnv = Record<string, string | undefined>;

function normalizeBaseURL(value?: string): string | null {
    const trimmed = value?.trim();
    if (!trimmed) {
        return null;
    }

    const withoutTrailingSlash = trimmed.replace(/\/$/, '');
    return withoutTrailingSlash.endsWith('/trpc')
        ? withoutTrailingSlash.slice(0, -5)
        : withoutTrailingSlash;
}

function resolveBrowserHost(host: string): string {
    return host === '0.0.0.0' || host === '::' || host === '[::]'
        ? '127.0.0.1'
        : host;
}

export function resolveTormentNexusConfigDir(env: OrchestratorEnv = process.env): string {
    const configuredDir = env.TORMENTNEXUS_CONFIG_DIR?.trim();
    if (configuredDir) {
        return configuredDir;
    }

    return path.join(os.homedir(), '.tormentnexus');
}

export function resolveTormentNexusLockPath(env: OrchestratorEnv = process.env): string {
    return path.join(resolveTormentNexusConfigDir(env), 'lock');
}

export function resolveLockedTormentNexusBase(env: OrchestratorEnv = process.env): string | null {
    const lockPath = resolveTormentNexusLockPath(env);
    if (!existsSync(lockPath)) {
        return null;
    }

    try {
        const parsed = JSON.parse(readFileSync(lockPath, 'utf8')) as TormentNexusLockRecord;
        if (!parsed || typeof parsed.port !== 'number' || parsed.port <= 0) {
            return null;
        }

        const host = typeof parsed.host === 'string' && parsed.host.trim().length > 0
            ? resolveBrowserHost(parsed.host.trim())
            : '127.0.0.1';

        return `http://${host}:${parsed.port}`;
    } catch {
        return null;
    }
}

export function resolveOrchestratorBase(env: OrchestratorEnv = process.env): string | null {
    return normalizeBaseURL(env.TORMENTNEXUS_ORCHESTRATOR_URL)
        ?? normalizeBaseURL(env.TORMENTNEXUS_TRPC_UPSTREAM)
        ?? resolveLockedTormentNexusBase(env)
        ?? normalizeBaseURL(env.NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL)
        ?? normalizeBaseURL(env.NEXT_PUBLIC_AUTOPILOT_URL);
}
