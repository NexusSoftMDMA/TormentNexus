package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTrackParcel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	trackingNumber, _ :=getString(args, "tracking_number")
	if trackingNumber == "" {
		return err("tracking_number is required")
}

	url := fmt.Sprintf("https://api.parcelapp.net/v1/track?q=%s", trackingNumber)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("%v", result))
}