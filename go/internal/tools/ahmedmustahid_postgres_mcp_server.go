//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("json error: " + e.Error())
}

	return success("query executed")
}
