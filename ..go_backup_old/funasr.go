package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleAsr(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	audio, _ :=getString(args, "audio")
	if audio == "" {
		return err("audio input is required")
}

	result := map[string]string{"transcription": "Hello, world!"}
	data, e := json.Marshal(result)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", e))
}

	return ok(string(data))
}