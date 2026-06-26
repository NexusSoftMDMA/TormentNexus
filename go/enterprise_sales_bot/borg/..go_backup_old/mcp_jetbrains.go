package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleOpenFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	project, _ :=getString(args, "project")
	baseURL := "http://localhost:63342/api"
	url := fmt.Sprintf("%s/open?file=%s", baseURL, filePath)
	if project != "" {
		url += "&project=" + project
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to open file: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("IDE returned status " + resp.Status)
}

	return ok("File opened successfully")
}

func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	baseURL := "http://localhost:63342/api"
	url := fmt.Sprintf("%s/run?command=%s", baseURL, command)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to run command: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("IDE returned status " + resp.Status)
}

	return ok("Command executed successfully")
}