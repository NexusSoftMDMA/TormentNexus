package orchestration

/**
 * @file a2a_logger.go
 * @module go/internal/orchestration
 *
 * WHAT: Go-native traffic logging and auditing for Agent-to-Agent (A2A) communication.
 * Captures all signals for session history and observability.
 *
 * WHY: Multi-agent coordination can be opaque. Persistent logging ensures 
 * that agent decisions, task handoffs, and consensus votes are diagnosable.
 */

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type A2ALogger struct {
	mu         sync.Mutex
	logPath    string
	buffer     []A2AMessage
	flushRate  time.Duration
}

func NewA2ALogger(workspaceRoot string) *A2ALogger {
	path := filepath.Join(workspaceRoot, ".tormentnexus", "logs", "a2a_traffic.jsonl")
	l := &A2ALogger{
		logPath:   path,
		buffer:    make([]A2AMessage, 0),
		flushRate: 5 * time.Second,
	}
	l.initialize()
	return l
}

func (l *A2ALogger) initialize() {
	dir := filepath.Dir(l.logPath)
	_ = os.MkdirAll(dir, 0755)

	go l.flushLoop()
}

func (l *A2ALogger) LogMessage(msg A2AMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer = append(l.buffer, msg)
}

func (l *A2ALogger) flushLoop() {
	ticker := time.NewTicker(l.flushRate)
	defer ticker.Stop()

	for range ticker.C {
		l.flush()
	}
}

func (l *A2ALogger) flush() {
	l.mu.Lock()
	if len(l.buffer) == 0 {
		l.mu.Unlock()
		return
	}
	msgs := l.buffer
	l.buffer = make([]A2AMessage, 0)
	l.mu.Unlock()

	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[Go A2A Logger] Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	for _, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		f.Write(data)
		f.Write([]byte("\n"))
	}
}

func (l *A2ALogger) GetRecentLogs(limit int) ([]A2AMessage, error) {
	f, err := os.Open(l.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []A2AMessage{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var logs []A2AMessage
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var msg A2AMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
			logs = append(logs, msg)
		}
	}

	if len(logs) > limit {
		return logs[len(logs)-limit:], nil
	}
	return logs, nil
}
