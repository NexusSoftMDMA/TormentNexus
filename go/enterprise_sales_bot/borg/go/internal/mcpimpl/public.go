package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPublicIP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.Get("https://api.ipify.org?format=json")
	if e != nil {
		return err(fmt.Sprintf("Failed to fetch IP: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Failed to read response: %v", e))
}

	var data struct {
		IP string `json:"ip"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("Failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Your public IP is %s", data.IP))
}