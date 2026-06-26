package httpapi

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type localAutoDevLoopConfig struct {
	MaxAttempts int    `json:"maxAttempts"`
	Type        string `json:"type"`
	Target      string `json:"target,omitempty"`
	Command     string `json:"command,omitempty"`
}

type localAutoDevLoop struct {
	ID             string                 `json:"id"`
	Config         localAutoDevLoopConfig `json:"config"`
	Status         string                 `json:"status"`
	CurrentAttempt int                    `json:"currentAttempt"`
	StartTime      int64                  `json:"startTime"`
	LastOutput     string                 `json:"lastOutput"`
}

type localAutoDevManager struct {
	mu         sync.Mutex
	active     map[string]*localAutoDevLoop
	nextID     int
	workDir    string
}

func newLocalAutoDevManager(workDir string) *localAutoDevManager {
	return &localAutoDevManager{
		active:  make(map[string]*localAutoDevLoop),
		nextID:  1,
		workDir: workDir,
	}
}

func (m *localAutoDevManager) startLoop(config localAutoDevLoopConfig) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("loop-%d", m.nextID)
	m.nextID++
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	loop := &localAutoDevLoop{
		ID:             id,
		Config:         config,
		Status:         "running",
		CurrentAttempt: 0,
		StartTime:      time.Now().UnixMilli(),
	}
	m.active[id] = loop

	go m.runLoop(id)

	return id
}

func (m *localAutoDevManager) cancelLoop(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	loop, ok := m.active[id]
	if ok && loop.Status == "running" {
		loop.Status = "cancelled"
		return true
	}
	return false
}

func (m *localAutoDevManager) getLoops() []*localAutoDevLoop {
	m.mu.Lock()
	defer m.mu.Unlock()

	loops := make([]*localAutoDevLoop, 0, len(m.active))
	for _, l := range m.active {
		loops = append(loops, l)
	}
	return loops
}

func (m *localAutoDevManager) getLoop(id string) (*localAutoDevLoop, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, ok := m.active[id]
	return l, ok
}

func (m *localAutoDevManager) clearCompleted() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for id, l := range m.active {
		if l.Status != "running" {
			delete(m.active, id)
			count++
		}
	}
	return count
}

func (m *localAutoDevManager) runLoop(id string) {
	for {
		m.mu.Lock()
		loop, ok := m.active[id]
		if !ok || loop.Status != "running" {
			m.mu.Unlock()
			return
		}

		if loop.CurrentAttempt >= loop.Config.MaxAttempts {
			loop.Status = "failed"
			m.mu.Unlock()
			return
		}

		loop.CurrentAttempt++
		config := loop.Config
		m.mu.Unlock()

		output, success := m.executeAttempt(config)

		m.mu.Lock()
		loop, ok = m.active[id]
		if !ok {
			m.mu.Unlock()
			return
		}

		loop.LastOutput = output
		if success {
			loop.Status = "success"
			m.mu.Unlock()
			return
		}

		if loop.Status != "running" {
			m.mu.Unlock()
			return
		}

		attempt := loop.CurrentAttempt
		m.mu.Unlock()

		if attempt >= config.MaxAttempts {
			m.mu.Lock()
			if loop, ok := m.active[id]; ok {
				loop.Status = "failed"
			}
			m.mu.Unlock()
			return
		}

		// Wait before retry
		delay := time.Duration(1<<uint(attempt-1)) * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		time.Sleep(delay)
	}
}

func (m *localAutoDevManager) executeAttempt(config localAutoDevLoopConfig) (string, bool) {
	command := strings.TrimSpace(config.Command)
	if command == "" {
		command = m.defaultCommand(config)
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "empty command", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = m.workDir
	output, err := cmd.CombinedOutput()

	if err == nil {
		return string(output), true
	}

	return string(output) + "\nError: " + err.Error(), false
}

func (m *localAutoDevManager) defaultCommand(config localAutoDevLoopConfig) string {
	switch config.Type {
	case "test":
		if strings.TrimSpace(config.Target) != "" {
			return "npx vitest run " + config.Target
		}
		return "npm test"
	case "lint":
		if strings.TrimSpace(config.Target) != "" {
			return "npx eslint --fix " + config.Target
		}
		return "npm run lint -- --fix"
	case "build":
		return "npm run build"
	default:
		return "npm test"
	}
}
