package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
	}
	resp, e := http.DefaultClient.Post("http://localhost:8080/execute", "application/json", strings.NewReader(command))
	if e != nil {
		return err("HTTP error: "+e.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
	}
	resp, e := http.DefaultClient.Get("http://localhost:8080/read?path=" + path)
	if e != nil {
		return err("HTTP error: "+e.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}