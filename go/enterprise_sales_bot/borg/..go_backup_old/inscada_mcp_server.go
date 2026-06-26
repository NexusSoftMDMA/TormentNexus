package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleReadTag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tag, _ :=getString(args, "tag")
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	fullURL := fmt.Sprintf("%s/api/read?tag=%s", url, tag)
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err("failed to read tag: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	value, found := result["value"]
	if !found {
		return err("no value in response")
}

	return success(fmt.Sprintf("Tag '%s' value: %v", tag, value))
}

func HandleWriteTag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tag, _ :=getString(args, "tag")
	value, _ :=getString(args, "value")
	url, _ :=getString(args, "url")
	if url == "" || tag == "" || value == "" {
		return err("url, tag, and value are required")
}

	fullURL := fmt.Sprintf("%s/api/write?tag=%s&value=%s", url, tag, value)
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err("failed to write tag: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	status, found := result["status"]
	if !found {
		return err("no status in response")
}

	return success(fmt.Sprintf("Write to tag '%s' status: %v", tag, status))
}