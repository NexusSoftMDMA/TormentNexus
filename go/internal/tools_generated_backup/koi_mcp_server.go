package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleVulnerabilitiesList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.koi.security/v1/vulnerabilities"
	severity, _ :=getString(args, "severity")
	status, _ :=getString(args, "status")
	if severity != "" {
		url += "?severity=" + severity
	}
	if status != "" {
		if severity != "" {
			url += "&status=" + status
		} else {
			url += "?status=" + status
		}
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("vulnerabilities: %s", string(body)))
}

func HandleVulnerabilityGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing 'id'")
}

	url := fmt.Sprintf("https://api.koi.security/v1/vulnerabilities/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success(fmt.Sprintf("vulnerability: %s", string(body)))
}