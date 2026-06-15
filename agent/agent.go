package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/tormentnexus"
	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
	"github.com/NexusSoftMDMA/TormentNexus/tools"
	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	client       *openai.Client
	messages     []openai.ChatCompletionMessage
	tools        *tools.Registry
	TormentNexusAdapter  *tormentnexus.Adapter
	HyperAdapter *adapters.TormentNexusAdapter
}

func NewAgent() *Agent {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "dummy"
	}

	registry := tools.NewRegistry()
	tormentnexusAdapter := tormentnexus.NewAdapter()
	cwd, _ := os.Getwd()
	hyperAdapter := adapters.NewTormentNexusAdapter(cwd)
	systemPrompt := strings.Join([]string{
		"You are TormentNexus, a Go-native coding and terminal assistant integrated with TormentNexus and TormentNexus.",
		"Prefer the exact-name Pi-compatible tools read, write, edit, and bash when solving coding tasks.",
		"Use repomap for repository-wide context when a condensed map would help.",
		"Additional legacy tools may exist for compatibility, but exact-contract tools are preferred.",
		hyperAdapter.BuildSystemContext(),
	}, "\n\n")

	return &Agent{
		client: openai.NewClient(apiKey),
		messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
		},
		tools:        registry,
		TormentNexusAdapter:  tormentnexusAdapter,
		HyperAdapter: hyperAdapter,
	}
}

func (a *Agent) buildOpenAITools() []openai.Tool {
	openAITools := make([]openai.Tool, 0, len(a.tools.Tools))
	for _, t := range a.tools.Tools {
		openAITools = append(openAITools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  append(json.RawMessage(nil), t.Parameters...),
			},
		})
	}
	return openAITools
}

func (a *Agent) Chat(input string) (string, error) {
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: a.messages,
		Tools:    a.buildOpenAITools(),
	}

	resp, err := a.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	msg := resp.Choices[0].Message
	a.messages = append(a.messages, msg)

	if len(msg.ToolCalls) > 0 {
		return a.handleToolCalls(msg.ToolCalls)
	}

	return msg.Content, nil
}

func (a *Agent) handleToolCalls(toolCalls []openai.ToolCall) (string, error) {
	resultSummary := ""

	for _, tc := range toolCalls {
		var args map[string]interface{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

		var toolResult string
		found := false
		for _, t := range a.tools.Tools {
			if t.Name == tc.Function.Name {
				found = true
				out, err := t.Execute(args)
				if err != nil {
					if out != "" {
						toolResult = out
					} else {
						toolResult = fmt.Sprintf("Error executing %s: %v", t.Name, err)
					}
				} else {
					toolResult = out
				}
				break
			}
		}
		if !found {
			toolResult = fmt.Sprintf("Unknown tool: %s", tc.Function.Name)
		}

		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    toolResult,
			Name:       tc.Function.Name,
			ToolCallID: tc.ID,
		})

		resultSummary += fmt.Sprintf("Executed tool %s\n", tc.Function.Name)
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: a.messages,
		Tools:    a.buildOpenAITools(),
	}

	resp, err := a.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return resultSummary, err
	}

	msg := resp.Choices[0].Message
	a.messages = append(a.messages, msg)

	return resultSummary + "\n" + msg.Content, nil
}
