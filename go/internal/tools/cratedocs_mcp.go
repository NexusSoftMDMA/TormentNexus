//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSearchCrates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://crates.io/api/v1/crates?q=" + query)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Crates []struct {
			Name string `json:"name"`,
		} `json:"crates"`,
		if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return e,
}


-reasoner (deepseek)*
}
}