package mcpimpl

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "io"
    "net/http"
    "strings"
)

func HandleTTS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    text, _ :=getString(args, "text")
    if text == "" {
        return err("text is required")
}

    apiKey, _ :=getString(args, "api_key")
    voiceID, _ :=getString(args, "voice_id")
    body := map[string]interface{}{
        "text": text,
    }
    if voiceID != "" {
        body["voice_id"] = voiceID
    }
    jsonBody, e := json.Marshal(body)
    if e != nil {
        return err("failed to marshal request: " + e.Error())
}

    req, e := http.NewRequestWithContext(ctx, "POST", "https://api.fish.audio/v1/tts", strings.NewReader(string(jsonBody)))
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("request failed: " + e.Error())
}

    defer resp.Body.Close()
    audioBytes, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    if resp.StatusCode != 200 {
        return err("API error: " + string(audioBytes))
}

    base64Audio := base64.StdEncoding.EncodeToString(audioBytes)
    return ok(base64Audio)
}