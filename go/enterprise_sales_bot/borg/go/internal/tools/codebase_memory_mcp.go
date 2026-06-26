package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const memoryFile = ".codebase_memory.json"

// Memory represents a single codebase memory entry
type Memory struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	FilePath  string `json:"file_path"`
	Timestamp int64  `json:"timestamp"`
}

// loadMemories reads the memory store from disk
func loadMemories() ([]Memory, error) {
	data, readErr := os.ReadFile(memoryFile)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return []Memory{}, nil
		}
		return nil, readErr
	}

	var mems []Memory
	parseErr := json.Unmarshal(data, &mems)
	if parseErr != nil {
		return nil, parseErr
	}
	return mems, nil
}

// saveMemories writes the memory store to disk
func saveMemories(mems []Memory) error {
	data, jsonErr := json.MarshalIndent(mems, "", "  ")
	if jsonErr != nil {
		return jsonErr
	}
	writeErr := os.WriteFile(memoryFile, data, 0644)
	return writeErr
}

// HandleAddMemory adds a new code snippet to memory
func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	filePath, _ :=getString(args, "file_path")

	if content == "" {
		return err("content is required")
}

	mems, loadErr := loadMemories()
	if loadErr != nil {
		return err(fmt.Sprintf("failed to load memories: %s", loadErr.Error()))
}

	newMem := Memory{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   content,
		FilePath:  filePath,
		Timestamp: time.Now().Unix(),
	}

	mems = append(mems, newMem)

	saveErr := saveMemories(mems)
	if saveErr != nil {
		return err(fmt.Sprintf("failed to save memory: %s", saveErr.Error()))
}

	return ok("memory added successfully")
}