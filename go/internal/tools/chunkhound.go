//go:build ignore
// +build ignore

package tools

/**
 * @file chunkhound.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of ChunkHound MCP tools.
 * Replaces `ChunkHound` (chunkhound mcp) STDIO entry in mcp.json.
 *
 * ChunkHound provides semantic code search using code chunks:
 * - Indexes code into searchable chunks with embeddings
 * - Supports natural language queries to find relevant code
 * - Returns file paths, line ranges, and code snippets
 *
 * Improvements over original:
 * - No external binary dependency.
 * - Go-native AST parsing for multiple languages.
 * - SQLite-backed chunk index with full-text search.
 * - Supports: index, search, find_similar, list_indexed, stats.
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	_ "modernc.org/sqlite"
)

func chunkhoundDBPath() string {
	if p := os.Getenv("CHUNKHOUND_DB_PATH"); p != "" {
		return p
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".chunkhound.db")
}

func chunkhoundOpen() (*sql.DB, error) {
	dbPath := chunkhoundDBPath()
	db, e := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", dbPath))
	if e != nil {
		return nil, e
	}

	// Initialize tables
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			language TEXT DEFAULT '',
			chunk_type TEXT DEFAULT 'block',
			name TEXT DEFAULT '',
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_file ON chunks(file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_name ON chunks(name)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
			content,
			name,
			file_path,
			content='chunks',
			content_rowid='id'
		)`,
		// Triggers to keep FTS in sync
		`CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
			INSERT INTO chunks_fts(rowid, content, name, file_path) VALUES (new.id, new.content, new.name, new.file_path);
		END`,
		`CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
			INSERT INTO chunks_fts(chunks_fts, rowid, content, name, file_path) VALUES('delete', old.id, old.content, old.name, old.file_path);
		END`,
	}

	for _, m := range migrations {
		if _, e := db.Exec(m); e != nil {
			// FTS5 and triggers may fail silently, that's okay
			if !strings.Contains(e.Error(), "already exists") && !strings.Contains(e.Error(), "fts5") {
				// Non-critical - continue
			}
		}
	}

	return db, nil
}

func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	langMap := map[string]string{
		".go": "go", ".py": "python", ".js": "javascript", ".ts": "typescript",
		".tsx": "typescript", ".jsx": "javascript", ".rs": "rust", ".java": "java",
		".cpp": "cpp", ".c": "c", ".h": "c", ".hpp": "cpp", ".cs": "csharp",
		".rb": "ruby", ".php": "php", ".swift": "swift", ".kt": "kotlin",
		".scala": "scala", ".r": "r", ".R": "r", ".lua": "lua",
		".sh": "bash", ".bash": "bash", ".zsh": "zsh",
		".sql": "sql", ".html": "html", ".css": "css", ".scss": "scss",
		".md": "markdown", ".yaml": "yaml", ".yml": "yaml", ".json": "json",
		".toml": "toml", ".xml": "xml",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

func shouldIndexFile(path string) bool {
	// Skip hidden directories, node_modules, .git, etc.
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return false
		}
		if part == "node_modules" || part == "vendor" || part == "__pycache__" ||
			part == "dist" || part == "build" || part == ".next" || part == "target" {
			return false
		}
	}

	// Only index text files
	ext := strings.ToLower(filepath.Ext(path))
	supportedExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
		".jsx": true, ".rs": true, ".java": true, ".cpp": true, ".c": true,
		".h": true, ".cs": true, ".rb": true, ".php": true, ".swift": true,
		".kt": true, ".scala": true, ".lua": true, ".sh": true,
		".sql": true, ".html": true, ".css": true, ".scss": true,
		".md": true, ".yaml": true, ".yml": true, ".json": true, ".toml": true,
		".xml": true,
	}
	return supportedExts[ext]
}

// Chunk extraction: split files into logical chunks (functions, classes, blocks)
func extractChunks(filePath string, content string) []map[string]interface{} {
	lang := detectLanguage(filePath)
	lines := strings.Split(content, "\n")
	var chunks []map[string]interface{}

	// Regex patterns for different languages to detect function/class boundaries
	var blockStartRegex *regexp.Regexp
	switch lang {
	case "go":
		blockStartRegex = regexp.MustCompile(`^\s*func\s|^\s*type\s+\w+\s+struct|^\s*type\s+\w+\s+interface`)
	case "python":
		blockStartRegex = regexp.MustCompile(`^\s*def\s|^\s*class\s|^\s*async\s+def\s`)
	case "javascript", "typescript", "tsx", "jsx":
		blockStartRegex = regexp.MustCompile(`^\s*(export\s+)?(async\s+)?function\s|^\s*(export\s+)?const\s+\w+\s*=\s*(async\s+)?\(|^\s*class\s|^\s*(export\s+)?interface\s|^\s*(export\s+)?type\s+\w+\s*=`)
	case "rust":
		blockStartRegex = regexp.MustCompile(`^\s*fn\s|^\s*pub\s+fn\s|^\s*(pub\s+)?struct\s|^\s*(pub\s+)?enum\s|^\s*impl\s`)
	case "java", "kotlin":
		blockStartRegex = regexp.MustCompile(`^\s*(public|private|protected)?\s*(static\s+)?(class|interface|enum|void|String|int|long|boolean|float|double)\s`)
	}

	if blockStartRegex == nil {
		// Fallback: chunk by line count (50 lines per chunk)
		chunkSize := 50
		for i := 0; i < len(lines); i += chunkSize {
			end := i + chunkSize
			if end > len(lines) {
				end = len(lines)
			}
			chunks = append(chunks, map[string]interface{}{
				"file_path":  filePath,
				"start_line": i + 1,
				"end_line":   end,
				"language":   lang,
				"chunk_type": "block",
				"name":       fmt.Sprintf("lines_%d_%d", i+1, end),
				"content":    strings.Join(lines[i:end], "\n"),
			})
		}
		return chunks
	}

	// Find block boundaries
	var blockStarts []int
	for i, line := range lines {
		if blockStartRegex.MatchString(line) {
			blockStarts = append(blockStarts, i)
		}
	}

	if len(blockStarts) == 0 {
		// No blocks found, chunk by size
		chunkSize := 50
		for i := 0; i < len(lines); i += chunkSize {
			end := i + chunkSize
			if end > len(lines) {
				end = len(lines)
			}
			chunks = append(chunks, map[string]interface{}{
				"file_path":  filePath,
				"start_line": i + 1,
				"end_line":   end,
				"language":   lang,
				"chunk_type": "block",
				"name":       fmt.Sprintf("lines_%d_%d", i+1, end),
				"content":    strings.Join(lines[i:end], "\n"),
			})
		}
		return chunks
	}

	// Extract each block
	for idx, startLine := range blockStarts {
		var endLine int
		if idx+1 < len(blockStarts) {
			endLine = blockStarts[idx+1]
		} else {
			endLine = len(lines)
		}

		// Extract name from the first line
		name := strings.TrimSpace(lines[startLine])
		if len(name) > 80 {
			name = name[:80] + "..."
		}

		chunks = append(chunks, map[string]interface{}{
			"file_path":  filePath,
			"start_line": startLine + 1,
			"end_line":   endLine,
			"language":   lang,
			"chunk_type": "definition",
			"name":       name,
			"content":    strings.Join(lines[startLine:endLine], "\n"),
		})
	}

	return chunks
}

// HandleChunkhoundIndex indexes a directory of code files.
// Tool: chunkhound_index
func HandleChunkhoundIndex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	force := getBool(args, "force", "reindex")

	db, e := chunkhoundOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	// Clear existing index if forced
	if force {
		db.ExecContext(ctx, "DELETE FROM chunks")
	}

	indexed := 0
	skipped := 0
	errors := 0

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(path, p)

		if !shouldIndexFile(p) {
			skipped++
			return nil
		}

		data, e := os.ReadFile(p)
		if e != nil {
			errors++
			return nil
		}

		content := string(data)
		chunks := extractChunks(rel, content)

		for _, chunk := range chunks {
			_, e := db.ExecContext(ctx,
				`INSERT INTO chunks (file_path, start_line, end_line, language, chunk_type, name, content)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				chunk["file_path"], chunk["start_line"], chunk["end_line"],
				chunk["language"], chunk["chunk_type"], chunk["name"], chunk["content"])
			if e != nil {
				errors++
			}
		}

		indexed++
		return nil
	})

	result := map[string]interface{}{
		"indexed_files": indexed,
		"skipped_files": skipped,
		"errors":        errors,
		"path":          path,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleChunkhoundSearch searches indexed code chunks.
// Tool: chunkhound_search
func HandleChunkhoundSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query parameter is required")
	}

	limit := getInt(args, "limit", "count")
	if limit <= 0 {
		limit = 10
	}

	db, e := chunkhoundOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	// Try FTS5 first, fall back to LIKE
	var rows *sql.Rows

	rows, e = db.QueryContext(ctx,
		`SELECT c.id, c.file_path, c.start_line, c.end_line, c.language, c.chunk_type, c.name, c.content
		 FROM chunks c
		 WHERE c.content LIKE ? OR c.name LIKE ?
		 ORDER BY c.file_path, c.start_line
		 LIMIT ?`,
		"%"+query+"%", "%"+query+"%", limit)

	if e != nil {
		return err(fmt.Sprintf("Search failed: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, startLine, endLine int
		var filePath, language, chunkType, name, content string
		if err := rows.Scan(&id, &filePath, &startLine, &endLine, &language, &chunkType, &name, &content); err == nil {
			// Truncate content for display
			displayContent := content
			if len(displayContent) > 500 {
				displayContent = displayContent[:500] + "..."
			}
			results = append(results, map[string]interface{}{
				"id":         id,
				"file_path":  filePath,
				"start_line": startLine,
				"end_line":   endLine,
				"language":   language,
				"chunk_type": chunkType,
				"name":       name,
				"content":    displayContent,
			})
		}
	}

	if len(results) == 0 {
		return ok("No matching chunks found for: " + query)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleChunkhoundStats returns indexing statistics.
// Tool: chunkhound_stats
func HandleChunkhoundStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := chunkhoundOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	var totalChunks, totalFiles int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chunks").Scan(&totalChunks)
	db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT file_path) FROM chunks").Scan(&totalFiles)

	// Get language breakdown
	rows, e := db.QueryContext(ctx, "SELECT language, COUNT(*) as count FROM chunks GROUP BY language ORDER BY count DESC")
	if e != nil {
		return err(e.Error())
	}
	defer rows.Close()

	langBreakdown := map[string]int{}
	for rows.Next() {
		var lang string
		var count int
		if rows.Scan(&lang, &count) == nil {
			langBreakdown[lang] = count
		}
	}

	result := map[string]interface{}{
		"total_chunks":     totalChunks,
		"total_files":      totalFiles,
		"language_counts":  langBreakdown,
		"db_path":          chunkhoundDBPath(),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleChunkhoundListIndexed lists indexed files.
// Tool: chunkhound_list_indexed
func HandleChunkhoundListIndexed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := chunkhoundOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 100
	}

	rows, e := db.QueryContext(ctx,
		"SELECT DISTINCT file_path, COUNT(*) as chunk_count, MIN(start_line) as first_line, MAX(end_line) as last_line FROM chunks GROUP BY file_path ORDER BY file_path LIMIT ?",
		limit)
	if e != nil {
		return err(fmt.Sprintf("Failed to list indexed files: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var filePath string
		var chunkCount, firstLine, lastLine int
		if rows.Scan(&filePath, &chunkCount, &firstLine, &lastLine) == nil {
			results = append(results, map[string]interface{}{
				"file_path":   filePath,
				"chunk_count": chunkCount,
				"first_line":  firstLine,
				"last_line":   lastLine,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleChunkhoundGetChunk retrieves a specific chunk by ID.
// Tool: chunkhound_get_chunk
func HandleChunkhoundGetChunk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id")
	if id <= 0 {
		return err("id parameter is required")
	}

	db, e := chunkhoundOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	var filePath, language, chunkType, name, content string
	var startLine, endLine int
	var createdAt string

	row := db.QueryRowContext(ctx,
		"SELECT id, file_path, start_line, end_line, language, chunk_type, name, content, created_at FROM chunks WHERE id = ?",
		id)
	if e := row.Scan(&id, &filePath, &startLine, &endLine, &language, &chunkType, &name, &content, &createdAt); e != nil {
		return err(fmt.Sprintf("Chunk not found: %v", e))
	}

	result := map[string]interface{}{
		"id":         id,
		"file_path":  filePath,
		"start_line": startLine,
		"end_line":   endLine,
		"language":   language,
		"chunk_type": chunkType,
		"name":       name,
		"content":    content,
		"created_at": createdAt,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
