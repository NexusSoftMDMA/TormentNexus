package tools

import (
	"context"
	"io/ioutil"
	"net/http"
)

func HandleListDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	if baseURL == "" || token == "" {
		return err("missing base_url or api_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/deployments", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok(string(body))
}

func HandleGetDeployment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "api_token")
	id, _ :=getString(args, "deployment_id")
	if baseURL == "" || token == "" || id == "" {
		return err("missing required arguments")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/deployments/"+id, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok(string(body))
}