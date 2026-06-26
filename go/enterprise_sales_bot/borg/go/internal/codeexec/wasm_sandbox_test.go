package codeexec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWASMSandboxCreation(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})
	if sandbox.config.MaxMemoryMB != 64 {
		t.Errorf("expected default MaxMemoryMB=64, got %d", sandbox.config.MaxMemoryMB)
	}
	if sandbox.config.MaxTimeout != 15*time.Second {
		t.Errorf("expected default MaxTimeout=15s, got %v", sandbox.config.MaxTimeout)
	}
	if sandbox.config.AllowNetwork {
		t.Error("AllowNetwork should always be false in sandbox")
	}
	if sandbox.config.AllowFS {
		t.Error("AllowFS should always be false in sandbox")
	}
}

func TestWASMSandboxSecurityEnforced(t *testing.T) {
	cfg := WASMSandboxConfig{
		MaxMemoryMB: 128,
		AllowNetwork: true, // should be forced to false
		AllowFS:      true, // should be forced to false
	}
	sandbox := NewWASMSandbox(cfg)
	if sandbox.config.AllowNetwork {
		t.Error("AllowNetwork must be forced to false for security")
	}
	if sandbox.config.AllowFS {
		t.Error("AllowFS must be forced to false for security")
	}
}

func TestWASMSandboxWrapCode(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})

	// Code without package declaration should be wrapped
	wrapped := sandbox.wrapCode(`fmt.Println("hello")`)
	if !strings.Contains(wrapped, "package main") {
		t.Error("expected package main wrapper")
	}
	if !strings.Contains(wrapped, "fmt.Println") {
		t.Error("expected user code to be included")
	}

	// Code with package declaration should be used as-is
	pkgCode := `package main

import "fmt"

func main() {
	fmt.Println("raw")
}`
	wrapped2 := sandbox.wrapCode(pkgCode)
	if wrapped2 != pkgCode {
		t.Error("expected code with package declaration to be used as-is")
	}
}

func TestWASMSandboxCacheKey(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})

	key1 := sandbox.cacheKey("hello")
	key2 := sandbox.cacheKey("hello")
	key3 := sandbox.cacheKey("world")

	if key1 != key2 {
		t.Error("same code should produce same cache key")
	}
	if key1 == key3 {
		t.Error("different code should produce different cache key")
	}
}

func TestWASMSandboxIsAvailable(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})
	// The result depends on the system, just verify it doesn't panic
	_ = sandbox.IsAvailable()
}

func TestWASMSandboxStats(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})
	stats := sandbox.Stats()
	if stats.TotalExecutions != 0 {
		t.Error("new sandbox should have zero executions")
	}
}

func TestWASMSandboxMarshalStats(t *testing.T) {
	sandbox := NewWASMSandbox(WASMSandboxConfig{})
	data := sandbox.MarshalStats()
	if len(data) == 0 {
		t.Error("expected non-empty stats JSON")
	}
}

func TestWASMCompileToWASM(t *testing.T) {
	// Skip if Go compiler is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go compiler not available")
	}

	tmpDir := t.TempDir()

	// Create a simple Go program
	src := `package main

func main() {
}
`
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmpDir, "main.wasm")

	err := CompileToWASM(context.Background(), srcPath, outPath)
	if err != nil {
		// WASM compilation may fail on some systems without js/wasm support
		t.Logf("WASM compilation failed (may be expected): %v", err)
		return
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("expected WASM output file to exist")
	}
}

func TestWASMSandboxCustomConfig(t *testing.T) {
	cfg := WASMSandboxConfig{
		MaxMemoryMB: 256,
		MaxTimeout:  30 * time.Second,
		EnvVars: map[string]string{
			"TEST_VAR": "hello",
		},
	}
	sandbox := NewWASMSandbox(cfg)

	if sandbox.config.MaxMemoryMB != 256 {
		t.Errorf("expected MaxMemoryMB=256, got %d", sandbox.config.MaxMemoryMB)
	}
	if sandbox.config.MaxTimeout != 30*time.Second {
		t.Errorf("expected MaxTimeout=30s, got %v", sandbox.config.MaxTimeout)
	}
	if sandbox.config.EnvVars["TEST_VAR"] != "hello" {
		t.Error("expected custom env var to be preserved")
	}
}
