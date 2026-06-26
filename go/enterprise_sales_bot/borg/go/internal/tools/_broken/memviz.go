package mcpimpl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMemoryStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/memory")
	if e != nil {
		return err("failed to fetch memory stats")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON")
}

	return ok(fmt.Sprintf("Memory: %v", result))
}

func HandleGenerateViz(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	kind, _ :=getString(args, "type")
	resp, e := http.DefaultClient.Post("http://localhost:8080/viz", "application/json", nil)
	if e != nil {
		return err("failed to generate visualization")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(fmt.Sprintf("Visualization of %s: %s", kind, string(body)))
}