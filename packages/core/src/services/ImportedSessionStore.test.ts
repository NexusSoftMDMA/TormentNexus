import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import zlib from 'zlib';

import Database from 'better-sqlite3';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

let sqliteForTest: Database.Database;
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

vi.mock('../db/index.js', () => ({
    get sqliteInstance() {
        return sqliteForTest;
    },
}));

type ImportedSessionStoreModule = typeof import('./ImportedSessionStore.js');

function createSchema(database: Database.Database): void {
    database.pragma('foreign_keys = ON');
    database.exec(`
        CREATE TABLE imported_sessions (
            uuid TEXT PRIMARY KEY,
            source_tool TEXT NOT NULL,
            source_path TEXT NOT NULL,
            source_size INTEGER,
            source_mtime INTEGER,
            external_session_id TEXT,
            title TEXT,
            session_format TEXT NOT NULL,
            transcript TEXT NOT NULL,
            excerpt TEXT,
            working_directory TEXT,
            transcript_hash TEXT NOT NULL UNIQUE,
            transcript_archive_path TEXT,
            transcript_metadata_archive_path TEXT,
            transcript_archive_format TEXT,
            transcript_stored_bytes INTEGER,
            normalized_session TEXT NOT NULL DEFAULT '{}',
            metadata TEXT NOT NULL DEFAULT '{}',
            discovered_at INTEGER NOT NULL,
            imported_at INTEGER NOT NULL,
            last_modified_at INTEGER,
            created_at INTEGER NOT NULL,
            updated_at INTEGER NOT NULL
        );

        CREATE TABLE imported_session_memories (
            uuid TEXT PRIMARY KEY,
            imported_session_uuid TEXT NOT NULL,
            memory_index INTEGER NOT NULL,
            kind TEXT NOT NULL,
            content TEXT NOT NULL,
            tags TEXT NOT NULL DEFAULT '[]',
            source TEXT NOT NULL,
            metadata TEXT NOT NULL DEFAULT '{}',
            created_at INTEGER NOT NULL,
            FOREIGN KEY (imported_session_uuid) REFERENCES imported_sessions(uuid) ON DELETE CASCADE,
            UNIQUE(imported_session_uuid, memory_index)
        );
    `);
}

async function createTempRoot(): Promise<string> {
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-imported-session-store-'));
    tempRoots.push(root);
    return root;
}

function createSessionInput(hash: string, transcript: string, memoryContent: string) {
    return {
        sourceTool: 'antigravity',
        sourcePath: `C:\\temp\\${hash}.jsonl`,
        externalSessionId: `ext-${hash}`,
        title: `Imported ${hash}`,
        sessionFormat: 'jsonl',
        transcript,
        excerpt: transcript.slice(0, 40),
        workingDirectory: 'C:\\temp',
        transcriptHash: hash,
        normalizedSession: {
            sourceTool: 'antigravity',
            transcriptHash: hash,
        },
        metadata: {
            antigravityImportSurface: 'experimental',
        },
        discoveredAt: 1_700_000_000_000,
        importedAt: 1_700_000_000_100,
        lastModifiedAt: 1_700_000_000_050,
        parsedMemories: [
            {
                kind: 'instruction' as const,
                content: memoryContent,
                tags: ['antigravity', 'instruction'],
                source: 'heuristic' as const,
                metadata: { extraction: 'heuristic' },
            },
            {
                kind: 'memory' as const,
                content: `memory-${hash}`,
                tags: ['antigravity', 'memory'],
                source: 'llm' as const,
                metadata: { extraction: 'llm' },
            },
        ],
    };
}

describe('ImportedSessionStore', () => {
    beforeEach(() => {
        if (!sqliteAvailable) {
            return;
        }
        sqliteForTest = new Database(':memory:');
        createSchema(sqliteForTest);
    });

    afterEach(async () => {
        sqliteForTest?.close();
        await Promise.all(tempRoots.splice(0).map((root) => fs.rm(root, { recursive: true, force: true })));
        vi.restoreAllMocks();
    });

    it('returns false and logs a concise warning when transcript deduplication is unavailable', async () => {
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const prepareSpy = vi.spyOn(sqliteForTest, 'prepare').mockImplementation(() => {
            throw new Error('SQLite runtime is unavailable for this run.');
        });

        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const store = new ImportedSessionStore(await createTempRoot());

        expect(store.hasTranscriptHash('hash-missing-sqlite')).toBe(false);
        expect(warnSpy).toHaveBeenCalledWith(
            '[ImportedSessionStore] Transcript deduplication is unavailable: SQLite runtime is unavailable for this run.',
        );
        prepareSpy.mockRestore();
    });

    it('upserts, lists, and fetches imported sessions with parsed memories', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const store = new ImportedSessionStore(await createTempRoot());
        const input = createSessionInput('hash-a', 'User: keep imports truthful.', 'Keep imports truthful.');

        const created = store.upsertSession(input);
        const fetched = store.getImportedSession(created.id);
        const listed = store.listImportedSessions(10);

        expect(store.hasTranscriptHash('hash-a')).toBe(true);
        expect(created.sourceTool).toBe('antigravity');
        expect(created.parsedMemories).toHaveLength(2);
        expect(fetched).not.toBeNull();
        expect(fetched?.parsedMemories.map((entry) => entry.content)).toEqual([
            'Keep imports truthful.',
            'memory-hash-a',
        ]);
        expect(created.transcript).toContain('keep imports truthful');
        expect(listed).toHaveLength(1);
        expect(listed[0]?.id).toBe(created.id);
    });

    it('reuses the same session row for an existing transcript hash and replaces parsed memories', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const store = new ImportedSessionStore(await createTempRoot());

        const first = store.upsertSession(createSessionInput('hash-b', 'User: first transcript.', 'First instruction.'));
        const second = store.upsertSession({
            ...createSessionInput('hash-b', 'User: updated transcript.', 'Updated instruction.'),
            transcript: 'User: updated transcript.\n\nAssistant: keep antigravity experimental.',
            metadata: {
                antigravityImportSurface: 'experimental',
                revision: 2,
            },
        });

        expect(second.id).toBe(first.id);
        expect(second.transcript).toContain('updated transcript');
        expect(second.metadata).toMatchObject({
            antigravityImportSurface: 'experimental',
            revision: 2,
        });
        expect(second.parsedMemories).toHaveLength(2);
        expect(second.parsedMemories[0]?.content).toBe('Updated instruction.');

        const memoryCount = sqliteForTest
            .prepare('SELECT COUNT(*) AS count FROM imported_session_memories WHERE imported_session_uuid = ?')
            .get(first.id) as { count: number };
        expect(memoryCount.count).toBe(2);
    });

    it('lists only instruction memories in newest-first order', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const nowSpy = vi.spyOn(Date, 'now');
        nowSpy.mockReturnValueOnce(1_700_000_000_100).mockReturnValueOnce(1_700_000_000_200);

        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const store = new ImportedSessionStore(await createTempRoot());

        store.upsertSession(createSessionInput('hash-c', 'User: first transcript.', 'First instruction.'));
        store.upsertSession(createSessionInput('hash-d', 'User: second transcript.', 'Second instruction.'));

        const instructionMemories = store.listInstructionMemories(10);

        expect(instructionMemories).toHaveLength(2);
        expect(instructionMemories.map((entry) => entry.kind)).toEqual(['instruction', 'instruction']);
        expect(instructionMemories.map((entry) => entry.content)).toEqual([
            'Second instruction.',
            'First instruction.',
        ]);
    });

    it('compacts legacy inline transcripts into gzip archives and keeps them readable', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const archiveRoot = await createTempRoot();
        const store = new ImportedSessionStore(archiveRoot);

        sqliteForTest.prepare(`
            INSERT INTO imported_sessions (
                uuid,
                source_tool,
                source_path,
                external_session_id,
                title,
                session_format,
                transcript,
                excerpt,
                working_directory,
                transcript_hash,
                normalized_session,
                metadata,
                transcript_archive_path,
                transcript_metadata_archive_path,
                transcript_archive_format,
                transcript_stored_bytes,
                discovered_at,
                imported_at,
                last_modified_at,
                created_at,
                updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `).run(
            'legacy-session',
            'antigravity',
            'C:\\temp\\legacy.jsonl',
            'legacy-ext',
            'Legacy import',
            'jsonl',
            'User: keep compact storage truthful.\nAssistant: store only valuable memories in the fast path.',
            'User: keep compact storage truthful.',
            'C:\\temp',
            'legacy-hash',
            JSON.stringify({ sourceTool: 'antigravity', transcriptHash: 'legacy-hash' }),
            JSON.stringify({ legacy: true }),
            null,
            null,
            null,
            null,
            1_700_000_000_000,
            1_700_000_000_100,
            1_700_000_000_050,
            1_700_000_000_100,
            1_700_000_000_100,
        );

        const compacted = store.compactInlineTranscripts(10);
        const fetched = store.getImportedSession('legacy-session');
        const row = sqliteForTest
            .prepare('SELECT transcript, transcript_archive_path, transcript_metadata_archive_path, transcript_archive_format, transcript_stored_bytes FROM imported_sessions WHERE uuid = ?')
            .get('legacy-session') as Record<string, unknown>;

        expect(compacted).toBe(1);
        expect(String(row.transcript ?? '')).toBe('');
        expect(typeof row.transcript_archive_path).toBe('string');
        expect(typeof row.transcript_metadata_archive_path).toBe('string');
        expect(row.transcript_archive_format).toBe('gzip-text-v1');
        expect(Number(row.transcript_stored_bytes ?? 0)).toBeGreaterThan(0);
        expect(fetched?.transcript).toContain('keep compact storage truthful');
        expect(fetched?.transcript).toContain('valuable memories');
    });

    it('writes retention summary into compressed archive metadata sidecars', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const archiveRoot = await createTempRoot();
        const store = new ImportedSessionStore(archiveRoot);

        const created = store.upsertSession({
            ...createSessionInput('hash-retention', 'User: discuss implementation detail.\nAssistant: keep only durable defaults.', 'Keep only durable defaults.'),
            metadata: {
                retentionSummary: {
                    archiveDisposition: 'archive_only',
                    summary: 'Keep defaults durable; archive the rest.',
                },
            },
        });

        const row = sqliteForTest
            .prepare('SELECT transcript_metadata_archive_path FROM imported_sessions WHERE uuid = ?')
            .get(created.id) as Record<string, unknown>;
        const relativePath = String(row.transcript_metadata_archive_path ?? '');
        const metadataPath = path.join(archiveRoot, ...relativePath.split('/'));
        const archiveMetadata = JSON.parse(
            zlib.gunzipSync(await fs.readFile(metadataPath)).toString('utf-8'),
        ) as Record<string, unknown>;

        expect(archiveMetadata.retentionSummary).toMatchObject({
            archiveDisposition: 'archive_only',
            summary: 'Keep defaults durable; archive the rest.',
        });
    });

    it('reports maintenance stats for inline, archived, and missing-retention sessions', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const archiveRoot = await createTempRoot();
        const store = new ImportedSessionStore(archiveRoot);

        const archived = store.upsertSession(createSessionInput(
            'hash-stats-archived',
            'User: keep durable defaults.\nAssistant: archive the rest.',
            'Keep durable defaults.',
        ));

        sqliteForTest.prepare(`
            INSERT INTO imported_sessions (
                uuid,
                source_tool,
                source_path,
                external_session_id,
                title,
                session_format,
                transcript,
                excerpt,
                working_directory,
                transcript_hash,
                normalized_session,
                metadata,
                transcript_archive_path,
                transcript_metadata_archive_path,
                transcript_archive_format,
                transcript_stored_bytes,
                discovered_at,
                imported_at,
                last_modified_at,
                created_at,
                updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `).run(
            'inline-session',
            'antigravity',
            'C:\\temp\\inline.jsonl',
            'inline-ext',
            'Inline import',
            'jsonl',
            'User: keep this inline for migration stats.',
            'User: keep this inline for migration stats.',
            'C:\\temp',
            'inline-hash',
            JSON.stringify({ sourceTool: 'antigravity', transcriptHash: 'inline-hash' }),
            JSON.stringify({ legacy: true }),
            null,
            null,
            null,
            null,
            1_700_000_000_000,
            1_700_000_000_100,
            1_700_000_000_050,
            1_700_000_000_100,
            1_700_000_000_100,
        );

        sqliteForTest.prepare('UPDATE imported_sessions SET metadata = ? WHERE uuid = ?').run(
            JSON.stringify({ antigravityImportSurface: 'experimental' }),
            archived.id,
        );

        const stats = store.getMaintenanceStats();

        expect(stats).toEqual({
            totalSessions: 2,
            inlineTranscriptCount: 1,
            archivedTranscriptCount: 1,
            missingRetentionSummaryCount: 2,
        });
    });

    it('backfills retention summaries for legacy archived sessions', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const archiveRoot = await createTempRoot();
        const store = new ImportedSessionStore(archiveRoot);

        const created = store.upsertSession(createSessionInput(
            'hash-legacy-retention',
            'User: keep durable defaults.\nAssistant: archive the rest.',
            'Keep durable defaults.',
        ));

        sqliteForTest.prepare('UPDATE imported_sessions SET metadata = ? WHERE uuid = ?').run(
            JSON.stringify({ antigravityImportSurface: 'experimental' }),
            created.id,
        );

        const updated = store.backfillRetentionSummaries((session) => ({
            archiveDisposition: 'archive_only',
            summary: `Backfilled for ${session.sourceTool}`,
            strategy: 'heuristic',
        }), 10);

        const fetched = store.getImportedSession(created.id);
        const row = sqliteForTest
            .prepare('SELECT transcript_metadata_archive_path, metadata FROM imported_sessions WHERE uuid = ?')
            .get(created.id) as Record<string, unknown>;
        const metadataPath = path.join(archiveRoot, ...String(row.transcript_metadata_archive_path ?? '').split('/'));
        const archiveMetadata = JSON.parse(
            zlib.gunzipSync(await fs.readFile(metadataPath)).toString('utf-8'),
        ) as Record<string, unknown>;

        expect(updated).toBe(1);
        expect(fetched?.metadata).toMatchObject({
            retentionSummary: {
                archiveDisposition: 'archive_only',
                summary: 'Backfilled for antigravity',
            },
        });
        expect(archiveMetadata.retentionSummary).toMatchObject({
            archiveDisposition: 'archive_only',
            summary: 'Backfilled for antigravity',
        });
    });
});
