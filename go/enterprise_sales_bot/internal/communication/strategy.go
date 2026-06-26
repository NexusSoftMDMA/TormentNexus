package communication

import (
	"context"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Action represents a decided path of action by the sales engine.
type Action string

const (
	ActionRespond     Action = "Respond"
	ActionEscalate    Action = "Escalate"
	ActionAdvanceState Action = "AdvanceState"
	ActionWait        Action = "Wait"
)

// SalesContext encapsulates the data needed for the sales engine to make decisions.
type SalesContext struct {
	Company      db.Company
	Deal         db.Deal
	Contact      db.Contact
	Interactions []db.Interaction
	LatestIntent Intent
}

// SalesStrategy defines the interface for the autonomous sales workflow engine.
type SalesStrategy interface {
	// Decide determines the next action based on the current sales context.
	Decide(ctx context.Context, salesCtx SalesContext) (Action, error)
}
