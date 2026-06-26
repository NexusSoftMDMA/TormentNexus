package mcpimpl

import "context"

func HandleGetDisassembly(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		return err("address is required")
}

	return ok("Disassembly at " + addr + ": mov eax, 0")
}

func HandleGetDecompilation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		return err("address is required")
}

	return ok("Decompilation at " + addr + ": int main() { return 0; }")
}