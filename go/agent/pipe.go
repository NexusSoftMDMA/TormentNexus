package agent

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
)

// PipeProcessor mimics Simon Willison's "LLM CLI" and Charmbracelet's "Crush".
// It takes data from standard input, processes it through the LLM with a prompt,
// and streams or returns the result.
func (a *Agent) ProcessPipe(prompt string) (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if data is being piped to Stdin
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("no data piped to stdin")
	}

	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	execution := adapters.PrepareProviderExecution(adapters.ProviderExecutionRequest{Prompt: prompt, TaskType: "analysis", CostPreference: "budget"})
	combinedPrompt := fmt.Sprintf("%s\n\nExecution Hint:\n%s\n\nInput Data:\n%s", prompt, execution.ExecutionHint, string(inputData))

	fmt.Printf("[PipeProcessor] Processing %d bytes of piped data via %s/%s...\n", len(inputData), execution.Route.Provider, execution.Route.Model)

	response, err := a.Chat(combinedPrompt)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(execution.ExecutionHint) != "" {
		return fmt.Sprintf("[Pipe Execution]\n%s\n\n%s", execution.ExecutionHint, response), nil
	}
	return response, nil
}
