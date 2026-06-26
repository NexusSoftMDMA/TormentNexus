package agent

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// SuggestShellCommand mimics Copilot CLI and Shell Pilot.
// It translates natural language into a runnable shell command.
func (a *Agent) SuggestShellCommand(query string) (string, error) {
	prompt := fmt.Sprintf("Translate this natural language request into a single, valid shell command for a Windows environment. Output ONLY the command, nothing else. Request: %s", query)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	}

	resp, err := a.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	command := resp.Choices[0].Message.Content
	return command, nil
}
