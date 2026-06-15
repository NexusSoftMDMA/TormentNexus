//go:build ignore
// +build ignore

package tools

import (
    "context"
    "fmt"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    notes := []string{"Note 1", "Note 2", "Note 3"}
    return ok(fmt.Sprintf("Notes: %v", notes))
}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "id")
    if id == "" {
        return err("Missing id argument")
}

    content := fmt.Sprintf("Content of note %s: This is a sample note.", id)
    return ok(content)
}
