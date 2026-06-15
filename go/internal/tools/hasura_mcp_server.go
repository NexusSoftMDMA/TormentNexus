//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleExecuteGraphQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	query, _ :=getString(args, "query")
	if endpoint == "" || query == "" {
		return err("endpoint and query are required")
}

	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", endpoint, io.N


-reasoner (deepseek)*
}