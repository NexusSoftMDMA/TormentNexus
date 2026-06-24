package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListMcpServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch servers: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var servers []map[string]interface{}
	if e := json.Unmarshal(body, &servers); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	var list string
	for _, s := range servers {
		name, _ := s["name"].(string)
		list += fmt.Sprintf("- %s\n", name)

	if list == "" {
		list = "No servers found."
	}

	return ok(list)
}

}

func HandleGetMcpServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch server: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var server map[string]interface{}
	if e := json.Unmarshal(body, &server); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	desc, found := server["description"].(string)
	if !found {
		desc = "No description"
	}
	return ok(fmt.Sprintf("Server: %v\nDescription: %s", server["name"], desc))
}