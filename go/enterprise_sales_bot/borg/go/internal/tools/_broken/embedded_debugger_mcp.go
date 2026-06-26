package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleDebugRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	addr, _ :=getInt(args, "address")
	url := fmt.Sprintf("http://%s:%d/read?addr=%d", host, port, addr)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to read memory: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	value, found := result["value"].(float64)
	if !found {
		return err("missing 'value'")
}

	return ok(fmt.Sprintf("Read %d: %d", addr, int(value)))
}