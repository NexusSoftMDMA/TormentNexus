//go:build ignore
// +build ignore

package tools

/**
 * @file context_server.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of MCP Context Server — SQLite-backed context management.
 * Replaces: github.com/alex-feel/mcp-context-server
 *
 * Designed for managing AI agent contexts: store, search, retrieve,
 * and manage conversation threads with semantic search capability.
 *
 * Tools:
 *  - context_store — store a context entry
 *  - context_search — search contexts by keyword
 *  - context_get — get context entries by ID
 *  - context_delete — delete a context entry
 *  - context_list_threads — list all threads
 *  - context_stats — get storage statistics
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	contextDB   *sql.DB
	contextOnce sync.Once
)

func getContextDB() (*sql.DB, error) {
	var initErr error
	contextOnce.Do(func() {
		dbPath := ".tormentnexus/context.db"
		os.MkdirAll(".tormentnexus", 0755)
		db, e := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", dbPath))
		if e != nil {
			initErr = e
			return
		}
		_, e = db.Exec(`CREATE TABLE IF NOT EXISTS contexts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			thread_id TEXT NOT NULL DEFAULT 'default',
			role TEXT DEFAULT 'user',
			content TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
		if e != nil {
			initErr = e
			return
		}
		_, e = db.Exec(`CREATE INDEX IF NOT EXISTS idx_contexts_thread ON contexts(thread_id)`)
		if e != nil {
			initErr = e
			return
		}
		contextDB = db
	})
	return contextDB, initErr
}

// HandleContextStore stores a context entry.
func HandleContextStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	content, _ := getString(args, "content", "text")
	if content == "" {
		return err("content is required")
	}
	threadID, _ := getString(args, "thread_id", "thread", "session")
	if threadID == "" {
		threadID = "default"
	}
	role, _ := getString(args, "role")
	if role == "" {
		role = "user"
	}
	meta, _ := getString(args, "metadata")
	if meta == "" {
		meta = "{}"
	}

	res, e := db.ExecContext(ctx,
		"INSERT INTO contexts (thread_id, role, content, metadata) VALUES (?, ?, ?, ?)",
		threadID, role, content, meta)
	if e != nil {
		return err(fmt.Sprintf("store failed: %v", e))
	}
	id, _ := res.LastInsertId()
	return ok(fmt.Sprintf("Context stored with ID: %d", id))
}

// HandleContextSearch searches contexts by keyword.
func HandleContextSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	query, _ := getString(args, "query", "q")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	threadID, _ := getString(args, "thread_id", "thread")

	var rows *sql.Rows
	var err2 error
	q := strings.ReplaceAll(query, "'", "''")

	if threadID != "" {
		rows, err2 = db.QueryContext(ctx,
			fmt.Sprintf("SELECT id, thread_id, role, content, metadata, created_at FROM contexts WHERE content LIKE '%%%s%%' AND thread_id = ? ORDER BY created_at DESC LIMIT ?", q),
			threadID, limit)
	} else {
		rows, err2 = db.QueryContext(ctx,
			fmt.Sprintf("SELECT id, thread_id, role, content, metadata, created_at FROM contexts WHERE content LIKE '%%%s%%' ORDER BY created_at DESC LIMIT ?", q),
			limit)
	}
	if err2 != nil {
		return err(fmt.Sprintf("search failed: %v", err2))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var threadID, role, content, meta, createdAt string
		rows.Scan(&id, &threadID, &role, &content, &meta, &createdAt)
		results = append(results, map[string]interface{}{
			"id": id, "thread_id": threadID, "role": role,
			"content": content, "metadata": meta, "created_at": createdAt,
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleContextGet retrieves context entries by ID.
func HandleContextGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	idStr, _ := getString(args, "id", "ids")
	if idStr == "" {
		threadID, _ := getString(args, "thread_id", "thread")
		if threadID != "" {
			limit := getInt(args, "limit")
			if limit <= 0 || limit > 100 {
				limit = 50
			}
			rows, e := db.QueryContext(ctx,
				"SELECT id, thread_id, role, content, metadata, created_at FROM contexts WHERE thread_id = ? ORDER BY created_at DESC LIMIT ?",
				threadID, limit)
			if e != nil {
				return err(fmt.Sprintf("query failed: %v", e))
			}
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() {
				var id int
				var t, r, c, m, ca string
				rows.Scan(&id, &t, &r, &c, &m, &ca)
				results = append(results, map[string]interface{}{
					"id": id, "thread_id": t, "role": r, "content": c, "metadata": m, "created_at": ca,
				})
			}
			out, _ := json.MarshalIndent(results, "", "  ")
			return ok(string(out))
		}
		return err("id or thread_id is required")
	}

	row := db.QueryRowContext(ctx,
		"SELECT id, thread_id, role, content, metadata, created_at FROM contexts WHERE id = ?", idStr)
	var id int
	var t, r, c, m, ca string
	if e := row.Scan(&id, &t, &r, &c, &m, &ca); e != nil {
		return err("context not found: " + idStr)
	}
	out, _ := json.MarshalIndent(map[string]interface{}{
		"id": id, "thread_id": t, "role": r, "content": c, "metadata": m, "created_at": ca,
	}, "", "  ")
	return ok(string(out))
}

// HandleContextDelete deletes a context entry.
func HandleContextDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	id := getInt(args, "id")
	if id <= 0 {
		return err("valid id is required")
	}

	res, e := db.ExecContext(ctx, "DELETE FROM contexts WHERE id = ?", id)
	if e != nil {
		return err(fmt.Sprintf("delete failed: %v", e))
	}
	n, _ := res.RowsAffected()
	return ok(fmt.Sprintf("Deleted %d context entry", n))
}

// HandleContextListThreads lists all threads.
func HandleContextListThreads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	rows, e := db.QueryContext(ctx,
		"SELECT thread_id, COUNT(*), MAX(created_at) FROM contexts GROUP BY thread_id ORDER BY MAX(created_at) DESC")
	if e != nil {
		return err(fmt.Sprintf("query failed: %v", e))
	}
	defer rows.Close()

	var threads []map[string]interface{}
	for rows.Next() {
		var t string
		var count int
		var last string
		rows.Scan(&t, &count, &last)
		threads = append(threads, map[string]interface{}{
			"thread_id": t, "messages": count, "last_active": last,
		})
	}
	if threads == nil {
		threads = []map[string]interface{}{}
	}
	out, _ := json.MarshalIndent(threads, "", "  ")
	return ok(string(out))
}

// HandleContextStats returns storage statistics.
func HandleContextStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := getContextDB()
	if e != nil {
		return err(fmt.Sprintf("db error: %v", e))
	}

	var totalEntries, totalThreads int
	var totalSize int64
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM contexts").Scan(&totalEntries)
	db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT thread_id) FROM contexts").Scan(&totalThreads)
	db.QueryRowContext(ctx, "SELECT SUM(LENGTH(content)) FROM contexts").Scan(&totalSize)

	stats := map[string]interface{}{
		"total_entries": totalEntries,
		"total_threads": totalThreads,
		"total_size_bytes": totalSize,
	}
	out, _ := json.MarshalIndent(stats, "", "  ")
	return ok(string(out))
}
