package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServices_jaeger_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/api/services", base)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleFindTraces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	service, _ :=getString(args, "service")
	if base == "" || service == "" {
		return err("base_url and service are required")
}

	url := fmt.Sprintf("%s/api/traces?service=%s", base, service)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return ok(string(body))
}