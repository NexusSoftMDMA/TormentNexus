package mcpimpl

import (
	"context"
	"encoding/json"
	"net/url"
)

func HandleValidateJson(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	s, _ :=getString(args, "json_string")
	if s == "" {
		return err("json_string is required")
}

	if json.Valid([]byte(s)) {
		return ok("valid JSON")
}

	return err("invalid JSON")
}

func HandleValidateUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	u, _ :=getString(args, "url")
	if u == "" {
		return err("url is required")
}

	_, e := url.ParseRequestURI(u)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	return ok("valid URL")
}