//go:build ignore
// +build ignore

package tools

import "context"

func HandleGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok("Git status for " + path + " retrieved")
}

func HandleGitLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok("Git log for " + path + " retrieved")
}
