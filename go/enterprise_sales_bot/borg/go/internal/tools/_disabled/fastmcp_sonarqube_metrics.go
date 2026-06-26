package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetProjectMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectKey, _ :=getString(args, "projectKey")
	metricKeys, _ :=getString(args, "metricKeys")
	if projectKey == "" || metricKeys == "" {
		return err("projectKey and metricKeys are required")
}

	host := os.Getenv("SONAR_HOST_URL")
	token := os.Getenv("SONAR_TOKEN")
	url := fmt.Sprintf("%s/api/measures/component?component=%s&metricKeys=%s", host, projectKey, metricKeys)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(token, "")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host := os.Getenv("SONAR_HOST_URL")
	token := os.Getenv("SONAR_TOKEN")
	url := fmt.Sprintf("%s/api/projects/search", host)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(token, "")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}