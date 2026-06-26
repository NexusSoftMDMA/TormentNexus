package enrichment_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
)

func setupTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()
	// Run migrations
	if err := database.RunMigrations(ctx); err != nil {
		_ = database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func cleanupDeal(t *testing.T, database *db.DB, companyName string) {
	ctx := context.Background()
	deals, _ := database.ListRecentDeals(ctx, 100)
	for _, deal := range deals {
		company, _ := database.GetCompanyByID(ctx, deal.CompanyID)
		if company != nil && company.Name == companyName {
			_ = database.UpdateDealState(ctx, deal.ID, db.StateClosedLost)
		}
	}
}

func TestEnricher_Integration_EnrichesDiscoveredDeals(t *testing.T) {
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	// Create a discovered company that the mock source will match
	company := &db.Company{
		Name:          "AI Dynamics Corp",
		Domain:        "aidynamics.com",
		MarketCapTier: "Mid-Market",
	}
	err := database.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:   company.ID,
		CurrentState: db.StateDiscovered,
	}
	err = database.CreateDeal(ctx, deal)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	defer cleanupDeal(t, database, company.Name)

	// Create enricher with mock source
	mockSource := &enrichment.MockApolloSource{}
	mockCRM := crm.NewMockCRMClient()
	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{mockSource}, mockCRM)

	// Run enrichment once
	e.ExecuteEnrichment(ctx)

	// Verify deal moved to Researched
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get updated deal: %v", err)
	}
	if updatedDeal.CurrentState != db.StateResearched {
		t.Errorf("Expected deal state to be Researched, got %s", updatedDeal.CurrentState)
	}

	// Verify contacts were created
	contacts, err := database.ListContactsByCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to list contacts: %v", err)
	}
	if len(contacts) == 0 {
		t.Error("Expected contacts to be created during enrichment")
	}

	// Verify the contact has the expected details from mock source
	found := false
	for _, contact := range contacts {
		if contact.Name == "Sarah Chen" && contact.Email == "sarah.chen@aidynamics.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Sarah Chen contact with correct email")
	}
}

func TestEnricher_Integration_SkipsNilDB(t *testing.T) {
	mockSource := &enrichment.MockApolloSource{}
	mockCRM := crm.NewMockCRMClient()
	e := enrichment.NewEnricher(nil, []enrichment.EnrichmentSource{mockSource}, mockCRM)

	ctx := context.Background()
	e.ExecuteEnrichment(ctx)
	// If this doesn't panic and return cleanly, the nil check works
}
