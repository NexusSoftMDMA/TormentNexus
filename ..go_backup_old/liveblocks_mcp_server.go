package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleGetRooms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	secretKey := os.Getenv("LIVEBLOCKS_SECRET_KEY")
	if secretKey == "" {
		return err("LIVEBLOCKS_SECRET_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.liveblocks.io/v2/rooms", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %s", e.Error()))
}

	req.Header.Set("Authorization", "Bearer "+secretKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %s", e.Error()))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %s", e.Error()))
}

	data, found := result["data"]
	if !found {
		return err("no data field in response")
}

	return ok(fmt.Sprintf("rooms: %v", data))
}