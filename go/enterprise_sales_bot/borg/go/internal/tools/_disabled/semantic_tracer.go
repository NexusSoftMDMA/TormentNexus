package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSemanticTracer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing required field: text")
}

	lang, _ :=getString(args, "lang")

	u := fmt.Sprintf("https://api.semantictracer.com/trace?text=%s", url.QueryEscape(text))
	if lang != "" {
		u += "&lang=" + url.QueryEscape(lang)

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if !result.Success {
		return err("API returned unsuccessful")
}

	return ok("trace result: " + result.Data)
}
}