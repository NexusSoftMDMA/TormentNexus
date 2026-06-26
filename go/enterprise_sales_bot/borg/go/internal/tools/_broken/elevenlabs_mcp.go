package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetVoices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return err("ELEVENLABS_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/voices", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("xi-api-key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	voices, found := result["voices"].([]interface{})
	if !found {
		return err("no voices found")
}

	var list string
	for _, v := range voices {
		vm, _ := v.(map[string]interface{})
		name, _ := vm["name"].(string)
		vid, _ := vm["voice_id"].(string)
		list += fmt.Sprintf("%s (%s)\n", name, vid)

	return ok(list)
}

}

func HandleTextToSpeech_elevenlabs_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return err("ELEVENLABS_API_KEY not set")
}

	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	voiceID, _ :=getString(args, "voice_id")
	if voiceID == "" {
		voiceID = "21m00Tcm4TlvDq8ikWAM" // default Rachel
	}
	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return success("Audio generated successfully")
}