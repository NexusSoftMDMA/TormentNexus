package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetMDKProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "path")
	if projectPath == "" {
		return err("missing project path")
	}
	url := "http://localhost:8080/api/mdk/project?path=" + projectPath
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return ok("project loaded")
}

func HandleValidateMDK(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "path")
	if projectPath == "" {
		return err("missing project path")
	}
	url := "http://localhost:8080/api/mdk/validate?path=" + projectPath
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	return success("validation complete")
}// touch 1781132140
