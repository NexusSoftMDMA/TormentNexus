package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSerpapiSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "q")
	if query == "" {
		return err("'q' argument is required")
}

	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		return err("SERPAPI_API_KEY environment variable not set")
}

	url := fmt.Sprintf("https://serpapi.com/search?q=%s&api_key=%s", query, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	out, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	return success(string(out))
}