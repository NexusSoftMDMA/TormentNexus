package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	title, _ :=getString(args, "title")
	token, _ :=getString(args, "token")
	payload := map[string]string{"project_id": projectID, "title": title}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.launchsecure.io/v1/feedback", bytes.NewReader(body))
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	return success("Feedback created")
}

func HandleListWorkItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.launchsecure.io/v1/projects/%s/work-items", projectID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("API returned status " + resp.Status)
}

	var result []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(fmt.Sprintf("Found %d work items", len(result)))
}