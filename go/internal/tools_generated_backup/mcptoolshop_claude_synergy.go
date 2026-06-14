package tools

import (
	"context"
	"fmt"
	"os"
)

func HandleGetChangelogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	product, _ :=getString(args, "product")
	data, e := os.ReadFile("changelogs.json")
	if e != nil {
		return err(fmt.Sprintf("failed to read changelogs: %v", e))
}

	if product != "" {
		// Filtering not implemented; return all for now.
	}
	return ok(string(data))
}

func HandleGetSynergies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	product, _ :=getString(args, "product")
	data, e := os.ReadFile("synergies.json")
	if e != nil {
		return err(fmt.Sprintf("failed to read synergies: %v", e))
}

	if product != "" {
		// Filtering not implemented; return all for now.
	}
	return ok(string(data))
}