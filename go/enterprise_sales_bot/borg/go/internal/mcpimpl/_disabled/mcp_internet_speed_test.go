package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func HandleSpeedTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://www.google.com"
	}
	start := time.Now()
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach URL: " + e.Error())
}

	defer resp.Body.Close()
	elapsed := time.Since(start)
	return ok(fmt.Sprintf("Ping to %s took %v", url, elapsed))
}