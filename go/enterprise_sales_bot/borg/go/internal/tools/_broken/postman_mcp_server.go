package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleListCollections_postman_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("POSTMAN_API_KEY")

	if apiKey == "" {
		return err("missing Postman API key")
}

	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "https://api.getpostman.com"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/collections", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
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

	return ok(fmt.Sprintf("Collections: %v", result["collections"]))
}

}

func HandleGetCollection_postman_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("POSTMAN_API_KEY")

	if apiKey == "" {
		return err("missing Postman API key")
}

	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "https://api.getpostman.com"
	}
	uid, _ :=getString(args, "uid")
	if uid == "" {
		return err("missing collection uid")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/collections/"+url.PathEscape(uid), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
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

	return ok(fmt.Sprintf("Collection: %v", result["collection"]))
}
}