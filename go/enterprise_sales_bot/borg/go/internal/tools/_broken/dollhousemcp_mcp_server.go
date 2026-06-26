package mcpimpl

import (
    "context"
    "os"
    "path/filepath"
    "strings"
)

var currentPersona string

func HandleListPersonas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    dir, _ :=getString(args, "directory")
    if dir == "" {
        dir = "personas"
    }
    entries, e := os.ReadDir(dir)
    if e != nil {
        return err("failed to read personas directory: " + e.Error())
}

    var personas []string
    for _, entry := range entries {
        if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
            personas = append(personas, strings.TrimSuffix(entry.Name(), ".md"))

    }
    return ok("personas: " + strings.Join(personas, ", "))
}

}

func HandleActivatePersona(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    path := filepath.Join("personas", name+".md")
    if _, e := os.Stat(path); os.IsNotExist(e) {
        return err("persona not found: " + name)
}

    currentPersona = name
    return success("activated persona: " + name)
}