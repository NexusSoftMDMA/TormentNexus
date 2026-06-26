package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandlePurgePullzone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zoneID, _ :=getInt(args, "zone_id")
	if zoneID == 0 {
		return err("zone_id is required")
}

	pattern, _ :=getString(args, "pattern")
	apiKey := os.Getenv("BUNNY_API_KEY")
	if apiKey == "" {
		return err("BUNNY_API_KEY environment variable not set")
}

	url := fmt.Sprintf("https://api.bunny.net/pullzone/%d/purgeCache", zoneID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("AccessKey", apiKey)
	if pattern != "" {
		q := req.URL.Query()
		q.Set("pattern", pattern)
		req.URL.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("purge failed (status %d): %s", resp.StatusCode, string(body)))
}

	return ok("Cache purged successfully for pull zone " + fmt.Sprint(zoneID))
}

}

func HandleListStorageZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BUNNY_API_KEY")
	if apiKey == "" {
		return err("BUNNY_API_KEY environment variable not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.bunny.net/storagezone", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("AccessKey", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("list failed (status %d): %s", resp.StatusCode, string(body)))
}

	var zones []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&zones); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d storage zones", len(zones)))
}