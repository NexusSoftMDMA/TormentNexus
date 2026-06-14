package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListFolder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	folderID, _ :=getString(args, "folder_id")
	if token == "" || folderID == "" {
		return err("token and folder_id are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.box.com/2.0/folders/%s/items", folderID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Items: %v", result["entries"]))
}

func HandleGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	fileID, _ :=getString(args, "file_id")
	if token == "" || fileID == "" {
		return err("token and file_id are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.box.com/2.0/files/%s", fileID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("File: %v", result))
}