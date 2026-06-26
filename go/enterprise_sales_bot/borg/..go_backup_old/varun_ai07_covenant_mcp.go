package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url required")
}

	resp, e := http.DefaultClient.Get(base + "/tasks")
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var tasks []interface{}
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err("decode error: " + e.Error())
}

	return ok("fetched tasks")
}

func HandleGetTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	id, _ :=getString(args, "id")
	if base == "" || id == "" {
		return err("base_url and id required")
}

	resp, e := http.DefaultClient.Get(base + "/tasks/" + id)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var task map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&task); e != nil {
		return err("decode error: " + e.Error())
}

	return ok("fetched task")
}