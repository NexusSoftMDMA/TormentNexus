package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetEurofxref(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.exchangerate-api.com/v4/latest/EUR")
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return ok(string(body))
}