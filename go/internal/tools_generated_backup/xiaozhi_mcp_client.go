package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleXiaozhiConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	port, _ :=getInt(args, "port")
	if name == "" {
		return err("missing name")
	}
	config := map[string]interface{}{"name": name, "port": port, "status": "active"}
	data, e := json.Marshal(config)
	if e != nil {
		return err("marshal failed")
	}
	return success(string(data))
}

func HandleXiaozhiManage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("missing action")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost/manage", nil)
	if e != nil {
		return err("request failed")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do failed")
	}
	defer resp.Body.Close()
	return ok("managed")
}