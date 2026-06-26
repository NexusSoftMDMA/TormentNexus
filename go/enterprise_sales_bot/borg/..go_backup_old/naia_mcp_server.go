package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleVisibility(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	radius, _ :=getInt(args, "radius")
	url := "https://api.naia.ai/visibility?location=" + location + "&radius=" + fmt.Sprint(radius)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Visibility analysis: %v", result))
}

func HandleContentGeneration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	maxTokens, _ :=getInt(args, "max_tokens")
	url := "https://api.naia.ai/content?prompt=" + prompt + "&max_tokens=" + fmt.Sprint(maxTokens)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok("Generated content: " + string(body))
}