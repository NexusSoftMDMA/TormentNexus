package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleListCatalogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url is required")
}

	u, e := url.JoinPath(base, "/api/2.1/unity-catalog/catalogs")
	if e != nil {
		return err(fmt.Sprintf("invalid URL: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Catalogs: %+v", result["catalogs"]))
}

func HandleListSchemas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	catalog, _ :=getString(args, "catalog")
	if base == "" || catalog == "" {
		return err("base_url and catalog are required")
}

	u, e := url.JoinPath(base, "/api/2.1/unity-catalog/schemas", url.PathEscape(catalog))
	if e != nil {
		return err(fmt.Sprintf("invalid URL: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Schemas for catalog %s: %+v", catalog, result["schemas"]))
}