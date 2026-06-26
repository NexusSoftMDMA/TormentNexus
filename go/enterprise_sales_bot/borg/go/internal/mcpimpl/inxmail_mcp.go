package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandlePing_inxmail_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleListContacts_inxmail_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url := "https://api.inxmail.com/v1/contacts"
	if name != "" {
		url += "?name=" + name
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch contacts: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}