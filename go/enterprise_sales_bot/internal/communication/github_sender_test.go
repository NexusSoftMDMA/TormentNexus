package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// --- CalculateRelevance tests ---

func TestCalculateRelevance_HighValueKeywordsInTitle(t *testing.T) {
	score := CalculateRelevance("MCP", "MCP tool routing bottleneck", "some body text")
	if score < 3 {
		t.Errorf("expected score >= 3 for high-value keyword in title, got %d", score)
	}
}

func TestCalculateRelevance_HighValueKeywordsInBody(t *testing.T) {
	score := CalculateRelevance("agent", "some generic title", "we need agent orchestration for our LLM pipeline")
	if score < 2 {
		t.Errorf("expected score >= 2 for high-value keywords in body, got %d", score)
	}
}

func TestCalculateRelevance_MediumValueKeywords(t *testing.T) {
	score := CalculateRelevance("infra", "CI/CD pipeline automation", "deploy workflow integration")
	if score < 2 {
		t.Errorf("expected score >= 2 for medium-value keywords, got %d", score)
	}
}

func TestCalculateRelevance_NoRelevantKeywords(t *testing.T) {
	score := CalculateRelevance("unrelated", "fix CSS alignment", "button color is wrong")
	if score != 0 {
		t.Errorf("expected score 0 for irrelevant content, got %d", score)
	}
}

func TestCalculateRelevance_MultipleHighValueMatches(t *testing.T) {
	score := CalculateRelevance("MCP", "MCP agent orchestration with LLM tool routing", "agent workflow for model context protocol")
	// Title: MCP (+3), agent (+3), orchestrat (+3), tool routing (+3), LLM (+3) = 15
	// Body: agent (+1), model context protocol (+1) = 2
	// Total should be high
	if score < 10 {
		t.Errorf("expected score >= 10 for multiple high-value matches, got %d", score)
	}
}

// --- GenerateTechHookComment tests ---

func TestGenerateTechHookComment_ContainsIssueTitle(t *testing.T) {
	issue := IssueTarget{
		Owner:       "acme-ai",
		Repo:        "ml-platform",
		IssueNumber: 42,
		Title:       "MCP tool routing bottleneck",
		URL:         "https://github.com/acme-ai/ml-platform/issues/42",
		Relevance:   10,
	}

	comment := GenerateTechHookComment(issue)

	if comment == "" {
		t.Fatal("expected non-empty comment")
	}

	// Comment should mention the issue title
	if !testContainsStr(comment, issue.Title) {
		t.Errorf("expected comment to contain issue title %q", issue.Title)
	}

	// Comment should mention HyperNexus
	if !testContainsStr(comment, "HyperNexus") {
		t.Error("expected comment to mention HyperNexus")
	}

	// Comment should mention key selling points
	if !testContainsStr(comment, "Progressive MCP Tool Routing") {
		t.Error("expected comment to mention Progressive MCP Tool Routing")
	}
	if !testContainsStr(comment, "Cross-Harness Tool Parity") {
		t.Error("expected comment to mention Cross-Harness Tool Parity")
	}
	if !testContainsStr(comment, "LLM Waterfall") {
		t.Error("expected comment to mention LLM Waterfall")
	}
}

// --- extractOrgFromDomain tests ---

func TestExtractOrgFromDomain_SimpleDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"acme.io", "acme"},
		{"deepmind.tech", "deepmind"},
		{"openai.com", "openai"},
		{"example.co.uk", "example"},
	}

	for _, tt := range tests {
		got := extractOrgFromDomain(tt.domain)
		if got != tt.expected {
			t.Errorf("extractOrgFromDomain(%q) = %q, want %q", tt.domain, got, tt.expected)
		}
	}
}

func TestExtractOrgFromDomain_WithURL(t *testing.T) {
	got := extractOrgFromDomain("https://acme.io/about")
	if got != "acme" {
		t.Errorf("extractOrgFromDomain with URL = %q, want %q", got, "acme")
	}
}

func TestExtractOrgFromDomain_SinglePart(t *testing.T) {
	got := extractOrgFromDomain("localhost")
	if got != "localhost" {
		t.Errorf("extractOrgFromDomain(%q) = %q, want %q", "localhost", got, "localhost")
	}
}

// --- IssueTarget struct tests ---

func TestIssueTarget_Fields(t *testing.T) {
	target := IssueTarget{
		Owner:       "acme-ai",
		Repo:        "ml-platform",
		IssueNumber: 42,
		Title:       "Tool routing issue",
		URL:         "https://github.com/acme-ai/ml-platform/issues/42",
		Relevance:   5,
	}

	if target.Owner != "acme-ai" {
		t.Errorf("expected owner 'acme-ai', got %q", target.Owner)
	}
	if target.IssueNumber != 42 {
		t.Errorf("expected issue number 42, got %d", target.IssueNumber)
	}
	if target.Relevance != 5 {
		t.Errorf("expected relevance 5, got %d", target.Relevance)
	}
}

// --- GitHubCommentSender construction tests ---

func TestNewGitHubCommentSender_WithoutToken(t *testing.T) {
	// Ensure GITHUB_TOKEN is unset for this test
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_BOT_USERNAME", "")

	sender := NewGitHubCommentSender("robertpelloni/enterprise_sales_bot")
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}
	if sender.repo != "robertpelloni/enterprise_sales_bot" {
		t.Errorf("expected repo to be set, got %q", sender.repo)
	}
	if sender.username != "hypernexus-bot" {
		t.Errorf("expected default username 'hypernexus-bot', got %q", sender.username)
	}
}

func TestGitHubCommentSender_SendComment_NilClient(t *testing.T) {
	sender := &GitHubCommentSender{
		client: nil,
		repo:   "test/repo",
	}

	err := sender.SendComment(context.Background(), "owner", "repo", 1, "test comment")
	if err == nil {
		t.Error("expected error when client is nil")
	}
}

func TestGitHubCommentSender_SearchRelevantIssues_NilClient(t *testing.T) {
	sender := &GitHubCommentSender{
		client: nil,
		repo:   "test/repo",
	}

	_, err := sender.SearchRelevantIssues(context.Background(), "acme.io")
	if err == nil {
		t.Error("expected error when client is nil")
	}
}

func TestGitHubCommentSender_FindAndComment_NilClient(t *testing.T) {
	sender := &GitHubCommentSender{
		client: nil,
		repo:   "test/repo",
	}

	company := db.Company{Domain: "acme.io", Name: "Acme"}
	contact := db.Contact{Email: "test@acme.io"}

	err := sender.FindAndComment(context.Background(), company, contact)
	if err == nil {
		t.Error("expected error when client is nil")
	}
}

// --- helpers ---

func testContainsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
