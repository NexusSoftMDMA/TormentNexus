package compat

import "encoding/json"

// ParityLevel tracks how close a native implementation is to an upstream tool contract.
type ParityLevel string

const (
	ParityPlanned  ParityLevel = "planned"
	ParityBridged  ParityLevel = "bridged"
	ParitySpeced   ParityLevel = "speced"
	ParityNative   ParityLevel = "native"
	ParityVerified ParityLevel = "verified"
)

// ResultContract documents the expected observable output shape for a tool.
type ResultContract struct {
	Format        string   `json:"format"`
	Deterministic bool     `json:"deterministic"`
	Notes         []string `json:"notes,omitempty"`
}

// ToolContract describes a single model-facing tool surface that must remain stable.
type ToolContract struct {
	Source            string          `json:"source"`
	Name              string          `json:"name"`
	Description       string          `json:"description,omitempty"`
	Parameters        json.RawMessage `json:"parameters,omitempty"`
	Result            ResultContract  `json:"result"`
	ExactName         bool            `json:"exactName"`
	ExactParameters   bool            `json:"exactParameters"`
	ExactResultShape  bool            `json:"exactResultShape"`
	Status            ParityLevel     `json:"status"`
	ImplementationRef string          `json:"implementationRef,omitempty"`
	Notes             []string        `json:"notes,omitempty"`
}

// Clone returns a safe copy for callers.
func (c ToolContract) Clone() ToolContract {
	out := c
	if c.Parameters != nil {
		out.Parameters = append(json.RawMessage(nil), c.Parameters...)
	}
	if c.Result.Notes != nil {
		out.Result.Notes = append([]string(nil), c.Result.Notes...)
	}
	if c.Notes != nil {
		out.Notes = append([]string(nil), c.Notes...)
	}
	return out
}
