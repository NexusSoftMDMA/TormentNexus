package compat

import "encoding/json"

// DefaultCatalog seeds the current exact-name tool compatibility surface.
func DefaultCatalog() *Catalog {
	catalog := NewCatalog()
	contracts := []ToolContract{
		{
			Source:            "pi",
			Name:              "read",
			Description:       "Read file contents by path with optional line offsets.",
			Parameters:        json.RawMessage(`{"type":"object","required":["path"],"properties":{"path":{"type":"string"},"offset":{"type":"integer","minimum":1},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "write",
			Description:       "Create or overwrite a file with content.",
			Parameters:        json.RawMessage(`{"type":"object","required":["path","content"],"properties":{"path":{"type":"string"},"content":{"type":"string"}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "edit",
			Description:       "Apply exact text replacements to a file.",
			Parameters:        json.RawMessage(`{"type":"object","required":["path","edits"],"properties":{"path":{"type":"string"},"edits":{"type":"array","items":{"type":"object","required":["oldText","newText"],"properties":{"oldText":{"type":"string"},"newText":{"type":"string"}},"additionalProperties":false},"minItems":1}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "bash",
			Description:       "Execute a shell command with optional timeout seconds.",
			Parameters:        json.RawMessage(`{"type":"object","required":["command"],"properties":{"command":{"type":"string"},"timeout":{"type":"number","exclusiveMinimum":0}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "grep",
			Description:       "Search file contents for a pattern with file paths and line numbers. Respects .gitignore.",
			Parameters:        json.RawMessage(`{"type":"object","required":["pattern"],"properties":{"pattern":{"type":"string"},"path":{"type":"string"},"glob":{"type":"string"},"ignoreCase":{"type":"boolean"},"literal":{"type":"boolean"},"context":{"type":"integer","minimum":0},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "find",
			Description:       "Search for files by glob pattern. Respects .gitignore. Returns relative paths.",
			Parameters:        json.RawMessage(`{"type":"object","required":["pattern"],"properties":{"pattern":{"type":"string"},"path":{"type":"string"},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
		{
			Source:            "pi",
			Name:              "ls",
			Description:       "List directory contents sorted alphabetically with / suffix for directories. Includes dotfiles.",
			Parameters:        json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"limit":{"type":"integer","minimum":1}},"additionalProperties":false}`),
			Result:            ResultContract{Format: "tool-specific", Deterministic: false},
			ExactName:         true,
			ExactParameters:   true,
			ExactResultShape:  true,
			Status:            ParityNative,
			ImplementationRef: "foundation/pi.DefaultToolHandlers",
		},
	}
	for _, contract := range contracts {
		catalog.MustRegister(contract)
	}
	return catalog
}
