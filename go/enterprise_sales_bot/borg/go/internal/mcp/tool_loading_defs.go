package mcp

// ToolLoadingDef provides metadata for loading a tool from an MCP server.
type ToolLoadingDef struct {
	ServerName         string   `json:"serverName"`
	ToolName           string   `json:"toolName"`
	NamespacedName     string   `json:"namespacedName"`
	RequiredCapability string   `json:"requiredCapability,omitempty"`
	LoadPriority       int      `json:"loadPriority"` // higher = loaded first
	Tags               []string `json:"tags"`
	AlwaysOn           bool     `json:"alwaysOn"`
}

// ToolLoadingDefinitions manages definitions for how tools should be loaded.
type ToolLoadingDefinitions struct {
	defs map[string]*ToolLoadingDef // keyed by namespaced name
}

// NewToolLoadingDefinitions creates a new tool loading definitions manager.
func NewToolLoadingDefinitions() *ToolLoadingDefinitions {
	return &ToolLoadingDefinitions{
		defs: make(map[string]*ToolLoadingDef),
	}
}

// Register adds a tool loading definition.
func (tld *ToolLoadingDefinitions) Register(def *ToolLoadingDef) {
	if def.NamespacedName == "" {
		def.NamespacedName = NamespaceToolName(def.ServerName, def.ToolName)
	}
	tld.defs[def.NamespacedName] = def
}

// Get returns the loading definition for a tool.
func (tld *ToolLoadingDefinitions) Get(namespacedName string) *ToolLoadingDef {
	return tld.defs[namespacedName]
}

// List returns all registered definitions.
func (tld *ToolLoadingDefinitions) List() []*ToolLoadingDef {
	defs := make([]*ToolLoadingDef, 0, len(tld.defs))
	for _, d := range tld.defs {
		defs = append(defs, d)
	}
	return defs
}

// ListAlwaysOn returns definitions for tools marked as always-on.
func (tld *ToolLoadingDefinitions) ListAlwaysOn() []*ToolLoadingDef {
	var result []*ToolLoadingDef
	for _, d := range tld.defs {
		if d.AlwaysOn {
			result = append(result, d)
		}
	}
	return result
}

// Remove deletes a tool loading definition.
func (tld *ToolLoadingDefinitions) Remove(namespacedName string) {
	delete(tld.defs, namespacedName)
}

// Len returns the number of registered definitions.
func (tld *ToolLoadingDefinitions) Len() int {
	return len(tld.defs)
}
