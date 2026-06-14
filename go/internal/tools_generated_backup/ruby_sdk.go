package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	servers := []string{"main", "staging", "dev"}
	data, e := json.Marshal(servers)
	if e != nil {
		return err("failed to marshal servers")
}

	return ok(string(data))
}

func HandleGetServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	resp, e := http.DefaultClient.Get("https://api.example.com/servers/" + name)
	if e != nil {
		return err("failed to fetch server")
}

	defer resp.Body.Close()
	return success("fetched server")
}