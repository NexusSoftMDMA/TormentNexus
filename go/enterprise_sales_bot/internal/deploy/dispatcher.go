package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// WorkflowDispatcher defines the interface for triggering remote workflows.
type WorkflowDispatcher interface {
	Dispatch(ctx context.Context, workflowFile string, ref string, inputs map[string]interface{}) error
}

// GitHubDispatcher implements WorkflowDispatcher for GitHub Actions.
type GitHubDispatcher struct {
	owner string
	repo  string
	token string
}

// NewGitHubDispatcher creates a new dispatcher instance.
func NewGitHubDispatcher(owner, repo string) *GitHubDispatcher {
	return &GitHubDispatcher{
		owner: owner,
		repo:  repo,
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

func (g *GitHubDispatcher) Dispatch(ctx context.Context, workflowFile string, ref string, inputs map[string]interface{}) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", g.owner, g.repo, workflowFile)

	payload := map[string]interface{}{
		"ref":    ref,
		"inputs": inputs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	return nil
}
