package orchestration

import (
	"context"
)

type FleetSignalProcessor interface {
	ProcessSignal(ctx context.Context, msg A2AMessage)
}
