package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCheckProfanity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return ok("No text provided")
}

	url := "https://www.purgomalum.com/service/containsprofanity?text=" + text
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse JSON: " + e.Error())
}

	found, _ := result["containsProfanity"].(bool)
	if found {
		return success("Profanity detected")
}

	return success("No profanity found")
}

func HandleSanitizeText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return ok("No text provided")
}

	url := "https://www.purgomalum.com/service/plain?text=" + text
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return success(string(body))
}// touch 1781132138
