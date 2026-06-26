package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleIngestLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	level, _ :=getString(args, "level")
	if level == "" {
		level = "info"
	}
	source, _ :=getString(args, "source")
	if source == "" {
		source = "unknown"
	}
	payload := map[string]string{"message": message, "level": level, "source": source}
	b, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload")
}

	u := os.Getenv("LOG_SINK_URL")
	if u == "" {
		return err("LOG_SINK_URL not set")
}

	req, e := http.NewRequestWithContext(ctx, "POST", u+"/ingest", bytes.NewReader(b))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("log ingested")
}

func HandleAnalyzeLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := os.Getenv("LOG_SINK_URL")
	if u == "" {
		return err("LOG_SINK_URL not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", u+"/analyze?query="+url.QueryEscape(query), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}