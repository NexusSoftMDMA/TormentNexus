package mcp

import "strings"

type MetadataTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
	AlwaysOn    bool        `json:"alwaysOn,omitempty"`
}

func MetadataToolsFromAny(raw any) []MetadataTool {
	items, _ := raw.([]any)
	if len(items) == 0 {
		return []MetadataTool{}
	}
	tools := make([]MetadataTool, 0, len(items))
	for _, item := range items {
		toolMap, _ := item.(map[string]any)
		name := strings.TrimSpace(stringFromAny(toolMap["name"]))
		if name == "" {
			continue
		}
		tools = append(tools, MetadataTool{
			Name:        name,
			Description: stringFromAny(toolMap["description"]),
			InputSchema: toolMap["inputSchema"],
			AlwaysOn:    boolFromAny(toolMap["alwaysOn"]),
		})
	}
	return tools
}

func MetadataToolsToAny(tools []MetadataTool) []any {
	if len(tools) == 0 {
		return []any{}
	}
	result := make([]any, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		item := map[string]any{
			"name":        name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
		if tool.AlwaysOn {
			item["alwaysOn"] = true
		}
		result = append(result, item)
	}
	return result
}

func ToolEntryFromMetadata(serverName string, tool MetadataTool) ToolEntry {
	serverName = strings.TrimSpace(serverName)
	name := strings.TrimSpace(tool.Name)
	advertisedName := name
	if serverName != "" {
		advertisedName = serverName + "__" + name
	}
	return ToolEntry{
		Name:              advertisedName,
		Description:       tool.Description,
		Server:            serverName,
		ServerDisplayName: serverName,
		AdvertisedName:    advertisedName,
		OriginalName:      name,
		AlwaysOn:          tool.AlwaysOn,
		InputSchema:       tool.InputSchema,
	}
}

func stringFromAny(value any) string {
	stringValue, _ := value.(string)
	return stringValue
}

func boolFromAny(value any) bool {
	boolValue, _ := value.(bool)
	return boolValue
}
