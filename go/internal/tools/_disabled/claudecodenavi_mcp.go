package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(fmt.Sprintf("response status %d, body length %d", resp.StatusCode, len(body)))
}

	return ok(fmt.Sprintf("response status %d, parsed JSON keys: %v", resp.StatusCode, keys(result)))
}

func keys(m map[string]interface{}) []string {
	var k []string
	for key := range m {
		k = append(k, key)

	return k
}
}