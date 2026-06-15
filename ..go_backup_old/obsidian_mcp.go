package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "apiUrl")
	if base == "" {
		base = "http://localhost:27123"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/notes", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Obsidian API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("reading response: " + e.Error())
}

	var notes []string
	if e := json.Unmarshal(body, &notes); e != nil {
		return err("JSON decode: " + e.Error())
}

	return ok(fmt.Sprintf("Notes: %v", notes))
}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "apiUrl")
	if base == "" {
		base = "http://localhost:27123"
	}
	name, _ :=getString(args, "name")
	if name == "" {
		return err("note name required")
}

	u := base + "/notes/" + url.PathEscape(name)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Obsidian API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("reading response: " + e.Error())
}

	return ok(string(body))
}