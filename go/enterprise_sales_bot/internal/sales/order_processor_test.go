package sales

import (
	"context"
	"errors"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Mock Billing Client for detailed testing
type errorBillingClient struct {
	billing.MockBillingClient
}

func (e *errorBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "", errors.New("billing service down")
}

// Mock CRM Client for detailed testing
type errorCRMClient struct {
	crm.MockCRMClient
}

func (e *errorCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	return errors.New("crm service down")
}

// Mock objects for testing
type mockDB struct {
	company *db.Company
	err     error
}

func (m *mockDB) GetCompanyByID(ctx context.Context, id int64) (*db.Company, error) {
	return m.company, m.err
}

func TestProcessOrder_BillingFailure(t *testing.T) {
	billingClient := &errorBillingClient{}
	crmClient := &crm.MockCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err == nil {
		t.Fatal("Expected billing failure error, got nil")
	}
	if err.Error() != "failed to create invoice: billing service down" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestProcessOrder_Success(t *testing.T) {
	billingClient := &billing.MockBillingClient{}
	crmClient := &crm.MockCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err != nil {
		t.Fatalf("ProcessOrder failed: %v", err)
	}
}

func TestProcessOrder_CRMWarning(t *testing.T) {
	// CRM sync failure should only log a warning and not return an error in the current implementation.
	billingClient := &billing.MockBillingClient{}
	crmClient := &errorCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err != nil {
		t.Fatalf("ProcessOrder should not fail on CRM error, but got: %v", err)
	}
}
