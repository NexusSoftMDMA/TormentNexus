package mcpimpl

import (
	"context"
	"os"
)

func HandleCreateTmpFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prefix, _ :=getString(args, "prefix")
	f, e := os.CreateTemp("", prefix)
	if e != nil {
		return err("failed to create temp file: " + e.Error())
}

	defer f.Close()
	path := f.Name()
	return ok("Created temp file: " + path)
}