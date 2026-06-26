package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleEcho_cashcon57(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message argument required")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/get?msg=" + message)
	if e != nil {
		return err(fmt.Sprintf("HTTP error: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	return success(fmt.Sprintf("Echo response: %v", result))
}