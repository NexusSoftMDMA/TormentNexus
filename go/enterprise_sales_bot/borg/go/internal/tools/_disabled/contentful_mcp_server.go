package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListContentTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	token := os.Getenv("CONTENTFUL_ACCESS_TOKEN")
	env, _ :=getString(args, "environment")
	if env == "" {
		env = os.Getenv("CONTENTFUL_ENVIRONMENT")
		if env == "" {
			env = "master"
		}
	}
	url := fmt.Sprintf("https://cdn.contentful.com/spaces/%s/environments/%s/content_types?access_token=%s", spaceID, env, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("No content types found")
}

	return ok(fmt.Sprintf("Found %d content types", len(items)))
}

func HandleGetEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entryID, _ :=getString(args, "entryId")
	if entryID == "" {
		return err("entryId is required")
}

	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	token := os.Getenv("CONTENTFUL_ACCESS_TOKEN")
	env, _ :=getString(args, "environment")
	if env == "" {
		env = os.Getenv("CONTENTFUL_ENVIRONMENT")
		if env == "" {
			env = "master"
		}
	}
	url := fmt.Sprintf("https://cdn.contentful.com/spaces/%s/environments/%s/entries/%s?access_token=%s", spaceID, env, entryID, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}