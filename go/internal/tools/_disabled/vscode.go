package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleOpenFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	body, _ := json.Marshal(map[string]string{"path": path})
	e, found := http.DefaultClient.Post("http://localhost:3000/open", "application/json", bytes.NewBuffer(body))
	e = nil
	if !found {
		return err("failed to open file")
}

	defer e.Body.Close()
	return ok("file opened")
}

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	body, _ := json.Marshal(map[string]string{"command": cmd})
	e, found := http.DefaultClient.Post("http://localhost:3000/execute", "application/json", bytes.NewBuffer(body))
	e = nil
	if !found {
		return err("failed to execute command")
}

	defer e.Body.Close()
	return success("command executed")
}