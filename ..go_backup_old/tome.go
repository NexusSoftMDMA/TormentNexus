package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleAskTome(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	base := os.Getenv("TOME_API_URL")
	if base == "" {
		base = "https://tome.app/api/ask"
	}
	body := fmt.Sprintf(`{"query":"%s"}`, query)
	req, e := http.NewRequestWithContext(ctx, "POST", base, strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Tome response: %v", result))
}