package tools

import (
	"context"
	"encoding/json"
	"runtime"
)

func HandleGetCpuUsage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cpuCount := runtime.NumCPU()
	info := map[string]interface{}{"cpu_count": cpuCount}
	data, e := json.Marshal(info)
	if e != nil {
		return err("failed to marshal cpu info")
}

	return ok(string(data))
}

func HandleGetMemoryUsage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info := map[string]interface{}{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      m.NumGC,
	}
	data, e := json.Marshal(info)
	if e != nil {
		return err("failed to marshal memory info")
}

	return ok(string(data))
}