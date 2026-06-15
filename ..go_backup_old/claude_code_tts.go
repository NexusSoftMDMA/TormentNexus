package tools

import (
    "context"
)

func HandleTts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    text, _ :=getString(args, "text")
    if text == "" {
        return err("text is required")
}

    voice, _ :=getString(args, "voice")
    if voice == "" {
        voice = "default"
    }
    return ok("TTS: text=%s, voice=%s", text, voice)
}