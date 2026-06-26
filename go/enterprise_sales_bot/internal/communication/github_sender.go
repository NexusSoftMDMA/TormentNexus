package communication

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"golang.org/x/oauth2"
)

// GitHubCommentSender posts technical comments on GitHub Issues and PRs
// as a "technical hook" outreach for target companies.
type GitHubCommentSender struct {
	client		*github.Client
	repo		string	// e.g. "robertpelloni/enterprise_sales_bot"
	username	string	// GitHub username for the bot account
}

// NewGitHubCommentSender creates a GitHubCommentSender.
// If GITHUB_TOKEN is set, it creates an authenticated client.
// repo is the bot's repository for generating relevant comments.
func NewGitHubCommentSender(repo string) *GitHubCommentSender {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Info("GitHubCommentSender: Warning: GITHUB_TOKEN not set, will use unauthenticated client (rate limited)")
		return &GitHubCommentSender{
			client:		github.NewClient(nil),
			repo:		repo,
			username:	"hypernexus-bot",
		}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return &GitHubCommentSender{
		client:		github.NewClient(tc),
		repo:		repo,
		username:	os.Getenv("GITHUB_BOT_USERNAME"),
	}
}

// SendComment posts a technical comment on a GitHub Issue or PR.
// owner/repo: the target repository
// issueNumber: the issue or PR number to comment on
// commentBody: the message body to post
func (g *GitHubCommentSender) SendComment(ctx context.Context, owner, repo string, issueNumber int, commentBody string) error {
	if g.client == nil {
		return fmt.Errorf("GitHubCommentSender: client not initialized")
	}

	comment := &github.IssueComment{
		Body: github.String(commentBody),
	}

	_, _, err := g.client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)
	if err != nil {
		return fmt.Errorf("GitHubCommentSender: failed to create comment on %s/%s#%d: %w", owner, repo, issueNumber, err)
	}

	slog.Info(fmt.Sprintf("GitHubCommentSender: Comment posted on %s/%s#%d", owner, repo, issueNumber))
	return nil
}

// SearchRelevantIssues searches a target company's GitHub repos for issues
// related to AI infrastructure, LLM orchestration, or MCP (TormentNexus's niche).
// Returns issue URLs and titles.
func (g *GitHubCommentSender) SearchRelevantIssues(ctx context.Context, companyDomain string) ([]IssueTarget, error) {
	if g.client == nil {
		return nil, fmt.Errorf("GitHubCommentSender: client not initialized")
	}

	// Extract owner from domain or use a search approach
	// We search for issues mentioning keywords in the company's GitHub repos
	searchTerms := []string{
		"AI infrastructure",
		"LLM orchestration",
		"MCP server",
		"model routing",
		"agent workflow",
		"multi-agent",
		"tool orchestration",
		"prompt management",
		"AI observability",
	}

	var targets []IssueTarget

	for _, term := range searchTerms {
		query := fmt.Sprintf("%s org:%s is:issue is:open", term, companyDomain)
		results, _, err := g.client.Search.Issues(ctx, query, &github.SearchOptions{
			Sort:	"updated",
			Order:	"desc",
			ListOptions: github.ListOptions{
				PerPage: 3,
			},
		})
		if err != nil {
			slog.Info(fmt.Sprintf("GitHubCommentSender: Search error for %q: %v", query, err))
			continue
		}

		for _, issue := range results.Issues {
			// Split owner/repo from issue URL
			parts := strings.Split(issue.GetRepositoryURL(), "/")
			repoOwner := ""
			repoName := ""
			if len(parts) >= 2 {
				repoOwner = parts[len(parts)-2]
				repoName = parts[len(parts)-1]
			}

			targets = append(targets, IssueTarget{
				Owner:		repoOwner,
				Repo:		repoName,
				IssueNumber:	issue.GetNumber(),
				Title:		issue.GetTitle(),
				URL:		issue.GetHTMLURL(),
				Relevance:	CalculateRelevance(term, issue.GetTitle(), issue.GetBody()),
			})
		}
	}

	// Deduplicate by URL
	seen := make(map[string]bool)
	var unique []IssueTarget
	for _, t := range targets {
		if !seen[t.URL] {
			seen[t.URL] = true
			unique = append(unique, t)
		}
	}

	return unique, nil
}

// IssueTarget represents a GitHub issue or PR that is relevant for outreach.
type IssueTarget struct {
	Owner		string
	Repo		string
	IssueNumber	int
	Title		string
	URL		string
	Relevance	int	// higher = more relevant
}

// CalculateRelevance scores how relevant an issue is to TormentNexus.
func CalculateRelevance(term, title, body string) int {
	score := 0
	lowerTitle := strings.ToLower(title)
	lowerBody := strings.ToLower(body)

	// High-value keywords
	highValue := []string{"MCP", "model context protocol", "agent", "orchestrat", "tool routing", "LLM"}
	for _, kw := range highValue {
		if strings.Contains(lowerTitle, kw) {
			score += 3
		}
		if strings.Contains(lowerBody, kw) {
			score += 1
		}
	}

	// Medium-value keywords
	mediumValue := []string{"infrastructure", "pipeline", "workflow", "automation", "integration", "deploy"}
	for _, kw := range mediumValue {
		if strings.Contains(lowerTitle, kw) {
			score += 2
		}
		if strings.Contains(lowerBody, kw) {
			score += 1
		}
	}

	// Bonus for matching search term
	if strings.Contains(strings.ToLower(term), lowerTitle) || strings.Contains(lowerTitle, strings.ToLower(term)) {
		score += 2
	}

	return score
}

// GenerateTechHookComment generates a helpful, value-first comment related to
// the issue that positions TormentNexus as a solution.
func GenerateTechHookComment(issue IssueTarget) string {
	comment := fmt.Sprintf(`Hi there! 👋

I noticed this issue about %s — we've been working on similar challenges with **HyperNexus** (hypernexus.site).

HyperNexus is the enterprise cloud-hosted version of TormentNexus, built using our stable open-source fork located at github.com/HyperNexusSoft/HyperNexus. It coordinates multi-agent LLM workflows with:
- **Progressive MCP Tool Routing** — semantic router that injects only the 3 most relevant tools per request (no 50K-token tool dumps)
- **Cross-Harness Tool Parity** — byte-for-byte identical tool signatures for Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf
- **LLM Waterfall** — seamless cascading through cloud providers → aggregators → local models on 429/5xx errors
- **Local-First Memory** — 14K+ persisted memories with sqlite-vec semantic search, surviving restarts

If this resonates with what you're building, I'd love to hear your thoughts. We're always looking for feedback from teams pushing the boundaries of AI infrastructure.

— HyperNexus Bot`, issue.Title)

	return comment
}

// FindAndComment is a high-level operation that searches for relevant issues
// in a target organization's GitHub repos and posts a technical hook comment.
func (g *GitHubCommentSender) FindAndComment(ctx context.Context, company db.Company, contact db.Contact) error {
	// Search for relevant issues
	targets, err := g.SearchRelevantIssues(ctx, extractOrgFromDomain(company.Domain))
	if err != nil {
		return fmt.Errorf("GitHubCommentSender: search failed: %w", err)
	}

	if len(targets) == 0 {
		slog.Info(fmt.Sprintf("GitHubCommentSender: No relevant issues found for %s", company.Domain))
		return nil
	}

	// Pick the most relevant issue
	bestTarget := targets[0]
	for _, t := range targets[1:] {
		if t.Relevance > bestTarget.Relevance {
			bestTarget = t
		}
	}

	// Generate and post the comment
	comment := GenerateTechHookComment(bestTarget)
	if err := g.SendComment(ctx, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber, comment); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("GitHubCommentSender: Technical hook posted for %s on %s/%s#%d",
		company.Name, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber))

	return nil
}

// extractOrgFromDomain extracts the likely GitHub org name from a company domain.
func extractOrgFromDomain(domain string) string {
	// Handle URLs
	if strings.HasPrefix(domain, "http") {
		parsed, err := url.Parse(domain)
		if err == nil {
			domain = parsed.Host
		}
	}

	// Remove TLD and subdomain
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[0]
	}
	return domain
}
