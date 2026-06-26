package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCourses_canvas_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		base = "https://canvas.instructure.com"
	}
	token, _ :=getString(args, "apiToken")
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/v1/courses", nil)
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
		return err("read failed: " + e.Error())
}

	var courses []map[string]interface{}
	if e := json.Unmarshal(body, &courses); e != nil {
		return err("parse failed: " + e.Error())
}

	result, _ := json.Marshal(courses)
	return ok(string(result))
}

func HandleGetAssignments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		base = "https://canvas.instructure.com"
	}
	token, _ :=getString(args, "apiToken")
	courseID, _ :=getInt(args, "courseId")
	url := fmt.Sprintf("%s/api/v1/courses/%d/assignments", base, courseID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("read failed: " + e.Error())
}

	var assignments []map[string]interface{}
	if e := json.Unmarshal(body, &assignments); e != nil {
		return err("parse failed: " + e.Error())
}

	result, _ := json.Marshal(assignments)
	return ok(string(result))
}