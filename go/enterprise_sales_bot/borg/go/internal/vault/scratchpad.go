package vault

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Scratchpad struct {
	db        *sql.DB
	sessionID string
	mu        sync.Mutex
	stmtAppend   *sql.Stmt
	stmtGetAll   *sql.Stmt
	stmtComplete *sql.Stmt
}

func NewScratchpad() (*Scratchpad, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil { return nil, fmt.Errorf("open L1: %w", err) }
	if err := InitL1Schema(db); err != nil { db.Close(); return nil, fmt.Errorf("init L1: %w", err) }
	sp := &Scratchpad{db: db}
	if err := sp.prepareStatements(); err != nil { db.Close(); return nil, err }
	return sp, nil
}

func (sp *Scratchpad) prepareStatements() error {
	var err error
	sp.stmtAppend, err = sp.db.Prepare(`INSERT INTO l1_scratch (id, session_id, role, content, tool_name, tool_args, tool_result) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil { return fmt.Errorf("prepare append: %w", err) }
	sp.stmtGetAll, err = sp.db.Prepare(`SELECT id, session_id, role, content, tool_name, tool_args, tool_result, created_at FROM l1_scratch WHERE session_id = ? ORDER BY created_at ASC`)
	if err != nil { return fmt.Errorf("prepare getall: %w", err) }
	sp.stmtComplete, err = sp.db.Prepare(`UPDATE l1_sessions SET status = ?, completed_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`)
	if err != nil { return fmt.Errorf("prepare complete: %w", err) }
	return nil
}

func (sp *Scratchpad) InitSession(agentName, project, prompt, systemCtx string) error {
	sp.mu.Lock(); defer sp.mu.Unlock()
	sp.sessionID = uuid.New().String()
	_, err := sp.db.Exec(`INSERT INTO l1_sessions (id, agent_name, project, status, prompt, system_ctx) VALUES (?, ?, ?, 'active', ?, ?)`, sp.sessionID, agentName, project, prompt, systemCtx)
	if err != nil { return fmt.Errorf("init session: %w", err) }
	if systemCtx != "" {
		sp.stmtAppend.Exec(uuid.New().String(), sp.sessionID, "system", systemCtx, "", "", "")
	}
	_, err = sp.stmtAppend.Exec(uuid.New().String(), sp.sessionID, "user", prompt, "", "", "")
	return err
}

func (sp *Scratchpad) SessionID() string { return sp.sessionID }

func (sp *Scratchpad) Append(role, content string) error {
	return sp.AppendTool(role, content, "", "", "")
}

func (sp *Scratchpad) AppendTool(role, content, toolName, toolArgs, toolResult string) error {
	sp.mu.Lock(); defer sp.mu.Unlock()
	if sp.sessionID == "" { return fmt.Errorf("no active session") }
	_, err := sp.stmtAppend.Exec(uuid.New().String(), sp.sessionID, role, content, toolName, toolArgs, toolResult)
	return err
}

func (sp *Scratchpad) GetAll() ([]ScratchEntry, error) {
	sp.mu.Lock(); defer sp.mu.Unlock()
	if sp.sessionID == "" { return nil, nil }
	rows, err := sp.stmtGetAll.Query(sp.sessionID)
	if err != nil { return nil, err }
	defer rows.Close()
	var entries []ScratchEntry
	for rows.Next() {
		var e ScratchEntry; var createdAt string
		if rows.Scan(&e.ID, &e.SessionID, &e.Role, &e.Content, &e.ToolName, &e.ToolArgs, &e.ToolResult, &createdAt) != nil { continue }
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, e)
	}
	return entries, nil
}

func (sp *Scratchpad) Drain() (*Session, []ScratchEntry, error) {
	sp.mu.Lock(); defer sp.mu.Unlock()
	if sp.sessionID == "" { return nil, nil, fmt.Errorf("no active session") }
	var sess Session; var createdAt, completedAt sql.NullString
	err := sp.db.QueryRow(`SELECT id, agent_name, project, status, prompt, system_ctx, created_at, completed_at FROM l1_sessions WHERE id = ?`, sp.sessionID).Scan(&sess.ID, &sess.AgentName, &sess.Project, &sess.Status, &sess.Prompt, &sess.SystemCtx, &createdAt, &completedAt)
	if err != nil { return nil, nil, fmt.Errorf("read session: %w", err) }
	if createdAt.Valid { sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String) }
	// Query entries directly (we already hold the lock, cannot call GetAll).
	rows, err := sp.stmtGetAll.Query(sp.sessionID)
	if err != nil { return nil, nil, err }
	defer rows.Close()
	var entries []ScratchEntry
	for rows.Next() {
		var e ScratchEntry; var ca string
		if rows.Scan(&e.ID, &e.SessionID, &e.Role, &e.Content, &e.ToolName, &e.ToolArgs, &e.ToolResult, &ca) != nil { continue }
		e.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		entries = append(entries, e)
	}
	now := time.Now().UTC(); sess.Status = StatusCompleted; sess.CompletedAt = &now
	sp.stmtComplete.Exec(StatusCompleted, sp.sessionID)
	return &sess, entries, nil
}

func (sp *Scratchpad) Transcript() (string, error) {
	entries, err := sp.GetAll()
	if err != nil { return "", err }
	var sb strings.Builder
	for _, e := range entries {
		if e.ToolName != "" {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n  Tool: %s(%s)\n  Result: %s\n", e.CreatedAt.Format("15:04:05"), e.Role, e.Content, e.ToolName, e.ToolArgs, truncate(e.ToolResult, 200)))
		} else {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", e.CreatedAt.Format("15:04:05"), e.Role, e.Content))
		}
	}
	return sb.String(), nil
}

func (sp *Scratchpad) Close() error {
	if sp.stmtAppend != nil { sp.stmtAppend.Close() }
	if sp.stmtGetAll != nil { sp.stmtGetAll.Close() }
	if sp.stmtComplete != nil { sp.stmtComplete.Close() }
	return sp.db.Close()
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n-3] + "..."
}
