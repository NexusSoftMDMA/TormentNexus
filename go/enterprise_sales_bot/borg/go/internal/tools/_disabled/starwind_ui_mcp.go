package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleFetchUI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint is required")
}

	resp, e := http.DefaultClient.Get("https://starwind-ui.example.com/" + endpoint)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}

func HandleCreateUIComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	component, _ :=getString(args, "component")
	if name == "" || component == "" {
		return err("name and component are required")
}

	payload, e := json.Marshal(map[string]string{"name": name, "component": component})
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://starwind-ui.example.com/api/create", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("create failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Component created")
}