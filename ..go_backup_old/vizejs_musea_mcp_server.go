package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleAnalyzeComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	project, _ :=getString(args, "project")
	url := "https://api.musea.dev/analyze"
	body, _ := json.Marshal(map[string]string{"component": component, "project": project})
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to analyze component: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return success(string(data))
}

func HandleGenerateVariant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	variant, _ :=getString(args, "variant")
	tokens, _ :=getString(args, "tokens")
	url := "https://api.musea.dev/generate-variant"
	payload := map[string]string{"component": component, "variant": variant, "tokens": tokens}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to generate variant: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}