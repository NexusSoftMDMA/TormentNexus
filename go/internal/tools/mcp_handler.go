//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleVercelList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("VERCEL_TOKEN")
	if token == "" {
		return err("VERCEL_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.vercel.com/v9/projects", nil)
	if e != nil {
		return err("create request")
}


-reasoner (deepseek)*
}