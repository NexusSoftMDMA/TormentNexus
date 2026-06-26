package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

// AutoCommit creates a commit with the given message automatically handling staging
func (rm *RepoManager) AutoCommit(message string) (string, error) {
	w, err := rm.repo.Worktree()
	if err != nil {
		return "", err
	}

	_, err = w.Add(".")
	if err != nil {
		return "", err
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "SuperCLI Agent",
			Email: "agent@supercli.dev",
			When:  time.Now(),
		},
	})

	if err != nil {
		return "", err
	}

	return commit.String(), nil
}
