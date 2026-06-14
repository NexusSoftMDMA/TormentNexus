package tools

import (
	"context"
	"os"
)

func HandleCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	e := os.WriteFile(name, []byte(content), 0644)
	if e != nil {
		return err("failed to create file: " + e.Error())
}

	return ok("file created: " + name)
}

func HandleDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	e := os.Remove(name)
	if e != nil {
		return err("failed to delete file: " + e.Error())
}

	return ok("file deleted: " + name)
}