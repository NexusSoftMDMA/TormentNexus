package enrichment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Enricher worker started...")

	// Run immediately on startup
	e.executeEnrichment(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Enricher worker stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			e.executeEnrichment(ctx)
		}
	}
}

// ExecuteEnrichment manually triggers an enrichment cycle (exported for testing).
func (e *Enricher) ExecuteEnrichment(ctx context.Context) {
	e.executeEnrichment(ctx)
}

func (e *Enricher) executeEnrichment(ctx context.Context) {
	if e.db == nil {
		// DB not available – skip enrichment to avoid panic in dev/test environments
		slog.Info("Enricher: DB unavailable, skipping enrichment cycle")
		return
	}
	// 1. Find deals in Discovered state
	deals, err := e.db.ListDealsByState(ctx, db.StateDiscovered)
	if err != nil {
		slog.Info(fmt.Sprintf("Enricher: Error listing discovered deals: %v", err))
		return
	}

	for _, deal := range deals {
		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error getting company %d: %v", deal.CompanyID, err))
			continue
		}

		err = e.enrichCompany(ctx, deal, *company)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error enriching company %s: %v", company.Name, err))
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) error {
	for _, source := range e.sources {
		contacts, err := source.Enrich(ctx, company)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error from source: %v", err))
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
				slog.Info(fmt.Sprintf("Enricher: Error persisting contact %s: %v", contact.Name, err))
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return fmt.Errorf("failed to update deal state: %w", err)
			}

			// Synchronize newly found contacts with the CRM (with retry logic)
			if e.crmClient != nil {
				go func() {
					maxRetries := 3
					for i := 0; i < maxRetries; i++ {
						if err := e.crmClient.SyncContacts(ctx, company.ID, contacts); err != nil {
							slog.Info(fmt.Sprintf("Enricher Warning: Failed to sync contacts to CRM (attempt %d/%d): %v", i+1, maxRetries, err))
							time.Sleep(time.Duration(i+1) * 2 * time.Second)
							continue
						}
						return
					}
					slog.Info(fmt.Sprintf("Enricher Error: CRM contact synchronization failed after %d attempts for company %d", maxRetries, company.ID))
				}()
			}

			slog.Info(fmt.Sprintf("Enricher: Successfully enriched %s with %d contacts", company.Name, len(contacts)))
			return nil
		}
	}
	return nil
}

// MockApolloSource is a simulated enrichment source.
type MockApolloSource struct{}

func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	slog.Info(fmt.Sprintf("MockApolloSource: Searching for contacts at %s", company.Domain))

	// Simulate finding contacts based on domain
	switch company.Domain {
	case "aidynamics.com":
		return []db.Contact{
			{
				Name:          "Sarah Chen",
				Role:          "Director of AI",
				Email:         "sarah.chen@aidynamics.com",
				GitHubHandle:  "schen-ai",
			},
		}, nil
	case "neuralsystems.io":
		return []db.Contact{
			{
				Name:          "James Wilson",
				Role:          "Principal Systems Architect",
				Email:         "j.wilson@neuralsystems.io",
				GitHubHandle:  "jwilson-sys",
			},
		}, nil
	}

	return nil, nil
}
