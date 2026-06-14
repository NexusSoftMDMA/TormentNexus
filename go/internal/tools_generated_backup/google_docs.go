package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	accessToken, _ :=getString(args, "access_token")
	body := map[string]interface{}{"title": title}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://docs.googleapis.com/v1/documents", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
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
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("Document created: %v", result["documentId"]))
}

func HandleGetDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "documentId")
	accessToken, _ :=getString(args, "access_token")
	url := fmt.Sprintf("https://docs.googleapis.com/v1/documents/%s", docID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
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
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("Document: %v", result["title"]))
}