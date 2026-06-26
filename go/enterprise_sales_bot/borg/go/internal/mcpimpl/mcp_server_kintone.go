package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetRecord_mcp_server_kintone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	app, _ :=getInt(args, "app")
	id, _ :=getInt(args, "id")
	baseURL := os.Getenv("KINTONE_BASE_URL")
	if baseURL == "" {
		return err("KINTONE_BASE_URL not set")
}

	url := fmt.Sprintf("%s/k/v1/record.json?app=%d&id=%d", baseURL, app, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	return ok(fmt.Sprintf("Record: %v", result))
}