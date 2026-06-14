package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleGetNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	nodeID, _ :=getInt(args, "node_id")
	if nodeID == 0 {
		return err("node_id is required")
}

	url := fmt.Sprintf("%s/node/%d?_format=json", baseURL, nodeID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch node: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Node %d: %v", nodeID, data["title"]))
}

func HandleListNodeTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/entity/node_type?_format=json", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch node types: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var types []map[string]interface{}
	if e := json.Unmarshal(body, &types); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	names := make([]string, 0, len(types))
	for _, t := range types {
		if name, found := t["name"]; found {
			names = append(names, fmt.Sprintf("%v", name))

	}
	return ok(fmt.Sprintf("Node types: %v", names))
}
}