package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetCodeContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}

func HandleAnalyzeCodeContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	lines := strings.Split(code, "\n")
	lineCount := len(lines)
	charCount := len(code)
	result := fmt.Sprintf("Lines: %d, Characters: %d", lineCount, charCount)
	return success(result)
}