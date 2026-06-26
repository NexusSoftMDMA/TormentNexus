package codeexec

/**
 * @file wasm_sandbox.go
 * @module go/internal/codeexec
 *
 * WHAT: WebAssembly sandbox for safe, isolated Go code execution.
 * Compiles Go code to WASM and runs it in a controlled environment
 * with CPU/memory limits and no filesystem/network access.
 *
 * WHY: Process-based `go run` is too permissive for untrusted code.
 * WASM provides true sandboxing: no OS access, deterministic limits,
 * fast cold start via precompiled modules.
 *
 * ADDED: v1.0.0-alpha.50
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WASMSandboxConfig controls the WASM execution environment.
type WASMSandboxConfig struct {
	// MaxMemoryMB is the memory limit in megabytes (default: 64).
	MaxMemoryMB int

	// MaxTimeout is the maximum execution timeout (default: 15s).
	MaxTimeout time.Duration

	// AllowNetwork controls whether the WASM module can access network.
	// Always false for untrusted code.
	AllowNetwork bool

	// AllowFS controls whether the WASM module can access the filesystem.
	// Always false for untrusted code.
	AllowFS bool

	// EnvVars are the environment variables to expose to the WASM module.
	EnvVars map[string]string
}

// WASMExecutionResult is the output of a WASM sandbox execution.
type WASMExecutionResult struct {
	Success    bool              `json:"success"`
	ExitCode   int               `json:"exitCode"`
	Stdout     string            `json:"stdout"`
	Stderr     string            `json:"stderr"`
	Duration   string            `json:"duration"`
	MemoryUsed string            `json:"memoryUsed,omitempty"`
	WASMSize   int64             `json:"wasmSize"`
	Sandboxed  bool              `json:"sandboxed"`
	Error      string            `json:"error,omitempty"`
	ExecutedAt time.Time          `json:"executedAt"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// WASMSandbox provides isolated Go code execution via WebAssembly.
type WASMSandbox struct {
	mu          sync.Mutex
	cacheDir    string
	wasmRuntime string // path to wasmtime or wasmer binary
	config      WASMSandboxConfig
	stats       WASMSandboxStats
}

// WASMSandboxStats tracks cumulative sandbox statistics.
type WASMSandboxStats struct {
	TotalExecutions int           `json:"totalExecutions"`
	SuccessCount    int           `json:"successCount"`
	FailureCount    int           `json:"failureCount"`
	TotalDuration   time.Duration `json:"totalDuration"`
	AvgDuration     time.Duration `json:"avgDuration"`
	CacheHits       int           `json:"cacheHits"`
	CacheMisses     int           `json:"cacheMisses"`
}

// NewWASMSandbox creates a new WASM sandbox with the given configuration.
func NewWASMSandbox(cfg WASMSandboxConfig) *WASMSandbox {
	if cfg.MaxMemoryMB <= 0 {
		cfg.MaxMemoryMB = 64
	}
	if cfg.MaxTimeout <= 0 {
		cfg.MaxTimeout = 15 * time.Second
	}
	// Security: never allow network/FS for sandbox
	cfg.AllowNetwork = false
	cfg.AllowFS = false

	cacheDir := filepath.Join(os.TempDir(), "tormentnexus-wasm-cache")
	os.MkdirAll(cacheDir, 0755)

	sandbox := &WASMSandbox{
		cacheDir: cacheDir,
		config:   cfg,
	}

	// Detect available WASM runtime
	sandbox.wasmRuntime = sandbox.detectWASMRuntime()

	return sandbox
}

// IsAvailable returns whether WASM execution is possible on this system.
func (ws *WASMSandbox) IsAvailable() bool {
	return ws.wasmRuntime != "" && ws.canCompileWASM()
}

// Execute compiles Go code to WASM and runs it in the sandbox.
func (ws *WASMSandbox) Execute(ctx context.Context, code string) (*WASMExecutionResult, error) {
	start := time.Now()

	if !ws.IsAvailable() {
		return nil, fmt.Errorf("WASM sandbox not available: need both 'go' compiler and a WASM runtime (wasmtime/wasmer)")
	}

	// Check cache
	cacheKey := ws.cacheKey(code)
	cachedWASM := filepath.Join(ws.cacheDir, cacheKey+".wasm")
	if _, err := os.Stat(cachedWASM); err == nil {
		ws.mu.Lock()
		ws.stats.CacheHits++
		ws.mu.Unlock()
		return ws.runWASM(ctx, cachedWASM, start)
	}

	ws.mu.Lock()
	ws.stats.CacheMisses++
	ws.mu.Unlock()

	// 1. Write Go source to temp file
	tmpDir, err := os.MkdirTemp("", "tormentnexus-wasm-build-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcFile, []byte(ws.wrapCode(code)), 0644); err != nil {
		return nil, fmt.Errorf("write source: %w", err)
	}

	// 2. Compile Go → WASM
	wasmFile := filepath.Join(tmpDir, "main.wasm")
	compileCtx, compileCancel := context.WithTimeout(ctx, 30*time.Second)
	defer compileCancel()

	compileCmd := exec.CommandContext(compileCtx, "go", "build",
		"-o", wasmFile,
		"-trimpath",
		"-ldflags", "-s -w",
		"GOOS=js", "GOARCH=wasm",
		srcFile,
	)
	// Set environment for cross-compilation
	compileCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	var compileStdout, compileStderr bytes.Buffer
	compileCmd.Stdout = &compileStdout
	compileCmd.Stderr = &compileStderr

	if err := compileCmd.Run(); err != nil {
		return &WASMExecutionResult{
			Success:    false,
			ExitCode:   -1,
			Stderr:     compileStderr.String(),
			Duration:   time.Since(start).String(),
			Sandboxed:  true,
			Error:      fmt.Sprintf("WASM compilation failed: %v", err),
			ExecutedAt: start.UTC(),
		}, nil
	}

	// 3. Cache the compiled WASM
	wasmData, err := os.ReadFile(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("read compiled WASM: %w", err)
	}
	os.WriteFile(cachedWASM, wasmData, 0644)

	// 4. Execute
	return ws.runWASM(ctx, wasmFile, start)
}

// ExecutePrecompiled runs an already-compiled WASM module.
func (ws *WASMSandbox) ExecutePrecompiled(ctx context.Context, wasmPath string) (*WASMExecutionResult, error) {
	start := time.Now()
	return ws.runWASM(ctx, wasmPath, start)
}

// Stats returns cumulative sandbox execution statistics.
func (ws *WASMSandbox) Stats() WASMSandboxStats {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.stats
}

// --- Internal ---

func (ws *WASMSandbox) runWASM(ctx context.Context, wasmPath string, start time.Time) (*WASMExecutionResult, error) {
	timeout := ws.config.MaxTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	switch {
	case strings.Contains(ws.wasmRuntime, "wasmtime"):
		cmd = exec.CommandContext(ctx, ws.wasmRuntime,
			"--max-wasm-stack", "1048576",
			"--wasm-max-memory", fmt.Sprintf("%d", ws.config.MaxMemoryMB*1024*1024),
			wasmPath,
		)
	case strings.Contains(ws.wasmRuntime, "wasmer"):
		cmd = exec.CommandContext(ctx, ws.wasmRuntime,
			"run",
			"--max-memory", fmt.Sprintf("%dM", ws.config.MaxMemoryMB),
			wasmPath,
		)
	default:
		// Fallback: just run with basic flags
		cmd = exec.CommandContext(ctx, ws.wasmRuntime, wasmPath)
	}

	// Build controlled environment (no host env leakage)
	envPairs := []string{}
	for k, v := range ws.config.EnvVars {
		envPairs = append(envPairs, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = envPairs

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	wasmSize := int64(0)
	if fi, err := os.Stat(wasmPath); err == nil {
		wasmSize = fi.Size()
	}

	err := cmd.Run()
	duration := time.Since(start)

	result := &WASMExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Duration:   duration.String(),
		WASMSize:   wasmSize,
		Sandboxed:  true,
		ExecutedAt: start.UTC(),
		Metadata: map[string]string{
			"runtime":    filepath.Base(ws.wasmRuntime),
			"maxMemory":  fmt.Sprintf("%dMB", ws.config.MaxMemoryMB),
			"timeout":    timeout.String(),
			"network":    "denied",
			"filesystem": "denied",
		},
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else if ctx.Err() == context.DeadlineExceeded {
			result.ExitCode = -1
			result.Success = false
			result.Error = "execution timed out"
		} else {
			result.ExitCode = -1
			result.Success = false
			result.Error = err.Error()
		}
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	// Truncate large outputs
	if len(result.Stdout) > 50000 {
		result.Stdout = result.Stdout[:50000] + "\n... (truncated)"
	}
	if len(result.Stderr) > 25000 {
		result.Stderr = result.Stderr[:25000] + "\n... (truncated)"
	}

	// Update stats
	ws.mu.Lock()
	ws.stats.TotalExecutions++
	if result.Success {
		ws.stats.SuccessCount++
	} else {
		ws.stats.FailureCount++
	}
	ws.stats.TotalDuration += duration
	if ws.stats.TotalExecutions > 0 {
		ws.stats.AvgDuration = time.Duration(int64(ws.stats.TotalDuration) / int64(ws.stats.TotalExecutions))
	}
	ws.mu.Unlock()

	return result, nil
}

// wrapCode wraps user code in a proper Go main package if needed.
func (ws *WASMSandbox) wrapCode(code string) string {
	code = strings.TrimSpace(code)

	// If the code already has a package declaration, use as-is
	if strings.HasPrefix(code, "package ") {
		return code
	}

	// Otherwise, wrap in a main package
	return fmt.Sprintf(`package main

import (
	"fmt"
)

func main() {
	// User code below
%s
}`, indentCode(code))
}

func indentCode(code string) string {
	lines := strings.Split(code, "\n")
	var buf strings.Builder
	for _, line := range lines {
		buf.WriteString("\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return buf.String()
}

// cacheKey generates a cache key from the code content.
func (ws *WASMSandbox) cacheKey(code string) string {
	// Simple hash — use a truncated SHA-like approach
	hash := uint32(0)
	for _, ch := range code {
		hash = hash*31 + uint32(ch)
	}
	return fmt.Sprintf("tormentnexus-%08x", hash)
}

// detectWASMRuntime finds an available WASM runtime on the system.
func (ws *WASMSandbox) detectWASMRuntime() string {
	candidates := []string{"wasmtime", "wasmer", "wasm3"}
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// canCompileWASM checks if the Go compiler supports WASM targets.
func (ws *WASMSandbox) canCompileWASM() bool {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	// Go has built-in WASM support since 1.11
	return true
}

// CompileToWASM compiles Go source to a WASM binary at the given output path.
// This is useful for pre-building sandbox modules.
func CompileToWASM(ctx context.Context, srcPath, outputPath string) error {
	compileCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(compileCtx, "go", "build",
		"-o", outputPath,
		"-trimpath",
		"-ldflags", "-s -w",
	)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Dir = filepath.Dir(srcPath)
	cmd.Args = append(cmd.Args, filepath.Base(srcPath))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WASM compilation failed: %w\n%s", err, stderr.String())
	}
	return nil
}

// MarshalStats returns the sandbox stats as JSON.
func (ws *WASMSandbox) MarshalStats() json.RawMessage {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	data, _ := json.Marshal(ws.stats)
	return data
}
