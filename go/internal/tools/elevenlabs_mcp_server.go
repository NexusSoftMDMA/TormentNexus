//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleListVoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		return err("ELEVENLABS_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/voices", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("xi-api-key", key)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	return ok(string(body))
}

func HandleTextToSpeech(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		return err("ELEVENLABS_API_KEY not set")
}

	voiceID, _ :=getString(args, "voice_id")
	if voiceID == "" {
		return err("voice_id is required")
}

	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	payload := fmt.Sprintf(`{"text":"%s"}`, strings.ReplaceAll(text, `"`, `\"`))
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", key)
	req.Header.Set("Accept", "audio/mpeg")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	audio, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read audio failed: " + e.Error())
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(audio)))
}

	// Return base64? For simplicity, return hex or just "Audio received" with length.
	// But MCP expects content. We'll return a message.
	return ok(fmt.Sprintf("Audio received, length=%d bytes", len(audio)))
}
}
}
}