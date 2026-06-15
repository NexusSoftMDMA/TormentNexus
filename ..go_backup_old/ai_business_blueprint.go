package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetBlueprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("blueprint id is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/blueprints/" + id)
	if e != nil {
		return err("failed to fetch blueprint: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleCreateBlueprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("blueprint name is required")
}

	payload, e := json.Marshal(map[string]string{"name": name})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.example.com/blueprints", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create blueprint: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}