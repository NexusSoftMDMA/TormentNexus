package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	url := fmt.Sprintf("https://api.lore.example.com/lore?topic=%s", topic)
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	lore, found := result["lore"].(string)
	if !found {
		return err("unexpected response format")
}

	return ok(lore)
}

func HandleListLore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topics := []string{"dragons", "elves", "ancient ruins", "magic"}
	data, e := json.Marshal(topics)
	if e != nil {
		return err("failed to marshal topics")
}

	return ok(string(data))
}