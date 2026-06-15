package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleFetchGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("Missing url argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleAnalyzeGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("Missing code argument")
}

	lines := 0
	for _, c := range code {
		if c == '\n' {
			lines++
		}
	}
	result := map[string]interface{}{
		"lines": lines,
		"chars": len(code),
	}
	jsonBytes, e := json.Marshal(result)
	if e != nil {
		return err("Failed to marshal result")
}

	return ok(string(jsonBytes))
}