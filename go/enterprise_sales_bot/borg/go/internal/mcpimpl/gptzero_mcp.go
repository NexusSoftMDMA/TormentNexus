package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type detectRequest struct {
	Document string `json:"document"`
}

type detectResponse struct {
	Result string `json:"result"`
}

func HandleDetectText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	apiKey := os.Getenv("GPTZERO_API_KEY")
	if apiKey == "" {
		return err("GPTZERO_API_KEY not set")
}

	body, e := json.Marshal(detectRequest{Document: text})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.gptzero.me/v2/detect", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request create error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	var result detectResponse
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err(fmt.Sprintf("unmarshal error: %v", e))
}

	return ok(result.Result)
}