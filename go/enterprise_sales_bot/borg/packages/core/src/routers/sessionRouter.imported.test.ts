import { beforeEach, describe, expect, it, vi } from 'vitest';

let mockSessionImportService: {
    scanAndImport: ReturnType<typeof vi.fn>;
    listImportedSessions: ReturnType<typeof vi.fn>;
    getImportedSession: ReturnType<typeof vi.fn>;
    getImportedMaintenanceStats?: ReturnType<typeof vi.fn>;
    listInstructionDocs?: ReturnType<typeof vi.fn>;
} | undefined;

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getSessionImportService: () => mockSessionImportService,
        getAgentMemoryService: () => {
            throw new Error('getAgentMemoryService should not be called in imported-session router tests.');
        },
        getSessionManager: () => {
            throw new Error('getSessionManager should not be called in imported-session router tests.');
        },
        getSessionSupervisor: () => {
            throw new Error('getSessionSupervisor should not be called in imported-session router tests.');
        },
        getShellService: () => {
            throw new Error('getShellService should not be called in imported-session router tests.');
        },
    };
});

import { sessionRouter } from './sessionRouter.js';

function createCaller() {
    return sessionRouter.createCaller({} as never);
}

function createImportedSession(id: string) {
    return {
        id,
        sourceTool: 'antigravity',
        sourcePath: `C:\\temp\\${id}.jsonl`,
        externalSessionId: id,
        title: `Imported ${id}`,
        sessionFormat: 'jsonl',
        transcript: 'User: Keep the import lane experimental.',
        excerpt: 'Keep the import lane experimental.',
        workingDirectory: 'C:\\temp',
        transcriptHash: `hash-${id}`,
        normalizedSession: {
            id,
            sourceTool: 'antigravity',
        },
        metadata: {
            antigravityImportSurface: 'experimental',
        },
        discoveredAt: 1_700_000_000_000,
        importedAt: 1_700_000_000_100,
        lastModifiedAt: 1_700_000_000_200,
        createdAt: 1_700_000_000_100,
        updatedAt: 1_700_000_000_100,
        parsedMemories: [
            {
                id: `memory-${id}`,
                importedSessionId: id,
                kind: 'instruction' as const,
                content: 'Keep the Antigravity importer explicitly experimental.',
                tags: ['antigravity', 'instruction'],
                source: 'heuristic' as const,
                metadata: {
                    extraction: 'heuristic',
                },
                createdAt: 1_700_000_000_150,
            },
        ],
    };
}

describe('sessionRouter imported-session procedures', () => {
    beforeEach(() => {
        mockSessionImportService = undefined;
    });

    it('lists and fetches imported sessions through the session import runtime', async () => {
        const importedSession = createImportedSession('session-1');
        mockSessionImportService = {
            scanAndImport: vi.fn(),
            listImportedSessions: vi.fn(() => [importedSession]),
            getImportedSession: vi.fn((id: string) => (id === 'session-1' ? importedSession : null)),
            listInstructionDocs: vi.fn(),
        };

        const caller = createCaller();
        const listResult = await caller.importedList({ limit: 1 });
        const getResult = await caller.importedGet({ id: 'session-1' });

        expect(mockSessionImportService.listImportedSessions).toHaveBeenCalledWith(1);
        expect(listResult).toEqual([importedSession]);
        expect(mockSessionImportService.getImportedSession).toHaveBeenCalledWith('session-1');
        expect(getResult).toEqual(importedSession);
    });

    it('returns safe empty defaults when no session import runtime is available', async () => {
        const caller = createCaller();

        await expect(caller.importedList({ limit: 10 })).resolves.toEqual([]);
        await expect(caller.importedGet({ id: 'missing' })).resolves.toBeNull();
        await expect(caller.importedMaintenanceStats()).resolves.toEqual({
            totalSessions: 0,
            inlineTranscriptCount: 0,
            archivedTranscriptCount: 0,
            missingRetentionSummaryCount: 0,
        });
        await expect(caller.importedInstructionDocs()).resolves.toEqual([]);
        await expect(caller.importedScan({ force: true })).resolves.toEqual({
            discoveredCount: 0,
            importedCount: 0,
            skippedCount: 0,
            storedMemoryCount: 0,
            instructionDocPath: null,
            tools: [],
        });
    });

    it('passes through imported scan results and forwards the force flag', async () => {
        const scanSummary = {
            discoveredCount: 3,
            importedCount: 2,
            skippedCount: 1,
            storedMemoryCount: 4,
            instructionDocPath: 'C:\\temp\\auto-imported-agent-instructions.md',
            tools: ['antigravity', 'openai'],
        };
        const instructionDocs = [
            {
                path: 'C:\\temp\\auto-imported-agent-instructions.md',
                updatedAt: 1_700_000_000_200,
                size: 512,
            },
        ];

        mockSessionImportService = {
            scanAndImport: vi.fn(async (options?: { force?: boolean }) => ({
                ...scanSummary,
                receivedForce: options?.force ?? false,
            })),
            listImportedSessions: vi.fn(() => []),
            getImportedSession: vi.fn(() => null),
            listInstructionDocs: vi.fn(async () => instructionDocs),
        };

        const caller = createCaller();
        const scanResult = await caller.importedScan({ force: true });
        const docsResult = await caller.importedInstructionDocs();

        expect(mockSessionImportService.scanAndImport).toHaveBeenCalledWith({ force: true });
        expect(scanResult).toEqual(scanSummary);
        expect(mockSessionImportService.listInstructionDocs).toHaveBeenCalledTimes(1);
        expect(docsResult).toEqual(instructionDocs);
    });

    it('returns imported maintenance stats through the session import runtime', async () => {
        const stats = {
            totalSessions: 12,
            inlineTranscriptCount: 3,
            archivedTranscriptCount: 9,
            missingRetentionSummaryCount: 2,
        };

        mockSessionImportService = {
            scanAndImport: vi.fn(),
            listImportedSessions: vi.fn(() => []),
            getImportedSession: vi.fn(() => null),
            getImportedMaintenanceStats: vi.fn(() => stats),
            listInstructionDocs: vi.fn(async () => []),
        };

        const caller = createCaller();
        const result = await caller.importedMaintenanceStats();

        expect(mockSessionImportService.getImportedMaintenanceStats).toHaveBeenCalledTimes(1);
        expect(result).toEqual(stats);
    });
});
