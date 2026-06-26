package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTriggerJob_jenkins_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jenkinsURL, _ :=getString(args, "jenkins_url")
	jobName, _ :=getString(args, "job_name")
	token, _ :=getString(args, "token")
	if jenkinsURL == "" || jobName == "" {
		return err("jenkins_url and job_name are required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/job/%s/build", jenkinsURL, jobName), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	return ok("Build triggered successfully")
}

}

func HandleGetJobStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jenkinsURL, _ :=getString(args, "jenkins_url")
	jobName, _ :=getString(args, "job_name")
	token, _ :=getString(args, "token")
	if jenkinsURL == "" || jobName == "" {
		return err("jenkins_url and job_name are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/job/%s/api/json", jenkinsURL, jobName), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	var result struct {
		LastBuild struct {
			Number int    `json:"number"`
			Result string `json:"result"`
		} `json:"lastBuild"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	status := fmt.Sprintf("Job '%s' last build #%d result: %s", jobName, result.LastBuild.Number, result.LastBuild.Result)
	return success(status)
}
}