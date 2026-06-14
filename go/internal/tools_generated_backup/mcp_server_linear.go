package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleLinearListTeams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	query := `{"query":"{ teams { nodes { id name } } }"}`
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.linear.app/graphql", bytes.NewBufferString(query))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	teams, found := data["teams"].(map[string]interface{})
	if !found {
		return err("no teams in response")
}

	nodes, found := teams["nodes"].([]interface{})
	if !found {
		return err("no nodes in teams")
}

	teamNames := []string{}
	for _, n := range nodes {
		team, found := n.(map[string]interface{})
		if found {
			if name, found := team["name"].(string); found {
				teamNames = append(teamNames, name)

		}
	}
	return ok(fmt.Sprintf("Linear teams: %v", teamNames))
}
}