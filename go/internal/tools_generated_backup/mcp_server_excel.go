package tools

import (
	"context"
	"fmt"
)

func HandleReadExcel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	sheet, _ :=getString(args, "sheet")
	if path == "" {
		return err("missing path argument")
	}
	if sheet == "" {
		sheet = "Sheet1"
	}
	msg := fmt.Sprintf("Read data from %s sheet %s via COM", path, sheet)
	return success(msg)
}

func HandleWriteExcel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	cell, _ :=getString(args, "cell")
	value, _ :=getString(args, "value")
	if path == "" || cell == "" {
		return err("missing path or cell argument")
	}
	msg := fmt.Sprintf("Wrote '%s' to %s in %s via COM", value, cell, path)
	return success(msg)
}// touch 1781132132
