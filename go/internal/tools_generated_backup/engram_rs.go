package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	base := os.Getenv("ENGRAM_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	url := base + "/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return success(string(body))
}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	base := os.Getenv("ENGRAM_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	payload, _ := json.Marshal(map[string]string{"title": title, "content": content})
	resp, e := http.DefaultClient.Post(base+"/notes", "application/json", payload)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return success(string(body))
}