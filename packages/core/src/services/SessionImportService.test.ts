import fs from 'fs/promises';
import os from 'os';
import path from 'path';

import Database from 'better-sqlite3';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { SessionImportService } from './SessionImportService.js';

type FakeImportedSession = {
    id: string;
    transcriptHash: string;
    parsedMemories: Array<{
        id: string;
        importedSessionId: string;
        kind: 'memory' | 'instruction';
        content: string;
        tags: string[];
        source: 'llm' | 'heuristic';
        metadata: Record<string, unknown>;
        createdAt: number;
    }>;
    [key: string]: unknown;
};

const tempRoots: string[] = [];
const sqliteAvailable = (() => {
    try {
        const db = new Database(':memory:');
        db.close();
        return true;
    } catch {
        return false;
    }
})();

async function createTempRoot(): Promise<string> {
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-session-import-'));
    tempRoots.push(root);
    return root;
}

function createFakeStore() {
    const sessions: FakeImportedSession[] = [];

    return {
        sessions,
        hasTranscriptHash: vi.fn((hash: string) => sessions.some((session) => session.transcriptHash === hash)),
        upsertSession: vi.fn((input: any) => {
            const id = `session-${sessions.length + 1}`;
            const now = Date.now();
            const parsedMemories = (input.parsedMemories ?? []).map((memory: any, index: number) => ({
                id: `memory-${sessions.length + 1}-${index + 1}`,
                importedSessionId: id,
                kind: memory.kind,
                content: memory.content,
                tags: memory.tags ?? [],
                source: memory.source,
                metadata: memory.metadata ?? {},
                createdAt: now,
            }));

            const record: FakeImportedSession = {
                id,
                ...input,
                externalSessionId: input.externalSessionId ?? null,
                title: input.title ?? null,
                excerpt: input.excerpt ?? null,
                workingDirectory: input.workingDirectory ?? null,
                metadata: input.metadata ?? {},
                parsedMemories,
                createdAt: now,
                updatedAt: now,
            };

            sessions.push(record);
            return record;
        }),
        listImportedSessions: vi.fn((limit: number = 50) => sessions.slice(-limit).reverse()),
        getImportedSession: vi.fn((id: string) => sessions.find((session) => session.id === id) ?? null),
        listInstructionMemories: vi.fn((limit: number = 200) => sessions
            .flatMap((session) => session.parsedMemories)
            .filter((memory) => memory.kind === 'instruction')
            .slice(0, limit)),
        compactInlineTranscripts: vi.fn(),
        backfillRetentionSummaries: vi.fn(),
    };
}

afterEach(async () => {
    await Promise.all(tempRoots.splice(0).map((root) => fs.rm(root, { recursive: true, force: true })));
    vi.restoreAllMocks();
});

describe('SessionImportService', () => {
    it('discovers session transcripts, stores imported sessions, derives memories, and writes instruction docs', async () => {
        const root = await createTempRoot();
        const claudeDir = path.join(root, '.claude');
        await fs.mkdir(claudeDir, { recursive: true });
        await fs.writeFile(
            path.join(claudeDir, 'session-1.jsonl'),
            [
                JSON.stringify({ role: 'user', content: 'Use port 4000 for the TormentNexus control plane.' }),
                JSON.stringify({ role: 'assistant', content: 'Always prefer PowerShell on Windows for TormentNexus shell commands.' }),
            ].join('\n'),
            'utf-8',
        );

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const llmService = {
            generateText: vi.fn(async () => {
                throw new Error('OpenAI API Key not configured.');
            }),
        } as any;

        const service = new SessionImportService(llmService, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: false,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.storedMemoryCount).toBeGreaterThan(0);
        expect(store.upsertSession).toHaveBeenCalledTimes(1);
        expect(store.sessions[0]?.metadata).toMatchObject({
            retentionSummary: {
                archiveDisposition: 'archive_only',
                strategy: 'heuristic',
            },
        });
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);

        const docs = await service.listInstructionDocs();
        expect(docs).toHaveLength(1);
        const content = await fs.readFile(docs[0].path, 'utf-8');
        expect(content).toContain('Auto-imported Agent Instructions');
        expect(content).toContain('Always prefer PowerShell on Windows for TormentNexus shell commands.');
    });

    it('skips already imported transcript hashes unless force is requested', async () => {
        const root = await createTempRoot();
        const copilotDir = path.join(root, '.copilot', 'session-state');
        await fs.mkdir(copilotDir, { recursive: true });
        await fs.writeFile(
            path.join(copilotDir, 'session-1.md'),
            'Remember to use the dashboard on port 3000 and the control plane on port 4000.',
            'utf-8',
        );

        const store = createFakeStore();
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: false,
            maxFilesPerRoot: 20,
        });

        const first = await service.scanAndImport();
        const second = await service.scanAndImport();
        const forced = await service.scanAndImport({ force: true });

        expect(first.importedCount).toBe(1);
        expect(second.importedCount).toBe(0);
        expect(second.skippedCount).toBe(1);
        expect(forced.importedCount).toBe(1);
    });

    it('logs concise degraded-mode warnings when imported-session SQLite maintenance is unavailable', async () => {
        const root = await createTempRoot();
        const store = createFakeStore();
        const sqliteUnavailable = new Error(
            'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
        );
        store.compactInlineTranscripts.mockImplementation(() => {
            throw sqliteUnavailable;
        });
        store.backfillRetentionSummaries.mockImplementation(() => {
            throw sqliteUnavailable;
        });
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: false,
            maxFilesPerRoot: 20,
        });

        expect(warnSpy).toHaveBeenCalledWith(
            '[SessionImport] Failed to compact inline imported transcripts: SQLite runtime is unavailable for this run.',
        );
        expect(warnSpy).toHaveBeenCalledWith(
            '[SessionImport] Failed to backfill imported transcript retention summaries: SQLite runtime is unavailable for this run.',
        );
        expect(warnSpy.mock.calls.flat().join('\n')).not.toContain('Could not locate the bindings file');
    });

    it('skips imported-session persistence when SQLite-backed store operations are unavailable during scan', async () => {
        const root = await createTempRoot();
        const copilotDir = path.join(root, '.copilot', 'session-state');
        await fs.mkdir(copilotDir, { recursive: true });
        await fs.writeFile(
            path.join(copilotDir, 'session-1.md'),
            'Remember to keep startup running even when imported-session SQLite storage is offline.',
            'utf-8',
        );

        const store = createFakeStore();
        const sqliteUnavailable = new Error(
            'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
        );
        store.hasTranscriptHash.mockImplementation(() => {
            throw sqliteUnavailable;
        });
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: false,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(0);
        expect(result.skippedCount).toBe(1);
        expect(warnSpy).toHaveBeenCalledTimes(1);
        expect(String(warnSpy.mock.calls[0]?.[0] ?? '')).toContain('[SessionImport] Skipping imported-session persistence for ');
        expect(String(warnSpy.mock.calls[0]?.[0] ?? '')).toContain('.copilot/session-state/session-1.md: SQLite runtime is unavailable for this run.');
        expect(warnSpy.mock.calls.flat().join('\n')).not.toContain('Could not locate the bindings file');
    });

    it('imports VS Code Copilot Chat home-directory sessions with UUID filenames', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const fakeAppData = await createTempRoot();
        const emptyWindowDir = path.join(fakeAppData, 'Code - Insiders', 'User', 'globalStorage', 'emptyWindowChatSessions');
        await fs.mkdir(emptyWindowDir, { recursive: true });
        await fs.writeFile(
            path.join(emptyWindowDir, '6eebfd4c-8213-460d-bc4e-219ff6aada7b.jsonl'),
            [
                JSON.stringify({ request: { message: 'Remember that TormentNexus runs on port 4000.' } }),
                JSON.stringify({ response: { message: 'Always use PowerShell paths on Windows.' } }),
            ].join('\n'),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);
        vi.stubEnv('APPDATA', fakeAppData);
        vi.stubEnv('LOCALAPPDATA', fakeAppData);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('copilot-chat');
        expect(store.sessions[0]?.sourceTool).toBe('copilot-chat');
        expect(store.sessions[0]?.sourcePath).toContain('emptyWindowChatSessions');
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('imports VS Code extension chat history from global storage without duplicating copilot-chat roots', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const fakeAppData = await createTempRoot();
        const extensionDir = path.join(fakeAppData, 'Code', 'User', 'globalStorage', 'some-extension');
        await fs.mkdir(extensionDir, { recursive: true });
        await fs.writeFile(
            path.join(extensionDir, 'chat-history.jsonl'),
            [
                JSON.stringify({ role: 'user', content: 'Remember to import VS Code extension chats.' }),
                JSON.stringify({ role: 'assistant', content: 'Keep durable instructions in long-term memory.' }),
            ].join('\n'),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);
        vi.stubEnv('APPDATA', fakeAppData);
        vi.stubEnv('LOCALAPPDATA', fakeAppData);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('vscode-extensions');
        expect(store.sessions[0]?.sourceTool).toBe('vscode-extensions');
        expect(store.sessions[0]?.sourcePath).toContain('globalStorage');
        expect(store.sessions[0]?.sourcePath).toContain('some-extension');
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('imports OpenAI or ChatGPT JSON history from explicit home-directory roots', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const openAiDir = path.join(fakeHome, '.openai');
        await fs.mkdir(openAiDir, { recursive: true });
        await fs.writeFile(
            path.join(openAiDir, 'chatgpt-conversation-history.json'),
            JSON.stringify([
                {
                    role: 'user',
                    content: 'Remember that TormentNexus should import durable chat history automatically.',
                    created_at: 1711500000,
                },
                {
                    role: 'assistant',
                    content: 'Prefer truthful, stabilization-first imports over noisy crawling.',
                    tool_calls: [
                        {
                            function: {
                                name: 'session_search_memory',
                            },
                        },
                    ],
                    created_at: 1711500010,
                },
            ]),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('openai');
        expect(store.sessions[0]?.sourceTool).toBe('openai');
        expect(store.sessions[0]?.transcript).toContain('User: Remember that TormentNexus should import durable chat history automatically.');
        expect(store.sessions[0]?.transcript).toContain('Assistant: Prefer truthful, stabilization-first imports over noisy crawling.');
        expect(store.sessions[0]?.transcript).toContain('[Tool Use: session_search_memory]');
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('imports ChatGPT exported conversations.json mapping trees as per-conversation sessions', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const chatGptDir = path.join(fakeHome, 'ChatGPT');
        await fs.mkdir(chatGptDir, { recursive: true });
        await fs.writeFile(
            path.join(chatGptDir, 'conversations.json'),
            JSON.stringify([
                {
                    id: 'conv-1',
                    title: 'TormentNexus stabilization plan',
                    create_time: 1711500100,
                    update_time: 1711500500,
                    current_node: 'assistant-final',
                    mapping: {
                        root: {
                            id: 'root',
                            parent: null,
                            children: ['user-1'],
                            message: null,
                        },
                        'user-1': {
                            id: 'user-1',
                            parent: 'root',
                            children: ['assistant-1', 'assistant-alt'],
                            message: {
                                id: 'msg-user-1',
                                author: { role: 'user' },
                                create_time: 1711500200,
                                content: {
                                    content_type: 'text',
                                    parts: ['Remember the next stabilization fix should be truthful and testable.'],
                                },
                            },
                        },
                        'assistant-alt': {
                            id: 'assistant-alt',
                            parent: 'user-1',
                            children: [],
                            message: {
                                id: 'msg-assistant-alt',
                                author: { role: 'assistant' },
                                create_time: 1711500300,
                                content: {
                                    content_type: 'text',
                                    parts: ['This branch should not be imported because it is not current.'],
                                },
                            },
                        },
                        'assistant-1': {
                            id: 'assistant-1',
                            parent: 'user-1',
                            children: ['assistant-final'],
                            message: {
                                id: 'msg-assistant-1',
                                author: { role: 'assistant' },
                                create_time: 1711500400,
                                content: {
                                    content_type: 'text',
                                    parts: ['Prefer fixes that close real runtime gaps before broad expansion.'],
                                },
                            },
                        },
                        'assistant-final': {
                            id: 'assistant-final',
                            parent: 'assistant-1',
                            children: [],
                            message: {
                                id: 'msg-assistant-final',
                                author: { role: 'assistant' },
                                create_time: 1711500500,
                                content: {
                                    content_type: 'text',
                                    parts: ['Then validate with focused tests and the repo suite.'],
                                },
                            },
                        },
                    },
                },
            ]),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('openai');
        expect(store.sessions[0]?.sourceTool).toBe('openai');
        expect(store.sessions[0]?.sessionFormat).toBe('chatgpt-export');
        expect(String(store.sessions[0]?.sourcePath)).toContain('#conversation:conv-1');
        expect(store.sessions[0]?.title).toBe('TormentNexus stabilization plan');
        expect(String(store.sessions[0]?.transcript)).toContain('User: Remember the next stabilization fix should be truthful and testable.');
        expect(String(store.sessions[0]?.transcript)).toContain('Assistant: Prefer fixes that close real runtime gaps before broad expansion.');
        expect(String(store.sessions[0]?.transcript)).toContain('Assistant: Then validate with focused tests and the repo suite.');
        expect(String(store.sessions[0]?.transcript)).not.toContain('This branch should not be imported because it is not current.');
        expect(store.sessions[0]?.metadata).toMatchObject({
            openaiExportType: 'chatgpt-conversations',
            openaiConversationId: 'conv-1',
        });
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('imports Opencode home-directory message history', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const opencodeMessagesDir = path.join(fakeHome, '.opencode', 'messages', 'ses_abc123');
        await fs.mkdir(opencodeMessagesDir, { recursive: true });
        await fs.writeFile(
            path.join(opencodeMessagesDir, 'msg_001.json'),
            JSON.stringify({
                message: {
                    role: 'user',
                    content: 'Remember that Opencode sessions should be imported from the home directory.',
                },
                response: {
                    role: 'assistant',
                    content: 'Use the shared session importer instead of a one-off migration path.',
                },
            }),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('opencode');
        expect(store.sessions[0]?.sourceTool).toBe('opencode');
        expect(String(store.sessions[0]?.transcript)).toContain('Remember that Opencode sessions should be imported from the home directory.');
        expect(String(store.sessions[0]?.transcript)).toContain('Use the shared session importer instead of a one-off migration path.');
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('imports Experimental Antigravity brain logs from explicit home-directory roots', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const antigravityBrainDir = path.join(fakeHome, '.gemini', 'antigravity', 'brain');
        await fs.mkdir(antigravityBrainDir, { recursive: true });
        await fs.writeFile(
            path.join(antigravityBrainDir, '2026-03-28-session.jsonl'),
            [
                JSON.stringify({ request: { message: 'Remember that Antigravity import should stay explicitly experimental.' } }),
                JSON.stringify({ response: { message: 'Only use the reverse-engineered brain root until a stable export contract exists.' } }),
            ].join('\n'),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('antigravity');
        expect(store.sessions[0]?.sourceTool).toBe('antigravity');
        expect(String(store.sessions[0]?.sourcePath).replaceAll('\\', '/')).toContain('/.gemini/antigravity/brain/');
        expect(String(store.sessions[0]?.transcript)).toContain('Remember that Antigravity import should stay explicitly experimental.');
        expect(String(store.sessions[0]?.transcript)).toContain('Only use the reverse-engineered brain root until a stable export contract exists.');
        expect(store.sessions[0]?.metadata).toMatchObject({
            antigravityImportSurface: 'experimental',
            antigravityDiscoveryRoot: 'brain',
            antigravitySource: 'reverse-engineered',
        });
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });

    it('skips a discovered session file that disappears before import', async () => {
        const root = await createTempRoot();
        const tormentnexusSessionDir = path.join(root, '.tormentnexus', 'sessions');
        const sessionPath = path.join(tormentnexusSessionDir, 'session_123.json');
        await fs.mkdir(tormentnexusSessionDir, { recursive: true });
        await fs.writeFile(sessionPath, JSON.stringify({ summary: 'Transient session file' }), 'utf-8');

        const realReadFile = fs.readFile.bind(fs);
        vi.spyOn(fs, 'readFile').mockImplementation(async (targetPath, ...args) => {
            if (path.normalize(String(targetPath)) === path.normalize(sessionPath)) {
                const error = new Error(`ENOENT: no such file or directory, open '${sessionPath}'`) as NodeJS.ErrnoException;
                error.code = 'ENOENT';
                throw error;
            }

            return realReadFile(targetPath as Parameters<typeof fs.readFile>[0], ...(args as []));
        });

        const store = createFakeStore();
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: false,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(0);
        expect(result.skippedCount).toBe(1);
        expect(store.sessions).toHaveLength(0);
    });

    it('ignores Cursor workspace metadata files that are not actual chat sessions', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const fakeAppData = await createTempRoot();
        const cursorStorageDir = path.join(
            fakeAppData,
            'Cursor',
            'User',
            'workspaceStorage',
            'a49e254c57e8a9392ef9ecab474d23ec',
            'anysphere.cursor-retrieval',
        );
        await fs.mkdir(cursorStorageDir, { recursive: true });
        await fs.writeFile(
            path.join(cursorStorageDir, 'high_level_folder_description.txt'),
            'This file describes the repository for retrieval indexing.',
            'utf-8',
        );
        await fs.writeFile(
            path.join(path.dirname(cursorStorageDir), 'workspace.json'),
            JSON.stringify({ folder: 'C:\\Users\\hyper\\workspace\\tormentnexus' }),
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);
        vi.stubEnv('APPDATA', fakeAppData);
        vi.stubEnv('LOCALAPPDATA', fakeAppData);

        const store = createFakeStore();
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(0);
        expect(result.importedCount).toBe(0);
        expect(store.sessions).toHaveLength(0);
    });

    it('ignores Copilot Chat embeddings and helper docs that are not session transcripts', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const fakeAppData = await createTempRoot();
        const copilotChatDir = path.join(fakeAppData, 'Code - Insiders', 'User', 'globalStorage', 'github.copilot-chat');
        await fs.mkdir(path.join(copilotChatDir, 'memory-tool', 'memories'), { recursive: true });
        await fs.mkdir(path.join(copilotChatDir, 'ask-agent'), { recursive: true });
        await fs.writeFile(
            path.join(copilotChatDir, 'commandEmbeddings.json'),
            JSON.stringify({ embedding: [1, 2, 3] }),
            'utf-8',
        );
        await fs.writeFile(
            path.join(copilotChatDir, 'memory-tool', 'memories', 'preferences.md'),
            'Remember the user likes dark themes.',
            'utf-8',
        );
        await fs.writeFile(
            path.join(copilotChatDir, 'ask-agent', 'Ask.agent.md'),
            'Agent instructions, not a session transcript.',
            'utf-8',
        );

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);
        vi.stubEnv('APPDATA', fakeAppData);
        vi.stubEnv('LOCALAPPDATA', fakeAppData);

        const store = createFakeStore();
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm: vi.fn(async () => ({})),
            captureSessionSummary: vi.fn(async () => ({})),
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(0);
        expect(result.importedCount).toBe(0);
        expect(store.sessions).toHaveLength(0);
    });

    (sqliteAvailable ? it : it.skip)('imports TormentNexus ledger and handoff entries from the local data.db store', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const tormentnexusDir = path.join(fakeHome, '.tormentnexus-mcp');
        const tormentnexusDbPath = path.join(tormentnexusDir, 'data.db');
        await fs.mkdir(tormentnexusDir, { recursive: true });

        const tormentnexusDb = new Database(tormentnexusDbPath);
        tormentnexusDb.exec(`
            CREATE TABLE session_ledger (
                id TEXT PRIMARY KEY,
                project TEXT NOT NULL,
                conversation_id TEXT NOT NULL,
                summary TEXT NOT NULL,
                todos TEXT DEFAULT '[]',
                files_changed TEXT DEFAULT '[]',
                decisions TEXT DEFAULT '[]',
                keywords TEXT DEFAULT '[]',
                role TEXT DEFAULT 'dev',
                event_type TEXT DEFAULT 'session',
                confidence_score INTEGER DEFAULT NULL,
                importance INTEGER DEFAULT 0,
                created_at TEXT DEFAULT '2026-03-27T00:00:00.000Z'
            );
            CREATE TABLE session_handoffs (
                project TEXT PRIMARY KEY,
                last_summary TEXT DEFAULT NULL,
                pending_todo TEXT DEFAULT '[]',
                active_decisions TEXT DEFAULT '[]',
                keywords TEXT DEFAULT '[]',
                key_context TEXT DEFAULT NULL,
                active_branch TEXT DEFAULT NULL,
                version INTEGER DEFAULT 1,
                metadata TEXT DEFAULT '{}',
                created_at TEXT DEFAULT '2026-03-27T00:00:00.000Z',
                updated_at TEXT DEFAULT '2026-03-27T01:00:00.000Z'
            );
        `);
        tormentnexusDb
            .prepare(`
                INSERT INTO session_ledger (
                    id, project, conversation_id, summary, todos, files_changed, decisions, keywords, role, event_type, confidence_score, importance, created_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            `)
            .run(
                'ledger-1',
                'tormentnexus',
                'conversation-1',
                'Integrated the TormentNexus session importer into TormentNexus.',
                JSON.stringify(['Add dashboard visibility for TormentNexus memory state']),
                JSON.stringify(['packages/core/src/services/SessionImportService.ts']),
                JSON.stringify(['Prefer importing TormentNexus ledger summaries into TormentNexus memory']),
                JSON.stringify(['tormentnexus', 'memory', 'import']),
                'dev',
                'correction',
                92,
                5,
                '2026-03-27T02:00:00.000Z',
            );
        tormentnexusDb
            .prepare(`
                INSERT INTO session_handoffs (
                    project, last_summary, pending_todo, active_decisions, keywords, key_context, active_branch, version, metadata, updated_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            `)
            .run(
                'tormentnexus',
                'TormentNexus handoff summary for TormentNexus.',
                JSON.stringify(['Keep assimilating TormentNexus features carefully']),
                JSON.stringify(['Use TormentNexus as the operator-facing control plane']),
                JSON.stringify(['tormentnexus', 'handoff']),
                'Current focus: bridge durable memory and imported sessions.',
                'main',
                7,
                JSON.stringify({ cwd: 'C:\\Users\\hyper\\workspace\\tormentnexus' }),
                '2026-03-27T03:00:00.000Z',
            );
        tormentnexusDb.close();

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(2);
        expect(result.importedCount).toBe(2);
        expect(result.tools).toContain('tormentnexus-mcp');
        expect(store.sessions.map((session) => session.sessionFormat)).toEqual(
            expect.arrayContaining(['tormentnexus-ledger', 'tormentnexus-handoff']),
        );
        expect(store.sessions.map((session) => session.sourcePath)).toEqual(
            expect.arrayContaining([
                expect.stringContaining('#session_ledger:ledger-1'),
                expect.stringContaining('#session_handoffs:tormentnexus'),
            ]),
        );
        expect(store.sessions.map((session) => session.workingDirectory)).toContain('C:\\Users\\hyper\\workspace\\tormentnexus');
        const tormentnexusLedger = store.sessions.find((session) => session.sessionFormat === 'tormentnexus-ledger');
        expect(tormentnexusLedger?.metadata).toMatchObject({
            tormentnexusEventType: 'correction',
            tormentnexusConfidenceScore: 92,
            tormentnexusImportance: 5,
            behavioralWarnings: ['Integrated the TormentNexus session importer into TormentNexus.'],
        });
        expect(String(tormentnexusLedger?.transcript)).toContain('Event type: correction');
        expect(String(tormentnexusLedger?.transcript)).toContain('Confidence score: 92');
        expect(String(tormentnexusLedger?.transcript)).toContain('Importance: 5');
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(2);
    });

    (sqliteAvailable ? it : it.skip)('imports llm CLI conversations from logs.db', async () => {
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        const llmDir = path.join(fakeHome, '.llm');
        const llmDbPath = path.join(llmDir, 'logs.db');
        await fs.mkdir(llmDir, { recursive: true });

        const llmDb = new Database(llmDbPath);
        llmDb.exec(`
            CREATE TABLE conversations (
                id TEXT PRIMARY KEY,
                name TEXT,
                model TEXT
            );
            CREATE TABLE responses (
                id TEXT PRIMARY KEY,
                model TEXT,
                prompt TEXT,
                system TEXT,
                prompt_json TEXT,
                options_json TEXT,
                response TEXT,
                response_json TEXT,
                conversation_id TEXT,
                duration_ms INTEGER,
                datetime_utc TEXT,
                input_tokens INTEGER,
                output_tokens INTEGER,
                token_details TEXT,
                schema_id TEXT,
                resolved_model TEXT
            );
            CREATE TABLE tool_calls (
                id INTEGER PRIMARY KEY,
                response_id TEXT,
                tool_id INTEGER,
                name TEXT,
                arguments TEXT,
                tool_call_id TEXT
            );
            CREATE TABLE tool_results (
                id INTEGER PRIMARY KEY,
                response_id TEXT,
                tool_id INTEGER,
                name TEXT,
                output TEXT,
                tool_call_id TEXT,
                instance_id INTEGER,
                exception TEXT
            );
        `);
        llmDb
            .prepare('INSERT INTO conversations (id, name, model) VALUES (?, ?, ?)')
            .run('conv-1', 'TormentNexus import strategy', 'anthropic/claude-3-5-sonnet');
        llmDb
            .prepare(`
                INSERT INTO responses (
                    id, model, prompt, system, response, conversation_id, datetime_utc, input_tokens, output_tokens, resolved_model
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            `)
            .run(
                'resp-1',
                'anthropic/claude-3-5-sonnet',
                'How should TormentNexus import llm CLI logs?',
                'Be concise and durable.',
                'Use the shared session importer so logs become searchable memory.',
                'conv-1',
                '2026-03-28T01:00:00.000Z',
                120,
                40,
                'anthropic/claude-3-5-sonnet',
            );
        llmDb
            .prepare(`
                INSERT INTO responses (
                    id, model, prompt, system, response, conversation_id, datetime_utc, input_tokens, output_tokens, resolved_model
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            `)
            .run(
                'resp-2',
                'anthropic/claude-3-5-sonnet',
                'Should tool usage be preserved?',
                '',
                'Yes, keep tool markers in the imported transcript.',
                'conv-1',
                '2026-03-28T01:05:00.000Z',
                90,
                30,
                'anthropic/claude-3-5-sonnet',
            );
        llmDb
            .prepare('INSERT INTO tool_calls (response_id, name, arguments, tool_call_id) VALUES (?, ?, ?, ?)')
            .run('resp-2', 'search_docs', '{"query":"llm logs"}', 'call-1');
        llmDb
            .prepare('INSERT INTO tool_results (response_id, name, output, tool_call_id) VALUES (?, ?, ?, ?)')
            .run('resp-2', 'search_docs', 'Found logging documentation.', 'call-1');
        llmDb.close();

        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);

        const store = createFakeStore();
        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store: store as any,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('llm-cli');
        expect(store.sessions[0]?.sourceTool).toBe('llm-cli');
        expect(store.sessions[0]?.sessionFormat).toBe('llm-conversation');
        expect(String(store.sessions[0]?.sourcePath)).toContain('#conversation:conv-1');
        expect(String(store.sessions[0]?.transcript)).toContain('System: Be concise and durable.');
        expect(String(store.sessions[0]?.transcript)).toContain('User: How should TormentNexus import llm CLI logs?');
        expect(String(store.sessions[0]?.transcript)).toContain('Assistant: Use the shared session importer so logs become searchable memory.');
        expect(String(store.sessions[0]?.transcript)).toContain('[Tool Call: search_docs]');
        expect(String(store.sessions[0]?.transcript)).toContain('[Tool Result: search_docs]');
        expect(store.sessions[0]?.metadata).toMatchObject({
            llmConversationId: 'conv-1',
            llmResponseCount: 2,
            llmInputTokens: 210,
            llmOutputTokens: 70,
        });
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });
});
