package memorystore

import (
	"context"
)

type MemoryArchiver struct {
	workspaceRoot string
	vectorStore   *VectorStore
}

func NewMemoryArchiver(workspaceRoot string, vs *VectorStore) *MemoryArchiver {
	return &MemoryArchiver{
		workspaceRoot: workspaceRoot,
		vectorStore:   vs,
	}
}

func (a *MemoryArchiver) ArchiveAndExtract(ctx context.Context, sessionData map[string]interface{}) (any, error) {
	return nil, nil
}
