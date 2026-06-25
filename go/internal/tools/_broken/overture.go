package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HandleOvertureSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	client := http.DefaultClient
	apiURL := "https://api.github.com/search/repositories?q=" + url.QueryEscape(query)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Overture-MCP-Tool")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch data: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
		} `json:"items"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	if len(result.Items) == 0 {
		return ok("No repositories found matching your query")
}

	var response strings.Builder
	response.WriteString("GitHub Search Results:\n\n")

	for i, repo := range result.Items {
		response.WriteString(fmt.Sprintf("%d. %s\n", i+1, repo.Name))
		response.WriteString(fmt.Sprintf("   Description: %s\n", repo.Description))
		response.WriteString(fmt.Sprintf("   URL: %s\n\n", repo.HTMLURL))

	return ok(response.String())
}

}

func HandleOverturePlan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	planID, _ :=getString(args, "plan_id")
	if planID == "" {
		return err("plan_id parameter is required")
}

	// In a real implementation, this would interact with the Overture API
	// For this example, we'll simulate a response
	response := fmt.Sprintf("Plan %s details would be shown here in a real implementation", planID)
	return ok(response)
}

func HandleOvertureBranch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	branchName, _ :=getString(args, "branch_name")
	if branchName == "" {
		return err("branch_name parameter is required")
}

	// Simulate branch creation
	response := fmt.Sprintf("Branch %s would be created in a real implementation", branchName)
	return ok(response)
}

func HandleOvertureNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeID, _ :=getString(args, "node_id")
	if nodeID == "" {
		return err("node_id parameter is required")
}

	// Simulate node operation
	response := fmt.Sprintf("Operation on node %s would be performed in a real implementation", nodeID)
	return ok(response)
}

func HandleOvertureStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Simulate status check
	response := "System status: Operational\n" +
		"Current plans: 5\n" +
		"Active branches: 3\n" +
		"Nodes processed: 42"

	return ok(response)
}

func HandleOvertureHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Simulate help response
	response := "Overture MCP Tool Help:\n\n" +
		"Available commands:\n" +
		"1. search - Search GitHub repositories\n" +
		"2. plan - Get plan details\n" +
		"3. branch - Create a new branch\n" +
		"4. node - Perform node operations\n" +
		"5. status - Check system status\n" +
		"6. help - Show this help message"

	return ok(response)
}