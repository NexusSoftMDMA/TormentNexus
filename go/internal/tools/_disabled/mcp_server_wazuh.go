package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host := os.Getenv("WAZUH_HOST")
	user := os.Getenv("WAZUH_USER")
	pass := os.Getenv("WAZUH_PASS")
	if host == "" || user == "" || pass == "" {
		return err("missing WAzuH_HOST, WAzuH_USER, or WAzuH_PASS environment variables")
}

	url := fmt.Sprintf("%s/agents", host)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	if agentID == "" {
		return err("missing required argument 'agent_id'")
}

	host := os.Getenv("WAZUH_HOST")
	user := os.Getenv("WAZUH_USER")
	pass := os.Getenv("WAZUH_PASS")
	if host == "" || user == "" || pass == "" {
		return err("missing WAzuH_HOST, WAzuH_USER, or WAzuH_PASS environment variables")
}

	url := fmt.Sprintf("%s/agents/%s", host, agentID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}