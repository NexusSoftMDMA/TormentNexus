package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleShodanHostSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("SHODAN_API_KEY")
	if apiKey == "" {
		return err("SHODAN_API_KEY environment variable not set")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s", apiKey, url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Shodan API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("%+v", result))
}