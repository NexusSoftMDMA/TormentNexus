package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func HandleKnowledgeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 5
	}
	u := "https://api.quelvio.com/search?q=" + url.QueryEscape(query) + "&count=" + strconv.Itoa(count)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e == nil {
		bytes, _ := json.MarshalIndent(result, "", "  ")
		return success(string(bytes))
}

	return success(string(body))
}