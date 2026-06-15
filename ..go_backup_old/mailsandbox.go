package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inboxID, _ :=getString(args, "inbox_id")
	if inboxID == "" {
		return err("inbox_id is required")
}

	apiKey := os.Getenv("MAILSANDBOX_API_KEY")
	if apiKey == "" {
		return err("MAILSANDBOX_API_KEY not set")
}

	url := fmt.Sprintf("https://api.mailsandbox.com/v1/inboxes/%s/messages", inboxID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var messages []map[string]interface{}
	if e := json.Unmarshal(body, &messages); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	out, _ := json.MarshalIndent(messages, "", "  ")
	return ok(string(out))
}