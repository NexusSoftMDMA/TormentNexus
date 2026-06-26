package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleLinearListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "teamId")
	state, _ :=getString(args, "state")
	query := `query { issues`
	if teamID != "" {
		query += fmt.Sprintf(`(filter: {team: {id: {eq: "%s"}}}`, teamID)

	if state != "" {
		if teamID == "" {
			query += fmt.Sprintf(`(filter: {state: {name: {eq: "%s"}}}`, state)
		} else {
			query += fmt.Sprintf(`, state: {name: {eq: "%s"}}`, state)

	}
	query += `) { nodes { id title state { name } team { name } } } }`

	apiURL := os.Getenv("LINEAR_API_URL")
	if apiURL == "" {
		apiURL = "https://api.linear.app/graphql"
	}
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		return err("LINEAR_API_KEY not set")
}

	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	issues, found := data["issues"].(map[string]interface{})
	if !found {
		return err("no issues in response")
}

	nodes, found := issues["nodes"].([]interface{})
	if !found {
		return err("no nodes in issues")
}

	out, e := json.MarshalIndent(nodes, "", "  ")
	if e != nil {
		return err("marshal output failed: " + e.Error())
}

	return ok(string(out))
}

}
}

func HandleLinearGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueID, _ :=getString(args, "issueId")
	if issueID == "" {
		return err("issueId is required")
}

	query := fmt.Sprintf(`query { issue(id: "%s") { id title description state { name } team { name } } }`, issueID)

	apiURL := os.Getenv("LINEAR_API_URL")
	if apiURL == "" {
		apiURL = "https://api.linear.app/graphql"
	}
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		return err("LINEAR_API_KEY not set")
}

	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("no data in response")
}

	issue, found := data["issue"].(map[string]interface{})
	if !found {
		return err("no issue in response")
}

	out, e := json.MarshalIndent(issue, "", "  ")
	if e != nil {
		return err("marshal output failed: " + e.Error())
}

	return ok(string(out))
}