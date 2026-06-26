import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { describe, expect, it } from 'vitest';

import { resolveTormentNexusConfigDir, resolveTormentNexusLockPath, resolveLockedTormentNexusBase, resolveOrchestratorBase } from './tormentnexus-orchestrator.js';

describe('tormentnexus orchestrator helpers', () => {
    it('uses explicit env bases before lock-derived values', () => {
        expect(resolveOrchestratorBase({
            TORMENTNEXUS_ORCHESTRATOR_URL: 'http://127.0.0.1:4100/',
            TORMENTNEXUS_TRPC_UPSTREAM: 'http://127.0.0.1:4200/trpc',
        })).toBe('http://127.0.0.1:4100');

        expect(resolveOrchestratorBase({
            TORMENTNEXUS_TRPC_UPSTREAM: 'http://127.0.0.1:4200/trpc/',
        })).toBe('http://127.0.0.1:4200');
    });

    it('resolves the live lock-file base before public env fallbacks', () => {
        const configDir = mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-lock-'));
        writeFileSync(path.join(configDir, 'lock'), JSON.stringify({ host: '0.0.0.0', port: 4312 }));

        expect(resolveTormentNexusConfigDir({ TORMENTNEXUS_CONFIG_DIR: configDir })).toBe(configDir);
        expect(resolveTormentNexusLockPath({ TORMENTNEXUS_CONFIG_DIR: configDir })).toBe(path.join(configDir, 'lock'));
        expect(resolveLockedTormentNexusBase({ TORMENTNEXUS_CONFIG_DIR: configDir })).toBe('http://127.0.0.1:4312');
        expect(resolveOrchestratorBase({
            TORMENTNEXUS_CONFIG_DIR: configDir,
            NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL: 'http://127.0.0.1:3847',
        })).toBe('http://127.0.0.1:4312');
    });

    it('falls back to configured public envs when no live lock exists', () => {
        const configDir = mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-lock-empty-'));
        mkdirSync(configDir, { recursive: true });

        expect(resolveOrchestratorBase({
            TORMENTNEXUS_CONFIG_DIR: configDir,
            NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL: 'http://127.0.0.1:5001/',
        })).toBe('http://127.0.0.1:5001');

        expect(resolveOrchestratorBase({
            TORMENTNEXUS_CONFIG_DIR: configDir,
            NEXT_PUBLIC_AUTOPILOT_URL: 'http://127.0.0.1:3847',
        })).toBe('http://127.0.0.1:3847');
    });
});
