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

func HandleSearchIntents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	apiURL := os.Getenv("MARKIDY_API_URL")
	if apiURL == "" {
		apiURL = "https://api.markidy.com"
	}
	url := fmt.Sprintf("%s/intents?q=%s&limit=%d", apiURL, query, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status + " - " + string(body))
}

	return ok(string(body))
}