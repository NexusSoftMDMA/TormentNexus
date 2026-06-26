package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	body, _ := json.Marshal(map[string]string{"title": title, "content": content})
	req, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost:27123/notes", bytes.NewReader(body))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("create note failed: %s", string(respBody)))
}

	return ok(fmt.Sprintf("Note '%s' created", title))
}

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:27123/search/%s", query), nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("search failed: %s", string(respBody)))
}

	var results []string
	json.Unmarshal(respBody, &results)
	return ok(fmt.Sprintf("Found %d notes", len(results)))
}