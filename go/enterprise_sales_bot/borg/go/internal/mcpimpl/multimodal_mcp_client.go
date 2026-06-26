package mcpimpl

import (
	"context"
	"net/http"
	"io"
)

func HandleProcessVoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("missing input")
}

	resp, e := http.DefaultClient.Post("https://api.example.com/voice", "text/plain", nil)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}

func HandleMultimodalQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	image, _ :=getString(args, "image")
	if text == "" && image == "" {
		return err("provide text or image")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/multimodal?text=" + text + "&image=" + image)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}