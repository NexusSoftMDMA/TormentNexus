package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleListJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.trigger.dev/v1/jobs", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}

func HandleTriggerJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	jobId, _ :=getString(args, "jobId")
	if jobId == "" {
		return err("jobId is required")
}

	payload, _ :=getString(args, "payload")
	url := fmt.Sprintf("https://api.trigger.dev/v1/jobs/%s/trigger", jobId)
	var bodyReader *strings.Reader
	if payload != "" {
		bodyReader = strings.NewReader(payload)

	req, e := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}
}