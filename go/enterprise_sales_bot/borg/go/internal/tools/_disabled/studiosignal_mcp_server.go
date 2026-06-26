package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleStudioSignalQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	urlStr := fmt.Sprintf("https://api.studiosignal.com/intelligence?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err("failed to make request")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	msg := fmt.Sprintf("Result: %v", result)
	return ok(msg)
}

func HandleStudioSignalPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}// touch 1781132141
