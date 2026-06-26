package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchProfiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("http://localhost:8181/cxs/profiles/search?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Unomi: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Unomi returned status: " + resp.Status)
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleGetProfile_inoyu_mcp_unomi_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("http://localhost:8181/cxs/profiles/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Unomi: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Unomi returned status: " + resp.Status)
}

	var profile map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&profile); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.MarshalIndent(profile, "", "  ")
	return ok(string(data))
}