package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleReadPDF(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	out, e := exec.Command("pdftotext", path, "-").Output()
	if e != nil {
		return err(fmt.Sprintf("failed to read PDF: %v", e))
}

	return ok(string(out))
}