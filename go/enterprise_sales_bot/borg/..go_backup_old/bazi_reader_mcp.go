package tools

import (
	"context"
	"fmt"
)

func HandleReadBazi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	year, _ :=getString(args, "year")
	month, _ :=getString(args, "month")
	day, _ :=getString(args, "day")
	hour, _ :=getString(args, "hour")
	gender, _ :=getString(args, "gender")
	result := fmt.Sprintf("Your Bazi: Year=%s, Month=%s, Day=%s, Hour=%s, Gender=%s", year, month, day, hour, gender)
	return ok(result)
}