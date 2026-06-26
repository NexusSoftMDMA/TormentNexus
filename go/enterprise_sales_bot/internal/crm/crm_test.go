package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestRestCRMClient_PushDeal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deals" {
			t.Errorf("Expected to POST to /deals, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	err := client.PushDeal(context.Background(), db.Deal{ID: 1}, db.Company{Name: "TestCorp"}, "test-route")
	if err != nil {
		t.Fatalf("PushDeal failed: %v", err)
	}
}

func TestRestCRMClient_GetLeadUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"ID": "123", "NewState": "Negotiating"}]`))
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	updates, err := client.GetLeadUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetLeadUpdates failed: %v", err)
	}

	if len(updates) != 1 || updates[0].ID != "123" {
		t.Errorf("Unexpected updates: %+v", updates)
	}
}

func TestRestCRMClient_SyncContacts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/companies/1/contacts" {
			t.Errorf("Expected to POST to /companies/1/contacts, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	err := client.SyncContacts(context.Background(), 1, []db.Contact{{Name: "Test Contact"}})
	if err != nil {
		t.Fatalf("SyncContacts failed: %v", err)
	}
}
