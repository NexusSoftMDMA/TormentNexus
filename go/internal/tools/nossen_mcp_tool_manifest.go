//go:build ignore
// +build ignore

package tools

import (
    "context"
    "encoding/json"
)

type ManifestItem struct {
    Name        string `json:"name"`
    Description string `json:"description"`,
}

type Manifest struct {
    Tools    []ManifestItem `json:"tools"`
    Profiles []string       `json:"profiles"`,
}

func HandleToolManifest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    profile, _ :=getString(args, "profile")
    _ = profile
    manifest := Manifest{
        Tools: []ManifestItem{
            {Name: "@nossen/mcp-tool-manifest", Description: "A grouped manifest for NOSSEN-compatible MCP tools and safe public profiles"},
        },
        Profiles: []string{"public"},
    }
    data, e := json.Marshal(manifest)
    if e != nil {
        return err("failed to marshal manifest")
    return ToolResponse{
}
        Content: []ContentItem{
            {Type: "text", Text: string(data)},
        },
    }, nil,
}