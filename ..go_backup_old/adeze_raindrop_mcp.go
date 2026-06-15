package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListBookmarks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collectionID, _ :=getInt(args, "collectionId")
	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.raindrop.io/rest/v1/raindrops/%d", collectionID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to parse response")
}

	items, found := result["items"].([]interface{})
	if !found {
		return err("no items in response")
}

	data, e := json.MarshalIndent(items, "", "  ")
	if e != nil {
		return err("failed to marshal items")
}

	return ok(string(data))
}