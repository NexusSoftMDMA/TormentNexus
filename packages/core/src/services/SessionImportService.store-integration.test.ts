import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import zlib from 'zlib';

import Database from 'better-sqlite3';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

let sqliteForTest: Database.Database;
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
type SessionImportServiceModule = typeof import('./SessionImportService.js');

const tempRoots: string[] = [];

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
    const root = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-session-import-store-'));
    tempRoots.push(root);
    return root;
}

describe('SessionImportService with ImportedSessionStore', () => {
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

    it('persists discovered sessions into the real imported-session store and writes instruction docs', async () => {
        if (!sqliteAvailable) {
            return;
        }
        const root = await createTempRoot();
        const fakeHome = await createTempRoot();
        vi.spyOn(os, 'homedir').mockReturnValue(fakeHome);
        vi.stubEnv('APPDATA', fakeHome);
        vi.stubEnv('LOCALAPPDATA', fakeHome);
        const antigravityBrainDir = path.join(fakeHome, '.gemini', 'antigravity', 'brain');
        await fs.mkdir(antigravityBrainDir, { recursive: true });
        await fs.writeFile(
            path.join(antigravityBrainDir, 'session.jsonl'),
            [
                JSON.stringify({ request: { message: 'Keep Antigravity import explicitly experimental.' } }),
                JSON.stringify({ response: { message: 'Use the reverse-engineered brain root until a stable contract exists.' } }),
            ].join('\n'),
            'utf-8',
        );

        const addLongTerm = vi.fn(async () => ({}));
        const captureSessionSummary = vi.fn(async () => ({}));
        const archiveRoot = path.join(root, '.tormentnexus', 'imported_sessions', 'archive');

        const { ImportedSessionStore } = await import('./ImportedSessionStore.js') as ImportedSessionStoreModule;
        const { SessionImportService } = await import('./SessionImportService.js') as SessionImportServiceModule;
        const store = new ImportedSessionStore(archiveRoot);
        const service = new SessionImportService({
            generateText: vi.fn(async () => {
                throw new Error('no llm');
            }),
        } as any, {
            addLongTerm,
            captureSessionSummary,
        } as any, root, {
            store,
            includeHomeDirectories: true,
            maxFilesPerRoot: 20,
        });

        const result = await service.scanAndImport();
        const sessions = store.listImportedSessions(10);
        const instructionDocs = await service.listInstructionDocs();

        expect(result.discoveredCount).toBe(1);
        expect(result.importedCount).toBe(1);
        expect(result.tools).toContain('antigravity');
        expect(sessions).toHaveLength(1);
        expect(sessions[0]?.sourceTool).toBe('antigravity');
        expect(sessions[0]?.transcript).toContain('Keep Antigravity import explicitly experimental.');
        expect(sessions[0]?.metadata).toMatchObject({
            antigravityImportSurface: 'experimental',
            antigravityDiscoveryRoot: 'brain',
            antigravitySource: 'reverse-engineered',
            retentionSummary: {
                archiveDisposition: 'archive_only',
                strategy: 'heuristic',
            },
        });
        expect(sessions[0]?.parsedMemories.some((memory) => memory.kind === 'instruction')).toBe(true);
        const archiveSidecars = await findFiles(archiveRoot, '.meta.json.gz');
        expect(archiveSidecars.length).toBeGreaterThan(0);
        const sidecar = JSON.parse(
            zlib.gunzipSync(await fs.readFile(archiveSidecars[0]!)).toString('utf-8'),
        ) as Record<string, unknown>;
        expect(sidecar.retentionSummary).toMatchObject({
            archiveDisposition: 'archive_only',
        });
        expect(instructionDocs).toHaveLength(1);
        expect(addLongTerm).toHaveBeenCalled();
        expect(captureSessionSummary).toHaveBeenCalledTimes(1);
    });
});

async function findFiles(root: string, suffix: string): Promise<string[]> {
    const entries = await fs.readdir(root, { withFileTypes: true });
    const files = await Promise.all(entries.map(async (entry) => {
        const fullPath = path.join(root, entry.name);
        if (entry.isDirectory()) {
            return findFiles(fullPath, suffix);
        }

        return entry.name.endsWith(suffix) ? [fullPath] : [];
    }));

    return files.flat();
}
