package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// HandleRepoSync handles automated dependency management for a project.
func HandleRepoSync(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ := args["project_path"].(string)
	if projectPath == "" {
		return err("project_path is required")
	}

	absPath, e := filepath.Abs(projectPath)
	if e != nil {
		return err(e.Error())
	}

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/repo_sync/main.go")
	cmd.Dir = absPath

	output, e := cmd.CombinedOutput()
	if e != nil {
		return ok(fmt.Sprintf("Sync failed: %v\nOutput: %s", e, string(output)))
	}

	return ok(fmt.Sprintf("Repository synchronized successfully at %s", absPath))
}

// HandleProjectDeploy handles automated testing and deployment for a project.
func HandleProjectDeploy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ := args["project_path"].(string)
	environment, _ := args["environment"].(string)
	if projectPath == "" {
		return err("project_path is required")
	}
	if environment == "" {
		environment = "production"
	}

	absPath, e := filepath.Abs(projectPath)
	if e != nil {
		return err(e.Error())
	}

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/deployment_manager/main.go")
	cmd.Dir = absPath

	output, e := cmd.CombinedOutput()
	if e != nil {
		return ok(fmt.Sprintf("Deployment failed: %v\nOutput: %s", e, string(output)))
	}

	return ok(fmt.Sprintf("Project deployed to %s successfully from %s", environment, absPath))
}
