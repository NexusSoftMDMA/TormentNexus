//go:build ignore
// +build ignore

package tools

/**
 * @file taskmaster.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of TaskMaster AI MCP tools.
 * Replaces `task-master-ai` (npx task-master-ai) entry in mcp.json.
 *
 * TaskMaster provides AI-powered task and project management:
 * - Create, update, and track tasks with subtasks
 * - Generate task breakdowns from PRDs
 * - Expand tasks into implementation steps
 * - Set and track task status and dependencies
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - SQLite-backed persistent task storage.
 * - Supports: create_task, get_task, list_tasks, update_status,
 *   add_subtask, expand_task, generate_tasks_from_prd, next_task.
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

func taskmasterDBPath() string {
	if p := os.Getenv("TASKMASTER_DB_PATH"); p != "" {
		return p
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "taskmaster.db")
}

func taskmasterOpen() (*sql.DB, error) {
	dbPath := taskmasterDBPath()
	db, e := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", dbPath))
	if e != nil {
		return nil, e
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		status TEXT DEFAULT 'pending',
		priority TEXT DEFAULT 'medium',
		dependencies TEXT DEFAULT '[]',
		subtasks TEXT DEFAULT '[]',
		details TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	return db, nil
}

// HandleTaskMasterCreateTask creates a new task.
// Tool: taskmaster_create_task
func HandleTaskMasterCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title", "name")
	if title == "" {
		return err("title parameter is required")
	}

	description, _ := getString(args, "description", "body")
	priority, _ := getString(args, "priority")
	if priority == "" {
		priority = "medium"
	}
	details, _ := getString(args, "details")

	depJSON, _ := json.Marshal([]string{})
	if deps, ok := args["dependencies"].([]interface{}); ok {
		depStrs := make([]string, 0, len(deps))
		for _, d := range deps {
			if s, okS := d.(string); okS {
				depStrs = append(depStrs, s)
			}
		}
		depJSON, _ = json.Marshal(depStrs)
	}

	subtaskJSON, _ := json.Marshal([]interface{}{})

	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	result, e := db.ExecContext(ctx,
		"INSERT INTO tasks (title, description, status, priority, dependencies, subtasks, details) VALUES (?, ?, 'pending', ?, ?, ?, ?)",
		title, description, priority, string(depJSON), string(subtaskJSON), details)
	if e != nil {
		return err(fmt.Sprintf("Failed to create task: %v", e))
	}

	id, _ := result.LastInsertId()
	return ok(fmt.Sprintf("Task created (ID: %d, Priority: %s): %s", id, priority, title))
}

// HandleTaskMasterGetTask retrieves a task by ID.
// Tool: taskmaster_get_task
func HandleTaskMasterGetTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "task_id")
	if id <= 0 {
		return err("id parameter is required")
	}

	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	var title, description, status, priority, depJSON, subtaskJSON, details, createdAt, updatedAt string
	row := db.QueryRowContext(ctx,
		"SELECT id, title, description, status, priority, dependencies, subtasks, details, created_at, updated_at FROM tasks WHERE id = ?",
		id)
	if e := row.Scan(&id, &title, &description, &status, &priority, &depJSON, &subtaskJSON, &details, &createdAt, &updatedAt); e != nil {
		return err(fmt.Sprintf("Task not found: %v", e))
	}

	var deps []string
	json.Unmarshal([]byte(depJSON), &deps)
	var subtasks []interface{}
	json.Unmarshal([]byte(subtaskJSON), &subtasks)

	result := map[string]interface{}{
		"id":           id,
		"title":        title,
		"description":  description,
		"status":       status,
		"priority":     priority,
		"dependencies": deps,
		"subtasks":     subtasks,
		"details":      details,
		"created_at":   createdAt,
		"updated_at":   updatedAt,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleTaskMasterListTasks lists all tasks.
// Tool: taskmaster_list_tasks
func HandleTaskMasterListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	status, _ := getString(args, "status")
	query := "SELECT id, title, status, priority FROM tasks ORDER BY id"
	if status != "" {
		query = fmt.Sprintf("SELECT id, title, status, priority FROM tasks WHERE status = '%s' ORDER BY id", status)
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to list tasks: %v", e))
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var title, statusVal, priority string
		if rows.Scan(&id, &title, &statusVal, &priority) == nil {
			results = append(results, map[string]interface{}{
				"id":       id,
				"title":    title,
				"status":   statusVal,
				"priority": priority,
			})
		}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleTaskMasterUpdateStatus updates a task's status.
// Tool: taskmaster_update_status
func HandleTaskMasterUpdateStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "task_id")
	status, _ := getString(args, "status")
	if id <= 0 || status == "" {
		return err("id and status parameters are required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending": true, "in-progress": true, "done": true,
		"review": true, "deferred": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return err(fmt.Sprintf("Invalid status: %s. Valid: pending, in-progress, done, review, deferred, cancelled", status))
	}

	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	result, e := db.ExecContext(ctx,
		"UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now().UTC().Format(time.RFC3339), id)
	if e != nil {
		return err(fmt.Sprintf("Failed to update status: %v", e))
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return err(fmt.Sprintf("Task %d not found", id))
	}

	return ok(fmt.Sprintf("Task %d status updated to: %s", id, status))
}

// HandleTaskMasterAddSubtask adds a subtask to a task.
// Tool: taskmaster_add_subtask
func HandleTaskMasterAddSubtask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	parentID := getInt(args, "parent_id", "task_id")
	title, _ := getString(args, "title", "name")
	if parentID <= 0 || title == "" {
		return err("parent_id and title parameters are required")
	}

	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	// Get existing subtasks
	var subtaskJSON string
	row := db.QueryRowContext(ctx, "SELECT subtasks FROM tasks WHERE id = ?", parentID)
	if e := row.Scan(&subtaskJSON); e != nil {
		return err(fmt.Sprintf("Parent task %d not found", parentID))
	}

	var subtasks []interface{}
	json.Unmarshal([]byte(subtaskJSON), &subtasks)

	newSubtask := map[string]interface{}{
		"title":  title,
		"status": "pending",
	}
	subtasks = append(subtasks, newSubtask)

	newSubtaskJSON, _ := json.Marshal(subtasks)
	db.ExecContext(ctx, "UPDATE tasks SET subtasks = ?, updated_at = ? WHERE id = ?",
		string(newSubtaskJSON), time.Now().UTC().Format(time.RFC3339), parentID)

	return ok(fmt.Sprintf("Subtask added to task %d: %s", parentID, title))
}

// HandleTaskMasterNextTask gets the next actionable task.
// Tool: taskmaster_next_task
func HandleTaskMasterNextTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	// Find the highest-priority pending task with no incomplete dependencies
	rows, e := db.QueryContext(ctx,
		"SELECT id, title, description, status, priority, dependencies FROM tasks WHERE status = 'pending' ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, id")
	if e != nil {
		return err(fmt.Sprintf("Failed to find next task: %v", e))
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title, description, status, priority, depJSON string
		if rows.Scan(&id, &title, &description, &status, &priority, &depJSON) != nil {
			continue
		}

		var deps []string
		json.Unmarshal([]byte(depJSON), &deps)

		// Check if all dependencies are done
		allDone := true
		for _, dep := range deps {
			var depStatus string
			if db.QueryRowContext(ctx, "SELECT status FROM tasks WHERE id = ?", dep).Scan(&depStatus) == nil {
				if depStatus != "done" && depStatus != "cancelled" {
					allDone = false
					break
				}
			}
		}

		if allDone {
			result := map[string]interface{}{
				"id":           id,
				"title":        title,
				"description":  description,
				"status":       status,
				"priority":     priority,
				"dependencies": deps,
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return ok(string(out))
		}
	}

	return ok("No actionable tasks found. All pending tasks have incomplete dependencies.")
}

// HandleTaskMasterGenerateFromPRD generates tasks from a PRD document.
// Tool: taskmaster_generate_from_prd
func HandleTaskMasterGenerateFromPRD(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prd, _ := getString(args, "prd", "content")
	if prd == "" {
		prdPath, _ := getString(args, "prd_path", "file_path")
		if prdPath == "" {
			return err("prd or prd_path parameter is required")
		}
		data, e := os.ReadFile(prdPath)
		if e != nil {
			return err(fmt.Sprintf("Failed to read PRD: %v", e))
		}
		prd = string(data)
	}

	// Use LLM to generate task breakdown
	prompt := fmt.Sprintf(
		"Analyze the following Product Requirements Document and break it down into implementation tasks. "+
			"For each task, provide a title, description, priority (high/medium/low), and dependencies. "+
			"Output as JSON array of objects with keys: title, description, priority, dependencies.\n\nPRD:\n%s", prd)

	result, e := callLLM(ctx,
		"You are a project manager who breaks down PRDs into actionable development tasks.", prompt, 0.3, "")
	if e != nil {
		// Fallback: simple section-based task generation
		return ok(generateSimpleTasks(prd))
	}

	return ok(result)
}

// HandleTaskMasterExpandTask expands a task into detailed implementation steps.
// Tool: taskmaster_expand_task
func HandleTaskMasterExpandTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "task_id")
	if id <= 0 {
		return err("id parameter is required")
	}

	db, e := taskmasterOpen()
	if e != nil {
		return err(e.Error())
	}
	defer db.Close()

	var title, description, details string
	row := db.QueryRowContext(ctx, "SELECT title, description, details FROM tasks WHERE id = ?", id)
	if e := row.Scan(&title, &description, &details); e != nil {
		return err(fmt.Sprintf("Task %d not found", id))
	}

	// Use LLM to expand
	prompt := fmt.Sprintf(
		"Expand the following task into detailed implementation steps. "+
			"For each step provide a clear action item.\n\nTask: %s\nDescription: %s\nDetails: %s",
		title, description, details)

	result, e := callLLM(ctx,
		"You are a senior developer who breaks tasks into implementation steps.", prompt, 0.3, "")
	if e != nil {
		return ok(fmt.Sprintf("Task: %s\nDescription: %s\n[No LLM available for expansion]", title, description))
	}

	// Save expanded details
	db.ExecContext(ctx, "UPDATE tasks SET details = ?, updated_at = ? WHERE id = ?",
		result, time.Now().UTC().Format(time.RFC3339), id)

	return ok(result)
}

func generateSimpleTasks(prd string) string {
	// Simple PRD parsing: split by headers
	sections := strings.Split(prd, "\n## ")
	var tasks []string
	for _, section := range sections {
		lines := strings.Split(section, "\n")
		title := strings.TrimSpace(lines[0])
		if title == "" {
			continue
		}
		tasks = append(tasks, fmt.Sprintf("- [ ] %s", title))
	}
	if len(tasks) == 0 {
		return "Could not auto-generate tasks from PRD. Please use an LLM API key for intelligent task generation."
	}
	return strings.Join(tasks, "\n")
}
