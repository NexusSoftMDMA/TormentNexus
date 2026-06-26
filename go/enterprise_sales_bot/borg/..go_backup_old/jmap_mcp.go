package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleJmapListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "http://localhost:8080"
	resp, e := http.DefaultClient.Get(base + "/processes")
	if e != nil {
		return err("failed to fetch processes")
}

	defer resp.Body.Close()
	var processes []string
	if found := json.NewDecoder(resp.Body).Decode(&processes); found != nil {
		return err("decode error")
}

	msg := fmt.Sprintf("Processes: %v", processes)
	return ok(msg)
}

func HandleJmapHeapHisto(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getString(args, "pid")
	if pid == "" {
		return err("pid is required")
}

	base := "http://localhost:8080"
	url := base + "/heap/" + pid
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch heap histogram")
}

	defer resp.Body.Close()
	var histo []string
	if found := json.NewDecoder(resp.Body).Decode(&histo); found != nil {
		return err("decode error")
}

	msg := fmt.Sprintf("Heap histogram for PID %s: %v", pid, histo)
	return ok(msg)
}