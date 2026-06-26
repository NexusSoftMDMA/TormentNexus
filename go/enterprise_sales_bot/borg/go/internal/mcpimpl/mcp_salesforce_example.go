package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSalesforceQuery_mcp_salesforce_example(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceURL, _ :=getString(args, "instance_url")
	accessToken, _ :=getString(args, "access_token")
	query, _ :=getString(args, "query")
	if instanceURL == "" || accessToken == "" || query == "" {
		return err("Missing required arguments: instance_url, access_token, query")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/services/data/v58.0/query?q=%s", instanceURL, query), nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Salesforce returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok("Query response: " + string(body))
}

	return success("Query executed successfully: " + string(body))
}