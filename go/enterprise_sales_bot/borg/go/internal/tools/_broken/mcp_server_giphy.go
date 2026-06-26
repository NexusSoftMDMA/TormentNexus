package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleSearchGiphy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	apiKey := os.Getenv("GIPHY_API_KEY")
	if apiKey == "" {
		return err("GIPHY_API_KEY not set")
}

	params := url.Values{}
	params.Set("api_key", apiKey)
	params.Set("q", q)
	resp, e := http.DefaultClient.Get("https://api.giphy.com/v1/gifs/search?" + params.Encode())
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	if len(result.Data) == 0 {
		return ok("No gifs found")
}

	out := "Gif URLs:\n"
	for _, g := range result.Data {
		out += fmt.Sprintf("- %s\n", g.URL)

	return ok(out)
}
}