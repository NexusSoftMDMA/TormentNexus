package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleRender(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagram, _ :=getString(args, "diagram")
	if diagram == "" {
		return err("diagram parameter is required")
}

	format, _ :=getString(args, "format")
	if format == "" {
		format = "svg"
	}
	baseURL := os.Getenv("DGMO_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	body := map[string]string{"diagram": diagram, "format": format}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(baseURL+"/render", "application/json", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(data))
}

func HandlePreview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagram, _ :=getString(args, "diagram")
	if diagram == "" {
		return err("diagram parameter is required")
}

	baseURL := os.Getenv("DGMO_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	body := map[string]string{"diagram": diagram}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(baseURL+"/preview", "application/json", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(data))
}// touch 1781132124
