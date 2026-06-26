package mcpimpl

import (
	"context"
	"fmt"
	"time"
)

func HandleGetCost(ctx context.Context, args map[string]string) (ToolResponse, error) {
	granularity, _ :=getString(args, "granularity")
	start, _ :=getString(args, "start_date")
	end, _ :=getString(args, "end_date")
	if start == "" || end == "" {
		return err("start_date and end_date are required")
	}
	_, e := time.Parse("2006-01-02", start)
	if e != nil {
		return err("invalid start_date format")
	}
	_, e = time.Parse("2006-01-02", end)
	if e != nil {
		return err("invalid end_date format")
	}
	if granularity == "" {
		granularity = "DAILY"
	}
	return ok(fmt.Sprintf("Mock cost for %s from %s to %s: $5000.00", granularity, start, end))
}

func HandleGetForecast_aws_cost_explorer_mcp_server(ctx context.Context, args map[string]string) (ToolResponse, error) {
	granularity, _ :=getString(args, "granularity")
	start, _ :=getString(args, "start_date")
	if start == "" {
		return err("start_date is required")
	}
	_, e := time.Parse("2006-01-02", start)
	if e != nil {
		return err("invalid start_date format")
	}
	if granularity == "" {
		granularity = "MONTHLY"
	}
	return ok(fmt.Sprintf("Mock forecast for %s starting %s: $6000.00", granularity, start))
}