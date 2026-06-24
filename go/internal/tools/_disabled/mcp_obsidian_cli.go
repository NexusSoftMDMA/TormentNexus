package tools

import (
	"context"
	"os"
)

func HandleSaveNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	title, _ :=getString(args, "title")
	vaultPath, _ :=getString(args, "vault_path")
	if title == "" {
		title = "Untitled"
	}
	filePath := vaultPath + "/" + title + ".md"
	e := os.WriteFile(filePath, []byte(content), 0644)
	if e != nil {
		return err("Failed to save note: " + e.Error())
}

	return ok("Note saved to " + filePath)
}// touch 1781132132
