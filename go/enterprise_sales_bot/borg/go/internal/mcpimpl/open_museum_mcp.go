package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleSearchMuseum(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.example.com/museum/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search museum: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("Search results: " + string(body))
}

func HandleGetObject_open_museum_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://api.example.com/museum/object/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get object: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok("Object data: " + string(body))
}