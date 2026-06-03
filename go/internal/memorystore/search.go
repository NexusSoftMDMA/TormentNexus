package memorystore

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type SearchResult struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	CreatedAt int64  `json:"createdAt"`
	Source    string `json:"source"`
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
}

// Search executes a fast full-text or LIKE search directly across the TormentNexus local SQLite database.
// This serves as the Go sidecar fallback for the more complex LanceDB vector router in TypeScript.
func Search(workspaceRoot string, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return []SearchResult{}, nil
		}
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	defer db.Close()

	queryStr := fmt.Sprintf("%%%s%%", query)

	// Combine results from both web_memories and imported_session_memories
	// using a UNION ALL query for high-performance retrieval
	stmt := `
		SELECT id, content, tags as type, saved_at as createdAt, source, url, title
		FROM web_memories 
		WHERE content LIKE ? OR title LIKE ?
		UNION ALL
		SELECT uuid as id, content, kind as type, created_at as createdAt, source, '' as url, '' as title
		FROM imported_session_memories
		WHERE content LIKE ?
		ORDER BY createdAt DESC
		LIMIT ?
	`

	rows, err := db.Query(stmt, queryStr, queryStr, queryStr, limit)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") || strings.Contains(err.Error(), "unable to open database file") {
			return []SearchResult{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var entry SearchResult
		err := rows.Scan(&entry.ID, &entry.Content, &entry.Type, &entry.CreatedAt, &entry.Source, &entry.URL, &entry.Title)
		if err == nil {
			results = append(results, entry)
		}
	}

	return results, nil
}
