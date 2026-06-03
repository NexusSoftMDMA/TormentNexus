package orchestration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type TrafficObserver struct {
	vault controlplane.MemoryVault
	bus   interface {
		EmitEvent(eventType string, source string, payload interface{})
	}
}

func NewTrafficObserver(vault controlplane.MemoryVault, bus interface {
	EmitEvent(eventType string, source string, payload interface{})
}) *TrafficObserver {
	return &TrafficObserver{vault: vault, bus: bus}
}

func (o *TrafficObserver) Observe(ctx context.Context, msg A2AMessage) {
	if msg.Type != TaskResponse && msg.Type != Critique {
		return
	}

	go func() {
		fact, err := o.extractFact(ctx, msg)
		if err != nil || fact == "" {
			return
		}

		entry := controlplane.L2VaultRecord{
			ID:            fmt.Sprintf("fact-%d", time.Now().UnixNano()),
			SessionID:     msg.Sender,
			Type:          controlplane.MemoryLongTerm,
			Content:       fact,
			Importance:    0.8,
			HeatScore:     100.0,
			LastAccessedAt: time.Now(),
			CreatedAt:     time.Now(),
		}

		_ = o.vault.Commit(ctx, entry)

		if o.bus != nil {
			o.bus.EmitEvent("memory:fact_discovered", "TrafficObserver", map[string]interface{}{
				"fact":    fact,
				"source":  msg.Sender,
				"context": msg.Type,
			})
		}
	}()
}

func (o *TrafficObserver) extractFact(ctx context.Context, msg A2AMessage) (string, error) {
	content := fmt.Sprintf("%v", msg.Payload)
	if len(content) < 50 {
		return "", nil
	}

	prompt := fmt.Sprintf("Extract a single high-value technical fact or lesson learned from this message. Return ONLY the fact or empty string if none found: %s", content)
	resp, err := ai.AutoRoute(ctx, []ai.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}
