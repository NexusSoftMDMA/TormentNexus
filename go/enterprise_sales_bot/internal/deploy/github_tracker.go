package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// GitHubCITracker implements the CITracker interface using the GitHub Actions API.
type GitHubCITracker struct {
	owner string
	repo  string
	token string
}

// NewGitHubCITracker creates a new tracker instance.
func NewGitHubCITracker(owner, repo string) *GitHubCITracker {
	return &GitHubCITracker{
		owner: owner,
		repo:  repo,
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

type workflowRunsResponse struct {
	WorkflowRuns []struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"workflow_runs"`
}

func (g *GitHubCITracker) GetLatestStatus(ctx context.Context, branch string) (CIStatus, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?branch=%s&per_page=1", g.owner, g.repo, branch)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return CIStatusUnknown, err
	}

	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CIStatusUnknown, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return CIStatusUnknown, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var data workflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return CIStatusUnknown, err
	}

	if len(data.WorkflowRuns) == 0 {
		return CIStatusPending, nil
	}

	run := data.WorkflowRuns[0]
	if run.Status != "completed" {
		return CIStatusPending, nil
	}

	switch run.Conclusion {
	case "success":
		return CIStatusSuccess, nil
	case "failure", "cancelled", "timed_out", "action_required":
		return CIStatusFailure, nil
	case "skipped":
		return CIStatusSuccess, nil // Treat skipped as non-failure for gating
	default:
		return CIStatusUnknown, nil
	}
}

func (g *GitHubCITracker) GetSystemHealth(ctx context.Context) (string, error) {
	// For health, we can check the status of the main branch
	status, err := g.GetLatestStatus(ctx, "main")
	if err != nil {
		return "Health check failed", err
	}
	return fmt.Sprintf("Main branch status: %s", status), nil
}
