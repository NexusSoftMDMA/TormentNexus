package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGenerateCodeGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo_url")
	code, _ :=getString(args, "code")
	if repo == "" && code == "" {
		return err("either repo_url or code is required")
}

	payload := map[string]string{}
	if repo != "" {
		payload["repo_url"] = repo
	} else {
		payload["code"] = code
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.supermodel.ai/v1/codegraph", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(respBytes))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBytes, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Code graph generated: %v", result["graph_id"]))
}

func HandleGetGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	graphID, _ :=getString(args, "graph_id")
	if graphID == "" {
		return err("graph_id is required")
}

	url := "https://api.supermodel.ai/v1/codegraph/" + graphID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(respBytes))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBytes, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Graph data: %v", result))
}