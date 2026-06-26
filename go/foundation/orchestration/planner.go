package orchestration

import (
	"fmt"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
	foundationrepomap "github.com/NexusSoftMDMA/TormentNexus/foundation/repomap"
)

type PlanRequest struct {
	Prompt       string `json:"prompt"`
	WorkingDir   string `json:"workingDir,omitempty"`
	IncludeRepo  bool   `json:"includeRepo"`
	MaxRepoFiles int    `json:"maxRepoFiles,omitempty"`
	TaskType     string `json:"taskType,omitempty"`
	Cost         string `json:"cost,omitempty"`
	RequireLocal bool   `json:"requireLocal,omitempty"`
}

type PlanResult struct {
	Prompt            string                           `json:"prompt"`
	TaskType          string                           `json:"taskType"`
	Execution         adapters.ProviderExecutionResult `json:"execution"`
	RepoMap           string                           `json:"repoMap,omitempty"`
	RepoMapIncluded   bool                             `json:"repoMapIncluded"`
	Steps             []string                         `json:"steps"`
	SystemContextHint string                           `json:"systemContextHint,omitempty"`
}

func BuildPlan(req PlanRequest) (PlanResult, error) {
	cwd := strings.TrimSpace(req.WorkingDir)
	if cwd == "" {
		resolved, err := os.Getwd()
		if err != nil {
			return PlanResult{}, err
		}
		cwd = resolved
	}
	if req.MaxRepoFiles <= 0 {
		req.MaxRepoFiles = 8
	}
	execution := adapters.PrepareProviderExecution(adapters.ProviderExecutionRequest{
		Prompt:         req.Prompt,
		TaskType:       req.TaskType,
		CostPreference: req.Cost,
		RequireLocal:   req.RequireLocal,
	})
	steps := deriveSteps(execution.TaskType, req.Prompt)
	result := PlanResult{
		Prompt:            req.Prompt,
		TaskType:          execution.TaskType,
		Execution:         execution,
		Steps:             steps,
		SystemContextHint: execution.ExecutionHint,
	}
	if req.IncludeRepo || shouldIncludeRepoMap(req.Prompt) {
		mapResult, err := foundationrepomap.Generate(foundationrepomap.Options{BaseDir: cwd, MaxFiles: req.MaxRepoFiles})
		if err == nil {
			result.RepoMap = mapResult.Map
			result.RepoMapIncluded = true
		} else {
			result.Steps = append(result.Steps, fmt.Sprintf("Repo map unavailable: %v", err))
		}
	}
	return result, nil
}

func shouldIncludeRepoMap(prompt string) bool {
	lower := strings.ToLower(prompt)
	for _, needle := range []string{"repo", "repository", "codebase", "file", "files", "refactor", "architecture", "search"} {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func deriveSteps(taskType, prompt string) []string {
	steps := []string{"Interpret the request and confirm the primary objective."}
	switch taskType {
	case "coding":
		steps = append(steps,
			"Inspect relevant files and gather repository context.",
			"Prepare an execution route using the provider adapter.",
			"Apply or propose the required code changes with exact-name tools.",
			"Verify the outcome and summarize next actions.",
		)
	case "analysis":
		steps = append(steps,
			"Collect repository context and relevant files.",
			"Route the analysis through the selected provider profile.",
			"Synthesize findings and identify concrete next actions.",
		)
	case "local":
		steps = append(steps,
			"Confirm local execution requirements.",
			"Use the local/provider route selected by the adapter.",
			"Summarize constraints, findings, and next actions.",
		)
	default:
		steps = append(steps,
			"Gather enough context to answer accurately.",
			"Select an execution route appropriate for the task.",
			"Deliver the result with follow-up recommendations.",
		)
	}
	if strings.TrimSpace(prompt) != "" {
		steps = append(steps, fmt.Sprintf("Original request: %s", strings.TrimSpace(prompt)))
	}
	return steps
}
