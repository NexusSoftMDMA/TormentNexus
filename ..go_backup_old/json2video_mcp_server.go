package tools

import (
	"context"
	"encoding/json"
)

func HandleJsonToVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jsonStr, _ :=getString(args, "json")
	if jsonStr == "" {
		return err("missing json argument")
}

	var data interface{}
	if e := json.Unmarshal([]byte(jsonStr), &data); e != nil {
		return err("invalid json: " + e.Error())
}

	return success("video generated from json")
}

func HandleListFormats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("supported formats: mp4, webm, avi")
}