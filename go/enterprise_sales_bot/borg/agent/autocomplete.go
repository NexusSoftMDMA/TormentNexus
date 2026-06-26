package agent

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// SuggestCompletion mimics Auggie's context-aware completions.
// It analyzes code before and after the cursor to suggest completions.
func (a *Agent) SuggestCompletion(prefix, suffix string) (string, error) {
	prompt := fmt.Sprintf("Provide the code completion that goes exactly between this prefix and suffix.\n\nPrefix:\n%s\n\nSuffix:\n%s\n\nOutput ONLY the completion string.", prefix, suffix)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o, // Use appropriate fast model like gpt-4o-mini or a specialized coder model
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.2, // Lower temperature for more deterministic completions
	}

	resp, err := a.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
