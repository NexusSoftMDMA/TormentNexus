package orchestration

/**
 * @file director.go
 * @module go/internal/orchestration
 *
 * WHAT: Go-native implementation of the Director loop.
 * Coordinates autonomous task execution using the SwarmController and CoderAgent.
 */

import (
	"context"
	"fmt"
	"time"
)

type Director struct {
	swarm  *SwarmController
	coder  *CoderAgent
	broker *A2ABroker
}

func NewDirector(swarm *SwarmController, coder *CoderAgent, broker *A2ABroker) *Director {
	return &Director{
		swarm:  swarm,
		coder:  coder,
		broker: broker,
	}
}

func (d *Director) StartAutonomousTask(ctx context.Context, goal string) error {
	fmt.Printf("[Go Director] 🎬 Starting autonomous task: %s\n", goal)

	// 1. Negotiate Task via Handshake (Multi-turn A2A)
	fmt.Println("[Go Director] 🤝 Negotiating with local agents...")

	negID := fmt.Sprintf("neg-%d", nowMillis())
	queryMsg := A2AMessage{
		ID:        negID,
		Timestamp: nowMillis(),
		Sender:    "DIRECTOR",
		Type:      TaskNegotiation,
		Payload:   map[string]interface{}{"task": goal},
	}

	// Request with timeout
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := d.broker.Query(queryCtx, queryMsg)
	if err == nil {
		fmt.Printf("[Go Director] ✅ Negotiation successful! Picked agent: %s\n", resp.Sender)
	} else {
		fmt.Printf("[Go Director] ⚠️ Negotiation timed out or failed: %v. Using default fallback.\n", err)
	}

	// 2. Run a Swarm session to plan and review
	fmt.Println("[Go Director] 🧠 Convening Swarm for planning...")
	swarmResult, err := d.swarm.StartSession(ctx, goal, SwarmSessionConfig{
		MaxTurns:            3,
		CompletionThreshold: 0.8,
		AutoRotate:          true,
	})
	if err != nil {
		return fmt.Errorf("swarm planning failed: %w", err)
	}

	if !swarmResult.Success {
		fmt.Println("[Go Director] ⚠️ Swarm did not reach full consensus, but proceeding with current plan.")
	}

	// 3. Delegate implementation to the selected agent (or default Go Coder)
	recipient := d.coder.ID
	if resp.Sender != "" {
		recipient = resp.Sender
	}

	fmt.Printf("[Go Director] 🤖 Delegating implementation to %s...\n", recipient)
	
	taskID := fmt.Sprintf("task-%d", time.Now().Unix())
	msg := A2AMessage{
		ID:        fmt.Sprintf("a2a-%d", nowMillis()),
		Timestamp: nowMillis(),
		Sender:    "DIRECTOR",
		Recipient: recipient,
		Type:      TaskRequest,
		Payload: map[string]interface{}{
			"task":   goal,
			"plan":   swarmResult.Transcript,
			"taskId": taskID,
		},
	}

	d.broker.RouteMessage(msg)

	return nil
}
