package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

func ok(text string) ToolResponse {
    return ToolResponse{
}
        TextContent: text,
        Ok:          true,
        Err:         nil,
    }
}

func err(e error) ToolResponse {
    return ToolResponse{
}
        TextContent: "",
        Ok:          false,
        Err:         e,
    }
}

func getString(args map[string]interface{}, key string) string {
    if value, found := args[key]; found {
        if str, found := value.(string); found {
            return str
        }
    }
    return ""
}

func getInt(args map[string]interface{}, key string) (int, bool) {
    if value, found := args[key]; found {
        if intVal, found := value.(int); found {
            return intVal, true
        }
    }
    return 0, false
}

func getBool(args map[string]interface{}, key string) (bool, bool) {
    if value, found := args[key]; found {
        if boolVal, found := value.(bool); found {
            return boolVal, true
        }
    }
    return false, false
}

type TextContent string

var marketplace Marketplace

func init() {
    marketplaceData := `{
        "name": "levnikolaevich-skills-marketplace",
        "owner": {
            "name": "Lev Nikolaevich",
            "email": "levnikolaevich.com@gmail.com",
            "url": "https://github.com/levnikolaevich"
        },
        "metadata": {
            "description": "Agile workflow (Planning, Execution, Quality) + Documentation pipeline + Codebase audit suite + Project Bootstrap (scaffolding, Docker, CI/CD) + Optimization suite (Performance, Dependencies, Modernization) + Community engagement (Announcements, RFC debates, Triage) + Setup environment (agent install, MCP config, settings sync).",
            "version": "2026.05.06-plugin-first"
        },
        "plugins": [
            // ... (same as original)
        ]
    }`
    var parseErr error
    if parseErr = json.Unmarshal([]byte(marketplaceData), &marketplace); parseErr != nil {
        panic(parseErr)

}

func HandleGetMarketplaceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("Marketplace Name: %s\n", marketplace.Name))
    sb.WriteString(fmt.Sprintf("Owner: %s (%s) - %s\n", marketplace.Owner.Name, marketplace.Owner.Email, marketplace.Owner.URL))
    sb.WriteString(fmt.Sprintf("Version: %s\n", marketplace.Metadata.Version))
    sb.WriteString(fmt.Sprintf("Description: %s\n", marketplace.Metadata.Description))
    sb.WriteString(fmt.Sprintf("Total Plugins: %d\n", len(marketplace.Plugins)))
    return ok(sb.String())
}

func HandleListPlugins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    var sb strings.Builder
    sb.WriteString("Available Plugins:\n\n")
    for _, plugin := range marketplace.Plugins {
        sb.WriteString(fmt.Sprintf("- Name: %s\n", plugin.Name))
        sb.WriteString(fmt.Sprintf("  Category: %s\n", plugin.Category))
        sb.WriteString(fmt.Sprintf("  Description: %s\n", plugin.Description))
        sb.WriteString(fmt.Sprintf("  Skills Count: %d\n\n", len(plugin.Skills)))

    return ok(sb.String())
}

func HandleGetPluginDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    pluginName, _ :=getString(args, "plugin_name")
    if pluginName == "" {
        return err("plugin_name parameter is required")
    }
    var foundPlugin *Plugin
    for _, plugin := range marketplace.Plugins {
        if plugin.Name == pluginName {
            foundPlugin = &plugin
            break
        }
    }
    if foundPlugin == nil {
        return err(fmt.Sprintf("plugin '%s' not found", pluginName))
    }
    var sb strings.Builder
    if len(foundPlugin.Skills) > 0 {
        sb.WriteString(fmt.Sprintf("Plugin: %s\n", foundPlugin.Name))
        sb.WriteString(fmt.Sprintf("Category: %s\n", foundPlugin.Category))
        sb.WriteString(fmt.Sprintf("Description: %s\n", foundPlugin.Description))
        sb.WriteString("Skills:\n")
        for _, skill := range foundPlugin.Skills {
            sb.WriteString(fmt.Sprintf("- %s\n", skill))

    } else {
        return err("No skills found for this plugin")
    }
    return ok(sb.String())
}

// ... (other handlers as per original code)
}