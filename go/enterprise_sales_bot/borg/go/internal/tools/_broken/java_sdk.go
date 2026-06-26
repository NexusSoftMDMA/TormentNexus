package mcpimpl

import (
	"bytes"
	"context"
	"os/exec"
)

func HandleJavaVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "java", "-version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	e := cmd.Run()
	if e != nil {
		return err("failed to get Java version: " + e.Error())
}

	return success(stderr.String())
}

func HandleJavaCompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("missing 'source' argument")
}

	tmpFile, e := os.CreateTemp("", "*.java")
	if e != nil {
		return err("failed to create temp file: " + e.Error())
}

	defer os.Remove(tmpFile.Name())
	_, e = tmpFile.WriteString(source)
	if e != nil {
		return err("failed to write source: " + e.Error())
}

	tmpFile.Close()
	cmd := exec.CommandContext(ctx, "javac", tmpFile.Name())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	e = cmd.Run()
	if e != nil {
		return err("compilation failed: " + out.String())
}

	return success("compilation successful")
}