//go:build ignore
// +build ignore

package tools

/**
 * @file conport.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of ConPort (Context Portal) MCP tools.
 * Replaces `conport` (uvx context-portal-mcp) entry in mcp.json.
 *
 * ConPort provides a structured context management system for AI coding:
 * - Product context (goals, architecture, conventions)
 * - Decision logs (architectural decisions with rationale)
 * - System patterns (coding patterns, templates)
 * - Active context (current task focus)
 * - Progress entries
 *
 * Improvements over original:
 * - No uvx/Python dependency.
 * - SQLite-backed persistent storage.
 * - Supports: get_context, update_context, log_decision, get_decisions,
 *   add_pattern, get_patterns, set_active_context, get_active_context,
 *   log_progress, get_progress.
 * - Context-aware with timeout.
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func conportDBPath() string {
	if p := os.Getenv("CONPORT_DB_PATH"); p != "" {
		return p
	}
	// Default to project-level .conport.db
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".conport.db")
}

func conportOpen() (*sql.DB, error) {
	dbPath := conportDBPath()
	db, e := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", dbPath))
	if e != nil {
		return nil, e
	}

	// Initialize tables if they don't exist
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS product_context (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS decisions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scope TEXT NOT NULL DEFAULT 'project',
			decision TEXT NOT NULL,
			rationale TEXT DEFAULT '',
			tags TEXT DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS system_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			pattern TEXT NOT NULL,
			category TEXT DEFAULT 'general',
			tags TEXT DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS active_context (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			context TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS progress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL DEFAULT 'in_progress',
			description TEXT NOT NULL,
			parent_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, m := range migrations {
		if _, e := db.Exec(m); e != nil {
			return nil, fmt.Errorf("migration failed: %v", e)
		}
	}

	return db, nil
}

// HandleConPortGetContext retrieves all product context.
// Tool: conport_get_context
func HandleConPortGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	rows, e := db.QueryContext(ctx, "SELECT key, value, updated_at FROM product_context ORDER BY key")
	if e != nil {
		return err(fmt.Sprintf("Failed to get context: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var key, value, updatedAt string
		if err := rows.Scan(&key, &value, &updatedAt); err == nil {
			results = append(results, map[string]interface{}{
				"key":        key,
				"value":      value,
				"updated_at": updatedAt,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleConPortUpdateContext sets or updates a product context entry.
// Tool: conport_update_context
func HandleConPortUpdateContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ := getString(args, "key")
	value, _ := getString(args, "value")
	if key == "" || value == "" {
		return err("key and value parameters are required")
	}

	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	_, e = db.ExecContext(ctx,
		`INSERT INTO product_context (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		key, value, time.Now().UTC().Format(time.RFC3339))
	if e != nil {
		return err(fmt.Sprintf("Failed to update context: %v", e))
	}

	return ok(fmt.Sprintf("Context updated: %s", key))
}

// HandleConPortLogDecision logs an architectural decision.
// Tool: conport_log_decision
func HandleConPortLogDecision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	decision, _ := getString(args, "decision")
	if decision == "" {
		return err("decision parameter is required")
	}

	rationale, _ := getString(args, "rationale")
	scope, _ := getString(args, "scope")
	if scope == "" {
		scope = "project"
	}

	tagsJSON, _ := json.Marshal([]string{})
	if tags, ok := args["tags"].([]interface{}); ok {
		tagStrs := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				tagStrs = append(tagStrs, s)
			}
		}
		tagsJSON, _ = json.Marshal(tagStrs)
	}

	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	result, e := db.ExecContext(ctx,
		"INSERT INTO decisions (scope, decision, rationale, tags) VALUES (?, ?, ?, ?)",
		scope, decision, rationale, string(tagsJSON))
	if e != nil {
		return err(fmt.Sprintf("Failed to log decision: %v", e))
	}

	id, _ := result.LastInsertId()
	return ok(fmt.Sprintf("Decision logged (ID: %d): %s", id, decision))
}

// HandleConPortGetDecisions retrieves decision log entries.
// Tool: conport_get_decisions
func HandleConPortGetDecisions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	query := "SELECT id, scope, decision, rationale, tags, created_at FROM decisions ORDER BY id DESC"
	filterTags, _ := args["tags"].([]interface{})
	scope, _ := getString(args, "scope")

	if scope != "" {
		query = fmt.Sprintf("SELECT id, scope, decision, rationale, tags, created_at FROM decisions WHERE scope = '%s' ORDER BY id DESC", scope)
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to get decisions: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var scopeVal, decisionVal, rationaleVal, tagsVal, createdAt string
		if err := rows.Scan(&id, &scopeVal, &decisionVal, &rationaleVal, &tagsVal, &createdAt); err == nil {
			var tags []string
			json.Unmarshal([]byte(tagsVal), &tags)

			// Filter by tags if specified
			if len(filterTags) > 0 {
				matched := false
				for _, ft := range filterTags {
					if ftStr, ok := ft.(string); ok {
						for _, t := range tags {
							if strings.EqualFold(t, ftStr) {
								matched = true
								break
							}
						}
					}
				}
				if !matched {
					continue
				}
			}

			results = append(results, map[string]interface{}{
				"id":         id,
				"scope":      scopeVal,
				"decision":   decisionVal,
				"rationale":  rationaleVal,
				"tags":       tags,
				"created_at": createdAt,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleConPortAddPattern adds a system pattern.
// Tool: conport_add_pattern
func HandleConPortAddPattern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name")
	pattern, _ := getString(args, "pattern")
	if name == "" || pattern == "" {
		return err("name and pattern parameters are required")
	}

	description, _ := getString(args, "description")
	category, _ := getString(args, "category")
	if category == "" {
		category = "general"
	}

	tagsJSON, _ := json.Marshal([]string{})
	if tags, ok := args["tags"].([]interface{}); ok {
		tagStrs := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				tagStrs = append(tagStrs, s)
			}
		}
		tagsJSON, _ = json.Marshal(tagStrs)
	}

	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	result, e := db.ExecContext(ctx,
		"INSERT INTO system_patterns (name, description, pattern, category, tags) VALUES (?, ?, ?, ?, ?)",
		name, description, pattern, category, string(tagsJSON))
	if e != nil {
		return err(fmt.Sprintf("Failed to add pattern: %v", e))
	}

	id, _ := result.LastInsertId()
	return ok(fmt.Sprintf("Pattern added (ID: %d): %s", id, name))
}

// HandleConPortGetPatterns retrieves system patterns.
// Tool: conport_get_patterns
func HandleConPortGetPatterns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	query := "SELECT id, name, description, pattern, category, tags, created_at FROM system_patterns ORDER BY category, name"
	category, _ := getString(args, "category")
	if category != "" {
		query = fmt.Sprintf("SELECT id, name, description, pattern, category, tags, created_at FROM system_patterns WHERE category = '%s' ORDER BY name", category)
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to get patterns: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var name, description, patternVal, cat, tagsVal, createdAt string
		if err := rows.Scan(&id, &name, &description, &patternVal, &cat, &tagsVal, &createdAt); err == nil {
			var tags []string
			json.Unmarshal([]byte(tagsVal), &tags)
			results = append(results, map[string]interface{}{
				"id":          id,
				"name":        name,
				"description": description,
				"pattern":     patternVal,
				"category":    cat,
				"tags":        tags,
				"created_at":  createdAt,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleConPortSetActiveContext sets the current active context.
// Tool: conport_set_active_context
func HandleConPortSetActiveContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contextStr, _ := getString(args, "context")
	if contextStr == "" {
		return err("context parameter is required")
	}

	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	// Clear previous and insert new
	db.ExecContext(ctx, "DELETE FROM active_context")
	_, e = db.ExecContext(ctx, "INSERT INTO active_context (context, updated_at) VALUES (?, ?)",
		contextStr, time.Now().UTC().Format(time.RFC3339))
	if e != nil {
		return err(fmt.Sprintf("Failed to set active context: %v", e))
	}

	return ok("Active context set: " + contextStr)
}

// HandleConPortGetActiveContext retrieves the current active context.
// Tool: conport_get_active_context
func HandleConPortGetActiveContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	var contextStr, updatedAt string
	row := db.QueryRowContext(ctx, "SELECT context, updated_at FROM active_context ORDER BY id DESC LIMIT 1")
	if e := row.Scan(&contextStr, &updatedAt); e != nil {
		if e == sql.ErrNoRows {
			return ok("No active context set.")
		}
		return err(fmt.Sprintf("Failed to get active context: %v", e))
	}

	result := map[string]interface{}{
		"context":    contextStr,
		"updated_at": updatedAt,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleConPortLogProgress logs a progress entry.
// Tool: conport_log_progress
func HandleConPortLogProgress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ := getString(args, "description")
	if description == "" {
		return err("description parameter is required")
	}

	status, _ := getString(args, "status")
	if status == "" {
		status = "in_progress"
	}

	parentID := getInt(args, "parent_id")

	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	result, e := db.ExecContext(ctx,
		"INSERT INTO progress (status, description, parent_id, updated_at) VALUES (?, ?, ?, ?)",
		status, description, parentID, time.Now().UTC().Format(time.RFC3339))
	if e != nil {
		return err(fmt.Sprintf("Failed to log progress: %v", e))
	}

	id, _ := result.LastInsertId()
	return ok(fmt.Sprintf("Progress logged (ID: %d, Status: %s): %s", id, status, description))
}

// HandleConPortGetProgress retrieves progress entries.
// Tool: conport_get_progress
func HandleConPortGetProgress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := conportOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	status, _ := getString(args, "status")

	query := "SELECT id, status, description, parent_id, updated_at FROM progress ORDER BY id DESC"
	if status != "" {
		query = fmt.Sprintf("SELECT id, status, description, parent_id, updated_at FROM progress WHERE status = '%s' ORDER BY id DESC", status)
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to get progress: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, parentID int
		var status, description, updatedAt string
		if err := rows.Scan(&id, &status, &description, &parentID, &updatedAt); err == nil {
			results = append(results, map[string]interface{}{
				"id":          id,
				"status":      status,
				"description": description,
				"parent_id":   parentID,
				"updated_at":  updatedAt,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}
