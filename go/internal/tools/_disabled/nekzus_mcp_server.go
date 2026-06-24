package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPackage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "package")
	if name == "" {
		return err("package name is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://registry.npmjs.org/%s", name))
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("Package info: %s", string(body)))
}

func HandleSearchPackages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s", query))
	if e != nil {
		return err("failed to search")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}