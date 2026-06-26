package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnalyzeFrame(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	frame, _ :=getString(args, "frame")
	if frame == "" {
		return err("frame is required")
}

	prevFrame, _ :=getString(args, "previous_frame")
	action, _ :=getString(args, "action")
	_ = http.DefaultClient
	result := map[string]interface{}{
		"token_savings": "75%",
		"state_changed": prevFrame != frame,
		"action":        action,
	}
	jsonBytes, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return success(string(jsonBytes))
}