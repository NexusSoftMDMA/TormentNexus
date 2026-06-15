package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type npmSearchResult struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
		} `json:"package"`
	} `json:"objects"`
}

func HandleNpmSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s&size=5", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search npm: " + e.Error())
}

	defer resp.Body.Close()
	var result npmSearchResult
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Objects) == 0 {
		return ok("No packages found")
}

	var out string
	for i, obj := range result.Objects {
		pkg := obj.Package
		out += fmt.Sprintf("%d. %s v%s - %s\n", i+1, pkg.Name, pkg.Version, pkg.Description)

	return success(out)
}
}