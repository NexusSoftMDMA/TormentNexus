package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListOperators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/operators")
	if e != nil {
		return err(fmt.Sprintf("failed to list operators: %v", e))
}

	defer resp.Body.Close()
	var result []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("Found %d operators", len(result)))
}

func HandleSetOperatorParameter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	opID, _ :=getString(args, "operator_id")
	param, _ :=getString(args, "parameter")
	value, _ :=getString(args, "value")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	if opID == "" || param == "" {
		return err("operator_id and parameter are required")
}

	url := fmt.Sprintf("%s/operator/%s/parameter/%s?value=%s", baseURL, opID, param, value)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to set parameter: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("received status %d", resp.StatusCode))
}

	return success("parameter set")
}