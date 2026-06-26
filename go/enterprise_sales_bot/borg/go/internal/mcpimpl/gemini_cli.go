package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleGenerateText_gemini_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	apiKey, _ :=getString(args, "apiKey")
	if prompt == "" || apiKey == "" {
		return err("prompt and apiKey are required")
}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + apiKey

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to call Gemini API")
}

	defer resp.Body.Close()

	respData, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if e := json.Unmarshal(respData, &result); e != nil {
		return err("failed to parse response")
}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return err("no response from Gemini")
}

	text := result.Candidates[0].Content.Parts[0].Text
	return success(text)
}