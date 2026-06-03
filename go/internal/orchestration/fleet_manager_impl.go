package orchestration

import (
	"context"

	"github.com/tormentnexushq/tormentnexus-go/internal/session"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	"github.com/tormentnexushq/tormentnexus-go/internal/supervisor"
)

type FleetManagerPlus struct {
	*session.FleetManager
	supervisor *supervisor.Manager
	observer   *TrafficObserver
}

func NewFleetManagerPlus(vault controlplane.MemoryVault, bus any, sup *supervisor.Manager) *FleetManagerPlus {
	// bus is expected to satisfy the string-based EmitEvent interface
	// (e.g., httpapi.eventBusAdapter wrapping *eventbus.EventBus)
	var observerBus interface {
		EmitEvent(eventType string, source string, payload interface{})
	}
	if b, ok := bus.(interface {
		EmitEvent(eventType string, source string, payload interface{})
	}); ok {
		observerBus = b
	}

	return &FleetManagerPlus{
		FleetManager: session.NewFleetManager(),
		supervisor:   sup,
		observer:     NewTrafficObserver(vault, observerBus),
	}
}

func (f *FleetManagerPlus) ProcessSignal(ctx context.Context, msg A2AMessage) {
	f.observer.Observe(ctx, msg)
}

func (f *FleetManagerPlus) GetFleetStatus() []*session.FleetMember {
	sessions := f.supervisor.ListSessions()
	for _, s := range sessions {
		if s.PID > 0 {
			f.FleetManager.Register(s.ID, s.PID)
		}
	}
	return f.FleetManager.GetFleetStatus()
}
