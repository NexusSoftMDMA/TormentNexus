package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type container struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
}

func HandleListContainers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost:2375"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "http://"+host+"/containers/json?all=true", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to list containers: " + e.Error())
}

	defer resp.Body.Close()
	var containers []container
	if e := json.NewDecoder(resp.Body).Decode(&containers); e != nil {
		return err("failed to decode response: " + e.Error())
}

	names := ""
	for _, c := range containers {
		for _, n := range c.Names {
			names += n[1:] + ", "
		}
	}
	if names == "" {
		return ok("No containers found")
}

	return ok("Containers: " + names[:len(names)-2])
}

func HandleStartContainer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost:2375"
	}
	id, _ :=getString(args, "container_id")
	if id == "" {
		return err("container_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://"+host+"/containers/"+id+"/start", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to start container: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return err("failed to start container, status: " + resp.Status)
}

	return success("Container " + id + " started")
}