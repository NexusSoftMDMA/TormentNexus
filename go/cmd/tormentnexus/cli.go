package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/buildinfo"
)

type managedProcess struct {
	Name     string
	Cmd      *exec.Cmd
	Port     int
	ReadyURL string
	Stop     func() error
}

func cmdStart(args []string) int {
	workspaceRoot := detectWorkspaceRoot()
	goBinary := findGoBinary(workspaceRoot)
	if goBinary == "" {
		fmt.Println("[CLI] No Go binary found. Build with: cd go && go build -o tormentnexus.exe ./cmd/tormentnexus")
		return 1
	}

	fmt.Printf("[CLI] TormentNexus v%s\n", buildinfo.Version)
	fmt.Printf("[CLI] Workspace: %s\n", workspaceRoot)
	fmt.Printf("[CLI] Go binary: %s\n", goBinary)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Parse --runtime flag
	runtime := "go"
	for i, a := range args {
		if a == "--runtime" && i+1 < len(args) {
			runtime = args[i+1]
		}
	}

	processes := []*managedProcess{}

	// 1. Start Go sidecar
	if runtime == "go" || runtime == "auto" {
		goPort := 7778
		goProc := startGoSidecar(goBinary, workspaceRoot, goPort)
		if goProc == nil {
			goPort = 7777
			goProc = startGoSidecar(goBinary, workspaceRoot, goPort)
		}
		if goProc != nil {
			processes = append(processes, &managedProcess{
				Name:     "Go sidecar",
				Cmd:      goProc,
				Port:     goPort,
				ReadyURL: fmt.Sprintf("http://127.0.0.1:%d/health", goPort),
				Stop: func() error {
					return goProc.Process.Signal(os.Interrupt)
				},
			})
			fmt.Printf("[CLI] ✅ Go sidecar started on port %d\n", goPort)
		} else {
			fmt.Println("[CLI] ❌ Failed to start Go sidecar")
		}
	}

	// 2. Start Next.js dashboard
	dashProc := startDashboard(workspaceRoot)
	if dashProc != nil {
		processes = append(processes, &managedProcess{
			Name: "Dashboard",
			Cmd:  dashProc,
			Port: 3000,
			Stop: func() error {
				return dashProc.Process.Signal(os.Interrupt)
			},
		})
		fmt.Println("[CLI] ✅ Dashboard starting on port 3000")
	}

	fmt.Println()
	fmt.Println("[CLI] Services starting... Press Ctrl+C to stop.")
	fmt.Println()

	// Wait for ready signals
	for _, p := range processes {
		if p.ReadyURL != "" {
			waitForReady(p.ReadyURL, 15*time.Second)
		}
	}

	// Print status
	fmt.Println("[CLI] ─────────────────────────────")
	for _, p := range processes {
		status := checkHealth(p.ReadyURL)
		fmt.Printf("[CLI] %s: %s (port %d)\n", p.Name, status, p.Port)
	}
	fmt.Println("[CLI] ─────────────────────────────")
	fmt.Println()

	// Write PID file for external management
	writePIDFile(filepath.Join(workspaceRoot, "data", "go_cli.pid"), os.Getpid())

	// Wait for shutdown signal
	<-ctx.Done()
	fmt.Println("\n[CLI] Shutting down...")

	// Stop processes in reverse order
	for i := len(processes) - 1; i >= 0; i-- {
		p := processes[i]
		fmt.Printf("[CLI] Stopping %s...\n", p.Name)
		if err := p.Stop(); err != nil {
			fmt.Printf("[CLI] Error stopping %s: %v\n", p.Name, err)
		}
	}

	os.Remove(filepath.Join(workspaceRoot, "data", "go_cli.pid"))
	fmt.Println("[CLI] All services stopped.")
	return 0
}

func cmdStop(args []string) int {
	// Kill Go sidecar
	exec.Command("taskkill", "/F", "/IM", "tormentnexus.exe").Run()

	// Kill dashboard node processes on relevant ports
	exec.Command("taskkill", "/F", "/FI", "WINDOWTITLE eq *next*").Run()

	fmt.Println("[CLI] All TormentNexus services stopped.")
	return 0
}

func cmdStatus(args []string) int {
	checks := []struct {
		Name string
		URL  string
	}{
		{"Go sidecar", "http://127.0.0.1:7778/health"},
		{"Go sidecar (alt)", "http://127.0.0.1:7777/health"},
		{"Dashboard", "http://127.0.0.1:3000/dashboard"},
		{"Dashboard (dev)", "http://127.0.0.1:7779/"},
		{"Dashboard (7779)", "http://127.0.0.1:7779/"},
		{"TS control plane", "http://127.0.0.1:4100/health"},
	}

	fmt.Println("[CLI] TormentNexus Status")
	fmt.Println("[CLI] ─────────────────────────────")
	for _, c := range checks {
		status := checkHealth(c.URL)
		fmt.Printf("[CLI] %-20s %s\n", c.Name+":", status)
	}
	fmt.Println("[CLI] ─────────────────────────────")
	fmt.Printf("[CLI] Version: %s\n", buildinfo.Version)
	return 0
}

// ─── Helpers ───

func detectWorkspaceRoot() string {
	if w := os.Getenv("TORMENTNEXUS_WORKSPACE"); w != "" {
		return w
	}
	// Walk up from cwd looking for go.mod in a 'go' subdirectory
	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go", "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}

func findGoBinary(workspaceRoot string) string {
	candidates := []string{
		filepath.Join(workspaceRoot, "tormentnexus.exe"),
		filepath.Join(workspaceRoot, "bin", "tormentnexus.exe"),
		filepath.Join(workspaceRoot, "go", "tormentnexus.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func startGoSidecar(binary, workspaceRoot string, port int) *exec.Cmd {
	cmd := exec.Command(binary, "serve", "--port", fmt.Sprintf("%d", port))
	cmd.Dir = workspaceRoot
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		fmt.Printf("[CLI] Failed to start Go sidecar on port %d: %v\n", port, err)
		return nil
	}
	return cmd
}

func startDashboard(workspaceRoot string) *exec.Cmd {
	dashDir := filepath.Join(workspaceRoot, "apps", "web")
	if _, err := os.Stat(dashDir); os.IsNotExist(err) {
		fmt.Println("[CLI] Dashboard directory not found, skipping.")
		return nil
	}

	npxPath := findExecutable("npx")
	if npxPath == "" {
		fmt.Println("[CLI] npx not found, skipping dashboard.")
		return nil
	}

	cmd := exec.Command(npxPath, "next", "dev", "--port", "3000")
	cmd.Dir = dashDir
	cmd.Env = append(os.Environ(), "NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		fmt.Printf("[CLI] Failed to start dashboard: %v\n", err)
		return nil
	}
	return cmd
}

func findExecutable(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	// Check common locations
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "nodejs", name+".cmd"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "pnpm", name+".cmd"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func waitForReady(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func checkHealth(url string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "❌ DOWN"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// Try to extract version from JSON response
		var data map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if v, ok := data["version"].(string); ok {
				return fmt.Sprintf("✅ OK (v%s)", v)
			}
		}
		return "✅ OK"
	}
	return fmt.Sprintf("❌ HTTP %d", resp.StatusCode)
}

func writePIDFile(path string, pid int) {
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(fmt.Sprintf("%d", pid)), 0644)
}
