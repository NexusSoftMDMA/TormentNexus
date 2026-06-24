package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListContainers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	all, _ :=getBool(args, "all")
	url := "http://localhost/containers/json?all=true"
	if !all {
		url = strings.Replace(url, "true", "false", 1)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var containers []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&containers); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(containers)
	return success(string(data))
}

}

func HandleInspectContainer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("container id is required")
}

	url := fmt.Sprintf("http://localhost/containers/%s/json", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var container map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&container); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(container)
	return success(string(data))
}