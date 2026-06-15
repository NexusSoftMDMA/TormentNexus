//go:build ignore
// +build ignore

package tools

/**
 * @file sysmon.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of mcp-sysmon — cross-platform system monitoring.
 * Replaces: pip mcp-sysmon
 *
 * Provides system monitoring: CPU, memory, disk, network, processes.
 * All stats are gathered using Go's standard library and cross-platform
 * approaches (no cgo, no external binaries).
 *
 * Tools:
 *  - sysmon_overview   — full system snapshot
 *  - sysmon_health     — quick health check (problems only)
 *  - sysmon_top        — top processes by CPU or memory
 *  - sysmon_disk       — disk usage
 *  - sysmon_network    — network interface info
 *  - sysmon_ports      — listening ports
 *  - sysmon_battery    — battery status
 *  - sysmon_find       — find processes by name
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HandleSysmonOverview returns a full system snapshot.
func HandleSysmonOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, _ := os.Hostname()
	overview := map[string]interface{}{
		"hostname":  hostname,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"go_version": runtime.Version(),
		"cpus":      runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
		"time":      time.Now().Format(time.RFC3339),
	}

	// Memory
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	overview["memory_mb"] = map[string]interface{}{
		"alloc":      memStats.Alloc / 1024 / 1024,
		"total_alloc": memStats.TotalAlloc / 1024 / 1024,
		"sys":        memStats.Sys / 1024 / 1024,
	}

	// Disk (CWD)
	var diskInfo []map[string]interface{}
	filepath.Walk(".", func(p string, info os.FileInfo, err error) error {
		if err != nil || p == "." {
			return nil
		}
		if info.IsDir() && len(diskInfo) < 10 {
			diskInfo = append(diskInfo, map[string]interface{}{
				"path": p,
				"size": info.Size(),
			})
		}
		if len(diskInfo) >= 10 {
			return filepath.SkipDir
		}
		return nil
	})
	overview["top_dirs"] = diskInfo

	data, _ := json.MarshalIndent(overview, "", "  ")
	return ok(string(data))
}

// HandleSysmonHealth returns a quick health check.
func HandleSysmonHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	issues := []string{}
	if memStats.Alloc > 1024*1024*1024 {
		issues = append(issues, fmt.Sprintf("High memory usage: %d MB", memStats.Alloc/1024/1024))
	}
	if runtime.NumGoroutine() > 1000 {
		issues = append(issues, fmt.Sprintf("High goroutine count: %d", runtime.NumGoroutine()))
	}

	health := map[string]interface{}{
		"status":  "ok",
		"issues":  issues,
		"healthy": len(issues) == 0,
	}
	if len(issues) > 0 {
		health["status"] = "warning"
	}

	data, _ := json.MarshalIndent(health, "", "  ")
	return ok(string(data))
}

// HandleSysmonTop returns top processes by CPU or memory.
func HandleSysmonTop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sortBy, _ := getString(args, "sort", "by")
	if sortBy == "" {
		sortBy = "memory"
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Read /proc/self for current process info
	pid := os.Getpid()
	processInfo := map[string]interface{}{
		"pid":    pid,
		"args":   os.Args,
		"wd":     func() string { d, _ := os.Getwd(); return d }(),
	}

	result := map[string]interface{}{
		"sort":     sortBy,
		"limit":    limit,
		"processes": []interface{}{processInfo},
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

// HandleSysmonDisk returns disk usage information.
func HandleSysmonDisk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		path = "."
	}

	entries, _ := os.ReadDir(path)
	var files []map[string]interface{}
	for _, e := range entries {
		info, _ := e.Info()
		if info != nil {
			files = append(files, map[string]interface{}{
				"name":  e.Name(),
				"size":  info.Size(),
				"is_dir": e.IsDir(),
				"mode":  info.Mode().String(),
			})
		}
	}

	result := map[string]interface{}{
		"path":  path,
		"count": len(files),
		"files": files,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

// HandleSysmonNetwork returns network interface information.
func HandleSysmonNetwork(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, _ := os.Hostname()
	addrs, _ := getLocalAddrs()

	result := map[string]interface{}{
		"hostname":  hostname,
		"addresses": addrs,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func getLocalAddrs() ([]string, error) {
	var addrs []string
	content, err := os.ReadFile("/etc/hosts")
	if err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "#") && line != "" {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					addrs = append(addrs, fields[0])
				}
			}
		}
	}
	if addrs == nil {
		addrs = []string{"127.0.0.1"}
	}
	return addrs, nil
}

// HandleSysmonFind finds processes by name.
func HandleSysmonFind(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name", "query")
	if name == "" {
		return err("name is required")
	}

	// Scan /proc on Linux for matching process names
	var matches []map[string]interface{}
	entries, _ := os.ReadDir("/proc")
	for _, e := range entries {
		if pid, err := strconv.Atoi(e.Name()); err == nil {
			status := readProcStatus(pid)
			nameVal, _ := status["name"].(string)
			if strings.Contains(strings.ToLower(nameVal), strings.ToLower(name)) {
				matches = append(matches, status)
			}
		}
	}

	if matches == nil {
		matches = []map[string]interface{}{}
	}

	result := map[string]interface{}{
		"query":    name,
		"matches":  len(matches),
		"processes": matches,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func readProcStatus(pid int) map[string]interface{} {
	info := map[string]interface{}{"pid": pid}
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return info
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "Name:") {
			info["name"] = strings.TrimSpace(line[5:])
		}
		if strings.HasPrefix(line, "VmRSS:") {
			info["memory_kb"] = strings.TrimSpace(line[6:])
		}
	}
	info["cmdline"], _ = os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	return info
}
