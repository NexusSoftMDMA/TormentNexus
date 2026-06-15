//go:build ignore
// +build ignore

package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleCloneDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		repo = "https://github.com/AztecProtocol/aztec-docs"
	}
	cmd := exec.CommandContext(ctx, "git", "clone", repo, "aztec-docs")
	e := cmd.Run()
	if e != nil {
		return err("failed to clone: " + e.Error())
	}
	return ok("repo cloned")
}

func HandleSearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
	}
	cmd := exec.CommandContext(ctx, "grep", "-r", query, "aztec-docs")
	out, e := cmd.Output()
	if e != nil {
		return err("search failed: " + e.Error())
	}
	results := strings.TrimSpace(string(out))
	return success(results)
}
