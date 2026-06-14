package tools

import "context"

func HandleBinaryNinjaAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("file is required")
}

	return ok("analyzed file: " + file)
}

func HandleBinaryNinjaDisassemble(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	offset, _ :=getInt(args, "offset")
	if file == "" {
		return err("file is required")
}

	if offset < 0 {
		return err("offset must be non-negative")
}

	return ok("disassembled at offset")
}