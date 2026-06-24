package tools

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleGetDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("FORTIMONITOR_API_KEY")

	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = os.Getenv("FORTIMONITOR_BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.fortimonitor.com"
		}
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/devices", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status + ": " + string(body))
}

	return ok("devices: " + string(body))
}
}