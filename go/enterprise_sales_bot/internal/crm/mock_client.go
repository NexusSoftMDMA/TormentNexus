package crm

import (
	"context"
	"log/slog"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"fmt"
)

// MockCRMClient simulates an external CRM integration.
type MockCRMClient struct {
	PushDealCalled		bool
	SyncInteractionCalled	bool
	SyncContactsCalled	bool
	GetLeadUpdatesCalled	bool
	LatestNote		string
	UpdatesToReturn		[]LeadUpdate
}

// NewMockCRMClient creates a new mock CRM client.
func NewMockCRMClient() *MockCRMClient {
	return &MockCRMClient{}
}

func (m *MockCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	slog.Info(fmt.Sprintf("CRM: Pushing deal %d for company %s (Route: %s) to CRM", deal.ID, company.Name, route))
	m.PushDealCalled = true
	return nil
}

func (m *MockCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	slog.Info("CRM: Fetching updates from external CRM...")
	m.GetLeadUpdatesCalled = true
	if m.UpdatesToReturn != nil {
		return m.UpdatesToReturn, nil
	}
	return []LeadUpdate{}, nil
}

func (m *MockCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	slog.Info(fmt.Sprintf("CRM: Validating account for domain: %s", domain))
	return true, nil
}

func (m *MockCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	slog.Info(fmt.Sprintf("CRM: Syncing interaction for deal %d: %s", dealID, note))
	m.SyncInteractionCalled = true
	m.LatestNote = note
	return nil
}

func (m *MockCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	slog.Info(fmt.Sprintf("CRM: Syncing %d contacts for company %d", len(contacts), companyID))
	m.SyncContactsCalled = true
	return nil
}

func (m *MockCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	slog.Info(fmt.Sprintf("CRM: Fetching mock details for deal %d", dealID))
	return &DealDetails{
		ID:			dealID,
		Status:			db.StateNegotiating,
		QuotedPricing:		10000.0,
		CustomRequirements:	"Mock Requirement",
	}, nil
}
