package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "https://registry.npmjs.org/-/v1/search?text=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Objects []struct {
			Package struct {
				Name string `json:"name"`
			} `json:"package"`
		} `json:"objects"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	if len(result.Objects) == 0 {
		return ok("no results found")
}

	names := make([]string, 0, len(result.Objects))
	for _, obj := range result.Objects {
		names = append(names, obj.Package.Name)

	out := fmt.Sprintf("Found %d packages: %v", len(names), names)
	return ok(out)
}
}