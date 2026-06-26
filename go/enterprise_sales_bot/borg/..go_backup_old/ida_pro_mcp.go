package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleAnalyzeBinary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "binary_path")
	if path == "" {
		return err("binary_path is required")
}

	u := "http://localhost:8080/analyze?path=" + url.QueryEscape(path)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to call IDA: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}