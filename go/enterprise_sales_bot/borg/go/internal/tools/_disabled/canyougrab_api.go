package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCheckDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	resp, e := http.DefaultClient.Get("https://api.canyougrab.com/check?domain=" + domain)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Available bool `json:"available"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	if result.Available {
		return ok(fmt.Sprintf("Domain '%s' is available", domain))
}

	return ok(fmt.Sprintf("Domain '%s' is not available", domain))
}