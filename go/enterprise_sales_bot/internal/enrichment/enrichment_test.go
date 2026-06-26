package enrichment

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestMockApolloSource_Enrich(t *testing.T) {
	source := &MockApolloSource{}
	company := db.Company{
		Name:   "AI Dynamics Corp",
		Domain: "aidynamics.com",
	}

	contacts, err := source.Enrich(context.Background(), company)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	if contacts[0].Name != "Sarah Chen" {
		t.Errorf("Expected Sarah Chen, got %s", contacts[0].Name)
	}
}

func TestApolloSource_Enrich(t *testing.T) {
	apiKey := os.Getenv("APOLLO_API_KEY")
	if apiKey == "" {
		t.Skip("APOLLO_API_KEY not set, skipping integration test")
	}

	source := NewApolloSource(apiKey)
	company := db.Company{
		Name:   "OpenAI",
		Domain: "openai.com",
	}

	contacts, err := source.Enrich(context.Background(), company)
	if err != nil {
		// Apollo free plan doesn't include People Search API
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "API_INACCESSIBLE") {
			t.Skip("Apollo free plan blocks People Search: " + err.Error())
		}
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(contacts) == 0 {
		t.Log("No contacts returned from Apollo.io (may be rate limited or no matches)")
	} else {
		t.Logf("Found %d contacts:", len(contacts))
		for _, c := range contacts {
			t.Logf("  - %s (%s) <%s>", c.Name, c.Role, c.Email)
		}
	}
}

func TestApolloSource_HealthCheck(t *testing.T) {
	apiKey := os.Getenv("APOLLO_API_KEY")
	if apiKey == "" {
		t.Skip("APOLLO_API_KEY not set, skipping integration test")
	}

	source := NewApolloSource(apiKey)
	err := source.HealthCheck(context.Background())
	if err != nil {
		// Apollo free plan doesn't include People Search API
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "API_INACCESSIBLE") {
			t.Skip("Apollo free plan blocks People Search: " + err.Error())
		}
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestEnricher_Initialization(t *testing.T) {
	database := &db.DB{}
	sources := []EnrichmentSource{&MockApolloSource{}}
	e := NewEnricher(database, sources, nil)

	if e == nil {
		t.Fatal("Expected enricher instance, got nil")
	}

	if len(e.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(e.sources))
	}
}
// mockSourceForTest is a mock EnrichmentSource for unit testing.
type mockSourceForTest struct {
	name           string
	shouldFail     bool
	returnContacts int
}

func (m *mockSourceForTest) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("%s: simulated error", m.name)
	}
	if m.returnContacts == 0 {
		return nil, nil
	}
	contacts := make([]db.Contact, m.returnContacts)
	for i := 0; i < m.returnContacts; i++ {
		contacts[i] = db.Contact{
			Name:  fmt.Sprintf("%s Contact %d", company.Name, i+1),
			Email: fmt.Sprintf("contact%d@%s", i+1, strings.ToLower(strings.ReplaceAll(company.Domain, ".", "-"))),
			Role:  "Test Role",
		}
	}
	return contacts, nil
}

func TestFallbackSource_CustomSources(t *testing.T) {
	tests := []struct {
		name             string
		sources          []EnrichmentSource
		sourceNames      []string
		expectedSuccess  bool
		expectedContacts int
		checkError       bool
	}{
		{
			name: "first source succeeds",
			sources: []EnrichmentSource{
				&mockSourceForTest{name: "First", returnContacts: 1},
				&mockSourceForTest{name: "Second", returnContacts: 2},
			},
			sourceNames:      []string{"First", "Second"},
			expectedSuccess:  true,
			expectedContacts: 1,
		},
		{
			name: "first fails, second succeeds",
			sources: []EnrichmentSource{
				&mockSourceForTest{name: "First", shouldFail: true},
				&mockSourceForTest{name: "Second", returnContacts: 2},
			},
			sourceNames:      []string{"First", "Second"},
			expectedSuccess:  true,
			expectedContacts: 2,
		},
		{
			name: "first returns empty, second succeeds",
			sources: []EnrichmentSource{
				&mockSourceForTest{name: "First", returnContacts: 0},
				&mockSourceForTest{name: "Second", returnContacts: 3},
			},
			sourceNames:      []string{"First", "Second"},
			expectedSuccess:  true,
			expectedContacts: 3,
		},
		{
			name: "all sources fail",
			sources: []EnrichmentSource{
				&mockSourceForTest{name: "First", shouldFail: true},
				&mockSourceForTest{name: "Second", shouldFail: true},
				&mockSourceForTest{name: "Third", shouldFail: true},
			},
			sourceNames:      []string{"First", "Second", "Third"},
			expectedSuccess:  false,
			expectedContacts: 0,
		},
		{
			name: "all return empty",
			sources: []EnrichmentSource{
				&mockSourceForTest{name: "First", returnContacts: 0},
				&mockSourceForTest{name: "Second", returnContacts: 0},
			},
			sourceNames:      []string{"First", "Second"},
			expectedSuccess:  false,
			expectedContacts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallback := NewFallbackSource(tt.sources, tt.sourceNames)
			company := db.Company{Name: "Test Corp", Domain: "test.com"}

			contacts, err := fallback.Enrich(context.Background(), company)
			if tt.checkError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if (!tt.checkError || tt.expectedSuccess) && err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tt.expectedSuccess && len(contacts) == 0 {
				t.Fatal("Expected contacts, got none")
			}

			if len(contacts) != tt.expectedContacts {
				t.Errorf("Expected %d contacts, got %d", tt.expectedContacts, len(contacts))
			}
		})
	}
}

func TestFallbackSource_Status(t *testing.T) {
	sources := []EnrichmentSource{
		&mockSourceForTest{name: "Source1", returnContacts: 1},
		&mockSourceForTest{name: "Source2", returnContacts: 2},
	}
	names := []string{"Primary", "Backup"}

	fallback := NewFallbackSource(sources, names)
	status := fallback.Status()

	if !strings.Contains(status, "2 source(s)") {
		t.Errorf("Status should mention 2 sources, got: %s", status)
	}
	if !strings.Contains(status, "Primary") || !strings.Contains(status, "Backup") {
		t.Errorf("Status should contain source names, got: %s", status)
	}
}

func TestFallbackSource_Names(t *testing.T) {
	sources := []EnrichmentSource{
		&mockSourceForTest{name: "First", returnContacts: 1},
	}

	// Test with names
	names := []string{"MySource"}
	fallback := NewFallbackSource(sources, names)
	if len(fallback.Names()) != 1 || fallback.Names()[0] != "MySource" {
		t.Errorf("Expected custom name, got %v", fallback.Names())
	}

	// Test without names (should generate defaults)
	fallbackNoNames := NewFallbackSource(sources, nil)
	if len(fallbackNoNames.Names()) != 1 || fallbackNoNames.Names()[0] != "Source1" {
		t.Errorf("Expected Source1, got %v", fallbackNoNames.Names())
	}

	// Test with insufficient names
	namesPartial := []string{"OnlyOne"}
	fallbackPartial := NewFallbackSource(append(sources, &mockSourceForTest{name: "Second", returnContacts: 1}), namesPartial)
	if len(fallbackPartial.Names()) != 2 {
		t.Errorf("Expected 2 names, got %d", len(fallbackPartial.Names()))
	}
}
