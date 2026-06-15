package tools

import (
	"context"
	"io"
	"net/http"
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
	resp, e := http.DefaultClient.Get("https://api.kokoro.com/tts?text=" + text + "&voice=" + voice)
	if e != nil {
		return err("TTS API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleTextToSpeech(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok("synthesized: " + text)
}