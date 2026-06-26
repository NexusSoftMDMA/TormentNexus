package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleAddMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memory, _ :=getString(args, "memory")
	if memory == "" {
		return err("memory is required")
}

	body := fmt.Sprintf(`{"memory":"%s"}`, memory)
	resp, e := http.DefaultClient.Post("http://localhost:3000/memories", "application/json", strings.NewReader(body))
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("failed to add memory")
}

	return ok("memory added")
}

func HandleSearchMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("http://localhost:3000/memories?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(data))
}