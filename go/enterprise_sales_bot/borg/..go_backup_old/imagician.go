package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	size, _ :=getString(args, "size")
	if prompt == "" {
		return err("prompt is required")
}

	url := "https://api.imagician.dev/generate?prompt=" + prompt + "&size=" + size
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	imageURL, found := result["image_url"].(string)
	if !found {
		return err("missing image_url in response")
}

	return ok(imageURL)
}

func HandleDescribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	if imageURL == "" {
		return err("image_url is required")
}

	url := "https://api.imagician.dev/describe?url=" + imageURL
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	description, found := result["description"].(string)
	if !found {
		return err("missing description in response")
}

	return ok(description)
}