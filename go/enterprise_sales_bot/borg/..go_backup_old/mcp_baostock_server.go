package tools

import (
	"context"
)

func HandleGetStockData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	start, _ :=getString(args, "start_date")
	end, _ :=getString(args, "end_date")
	if code == "" || start == "" || end == "" {
		return err("missing required arguments: code, start_date, end_date")
}

	data := map[string]interface{}{
		"code": code,
		"start": start,
		"end": end,
		"data": []map[string]interface{}{
			{"date": "2024-01-02", "close": 10.5},
			{"date": "2024-01-03", "close": 10.7},
		},
	}
	return ok(data)
}

func HandleGetStockInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("missing required argument: code")
}

	info := map[string]interface{}{
		"code":   code,
		"name":   "Mock Stock",
		"market": "Shanghai",
	}
	return ok(info)
}