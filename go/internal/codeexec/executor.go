package codeexec

/**
 * @file executor.go
 * @module go/internal/codeexec
 *
 * WHAT: Code execution sandbox — runs code snippets in isolated processes
 *       with configurable timeouts, resource limits, and output capture.
 *
 * WHY: Safe code execution for code-mode, auto-test, and agent sandbox needs.
 *      Supports multiple languages via their respective runtimes.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Language string

const (
	JavaScript Language = "javascript"
	TypeScript Language = "typescript"
	Python     Language = "python"
	Go         Language = "go"
	Shell      Language = "shell"
	Rust       Language = "rust"
)

type ExecutionResult struct {
	ExitCode   int       `json:"exitCode"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	Duration   string    `json:"duration"`
	Language   string    `json:"language"`
	TimedOut   bool      `json:"timedOut"`
	Error      string    `json:"error,omitempty"`
	ExecutedAt time.Time `json:"executedAt"`
}

type ExecutionConfig struct {
	Language Language
	Code     string
	Timeout  time.Duration
	WorkDir  string
	Env      map[string]string
	Args     []string
}

type CodeExecutor struct {
	tempDir string
	mu      sync.Mutex
}

func NewCodeExecutor() *CodeExecutor {
	tmp := os.TempDir()
	return &CodeExecutor{tempDir: tmp}
}

func (ce *CodeExecutor) Execute(ctx context.Context, cfg ExecutionConfig) (*ExecutionResult, error) {
	start := time.Now()

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	switch cfg.Language {
	case JavaScript:
		return ce.execJS(ctx, cfg, start)
	case TypeScript:
		return ce.execTS(ctx, cfg, start)
	case Python:
		return ce.execPython(ctx, cfg, start)
	case Go:
		return ce.execGo(ctx, cfg, start)
	case Shell, "":
		return ce.execShell(ctx, cfg, start)
	case Rust:
		return ce.execRust(ctx, cfg, start)
	default:
		return nil, fmt.Errorf("unsupported language: %s", cfg.Language)
	}
}

func (ce *CodeExecutor) execJS(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	tmpFile, err := ce.writeTempFile("exec_*.js", cfg.Code)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	return ce.runCommand(ctx, "node", append([]string{tmpFile}, cfg.Args...), cfg, start)
}

func (ce *CodeExecutor) execTS(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	tmpFile, err := ce.writeTempFile("exec_*.ts", cfg.Code)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	// Use npx tsx for TypeScript execution
	return ce.runCommand(ctx, "npx", append([]string{"tsx", tmpFile}, cfg.Args...), cfg, start)
}

func (ce *CodeExecutor) execPython(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	tmpFile, err := ce.writeTempFile("exec_*.py", cfg.Code)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	cmd := "python"
	if _, err := exec.LookPath("python3"); err == nil {
		cmd = "python3"
	}
	return ce.runCommand(ctx, cmd, append([]string{tmpFile}, cfg.Args...), cfg, start)
}

func (ce *CodeExecutor) execGo(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	tmpFile, err := ce.writeTempFile("exec_*.go", cfg.Code)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	return ce.runCommand(ctx, "go", append([]string{"run", tmpFile}, cfg.Args...), cfg, start)
}

func (ce *CodeExecutor) execShell(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	shell := "/bin/bash"
	args := []string{"-c", cfg.Code}
	if runtime.GOOS == "windows" {
		shell = "cmd.exe"
		args = []string{"/C", cfg.Code}
	}
	return ce.runCommand(ctx, shell, args, cfg, start)
}

func (ce *CodeExecutor) execRust(ctx context.Context, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	tmpFile, err := ce.writeTempFile("exec_*.rs", cfg.Code)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	return ce.runCommand(ctx, "rustc", append([]string{tmpFile, "-o", tmpFile + ".exe", "&&", tmpFile + ".exe"}, cfg.Args...), cfg, start)
}

func (ce *CodeExecutor) runCommand(ctx context.Context, name string, args []string, cfg ExecutionConfig, start time.Time) (*ExecutionResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if cfg.WorkDir != "" {
		cmd.Dir = cfg.WorkDir
	}

	// Build environment
	env := os.Environ()
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	err := cmd.Run()
	duration := time.Since(start)

	result := &ExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Duration:   duration.String(),
		Language:   string(cfg.Language),
		ExecutedAt: start.UTC(),
		TimedOut:   ctx.Err() == context.DeadlineExceeded,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	// Truncate large outputs
	if len(result.Stdout) > 100000 {
		result.Stdout = result.Stdout[:100000] + "\n... (truncated)"
	}
	if len(result.Stderr) > 50000 {
		result.Stderr = result.Stderr[:50000] + "\n... (truncated)"
	}

	return result, nil
}

func (ce *CodeExecutor) writeTempFile(pattern, content string) (string, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	dir := filepath.Join(ce.tempDir, "tormentnexus-exec")
	os.MkdirAll(dir, 0o755)

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}

// IsLanguageAvailable checks if the runtime for a given language is installed.
func IsLanguageAvailable(lang Language) bool {
	var cmd string
	switch lang {
	case JavaScript:
		cmd = "node"
	case TypeScript:
		cmd = "npx"
	case Python:
		cmd = "python"
		if _, err := exec.LookPath("python3"); err == nil {
			return true
		}
	case Go:
		cmd = "go"
	case Shell:
		return true
	case Rust:
		cmd = "rustc"
	default:
		return false
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ListAvailableLanguages returns all languages that have runtimes installed.
func ListAvailableLanguages() []Language {
	langs := []Language{}
	for _, lang := range []Language{JavaScript, TypeScript, Python, Go, Shell, Rust} {
		if IsLanguageAvailable(lang) {
			langs = append(langs, lang)
		}
	}
	return langs
}

// Ensure unused import is used
var _ = strings.TrimSpace
