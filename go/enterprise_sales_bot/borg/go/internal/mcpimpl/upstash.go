package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListDatabases_upstash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key := os.Getenv("UPSTASH_API_KEY")
	if key == "" {
		return err("UPSTASH_API_KEY not set")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.upstash.com/v2/redis/databases", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}
	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
	}
	return ok(fmt.Sprintf("Databases: %s", string(body)))
}

func HandleGetDatabase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
	}
	key := os.Getenv("UPSTASH_API_KEY")
	if key == "" {
		return err("UPSTASH_API_KEY not set")
	}
	url := "https://api.upstash.com/v2/redis/databases/" + id
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}
	return ok(fmt.Sprintf("Database: %s", string(body)))
}