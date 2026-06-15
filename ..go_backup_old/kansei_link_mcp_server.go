package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleSearchTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.kansei-link.mcp/tools?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to search tools: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var tools []map[string]interface{}
	if e := json.Unmarshal(body, &tools); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(tools) == 0 {
		return success("No tools found")
}

	var result strings.Builder
	for _, t := range tools {
		name, found := t["name"].(string)
		if !found {
			name = "unknown"
		}
		desc, found := t["description"].(string)
		if !found {
			desc = ""
		}
		result.WriteString(fmt.Sprintf("- %s: %s\n", name, desc))

	return ok(result.String())
}
}