package agent

import (
	"fmt"
)

// Oracle Mode mimics Amp Code's "Oracle" and Deep Research agents.
// It executes a multi-step reasoning loop before returning a final answer.
func (a *Agent) OracleQuery(prompt string) (string, error) {
	fmt.Println("[Oracle Mode] Initiating deep reasoning loop...")

	// Step 1: Formulate a research plan
	planPrompt := fmt.Sprintf("You are in Oracle Mode. Create a multi-step research plan to answer this complex query: %s. Output ONLY the numbered steps.", prompt)
	plan, err := a.Chat(planPrompt)
	if err != nil {
		return "", err
	}

	fmt.Println("[Oracle Mode] Research Plan Generated:")
	fmt.Println(plan)

	// Step 2: Execute the plan using tools
	executionPrompt := fmt.Sprintf("Execute this research plan using your available tools. Gather all necessary context.\nPlan:\n%s\n\nOriginal Query: %s", plan, prompt)

	// We use the standard Chat function which handles tool execution
	researchData, err := a.Chat(executionPrompt)
	if err != nil {
		return "", err
	}

	// Step 3: Synthesize the final answer
	synthesisPrompt := fmt.Sprintf("Based on the original query: '%s', and the gathered research data:\n%s\n\nProvide a comprehensive, authoritative, and definitive answer.", prompt, researchData)

	finalAnswer, err := a.Chat(synthesisPrompt)
	if err != nil {
		return "", err
	}

	return finalAnswer, nil
}
