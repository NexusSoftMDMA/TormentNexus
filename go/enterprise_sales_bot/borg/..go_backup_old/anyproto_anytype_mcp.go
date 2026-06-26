package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateObject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spaceID, _ :=getString(args, "spaceId")
	objectType, _ :=getString(args, "objectType")
	name, _ :=getString(args, "name")
	body, _ := json.Marshal(map[string]interface{}{
		"spaceId": spaceID,
		"objectType": objectType,
		"name": name,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.anytype.io/v1/objects", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("object created")
}

func HandleSearchObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	body, _ := json.Marshal(map[string]interface{}{
		"query": query,
		"limit": limit,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.anytype.io/v1/search", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("search completed")
}