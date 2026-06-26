package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	userID, _ :=getString(args, "userID")
	if apiKey == "" || userID == "" {
		return err("apiKey and userID required")
}

	url := "https://api.zotero.org/users/" + userID + "/items?q=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Zotero-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d items", len(data.([]interface{}))))
}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemKey, _ :=getString(args, "itemKey")
	apiKey, _ :=getString(args, "apiKey")
	userID, _ :=getString(args, "userID")
	if apiKey == "" || userID == "" || itemKey == "" {
		return err("apiKey, userID, itemKey required")
}

	url := "https://api.zotero.org/users/" + userID + "/items/" + itemKey
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Zotero-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}