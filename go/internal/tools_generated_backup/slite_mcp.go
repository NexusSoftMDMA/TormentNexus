package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchSlite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if query == "" || apiKey == "" {
		return err("query and apiKey are required")
}

	url := fmt.Sprintf("https://api.slite.com/v1/search?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleGetSliteNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	noteID, _ :=getString(args, "noteId")
	apiKey, _ :=getString(args, "apiKey")
	if noteID == "" || apiKey == "" {
		return err("noteId and apiKey are required")
}

	url := fmt.Sprintf("https://api.slite.com/v1/notes/%s", noteID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var note interface{}
	if e := json.Unmarshal(body, &note); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Note: %v", note))
}