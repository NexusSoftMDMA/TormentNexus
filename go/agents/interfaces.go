package agents

import (
	"context"
)

// MessageRole defines the role of a message sender
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message represents a single chat message
type Message struct {
	Role       MessageRole
	Content    string
	Name       string
	ToolCallID string
	ToolCalls  []ToolCall
}

// ToolCall represents a tool invocation from the LLM
type ToolCall struct {
	ID   string
	Name string
	Args string
}

// Tool represents a capability the agent can use natively
type Tool struct {
	Name        string
	Description string
	Schema      map[string]interface{}
	Execute     func(args map[string]interface{}) (string, error)
}

// ILLMProvider defines the interface for interacting with various LLM backends (OpenAI, Anthropic, Gemini, Local)
type ILLMProvider interface {
	Chat(ctx context.Context, messages []Message, tools []Tool) (Message, error)
	Stream(ctx context.Context, messages []Message, tools []Tool, chunkChan chan<- string) error
	GetModelName() string
}

// IAgent is the core interface for all TormentNexus Agents, enforcing parity with the TS core
type IAgent interface {
	GetName() string
	GetRole() string
	HandleInput(ctx context.Context, input string) (string, error)
	InjectSystemContext(context string)
	GetState() map[string]interface{}
}

// ModelSelector interface matches the CoreModelSelector from TS
type ModelSelector interface {
	SelectModel(taskType string, costPreference string) ILLMProvider
}
