const Database = require('better-sqlite3');
const globalDb = new Database('tormentnexus.db');
const projectDb = new Database('../.tormentnexus/project.db');

const sessions = projectDb.prepare('SELECT * FROM imported_sessions').all();
const memories = projectDb.prepare('SELECT * FROM imported_session_memories').all();

const insertSession = globalDb.prepare(`
    INSERT OR REPLACE INTO imported_sessions (
        uuid, source_tool, source_path, source_size, source_mtime,
        external_session_id, title, session_format, transcript, excerpt,
        working_directory, transcript_hash, transcript_archive_path,
        transcript_metadata_archive_path, transcript_archive_format,
        transcript_stored_bytes, normalized_session, metadata,
        discovered_at, imported_at, last_modified_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`);

for (const session of sessions) {
    insertSession.run(
        session.uuid, session.source_tool, session.source_path, session.source_size, session.source_mtime,
        session.external_session_id, session.title, session.session_format, session.transcript, session.excerpt,
        session.working_directory, session.transcript_hash, session.transcript_archive_path,
        session.transcript_metadata_archive_path, session.transcript_archive_format,
        session.transcript_stored_bytes, session.normalized_session, session.metadata,
        session.discovered_at, session.imported_at, session.last_modified_at, session.created_at, session.updated_at
    );
}

const insertMemory = globalDb.prepare(`
    INSERT OR REPLACE INTO imported_session_memories (
        uuid, imported_session_uuid, memory_index, kind, content, tags, source, metadata, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`);

for (const memory of memories) {
    insertMemory.run(
        memory.uuid, memory.imported_session_uuid, memory.memory_index, memory.kind, memory.content, memory.tags, memory.source, memory.metadata, memory.created_at
    );
}

console.log('Restored ' + sessions.length + ' sessions and ' + memories.length + ' memories.');
