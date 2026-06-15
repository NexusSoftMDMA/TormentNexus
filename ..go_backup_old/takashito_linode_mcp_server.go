package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListLinodes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("LINODE_TOKEN")
	if token == "" {
		return err("LINODE_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.linode.com/v4/linode/instances", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Linodes: %+v", result))
}

func HandleGetLinode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("LINODE_TOKEN")
	if token == "" {
		return err("LINODE_TOKEN not set")
}

	id, _ :=getString(args, "linode_id")
	if id == "" {
		return err("linode_id is required")
}

	url := fmt.Sprintf("https://api.linode.com/v4/linode/instances/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Linode: %+v", result))
}