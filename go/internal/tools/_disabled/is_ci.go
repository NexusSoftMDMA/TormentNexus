package tools

import (
	"context"
	"os"
)

// HandleIsCI detects if the current environment is a CI server.
func HandleIsCI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "BUILD_NUMBER", "JENKINS_HOME", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "TRAVIS", "TEAMCITY_VERSION"}
	for _, v := range ciVars {
		if _, found := os.LookupEnv(v); found {
			return ok("true")
		}
	}
	return ok("false")
}