package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleHasdataSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.hasdata.com/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("API call failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}