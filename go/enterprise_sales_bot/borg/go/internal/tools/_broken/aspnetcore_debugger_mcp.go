package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListProcesses_aspnetcore_debugger_mcp(ctx context.Context, args map[string]string) (ToolResponse, error) {
	url := "http://localhost:5000/debugger/processes"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch processes: %v", e))
}

	defer resp.Body.Close()
	var processes []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&processes); e != nil {
		return err(fmt.Sprintf("failed to decode processes: %v", e))
}

	result := ""
	for _, p := range processes {
		name, found := p["name"].(string)
		if !found {
			name = "unknown"
		}
		pid, found := p["pid"].(float64)
		if found {
			result += fmt.Sprintf("PID: %.0f Name: %s\n", pid, name)
		} else {
			result += fmt.Sprintf("Name: %s (PID unknown)\n", name)

	}
	if result == "" {
		return ok("No processes found.")
}

	return ok(result)
}

}

func HandleAttachDebugger(ctx context.Context, args map[string]string) (ToolResponse, error) {
	pid, _ :=getString(args, "pid")
	if pid == "" {
		return err("pid is required")
}

	url := fmt.Sprintf("http://localhost:5000/debugger/attach?pid=%s", pid)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to attach debugger: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("attach failed with status %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("Debugger attached to process %s", pid))
}