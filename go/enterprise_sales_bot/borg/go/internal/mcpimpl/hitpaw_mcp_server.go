package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleEnhanceImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input_path")
	output, _ :=getString(args, "output_path")
	model, _ :=getString(args, "model")
	body := map[string]string{"input": input, "output": output, "model": model}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/enhance_image", bytes.NewBuffer(jsonData))
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	return success("image enhancement started")
}

func HandleEnhanceVideo_hitpaw_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input_path")
	output, _ :=getString(args, "output_path")
	model, _ :=getString(args, "model")
	body := map[string]string{"input": input, "output": output, "model": model}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/enhance_video", bytes.NewBuffer(jsonData))
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	return success("video enhancement started")
}