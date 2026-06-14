package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleListCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("CANVAS_API_URL")
	token := os.Getenv("CANVAS_API_TOKEN")
	if baseURL == "" || token == "" {
		return err("missing CANVAS_API_URL or CANVAS_API_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/courses", nil)
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
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleListAssignments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	courseID, _ :=getString(args, "course_id")
	if courseID == "" {
		return err("course_id is required")
}

	baseURL := os.Getenv("CANVAS_API_URL")
	token := os.Getenv("CANVAS_API_TOKEN")
	if baseURL == "" || token == "" {
		return err("missing CANVAS_API_URL or CANVAS_API_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/courses/"+courseID+"/assignments", nil)
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
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}