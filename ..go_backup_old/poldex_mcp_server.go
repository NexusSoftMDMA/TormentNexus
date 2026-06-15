package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleExtractInsuranceData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docURL, _ :=getString(args, "document_url")
	if docURL == "" {
		return err("document_url is required")
}

	endpoint := "https://api.poldex.com/extract?url=" + url.QueryEscape(docURL)
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
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

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response: " + e.Error())
}

	status, found := result["status"].(string)
	if !found || status != "ok" {
		return err("extraction failed")
}

	return success("extraction completed successfully")
}