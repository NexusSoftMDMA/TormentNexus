package mcpimpl

import (
	"context"
	"os"
	"path/filepath"
)

func HandleStoreMemory_jaumemory_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	content, _ :=getString(args, "content")
	if key == "" || content == "" {
		return err("key and content are required")
}

	dir := "memories"
	if e := os.MkdirAll(dir, 0755); e != nil {
		return err("failed to create memories dir: " + e.Error())
}

	filePath := filepath.Join(dir, key+".txt")
	if e := os.WriteFile(filePath, []byte(content), 0644); e != nil {
		return err("failed to write memory: " + e.Error())
}

	return ok("memory stored")
}

func HandleRetrieveMemory_jaumemory_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	filePath := filepath.Join("memories", key+".txt")
	data, e := os.ReadFile(filePath)
	if e != nil {
		return err("memory not found")
}

	return success(string(data))
}