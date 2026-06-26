package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// RepoAnalysis represents the technical analysis of a company's repository.
type RepoAnalysis struct {
	CompanyName	string		`json:"company_name"`
	ReposAnalyzed	int		`json:"repos_analyzed"`
	PrimaryLanguage	string		`json:"primary_language"`
	Languages	map[string]int	`json:"languages"`	// language -> bytes
	Topics		[]string	`json:"topics"`
	StarsTotal	int		`json:"stars_total"`
	ForksTotal	int		`json:"forks_total"`
	RecentActivity	string		`json:"recent_activity"`	// days since last commit
	OpenIssues	int		`json:"open_issues"`
	HasTests	bool		`json:"has_tests"`
	HasCI		bool		`json:"has_ci"`
	HasDocs		bool		`json:"has_docs"`
	HasDockerfile	bool		`json:"has_dockerfile"`
	Bottlenecks	[]string	`json:"bottlenecks"`		// detected problem areas
	TechStack	[]string	`json:"tech_stack"`		// detected technologies
	InsightSummary	string		`json:"insight_summary"`	// one-line hook from analysis
}

// GitHubAnalyzer analyzes repositories for technical insights.
type GitHubAnalyzer struct {
	client	*github.Client
	}

// NewGitHubAnalyzer creates a new GitHub analyzer.
// Token can be empty for unauthenticated access (lower rate limits).
func NewGitHubAnalyzer(ctx context.Context, token string) *GitHubAnalyzer {
	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}
	return &GitHubAnalyzer{client: client}
}

// AnalyzeOrg analyzes all public repos for a GitHub organization.
func (ga *GitHubAnalyzer) AnalyzeOrg(ctx context.Context, orgName string) (*RepoAnalysis, error) {
	analysis := &RepoAnalysis{
		CompanyName:	orgName,
		Languages:	make(map[string]int),
		TechStack:	make([]string, 0),
		Bottlenecks:	make([]string, 0),
	}

	// List repos for the org
	page := 1
	reposAnalyzed := 0
	for {
		repos, resp, err := ga.client.Repositories.ListByOrg(ctx, orgName, &github.RepositoryListByOrgOptions{
			Type:		"public",
			Sort:		"updated",
			ListOptions:	github.ListOptions{Page: page, PerPage: 30},
		})
		if err != nil {
			return nil, fmt.Errorf("listing org repos: %w", err)
		}
		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			ga.analyzeRepo(ctx, repo, analysis)
		}

		reposAnalyzed += len(repos)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	analysis.ReposAnalyzed = reposAnalyzed

	// Generate insight summary
	analysis.InsightSummary = ga.generateInsightSummary(analysis)

	return analysis, nil
}

// AnalyzeUser analyzes all public repos for a GitHub user.
func (ga *GitHubAnalyzer) AnalyzeUser(ctx context.Context, username string) (*RepoAnalysis, error) {
	analysis := &RepoAnalysis{
		CompanyName:	username,
		Languages:	make(map[string]int),
		TechStack:	make([]string, 0),
		Bottlenecks:	make([]string, 0),
	}

	opt := &github.RepositoryListByUserOptions{
		Type:		"public",
		Sort:		"updated",
		ListOptions:	github.ListOptions{PerPage: 30},
	}

	for {
		repos, resp, err := ga.client.Repositories.ListByUser(ctx, username, opt)
		if err != nil {
			return nil, fmt.Errorf("listing user repos: %w", err)
		}
		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			ga.analyzeRepo(ctx, repo, analysis)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	analysis.ReposAnalyzed = len(ga.listReposForUser(ctx, username))
	analysis.InsightSummary = ga.generateInsightSummary(analysis)

	return analysis, nil
}

// listReposForUser returns all public repos for a user (internal helper).
func (ga *GitHubAnalyzer) listReposForUser(ctx context.Context, username string) []string {
	var names []string
	opt := &github.RepositoryListByUserOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		repos, resp, err := ga.client.Repositories.ListByUser(ctx, username, opt)
		if err != nil {
			return names
		}
		for _, r := range repos {
			names = append(names, r.GetName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return names
}

// analyzeRepo extracts technical details from a single repository.
func (ga *GitHubAnalyzer) analyzeRepo(ctx context.Context, repo *github.Repository, analysis *RepoAnalysis) {
	name := repo.GetName()
	slog.Info(fmt.Sprintf("GitHubAnalyzer: analyzing %s/%s", analysis.CompanyName, name))

	// Languages
	langs, _, err := ga.client.Repositories.ListLanguages(ctx, analysis.CompanyName, name)
	if err == nil {
		for lang, bytes := range langs {
			analysis.Languages[lang] += bytes
			if bytes > 0 {
				analysis.TechStack = append(analysis.TechStack, lang)
			}
		}
	}

	// Primary language
	if repo.GetLanguage() != "" {
		if analysis.PrimaryLanguage == "" {
			analysis.PrimaryLanguage = repo.GetLanguage()
		}
		analysis.TechStack = append(analysis.TechStack, repo.GetLanguage())
	}

	// Topics
	analysis.Topics = append(analysis.Topics, repo.Topics...)

	// Stars & forks
	analysis.StarsTotal += repo.GetStargazersCount()
	analysis.ForksTotal += repo.GetForksCount()

	// Recent activity
	updated := repo.GetUpdatedAt().Time
	daysSinceUpdate := int(time.Since(updated).Hours() / 24)
	if daysSinceUpdate < ga.getRecentActivityDays(analysis.RecentActivity) || analysis.RecentActivity == "" {
		analysis.RecentActivity = fmt.Sprintf("%d days ago", daysSinceUpdate)
	}

	// Open issues
	analysis.OpenIssues += repo.GetOpenIssuesCount()

	// Detect infrastructure patterns
	topicsStr := strings.Join(repo.Topics, ",")
	desc := strings.ToLower(repo.GetDescription())

	if strings.Contains(topicsStr, "ci") || strings.Contains(topicsStr, "circleci") || strings.Contains(topicsStr, "github-actions") {
		analysis.HasCI = true
	}
	if strings.Contains(topicsStr, "docker") || strings.Contains(desc, "docker") {
		analysis.HasDockerfile = true
	}
	if strings.Contains(topicsStr, "docs") || strings.Contains(desc, "documentation") {
		analysis.HasDocs = true
	}
	if strings.Contains(topicsStr, "test") || strings.Contains(desc, "test") || strings.Contains(desc, "testing") {
		analysis.HasTests = true
	}

	// Check for common files
	fileContent, _, _, err := ga.client.Repositories.GetContents(ctx, analysis.CompanyName, name, "Dockerfile", nil)
	if err == nil && fileContent != nil {
		analysis.HasDockerfile = true
	}

	// Technology detection from description and topics
	techKeywords := map[string][]string{
		"AI/ML":		{"machine-learning", "deep-learning", "neural", "pytorch", "tensorflow", "ai", "ml"},
		"LLM":			{"llm", "gpt", "language-model", "transformer", "rag", "chain", "agent"},
		"Kubernetes":		{"kubernetes", "k8s", "helm", "istio"},
		"Cloud/Native":		{"aws", "gcp", "azure", "cloud", "serverless"},
		"Database":		{"postgresql", "postgres", "mysql", "mongodb", "redis", "sqlite"},
		"DevOps":		{"devops", "ci/cd", "terraform", "ansible", "prometheus", "grafana"},
		"Go":			{"go", "golang"},
		"TypeScript/JS":	{"typescript", "javascript", "node", "react", "angular", "vue"},
		"Rust":			{"rust", "wasm"},
		"Python":		{"python", "django", "flask", "fastapi"},
		"MCP/Protocol":		{"mcp", "model-context-protocol", "tool-calling"},
		"Orchestration":	{"orchestrator", "workflow", "pipeline", "scheduler"},
	}

	detected := make(map[string]bool)
	for tech, keywords := range techKeywords {
		for _, kw := range keywords {
			if strings.Contains(topicsStr, kw) || strings.Contains(desc, kw) {
				if !detected[tech] {
					analysis.TechStack = append(analysis.TechStack, tech)
					detected[tech] = true
				}
				break
			}
		}
	}

	// Detect bottlenecks from repo state
	if repo.GetOpenIssuesCount() > 50 {
		analysis.Bottlenecks = append(analysis.Bottlenecks, fmt.Sprintf("High open issue count (%d) in %s", repo.GetOpenIssuesCount(), name))
	}
	if daysSinceUpdate > 90 && repo.GetStargazersCount() > 100 {
		analysis.Bottlenecks = append(analysis.Bottlenecks, fmt.Sprintf("%s: popular but inactive (%d stars, %d days since update)", name, repo.GetStargazersCount(), daysSinceUpdate))
	}
}

// getRecentActivityDays parses the recent activity string and returns days.
func (ga *GitHubAnalyzer) getRecentActivityDays(activity string) int {
	if activity == "" {
		return 99999
	}
	var days int
	_, _ = fmt.Sscanf(activity, "%d days ago", &days)
	return days
}

// generateInsightSummary creates a one-line hook based on analysis findings.
func (ga *GitHubAnalyzer) generateInsightSummary(analysis *RepoAnalysis) string {
	var parts []string

	if analysis.PrimaryLanguage != "" {
		parts = append(parts, fmt.Sprintf("primarily %s", analysis.PrimaryLanguage))
	}
	if len(analysis.Topics) > 0 {
		topicStr := strings.Join(analysis.Topics[:min(3, len(analysis.Topics))], ", ")
		parts = append(parts, fmt.Sprintf("topics: %s", topicStr))
	}
	if analysis.StarsTotal > 1000 {
		parts = append(parts, fmt.Sprintf("%d+ stars across repos", analysis.StarsTotal))
	}
	if analysis.StarsTotal > 100 {
		parts = append(parts, fmt.Sprintf("%d stars", analysis.StarsTotal))
	}
	if len(analysis.Bottlenecks) > 0 {
		parts = append(parts, fmt.Sprintf("detected %d bottleneck(s)", len(analysis.Bottlenecks)))
	}

	if len(parts) > 0 {
		return strings.Join(parts, "; ")
	}
	return fmt.Sprintf("analyzed %d repos, %d languages detected", analysis.ReposAnalyzed, len(analysis.Languages))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
