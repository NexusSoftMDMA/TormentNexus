package mcpimpl

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func HandleGenerateTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	fset := token.NewFileSet()
	f, e := parser.ParseFile(fset, "", code, parser.ParseComments)
	if e != nil {
		return err("parse error: " + e.Error())
}

	var buf strings.Builder
	ast.Fprint(&buf, fset, f, nil)
	return ok(buf.String())
}