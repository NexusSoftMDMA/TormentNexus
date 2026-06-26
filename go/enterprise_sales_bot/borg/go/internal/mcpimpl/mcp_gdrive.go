package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGdriveList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	pageSize, _ :=getInt(args, "pageSize")
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return err("missing GOOGLE_API_KEY environment variable")
}

	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?key=%s&pageSize=%d", apiKey, pageSize)
	if query != "" {
		url += "&q=" + query
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	files, found := result["files"].([]interface{})
	if !found {
		return ok("No files found")
}

	out, _ := json.MarshalIndent(files, "", "  ")
	return success(string(out))
}