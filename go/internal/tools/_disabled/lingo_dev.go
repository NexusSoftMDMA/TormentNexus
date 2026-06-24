package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleTranslate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	target, _ :=getString(args, "target")
	if text == "" || target == "" {
		return err("text and target are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://lingo.dev/api/translate?text="+text+"&target="+target, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return success("translation: " + result["translated"].(string))
}

func HandleDefine(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	if word == "" {
		return err("word is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://lingo.dev/api/define?word="+word, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	def, found := result["definition"].(string)
	if !found {
		return err("no definition found")
}

	return success("definition: " + def)
}