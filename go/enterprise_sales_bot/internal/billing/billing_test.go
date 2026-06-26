package billing

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestStripeBillingClient_Mock(t *testing.T) {
	// Since real Stripe API calls require a valid key and network,
	// we use the New method and verify the interface implementation.
	client := NewStripeBillingClient("sk_test_123")
	if client == nil {
		t.Fatal("Failed to create StripeBillingClient")
	}
}

func TestMockBillingClient(t *testing.T) {
	client := &MockBillingClient{}
	id, err := client.CreateInvoice(context.Background(), db.Deal{ID: 1}, db.Company{Name: "Test"})
	if err != nil {
		t.Fatalf("Mock CreateInvoice failed: %v", err)
	}
	if id != "INV-MOCK-123" {
		t.Errorf("Expected INV-MOCK-123, got %s", id)
	}
}
