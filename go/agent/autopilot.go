package agent

import (
	"fmt"
)

// AutopilotMode mimics Opencode-Autopilot and CLI-Orchestrator.
// It enters an autonomous loop until a goal is achieved.
func (a *Agent) AutopilotMode(goal string) (string, error) {
	fmt.Printf("[Autopilot] Goal set: %s\n", goal)

	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		fmt.Printf("[Autopilot] Iteration %d/%d...\n", i+1, maxIterations)

		prompt := fmt.Sprintf("Goal: %s\nIteration: %d\nYou are in Autopilot Mode. If the goal is not yet achieved, use your tools to make progress. If achieved, respond with 'GOAL_ACHIEVED'.", goal, i+1)

		response, err := a.Chat(prompt)
		if err != nil {
			return "", err
		}

		if response == "GOAL_ACHIEVED" {
			return "Autopilot completed: Goal achieved.", nil
		}

		fmt.Printf("[Autopilot] Progress: %s\n", response)
	}

	return "Autopilot halted: Max iterations reached.", nil
}
