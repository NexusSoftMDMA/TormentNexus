import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import Database from 'better-sqlite3';

const WORKSPACE_ROOT = 'c:\\Users\\hyper\\workspace';
const GLOBAL_DB_PATH = path.join(WORKSPACE_ROOT, 'tormentnexus', 'tormentnexus.db');

function ensureDir(dirPath) {
    if (!fs.existsSync(dirPath)) {
        fs.mkdirSync(dirPath, { recursive: true });
    }
}

// Table schema initializers for the project-specific database
function initializeProjectSchema(projectDb) {
    projectDb.pragma("foreign_keys = ON");
    projectDb.exec(`
        CREATE TABLE IF NOT EXISTS imported_sessions (
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
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS imported_session_memories (
            uuid TEXT PRIMARY KEY,
            imported_session_uuid TEXT NOT NULL,
            memory_index INTEGER NOT NULL,
            kind TEXT NOT NULL DEFAULT 'memory',
            content TEXT NOT NULL,
            tags TEXT NOT NULL DEFAULT '[]',
            source TEXT NOT NULL DEFAULT 'heuristic',
            metadata TEXT NOT NULL DEFAULT '{}',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (imported_session_uuid) REFERENCES imported_sessions(uuid) ON DELETE CASCADE,
            UNIQUE(imported_session_uuid, memory_index)
        );

        CREATE INDEX IF NOT EXISTS imported_sessions_transcript_hash_idx ON imported_sessions(transcript_hash);
        CREATE INDEX IF NOT EXISTS imported_sessions_source_tool_idx ON imported_sessions(source_tool);
        CREATE INDEX IF NOT EXISTS imported_sessions_source_path_idx ON imported_sessions(source_path);
        CREATE UNIQUE INDEX IF NOT EXISTS imported_sessions_transcript_hash_unique ON imported_sessions(transcript_hash);

        CREATE INDEX IF NOT EXISTS imported_session_memories_session_idx ON imported_session_memories(imported_session_uuid);
        CREATE INDEX IF NOT EXISTS imported_session_memories_kind_idx ON imported_session_memories(kind);
    `);
}

// Check git remote URL for robertpelloni ownership
function isOwnedByRobertPelloni(dirPath) {
    try {
        const remoteUrl = execSync('git remote get-url origin', { cwd: dirPath, stdio: 'pipe' }).toString().trim();
        return remoteUrl.toLowerCase().includes('robertpelloni');
    } catch {
        return false;
    }
}

// Find all Git repositories recursively up to maxDepth
function findGitRepositories(dirPath, currentDepth = 1, maxDepth = 3) {
    const repos = [];
    if (currentDepth > maxDepth) return repos;

    try {
        const entries = fs.readdirSync(dirPath, { withFileTypes: true });
        
        // Check if current directory has a git repository
        const hasGit = entries.some(e => e.name === '.git');
        if (hasGit) {
            repos.push(dirPath);
        }

        // Recursively check children
        for (const entry of entries) {
            if (entry.isDirectory() && !entry.name.startsWith('.') && entry.name !== 'node_modules') {
                const childPath = path.join(dirPath, entry.name);
                repos.push(...findGitRepositories(childPath, currentDepth + 1, maxDepth));
            }
        }
    } catch {}

    return repos;
}

async function run() {
    console.log('==================================================');
    console.log('📦 ROBERTPELLONI RECURSIVE DATABASE SPLITTER');
    console.log('==================================================');
    
    if (!fs.existsSync(GLOBAL_DB_PATH)) {
        console.error(`Global database not found at ${GLOBAL_DB_PATH}. Exiting.`);
        process.exit(1);
    }
    
    const globalDb = new Database(GLOBAL_DB_PATH);
    console.log(`Connected to global database: ${GLOBAL_DB_PATH}`);
    
    // Find all Git repositories in the workspace
    console.log(`Scanning ${WORKSPACE_ROOT} recursively for Git repositories...`);
    const allRepos = findGitRepositories(WORKSPACE_ROOT);
    console.log(`Found ${allRepos.length} Git repositories in workspace.`);
    
    // Filter repositories owned by robertpelloni
    const robertRepos = allRepos.filter(repoPath => {
        const isOwned = isOwnedByRobertPelloni(repoPath);
        if (isOwned) {
            console.log(`  ✅ Verified ownership: ${path.relative(WORKSPACE_ROOT, repoPath)}`);
        }
        return isOwned;
    });
    
    console.log(`Found ${robertRepos.length} repositories owned by github/robertpelloni.`);
    
    // Sort repositories by path length descending to process subdirectories before parents
    robertRepos.sort((a, b) => b.length - a.length);
    
    let migratedSessionsTotal = 0;
    let migratedMemoriesTotal = 0;
    
    // Process migrations for each verified repository
    for (const repoPath of robertRepos) {
        const repoName = path.basename(repoPath);
        const relativePath = path.relative(WORKSPACE_ROOT, repoPath);
        
        // Query sessions that belong to this repo or subfolders
        // We normalize backslashes to forward slashes for matching since the DB has mixed formats
        const normalizedRepoPath = repoPath.replace(/\\/g, '/');
        const sessionsToMigrate = globalDb.prepare(`
            SELECT * FROM imported_sessions 
            WHERE lower(replace(working_directory, '\\', '/')) = lower(?) 
               OR lower(replace(working_directory, '\\', '/')) LIKE lower(?)
        `).all(normalizedRepoPath, `${normalizedRepoPath}/%`);
        
        if (sessionsToMigrate.length === 0) {
            continue; // No sessions in global DB to migrate for this repo
        }
        
        console.log(`\n📂 Migrating ${sessionsToMigrate.length} sessions for project: ${relativePath}`);
        
        const tormentnexusDir = path.join(repoPath, '.tormentnexus');
        ensureDir(tormentnexusDir);
        
        const projectDbPath = path.join(tormentnexusDir, 'project.db');
        const projectDb = new Database(projectDbPath);
        initializeProjectSchema(projectDb);
        
        const checkSession = projectDb.prepare(`SELECT updated_at FROM imported_sessions WHERE uuid = ?`);
        const insertSession = projectDb.prepare(`
            INSERT INTO imported_sessions (
                uuid, source_tool, source_path, source_size, source_mtime,
                external_session_id, title, session_format, transcript, excerpt,
                working_directory, transcript_hash, transcript_archive_path,
                transcript_metadata_archive_path, transcript_archive_format,
                transcript_stored_bytes, normalized_session, metadata,
                discovered_at, imported_at, last_modified_at, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `);
        const updateSession = projectDb.prepare(`
            UPDATE imported_sessions SET
                source_tool = ?, source_path = ?, source_size = ?, source_mtime = ?,
                external_session_id = ?, title = ?, session_format = ?, transcript = ?, excerpt = ?,
                working_directory = ?, transcript_hash = ?, transcript_archive_path = ?,
                transcript_metadata_archive_path = ?, transcript_archive_format = ?,
                transcript_stored_bytes = ?, normalized_session = ?, metadata = ?,
                discovered_at = ?, imported_at = ?, last_modified_at = ?, created_at = ?, updated_at = ?
            WHERE uuid = ?
        `);
        
        const insertMemory = projectDb.prepare(`
            INSERT OR IGNORE INTO imported_session_memories (
                uuid, imported_session_uuid, memory_index, kind, content, tags, source, metadata, created_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `);
        
        let migratedSessions = 0;
        let migratedMemories = 0;
        
        for (const session of sessionsToMigrate) {
            // Fetch memories from global DB
            const memories = globalDb.prepare('SELECT * FROM imported_session_memories WHERE imported_session_uuid = ?').all(session.uuid);
            
            // Check existing session
            const existingSession = checkSession.get(session.uuid);
            
            const sessionArgs = [
                session.uuid, session.source_tool, session.source_path, session.source_size, session.source_mtime,
                session.external_session_id, session.title, session.session_format, session.transcript, session.excerpt,
                session.working_directory, session.transcript_hash, session.transcript_archive_path,
                session.transcript_metadata_archive_path, session.transcript_archive_format,
                session.transcript_stored_bytes, session.normalized_session, session.metadata,
                session.discovered_at, session.imported_at, session.last_modified_at, session.created_at, session.updated_at
            ];
            
            if (!existingSession) {
                // Insert session in project-specific DB
                insertSession.run(...sessionArgs);
            } else if (session.updated_at > existingSession.updated_at) {
                // Global session is newer, update it
                const updateArgs = [...sessionArgs.slice(1), session.uuid]; // shift uuid to end
                updateSession.run(...updateArgs);
            }
            
            // Insert memories in project-specific DB
            for (const memory of memories) {
                insertMemory.run(
                    memory.uuid,
                    memory.imported_session_uuid,
                    memory.memory_index,
                    memory.kind,
                    memory.content,
                    memory.tags,
                    memory.source,
                    memory.metadata,
                    memory.created_at
                );
                migratedMemories++;
                migratedMemoriesTotal++;
            }
            
            // Remove from global database to complete the split
            globalDb.prepare('DELETE FROM imported_session_memories WHERE imported_session_uuid = ?').run(session.uuid);
            globalDb.prepare('DELETE FROM imported_sessions WHERE uuid = ?').run(session.uuid);
            
            migratedSessions++;
            migratedSessionsTotal++;
        }
        
        projectDb.close();
        console.log(`  - Persisted ${migratedSessions} sessions & ${migratedMemories} memories in project.db`);
        
        // Force Stage in Git
        console.log(`  - Staging project.db in project's Git repository`);
        const relDbPath = path.join('.tormentnexus', 'project.db');
        execSync(`git add -f ${relDbPath}`, { cwd: repoPath });
    }
    
    // Vacuum global database to recover space
    console.log('\n🧹 Vacuuming global database...');
    globalDb.exec('VACUUM');
    globalDb.close();
    
    console.log('==================================================');
    console.log('🎉 RECURSIVE DATABASE SPLIT COMPLETE');
    console.log('==================================================');
    console.log(`Total Sessions Migrated to Projects:  ${migratedSessionsTotal}`);
    console.log(`Total Memories Migrated to Projects:  ${migratedMemoriesTotal}`);
    console.log('All local project.db instances are successfully Git-tracked.');
    console.log('==================================================');
}

run();
