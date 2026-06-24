package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleGetDragon retrieves a dragon by name.
func HandleGetDragon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	url := fmt.Sprintf("https://api.dragonmcp.example.com/dragons/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status)
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Dragon: %+v", data))
}

// HandleListDragons lists all dragons.
func HandleListDragons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.dragonmcp.example.com/dragons"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status)
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Dragons: %+v", data))
}