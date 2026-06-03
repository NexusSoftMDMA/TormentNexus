import { mkdtempSync, mkdirSync } from 'node:fs';
import os from 'node:os';

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { sessionExportRouter } from './sessionExportRouter.js';

function createCaller() {
    return sessionExportRouter.createCaller({} as never);
}

describe('sessionExportRouter orchestrator base resolution', () => {
    const originalFetch = globalThis.fetch;
    const originalEnv = {
        TORMENTNEXUS_ORCHESTRATOR_URL: process.env.TORMENTNEXUS_ORCHESTRATOR_URL,
        TORMENTNEXUS_TRPC_UPSTREAM: process.env.TORMENTNEXUS_TRPC_UPSTREAM,
        TORMENTNEXUS_CONFIG_DIR: process.env.TORMENTNEXUS_CONFIG_DIR,
        NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL: process.env.NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL,
        NEXT_PUBLIC_AUTOPILOT_URL: process.env.NEXT_PUBLIC_AUTOPILOT_URL,
    };

    beforeEach(() => {
        delete process.env.TORMENTNEXUS_TRPC_UPSTREAM;
        delete process.env.TORMENTNEXUS_CONFIG_DIR;
        delete process.env.NEXT_PUBLIC_TORMENTNEXUS_ORCHESTRATOR_URL;
        delete process.env.NEXT_PUBLIC_AUTOPILOT_URL;
    });

    afterEach(() => {
        globalThis.fetch = originalFetch;
        for (const [key, value] of Object.entries(originalEnv)) {
            if (value === undefined) {
                delete process.env[key];
            } else {
                process.env[key] = value;
            }
        }
        vi.restoreAllMocks();
    });

    it('exports sessions from the configured orchestrator base', async () => {
        process.env.TORMENTNEXUS_ORCHESTRATOR_URL = 'http://127.0.0.1:5001';
        globalThis.fetch = vi.fn(async (input) => {
            expect(String(input)).toBe('http://127.0.0.1:5001/api/sessions');
            return {
                ok: true,
                json: async () => ([
                    { id: 'sess-1', currentTask: 'Ship export lane', status: 'running', startTime: 1234 },
                ]),
            } as Response;
        }) as typeof fetch;

        const caller = createCaller();
        const result = await caller.export({
            format: 'json',
            includeMemories: true,
            includeLogs: true,
            includeMetadata: true,
        });

        expect(result.package.sessionCount).toBe(1);
        expect(result.package.sessions).toEqual([
            expect.objectContaining({
                id: 'sess-1',
                name: 'Ship export lane',
                status: 'running',
                createdAt: 1234,
            }),
        ]);
    });

    it('filters exported sessions when sessionIds are provided', async () => {
        process.env.TORMENTNEXUS_ORCHESTRATOR_URL = 'http://127.0.0.1:5001';
        globalThis.fetch = vi.fn(async (input) => {
            expect(String(input)).toBe('http://127.0.0.1:5001/api/sessions');
            return {
                ok: true,
                json: async () => ([
                    { id: 'sess-1', currentTask: 'Ship export lane', status: 'running', startTime: 1234 },
                    { id: 'sess-2', currentTask: 'Other session', status: 'stopped', startTime: 5678 },
                ]),
            } as Response;
        }) as typeof fetch;

        const caller = createCaller();
        const result = await caller.export({
            format: 'json',
            includeMemories: true,
            includeLogs: true,
            includeMetadata: true,
            sessionIds: ['sess-2'],
        });

        expect(result.package.sessionCount).toBe(1);
        expect(result.package.sessions).toEqual([
            expect.objectContaining({
                id: 'sess-2',
                name: 'Other session',
                status: 'stopped',
                createdAt: 5678,
            }),
        ]);
    });

    it('reports a clear error when importing without an orchestrator base', async () => {
        const configDir = mkdtempSync(`${os.tmpdir()}\\tormentnexus-export-empty-`);
        mkdirSync(configDir, { recursive: true });
        process.env.TORMENTNEXUS_CONFIG_DIR = configDir;

        const fetchSpy = vi.fn();
        globalThis.fetch = fetchSpy as typeof fetch;

        const caller = createCaller();
        const result = await caller.import({
            data: JSON.stringify({
                version: '1.0',
                sessions: [{
                    id: 'sess-import-1',
                    name: 'Imported Session',
                    cliType: 'tormentnexus',
                    status: 'stopped',
                    createdAt: 1234,
                    workingDirectory: 'C:\\temp',
                    metadata: {},
                    logs: [],
                    memories: [],
                }],
            }),
            merge: true,
            dryRun: false,
        });

        expect(fetchSpy).not.toHaveBeenCalled();
        expect(result.imported).toBe(0);
        expect(result.errors).toEqual([
            'Failed to import session sess-import-1: No TormentNexus Orchestrator base configured.',
        ]);
        expect(result.details).toEqual([
            {
                sessionId: 'sess-import-1',
                action: 'error',
                reason: 'No TormentNexus Orchestrator base configured.',
            },
        ]);
    });
});
