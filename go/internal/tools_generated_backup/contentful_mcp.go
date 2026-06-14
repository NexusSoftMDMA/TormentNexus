package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleGetEntries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spaceID, _ :=getString(args, "space_id")
	accessToken, _ :=getString(args, "access_token")
	contentType, _ :=getString(args, "content_type")
	if spaceID == "" || accessToken == "" {
		return err("space_id and access_token are required")
}

	u := fmt.Sprintf("https://cdn.contentful.com/spaces/%s/entries?access_token=%s", url.PathEscape(spaceID), url.QueryEscape(accessToken))
	if contentType != "" {
		u += "&content_type=" + url.QueryEscape(contentType)

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return ok(string(body))
}

}

func HandleGetEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spaceID, _ :=getString(args, "space_id")
	accessToken, _ :=getString(args, "access_token")
	entryID, _ :=getString(args, "entry_id")
	if spaceID == "" || accessToken == "" || entryID == "" {
		return err("space_id, access_token, and entry_id are required")
}

	u := fmt.Sprintf("https://cdn.contentful.com/spaces/%s/entries/%s?access_token=%s",
		url.PathEscape(spaceID), url.PathEscape(entryID), url.QueryEscape(accessToken))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return ok(string(body))
}