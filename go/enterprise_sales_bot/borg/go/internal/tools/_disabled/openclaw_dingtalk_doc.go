package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func handleCreateDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	name, _ :=getString(args, "doc_name")
	content, _ :=getString(args, "content")
	if token == "" || name == "" {
		return err("missing required parameters")
}

	body, _ := json.Marshal(map[string]string{"name": name, "content": content})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.dingtalk.com/v1.0/doc/spaces/.../documents", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	return success("document created successfully")
}

func handleGetDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	docID, _ :=getString(args, "doc_id")
	if token == "" || docID == "" {
		return err("missing required parameters")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.dingtalk.com/v1.0/doc/spaces/.../documents/%s", docID), nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("x-acs-dingtalk-access-token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e = json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response")
}

	title, found := result["title"].(string)
	if !found {
		title = "untitled"
	}
	return ok(fmt.Sprintf("Document title: %s", title))
}