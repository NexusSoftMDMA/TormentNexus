package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestRestCRMClient_FetchDealDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 123,
			"status": "Negotiating",
			"quoted_pricing": 50000.0,
			"custom_requirements": "Custom SLA"
		}`))
	}))
	defer ts.Close()

	client := NewRestCRMClient(ts.URL, "test-api-key")
	details, err := client.FetchDealDetails(context.Background(), 123)

	if err != nil {
		t.Fatalf("FetchDealDetails failed: %v", err)
	}

	if details.ID != 123 {
		t.Errorf("Expected ID 123, got %d", details.ID)
	}
	if details.Status != db.StateNegotiating {
		t.Errorf("Expected status Negotiating, got %s", details.Status)
	}
	if details.QuotedPricing != 50000.0 {
		t.Errorf("Expected pricing 50000.0, got %f", details.QuotedPricing)
	}
}

func TestMockCRMClient_FetchDealDetails(t *testing.T) {
	client := NewMockCRMClient()
	details, err := client.FetchDealDetails(context.Background(), 456)

	if err != nil {
		t.Fatalf("Mock FetchDealDetails failed: %v", err)
	}

	if details.ID != 456 {
		t.Errorf("Expected ID 456, got %d", details.ID)
	}
}
