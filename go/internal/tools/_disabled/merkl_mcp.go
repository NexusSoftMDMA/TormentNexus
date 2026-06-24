package tools

import (
	"context"
	"encoding/json"
)

func HandleListOpportunities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	opps := []map[string]interface{}{
		{"id": "1", "name": "Opportunity A", "apr": 15.5},
		{"id": "2", "name": "Opportunity B", "apr": 12.3},
	}
	data, e := json.Marshal(opps)
	if e != nil {
		return err("failed to marshal opportunities")
	}
	return ok(string(data))
}

func HandleGetOpportunity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id parameter")
	}
	opp := map[string]interface{}{"id": id, "name": "Opportunity " + id, "apr": 10.0}
	data, e := json.Marshal(opp)
	if e != nil {
		return err("failed to marshal opportunity")
	}
	return ok(string(data))
}