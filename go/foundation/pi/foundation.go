package pi

import (
	"encoding/json"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/compat"
)

type ThinkingLevel string

type MessageDeliveryMode string

type TransportPreference string

type ToolExecutionMode string

type RunEventType string

const (
	ThinkingOff     ThinkingLevel = "off"
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
)

const (
	DeliveryOneAtATime MessageDeliveryMode = "one-at-a-time"
	DeliveryAll        MessageDeliveryMode = "all"
)

const (
	TransportAuto      TransportPreference = "auto"
	TransportSSE       TransportPreference = "sse"
	TransportWebSocket TransportPreference = "websocket"
)

const (
	ToolExecutionParallel   ToolExecutionMode = "parallel"
	ToolExecutionSequential ToolExecutionMode = "sequential"
)

const (
	EventAgentStart          RunEventType = "agent_start"
	EventTurnStart           RunEventType = "turn_start"
	EventMessageStart        RunEventType = "message_start"
	EventMessageUpdate       RunEventType = "message_update"
	EventMessageEnd          RunEventType = "message_end"
	EventToolExecutionStart  RunEventType = "tool_execution_start"
	EventToolExecutionUpdate RunEventType = "tool_execution_update"
	EventToolExecutionEnd    RunEventType = "tool_execution_end"
	EventTurnEnd             RunEventType = "turn_end"
	EventAgentEnd            RunEventType = "agent_end"
)

// ToolDescriptor is the model-facing tool contract used by the Go foundation.
type ToolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// AgentState mirrors pi's initial agent state shape at the contract level.
type AgentState struct {
	SystemPrompt  string           `json:"systemPrompt"`
	Model         string           `json:"model"`
	ThinkingLevel ThinkingLevel    `json:"thinkingLevel"`
	Tools         []ToolDescriptor `json:"tools"`
}

type ThinkingBudgets struct {
	Minimal int `json:"minimal,omitempty"`
	Low     int `json:"low,omitempty"`
	Medium  int `json:"medium,omitempty"`
	High    int `json:"high,omitempty"`
	XHigh   int `json:"xhigh,omitempty"`
}

type AgentConfig struct {
	InitialState    AgentState          `json:"initialState"`
	SteeringMode    MessageDeliveryMode `json:"steeringMode"`
	FollowUpMode    MessageDeliveryMode `json:"followUpMode"`
	Transport       TransportPreference `json:"transport"`
	ToolExecution   ToolExecutionMode   `json:"toolExecution"`
	ThinkingBudgets ThinkingBudgets     `json:"thinkingBudgets"`
}

type SessionConfig struct {
	AutoSave  bool   `json:"autoSave"`
	Ephemeral bool   `json:"ephemeral"`
	Directory string `json:"directory,omitempty"`
}

type FoundationSpec struct {
	Name             string         `json:"name"`
	Philosophy       string         `json:"philosophy"`
	Agent            AgentConfig    `json:"agent"`
	Session          SessionConfig  `json:"session"`
	RunEventSequence []RunEventType `json:"runEventSequence"`
	Features         []string       `json:"features"`
}

func DefaultFoundationSpec() FoundationSpec {
	return FoundationSpec{
		Name:       "pi-go-foundation",
		Philosophy: "Minimal terminal coding harness with exact model-facing tool contracts, strong extension seams, and native integration points for TormentNexus and TormentNexus.",
		Agent: AgentConfig{
			InitialState: AgentState{
				SystemPrompt:  "You are a helpful coding agent.",
				Model:         "provider/model",
				ThinkingLevel: ThinkingMinimal,
				Tools:         BuiltinTools(),
			},
			SteeringMode:  DeliveryOneAtATime,
			FollowUpMode:  DeliveryOneAtATime,
			Transport:     TransportAuto,
			ToolExecution: ToolExecutionParallel,
			ThinkingBudgets: ThinkingBudgets{
				Minimal: 128,
				Low:     512,
				Medium:  1024,
				High:    2048,
				XHigh:   4096,
			},
		},
		Session: SessionConfig{
			AutoSave: true,
		},
		RunEventSequence: []RunEventType{
			EventAgentStart,
			EventTurnStart,
			EventMessageStart,
			EventMessageUpdate,
			EventMessageEnd,
			EventToolExecutionStart,
			EventToolExecutionUpdate,
			EventToolExecutionEnd,
			EventTurnEnd,
			EventAgentEnd,
		},
		Features: []string{
			"interactive mode",
			"print/json mode",
			"rpc/daemon mode",
			"session branching",
			"session compaction",
			"extensions",
			"skills",
			"prompt templates",
			"themes",
			"exact tool contracts",
		},
	}
}

func BuiltinTools() []ToolDescriptor {
	return []ToolDescriptor{
		{
			Name:        "read",
			Description: "Read file contents by path with optional line offsets.",
			Parameters:  json.RawMessage(`{"type":"object","required":["path"],"properties":{"path":{"type":"string"},"offset":{"type":"integer","minimum":1},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
		},
		{
			Name:        "write",
			Description: "Create or overwrite a file with content.",
			Parameters:  json.RawMessage(`{"type":"object","required":["path","content"],"properties":{"path":{"type":"string"},"content":{"type":"string"}},"additionalProperties":false}`),
		},
		{
			Name:        "edit",
			Description: "Apply exact text replacements to a file.",
			Parameters:  json.RawMessage(`{"type":"object","required":["path","edits"],"properties":{"path":{"type":"string"},"edits":{"type":"array","items":{"type":"object","required":["oldText","newText"],"properties":{"oldText":{"type":"string"},"newText":{"type":"string"}},"additionalProperties":false},"minItems":1}},"additionalProperties":false}`),
		},
		{
			Name:        "bash",
			Description: "Execute a shell command with optional timeout seconds.",
			Parameters:  json.RawMessage(`{"type":"object","required":["command"],"properties":{"command":{"type":"string"},"timeout":{"type":"number","exclusiveMinimum":0}},"additionalProperties":false}`),
		},
		{
			Name:        "grep",
			Description: "Search file contents for a pattern. Returns matching lines with file paths and line numbers. Respects .gitignore. Output is truncated to 100 matches or 50KB (whichever is hit first). Long lines are truncated to 500 chars.",
			Parameters:  json.RawMessage(`{"type":"object","required":["pattern"],"properties":{"pattern":{"type":"string"},"path":{"type":"string"},"glob":{"type":"string"},"ignoreCase":{"type":"boolean"},"literal":{"type":"boolean"},"context":{"type":"integer","minimum":0},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
		},
		{
			Name:        "find",
			Description: "Search for files by glob pattern. Returns matching file paths relative to the search directory. Respects .gitignore. Output is truncated to 1000 results or 50KB (whichever is hit first).",
			Parameters:  json.RawMessage(`{"type":"object","required":["pattern"],"properties":{"pattern":{"type":"string"},"path":{"type":"string"},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
		},
		{
			Name:        "ls",
			Description: "List directory contents. Returns entries sorted alphabetically, with '/' suffix for directories. Includes dotfiles. Output is truncated to 500 entries or 50KB (whichever is hit first).",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
		},
	}
}

func BuiltinToolContracts() []compat.ToolContract {
	tools := BuiltinTools()
	contracts := make([]compat.ToolContract, 0, len(tools))
	for _, tool := range tools {
		contracts = append(contracts, compat.ToolContract{
			Source:            "pi",
			Name:              tool.Name,
			Description:       tool.Description,
			Parameters:        append(json.RawMessage(nil), tool.Parameters...),
			Result:            compat.ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            compat.ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
			Notes: []string{
				"These contracts are the initial exact-name compatibility surface.",
				"Runtime behavior should stay observationally compatible even when implementations evolve.",
			},
		})
	}
	return contracts
}
