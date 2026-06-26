package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleBuiltWithTechLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	apikey, _ :=getString(args, "apikey")
	if domain == "" || apikey == "" {
		return err("domain and apikey are required")
}

	url := fmt.Sprintf("https://api.builtwith.com/free1/api.json?KEY=%s&LOOKUP=%s", apikey, domain)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("BuiltWith tech lookup result: %+v", result))
}