package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleGetProblems fetches Dynatrace problems
func HandleGetProblems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	envURL, _ :=getString(args, "environmentUrl")
	apiToken, _ :=getString(args, "apiToken")
	if envURL == "" || apiToken == "" {
		return err("environmentUrl and apiToken are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", envURL+"/api/v2/problems", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Api-Token "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d problems", int(result["totalCount"].(float64))))
}

// HandleGetEntities retrieves Dynatrace entities
func HandleGetEntities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	envURL, _ :=getString(args, "environmentUrl")
	apiToken, _ :=getString(args, "apiToken")
	entityType, _ :=getString(args, "entityType")
	if envURL == "" || apiToken == "" {
		return err("environmentUrl and apiToken are required")
}

	url := envURL + "/api/v2/entities"
	if entityType != "" {
		url += "?entitySelector=type(" + entityType + ")"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Api-Token "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	totalCount := 0
	if tc, found := result["totalCount"].(float64); found {
		totalCount = int(tc)

	return ok(fmt.Sprintf("Found %d entities", totalCount))
}
}