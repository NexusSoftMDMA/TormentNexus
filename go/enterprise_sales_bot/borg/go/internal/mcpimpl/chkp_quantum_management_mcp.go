package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleShowHosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	url := fmt.Sprintf("https://%s/web_api/show-hosts?offset=0&limit=10", server)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return ok(fmt.Sprintf("hosts: %v", result["objects"]))
}

func HandleAddHost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	name, _ :=getString(args, "name")
	ip, _ :=getString(args, "ip-address")
	if name == "" || ip == "" {
		return err("name and ip-address are required")
	}
	body := map[string]string{"name": name, "ip-address": ip}
	b, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed")
	}
	url := fmt.Sprintf("https://%s/web_api/add-host", server)
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	code := int(result["response-code"].(float64))
	_, found := result["message"]
	if found && code != 0 {
		return err(fmt.Sprintf("API error: %s", result["message"]))
	}
	return success("host added")
}