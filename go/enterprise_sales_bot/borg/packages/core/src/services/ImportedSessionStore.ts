import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import zlib from 'zlib';

import { sqliteInstance } from '../db/index.js';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../db/sqliteAvailability.js';

export type ImportedSessionMemoryKind = 'memory' | 'instruction';
export type ImportedSessionMemorySource = 'llm' | 'heuristic';

export interface ImportedSessionMemoryInput {
    kind: ImportedSessionMemoryKind;
    content: string;
    tags: string[];
    source: ImportedSessionMemorySource;
    metadata?: Record<string, unknown>;
}

export interface ImportedSessionRecordInput {
    sourceTool: string;
    sourcePath: string;
    sourceSize?: number | null;
    sourceMtime?: number | null;
    externalSessionId?: string | null;
    title?: string | null;
    sessionFormat: string;
    transcript: string;
    excerpt?: string | null;
    workingDirectory?: string | null;
    transcriptHash: string;
    normalizedSession: Record<string, unknown>;
    metadata?: Record<string, unknown>;
    discoveredAt: number;
    importedAt: number;
    lastModifiedAt?: number | null;
    parsedMemories: ImportedSessionMemoryInput[];
}

export interface ImportedSessionMemoryRecord extends ImportedSessionMemoryInput {
    id: string;
    importedSessionId: string;
    createdAt: number;
}

export interface ImportedSessionRecord {
    id: string;
    sourceTool: string;
    sourcePath: string;
    sourceSize: number | null;
    sourceMtime: number | null;
    externalSessionId: string | null;
    title: string | null;
    sessionFormat: string;
    transcript: string;
    excerpt: string | null;
    workingDirectory: string | null;
    transcriptHash: string;
    normalizedSession: Record<string, unknown>;
    metadata: Record<string, unknown>;
    discoveredAt: number;
    importedAt: number;
    lastModifiedAt: number | null;
    createdAt: number;
    updatedAt: number;
    parsedMemories: ImportedSessionMemoryRecord[];
}

export interface ImportedSessionMaintenanceStats {
    totalSessions: number;
    inlineTranscriptCount: number;
    archivedTranscriptCount: number;
    missingRetentionSummaryCount: number;
}

type ImportedSessionRow = Record<string, unknown> & {
    transcript_archive_path?: string | null;
    transcript_metadata_archive_path?: string | null;
    transcript_archive_format?: string | null;
    transcript_stored_bytes?: number | null;
};

interface TranscriptArchiveInfo {
    transcriptArchivePath: string;
    transcriptMetadataArchivePath: string;
    transcriptArchiveFormat: string;
    transcriptStoredBytes: number;
}

function safeJsonParse<T>(value: unknown, fallback: T): T {
    if (typeof value !== 'string' || !value.trim()) {
        return fallback;
    }

    try {
        return JSON.parse(value) as T;
    } catch {
        return fallback;
    }
}

function mapMemoryRow(row: Record<string, unknown>): ImportedSessionMemoryRecord {
    return {
        id: String(row.uuid),
        importedSessionId: String(row.imported_session_uuid),
        kind: row.kind === 'instruction' ? 'instruction' : 'memory',
        content: String(row.content ?? ''),
        tags: safeJsonParse<string[]>(row.tags, []),
        source: row.source === 'llm' ? 'llm' : 'heuristic',
        metadata: safeJsonParse<Record<string, unknown>>(row.metadata, {}),
        createdAt: Number(row.created_at ?? 0),
    };
}

function ensureDir(dirPath: string): void {
    fs.mkdirSync(dirPath, { recursive: true });
}

function toRelativeArchivePath(root: string, filePath: string): string {
    return path.relative(root, filePath).split(path.sep).join('/');
}

export class ImportedSessionStore {
    private archiveRoot: string;

    constructor(archiveRoot: string = path.join(process.cwd(), '.tormentnexus', 'imported_sessions', 'archive')) {
        this.archiveRoot = archiveRoot;
        ensureDir(this.archiveRoot);
    }

    private getAbsoluteArchivePath(relativePath: string | null | undefined): string | null {
        if (typeof relativePath !== 'string' || !relativePath.trim()) {
            return null;
        }

        return path.join(this.archiveRoot, ...relativePath.split(/[\\/]+/));
    }

    private readArchivedTranscript(row: ImportedSessionRow): string {
        const archivePath = this.getAbsoluteArchivePath(
            typeof row.transcript_archive_path === 'string' ? row.transcript_archive_path : null,
        );

        if (archivePath && fs.existsSync(archivePath)) {
            return zlib.gunzipSync(fs.readFileSync(archivePath)).toString('utf-8');
        }

        return String(row.transcript ?? '');
    }

    private buildArchiveMetadata(input: ImportedSessionRecordInput): Record<string, unknown> {
        return {
            sourceTool: input.sourceTool,
            sourcePath: input.sourcePath,
            sessionFormat: input.sessionFormat,
            transcriptHash: input.transcriptHash,
            title: input.title ?? null,
            workingDirectory: input.workingDirectory ?? null,
            transcriptLength: input.transcript.length,
            excerpt: input.excerpt ?? null,
            durableMemoryCount: input.parsedMemories.length,
            durableInstructionCount: input.parsedMemories.filter((memory) => memory.kind === 'instruction').length,
            memoryTags: Array.from(new Set(input.parsedMemories.flatMap((memory) => memory.tags ?? []))).slice(0, 24),
            retentionSummary: input.metadata?.retentionSummary ?? null,
            archivedAt: Date.now(),
        };
    }

    private writeTranscriptArchive(sessionId: string, input: ImportedSessionRecordInput): TranscriptArchiveInfo {
        // Flattened structure to avoid thousands of subdirectories
        const archiveDir = path.join(this.archiveRoot, 'sessions');
        ensureDir(archiveDir);

        const transcriptFile = path.join(archiveDir, `${input.transcriptHash}.txt.gz`);
        const metadataFile = path.join(archiveDir, `${input.transcriptHash}.meta.json.gz`);
        const transcriptBuffer = Buffer.from(input.transcript, 'utf-8');
        const compressedTranscript = zlib.gzipSync(transcriptBuffer, { level: 9 });
        const compressedMetadata = zlib.gzipSync(
            Buffer.from(
                JSON.stringify(
                    {
                        sessionId,
                        ...this.buildArchiveMetadata(input),
                    },
                    null,
                    2,
                ),
                'utf-8',
            ),
            { level: 9 },
        );

        fs.writeFileSync(transcriptFile, compressedTranscript);
        fs.writeFileSync(metadataFile, compressedMetadata);

        return {
            transcriptArchivePath: toRelativeArchivePath(this.archiveRoot, transcriptFile),
            transcriptMetadataArchivePath: toRelativeArchivePath(this.archiveRoot, metadataFile),
            transcriptArchiveFormat: 'gzip-text-v1',
            transcriptStoredBytes: transcriptBuffer.byteLength,
        };
    }

    private mapSessionRow(row: ImportedSessionRow, parsedMemories: ImportedSessionMemoryRecord[]): ImportedSessionRecord {
        return {
            id: String(row.uuid),
            sourceTool: String(row.source_tool ?? ''),
            sourcePath: String(row.source_path ?? ''),
            sourceSize: row.source_size == null ? null : Number(row.source_size),
            sourceMtime: row.source_mtime == null ? null : Number(row.source_mtime),
            externalSessionId: typeof row.external_session_id === 'string' ? row.external_session_id : null,
            title: typeof row.title === 'string' ? row.title : null,
            sessionFormat: String(row.session_format ?? 'generic'),
            transcript: this.readArchivedTranscript(row),
            excerpt: typeof row.excerpt === 'string' ? row.excerpt : null,
            workingDirectory: typeof row.working_directory === 'string' ? row.working_directory : null,
            transcriptHash: String(row.transcript_hash ?? ''),
            normalizedSession: safeJsonParse<Record<string, unknown>>(row.normalized_session, {}),
            metadata: safeJsonParse<Record<string, unknown>>(row.metadata, {}),
            discoveredAt: Number(row.discovered_at ?? 0),
            importedAt: Number(row.imported_at ?? 0),
            lastModifiedAt: row.last_modified_at == null ? null : Number(row.last_modified_at),
            createdAt: Number(row.created_at ?? 0),
            updatedAt: Number(row.updated_at ?? 0),
            parsedMemories,
        };
    }

    hasTranscriptHash(transcriptHash: string): boolean {
        try {
            const row = sqliteInstance
                .prepare('SELECT uuid FROM imported_sessions WHERE transcript_hash = ? LIMIT 1')
                .get(transcriptHash) as Record<string, unknown> | undefined;

            return Boolean(row?.uuid);
        } catch (error) {
            if (isSqliteUnavailableError(error)) {
                console.warn(formatOptionalSqliteFailure(
                    '[ImportedSessionStore] Transcript deduplication is unavailable',
                    error,
                ));
                return false;
            }

            throw error;
        }
    }

    hasMatchingSource(sourcePath: string, size: number, mtime: number): boolean {
        try {
            const row = sqliteInstance
                .prepare('SELECT uuid FROM imported_sessions WHERE source_path = ? AND source_size = ? AND source_mtime = ? LIMIT 1')
                .get(sourcePath, size, mtime) as Record<string, unknown> | undefined;

            return Boolean(row?.uuid);
        } catch (error) {
            return false;
        }
    }

    compactInlineTranscripts(limit: number = 100): number {
        const rows = sqliteInstance.prepare(`
            SELECT *
            FROM imported_sessions
            WHERE (transcript_archive_path IS NULL OR transcript_archive_path = '')
              AND LENGTH(transcript) > 0
            ORDER BY imported_at ASC
            LIMIT ?
        `).all(limit) as ImportedSessionRow[];

        let compacted = 0;
        const updateStatement = sqliteInstance.prepare(`
            UPDATE imported_sessions
            SET transcript = '',
                transcript_archive_path = ?,
                transcript_metadata_archive_path = ?,
                transcript_archive_format = ?,
                transcript_stored_bytes = ?,
                updated_at = ?
            WHERE uuid = ?
        `);

        for (const row of rows) {
            const transcript = String(row.transcript ?? '');
            if (!transcript.trim()) {
                continue;
            }

            const input: ImportedSessionRecordInput = {
                sourceTool: String(row.source_tool ?? ''),
                sourcePath: String(row.source_path ?? ''),
                externalSessionId: typeof row.external_session_id === 'string' ? row.external_session_id : null,
                title: typeof row.title === 'string' ? row.title : null,
                sessionFormat: String(row.session_format ?? 'generic'),
                transcript,
                excerpt: typeof row.excerpt === 'string' ? row.excerpt : null,
                workingDirectory: typeof row.working_directory === 'string' ? row.working_directory : null,
                transcriptHash: String(row.transcript_hash ?? ''),
                normalizedSession: safeJsonParse<Record<string, unknown>>(row.normalized_session, {}),
                metadata: safeJsonParse<Record<string, unknown>>(row.metadata, {}),
                discoveredAt: Number(row.discovered_at ?? 0),
                importedAt: Number(row.imported_at ?? 0),
                lastModifiedAt: row.last_modified_at == null ? null : Number(row.last_modified_at),
                parsedMemories: [],
            };
            const archiveInfo = this.writeTranscriptArchive(String(row.uuid), input);
            updateStatement.run(
                archiveInfo.transcriptArchivePath,
                archiveInfo.transcriptMetadataArchivePath,
                archiveInfo.transcriptArchiveFormat,
                archiveInfo.transcriptStoredBytes,
                Date.now(),
                String(row.uuid),
            );
            compacted += 1;
        }

        return compacted;
    }

    backfillRetentionSummaries(
        buildRetentionSummary: (session: ImportedSessionRecord) => Record<string, unknown>,
        limit: number = 100,
    ): number {
        const rows = sqliteInstance.prepare(`
            SELECT uuid
            FROM imported_sessions
            WHERE json_extract(metadata, '$.retentionSummary') IS NULL
            ORDER BY imported_at ASC
            LIMIT ?
        `).all(limit) as Array<{ uuid?: string }>;

        let updated = 0;
        for (const row of rows) {
            const sessionId = typeof row.uuid === 'string' ? row.uuid : '';
            if (!sessionId) {
                continue;
            }

            const session = this.getImportedSession(sessionId);
            if (!session) {
                continue;
            }

            const metadata = {
                ...(session.metadata ?? {}),
                retentionSummary: buildRetentionSummary(session),
            };

            this.upsertSession({
                sourceTool: session.sourceTool,
                sourcePath: session.sourcePath,
                externalSessionId: session.externalSessionId,
                title: session.title,
                sessionFormat: session.sessionFormat,
                transcript: session.transcript,
                excerpt: session.excerpt,
                workingDirectory: session.workingDirectory,
                transcriptHash: session.transcriptHash,
                normalizedSession: session.normalizedSession,
                metadata,
                discoveredAt: session.discoveredAt,
                importedAt: session.importedAt,
                lastModifiedAt: session.lastModifiedAt,
                parsedMemories: session.parsedMemories.map((memory) => ({
                    kind: memory.kind,
                    content: memory.content,
                    tags: memory.tags,
                    source: memory.source,
                    metadata: memory.metadata,
                })),
            });
            updated += 1;
        }

        return updated;
    }

    upsertSession(input: ImportedSessionRecordInput): ImportedSessionRecord {
        const now = Date.now();
        const existing = sqliteInstance
            .prepare('SELECT uuid FROM imported_sessions WHERE transcript_hash = ? LIMIT 1')
            .get(input.transcriptHash) as Record<string, unknown> | undefined;

        const sessionId = typeof existing?.uuid === 'string' ? existing.uuid : crypto.randomUUID();
        const metadata = JSON.stringify(input.metadata ?? {});
        const normalizedSession = JSON.stringify(input.normalizedSession);
        const archiveInfo = this.writeTranscriptArchive(sessionId, input);

        sqliteInstance.prepare(`
            INSERT INTO imported_sessions (
                uuid,
                source_tool,
                source_path,
                source_size,
                source_mtime,
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
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(uuid) DO UPDATE SET
                source_tool = excluded.source_tool,
                source_path = excluded.source_path,
                source_size = excluded.source_size,
                source_mtime = excluded.source_mtime,
                external_session_id = excluded.external_session_id,
                title = excluded.title,
                session_format = excluded.session_format,
                transcript = excluded.transcript,
                excerpt = excluded.excerpt,
                working_directory = excluded.working_directory,
                transcript_hash = excluded.transcript_hash,
                normalized_session = excluded.normalized_session,
                metadata = excluded.metadata,
                transcript_archive_path = excluded.transcript_archive_path,
                transcript_metadata_archive_path = excluded.transcript_metadata_archive_path,
                transcript_archive_format = excluded.transcript_archive_format,
                transcript_stored_bytes = excluded.transcript_stored_bytes,
                discovered_at = excluded.discovered_at,
                imported_at = excluded.imported_at,
                last_modified_at = excluded.last_modified_at,
                updated_at = excluded.updated_at
        `).run(
            sessionId,
            input.sourceTool,
            input.sourcePath,
            input.sourceSize ?? null,
            input.sourceMtime ?? null,
            input.externalSessionId ?? null,
            input.title ?? null,
            input.sessionFormat,
            '',
            input.excerpt ?? null,
            input.workingDirectory ?? null,
            input.transcriptHash,
            normalizedSession,
            metadata,
            archiveInfo.transcriptArchivePath,
            archiveInfo.transcriptMetadataArchivePath,
            archiveInfo.transcriptArchiveFormat,
            archiveInfo.transcriptStoredBytes,
            input.discoveredAt,
            input.importedAt,
            input.lastModifiedAt ?? null,
            existing ? now : input.importedAt,
            now,
        );

        sqliteInstance
            .prepare('DELETE FROM imported_session_memories WHERE imported_session_uuid = ?')
            .run(sessionId);

        const insertMemory = sqliteInstance.prepare(`
            INSERT INTO imported_session_memories (
                uuid,
                imported_session_uuid,
                memory_index,
                kind,
                content,
                tags,
                source,
                metadata,
                created_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `);

        input.parsedMemories.forEach((memory, index) => {
            insertMemory.run(
                crypto.randomUUID(),
                sessionId,
                index,
                memory.kind,
                memory.content,
                JSON.stringify(memory.tags ?? []),
                memory.source,
                JSON.stringify(memory.metadata ?? {}),
                now,
            );
        });

        const session = this.getImportedSession(sessionId);
        if (!session) {
            throw new Error(`Imported session '${sessionId}' was not persisted.`);
        }

        return session;
    }

    listImportedSessions(limit: number = 50): ImportedSessionRecord[] {
        const rows = sqliteInstance.prepare(`
            SELECT *
            FROM imported_sessions
            ORDER BY imported_at DESC
            LIMIT ?
        `).all(limit) as ImportedSessionRow[];

        return rows.map((row) => this.mapSessionRow(row, this.listParsedMemories(String(row.uuid))));
    }

    getMaintenanceStats(): ImportedSessionMaintenanceStats {
        const row = sqliteInstance.prepare(`
            SELECT
                COUNT(*) AS totalSessions,
                SUM(CASE WHEN LENGTH(transcript) > 0 THEN 1 ELSE 0 END) AS inlineTranscriptCount,
                SUM(CASE WHEN transcript_archive_path IS NOT NULL AND transcript_archive_path != '' THEN 1 ELSE 0 END) AS archivedTranscriptCount,
                SUM(CASE WHEN json_extract(metadata, '$.retentionSummary') IS NULL THEN 1 ELSE 0 END) AS missingRetentionSummaryCount
            FROM imported_sessions
        `).get() as Record<string, unknown> | undefined;

        return {
            totalSessions: Number(row?.totalSessions ?? 0),
            inlineTranscriptCount: Number(row?.inlineTranscriptCount ?? 0),
            archivedTranscriptCount: Number(row?.archivedTranscriptCount ?? 0),
            missingRetentionSummaryCount: Number(row?.missingRetentionSummaryCount ?? 0),
        };
    }

    listParsedMemories(importedSessionId: string): ImportedSessionMemoryRecord[] {
        const rows = sqliteInstance.prepare(`
            SELECT *
            FROM imported_session_memories
            WHERE imported_session_uuid = ?
            ORDER BY memory_index ASC
        `).all(importedSessionId) as Record<string, unknown>[];

        return rows.map(mapMemoryRow);
    }

    getImportedSession(id: string): ImportedSessionRecord | null {
        const row = sqliteInstance.prepare(`
            SELECT *
            FROM imported_sessions
            WHERE uuid = ?
            LIMIT 1
        `).get(id) as ImportedSessionRow | undefined;

        if (!row) {
            return null;
        }

        return this.mapSessionRow(row, this.listParsedMemories(id));
    }

    listInstructionMemories(limit: number = 200): ImportedSessionMemoryRecord[] {
        const rows = sqliteInstance.prepare(`
            SELECT *
            FROM imported_session_memories
            WHERE kind = 'instruction'
            ORDER BY created_at DESC
            LIMIT ?
        `).all(limit) as Record<string, unknown>[];

        return rows.map(mapMemoryRow);
    }
}
