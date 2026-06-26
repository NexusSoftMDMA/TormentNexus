package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch_npmplus_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s&size=10", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(result)
}

func HandleGetPackageInfo_npmplus_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "package")
	if packageName == "" {
		return err("package is required")
}

	url := "https://registry.npmjs.org/" + packageName
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return err("package not found")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(result)
}