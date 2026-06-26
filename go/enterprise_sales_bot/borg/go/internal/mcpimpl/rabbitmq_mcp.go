package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListQueues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	if baseURL == "" {
		return err("base_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/queues", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(username, password)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandlePublishMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	vhost, _ :=getString(args, "vhost")
	queue, _ :=getString(args, "queue")
	message, _ :=getString(args, "message")
	if baseURL == "" || vhost == "" || queue == "" || message == "" {
		return err("base_url, vhost, queue, and message are required")
}

	url := baseURL + "/api/exchanges/" + vhost + "/amq.default/publish"
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok("message published")
}