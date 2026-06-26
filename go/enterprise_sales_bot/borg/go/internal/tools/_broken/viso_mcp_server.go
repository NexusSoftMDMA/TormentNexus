package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStatus_viso_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.example.com/status"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach Viso server: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Viso server status: %s", string(body)))
}

func HandleProcessVision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	if imageURL == "" {
		return err("image_url is required")
}

	payload := map[string]string{"url": imageURL}
	data, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://api.example.com/process", "application/json", io.NopCloser(bytes.NewReader(data)))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return success("Vision processed: " + fmt.Sprintf("%v", result))
}