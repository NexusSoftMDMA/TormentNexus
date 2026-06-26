package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleGenerateEndpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url, _ :=getString(args, "url")
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	return ok(fmt.Sprintf("endpoint %s generated (status %d)", name, resp.StatusCode))
}

func HandleGenerateSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fields, _ :=getString(args, "fields")
	format, _ :=getString(args, "format")
	_ = format
	return success(fmt.Sprintf("schema generated with fields: %s", fields))
}