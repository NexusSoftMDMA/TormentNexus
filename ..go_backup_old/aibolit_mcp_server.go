package tools

import (
	"context"
)

func HandleGetHealthTip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tip := "Eat balanced meals and exercise daily."
	return success(tip)
}

func HandleGetDoctorInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "doctor_name")
	if name == "" {
		name = "Dr. Aibolit"
	}
	return success(name + " is a qualified doctor.")
}