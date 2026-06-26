package tools

import (
	"context"
)

func HandleGenerateAudio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	voice, _ :=getString(args, "voice")
	result := "Generated audio for: " + text + " with voice: " + voice
	return success(result)
}