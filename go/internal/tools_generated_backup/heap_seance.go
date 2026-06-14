package tools

import (
	"context"
)

func HandleAnalyzeHeap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "heap_dump")
	return ok("Heap analysis started for: " + path)
}