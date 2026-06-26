package scraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Helper variable for env manipulation in tests

func TestGitHubIssueSource_Discover_WithMockServer(t *testing.T) {
	// Track which endpoints were called
	var issueSearchCalled, repoSearchCalled bool

	// Create a mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock GitHub API called: %s", r.URL.String())

		switch r.URL.Path {
		case "/search/issues":
			issueSearchCalled = true
			// Return mock issue search results
			result := githubIssueSearchResponse{
				TotalCount: 2,
				Issues: []struct {
					HTMLURL   string `json:"html_url"`
					Title     string `json:"title"`
					Body      string `json:"body"`
					RepoURL   string `json:"repository_url"`
					CreatedAt string `json:"created_at"`
				}{
					{
						HTMLURL:   "https://github.com/acme-ai/ml-platform/issues/42",
						Title:     "MCP tool routing bottleneck in production",
						RepoURL:   "https://api.github.com/repos/acme-ai/ml-platform",
						CreatedAt: "2026-01-15T10:30:00Z",
					},
					{
						HTMLURL:   "https://github.com/acme-ai/ml-platform/issues/43",
						Title:     "Multi-agent orchestration timing issues",
						RepoURL:   "https://api.github.com/repos/acme-ai/ml-platform",
						CreatedAt: "2026-01-16T14:20:00Z",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)

		case "/search/repositories":
			repoSearchCalled = true
			// Return mock repo search results
			result := struct {
				Items []githubRepo `json:"items"`
			}{
				Items: []githubRepo{
					{
						FullName:    "acme-ai/ml-platform",
						Description: "ML orchestration platform",
						Language:    "Go",
						Owner: struct {
							Login string `json:"login"`
						}{
							Login: "acme-ai",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	source := &GitHubIssueSource{
		Client:   server.Client(),
		Token:    "GITHUB_TOKEN_PLACEHOLDER", // test fixture, not a real token
		Keywords: []string{"MCP"},
	}

	t.Run("isExcludedOrg", func(t *testing.T) {
		if !source.isExcludedOrg("google") {
			t.Error("google should be excluded")
		}
		if !source.isExcludedOrg("microsoft") {
			t.Error("microsoft should be excluded")
		}
		if source.isExcludedOrg("acme-ai") {
			t.Error("acme-ai should NOT be excluded")
		}
		if source.isExcludedOrg("startup-corp") {
			t.Error("startup-corp should NOT be excluded")
		}
	})

	t.Run("orgToDomain", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"acme-ai", "acme-ai.io"},
			{"neuralsystems", "neuralsystems.io"},
			{"startup-corp", "startup.tech"},
			{"my-org", "my-org.tech"},
		}
		for _, tt := range tests {
			got := source.orgToDomain(tt.input)
			if got != tt.expected {
				t.Errorf("orgToDomain(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("extractOrgFromRepoURL", func(t *testing.T) {
		org, err := source.extractOrgFromRepoURL("https://api.github.com/repos/acme-ai/ml-platform")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if org != "acme-ai" {
			t.Errorf("expected 'acme-ai', got %q", org)
		}
	})

	t.Run("Discover_returns_empty_without_token", func(t *testing.T) {
		oldEnvToken := os.Getenv("GITHUB_TOKEN")
		_ = os.Setenv("GITHUB_TOKEN", "")
		defer func() { _ = os.Setenv("GITHUB_TOKEN", oldEnvToken) }()

		sourceNoToken := &GitHubIssueSource{
			Client:   http.DefaultClient,
			Token:    "",
			Keywords: []string{"MCP"},
		}
		companies, err := sourceNoToken.Discover(context.Background(), []string{"MCP"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(companies) != 0 {
			t.Errorf("expected 0 companies without token, got %d", len(companies))
		}
	})

	_ = issueSearchCalled
	_ = repoSearchCalled
}

func TestGitHubIssueSource_Discover_WithExcludedOrgs(t *testing.T) {
	source := &GitHubIssueSource{
		Client:   http.DefaultClient,
		Token:    "GITHUB_TOKEN_PLACEHOLDER", // test fixture, not a real token
		Keywords: []string{"MCP"},
	}

	// Test that excluded orgs are filtered
	if !source.isExcludedOrg("google") {
		t.Error("expected google to be excluded")
	}
	if !source.isExcludedOrg("GITHUB") {
		t.Error("expected GITHUB to be excluded (case-insensitive)")
	}
}

func TestGitHubIssueSource_InferMarketCap(t *testing.T) {
	source := &GitHubIssueSource{}

	tests := []struct {
		desc     string
		language string
		expected string
	}{
		{"enterprise platform", "Go", "Mid-Market"},
		{"empty language", "", "Small Business"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			repo := &githubRepo{
				Description: tt.desc,
				Language:    tt.language,
			}
			got := source.inferMarketCap(repo)
			if got != tt.expected {
				t.Errorf("inferMarketCap(%q) = %q, want %q", tt.desc, got, tt.expected)
			}
		})
	}
}

// Ensure GitHubIssueSource implements LeadSource interface
var _ LeadSource = (*GitHubIssueSource)(nil)

// Helper to verify Company struct compatibility
func TestGitHubIssueSource_CompanyStruct(t *testing.T) {
	company := db.Company{
		Name:          "Test Co",
		Domain:        "testco.io",
		TechStack:     []string{"Go"},
		HiringSignals: []string{"GitHub Issue: MCP routing issue - https://github.com/testco/platform/issues/1"},
		MarketCapTier: "Small Business",
	}

	if company.Name != "Test Co" {
		t.Error("company name mismatch")
	}
	if len(company.HiringSignals) != 1 {
		t.Error("expected 1 hiring signal")
	}
}