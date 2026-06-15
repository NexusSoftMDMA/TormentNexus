package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleKeywordResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	url := fmt.Sprintf("https://api.jackpotkeywords.com/keywords?q=%s", keyword)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Keyword research result: %v", result))
}

func HandleVisibilityScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	url := fmt.Sprintf("https://api.jackpotkeywords.com/visibility?domain=%s", domain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("AI-visibility scan result: %v", result))
}