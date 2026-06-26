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

func HandleUploadkitListUploads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("UPLOADKIT_BASE_URL")
	if base == "" {
		base = "https://api.uploadkit.com"
	}
	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("UPLOADKIT_TOKEN")

	reqURL, _ := url.JoinPath(base, "/v1/files")
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	return success(fmt.Sprintf("%v", result))
}

}

func HandleUploadkitGetUploadURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("UPLOADKIT_BASE_URL")
	if base == "" {
		base = "https://api.uploadkit.com"
	}
	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("UPLOADKIT_TOKEN")

	filename, _ :=getString(args, "filename")
	reqURL, _ := url.JoinPath(base, "/v1/upload")
	req, e := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Set("filename", filename)
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	return success(fmt.Sprintf("%v", result))
}
}