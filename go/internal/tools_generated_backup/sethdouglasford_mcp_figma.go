package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	token, _ :=getString(args, "token")
	if fileKey == "" || token == "" {
		return err("file_key and token are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey), nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleGetProjectFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	token, _ :=getString(args, "token")
	if projectID == "" || token == "" {
		return err("project_id and token are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/projects/%s/files", projectID), nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}