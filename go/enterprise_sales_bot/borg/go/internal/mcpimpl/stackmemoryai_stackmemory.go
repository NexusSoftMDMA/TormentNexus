package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleStackmemorySave(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	baseURL := os.Getenv("STACKMEMORY_URL")
	if baseURL == "" {
		return err("STACKMEMORY_URL not set")
}

	body, e := json.Marshal(map[string]string{"project": project, "key": key, "value": value})
	if e != nil {
		return err("marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post(baseURL+"/save", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("post: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("status: " + resp.Status)
}

	return ok("saved")
}

func HandleStackmemoryQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	query, _ :=getString(args, "query")
	baseURL := os.Getenv("STACKMEMORY_URL")
	if baseURL == "" {
		return err("STACKMEMORY_URL not set")
}

	body, e := json.Marshal(map[string]string{"project": project, "query": query})
	if e != nil {
		return err("marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post(baseURL+"/query", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("post: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
}

	return success(string(data))
}// touch 1781132141
