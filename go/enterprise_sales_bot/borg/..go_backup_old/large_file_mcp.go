package tools

import (
	"context"
	"os"
)

func HandleFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	fi, e := os.Stat(path)
	if e != nil {
		return err("failed to stat file: " + e.Error())
}

	return ok("file size: " + formatInt(fi.Size()) + " bytes")
}