package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFileMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileID, _ :=getString(args, "file_id")
	if fileID == "" {
		return err("file_id is required")
}

	url := fmt.Sprintf("https://api.rockhopper.co/files/%s/metadata", fileID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch metadata: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse metadata: " + e.Error())
}

	return ok("metadata retrieved")
}

func HandleGetFileReviews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileID, _ :=getString(args, "file_id")
	if fileID == "" {
		return err("file_id is required")
}

	url := fmt.Sprintf("https://api.rockhopper.co/files/%s/reviews", fileID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch reviews: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var reviews []map[string]interface{}
	if e := json.Unmarshal(body, &reviews); e != nil {
		return err("failed to parse reviews: " + e.Error())
}

	return ok("reviews retrieved")
}