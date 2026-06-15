package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	u := fmt.Sprintf("https://registry.opentofu.org/v1/providers?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to query providers: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Providers found: %v", result))
}

func HandleGetProvider(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	typ, _ :=getString(args, "type")
	u := fmt.Sprintf("https://registry.opentofu.org/v1/providers/%s/%s", namespace, typ)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to get provider: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Provider details: %v", result))
}