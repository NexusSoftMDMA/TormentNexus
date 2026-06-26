package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchBugs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspaceID, _ :=getString(args, "workspace_id")
	if workspaceID == "" {
		return err("workspace_id is required")
}

	url := fmt.Sprintf("https://api.tapd.cn/bugs?workspace_id=%s", workspaceID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Bugs: %v", result))
}

func HandleGetBug(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bugID, _ :=getString(args, "bug_id")
	if bugID == "" {
		return err("bug_id is required")
}

	workspaceID, _ :=getString(args, "workspace_id")
	if workspaceID == "" {
		return err("workspace_id is required")
}

	url := fmt.Sprintf("https://api.tapd.cn/bugs/%s?workspace_id=%s", bugID, workspaceID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Bug: %v", result))
}