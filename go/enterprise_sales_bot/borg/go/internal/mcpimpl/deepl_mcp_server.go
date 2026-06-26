package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleTranslate_deepl_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	targetLang, _ :=getString(args, "target_lang")
	sourceLang, _ :=getString(args, "source_lang")
	if text == "" || targetLang == "" {
		return err("text and target_lang are required")
}

	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		return err("DEEPL_API_KEY not set")
}

	url := "https://api-free.deepl.com/v2/translate?auth_key=" + apiKey + "&text=" + text + "&target_lang=" + targetLang
	if sourceLang != "" {
		url += "&source_lang=" + sourceLang
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Translations []struct {
			DetectedSourceLanguage string `json:"detected_source_language"`
			Text                   string `json:"text"`
		} `json:"translations"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	if len(result.Translations) == 0 {
		return err("no translations found")
}

	return ok(result.Translations[0].Text)
}