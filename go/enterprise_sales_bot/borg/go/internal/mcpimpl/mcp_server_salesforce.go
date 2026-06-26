package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQuerySalesforce(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	accessToken, _ :=getString(args, "accessToken")
	instanceURL, _ :=getString(args, "instanceURL")
	if query == "" || accessToken == "" || instanceURL == "" {
		return err("missing required parameters: query, accessToken, instanceURL")
}

	req, e := http.NewRequestWithContext(ctx, "GET", instanceURL+"/services/data/v58.0/query?q="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
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
		return err(fmt.Sprintf("Salesforce API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	records, found := result["records"]
	if !found {
		return ok("query returned no records")
}

	return success(fmt.Sprintf("records: %v", records))
}