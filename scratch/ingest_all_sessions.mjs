import fs from 'fs';
import path from 'path';
import crypto from 'crypto';
import zlib from 'zlib';
import Database from 'better-sqlite3';

const HOME_DIR = 'C:\\Users\\hyper';
const WORKSPACE_DIR = 'c:\\Users\\hyper\\workspace\\tormentnexus';
const DB_PATH = path.join(WORKSPACE_DIR, 'tormentnexus.db');
const ARCHIVE_ROOT = path.join(WORKSPACE_DIR, '.tormentnexus', 'imported_sessions', 'archive');
const MAX_FILE_SIZE = 25 * 1024 * 1024; // 25MB safety cap

function ensureDir(dirPath) {
    if (!fs.existsSync(dirPath)) {
        fs.mkdirSync(dirPath, { recursive: true });
    }
}

// Clean ANSI sequences, PowerShell progress bars, carriage returns, etc.
function cleanTranscript(rawText) {
    if (!rawText) return '';
    
    // Strip ANSI escape sequences
    let cleaned = rawText.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '');
    
    // Strip PowerShell terminal window title updates
    cleaned = cleaned.replace(/\x07/g, '');
    cleaned = cleaned.replace(/\]0;[^\r\n]*/g, '');
    
    // Normalize newlines and whitespace
    cleaned = cleaned.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
    
    // Remove sequences of dots or spaces that are part of rendering artifacts
    cleaned = cleaned.split('\n')
        .map(line => {
            // Trim whitespace
            let l = line.trim();
            // If it's a PowerShell prompt artifact or duplicate command echo lines, clean it up
            l = l.replace(/^❯\s*/, '');
            return l;
        })
        .filter(line => {
            // Filter out empty lines or pure terminal progress drawing lines
            if (!line) return false;
            if (/^[░▒▓█\-\.\=\s]+$/.test(line)) return false;
            return true;
        })
        .join('\n');
        
    return cleaned.trim();
}

// Parse transcript to find reference to working directories or project folders
function detectProject(filePath, cleanedText) {
    const fileName = path.basename(filePath).toLowerCase();
    
    // Check path itself first
    if (filePath.includes('workspace\\tormentnexus') || filePath.includes('.tormentnexus')) {
        return 'tormentnexus';
    }
    
    // Try to match paths in the transcript
    // e.g. cd 'c:\Users\hyper\workspace\<project>'
    const cdRegex = /cd\s+['"]?c:\\Users\\hyper\\workspace\\([^'"\n>\\/]+)/gi;
    let match;
    const detectedProjects = new Set();
    
    while ((match = cdRegex.exec(cleanedText)) !== null) {
        if (match[1]) {
            detectedProjects.add(match[1].trim());
        }
    }
    
    // Also check for workspace prompts like "chamber.Law.Backend"
    const promptRegex = /c:\\Users\\hyper\\workspace\\([^'"\n>\\/]+)/gi;
    while ((match = promptRegex.exec(cleanedText)) !== null) {
        if (match[1]) {
            detectedProjects.add(match[1].trim());
        }
    }
    
    if (detectedProjects.size > 0) {
        // Return first or most common project found
        return Array.from(detectedProjects)[0];
    }
    
    // Heuristic based on filename hints
    if (fileName.includes('aider')) return 'aider-session';
    if (fileName.includes('claude')) return 'claude-session';
    if (fileName.includes('cursor')) return 'cursor-session';
    
    return 'default-project';
}

function classifyMemoryKind(content) {
    const lowered = content.toLowerCase();
    return /(always|never|prefer|should|must|avoid|remember to|do not|don't|use\b)/.test(lowered)
        ? 'instruction'
        : 'memory';
}

function deriveTags(content, sourceTool) {
    const lowered = content.toLowerCase();
    const tags = [sourceTool, classifyMemoryKind(content)];

    if (/(port|localhost|127\.0\.0\.1|http|ws:|wss:)/.test(lowered)) tags.push('networking');
    if (/(memory|context|session|history)/.test(lowered)) tags.push('memory');
    if (/(sqlite|database|db)/.test(lowered)) tags.push('database');
    if (/(build|typecheck|test|vitest|tsc)/.test(lowered)) tags.push('validation');
    if (/(mcp|tool|server|catalog)/.test(lowered)) tags.push('mcp');
    if (/(dashboard|ui|widget|page)/.test(lowered)) tags.push('ui');

    return Array.from(new Set(tags));
}

function heuristicMemoryExtraction(text, sourceTool) {
    const candidateLines = text
        .split('\n')
        .map(line => line.trim().slice(0, 240))
        .filter(line => line.length >= 24)
        .filter(line => /(use|prefer|should|must|avoid|remember|fixed|fix|discovered|default|path|port|error|warning|supports|requires)/i.test(line));

    const candidateSentences = text
        .split(/(?<=[.!?])\s+/)
        .map(sentence => sentence.trim().slice(0, 220))
        .filter(sentence => sentence.length >= 30)
        .filter(sentence => /(use|prefer|should|must|avoid|remember|fixed|default|path|port|error|warning|supports|requires)/i.test(sentence));

    const facts = Array.from(new Set([...candidateLines, ...candidateSentences])).slice(0, 10);

    return facts.map(fact => ({
        kind: classifyMemoryKind(fact),
        content: fact,
        tags: deriveTags(fact, sourceTool),
        source: 'heuristic',
        metadata: { extraction: 'heuristic' }
    }));
}

async function run() {
    console.log('==================================================');
    console.log('🤖 UNIVERSAL SESSION DETECTOR & INGESTION PIPELINE');
    console.log('==================================================');
    
    if (!fs.existsSync(DB_PATH)) {
        console.error(`Database not found at ${DB_PATH}. Exiting.`);
        process.exit(1);
    }
    
    const db = new Database(DB_PATH);
    console.log(`Connected to tormentnexus.db at ${DB_PATH}`);
    
    // Ensure table structure exists (handled by packages/core/src/db/index.ts, but safe to verify)
    const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all().map(t => t.name);
    if (!tables.includes('imported_sessions') || !tables.includes('imported_session_memories')) {
        console.error('Session import tables are missing! Please make sure migrations have run.');
        process.exit(1);
    }
    
    // Discover all candidate session files
    const candidates = [];
    
    // 1. Scan ~ for PowerShell logs (2026-*.txt)
    console.log(`Scanning home directory ${HOME_DIR} for PowerShell logs...`);
    if (fs.existsSync(HOME_DIR)) {
        const homeFiles = fs.readdirSync(HOME_DIR);
        for (const file of homeFiles) {
            if (file.startsWith('2026-') && file.endsWith('.txt')) {
                candidates.push({
                    filePath: path.join(HOME_DIR, file),
                    sourceTool: 'powershell-shell',
                    sessionFormat: 'terminal-log'
                });
            }
        }
    }
    
    // 2. Scan workspace for markdown exports and local sessions
    console.log(`Scanning workspace ${WORKSPACE_DIR} for untracked session files...`);
    if (fs.existsSync(WORKSPACE_DIR)) {
        const workspaceFiles = fs.readdirSync(WORKSPACE_DIR);
        for (const file of workspaceFiles) {
            if (file.startsWith('chat_export_') && file.endsWith('.md')) {
                candidates.push({
                    filePath: path.join(WORKSPACE_DIR, file),
                    sourceTool: 'antigravity-chat',
                    sessionFormat: 'markdown'
                });
            }
            if (file === '.TormentNexus-session.json' || file === '.tormentnexus-session.json') {
                candidates.push({
                    filePath: path.join(WORKSPACE_DIR, file),
                    sourceTool: file.includes('tormentnexus') ? 'tormentnexus' : 'TormentNexus',
                    sessionFormat: 'json'
                });
            }
        }
    }
    
    console.log(`Discovered ${candidates.length} candidate session files/histories.`);
    
    let newlyIngestedCount = 0;
    let skippedDuplicateCount = 0;
    let skippedLargeCount = 0;
    let newlyFormedMemoriesCount = 0;
    const projectStats = {};
    
    const insertSessionStmt = db.prepare(`
        INSERT INTO imported_sessions (
            uuid, source_tool, source_path, source_size, source_mtime,
            external_session_id, title, session_format, transcript, excerpt,
            working_directory, transcript_hash, transcript_archive_path,
            transcript_metadata_archive_path, transcript_archive_format,
            transcript_stored_bytes, normalized_session, metadata,
            discovered_at, imported_at, last_modified_at, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    
    const insertMemoryStmt = db.prepare(`
        INSERT INTO imported_session_memories (
            uuid, imported_session_uuid, memory_index, kind, content, tags, source, metadata, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    
    for (const candidate of candidates) {
        const { filePath, sourceTool, sessionFormat } = candidate;
        
        try {
            const stat = fs.statSync(filePath);
            if (stat.size > MAX_FILE_SIZE) {
                console.warn(`⚠️ Skipping session too large (>25MB): ${path.basename(filePath)} (${(stat.size / 1024 / 1024).toFixed(1)} MB)`);
                skippedLargeCount++;
                continue;
            }
            
            // Read and clean the transcript
            const rawContent = fs.readFileSync(filePath, 'utf-8');
            let transcript = '';
            
            if (sessionFormat === 'json') {
                try {
                    const parsed = JSON.parse(rawContent);
                    transcript = JSON.stringify(parsed, null, 2);
                } catch {
                    transcript = cleanTranscript(rawContent);
                }
            } else {
                transcript = cleanTranscript(rawContent);
            }
            
            if (!transcript || transcript.length < 50) {
                // Too short to contain meaningful history or prompt information
                continue;
            }
            
            // Calculate transcript hash for strict deduplication
            const transcriptHash = crypto.createHash('sha256').update(transcript).digest('hex');
            
            // Check if hash already exists in DB
            const existing = db.prepare('SELECT uuid FROM imported_sessions WHERE transcript_hash = ? LIMIT 1').get(transcriptHash);
            if (existing) {
                skippedDuplicateCount++;
                continue;
            }
            
            // Detect project workspace and category
            const project = detectProject(filePath, transcript);
            projectStats[project] = (projectStats[project] || 0) + 1;
            
            const uuid = crypto.randomUUID();
            const now = Date.now();
            const title = `${sourceTool.toUpperCase()} Session (${project}) - ${new Date(stat.mtimeMs).toISOString().split('T')[0]}`;
            const excerpt = transcript.slice(0, 300).replace(/\n/g, ' ') + '...';
            
            // Write compressed archive file
            const archiveDir = path.join(ARCHIVE_ROOT, 'sessions');
            ensureDir(archiveDir);
            
            const archiveRelativePath = `sessions/${transcriptHash}.txt.gz`;
            const archiveFullPath = path.join(ARCHIVE_ROOT, archiveRelativePath);
            const compressed = zlib.gzipSync(Buffer.from(transcript, 'utf-8'), { level: 9 });
            fs.writeFileSync(archiveFullPath, compressed);
            
            const metadataObj = {
                sourceTool,
                sourcePath: filePath,
                sessionFormat,
                project,
                contentLength: transcript.length,
                ingestedAt: now
            };
            
            const metadataArchiveRelativePath = `sessions/${transcriptHash}.meta.json.gz`;
            const metadataArchiveFullPath = path.join(ARCHIVE_ROOT, metadataArchiveRelativePath);
            const metadataCompressed = zlib.gzipSync(Buffer.from(JSON.stringify({ uuid, ...metadataObj }, null, 2), 'utf-8'), { level: 9 });
            fs.writeFileSync(metadataArchiveFullPath, metadataCompressed);
            
            // Heuristically extract memories & instructions
            const parsedMemories = heuristicMemoryExtraction(transcript, sourceTool);
            newlyFormedMemoriesCount += parsedMemories.length;
            
            // Insert session
            insertSessionStmt.run(
                uuid,
                sourceTool,
                filePath,
                stat.size,
                stat.mtimeMs,
                path.basename(filePath, path.extname(filePath)),
                title,
                sessionFormat,
                '', // transcript is stored compressed in archive files
                excerpt,
                filePath.includes('workspace') ? WORKSPACE_DIR : HOME_DIR,
                transcriptHash,
                archiveRelativePath,
                metadataArchiveRelativePath,
                'gzip-text-v1',
                compressed.byteLength,
                JSON.stringify(metadataObj),
                JSON.stringify(metadataObj),
                now,
                now,
                stat.mtimeMs,
                Math.floor(stat.birthtimeMs / 1000) || Math.floor(now / 1000),
                Math.floor(now / 1000)
            );
            
            // Insert memories
            parsedMemories.forEach((memory, index) => {
                insertMemoryStmt.run(
                    crypto.randomUUID(),
                    uuid,
                    index,
                    memory.kind,
                    memory.content,
                    JSON.stringify(memory.tags),
                    memory.source,
                    JSON.stringify(memory.metadata),
                    now
                );
            });
            
            newlyIngestedCount++;
            if (newlyIngestedCount % 10 === 0) {
                console.log(`Ingested ${newlyIngestedCount} sessions...`);
            }
            
        } catch (err) {
            console.error(`❌ Failed to process ${filePath}:`, err.message);
        }
    }
    
    console.log('\n==================================================');
    console.log('🎉 INGESTION PIPELINE COMPLETE');
    console.log('==================================================');
    console.log(`Newly Ingested Sessions:  ${newlyIngestedCount}`);
    console.log(`Skipped Duplicates:       ${skippedDuplicateCount}`);
    console.log(`Skipped Large Files (>25MB): ${skippedLargeCount}`);
    console.log(`Extracted Facts/Memories:  ${newlyFormedMemoriesCount}`);
    console.log('\nSessions Ingested per Project:');
    Object.entries(projectStats).forEach(([proj, cnt]) => {
        console.log(`  - ${proj}: ${cnt} session(s)`);
    });
    console.log('==================================================');
    
    db.close();
}

run();
