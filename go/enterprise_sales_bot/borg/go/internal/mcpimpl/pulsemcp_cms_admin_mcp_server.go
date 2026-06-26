package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListContent_pulsemcp_cms_admin_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contentType, _ :=getString(args, "type")
	url := fmt.Sprintf("http://localhost:8080/api/content?type=%s", contentType)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch content: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("Content list: %v", data))
}

func HandleCreateContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	bodyContent, _ :=getString(args, "body")
	payload := map[string]string{"title": title, "body": bodyContent}
	jsonData, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal JSON: " + e.Error())
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/api/content", "application/json", bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to create content: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Content created successfully")
}