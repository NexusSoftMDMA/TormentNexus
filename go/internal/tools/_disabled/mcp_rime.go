package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListVoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.rime.ai/v1/voices")
	if e != nil {
		return err("failed to list voices: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Voices: %v", result))
}

func HandleTextToSpeech(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	voice, _ :=getString(args, "voice")
	if voice == "" {
		voice = "default"
	}
	payload := map[string]interface{}{
		"text":  text,
		"voice": voice,
	}
	bodyBytes, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://api.rime.ai/v1/synthesize", "application/json", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("synthesis failed: " + e.Error())
}

	defer resp.Body.Close()
	audioBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read audio: " + e.Error())
}

	return success(string(audioBytes))
}