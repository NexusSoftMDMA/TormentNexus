package billing

import (
	"context"
	"fmt"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// InvoiceStatus represents the current state of a deal's billing.
type InvoiceStatus string

const (
	InvoicePending InvoiceStatus = "Pending"
	InvoiceSent    InvoiceStatus = "Sent"
	InvoicePaid    InvoiceStatus = "Paid"
	InvoiceFailed  InvoiceStatus = "Failed"
)

// BillingClient defines the interface for interacting with financial systems (e.g., Stripe).
type BillingClient interface {
	// CreateInvoice generates a new billing record for a won deal.
	CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error)

	// GetInvoiceStatus retrieves the payment status of an invoice.
	GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error)
}

// MockBillingClient provides a simulated implementation for billing tasks.
type MockBillingClient struct{}

func (m *MockBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "INV-MOCK-123", nil
}

func (m *MockBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	return InvoicePending, nil
}

// StripeBillingClient implements BillingClient using the Stripe API.
type StripeBillingClient struct {
	APIKey string
}

// NewStripeBillingClient creates a new Stripe-based billing client.
func NewStripeBillingClient(apiKey string) *StripeBillingClient {
	return &StripeBillingClient{APIKey: apiKey}
}

func (s *StripeBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	stripe.Key = s.APIKey

	params := &stripe.InvoiceParams{
		Customer: stripe.String(company.Domain), // Simplified: in reality, would map to Stripe Customer ID
		AutoAdvance: stripe.Bool(true),
		CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodSendInvoice)),
		DaysUntilDue: stripe.Int64(30),
	}

	// Add line item for the deal
	// Note: In v81, line items are typically managed via InvoiceItem or Price APIs.
	// For this integration, we simulate the high-level orchestration.

	inv, err := invoice.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe invoice creation failed: %w", err)
	}

	return inv.ID, nil
}

func (s *StripeBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	stripe.Key = s.APIKey

	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return InvoiceFailed, fmt.Errorf("stripe invoice retrieval failed: %w", err)
	}

	switch inv.Status {
	case stripe.InvoiceStatusPaid:
		return InvoicePaid, nil
	case stripe.InvoiceStatusOpen:
		return InvoiceSent, nil
	case stripe.InvoiceStatusVoid, stripe.InvoiceStatusUncollectible:
		return InvoiceFailed, nil
	default:
		return InvoicePending, nil
	}
}
