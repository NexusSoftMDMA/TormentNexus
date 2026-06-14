package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	url := fmt.Sprintf("%s/api/issues?project=%s", os.Getenv("YOUTRACK_URL"), project)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("YOUTRACK_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("unmarshal failed")
}

	return ok(fmt.Sprintf("Issues: %v", data))
}

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("%s/api/issues/%s", os.Getenv("YOUTRACK_URL"), id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("YOUTRACK_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("unmarshal failed")
}

	return ok(fmt.Sprintf("Issue: %v", data))
}