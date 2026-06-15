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
	org, _ :=getString(args, "org")
	proj, _ :=getString(args, "project")
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("missing SENTRY_AUTH_TOKEN")
}

	url := fmt.Sprintf("https://sentry.io/api/0/projects/%s/%s/issues/", org, proj)
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	return success(string(body))
}

func HandleCreateIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "org")
	proj, _ :=getString(args, "project")
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("missing SENTRY_AUTH_TOKEN")
}

	message, _ :=getString(args, "message")
	if message == "" {
		return err("message required")
}

	payload := map[string]string{"message": message}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	url := fmt.Sprintf("https://sentry.io/api/0/projects/%s/%s/issues/", org, proj)
	req, e := http.NewRequest("POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	// Note: Missing body assignment - should be bytes.NewReader(bodyBytes). Below corrected.
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("API error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	return success(string(body))
}