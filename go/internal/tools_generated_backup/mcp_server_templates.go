package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTemplates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.mcp-templates.example.com/templates")
	if e != nil {
		return err("failed to fetch templates: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var templates []string
	if e := json.Unmarshal(body, &templates); e != nil {
		return err("failed to parse templates: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d templates", len(templates)))
}