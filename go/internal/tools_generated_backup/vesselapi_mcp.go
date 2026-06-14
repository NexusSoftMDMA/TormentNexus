package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetVessel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mmsi, _ :=getString(args, "mmsi")
	if mmsi == "" {
		return err("mmsi is required")
}

	url := fmt.Sprintf("https://api.vesselapi.com/v1/vessel?mmsi=%s", mmsi)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
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
		return err("failed to parse JSON: " + e.Error())
}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}