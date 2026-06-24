package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	url := "http://localhost:8080/memory/" + key
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get memory: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok(string(body))
}

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	url := "http://localhost:8080/memory/" + key
	req, e := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(value))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "text/plain")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to set memory: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return success("memory set")
}