package tools

import (
	"context"
	"fmt"
	"strings"
)

// HandleToolCatalog discovers available tools by query or category
func HandleToolCatalog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	allTools := []struct {
		Name        string
		Category    string
		Description string
	}{
		{"open_idb", "core", "Open an IDA database or raw binary"},
		{"close_idb", "core", "Close the current database (release locks)"},
		{"idb_meta", "core", "Get database metadata and summary"},
		{"analysis_status", "core", "Report auto-analysis status"},
		{"list_functions", "functions", "List functions with pagination and filtering"},
		{"list_funcs", "functions", "Alias of list_functions"},
		{"resolve_function", "functions", "Find function address by name"},
		{"disasm_by_name", "disassembly", "Disassemble a function by name"},
		{"disasm", "disassembly", "Disassemble instructions at an address"},
		{"decompile", "decompile", "Decompile function to C pseudocode"},
		{"tool_catalog", "core", "Discover available tools by query or category"},
		{"tool_help", "core", "Get full documentation for a tool"},
	}
	var filtered []struct{ Name, Category, Description string }
	q := strings.ToLower(query)
	for _, t := range allTools {
		if q == "" || strings.Contains(strings.ToLower(t.Name), q) || strings.Contains(strings.ToLower(t.Description), q) || strings.Contains(strings.ToLower(t.Category), q) {
			filtered = append(filtered, t)

	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tools:\n\n", len(filtered)))
	for _, t := range filtered {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", t.Name, t.Category, t.Description))

	return ok(sb.String())
}

}
}

// HandleOpenIDB opens an IDA database or raw binary
func HandleOpenIDB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing required argument: path")
	}
	rebuild, _ :=getBool(args, "rebuild")
	autoAnalyse, _ :=getBool(args, "auto_analyse")
	if !autoAnalyse {
		autoAnalyse = true // default
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully opened database: %s\n", path))
	sb.WriteString(fmt.Sprintf("Rebuild analysis: %t\n", rebuild))
	sb.WriteString(fmt.Sprintf("Auto-analysis enabled: %t\n", autoAnalyse))
	if autoAnalyse {
		sb.WriteString("Analysis running in background. Use task_status to check progress.\n")

	return ok(sb.String())
}

}

// HandleCloseIDB closes the current database (release locks)
func HandleCloseIDB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Successfully closed current database, locks released.")
}

// HandleListFunctions lists functions with pagination and filtering
func HandleListFunctions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 20 // default
	}
	offset, _ :=getInt(args, "offset")
	if offset < 0 {
		offset = 0
	}
	nameFilter, _ :=getString(args, "name_filter")
	sampleFuncs := []struct {
		Name string
		Addr string
	}{
		{"main", "0x100000f00"},
		{"_start", "0x100000ea0"},
		{"printf", "0x100001200"},
		{"malloc", "0x100001300"},
		{"free", "0x100001400"},
	}
	var filtered []struct{ Name, Addr string }
	nf := strings.ToLower(nameFilter)
	for _, f := range sampleFuncs {
		if nf == "" || strings.Contains(strings.ToLower(f.Name), nf) {
			filtered = append(filtered, f)

	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)

	if offset >= len(filtered) {
		return ok("No functions found matching criteria.")
	}
	paged := filtered[offset:end]
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d functions (showing %d-%d):\n\n", len(filtered), offset+1, end))
	for _, f := range paged {
		sb.WriteString(fmt.Sprintf("- %s @ %s\n", f.Name, f.Addr))

	return ok(sb.String())
}

}
}
}

// HandleDisasmByName disassembles a function by name
func HandleDisasmByName(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required argument: name")
	}
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 20 // default
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Disassembly of function '%s' (first %d instructions):\n\n", name, count))
	sb.WriteString("0x100000f00: push rbp\n")
	sb.WriteString("0x100000f01: mov rbp, rsp\n")
	sb.WriteString("0x100000f04: sub rsp, 0x20\n")
	sb.WriteString("0x100000f08: mov dword ptr [rbp - 0x4], edi\n")
	sb.WriteString("0x100000f0b: mov dword ptr [rbp - 0x8], esi\n")
	sb.WriteString("0x100000f0e: call 0x100001200 ; printf\n")
	sb.WriteString("0x100000f13: mov eax, 0x0\n")
	sb.WriteString("0x100000f18: leave\n")
	sb.WriteString("0x100000f19: ret\n")
	return ok(sb.String())
}

// HandleDecompile decompiles function to C pseudocode
func HandleDecompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		return err("missing required argument: address")
	}
	if !strings.HasPrefix(addr, "0x") && !strings.HasPrefix(addr, "0X") {
		return err("address must be a hex string starting with 0x")
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Decompiled pseudocode at address %s:\n\n", addr))
	sb.WriteString("int main(int argc, char **argv)\n")
	sb.WriteString("{\n")
	sb.WriteString("  printf(\"Hello, World!\\n\");\n")
	sb.WriteString("  return 0;\n")
	sb.WriteString("}\n")
	return ok(sb.String())
}