//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {

	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rest/api/content/search?cql=text~'"+query+"'&limit="+fmt.Sprint(limit), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "api_token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	results, found := result["results"].([]interface{})
	if !found {
		return err("no results field in response")
}

	output := ""
	for _, item := range results {
		page, found := item.(map[string]interface{})
		if found {
			title, _ :=


-reasoner (deepseek)*,
},
},
}
}