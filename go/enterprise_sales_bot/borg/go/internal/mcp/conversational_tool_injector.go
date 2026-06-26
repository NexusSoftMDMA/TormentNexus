package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ConversationalToolInjector injects tool schemas into conversation contexts
// to enable dynamic tool selection based on conversation intent.
type ConversationalToolInjector struct {
	catalog   *ToolCatalog
	inventory *CachedInventory
}

// NewConversationalToolInjector creates a new conversational tool injector.
func NewConversationalToolInjector(catalog *ToolCatalog, inventory *CachedInventory) *ConversationalToolInjector {
	return &ConversationalToolInjector{
		catalog:   catalog,
		inventory: inventory,
	}
}

// InjectedTool represents a tool that was injected into the conversation context.
type InjectedTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema,omitempty"`
	Server      string      `json:"server"`
	Score       float64     `json:"score"`
}

// InjectForContext selects and returns the most relevant tools for a given conversation context.
// It ranks tools by semantic similarity to the conversation topic.
func (cti *ConversationalToolInjector) InjectForContext(ctx context.Context, conversation string, maxTools int) ([]InjectedTool, error) {
	if maxTools <= 0 {
		maxTools = 10
	}

	// Get current inventory
	snapshot, err := cti.inventory.GetSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// Score tools based on keyword matching with conversation
	type scoredTool struct {
		tool  CachedMcpToolInventory
		score float64
	}

	var scored []scoredTool
	convLower := strings.ToLower(conversation)

	for _, tool := range snapshot.Tools {
		if !tool.AlwaysOn {
			continue
		}

		score := cti.scoreToolRelevance(tool, convLower)
		if score > 0 {
			scored = append(scored, scoredTool{tool: tool, score: score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Take top N
	if len(scored) > maxTools {
		scored = scored[:maxTools]
	}

	result := make([]InjectedTool, len(scored))
	for i, st := range scored {
		result[i] = InjectedTool{
			Name:        st.tool.Name,
			Description: st.tool.Description,
			InputSchema: st.tool.InputSchema,
			Server:      st.tool.Server,
			Score:       st.score,
		}
	}

	return result, nil
}

// scoreToolRelevance scores a tool's relevance to a conversation context.
func (cti *ConversationalToolInjector) scoreToolRelevance(tool CachedMcpToolInventory, convLower string) float64 {
	score := 0.0

	// Check tool name
	if strings.Contains(convLower, strings.ToLower(tool.Name)) {
		score += 40.0
	}

	// Check description
	descLower := strings.ToLower(tool.Description)
	for _, word := range strings.Fields(convLower) {
		if len(word) > 3 && strings.Contains(descLower, word) {
			score += 5.0
		}
	}

	// Check keywords
	for _, kw := range tool.Keywords {
		if strings.Contains(convLower, strings.ToLower(kw)) {
			score += 10.0
		}
	}

	// Boost exact matches
	for _, kw := range tool.Keywords {
		if strings.Contains(convLower, strings.ToLower(kw)) {
			score += 15.0
		}
	}

	return score
}

// InjectToolsForPrompt constructs a JSON schema fragment for the given tools.
func InjectToolsForPrompt(tools []InjectedTool) string {
	if len(tools) == 0 {
		return "{}"
	}

	schema := map[string]interface{}{
		"tools": tools,
	}

	bytes, _ := json.MarshalIndent(schema, "", "  ")
	return string(bytes)
}
