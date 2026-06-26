package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetNpmPackage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "package")
	if packageName == "" {
		return err("package name is required")
}

	version, _ :=getString(args, "version")
	url := fmt.Sprintf("https://registry.npmjs.org/%s", packageName)
	if version != "" {
		url = fmt.Sprintf("https://registry.npmjs.org/%s/%s", packageName, version)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch package info: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	jsonData, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("failed to format JSON: %v", e))
}

	return success(string(jsonData))
}
}