package scraper

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestMockJobBoardSource_Discover(t *testing.T) {
	source := &MockJobBoardSource{}
	keywords := []string{"AI Engineer"}
	companies, err := source.Discover(context.Background(), keywords)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(companies) == 0 {
		t.Error("Expected companies to be discovered, got none")
	}

	found := false
	for _, c := range companies {
		if c.Name == "AI Dynamics Corp" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find AI Dynamics Corp, but it was missing")
	}
}

// Note: Testing Scraper.executeDiscovery would require a mock/test database.
// For now, we verify the interface and discovery logic.
func TestScraper_Initialization(t *testing.T) {
	database := &db.DB{} // Empty DB for init test
	sources := []LeadSource{&MockJobBoardSource{}}
	s := NewScraper(database, sources)

	if s == nil {
		t.Fatal("Expected scraper instance, got nil")
	}

	if len(s.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(s.sources))
	}
}
