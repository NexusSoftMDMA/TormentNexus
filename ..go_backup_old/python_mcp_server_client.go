package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/docs?query=" + query)
	if e != nil {
		return err("failed to fetch documentation")
}

	defer resp.Body.Close()

	var data interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	return success("documentation fetched successfully: " + query + " data")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agent, _ :=getString(args, "agent")
	if agent == "" {
		return err("agent is required")
}

	return success("agent processed: " + agent)
}