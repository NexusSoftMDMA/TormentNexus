package mcpimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleTinifyOptimize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	apiKey, _ :=getString(args, "api_key")
	if imageURL == "" || apiKey == "" {
		return err("image_url and api_key are required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.tinify.com/shrink", strings.NewReader(fmt.Sprintf(`{"source":{"url":"%s"}}`, imageURL)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	auth := base64.StdEncoding.EncodeToString([]byte("api:" + apiKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("tinify API error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Output struct {
			URL string `json:"url"`
		} `json:"output"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success("Optimized image URL: " + result.Output.URL)
}