package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func HandleAnalyzeSql(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	parts := strings.Split(query, ";")
	count := 0
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	result, e := json.Marshal(map[string]int{"statements": count})
	if e != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", e))
}

	return ok(string(result))
}

func HandleValidateSql(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if strings.TrimSpace(query) == "" {
		return err("query is empty")
	}
	return ok("query is valid")
}