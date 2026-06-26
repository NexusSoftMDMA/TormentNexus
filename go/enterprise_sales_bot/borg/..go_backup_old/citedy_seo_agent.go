package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnalyzeSeo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get("https://example.com/seo?url=" + url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	return ok(string(data))
}