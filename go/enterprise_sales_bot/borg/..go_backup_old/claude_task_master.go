package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleCreateTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.PostForm("http://localhost:8000/tasks", url.Values{"name": {name}})
	if e != nil {
		return err("create failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return ok("task created: " + name)
}

func HandleListTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8000/tasks")
	if e != nil {
		return err("list failed: " + e.Error())
}

	defer resp.Body.Close()
	var tasks []string
	if e := json.NewDecoder(resp.Body).Decode(&tasks); e != nil {
		return err("decode failed: " + e.Error())
}

	// Convert to []interface{} for ToolResponse success
	items := make([]interface{}, len(tasks))
	for i, t := range tasks {
		items[i] = t
	}
	return success("tasks retrieved", items)
}