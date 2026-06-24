package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "instance_url")
	token, _ :=getString(args, "api_token")
	if base == "" || token == "" {
		return err("Missing required args: instance_url, api_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/v1/courses", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var courses []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&courses); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	msg := ""
	for _, c := range courses {
		msg += fmt.Sprintf("- %s (ID: %d)\n", c.Name, c.ID)

	if msg == "" {
		msg = "No courses found."
	}
	return ok(msg)
}
}