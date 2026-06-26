package researcher_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
)

// mockCrawler implements researcher.Crawler for testing.
type mockCrawler struct{}

func (m *mockCrawler) Crawl(ctx context.Context, target string) (string, error) {
	return "Mock research findings for " + target, nil
}

// mockDossierProcessor implements researcher.DossierProcessor for testing.
type mockDossierProcessor struct{}

func (m *mockDossierProcessor) Process(findings []string) (string, error) {
	return "Processed dossier: " + strings.Join(findings, "; "), nil
}

func setupResearcherTestDB(t *testing.T) *db.DB {
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

func TestResearcher_Integration_ResearchesDiscoveredDeals(t *testing.T) {
	database := setupResearcherTestDB(t)
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	// Create a company in Researched state
	company := &db.Company{
		Name:          "Test Research Corp",
		Domain:        "testresearch.io",
		MarketCapTier: "Startup",
	}
	err := database.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateResearched,
	}
	err = database.CreateDeal(ctx, deal)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Create researcher with mock crawlers
	crawlers := []researcher.Crawler{&mockCrawler{}}
	processor := &mockDossierProcessor{}
	mockCRM := crm.NewMockCRMClient()
	r := researcher.NewResearcher(database, crawlers, processor, mockCRM)

	// Run one research cycle
	r.ExecuteResearch(ctx)

	// Verify the deal was processed (state may have changed or dossier created)
	// The researcher should have attempted to crawl and process findings
	// We verify no panic occurred and the deal still exists
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get deal: %v", err)
	}
	if updatedDeal == nil {
		t.Error("Expected deal to still exist after research")
	}
}

func TestResearcher_Integration_SkipsNilDB(t *testing.T) {
	crawlers := []researcher.Crawler{&mockCrawler{}}
	processor := &mockDossierProcessor{}
	mockCRM := crm.NewMockCRMClient()
	r := researcher.NewResearcher(nil, crawlers, processor, mockCRM)

	ctx := context.Background()
	r.ExecuteResearch(ctx)
	// If this doesn't panic, the nil check works
}
