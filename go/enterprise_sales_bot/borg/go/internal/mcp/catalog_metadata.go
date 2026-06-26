package mcp

import (
	"strings"
)

// DeriveSemanticCatalogInput is the input for semantic catalog derivation.
type DeriveSemanticCatalogInput struct {
	ServerName  string
	Description string
	AlwaysOn    bool
	Tools       []DeriveSemanticToolInput
}

// DeriveSemanticToolInput represents a tool for semantic catalog derivation.
type DeriveSemanticToolInput struct {
	Name        string
	Title       string
	Description string
	InputSchema interface{}
	AlwaysOn    bool
}

// DeriveSemanticCatalogOutput is the result of semantic catalog derivation.
type DeriveSemanticCatalogOutput struct {
	ServerDisplayName string
	ServerTags        []string
	Tools             []DeriveSemanticToolOutput
	AlwaysOn          bool
}

// DeriveSemanticToolOutput is the result for a single tool's semantic derivation.
type DeriveSemanticToolOutput struct {
	Name               string
	ServerDisplayName  string
	ServerTags         []string
	ToolTags           []string
	SemanticGroup      string
	SemanticGroupLabel string
	AdvertisedName     string
	Keywords           []string
	AlwaysOn           bool
}

// semanticGroupIndicators maps keyword patterns to semantic groups.
var semanticGroupIndicators = []struct {
	Keywords []string
	Group    string
	Label    string
}{
	{[]string{"search", "query", "find", "lookup", "retrieve"}, "search-retrieval", "search & retrieval"},
	{[]string{"create", "write", "generate", "produce", "make"}, "content-generation", "content generation"},
	{[]string{"edit", "update", "modify", "change", "patch"}, "content-editing", "content editing"},
	{[]string{"delete", "remove", "clear", "destroy"}, "content-deletion", "content deletion"},
	{[]string{"read", "get", "fetch", "load", "open"}, "data-access", "data access"},
	{[]string{"list", "enumerate", "show", "display"}, "data-listing", "data listing"},
	{[]string{"analyze", "analyze", "evaluate", "assess", "score"}, "analysis", "analysis"},
	{[]string{"code", "compile", "build", "run", "execute"}, "code-execution", "code execution"},
	{[]string{"file", "directory", "path", "fs"}, "filesystem", "filesystem"},
	{[]string{"db", "database", "sql", "query", "table"}, "database", "database"},
	{[]string{"git", "commit", "branch", "repo", "pull"}, "version-control", "version control"},
	{[]string{"docker", "container", "image", "compose"}, "container-management", "container management"},
	{[]string{"network", "http", "api", "request", "url"}, "networking", "networking"},
	{[]string{"auth", "login", "token", "key", "credential"}, "authentication", "authentication"},
	{[]string{"memory", "store", "recall", "remember"}, "memory-management", "memory management"},
	{[]string{"plan", "task", "todo", "ticket"}, "task-management", "task management"},
}

// DeriveSemanticCatalogForServer derives semantic metadata for a server and its tools.
func DeriveSemanticCatalogForServer(input DeriveSemanticCatalogInput) DeriveSemanticCatalogOutput {
	serverDisplayName := input.ServerName
	if input.Description != "" {
		serverDisplayName = input.ServerName
	}

	var serverTags []string

	derivedTools := make([]DeriveSemanticToolOutput, 0, len(input.Tools))
	for _, tool := range input.Tools {
		derived := deriveToolMetadata(tool, input.ServerName)
		serverTags = append(serverTags, derived.ServerTags...)
		derivedTools = append(derivedTools, derived)
	}

	serverTags = uniqueStrings(serverTags)

	return DeriveSemanticCatalogOutput{
		ServerDisplayName: serverDisplayName,
		ServerTags:        serverTags,
		Tools:             derivedTools,
		AlwaysOn:          input.AlwaysOn,
	}
}

func deriveToolMetadata(tool DeriveSemanticToolInput, serverName string) DeriveSemanticToolOutput {
	displayName := tool.Title
	if displayName == "" {
		displayName = tool.Name
	}

	// Derive semantic group from tool name and description
	group, label := deriveSemanticGroup(tool)

	// Derive keywords
	keywords := deriveKeywords(tool)

	// Derive tags
	tags := deriveTags(tool, group)

	return DeriveSemanticToolOutput{
		Name:               tool.Name,
		ServerDisplayName:  displayName,
		ServerTags:         tags[:min(len(tags), 5)],
		ToolTags:           tags,
		SemanticGroup:      group,
		SemanticGroupLabel: label,
		AdvertisedName:     NamespaceToolName(serverName, tool.Name),
		Keywords:           keywords,
		AlwaysOn:           tool.AlwaysOn,
	}
}

func deriveSemanticGroup(tool DeriveSemanticToolInput) (string, string) {
	text := strings.ToLower(tool.Name + " " + tool.Description + " " + tool.Title)

	for _, indicator := range semanticGroupIndicators {
		for _, kw := range indicator.Keywords {
			if strings.Contains(text, kw) {
				return indicator.Group, indicator.Label
			}
		}
	}

	return "general-utility", "general utility"
}

func deriveKeywords(tool DeriveSemanticToolInput) []string {
	var keywords []string

	// Add words from name
	for _, word := range strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(tool.Name)) {
		word = strings.TrimSpace(word)
		if len(word) > 2 {
			keywords = append(keywords, strings.ToLower(word))
		}
	}

	// Add words from description
	if tool.Description != "" {
		words := strings.Fields(tool.Description)
		for _, w := range words {
			w = strings.Trim(strings.ToLower(w), ".,!?;:")
			if len(w) > 3 && !contains(keywords, w) {
				keywords = append(keywords, w)
				if len(keywords) >= 15 {
					break
				}
			}
		}
	}

	return keywords
}

func deriveTags(tool DeriveSemanticToolInput, group string) []string {
	tags := []string{group}

	// Extract potential tags from name
	parts := strings.Split(tool.Name, "_")
	if len(parts) == 1 {
		parts = strings.Split(tool.Name, "-")
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 2 && !contains(tags, strings.ToLower(p)) {
			tags = append(tags, strings.ToLower(p))
		}
	}

	return tags
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
