package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleCheckShield(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	resp, e := http.DefaultClient.Get("https://api.shieldapi.io/v1/check?domain=" + domain)
	if e != nil {
		return err("failed to call shield api")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}