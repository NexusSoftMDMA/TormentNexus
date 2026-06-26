// Package process provides process lifecycle management ported from
// packages/core/src/services/ProcessManager.ts.
//
// Supports spawning, writing to stdin, killing, and listing active processes.
package processmanager

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// ProcessConfig defines how a process should be spawned.
type ProcessConfig struct {
	SessionID string            `json:"sessionId"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Cwd       string            `json:"cwd"`
	Env       map[string]string `json:"env,omitempty"`
}

// ProcessOutput represents a chunk of process output.
type ProcessOutput struct {
	SessionID string `json:"sessionId"`
	Data      string `json:"data"`
	Type      string `json:"type"` // "stdout" or "stderr"
}

// ProcessExit represents a process exit event.
type ProcessExit struct {
	SessionID string `json:"sessionId"`
	Code      int    `json:"code"`
}

// ProcessManager manages spawned child processes.
type ProcessManager struct {
	mu       sync.RWMutex
	active   map[string]*managedProcess
	onOutput func(ProcessOutput)
	onExit   func(ProcessExit)
}

type managedProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	cancel func()
}

// NewProcessManager creates a new ProcessManager.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		active: make(map[string]*managedProcess),
	}
}

// OnOutput registers a callback for process output.
func (pm *ProcessManager) OnOutput(fn func(ProcessOutput)) {
	pm.onOutput = fn
}

// OnExit registers a callback for process exit.
func (pm *ProcessManager) OnExit(fn func(ProcessExit)) {
	pm.onExit = fn
}

// Spawn starts a new process and tracks its output.
func (pm *ProcessManager) Spawn(config ProcessConfig) (int, error) {
	ctx, cancel := createContext()
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	cmd.Dir = config.Cwd

	// Merge environment
	if len(config.Env) > 0 {
		env := os.Environ()
		for k, v := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Create pipes
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return -1, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return -1, fmt.Errorf("failed to start process: %w", err)
	}

	pid := cmd.Process.Pid

	mp := &managedProcess{
		cmd:    cmd,
		stdin:  stdinPipe,
		cancel: cancel,
	}

	pm.mu.Lock()
	pm.active[config.SessionID] = mp
	pm.mu.Unlock()

	// Stream stdout
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 && pm.onOutput != nil {
				pm.onOutput(ProcessOutput{
					SessionID: config.SessionID,
					Data:      string(buf[:n]),
					Type:      "stdout",
				})
			}
			if err != nil {
				break
			}
		}
	}()

	// Stream stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 && pm.onOutput != nil {
				pm.onOutput(ProcessOutput{
					SessionID: config.SessionID,
					Data:      string(buf[:n]),
					Type:      "stderr",
				})
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for exit
	go func() {
		err := cmd.Wait()
		code := 0
		if err != nil {
			if exitErr, ok := err.(interface{ ExitCode() int }); ok {
				code = exitErr.ExitCode()
			} else {
				code = -1
			}
		}

		pm.mu.Lock()
		delete(pm.active, config.SessionID)
		pm.mu.Unlock()

		if pm.onExit != nil {
			pm.onExit(ProcessExit{
				SessionID: config.SessionID,
				Code:      code,
			})
		}
	}()

	return pid, nil
}

// Write sends data to a process's stdin.
func (pm *ProcessManager) Write(sessionID string, data string) bool {
	pm.mu.RLock()
	mp, ok := pm.active[sessionID]
	pm.mu.RUnlock()

	if !ok || mp.stdin == nil {
		return false
	}

	_, err := mp.stdin.Write([]byte(data))
	return err == nil
}

// Kill terminates an active process.
func (pm *ProcessManager) Kill(sessionID string) bool {
	pm.mu.Lock()
	mp, ok := pm.active[sessionID]
	if ok {
		delete(pm.active, sessionID)
	}
	pm.mu.Unlock()

	if !ok {
		return false
	}

	mp.cancel()
	return true
}

// ListActiveSessions returns all active process session IDs.
func (pm *ProcessManager) ListActiveSessions() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	sessions := make([]string, 0, len(pm.active))
	for id := range pm.active {
		sessions = append(sessions, id)
	}
	return sessions
}

// ActiveCount returns the number of active processes.
func (pm *ProcessManager) ActiveCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.active)
}

// createContext creates a cancellable context.
func createContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

type ctxKey struct{}
