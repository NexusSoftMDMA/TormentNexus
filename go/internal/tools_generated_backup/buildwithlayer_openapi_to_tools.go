package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleConvertOpenApiToTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "spec_url")
	if url == "" {
		return err("spec_url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch spec: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var spec json.RawMessage
	if e := json.Unmarshal(body, &spec); e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	return success(fmt.Sprintf("OpenAPI spec loaded (%d bytes)", len(body)))
}