package crm_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func setupCRMTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		_ = database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestCRMWorker_Integration_SyncsNegotiatingDeals(t *testing.T) {
	database := setupCRMTestDB(t)
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	// Create a company and deal in Negotiating state
	company := &db.Company{
		Name:          "CRM Test Corp",
		Domain:        "crmtest.io",
		MarketCapTier: "Mid-Market",
	}
	err := database.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateNegotiating,
	}
	err = database.CreateDeal(ctx, deal)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Create CRM worker with mock CRM client
	mockCRM := crm.NewMockCRMClient()
	worker := crm.NewWorker(database, mockCRM)

	// Run sync cycle
	worker.ExecuteSync(ctx)

	// Verify no panic occurred and deal still exists
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get deal: %v", err)
	}
	if updatedDeal == nil {
		t.Error("Expected deal to still exist after sync")
	}
}

func TestCRMWorker_Integration_SkipsNilDB(t *testing.T) {
	mockCRM := crm.NewMockCRMClient()
	worker := crm.NewWorker(nil, mockCRM)

	ctx := context.Background()
	worker.ExecuteSync(ctx)
	// If this doesn't panic and returns cleanly, the nil check works
}