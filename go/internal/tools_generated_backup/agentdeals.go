package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type deal struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price"`
	Category string `json:"category"`
}

func HandleGetAgentDeals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	deals := []deal{
		{ID: "1", Title: "Summer Special", Price: 99, Category: "Travel"},
		{ID: "2", Title: "Electronics Bundle", Price: 149, Category: "Tech"},
		{ID: "3", Title: "Gourmet Dinner", Price: 75, Category: "Food"},
	}
	if category != "" {
		var filtered []deal
		for _, d := range deals {
			if d.Category == category {
				filtered = append(filtered, d)

		}
		deals = filtered
	}
	if limit < len(deals) {
		deals = deals[:limit]
	}
	jsonBytes, e := json.Marshal(deals)
	if e != nil {
		return err("failed to marshal deals")
	}
	return success(fmt.Sprintf("Found %d deals:\n%s", len(deals), string(jsonBytes)))
}
}