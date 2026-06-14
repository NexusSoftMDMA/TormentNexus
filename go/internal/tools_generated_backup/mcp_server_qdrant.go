package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/collections", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal: " + e.Error())
}

	return ok(fmt.Sprintf("collections: %v", result))
}

func HandleGetCollectionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	name, _ :=getString(args, "collection_name")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/collections/"+name, nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal: " + e.Error())
}

	return ok(fmt.Sprintf("collection info: %v", result))
}