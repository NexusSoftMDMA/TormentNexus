package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSerperSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey := os.Getenv("SERPER_API_KEY")
	if apiKey == "" {
		return err("SERPER_API_KEY environment variable not set")
}

	body, _ := json.Marshal(map[string]string{"q": query})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://google.serper.dev/search", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	organic, found := result["organic"].([]interface{})
	if !found {
		return err("no organic results")
}

	var output string
	for _, item := range organic {
		entry, _ := item.(map[string]interface{})
		output += fmt.Sprintf("Title: %v\nLink: %v\nSnippet: %v\n\n", entry["title"], entry["link"], entry["snippet"])

	return ok(output)
}
}