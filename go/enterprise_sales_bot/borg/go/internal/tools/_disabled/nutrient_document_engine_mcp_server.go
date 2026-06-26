package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleConvertDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileURL, _ :=getString(args, "file_url")
	format, _ :=getString(args, "format")
	apiURL, _ :=getString(args, "api_url")
	if fileURL == "" {
		return err("file_url is required")
}

	if apiURL == "" {
		return err("api_url is required")
}

	payload, e := json.Marshal(map[string]string{"file_url": fileURL, "format": format})
	if e != nil {
		return err("failed to marshal payload")
}

	resp, e := http.Post(apiURL, "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("http request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("conversion failed")
}

	return success("document converted")
}

func HandleExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileURL, _ :=getString(args, "file_url")
	apiURL, _ :=getString(args, "api_url")
	if fileURL == "" {
		return err("file_url is required")
}

	if apiURL == "" {
		return err("api_url is required")
}

	payload, e := json.Marshal(map[string]string{"file_url": fileURL})
	if e != nil {
		return err("failed to marshal payload")
}

	resp, e := http.Post(apiURL, "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("http request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("extraction failed")
}

	return success("text extracted")
}