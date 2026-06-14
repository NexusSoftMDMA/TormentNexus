package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleListCourses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiToken, _ :=getString(args, "apiToken")
	baseURL, _ :=getString(args, "baseURL")
	if apiToken == "" || baseURL == "" {
		return err("apiToken and baseURL are required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/courses", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var courses []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&courses); e != nil {
		return err("decode failed: " + e.Error())
	}
	names := make([]string, 0, len(courses))
	for _, c := range courses {
		if name, found := c["name"].(string); found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Found %d courses: %s", len(names), strings.Join(names, ", ")))
}

}

func HandleListAssignments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiToken, _ :=getString(args, "apiToken")
	baseURL, _ :=getString(args, "baseURL")
	courseID, _ :=getString(args, "courseID")
	if apiToken == "" || baseURL == "" || courseID == "" {
		return err("apiToken, baseURL, and courseID are required")
	}
	u, e := url.Parse(baseURL + "/api/v1/courses/" + courseID + "/assignments")
	if e != nil {
		return err("invalid URL: " + e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var assignments []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&assignments); e != nil {
		return err("decode failed: " + e.Error())
	}
	names := make([]string, 0, len(assignments))
	for _, a := range assignments {
		if name, found := a["name"].(string); found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Found %d assignments: %s", len(names), strings.Join(names, ", ")))
}
}