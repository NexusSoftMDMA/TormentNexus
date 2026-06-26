package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetAdminInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("missing host")
}

	resp, e := http.DefaultClient.Get("https://" + host + "/admin/info")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse error: " + e.Error())
}

	return ok(string(body))
}

func HandleListAdmins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("missing host")
}

	resp, e := http.DefaultClient.Get("https://" + host + "/admin/list")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var list []interface{}
	if e := json.Unmarshal(body, &list); e != nil {
		return err("parse error: " + e.Error())
}

	return ok(string(body))
}