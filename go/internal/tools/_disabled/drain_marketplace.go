package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleListDrains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	drains := []string{"Drain A", "Drain B", "Drain C"}
	if category != "" {
		drains = []string{"Drain " + category}
	}
	data, e := json.Marshal(drains)
	if e != nil {
		return err("failed to marshal drains")
}

	return ok(string(data))
}

func HandleGetDrain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	detail := fmt.Sprintf("Drain %s is available for $10", id)
	data, e := json.Marshal(map[string]string{"id": id, "detail": detail})
	if e != nil {
		return err("failed to marshal detail")
}

	return ok(string(data))
}