package tools

import (
	"context"
	"strconv"
)

func HandleGetEmployee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	department, _ :=getString(args, "department")
	msg := "Employee: " + name
	if department != "" {
		msg += " - " + department
	}
	return ok(msg)
}

func HandleListEmployees(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	return success("List of employees (limit: " + strconv.Itoa(limit) + ")")
}