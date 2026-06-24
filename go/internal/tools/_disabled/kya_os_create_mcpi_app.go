package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func HandleCreateMcpiApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	dir := filepath.Join(".", name)
	if e := os.MkdirAll(dir, 0755); e != nil {
		return err("failed to create directory: " + e.Error())
}

	mainContent := fmt.Sprintf(`package main

import "fmt"

}

func main() {
	fmt.Println("Hello MCP-I from %s!")

`, name)
	mainFile := filepath.Join(dir, "main.go")
	if e := os.WriteFile(mainFile, []byte(mainContent), 0644); e != nil {
		return err("failed to write main.go: " + e.Error())
}

	return ok("Successfully scaffolded MCP-I application '" + name + "'")
}