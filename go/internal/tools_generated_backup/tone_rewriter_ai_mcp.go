package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnalyzeTone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	body, _ := json.Marshal(map[string]string{"text": text})
	resp, e := http.DefaultClient.Post("https://api.meok.ai/tone-rewriter/analyze", "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	tone, found := result["tone"].(string)
	if !found {
		return err("invalid response")
}

	return ok(tone)
}

func HandleRewriteTone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	targetTone, _ :=getString(args, "target_tone")
	if text == "" || targetTone == "" {
		return err("text and target_tone are required")
}

	body, _ := json.Marshal(map[string]string{"text": text, "target_tone": targetTone})
	resp, e := http.DefaultClient.Post("https://api.meok.ai/tone-rewriter/rewrite", "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	rewritten, found := result["rewritten_text"].(string)
	if !found {
		return err("invalid response")
}

	return ok(rewritten)
}