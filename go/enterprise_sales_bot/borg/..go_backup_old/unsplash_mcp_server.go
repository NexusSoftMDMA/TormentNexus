package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearchUnsplash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	perPage, _ :=getInt(args, "perPage")
	if perPage == 0 {
		perPage = 10
	}
	apiKey := os.Getenv("UNSPLASH_ACCESS_KEY")
	url := fmt.Sprintf("https://api.unsplash.com/search/photos?query=%s&per_page=%d", query, perPage)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Client-ID "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleGetUnsplashPhoto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	apiKey := os.Getenv("UNSPLASH_ACCESS_KEY")
	url := fmt.Sprintf("https://api.unsplash.com/photos/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Client-ID "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(string(body))
}