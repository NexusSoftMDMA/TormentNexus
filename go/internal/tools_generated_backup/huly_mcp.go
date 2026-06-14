package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		server = "https://huly.app/api"
	}
	url := server + "/projects"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch projects: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Projects: %v", data))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "id")
	if projectID == "" {
		return err("missing project id")
}

	server, _ :=getString(args, "server")
	if server == "" {
		server = "https://huly.app/api"
	}
	url := server + "/projects/" + projectID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch project: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Project: %v", data))
}