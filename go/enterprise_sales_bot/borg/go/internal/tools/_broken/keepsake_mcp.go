package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetKeepsakes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	url := "https://api.keepsake.dev/keepsakes"
	if userID != "" {
		url += "?user_id=" + userID
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch keepsakes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Keepsakes: %s", string(body)))
}

func HandleCreateKeepsake(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name := strings.TrimSpace(getString(args, "name"))
	description := strings.TrimSpace(getString(args, "description"))
	if name == "" {
		return err("name is required")
}

	payload := fmt.Sprintf(`{"name":"%s","description":"%s"}`, name, description)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.keepsake.dev/keepsakes", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to create keepsake: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Keepsake created: %s", string(body)))
}