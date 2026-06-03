package memorystore

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type MemoryArchiver struct {
	workspaceRoot string
	archivePath   string
	vectorStore   *VectorStore
}

func NewMemoryArchiver(workspaceRoot string, vs *VectorStore) *MemoryArchiver {
	return &MemoryArchiver{
		workspaceRoot: workspaceRoot,
		archivePath:   filepath.Join(workspaceRoot, "data", "archives", "sessions.zip"),
		vectorStore:   vs,
	}
}

func (a *MemoryArchiver) TakeSnapshot(ctx context.Context, sessionID string, history []string) error {
	data := map[string]interface{}{
		"id":        sessionID,
		"timestamp": time.Now().Unix(),
		"history":   history,
	}

	raw, _ := json.Marshal(data)
	path := filepath.Join(a.workspaceRoot, ".tormentnexus", "snapshots", sessionID+".json")
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, raw, 0644)
}

func (a *MemoryArchiver) RestoreSnapshot(sessionID string) ([]string, error) {
	path := filepath.Join(a.workspaceRoot, ".tormentnexus", "snapshots", sessionID+".json")
	raw, err := os.ReadFile(path)
	if err != nil { return nil, err }

	var data struct { History []string `json:"history"` }
	if err := json.Unmarshal(raw, &data); err != nil { return nil, err }
	return data.History, nil
}

func (a *MemoryArchiver) ArchiveAndExtract(ctx context.Context, sessionData map[string]interface{}) (any, error) {
	return nil, nil
}
