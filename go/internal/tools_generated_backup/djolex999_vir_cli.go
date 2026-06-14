package tools

import (
	"context"
	"os"
	"path/filepath"
)

func HandleWikiAddEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vault, _ :=getString(args, "vault_path")
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if vault == "" || title == "" || content == "" {
		return err("missing required args: vault_path, title, content")
}

	path := filepath.Join(vault, title+".md")
	dir := filepath.Dir(path)
	if e := os.MkdirAll(dir, 0755); e != nil {
		return err("failed to create directory: " + e.Error())
}

	if e := os.WriteFile(path, []byte(content), 0644); e != nil {
		return err("failed to write file: " + e.Error())
}

	return success("wiki entry created at " + path)
}

func HandleWikiReadEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vault, _ :=getString(args, "vault_path")
	title, _ :=getString(args, "title")
	if vault == "" || title == "" {
		return err("missing required args: vault_path, title")
}

	path := filepath.Join(vault, title+".md")
	data, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
}

	return success(string(data))
}// touch 1781132125
