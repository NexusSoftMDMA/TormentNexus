package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleAlgorandBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	round, _ :=getInt(args, "round")
	baseURL := os.Getenv("ALGORAND_NODE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	url := fmt.Sprintf("%s/v2/blocks/%d", baseURL, round)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch block: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err("bad JSON: " + e.Error())
}

	hash, found := data["hash"].(string)
	if !found {
		hash = "unknown"
	}
	ts, found := data["timestamp"].(float64)
	if !found {
		ts = 0
	}
	msg := fmt.Sprintf("Block %d: hash=%s, timestamp=%.0f", round, hash, ts)
	return ok(msg)
}