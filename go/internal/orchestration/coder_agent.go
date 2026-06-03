package orchestration

/**
 * @file coder_agent.go
 * @module go/internal/orchestration
 *
 * WHAT: Go-native implementation of a Coder Agent.
 * Handles coding tasks via A2A protocol and local file system access.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type CoderAgent struct {
	ID         string
	Broker     *A2ABroker
	Workspace  string
	InMessages chan A2AMessage
}

func NewCoderAgent(broker *A2ABroker, workspace string) *CoderAgent {
	id := "go-coder"
	return &CoderAgent{
		ID:         id,
		Broker:     broker,
		Workspace:  workspace,
		InMessages: broker.RegisterAgent(id),
	}
}

func (a *CoderAgent) Start(ctx context.Context) {
	fmt.Printf("[Go Coder] 🤖 Agent started (Workspace: %s)\n", a.Workspace)
	go a.heartbeatLoop(ctx)
	go a.loop(ctx)
}

func (a *CoderAgent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.Broker.RouteMessage(A2AMessage{
				ID:        fmt.Sprintf("hb-%s-%d", a.ID, nowMillis()),
				Timestamp: nowMillis(),
				Sender:    a.ID,
				Type:      Heartbeat,
				Payload:   map[string]interface{}{},
			})
		}
	}
}

func (a *CoderAgent) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-a.InMessages:
			if !ok {
				return
			}
			if msg.Type == TaskRequest {
				a.handleTaskRequest(ctx, msg)
			}
			if msg.Type == TaskNegotiation {
				a.handleTaskNegotiation(ctx, msg)
			}
		}
	}
}

func (a *CoderAgent) handleTaskNegotiation(ctx context.Context, msg A2AMessage) {
	fmt.Printf("[Go Coder] 🤝 Bidding on task: %s\n", msg.Payload)

	a.Broker.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("a2a-bid-%d", nowMillis()),
		Timestamp: nowMillis(),
		Sender:    a.ID,
		Recipient: msg.Sender,
		Type:      CapabilityReport,
		ReplyTo:   msg.ID,
		Payload: map[string]interface{}{
			"agentId":            a.ID,
			"capabilities":       []string{"coding", "refactoring"},
			"canHandle":          true,
			"estimatedLatencyMs": 5000,
			"reasoning":          "Ready to implement Go-native code changes.",
		},
	})
}

func (a *CoderAgent) handleTaskRequest(ctx context.Context, msg A2AMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	// Handle Capability Report
	if msg.Type == StateUpdate && payload["action"] == "REPORT_CAPABILITIES" {
		a.Broker.RouteMessage(A2AMessage{
			ID:        fmt.Sprintf("a2a-cap-%d", nowMillis()),
			Timestamp: nowMillis(),
			Sender:    a.ID,
			Type:      StateUpdate,
			Payload: map[string]interface{}{
				"capabilities": []string{"coding", "refactoring", "filesystem"},
				"role":         "Go Native Coder",
			},
		})
		return
	}

	task, _ := payload["task"].(string)
	if task == "" {
		return
	}

	fmt.Printf("[Go Coder] 🧠 Processing task: %s\n", task)

	// 1. Plan and Generate Code
	prompt := fmt.Sprintf(`
		You are a Go-native Coder Agent.
		Task: %s

		Return JSON with:
		{
		  "filename": "string",
		  "content": "string",
		  "reasoning": "string"
		}
	`, task)

	resp, err := ai.AutoRoute(ctx, []ai.Message{
		{Role: "system", Content: "You are an expert software engineer."},
		{Role: "user", Content: prompt},
	})

	if err != nil {
		a.sendError(msg, err.Error())
		return
	}

	// 2. Parse and Write
	var plan struct {
		Filename  string `json:"filename"`
		Content   string `json:"content"`
		Reasoning string `json:"reasoning"`
	}

	// Extract JSON
	jsonStr := resp.Content
	if start := strings.Index(jsonStr, "{"); start != -1 {
		if end := strings.LastIndex(jsonStr, "}"); end != -1 {
			jsonStr = jsonStr[start : end+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		a.sendError(msg, "Failed to parse LLM plan: "+err.Error())
		return
	}

	targetPath := filepath.Join(a.Workspace, plan.Filename)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		a.sendError(msg, "FS Error: "+err.Error())
		return
	}

	if err := os.WriteFile(targetPath, []byte(plan.Content), 0644); err != nil {
		a.sendError(msg, "FS Error: "+err.Error())
		return
	}

	fmt.Printf("[Go Coder] 💾 Wrote to %s\n", targetPath)

	// 3. Respond
	a.Broker.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("a2a-%d", nowMillis()),
		Timestamp: nowMillis(),
		Sender:    a.ID,
		Recipient: msg.Sender,
		Type:      TaskResponse,
		ReplyTo:   msg.ID,
		Payload: map[string]interface{}{
			"status":       "completed",
			"filesChanged": []string{plan.Filename},
			"reasoning":    plan.Reasoning,
		},
	})
}

func (a *CoderAgent) sendError(orig A2AMessage, errMsg string) {
	a.Broker.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("a2a-%d", nowMillis()),
		Timestamp: nowMillis(),
		Sender:    a.ID,
		Recipient: orig.Sender,
		Type:      TaskResponse,
		ReplyTo:   orig.ID,
		Payload: map[string]interface{}{
			"status": "failed",
			"error":  errMsg,
		},
	})
}
