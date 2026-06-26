package mcp

// CompatibilityToolDef provides metadata for compatibility tools.
type CompatibilityToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// CompatToolName enumerates known compatibility tool names.
type CompatToolName string

const (
	CompatRunCode         CompatToolName = "run_code"
	CompatRunPython       CompatToolName = "run_python"
	CompatRunAgent        CompatToolName = "run_agent"
	CompatSaveMemory      CompatToolName = "save_memory"
	CompatSearchMemory    CompatToolName = "search_memory"
	CompatSaveScript      CompatToolName = "save_script"
	CompatSaveToolSet     CompatToolName = "save_tool_set"
	CompatLoadToolSet     CompatToolName = "load_tool_set"
	CompatToolsetList     CompatToolName = "toolset_list"
	CompatImportMCPConfig CompatToolName = "import_mcp_config"
	CompatAutoCallTool    CompatToolName = "auto_call_tool"
)

// baseCompatibilityToolDefs returns the standard definitions for all compatibility tools.
func baseCompatibilityToolDefs() map[CompatToolName]CompatibilityToolDef {
	return map[CompatToolName]CompatibilityToolDef{
		CompatRunCode: {
			Name:        string(CompatRunCode),
			Description: "Execute TypeScript/JavaScript code in a secure sandbox. Use this to chain multiple tool calls, process data, or perform logic.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The TypeScript/JavaScript code to execute. Top-level await is supported.",
					},
				},
				"required": []string{"code"},
			},
		},
		CompatRunPython: {
			Name:        string(CompatRunPython),
			Description: "Execute Python 3 code. Suitable for data processing or simple scripts.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The Python 3 code to execute.",
					},
				},
				"required": []string{"code"},
			},
		},
		CompatSaveMemory: {
			Name:        string(CompatSaveMemory),
			Description: "Save a fact, observation, or memory to the persistent memory store for recall in future sessions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Memory key/identifier.",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Memory content to store.",
					},
				},
				"required": []string{"key", "value"},
			},
		},
		CompatSearchMemory: {
			Name:        string(CompatSearchMemory),
			Description: "Search the persistent memory store for relevant facts, observations, or memories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query.",
					},
				},
				"required": []string{"query"},
			},
		},
		CompatSaveScript: {
			Name:        string(CompatSaveScript),
			Description: "Save a reusable script for later execution.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Script name.",
					},
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Script content.",
					},
				},
				"required": []string{"name", "code"},
			},
		},
		CompatImportMCPConfig: {
			Name:        string(CompatImportMCPConfig),
			Description: "Import an MCP server configuration from a JSON object.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"config": map[string]interface{}{
						"type": "object",
					},
				},
				"required": []string{"config"},
			},
		},
	}
}

// GetCompatibilityToolDef returns the definition for a named compatibility tool.
func GetCompatibilityToolDef(name CompatToolName) *CompatibilityToolDef {
	defs := baseCompatibilityToolDefs()
	if d, ok := defs[name]; ok {
		return &d
	}
	return nil
}

// ListCompatibilityToolNames returns all known compatibility tool names.
func ListCompatibilityToolNames() []CompatToolName {
	defs := baseCompatibilityToolDefs()
	names := make([]CompatToolName, 0, len(defs))
	for n := range defs {
		names = append(names, n)
	}
	return names
}
