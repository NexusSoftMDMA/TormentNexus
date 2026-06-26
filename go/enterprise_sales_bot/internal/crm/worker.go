package crm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Worker coordinates the synchronization between the local database and the external CRM.
type Worker struct {
	db	*db.DB
	client	CRMClient
}

// NewWorker creates a new CRM synchronization worker.
func NewWorker(database *db.DB, client CRMClient) *Worker {
	return &Worker{
		db:	database,
		client:	client,
	}
}

// Run starts the periodic CRM synchronization process.
func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("CRM Worker: Synchronization started (interval: %v)...", interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("CRM Worker: Synchronization stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			w.sync(ctx)
		}
	}
}

// ExecuteSync manually triggers a sync cycle (exported for testing).
func (w *Worker) ExecuteSync(ctx context.Context) {
	w.sync(ctx)
}

func (w *Worker) sync(ctx context.Context) {
	slog.Info("CRM Worker: Executing sync cycle...")

	if w.db == nil {
		slog.Info("CRM Worker: DB unavailable, skipping sync cycle")
		return
	}

	// 1. Reconcile updates from CRM
	updates, err := w.client.GetLeadUpdates(ctx)
	if err != nil {
		slog.Info(fmt.Sprintf("CRM Worker: Error fetching updates: %v", err))
	} else {
		for _, update := range updates {
			slog.Info(fmt.Sprintf("CRM Worker: Reconciling update for lead %s to %s", update.ID, update.NewState))
			// In a real system, we'd map external IDs. For now, we assume numeric mapping.
			var dealID int64
			if _, err := fmt.Sscanf(update.ID, "%d", &dealID); err == nil {
				if err := w.db.UpdateDealState(ctx, dealID, update.NewState); err != nil {
					slog.Info(fmt.Sprintf("CRM Worker: Failed to update deal %d: %v", dealID, err))
				}
			}
		}
	}

	// 2. Push local Negotiating/Closed deals to CRM
	// This is a simplified implementation
	deals, err := w.db.ListDealsByState(ctx, db.StateNegotiating)
	if err != nil {
		slog.Info(fmt.Sprintf("CRM Worker: Error listing negotiating deals: %v", err))
		return
	}

	for _, deal := range deals {
		company, _ := w.db.GetCompanyByID(ctx, deal.CompanyID)
		if company != nil {
			// 2a. Push local updates to CRM
			if err := w.client.PushDeal(ctx, deal, *company, "WorkerSync"); err != nil {
				slog.Info(fmt.Sprintf("CRM Worker: Error pushing deal %d: %v", deal.ID, err))
			}

			// 2b. Pull latest details from CRM to keep local state synchronized
			details, err := w.client.FetchDealDetails(ctx, deal.ID)
			if err != nil {
				slog.Info(fmt.Sprintf("CRM Worker: Error fetching deal details for %d: %v", deal.ID, err))
				continue
			}

			if details != nil {
				slog.Info(fmt.Sprintf("CRM Worker: Synchronizing details for deal %d from CRM", deal.ID))
				// Update local deal pricing and requirements if they differ
				if err := w.db.UpdateDealDetails(ctx, deal.ID, details.QuotedPricing, details.CustomRequirements); err != nil {
					slog.Info(fmt.Sprintf("CRM Worker: Error updating local deal %d: %v", deal.ID, err))
				}
			}
		}
	}
}
