package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSearch_elasticsearch_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	index, _ :=getString(args, "index")
	query, _ :=getString(args, "query")
	if host == "" || index == "" {
		return err("host and index are required")
}

	url := fmt.Sprintf("%s/%s/_search", strings.TrimRight(host, "/"), index)
	bodyReader := strings.NewReader(query)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(respBody))
}

func HandleIndexDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	index, _ :=getString(args, "index")
	doc, _ :=getString(args, "document")
	id, _ :=getString(args, "id")
	if host == "" || index == "" || doc == "" {
		return err("host, index, and document are required")
}

	url := fmt.Sprintf("%s/%s/_doc", strings.TrimRight(host, "/"), index)
	if id != "" {
		url += "/" + id
	}
	bodyReader := strings.NewReader(doc)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(respBody))
}