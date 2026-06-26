package memorystore

import (
	"context"
	"time"
)

type MemoryReactor struct {
	workspaceRoot string
	vectorStore   *VectorStore
}

func NewMemoryReactor(workspaceRoot string, vs *VectorStore) *MemoryReactor {
	r := &MemoryReactor{
		workspaceRoot: workspaceRoot,
		vectorStore:   vs,
	}
	go r.startDecayLoop()
	return r
}

func (r *MemoryReactor) startDecayLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if r.vectorStore != nil {
			_ = r.vectorStore.ApplyDecay(context.Background())
		}
	}
}

func (r *MemoryReactor) HandleFileChange(ctx context.Context, path string, content string) error {
	return nil
}

func (r *MemoryReactor) VectorStore() *VectorStore {
	return r.vectorStore
}
