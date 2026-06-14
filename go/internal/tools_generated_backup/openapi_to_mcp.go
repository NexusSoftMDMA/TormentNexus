package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleOpenapiToMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "openapi_spec")
	if spec == "" {
		return err("openapi_spec is required")
}

	resp, e := http.DefaultClient.Get(spec)
	if e != nil {
		return err("failed to fetch spec: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read spec: " + e.Error())
}

	return ok(string(body))
}