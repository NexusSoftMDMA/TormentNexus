package mcpimpl

import (
	"context"
	"fmt"
)

func HandleListVoices_voice_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	voices := []string{"en-US-Wavenet-D", "en-US-Wavenet-J", "en-GB-Wavenet-A"}
	return ok(fmt.Sprintf("Available voices: %v", voices))
}

func HandleSpeak_voice_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	voice, _ :=getString(args, "voice")
	if voice == "" {
		voice = "en-US-Wavenet-D"
	}
	return success(fmt.Sprintf("Speaking '%s' with voice %s", text, voice))
}