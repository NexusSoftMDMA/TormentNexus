package httpapi

import (
	"os"
	"path/filepath"
)

func localProjectContextPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".tormentnexus", "project_context.md")
}

func localWriteProjectContext(workspaceRoot string, content string) error {
	contextPath := localProjectContextPath(workspaceRoot)
	if err := os.MkdirAll(filepath.Dir(contextPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(contextPath, []byte(content), 0o644)
}
