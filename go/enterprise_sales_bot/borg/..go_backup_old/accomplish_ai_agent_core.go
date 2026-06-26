package tools

import (
    "context"
    "encoding/json"
)

func HandleListProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    providers := []string{"openai", "anthropic", "local"}
    data, _ := json.Marshal(providers)
    return success(string(data))
}

func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    skills := []string{"write_code", "analyze_data", "summarize"}
    data, _ := json.Marshal(skills)
    return success(string(data))
}