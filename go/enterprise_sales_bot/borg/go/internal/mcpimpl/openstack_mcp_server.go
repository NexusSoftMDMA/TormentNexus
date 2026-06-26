package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListServers_openstack_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	token, _ :=getString(args, "token")
	projectID, _ :=getString(args, "project_id")
	if baseURL == "" || token == "" {
		return err("missing required args: url, token")
}

	reqURL := fmt.Sprintf("%s/servers?project_id=%s", baseURL, projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Auth-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleGetServer_openstack_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	token, _ :=getString(args, "token")
	serverID, _ :=getString(args, "server_id")
	if baseURL == "" || token == "" || serverID == "" {
		return err("missing required args: url, token, server_id")
}

	reqURL := fmt.Sprintf("%s/servers/%s", baseURL, serverID)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Auth-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}