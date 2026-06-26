package tools

import (
    "context"
    "encoding/json"
    "fmt"
)

func HandleSkybridgeVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Skybridge MCP module (Go-native) v0.1.0")
}

func HandleSkybridgeListViews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    views := []string{"HomeView", "DashboardView", "SettingsView"}
    data, jsonErr := json.Marshal(views)
    if jsonErr != nil {
        return err(fmt.Sprintf("failed to marshal views: %v", jsonErr))
}

    return ok(string(data))
}