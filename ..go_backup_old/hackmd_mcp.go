package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiToken")
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	body := map[string]string{"title": title, "content": content}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.hackmd.io/v1/notes", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("HackMD API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}

func HandleReadNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiToken")
	noteID, _ :=getString(args, "noteId")
	url := fmt.Sprintf("https://api.hackmd.io/v1/notes/%s", noteID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HackMD API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
			return err("failed to parse response: " + e.Error())
}

	return ok(strings.TrimSpace(string(respBody)))
}