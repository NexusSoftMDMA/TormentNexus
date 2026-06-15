package tools

import (
	"context"
	"io"
	"net/http"
)

func ListWorkshops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:3001/api/workshops")
	if e != nil {
		return err("failed to fetch workshops")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func GetWorkshop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.Get("http://localhost:3001/api/workshops/" + name)
	if e != nil {
		return err("failed to fetch workshop")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}