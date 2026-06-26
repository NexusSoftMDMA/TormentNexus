package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		model = "default"
	}
	prompt, _ :=getString(args, "prompt")
	apiBase := "https://api.example.com"
	url := apiBase + "/generate?model=" + model + "&prompt=" + prompt
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	return success(string(body))
}

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/models")
	if e != nil {
		return err("list failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	return ok(string(body))
}