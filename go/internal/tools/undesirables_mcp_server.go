//go:build ignore
// +build ignore

package tools

import (
	"context"
	"strings"
)

var undesirables = map[string]bool{

func HandleCheckUndesirable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name := strings.ToLower(getString(args, "name"))
	if undesirables[name] {
		return ok("Name is undesirable")
	return ok("Name is not undesirable")
}

func HandleListUndesirables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	names := make([]string, 0, len(undesirables))
	for n := range undesirables {
		names = append(names, n)

	return success(strings.Join(names, ", "))
}
}
}
}