package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleGetNodeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("missing base_url")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/node/info", nil)
	if e != nil {
		return err(fmt.Sprintf("request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("do: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("unmarshal: %v", e))
}

	return ok(fmt.Sprintf("%v", result))
}