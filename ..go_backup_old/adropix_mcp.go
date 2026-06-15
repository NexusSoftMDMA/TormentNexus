package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGenerateImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	resp, e := http.DefaultClient.Post("https://api.adropix.com/generate-image", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to call Adropix API: " + e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	url, found := result["image_url"].(string)
	if !found {
		return err("unexpected response format")
}

	return success(fmt.Sprintf("Image generated: %s", url))
}

func HandleGenerateCopy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	body, _ := json.Marshal(map[string]string{"topic": topic})
	resp, e := http.DefaultClient.Post("https://api.adropix.com/generate-copy", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to call Adropix API: " + e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	copyText, found := result["copy"].(string)
	if !found {
		return err("unexpected response format")
}

	return success(fmt.Sprintf("Marketing copy: %s", copyText))
}