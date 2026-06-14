package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleQueryDatabase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	databaseID, _ :=getString(args, "database_id")
	if databaseID == "" {
		return err("missing database_id")
}

	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		return err("NOTION_TOKEN not set")
}

	body := bytes.NewBufferString(`{}`)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/databases/"+databaseID+"/query", body)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleCreatePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	parentID, _ :=getString(args, "parent_id")
	if parentID == "" {
		return err("missing parent_id")
}

	title, _ :=getString(args, "title")
	if title == "" {
		return err("missing title")
}

	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		return err("NOTION_TOKEN not set")
}

	payload := map[string]interface{}{
		"parent": map[string]string{"database_id": parentID},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": title}},
				},
			},
		},
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(bodyBytes))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Created page: %s", string(respBody)))
}