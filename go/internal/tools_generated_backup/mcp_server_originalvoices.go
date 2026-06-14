package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGenerateVoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	voice, _ :=getString(args, "voice")
	if text == "" {
		return err("text is required")
}

	apiURL := fmt.Sprintf("https://api.originalvoices.com/generate?text=%s&voice=%s", url.QueryEscape(text), url.QueryEscape(voice))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Generated voice: %v", result))
}