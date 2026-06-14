package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	uuid, _ :=getString(args, "uuid")
	if apiKey == "" || uuid == "" {
		return err("apiKey and uuid are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.getbring.com/rest/v1/bringusers/%s/lists", uuid), nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-Bring-API-Key", apiKey)
	req.Header.Set("X-Bring-Client", "1")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(string(body))
}