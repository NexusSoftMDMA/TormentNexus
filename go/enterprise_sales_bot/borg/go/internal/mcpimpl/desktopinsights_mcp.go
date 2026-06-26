package mcpimpl

import (
	"context"
	"runtime"
	"time"
)

func HandleGetSystemInfo_desktopinsights_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, _ :=getString(args, "hostname")
	if hostname == "" {
		hostname = "localhost"
	}
	cpus := runtime.NumCPU()
	osName := runtime.GOOS
	arch := runtime.GOARCH
	info := "System: " + osName + " " + arch + ", CPUs: " + string(rune(cpus+'0')) + ", Host: " + hostname + ", Time: " + time.Now().Format(time.RFC3339)
	return success(info)
}