package mcpimpl

import (
	"context"
	"net/http"
)

func HandleGetCodelens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	req, e := http.NewRequest("GET", "http://localhost:5000/codelens?path="+path, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok("Codelens retrieved")
}

func HandleAnalyzeCode_roslyn_codelens_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	req, e := http.NewRequest("POST", "http://localhost:5000/analyze", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return success("Code analyzed")
}