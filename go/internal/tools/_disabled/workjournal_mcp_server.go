package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleCreateJournalEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	base := os.Getenv("WORKJOURNAL_API_BASE")
	if base == "" {
		base = "https://api.workjournal.dev"
	}
	body, e := json.Marshal(map[string]string{"title": title, "content": content})
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", base+"/entries", jsonBody(body))
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok("Journal entry created successfully")
}

func HandleListJournalEntries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("WORKJOURNAL_API_BASE")
	if base == "" {
		base = "https://api.workjournal.dev"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/entries", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var entries []map[string]interface{}
	if e := json.Unmarshal(body, &entries); e != nil {
		return err(fmt.Sprintf("json unmarshal failed: %v", e))
}

	return ok(fmt.Sprintf("Found %d entries", len(entries)))
}

func jsonBody(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return io.NopCloser(bytes.NewReader(b))
}