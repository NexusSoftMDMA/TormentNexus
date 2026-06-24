package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateWorld(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080"
	}
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	body, e := json.Marshal(map[string]string{"name": name, "description": desc})
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	resp, e := http.Post(url+"/worlds", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create world: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	return ok("world created")
}

func HandleListWorlds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080"
	}
	resp, e := http.Get(url + "/worlds")
	if e != nil {
		return err(fmt.Sprintf("failed to list worlds: %v", e))
}

	defer resp.Body.Close()
	var worlds []interface{}
	if e := json.NewDecoder(resp.Body).Decode(&worlds); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	return success(fmt.Sprintf("worlds: %v", worlds))
}