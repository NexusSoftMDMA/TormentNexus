package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	goal, _ :=getString(args, "goal")
	apiKey := os.Getenv("SKYVERN_API_KEY")
	if apiKey == "" {
		return err("SKYVERN_API_KEY not set")
}

	payload := map[string]string{"url": url, "goal": goal}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.skyvern.com/api/v1/tasks", bytes.NewReader(body))
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(string(respBody))
}