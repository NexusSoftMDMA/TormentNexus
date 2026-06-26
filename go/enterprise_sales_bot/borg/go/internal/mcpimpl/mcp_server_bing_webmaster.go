package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSubmitUrl_mcp_server_bing_webmaster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteUrl, _ :=getString(args, "siteUrl")
	urlToSubmit, _ :=getString(args, "url")
	apiKey, _ :=getString(args, "apiKey")
	if siteUrl == "" || urlToSubmit == "" || apiKey == "" {
		return err("Missing required parameters: siteUrl, url, apiKey")
}

	body := map[string]string{"siteUrl": siteUrl, "url": urlToSubmit}
	bodyBytes, e := json.Marshal(body)
	if e != nil {
		return err("Failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://ssl.bing.com/webmaster/api.svc/json/SubmitUrl?apikey=%s", apiKey), bytes.NewReader(bodyBytes))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return success("URL submitted successfully")
}

func HandleGetCrawlInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	siteUrl, _ :=getString(args, "siteUrl")
	apiKey, _ :=getString(args, "apiKey")
	if siteUrl == "" || apiKey == "" {
		return err("Missing required parameters: siteUrl, apiKey")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://ssl.bing.com/webmaster/api.svc/json/GetCrawlInfo?apikey=%s&siteUrl=%s", apiKey, siteUrl), nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(respBody))
}