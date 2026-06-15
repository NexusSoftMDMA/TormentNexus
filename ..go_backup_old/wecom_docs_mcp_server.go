package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCreateDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	body := map[string]string{"title": title, "content": content}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.wecom.work/docs/create", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	e = json.Unmarshal(respBody, &result)
	if e != nil {
		return err("failed to parse response")
}

	return ok("Document created")
}

func HandleGetDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "doc_id")
	url := "https://api.wecom.work/docs/" + docID
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(respBody))
}