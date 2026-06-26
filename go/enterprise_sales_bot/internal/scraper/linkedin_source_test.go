package scraper

import (
	"context"
	"testing"
)

func TestLinkedInSource_Discover_Simulation(t *testing.T) {
	source := &LinkedInSource{
		Client:   nil,
		Username: "", // No credentials = simulation mode
		Password: "",
	}

	companies, err := source.Discover(context.Background(), []string{"AI", "ML"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(companies) != 3 {
		t.Errorf("expected 3 simulated companies, got %d", len(companies))
	}

	// Verify company structure
	for _, c := range companies {
		if c.Name == "" {
			t.Error("company name should not be empty")
		}
		if c.Domain == "" {
			t.Error("company domain should not be empty")
		}
		if len(c.TechStack) == 0 {
			t.Error("company tech stack should not be empty")
		}
		if len(c.HiringSignals) == 0 {
			t.Error("company hiring signals should not be empty")
		}
		if c.MarketCapTier == "" {
			t.Error("company market cap tier should not be empty")
		}
	}
}

func TestLinkedInSource_HealthCheck(t *testing.T) {
	t.Run("fails_without_credentials", func(t *testing.T) {
		source := &LinkedInSource{
			Username: "",
			Password: "",
		}
		err := source.HealthCheck(context.Background())
		if err == nil {
			t.Error("expected error without credentials")
		}
	})

	t.Run("fails_with_empty_credentials", func(t *testing.T) {
		source := &LinkedInSource{
			Username: "  ",
			Password: "  ",
		}
		err := source.HealthCheck(context.Background())
		if err == nil {
			t.Error("expected error with empty credentials")
		}
	})

	t.Run("passes_with_credentials", func(t *testing.T) {
		source := &LinkedInSource{
			Username: "test-user@example.com",
			Password: "test-value-for-testing-only", // test fixture only, not a real credential
		}
		err := source.HealthCheck(context.Background())
		if err != nil {
			t.Errorf("unexpected error with valid credentials: %v", err)
		}
	})
}

func TestLinkedInSource_SetCredentials(t *testing.T) {
	source := &LinkedInSource{}
	source.SetCredentials("test-user@example.com", "test-value-for-testing-only")

	if source.Username != "test-user@example.com" {
		t.Errorf("expected username 'test-user@example.com', got %q", source.Username)
	}
	if source.Password != "test-value-for-testing-only" {
		t.Errorf("expected password 'test-value-for-testing-only', got %q", source.Password)
	}
}

func TestLinkedInSource_SetTargetTitles(t *testing.T) {
	source := &LinkedInSource{}
	titles := []string{"CEO", "CTO", "VP Engineering"}
	source.SetTargetTitles(titles)

	if len(source.TargetTitles) != 3 {
		t.Errorf("expected 3 titles, got %d", len(source.TargetTitles))
	}
	if source.TargetTitles[0] != "CEO" {
		t.Errorf("expected first title 'CEO', got %q", source.TargetTitles[0])
	}
}

// Ensure LinkedInSource implements LeadSource interface
var _ LeadSource = (*LinkedInSource)(nil)
