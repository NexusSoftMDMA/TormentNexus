package tools

import (
	"fmt"
	"os"
	"runtime"
)

func (r *Registry) registerSystemTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "get_system_stats",
		Description: "Retrieves real-time system metrics (Warp/Pi CLI parity). Arguments: none",
		Execute: func(args map[string]interface{}) (string, error) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			hostname, _ := os.Hostname()

			stats := fmt.Sprintf("Hostname: %s\nOS: %s\nArch: %s\nCPUs: %d\nAllocated Memory: %v MB\nTotal Memory: %v MB\nGoroutines: %d",
				hostname, runtime.GOOS, runtime.GOARCH, runtime.NumCPU(),
				m.Alloc/1024/1024, m.TotalAlloc/1024/1024, runtime.NumGoroutine())

			return stats, nil
		},
	})
}
