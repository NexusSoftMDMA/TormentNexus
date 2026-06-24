package tools

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

func HandleCreatePresentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	theme, _ :=getString(args, "theme")
	message := fmt.Sprintf("Created presentation with theme '%s':\n%s", theme, content)
	return ok(message)
}

func HandleExportPresentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	theme, _ :=getString(args, "theme")
	cmd := exec.Command("marp", "--theme", theme, "--html", "-")
	cmd.Stdin = strings.NewReader(content)
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("Marp export failed: %v", e))
}

	return success(string(output))
}