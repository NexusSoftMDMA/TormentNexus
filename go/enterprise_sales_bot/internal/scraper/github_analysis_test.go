package scraper

import (
	"testing"
)

func TestGenerateInsightSummary_WithLanguage(t *testing.T) {
	analysis := &RepoAnalysis{
		CompanyName:     "testcorp",
		ReposAnalyzed:   5,
		PrimaryLanguage: "Go",
		Languages:       map[string]int{"Go": 10000, "JavaScript": 5000},
		StarsTotal:      500,
		Topics:          []string{"ai", "infrastructure"},
		Bottlenecks:     []string{"high open issues"},
	}

	ga := &GitHubAnalyzer{}
	summary := ga.generateInsightSummary(analysis)

	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	if !contains(summary, "primarily Go") {
		t.Errorf("expected summary to mention primarily Go, got: %s", summary)
	}
	if !contains(summary, "1 bottleneck") {
		t.Errorf("expected summary to mention bottleneck count, got: %s", summary)
	}
}

func TestGenerateInsightSummary_Empty(t *testing.T) {
	analysis := &RepoAnalysis{
		CompanyName:   "minimal",
		ReposAnalyzed: 2,
	}

	ga := &GitHubAnalyzer{}
	summary := ga.generateInsightSummary(analysis)

	if !contains(summary, "2 repos") {
		t.Errorf("expected summary to mention repo count, got: %s", summary)
	}
}

func TestGetRecentActivityDays(t *testing.T) {
	ga := &GitHubAnalyzer{}

	tests := []struct {
		input string
		want  int
	}{
		{"5 days ago", 5},
		{"30 days ago", 30},
		{"365 days ago", 365},
		{"", 99999},
		{"just now", 0},
	}

	for _, tc := range tests {
		got := ga.getRecentActivityDays(tc.input)
		if got != tc.want {
			t.Errorf("getRecentActivityDays(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Error("min(1,2) should be 1")
	}
	if min(5, 5) != 5 {
		t.Error("min(5,5) should be 5")
	}
	if min(-1, 0) != -1 {
		t.Error("min(-1,0) should be -1")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
