package orchestration

import (
	"context"
	"github.com/borghq/borg-go/internal/session"
	"github.com/borghq/borg-go/internal/controlplane"
)

type FleetManagerPlus struct {
	*session.FleetManager
	observer *TrafficObserver
}

func NewFleetManagerPlus(vault controlplane.MemoryVault, bus interface {
	EmitEvent(eventType string, source string, payload interface{})
}) *FleetManagerPlus {
	return &FleetManagerPlus{
		FleetManager: session.NewFleetManager(),
		observer:     NewTrafficObserver(vault, bus),
	}
}

func (f *FleetManagerPlus) ProcessSignal(ctx context.Context, msg A2AMessage) {
	f.observer.Observe(ctx, msg)
}
