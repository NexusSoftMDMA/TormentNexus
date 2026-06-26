package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "traceId")
	if id == "" {
		id = "latest"
	}
	url := fmt.Sprintf("https://api.digma.ai/traces/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Trace data: %v", result))
}