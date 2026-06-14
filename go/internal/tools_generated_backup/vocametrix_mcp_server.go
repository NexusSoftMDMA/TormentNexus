package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleAnalyzeVoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	audioURL, _ :=getString(args, "audio_url")
	if audioURL == "" {
		return err("audio_url is required")
}

	lang, _ :=getString(args, "language")
	body, e := json.Marshal(map[string]string{
		"audio_url": audioURL,
		"language":  lang,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.vocametrix.com/analyze", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	return ok(fmt.Sprintf("Analysis result: %v", result))
}