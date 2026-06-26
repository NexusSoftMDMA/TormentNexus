package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// GitHubIssueSource implements LeadSource by searching GitHub issues for companies with relevant technical challenges.
type GitHubIssueSource struct {
	Client		*http.Client
	Token		string
	Keywords	[]string
}

// githubIssueSearchResponse represents the GitHub Search API response for issues.
type githubIssueSearchResponse struct {
	TotalCount	int	`json:"total_count"`
	Issues		[]struct {
		HTMLURL		string	`json:"html_url"`
		Title		string	`json:"title"`
		Body		string	`json:"body"`
		RepoURL		string	`json:"repository_url"`
		CreatedAt	string	`json:"created_at"`
	}	`json:"items"`
}

// githubRepo represents a GitHub repository.
type githubRepo struct {
	FullName	string	`json:"full_name"`
	HTMLURL		string	`json:"html_url"`
	Description	string	`json:"description"`
	Language	string	`json:"language"`
	Owner		struct {
		Login string `json:"login"`
	}	`json:"owner"`
}

// Discover searches GitHub for open issues matching TormentNexus-relevant keywords,
// extracts organization information, and returns them as potential leads.
func (g *GitHubIssueSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	if g.Token == "" {
		g.Token = os.Getenv("GITHUB_TOKEN")
	}

	if g.Token == "" {
		slog.Info("GitHubIssueSource: No GITHUB_TOKEN set, returning empty results")
		return []db.Company{}, nil
	}

	if g.Client == nil {
		g.Client = http.DefaultClient
	}

	if len(g.Keywords) == 0 {
		g.Keywords = []string{
			"MCP",
			"model context protocol",
			"tool routing",
			"multi-agent",
			"agent orchestration",
			"LLM orchestration",
			"LLM management",
			"agent workflow",
			"tool calling",
			"function calling",
		}
	}

	slog.Info(fmt.Sprintf("GitHubIssueSource: Searching for issues with keywords: %v", g.Keywords))

	companiesMap := make(map[string]*db.Company)

	// Search for each keyword
	for _, keyword := range g.Keywords {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		issues, err := g.searchIssues(ctx, keyword)
		if err != nil {
			slog.Info(fmt.Sprintf("GitHubIssueSource: Error searching for '%s': %v", keyword, err))
			continue
		}

		// Process each issue to extract company information
		for _, issue := range issues {
			orgName, err := g.extractOrgFromRepoURL(issue.RepoURL)
			if err != nil {
				continue
			}

			// Skip well-known large orgs that aren't target customers
			if g.isExcludedOrg(orgName) {
				continue
			}

			domain := g.orgToDomain(orgName)
			if domain == "" {
				continue
			}

			// Check if we already have this company
			if existing, ok := companiesMap[domain]; ok {
				// Append this issue as an additional hiring signal
				signal := fmt.Sprintf("GitHub Issue: %s - %s", issue.Title, issue.HTMLURL)
				existing.HiringSignals = append(existing.HiringSignals, signal)
			} else {
				// Fetch org details to enrich the company
				repoInfo, err := g.getOrgInfo(ctx, orgName)
				if err != nil {
					continue
				}

				company := &db.Company{
					Name:		repoInfo.Owner.Login,
					Domain:		domain,
					TechStack:	[]string{repoInfo.Language},
					HiringSignals:	[]string{fmt.Sprintf("GitHub Issue: %s - %s", issue.Title, issue.HTMLURL)},
					MarketCapTier:	g.inferMarketCap(repoInfo),
				}
				companiesMap[domain] = company
			}
		}
	}

	// Convert map to slice
	var companies []db.Company
	for _, company := range companiesMap {
		companies = append(companies, *company)
	}

	slog.Info(fmt.Sprintf("GitHubIssueSource: Discovered %d companies from GitHub issues", len(companies)))
	return companies, nil
}

// searchIssues queries the GitHub Search API for issues matching the given keyword.
func (g *GitHubIssueSource) searchIssues(ctx context.Context, keyword string) ([]struct {
	HTMLURL		string	`json:"html_url"`
	Title		string	`json:"title"`
	Body		string	`json:"body"`
	RepoURL		string	`json:"repository_url"`
	CreatedAt	string	`json:"created_at"`
}, error) {
	// Build search query: keyword in title/body, open issues only, exclude very large orgs
	query := fmt.Sprintf("%s type:issue state:open", keyword)
	queryURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=10", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result githubIssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Issues, nil
}

// extractOrgFromRepoURL extracts the organization name from a GitHub repository URL.
// Example: "https://api.github.com/repos/acme-corp/my-repo" -> "acme-corp"
func (g *GitHubIssueSource) extractOrgFromRepoURL(repoURL string) (string, error) {
	// Repo URL format: https://api.github.com/repos/{owner}/{repo}
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid repo URL: %s", repoURL)
	}
	// The last two parts are owner and repo
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	// Return owner (org name)
	_ = repo	// repo name not used here
	return owner, nil
}

// getOrgInfo fetches repository information to extract organization details.
func (g *GitHubIssueSource) getOrgInfo(ctx context.Context, orgName string) (*githubRepo, error) {
	// Search for recent repos by this org to get org metadata
	query := fmt.Sprintf("user:%s", orgName)
	queryURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=updated&per_page=1", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for org %s", resp.StatusCode, orgName)
	}

	var result struct {
		Items []githubRepo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no repos found for org %s", orgName)
	}

	return &result.Items[0], nil
}

// isExcludedOrg checks if an organization should be excluded from targeting.
func (g *GitHubIssueSource) isExcludedOrg(orgName string) bool {
	// Exclude very large tech companies and GitHub itself
	excluded := []string{
		"microsoft", "google", "amazon", "facebook", "meta", "apple",
		"netflix", "uber", "airbnb", "twitter", "linkedin", "github",
		"gitlab", "atlassian", "jetbrains", "docker", "kubernetes",
		"canonical", "redhat", "ibm", "oracle", "salesforce", "slack",
		"stripe", "twilio", "mongodb", "elastic", "confluent",
	}

	for _, ex := range excluded {
		if strings.EqualFold(orgName, ex) {
			return true
		}
	}
	return false
}

// orgToDomain attempts to map a GitHub organization name to a domain.
// This is a heuristic - real implementation would use a lookup service.
func (g *GitHubIssueSource) orgToDomain(orgName string) string {
	// Try common patterns
	name := strings.ToLower(orgName)

	// Remove common suffixes
	name = strings.TrimSuffix(name, "-inc")
	name = strings.TrimSuffix(name, "-corp")
	name = strings.TrimSuffix(name, "-hq")
	name = strings.TrimSuffix(name, "-dev")
	name = strings.TrimSuffix(name, "-labs")
	name = strings.TrimSuffix(name, "-tech")

	// Add .tech or .io domains as defaults for tech companies
	if strings.Contains(name, "ai") || strings.Contains(name, "ml") ||
		strings.Contains(name, "neural") || strings.Contains(name, "logic") {
		return name + ".io"
	}

	return name + ".tech"
}

// inferMarketCap infers a company's market cap tier from repository metadata.
func (g *GitHubIssueSource) inferMarketCap(repo *githubRepo) string {
	// Heuristics based on repo metadata
	if repo.Language == "" {
		return "Small Business"
	}

	// Star count and activity heuristics
	// (In a real implementation, we'd fetch star count from the repo)
	if strings.Contains(repo.Description, "enterprise") ||
		strings.Contains(repo.Description, "platform") {
		return "Mid-Market"
	}

	return "Small Business"
}
