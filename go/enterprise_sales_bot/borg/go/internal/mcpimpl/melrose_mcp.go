package mcpimpl

import (
    "context"
    "fmt"
)

func HandleGetSong(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    return ok(fmt.Sprintf("Song '%s' found: artist Melrose, album 'Melrose Place', year 2024", name))
}

func HandleSearchSongs_melrose_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    artist, _ :=getString(args, "artist")
    if artist == "" {
        artist = "Melrose"
    }
    return ok(fmt.Sprintf("Found songs by %s: ['Melody', 'Rose', 'Mcp']", artist))
}