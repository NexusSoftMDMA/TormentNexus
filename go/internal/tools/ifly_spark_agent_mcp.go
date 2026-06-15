//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleSparkAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	resp, e := http.DefaultClient.Get("https://api.spark.com/chat?prompt=" + prompt)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	return ok("Spark response: " +
}


-reasoner (deepseek)*
}