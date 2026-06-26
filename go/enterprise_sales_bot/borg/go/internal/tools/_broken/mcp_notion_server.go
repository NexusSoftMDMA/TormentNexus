package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListDatabases_mcp_notion_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("NOTION_TOKEN")

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.notion.com/v1/databases", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(string(body))
}

}

func HandleCreatePage_mcp_notion_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("NOTION_TOKEN")

	parentID, _ :=getString(args, "parent_id")
	propsStr, _ :=getString(args, "properties")
	if parentID == "" || propsStr == "" {
		return err("parent_id and properties are required")
	}
	var properties interface{}
	if e := json.Unmarshal([]byte(propsStr), &properties); e != nil {
		return err(fmt.Sprintf("invalid properties JSON: %v", e))
	}
	payload := map[string]interface{}{
		"parent":     map[string]interface{}{"database_id": parentID},
		"properties": properties,
	}
	bodyBytes, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/pages", bytes.NewReader(bodyBytes))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return ok(string(respBody))
}
}