package mcpimpl

import (
	"context"
	"fmt"
	"syscall"
)

func HandleDf(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "/"
	}
	var stat syscall.Statfs_t
	e := syscall.Statfs(path, &stat)
	if e != nil {
		return err(fmt.Sprintf("failed to statfs: %v", e))
}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	msg := fmt.Sprintf("Path: %s\nTotal: %d bytes\nUsed: %d bytes\nFree: %d bytes", path, total, used, free)
	return ok(msg)
}