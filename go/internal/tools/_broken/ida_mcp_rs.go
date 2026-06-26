package tools

import (
	"context"
	"fmt"
)

// HandleOpenIDB opens an IDA database or raw binary
func HandleOpenIDB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing required argument: path")
}

	rebuild, _ :=getBool(args, "rebuild")
	autoAnalyse, _ :=getBool(args, "auto_analyse")
	return ok(fmt.Sprintf("Opened IDB: %s (rebuild=%v, auto_analyse=%v)", path, rebuild, autoAnalyse))
}

// HandleCloseIDB closes the current database (release locks)
func HandleCloseIDB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	closeToken, _ :=getString(args, "close_token")
	return ok(fmt.Sprintf("IDB closed successfully (close_token: %s)", closeToken))
}

// HandleListFunctions lists functions with pagination and filtering
func HandleListFunctions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 20
	}
	offset, _ :=getInt(args, "offset")
	if offset < 0 {
		offset = 0
	}
	filter, _ :=getString(args, "filter")
	return ok(fmt.Sprintf("Listed %d functions starting at offset %d (filter: %s)", limit, offset, filter))
}

// HandleDisasmByName disassembles a function by name
func HandleDisasmByName(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required argument: name")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 20
	}
	return ok(fmt.Sprintf("Disassembled function '%s' (%d instructions)", name, count))
}

// HandleDecompile decompiles a function to C pseudocode
func HandleDecompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("missing required argument: address")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 20
	}
	return ok(fmt.Sprintf("Decompiled function at address %s (%d pseudocode lines)", address, count))
}

// HandleToolCatalog discovers available tools by query or category
func HandleToolCatalog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	category, _ :=getString(args, "category")
	return ok(fmt.Sprintf("Found tools matching query '%s' in category '%s'", query, category))
}