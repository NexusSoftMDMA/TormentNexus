package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	protocol, _ :=getString(args, "protocol")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	u, e := url.Parse("https://api.402index.com/search")
	if e != nil {
		return err("invalid URL")
}

	q := u.Query()
	q.Set("q", query)
	if protocol != "" {
		q.Set("protocol", protocol)

	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return success(string(data))
}
}