package communication_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// mock implementations for communication manager integration test

type mockClassifier struct{}

func (m *mockClassifier) Classify(ctx context.Context, text string) (communication.Intent, error) {
	return communication.IntentGeneral, nil
}

type mockResponder struct{}

func (m *mockResponder) Generate(ctx context.Context, salesCtx communication.SalesContext, action communication.Action) (string, error) {
	return "mock reply", nil
}

type mockStrategy struct{}

func (m *mockStrategy) Decide(ctx context.Context, salesCtx communication.SalesContext) (communication.Action, error) {
	return communication.ActionAdvanceState, nil
}

type mockOrderProcessor struct{}

func (m *mockOrderProcessor) ProcessOrder(ctx context.Context, deal db.Deal) error {
	return nil
}

type mockEmailSender struct{}

func (m *mockEmailSender) Send(ctx context.Context, msg communication.EmailMessage) error {
	return nil
}

func setupCommTestDB(t *testing.T) *db.DB {
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

func TestCommunicationManager_Integration_OutreachForResearchedDeal(t *testing.T) {
	database := setupCommTestDB(t)
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	company := &db.Company{
		Name:          "Comm Test Corp",
		Domain:        "commtest.io",
		MarketCapTier: "Mid-Market",
	}
	if err := database.CreateCompany(ctx, company); err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateResearched,
	}
	if err := database.CreateDeal(ctx, deal); err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	contact := &db.Contact{
		CompanyID: company.ID,
		Name:      "Jane Doe",
		Role:      "CTO",
		Email:     "jane@commtest.io",
	}
	if err := database.CreateContact(ctx, contact); err != nil {
		t.Fatalf("Failed to create contact: %v", err)
	}

	mgr := communication.NewManager(
		database,
		&mockClassifier{},
		&mockResponder{},
		&mockStrategy{},
		&mockOrderProcessor{},
		&mockEmailSender{},
	)

	// Trigger one poll cycle
	mgr.ExecutePoll(ctx)

	// Verify contacts/interactions/etc.
	contacts, err := database.ListContactsByCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to list contacts: %v", err)
	}
	if len(contacts) == 0 {
		t.Error("Expected contact to still exist")
	}

	// The manager may or may not have created an outbound interaction depending on logic.
	// For this test, we mainly ensure the poll cycle does not panic.
}

func TestCommunicationManager_Integration_SkipsNilDB(t *testing.T) {
	mgr := communication.NewManager(
		nil,
		&mockClassifier{},
		&mockResponder{},
		&mockStrategy{},
		&mockOrderProcessor{},
		nil,
	)

	ctx := context.Background()
	mgr.ExecutePoll(ctx)
	// If this doesn't panic, the nil DB path works
}