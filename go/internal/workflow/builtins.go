package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ShellStep creates a step that runs a shell command
func ShellStep(id, name, command, cwd string, deps ...string) *Step {
	return &Step{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("Run: %s", command),
		DependsOn:   deps,
		Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			parts := strings.Fields(command)
			if len(parts) == 0 {
				return nil, fmt.Errorf("empty command")
			}

			cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
			cmd.Dir = cwd
			output, err := cmd.CombinedOutput()
			result := map[string]any{
				"stdout":   string(output),
				"exitCode": 0,
			}
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					result["exitCode"] = exitErr.ExitCode()
				}
				result["error"] = err.Error()
				return result, err
			}
			return result, nil
		},
	}
}

// FullBuildWorkflow creates the standard TormentNexus monorepo build workflow
func FullBuildWorkflow(workspaceRoot string) *Workflow {
	return NewWorkflow("full-build", "Full Monorepo Build", "Complete TormentNexus monorepo build pipeline", []*Step{
		ShellStep("install", "Install Dependencies", "pnpm install --frozen-lockfile", workspaceRoot),
		ShellStep("go-build", "Build Go Sidecar", "go build -buildvcs=false ./cmd/tormentnexus", workspaceRoot+"/go"),
		ShellStep("ts-build", "Build TypeScript Workspace", "pnpm run build:workspace", workspaceRoot, "install"),
	})
}

// SubmoduleSyncWorkflow creates a workflow to sync all submodules
func SubmoduleSyncWorkflow(workspaceRoot string) *Workflow {
	return NewWorkflow("submodule-sync", "Submodule Sync", "Sync all git submodules to latest remote commits", []*Step{
		ShellStep("fetch", "Fetch All", "git fetch --all --recurse-submodules", workspaceRoot),
		ShellStep("update", "Update Submodules", "git submodule update --remote --merge", workspaceRoot, "fetch"),
		ShellStep("status", "Check Status", "git submodule status", workspaceRoot, "update"),
	})
}

// LintAndTestWorkflow creates a quality-check workflow
func LintAndTestWorkflow(workspaceRoot string) *Workflow {
	return NewWorkflow("lint-test", "Lint & Test", "Run linting and testing across the monorepo", []*Step{
		ShellStep("go-vet", "Go Vet", "go vet ./...", workspaceRoot+"/go"),
		ShellStep("go-test", "Go Tests", "go test ./...", workspaceRoot+"/go"),
		ShellStep("ts-typecheck", "TypeScript Typecheck", "pnpm run build:workspace", workspaceRoot),
	})
}
