package agent

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"strings"
)

// CompareModels matches RowboatX's capability to evaluate multiple models side-by-side.
func (a *Agent) CompareModels(prompt string, models []string) (string, error) {
	var results strings.Builder
	results.WriteString(fmt.Sprintf("### Model Comparison for: %s ###\n\n", prompt))

	for _, modelName := range models {
		req := openai.ChatCompletionRequest{
			Model: modelName,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
		}

		resp, err := a.client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			results.WriteString(fmt.Sprintf("#### Model: %s ####\nError: %v\n\n", modelName, err))
			continue
		}

		results.WriteString(fmt.Sprintf("#### Model: %s ####\n%s\n\n", modelName, resp.Choices[0].Message.Content))
	}

	return results.String(), nil
}
