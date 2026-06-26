package tools

import (
	"context"
	"fmt"
	"os"
)

func HandleCiInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ciVars := []struct{ key, label string }{
		{"CI", "CI"},
		{"GITHUB_ACTIONS", "GitHub Actions"},
		{"GITLAB_CI", "GitLab CI"},
		{"JENKINS_HOME", "Jenkins"},
		{"TRAVIS", "Travis CI"},
		{"CIRCLECI", "CircleCI"},
	}
	info := "Current CI environment:\n"
	for _, v := range ciVars {
		val := os.Getenv(v.key)
		if val != "" {
			info += fmt.Sprintf("- %s: %s\n", v.label, val)

	}
	return ok(info)
}
}