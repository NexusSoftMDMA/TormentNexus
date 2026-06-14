package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleSearchCve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://cve.circl.lu/api/cve?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch CVE data: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	if len(result) == 0 {
		return ok("No CVEs found for query: " + query)
}

	summary := fmt.Sprintf("Found %d CVEs for query '%s':\n", len(result), query)
	for i, cve := range result {
		id, found := cve["id"].(string)
		if found {
			summary += fmt.Sprintf("%d. %s\n", i+1, id)

	}
	return ok(summary)
}
}