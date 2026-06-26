package orchestrator

import (
	"strings"

	foundationorchestration "github.com/NexusSoftMDMA/TormentNexus/foundation/orchestration"
)

func buildAutoDriveObjective(prompt, workingDir string) string {
	plan, err := foundationorchestration.BuildPlan(foundationorchestration.PlanRequest{
		Prompt:       prompt,
		WorkingDir:   workingDir,
		IncludeRepo:  true,
		MaxRepoFiles: 6,
		TaskType:     "coding",
	})
	if err != nil {
		return prompt
	}
	return strings.Join(plan.Steps, "\n")
}
