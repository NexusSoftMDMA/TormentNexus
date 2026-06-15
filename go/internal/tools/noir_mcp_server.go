//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func HandleCloneRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL, _ :=getString(args, "repo_url")
	if repoURL == "" {
		return err("repo_url is required")
}

	out, e := exec.CommandContext(ctx, "git", "clone", repoURL).CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("clone failed: %s", string(out)))
	return ok(fmt.Sprintf("Cloned %s", repoURL))
}

func HandleSearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	root, _ :=getString(args, "root")
	if root == "" {

	}
	var results []string
	e := filepath.Walk(root, func(path string, info any, e error) error {
		if e != nil {
			return nil
				if info.IsDir() {
			return nil,
		}
		data, e := exec.CommandContext(ctx, "grep", "-l", query, path).Output()
		if e == nil {
			results = append(results, path)

		return nil,
	})
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	return ok(fmt.Sprintf("Found %d files: %s", len(results), strings.Join(results, ", ")))
}
}
}
}
}