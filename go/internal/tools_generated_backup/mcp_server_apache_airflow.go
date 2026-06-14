package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleListDags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "base_url")
	if baseUrl == "" {
		return err("base_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dags", baseUrl), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	setBasicAuth(req, args)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("dags: %v", result["dags"]))
}

func HandleTriggerDag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "base_url")
	if baseUrl == "" {
		return err("base_url is required")
}

	dagId, _ :=getString(args, "dag_id")
	if dagId == "" {
		return err("dag_id is required")
}

	body := fmt.Sprintf(`{"conf": {}}`)
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/dags/%s/dagRuns", baseUrl, dagId), strings.NewReader(body)) //nolint:staticcheck
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	setBasicAuth(req, args)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("triggered: run_id=%v", result["dag_run_id"]))
}

func setBasicAuth(req *http.Request, args map[string]interface{}) {
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")
	if user != "" && pass != "" {
		req.SetBasicAuth(user, pass)

}
}