package tools

import (
	"context"
	"fmt"
	"net/http"
	"io"
)

func HandleGetFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	animal, _ :=getString(args, "animal")
	if animal == "" {
		animal = "panther"
	}
	url := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&prop=extracts&explaintext=true&exintro=true&titles=%s&format=json", animal)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch fact")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body)[:200])
}

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Panther MCP server v1.0")
}