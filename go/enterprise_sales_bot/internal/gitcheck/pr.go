package gitcheck

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// PRStatus represents the current state of a Pull Request.
type PRStatus string

const (
	PRStatusOpen	PRStatus	= "Open"
	PRStatusMerged	PRStatus	= "Merged"
	PRStatusClosed	PRStatus	= "Closed"
	PRStatusFailed	PRStatus	= "Failed"
)

// PullRequest encapsulates metadata for an autonomous PR.
type PullRequest struct {
	ID	string
	URL	string
	Branch	string
	Title	string
	Status	PRStatus
}

// PRManager defines the interface for autonomous pull request operations.
type PRManager interface {
	// CreatePullRequest creates a new PR from the source branch to the target.
	CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error)

	// GetPRStatus retrieves the current status of a PR.
	GetPRStatus(ctx context.Context, prID string) (PRStatus, error)

	// MergePullRequest merges an open PR.
	MergePullRequest(ctx context.Context, prID string) error

	// GetPRComments retrieves all comments for a PR.
	GetPRComments(ctx context.Context, prID string) ([]string, error)
}

// GitHubPRManager implements the PRManager interface using the GitHub API.
type GitHubPRManager struct {
	client	*github.Client
	owner	string
	repo	string
}

// NewGitHubPRManager creates a new GitHubPRManager instance.
func NewGitHubPRManager(owner, repo string) *GitHubPRManager {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Info("GitHubPRManager: Warning: GITHUB_TOKEN not set.")
		return &GitHubPRManager{owner: owner, repo: repo}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &GitHubPRManager{
		client:	client,
		owner:	owner,
		repo:	repo,
	}
}

func (g *GitHubPRManager) CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error) {
	if g.client == nil {
		return nil, fmt.Errorf("github client not initialized")
	}

	slog.Info(fmt.Sprintf("GitHubPRManager: Creating Pull Request for branch %s: %s", branch, title))

	head := branch
	base := "main"
	newPR := &github.NewPullRequest{
		Title:			github.String(title),
		Head:			github.String(head),
		Base:			github.String(base),
		Body:			github.String(body),
		MaintainerCanModify:	github.Bool(true),
	}

	pr, _, err := g.client.PullRequests.Create(ctx, g.owner, g.repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return &PullRequest{
		ID:	strconv.Itoa(pr.GetNumber()),
		URL:	pr.GetHTMLURL(),
		Branch:	branch,
		Title:	title,
		Status:	PRStatusOpen,
	}, nil
}

func (g *GitHubPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, error) {
	if g.client == nil {
		return PRStatusOpen, nil	// Fallback for simulation
	}

	number, err := strconv.Atoi(prID)
	if err != nil {
		return PRStatusFailed, err
	}

	pr, _, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, number)
	if err != nil {
		return PRStatusFailed, err
	}

	if pr.GetMerged() {
		return PRStatusMerged, nil
	}

	state := strings.ToLower(pr.GetState())
	switch state {
	case "open":
		return PRStatusOpen, nil
	case "closed":
		return PRStatusClosed, nil
	default:
		return PRStatusOpen, nil
	}
}

func (g *GitHubPRManager) MergePullRequest(ctx context.Context, prID string) error {
	if g.client == nil {
		slog.Info(fmt.Sprintf("GitHubPRManager: Simulating PR merge for %s", prID))
		return nil
	}

	slog.Info(fmt.Sprintf("GitHubPRManager: Merging Pull Request %s", prID))

	number, err := strconv.Atoi(prID)
	if err != nil {
		return err
	}

	opts := &github.PullRequestOptions{
		MergeMethod: "squash",
	}

	_, _, err = g.client.PullRequests.Merge(ctx, g.owner, g.repo, number, "Autonomous merge by sales-bot", opts)
	if err != nil {
		return fmt.Errorf("failed to merge PR %s: %w", prID, err)
	}

	return nil
}

func (g *GitHubPRManager) GetPRComments(ctx context.Context, prID string) ([]string, error) {
	if g.client == nil {
		return []string{}, nil
	}

	number, err := strconv.Atoi(prID)
	if err != nil {
		return nil, err
	}

	comments, _, err := g.client.Issues.ListComments(ctx, g.owner, g.repo, number, nil)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, c := range comments {
		result = append(result, c.GetBody())
	}
	return result, nil
}

// MockPRManager is a simulated implementation for testing.
type MockPRManager struct{}

func (m *MockPRManager) CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error) {
	return &PullRequest{
		ID:	"mock-456",
		Branch:	branch,
		Title:	title,
		Status:	PRStatusOpen,
	}, nil
}

func (m *MockPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, error) {
	return PRStatusOpen, nil
}

func (m *MockPRManager) MergePullRequest(ctx context.Context, prID string) error {
	return nil
}

func (m *MockPRManager) GetPRComments(ctx context.Context, prID string) ([]string, error) {
	return []string{"Mock comment 1", "Mock comment 2"}, nil
}
