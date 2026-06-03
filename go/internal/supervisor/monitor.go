package supervisor

/**
 * @file monitor.go
 * @module go/internal/supervisor
 *
 * WHAT: Go-native conversation monitor and tool predictor loop.
 * Watches session activity and triggers preemptive tool suggestions.
 */

import (
	"context"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
)

type ConversationMonitor struct {
	manager    *Manager
	predictor  *mcp.ToolPredictor
	interval   time.Duration
	cancelFunc context.CancelFunc
}

func NewConversationMonitor(manager *Manager, predictor *mcp.ToolPredictor) *ConversationMonitor {
	return &ConversationMonitor{
		manager:   manager,
		predictor: predictor,
		interval:  5 * time.Minute,
	}
}

func (m *ConversationMonitor) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	go m.loop(runCtx)
}

func (m *ConversationMonitor) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

func (m *ConversationMonitor) loop(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Initial delay
	time.Sleep(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runPrediction(ctx)
		}
	}
}

func (m *ConversationMonitor) runPrediction(ctx context.Context) {
	sessions := m.manager.ListSessions()
	for _, session := range sessions {
		if session.State != StateRunning {
			continue
		}

		// Extract recent logs as chat history
		logs, err := m.manager.GetSessionLogs(session.ID, 20)
		if err != nil {
			continue
		}

		history := ""
		for _, entry := range logs {
			history += fmt.Sprintf("[%s]: %s\n", entry.Stream, entry.Message)
		}

		if history == "" {
			continue
		}

		goal := ""
		if g, ok := session.Metadata["activeGoal"].(string); ok {
			goal = g
		}

		predicted, err := m.predictor.PredictAndPreload(ctx, history, goal)
		if err != nil {
			fmt.Printf("[Go Monitor] Prediction failed for session %s: %v\n", session.ID, err)
			continue
		}

		if len(predicted) > 0 {
			fmt.Printf("[Go Monitor] Predicted tools for session %s: %v\n", session.ID, predicted)
			// In a real implementation, we would update the session's advertised tools
			// or notify the control plane.
			if session.Metadata == nil {
				session.Metadata = make(map[string]any)
			}
			session.Metadata["predictedTools"] = predicted
			m.manager.mu.Lock()
			if s, ok := m.manager.sessions[session.ID]; ok {
				s.Metadata["predictedTools"] = predicted
			}
			m.manager.mu.Unlock()
		}
	}
}
