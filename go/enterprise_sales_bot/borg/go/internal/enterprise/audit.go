package enterprise

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// AuditEvent represents a single auditable action in the enterprise tier.
type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Result    string    `json:"result"`
	Metadata  any       `json:"metadata,omitempty"`
}

// Auditor handles the generation and storage of audit logs.
type Auditor struct {
	LogDir string
}

// NewAuditor creates a new enterprise auditor.
func NewAuditor(workspaceRoot string) *Auditor {
	logDir := filepath.Join(workspaceRoot, ".tormentnexus", "audit")
	_ = os.MkdirAll(logDir, 0755)
	return &Auditor{LogDir: logDir}
}

// Log records an enterprise audit event.
func (a *Auditor) Log(event AuditEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, _ := json.Marshal(event)

	// Structured logging to stdout
	log.Printf("[AUDIT] %s", string(data))

	// Persist to enterprise audit JSONL file
	logPath := filepath.Join(a.LogDir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		_, _ = f.WriteString(string(data) + "\n")
	}
}

// LogToolExecution records the execution of a native tool.
func (a *Auditor) LogToolExecution(userID string, toolName string, args any, result string) {
	a.Log(AuditEvent{
		UserID:   userID,
		Action:   "EXECUTE_TOOL",
		Resource: toolName,
		Result:   result,
		Metadata: args,
	})
}
