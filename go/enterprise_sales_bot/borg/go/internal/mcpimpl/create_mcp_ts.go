package mcpimpl

import (
	"context"
	"fmt"
	"os"
)

func HandleCreateMcpTs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	if name == "" {
		return err("name is required")
}

	e := os.WriteFile(name+".ts", []byte(content), 0644)
	if e != nil {
		return err(fmt.Sprintf("failed to write file: %v", e))
}

	return ok("created " + name + ".ts")
}