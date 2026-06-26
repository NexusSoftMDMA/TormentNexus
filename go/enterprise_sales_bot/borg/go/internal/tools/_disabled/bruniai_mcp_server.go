package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCompareImages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url1, _ :=getString(args, "url1")
	url2, _ :=getString(args, "url2")
	if url1 == "" || url2 == "" {
		return err("url1 and url2 are required")
}

	payload := map[string]string{"url1": url1, "url2": url2}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.bruniai.com/v1/compare", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(string(respBody))
}

func HandleAnalyzeImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("url is required")
}

	payload := map[string]string{"url": imageURL}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.bruniai.com/v1/analyze", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(string(respBody))
}