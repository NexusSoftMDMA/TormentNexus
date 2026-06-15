package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMemoryBank(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bankName, _ :=getString(args, "bankName")
	url := fmt.Sprintf("https://api.example.com/memory/%s", bankName)
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

	var result map[string]string
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	content, found := result["content"]
	if !found {
		return err("no content field in response")
}

	return success(content)
}

func HandleSetMemoryBank(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bankName, _ :=getString(args, "bankName")
	content, _ :=getString(args, "content")
	payload, _ := json.Marshal(map[string]string{"content": content})
	url := fmt.Sprintf("https://api.example.com/memory/%s", bankName)
	req, e := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Body = io.NopCloser(bytes.NewReader(payload))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("memory bank updated")
}