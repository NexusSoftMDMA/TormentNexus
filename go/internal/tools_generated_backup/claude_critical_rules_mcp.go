package tools

import (
	"context"
	"net/http"
)

func HandleGetRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := "https://critical-rules.example.com/rule/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("rule found")
}

func HandleListRules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("list of critical rules")
}