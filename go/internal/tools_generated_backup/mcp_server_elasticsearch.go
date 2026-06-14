package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	index, _ :=getString(args, "index")
	query, _ :=getString(args, "query")
	size, _ :=getInt(args, "size")
	if size <= 0 {
		size = 10
	}
	body := []byte(query)
	if len(body) == 0 {
		body = []byte(`{"query":{"match_all":{}}}`)

	req, e := http.NewRequestWithContext(ctx, "POST", url+"/"+index+"/_search", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("elasticsearch error: " + string(respBody))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	hits, found := result["hits"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
}

	return ok(fmt.Sprintf("found %v hits", hits["total"]))
}
}