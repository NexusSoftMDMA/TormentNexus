package sessionimport

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type ImportedSessionMemoryKind string

type ImportedSessionMemorySource string

const (
	ImportedSessionMemoryKindMemory      ImportedSessionMemoryKind   = "memory"
	ImportedSessionMemoryKindInstruction ImportedSessionMemoryKind   = "instruction"
	ImportedSessionMemorySourceLLM       ImportedSessionMemorySource = "llm"
	ImportedSessionMemorySourceHeuristic ImportedSessionMemorySource = "heuristic"
)

type ImportedSessionMemoryInput struct {
	Kind     ImportedSessionMemoryKind   `json:"kind"`
	Content  string                      `json:"content"`
	Tags     []string                    `json:"tags"`
	Source   ImportedSessionMemorySource `json:"source"`
	Metadata map[string]any              `json:"metadata,omitempty"`
}

type ImportedSessionRecordInput struct {
	SourceTool        string                       `json:"sourceTool"`
	SourcePath        string                       `json:"sourcePath"`
	ExternalSessionID *string                      `json:"externalSessionId,omitempty"`
	Title             *string                      `json:"title,omitempty"`
	SessionFormat     string                       `json:"sessionFormat"`
	Transcript        string                       `json:"transcript"`
	Excerpt           *string                      `json:"excerpt,omitempty"`
	WorkingDirectory  *string                      `json:"workingDirectory,omitempty"`
	TranscriptHash    string                       `json:"transcriptHash"`
	NormalizedSession map[string]any               `json:"normalizedSession"`
	Metadata          map[string]any               `json:"metadata,omitempty"`
	DiscoveredAt      int64                        `json:"discoveredAt"`
	ImportedAt        int64                        `json:"importedAt"`
	LastModifiedAt    *int64                       `json:"lastModifiedAt,omitempty"`
	ParsedMemories    []ImportedSessionMemoryInput `json:"parsedMemories"`
}

type ImportedSessionMemoryRecord struct {
	ID                string                      `json:"id"`
	ImportedSessionID string                      `json:"importedSessionId"`
	Kind              ImportedSessionMemoryKind   `json:"kind"`
	Content           string                      `json:"content"`
	Tags              []string                    `json:"tags"`
	Source            ImportedSessionMemorySource `json:"source"`
	Metadata          map[string]any              `json:"metadata"`
	CreatedAt         int64                       `json:"createdAt"`
}

type ImportedSessionRecord struct {
	ID                string                        `json:"id"`
	SourceTool        string                        `json:"sourceTool"`
	SourcePath        string                        `json:"sourcePath"`
	ExternalSessionID *string                       `json:"externalSessionId"`
	Title             *string                       `json:"title"`
	SessionFormat     string                        `json:"sessionFormat"`
	Transcript        string                        `json:"transcript"`
	Excerpt           *string                       `json:"excerpt"`
	WorkingDirectory  *string                       `json:"workingDirectory"`
	TranscriptHash    string                        `json:"transcriptHash"`
	NormalizedSession map[string]any                `json:"normalizedSession"`
	Metadata          map[string]any                `json:"metadata"`
	DiscoveredAt      int64                         `json:"discoveredAt"`
	ImportedAt        int64                         `json:"importedAt"`
	LastModifiedAt    *int64                        `json:"lastModifiedAt"`
	CreatedAt         int64                         `json:"createdAt"`
	UpdatedAt         int64                         `json:"updatedAt"`
	ParsedMemories    []ImportedSessionMemoryRecord `json:"parsedMemories"`
}

type ImportedSessionMaintenanceStats struct {
	TotalSessions                int `json:"totalSessions"`
	InlineTranscriptCount        int `json:"inlineTranscriptCount"`
	ArchivedTranscriptCount      int `json:"archivedTranscriptCount"`
	MissingRetentionSummaryCount int `json:"missingRetentionSummaryCount"`
}

type ImportedInstructionDoc struct {
	Path      string `json:"path"`
	UpdatedAt int64  `json:"updatedAt"`
	Size      int64  `json:"size"`
}

type importedSessionArchiveSidecar struct {
	SessionID               string         `json:"sessionId"`
	SourceTool              string         `json:"sourceTool"`
	SourcePath              string         `json:"sourcePath"`
	SessionFormat           string         `json:"sessionFormat"`
	TranscriptHash          string         `json:"transcriptHash"`
	Title                   *string        `json:"title"`
	WorkingDirectory        *string        `json:"workingDirectory"`
	TranscriptLength        int            `json:"transcriptLength"`
	Excerpt                 *string        `json:"excerpt"`
	DurableMemoryCount      int            `json:"durableMemoryCount"`
	DurableInstructionCount int            `json:"durableInstructionCount"`
	MemoryTags              []string       `json:"memoryTags"`
	RetentionSummary        map[string]any `json:"retentionSummary"`
	ArchivedAt              int64          `json:"archivedAt"`
}

type ImportedSessionStore struct {
	dbPath      string
	archiveRoot string
	docsDir     string
}

func NewImportedSessionStore(workspaceRoot string) *ImportedSessionStore {
	return &ImportedSessionStore{
		dbPath:      filepath.Join(workspaceRoot, "tormentnexus.db"),
		archiveRoot: filepath.Join(workspaceRoot, ".tormentnexus", "imported_sessions", "archive"),
		docsDir:     filepath.Join(workspaceRoot, ".tormentnexus", "imported_sessions", "docs"),
	}
}

func (s *ImportedSessionStore) HasTranscriptHash(ctx context.Context, transcriptHash string) (bool, error) {
	db, err := s.open(ctx)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var id string
	err = db.QueryRowContext(ctx, `SELECT uuid FROM imported_sessions WHERE transcript_hash = ? LIMIT 1`, strings.TrimSpace(transcriptHash)).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(id) != "", nil
}

func (s *ImportedSessionStore) UpsertSession(ctx context.Context, input ImportedSessionRecordInput) (*ImportedSessionRecord, error) {
	db, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if strings.TrimSpace(input.TranscriptHash) == "" {
		input.TranscriptHash = stableImportedTranscriptHash(input)
	}
	if input.NormalizedSession == nil {
		input.NormalizedSession = map[string]any{}
	}
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	if input.ImportedAt <= 0 {
		input.ImportedAt = time.Now().UTC().UnixMilli()
	}
	if input.DiscoveredAt <= 0 {
		input.DiscoveredAt = input.ImportedAt
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().UnixMilli()
	sessionID, existed, createdAt, err := s.resolveSessionIdentity(ctx, tx, input.TranscriptHash)
	if err != nil {
		return nil, err
	}
	archiveInfo, err := s.writeTranscriptArchive(sessionID, input, now)
	if err != nil {
		return nil, err
	}

	normalizedJSON, err := json.Marshal(input.NormalizedSession)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, err
	}

	if !existed {
		createdAt = input.ImportedAt
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO imported_sessions (
			uuid, source_tool, source_path, external_session_id, title, session_format, transcript, excerpt,
			working_directory, transcript_hash, normalized_session, metadata,
			transcript_archive_path, transcript_metadata_archive_path, transcript_archive_format, transcript_stored_bytes,
			discovered_at, imported_at, last_modified_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uuid) DO UPDATE SET
			source_tool = excluded.source_tool,
			source_path = excluded.source_path,
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
	`,
		sessionID,
		input.SourceTool,
		input.SourcePath,
		nullableString(input.ExternalSessionID),
		nullableString(input.Title),
		input.SessionFormat,
		"",
		nullableString(input.Excerpt),
		nullableString(input.WorkingDirectory),
		input.TranscriptHash,
		string(normalizedJSON),
		string(metadataJSON),
		archiveInfo.TranscriptArchivePath,
		archiveInfo.TranscriptMetadataArchivePath,
		archiveInfo.TranscriptArchiveFormat,
		archiveInfo.TranscriptStoredBytes,
		input.DiscoveredAt,
		input.ImportedAt,
		nullableInt64(input.LastModifiedAt),
		createdAt,
		now,
	)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM imported_session_memories WHERE imported_session_uuid = ?`, sessionID); err != nil {
		return nil, err
	}

	for index, memory := range input.ParsedMemories {
		tagsJSON, err := json.Marshal(memory.Tags)
		if err != nil {
			return nil, err
		}
		memoryMetadataJSON, err := json.Marshal(memory.Metadata)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO imported_session_memories (
				uuid, imported_session_uuid, memory_index, kind, content, tags, source, metadata, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, uuid.NewString(), sessionID, index, string(memory.Kind), memory.Content, string(tagsJSON), string(memory.Source), string(memoryMetadataJSON), now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetImportedSession(ctx, sessionID)
}

func (s *ImportedSessionStore) ListImportedSessions(ctx context.Context, limit int) ([]ImportedSessionRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	db, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT uuid, source_tool, source_path, external_session_id, title, session_format, transcript, excerpt,
		       working_directory, transcript_hash, normalized_session, metadata, transcript_archive_path,
		       transcript_metadata_archive_path, transcript_archive_format, transcript_stored_bytes,
		       discovered_at, imported_at, last_modified_at, created_at, updated_at
		FROM imported_sessions
		ORDER BY imported_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]ImportedSessionRecord, 0, limit)
	for rows.Next() {
		record, err := s.scanImportedSessionRow(ctx, db, rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *ImportedSessionStore) GetImportedSession(ctx context.Context, id string) (*ImportedSessionRecord, error) {
	db, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowContext(ctx, `
		SELECT uuid, source_tool, source_path, external_session_id, title, session_format, transcript, excerpt,
		       working_directory, transcript_hash, normalized_session, metadata, transcript_archive_path,
		       transcript_metadata_archive_path, transcript_archive_format, transcript_stored_bytes,
		       discovered_at, imported_at, last_modified_at, created_at, updated_at
		FROM imported_sessions
		WHERE uuid = ?
		LIMIT 1
	`, id)
	return s.scanImportedSessionRow(ctx, db, row)
}

func (s *ImportedSessionStore) ListInstructionMemories(ctx context.Context, limit int) ([]ImportedSessionMemoryRecord, error) {
	if limit <= 0 {
		limit = 200
	}
	db, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT uuid, imported_session_uuid, kind, content, tags, source, metadata, created_at
		FROM imported_session_memories
		WHERE kind = 'instruction'
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := make([]ImportedSessionMemoryRecord, 0, limit)
	for rows.Next() {
		memory, err := scanImportedSessionMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memories, nil
}

func (s *ImportedSessionStore) GetMaintenanceStats(ctx context.Context) (ImportedSessionMaintenanceStats, error) {
	records, err := s.ListImportedSessions(ctx, 10000)
	if err != nil {
		return ImportedSessionMaintenanceStats{}, err
	}
	stats := ImportedSessionMaintenanceStats{TotalSessions: len(records)}
	for _, record := range records {
		if strings.TrimSpace(record.Transcript) != "" {
			stats.InlineTranscriptCount++
		}
		if strings.TrimSpace(anyString(record.Metadata["archiveFormat"])) != "" {
			stats.ArchivedTranscriptCount++
		}
		if _, ok := record.Metadata["retentionSummary"]; !ok {
			stats.MissingRetentionSummaryCount++
		}
	}
	return stats, nil
}

func (s *ImportedSessionStore) WriteInstructionDoc(ctx context.Context, limit int) (*ImportedInstructionDoc, error) {
	instructions, err := s.ListInstructionMemories(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(instructions) == 0 {
		return nil, nil
	}
	if err := os.MkdirAll(s.docsDir, 0o755); err != nil {
		return nil, err
	}
	docPath := filepath.Join(s.docsDir, "auto-imported-agent-instructions.md")
	lines := []string{
		"# Auto-imported Agent Instructions",
		"",
		"Generated by the Go imported-session store.",
		"",
		"## Durable instructions",
		"",
	}
	for _, instruction := range instructions {
		sourceTool := anyString(instruction.Metadata["sourceTool"])
		if sourceTool == "" {
			sourceTool = "unknown"
		}
		sourcePath := anyString(instruction.Metadata["path"])
		if sourcePath == "" {
			sourcePath = "unknown source"
		}
		lines = append(lines, fmt.Sprintf("- **%s** — %s _(source: `%s`)_", sourceTool, instruction.Content, sourcePath))
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(docPath, []byte(content), 0o644); err != nil {
		return nil, err
	}
	info, err := os.Stat(docPath)
	if err != nil {
		return nil, err
	}
	return &ImportedInstructionDoc{Path: docPath, UpdatedAt: info.ModTime().UnixMilli(), Size: info.Size()}, nil
}

func (s *ImportedSessionStore) ListInstructionDocs() ([]ImportedInstructionDoc, error) {
	entries, err := os.ReadDir(s.docsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ImportedInstructionDoc{}, nil
		}
		return nil, err
	}
	docs := make([]ImportedInstructionDoc, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(s.docsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		docs = append(docs, ImportedInstructionDoc{Path: path, UpdatedAt: info.ModTime().UnixMilli(), Size: info.Size()})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].UpdatedAt > docs[j].UpdatedAt })
	return docs, nil
}

type transcriptArchiveInfo struct {
	TranscriptArchivePath         string
	TranscriptMetadataArchivePath string
	TranscriptArchiveFormat       string
	TranscriptStoredBytes         int64
}

func (s *ImportedSessionStore) writeTranscriptArchive(sessionID string, input ImportedSessionRecordInput, archivedAt int64) (*transcriptArchiveInfo, error) {
	if err := os.MkdirAll(s.archiveRoot, 0o755); err != nil {
		return nil, err
	}
	hash := input.TranscriptHash
	if len(hash) < 4 {
		hash = hash + strings.Repeat("0", 4-len(hash))
	}
	archiveDir := filepath.Join(s.archiveRoot, hash[:2], hash[2:4])
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return nil, err
	}
	transcriptPath := filepath.Join(archiveDir, input.TranscriptHash+".txt.gz")
	metadataPath := filepath.Join(archiveDir, input.TranscriptHash+".meta.json.gz")
	if err := writeGzipFile(transcriptPath, []byte(input.Transcript)); err != nil {
		return nil, err
	}
	sidecar := importedSessionArchiveSidecar{
		SessionID:               sessionID,
		SourceTool:              input.SourceTool,
		SourcePath:              input.SourcePath,
		SessionFormat:           input.SessionFormat,
		TranscriptHash:          input.TranscriptHash,
		Title:                   input.Title,
		WorkingDirectory:        input.WorkingDirectory,
		TranscriptLength:        len(input.Transcript),
		Excerpt:                 input.Excerpt,
		DurableMemoryCount:      len(input.ParsedMemories),
		DurableInstructionCount: countInstructionMemories(input.ParsedMemories),
		MemoryTags:              collectMemoryTags(input.ParsedMemories),
		RetentionSummary:        mapValue(input.Metadata, "retentionSummary"),
		ArchivedAt:              archivedAt,
	}
	payload, err := json.MarshalIndent(sidecar, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := writeGzipFile(metadataPath, payload); err != nil {
		return nil, err
	}
	return &transcriptArchiveInfo{
		TranscriptArchivePath:         toRelativeSlashPath(s.archiveRoot, transcriptPath),
		TranscriptMetadataArchivePath: toRelativeSlashPath(s.archiveRoot, metadataPath),
		TranscriptArchiveFormat:       "gzip-text-v1",
		TranscriptStoredBytes:         int64(len(input.Transcript)),
	}, nil
}

func (s *ImportedSessionStore) open(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return nil, err
	}
	if err := ensureImportedSessionSchema(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureImportedSessionSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS imported_sessions (
			uuid TEXT PRIMARY KEY,
			source_tool TEXT NOT NULL,
			source_path TEXT NOT NULL,
			external_session_id TEXT,
			title TEXT,
			session_format TEXT NOT NULL DEFAULT 'generic',
			transcript TEXT NOT NULL,
			excerpt TEXT,
			working_directory TEXT,
			transcript_hash TEXT NOT NULL UNIQUE,
			transcript_archive_path TEXT,
			transcript_metadata_archive_path TEXT,
			transcript_archive_format TEXT,
			transcript_stored_bytes INTEGER,
			normalized_session TEXT NOT NULL,
			metadata TEXT NOT NULL DEFAULT '{}',
			discovered_at INTEGER NOT NULL,
			imported_at INTEGER NOT NULL,
			last_modified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_imported_sessions_tool ON imported_sessions(source_tool);
		CREATE INDEX IF NOT EXISTS idx_imported_sessions_path ON imported_sessions(source_path);
		CREATE INDEX IF NOT EXISTS idx_imported_sessions_hash ON imported_sessions(transcript_hash);
		CREATE TABLE IF NOT EXISTS imported_session_memories (
			uuid TEXT PRIMARY KEY,
			imported_session_uuid TEXT NOT NULL,
			memory_index INTEGER NOT NULL,
			kind TEXT NOT NULL DEFAULT 'memory',
			content TEXT NOT NULL,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT 'heuristic',
			metadata TEXT NOT NULL DEFAULT '{}',
			created_at INTEGER NOT NULL,
			FOREIGN KEY (imported_session_uuid) REFERENCES imported_sessions(uuid) ON DELETE CASCADE,
			UNIQUE (imported_session_uuid, memory_index)
		);
		CREATE INDEX IF NOT EXISTS idx_imported_session_memories_session ON imported_session_memories(imported_session_uuid);
		CREATE INDEX IF NOT EXISTS idx_imported_session_memories_kind ON imported_session_memories(kind);
	`)
	return err
}

func (s *ImportedSessionStore) resolveSessionIdentity(ctx context.Context, tx *sql.Tx, transcriptHash string) (sessionID string, existed bool, createdAt int64, err error) {
	var existingID string
	var existingCreatedAt int64
	err = tx.QueryRowContext(ctx, `SELECT uuid, created_at FROM imported_sessions WHERE transcript_hash = ? LIMIT 1`, transcriptHash).Scan(&existingID, &existingCreatedAt)
	if err == sql.ErrNoRows {
		return uuid.NewString(), false, 0, nil
	}
	if err != nil {
		return "", false, 0, err
	}
	return existingID, true, existingCreatedAt, nil
}

type importedSessionScanner interface {
	Scan(dest ...any) error
}

func (s *ImportedSessionStore) scanImportedSessionRow(ctx context.Context, db *sql.DB, scanner importedSessionScanner) (*ImportedSessionRecord, error) {
	var (
		id                            string
		sourceTool                    string
		sourcePath                    string
		externalSessionID             sql.NullString
		title                         sql.NullString
		sessionFormat                 string
		transcriptInline              string
		excerpt                       sql.NullString
		workingDirectory              sql.NullString
		transcriptHash                string
		normalizedSessionRaw          string
		metadataRaw                   string
		transcriptArchivePath         sql.NullString
		transcriptMetadataArchivePath sql.NullString
		transcriptArchiveFormat       sql.NullString
		transcriptStoredBytes         sql.NullInt64
		discoveredAt                  int64
		importedAt                    int64
		lastModifiedAt                sql.NullInt64
		createdAt                     int64
		updatedAt                     int64
	)
	if err := scanner.Scan(
		&id, &sourceTool, &sourcePath, &externalSessionID, &title, &sessionFormat, &transcriptInline, &excerpt,
		&workingDirectory, &transcriptHash, &normalizedSessionRaw, &metadataRaw, &transcriptArchivePath,
		&transcriptMetadataArchivePath, &transcriptArchiveFormat, &transcriptStoredBytes,
		&discoveredAt, &importedAt, &lastModifiedAt, &createdAt, &updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	transcript := transcriptInline
	if strings.TrimSpace(transcript) == "" && transcriptArchivePath.Valid {
		archived, err := readGzipFile(filepath.Join(s.archiveRoot, filepath.FromSlash(transcriptArchivePath.String)))
		if err != nil {
			return nil, err
		}
		transcript = string(archived)
	}
	memories, err := s.listParsedMemories(ctx, db, id)
	if err != nil {
		return nil, err
	}
	metadata := parseJSONObject(metadataRaw)
	if strings.TrimSpace(transcriptArchiveFormat.String) != "" {
		metadata["archiveFormat"] = transcriptArchiveFormat.String
	}
	if transcriptStoredBytes.Valid {
		metadata["transcriptStoredBytes"] = transcriptStoredBytes.Int64
	}
	record := &ImportedSessionRecord{
		ID:                id,
		SourceTool:        sourceTool,
		SourcePath:        sourcePath,
		ExternalSessionID: nullStringPointer(externalSessionID),
		Title:             nullStringPointer(title),
		SessionFormat:     sessionFormat,
		Transcript:        transcript,
		Excerpt:           nullStringPointer(excerpt),
		WorkingDirectory:  nullStringPointer(workingDirectory),
		TranscriptHash:    transcriptHash,
		NormalizedSession: parseJSONObject(normalizedSessionRaw),
		Metadata:          metadata,
		DiscoveredAt:      discoveredAt,
		ImportedAt:        importedAt,
		LastModifiedAt:    nullInt64Pointer(lastModifiedAt),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		ParsedMemories:    memories,
	}
	return record, nil
}

func (s *ImportedSessionStore) listParsedMemories(ctx context.Context, db *sql.DB, importedSessionID string) ([]ImportedSessionMemoryRecord, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT uuid, imported_session_uuid, kind, content, tags, source, metadata, created_at
		FROM imported_session_memories
		WHERE imported_session_uuid = ?
		ORDER BY memory_index ASC
	`, importedSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memories := make([]ImportedSessionMemoryRecord, 0)
	for rows.Next() {
		memory, err := scanImportedSessionMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memories, nil
}

func scanImportedSessionMemoryRow(scanner importedSessionScanner) (ImportedSessionMemoryRecord, error) {
	var (
		id                string
		importedSessionID string
		kind              string
		content           string
		tagsRaw           string
		source            string
		metadataRaw       string
		createdAt         int64
	)
	if err := scanner.Scan(&id, &importedSessionID, &kind, &content, &tagsRaw, &source, &metadataRaw, &createdAt); err != nil {
		return ImportedSessionMemoryRecord{}, err
	}
	return ImportedSessionMemoryRecord{
		ID:                id,
		ImportedSessionID: importedSessionID,
		Kind:              ImportedSessionMemoryKind(kind),
		Content:           content,
		Tags:              parseJSONStringSlice(tagsRaw),
		Source:            ImportedSessionMemorySource(source),
		Metadata:          parseJSONObject(metadataRaw),
		CreatedAt:         createdAt,
	}, nil
}

func stableImportedTranscriptHash(input ImportedSessionRecordInput) string {
	hasher := sha256.New()
	hasher.Write([]byte(input.SourceTool))
	hasher.Write([]byte("\n"))
	hasher.Write([]byte(input.SourcePath))
	hasher.Write([]byte("\n"))
	hasher.Write([]byte(input.SessionFormat))
	hasher.Write([]byte("\n"))
	hasher.Write([]byte(input.Transcript))
	return hex.EncodeToString(hasher.Sum(nil))
}

func writeGzipFile(filePath string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := gzip.NewWriter(file)
	if _, err := writer.Write(payload); err != nil {
		writer.Close()
		return err
	}
	return writer.Close()
}

func readGzipFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func toRelativeSlashPath(root, fullPath string) string {
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return filepath.Base(fullPath)
	}
	return filepath.ToSlash(rel)
}

func parseJSONObject(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || parsed == nil {
		return map[string]any{}
	}
	return parsed
}

func parseJSONStringSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return []string{}
	}
	return parsed
}

func nullableString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value.String)
	return &trimmed
}

func nullInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	copy := value.Int64
	return &copy
}

func countInstructionMemories(memories []ImportedSessionMemoryInput) int {
	count := 0
	for _, memory := range memories {
		if memory.Kind == ImportedSessionMemoryKindInstruction {
			count++
		}
	}
	return count
}

func collectMemoryTags(memories []ImportedSessionMemoryInput) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, memory := range memories {
		for _, tag := range memory.Tags {
			normalized := strings.TrimSpace(tag)
			if normalized == "" {
				continue
			}
			key := strings.ToLower(normalized)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, normalized)
		}
	}
	sort.Strings(result)
	return result
}

func mapValue(input map[string]any, key string) map[string]any {
	if input == nil {
		return nil
	}
	value, ok := input[key]
	if !ok {
		return nil
	}
	mapped, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return mapped
}

func anyString(value any) string {
	text, _ := value.(string)
	return text
}
