package orchestration

import (
	"context"
	"github.com/borghq/borg-go/internal/session"
	"github.com/borghq/borg-go/internal/controlplane"
	"github.com/borghq/borg-go/internal/supervisor"
)

type FleetManagerPlus struct {
	*session.FleetManager
	supervisor *supervisor.Manager
	observer   *TrafficObserver
}

func NewFleetManagerPlus(vault controlplane.MemoryVault, bus any, sup *supervisor.Manager) *FleetManagerPlus {
    // We use 'any' and cast it to our internal interface inside NewTrafficObserver if needed,
    // or just pass it as is if it matches.
    // For now, I'll use a local interface to match eventbus.EventBus.
	return &FleetManagerPlus{
		FleetManager: session.NewFleetManager(),
		supervisor:   sup,
		observer:     NewTrafficObserver(vault, bus.(interface {
            EmitEvent(string, string, any)
        })),
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
