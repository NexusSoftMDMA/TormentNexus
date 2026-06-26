package mcpimpl

import (
	"context"
	"encoding/json"
	"runtime"
)

func HandleGetSystemInfo_mcp_sysmon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	info := map[string]interface{}{
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"numCPU":   runtime.NumCPU(),
		"goVersion": runtime.Version(),
	}
	data, e := json.Marshal(info)
	if e != nil {
		return err("failed to marshal system info")
}

	return success(string(data))
}

func HandleGetMemory_mcp_sysmon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mem := map[string]interface{}{
		"alloc":      m.Alloc,
		"totalAlloc": m.TotalAlloc,
		"sys":        m.Sys,
		"numGC":      m.NumGC,
	}
	data, e := json.Marshal(mem)
	if e != nil {
		return err("failed to marshal memory info")
}

	return success(string(data))
}