package tools

import (
	"context"
	"os"
	"os/exec"
)

func HandleReason(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	program, _ :=getString(args, "program")
	query, _ :=getString(args, "query")
	if program == "" || query == "" {
		return err("program and query are required")
}

	tmpFile, e := os.CreateTemp("", "prolog-*.pl")
	if e != nil {
		return err("failed to create temp file: " + e.Error())
}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	if _, e = tmpFile.WriteString(program); e != nil {
		return err("failed to write temp file: " + e.Error())
}

	output, e := exec.Command("swipl", "-f", tmpFile.Name(), "-g", query, "-t", "halt").CombinedOutput()
	if e != nil {
		return err("swipl execution failed: " + e.Error())
}

	return ok(string(output))
}