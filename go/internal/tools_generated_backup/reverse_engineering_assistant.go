package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleAnalyzeBinary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch binary: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(fmt.Sprintf("Fetched %d bytes, first bytes: %x", len(data), data[:16]))
}

func HandleStaticAnalysis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing path argument")
}

	return ok(fmt.Sprintf("Static analysis for %s: no issues found (simulated)", path))
}