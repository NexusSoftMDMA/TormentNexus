package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleGetTrustScore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		return err("server name required")
}

	resp, e := http.DefaultClient.Get("https://observatory.trust/v1/trust/" + server)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data struct {
		Score int `json:"score"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Trust score for %s: %d/100", server, data.Score))
}

func HandleCheckHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		return err("server name required")
}

	start := time.Now()
	resp, e := http.DefaultClient.Get("https://observatory.trust/v1/health/" + server)
	if e != nil {
		return err(fmt.Sprintf("health check failed: %v", e))
}

	resp.Body.Close()
	latency := time.Since(start).Milliseconds()
	return ok(fmt.Sprintf("%s is healthy, latency %dms", server, latency))
}