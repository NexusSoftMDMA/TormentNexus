import { mkdtempSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { afterEach, describe, expect, it } from 'vitest';

import { resolveSwarmOrchestratorBase } from './orchestrator-base.js';

const originalEnv = {
    TORMENTNEXUS_CONFIG_DIR: process.env.TORMENTNEXUS_CONFIG_DIR,
    TORMENTNEXUS_ORCHESTRATOR_URL: process.env.TORMENTNEXUS_ORCHESTRATOR_URL,
    TORMENTNEXUS_TRPC_UPSTREAM: process.env.TORMENTNEXUS_TRPC_UPSTREAM,
    NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL: process.env.NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL,
    NEXT_PUBLIC_AUTOPILOT_URL: process.env.NEXT_PUBLIC_AUTOPILOT_URL,
};

afterEach(() => {
    for (const [key, value] of Object.entries(originalEnv)) {
        if (value === undefined) {
            delete process.env[key];
        } else {
            process.env[key] = value;
        }
    }
});

describe('resolveSwarmOrchestratorBase', () => {
    it('preserves an explicit swarm base override', () => {
        expect(resolveSwarmOrchestratorBase('http://127.0.0.1:5001/')).toBe('http://127.0.0.1:5001');
    });

    it('reuses the live tormentnexus lock when no explicit override is supplied', () => {
        const configDir = mkdtempSync(path.join(os.tmpdir(), 'swarm-lock-'));
        process.env.TORMENTNEXUS_CONFIG_DIR = configDir;
        writeFileSync(path.join(configDir, 'lock'), JSON.stringify({ host: '0.0.0.0', port: 4321 }));

        expect(resolveSwarmOrchestratorBase()).toBe('http://127.0.0.1:4321');
    });

    it('returns null when no orchestrator base is configured', () => {
        const configDir = mkdtempSync(path.join(os.tmpdir(), 'swarm-lock-empty-'));
        process.env.TORMENTNEXUS_CONFIG_DIR = configDir;
        delete process.env.TORMENTNEXUS_ORCHESTRATOR_URL;
        delete process.env.TORMENTNEXUS_TRPC_UPSTREAM;
        delete process.env.NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL;
        delete process.env.NEXT_PUBLIC_AUTOPILOT_URL;

        expect(resolveSwarmOrchestratorBase()).toBeNull();
    });
});
