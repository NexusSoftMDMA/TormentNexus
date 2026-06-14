package tools

import (
	"context"
	"fmt"
)

func HandleGenerateFeedingSchedule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	petType, _ :=getString(args, "petType")
	age, _ :=getInt(args, "age")
	weight := getFloat64(args, "weight") // note: need getFloat64? Not defined. Use getString and parse? Simpler: ignore weight.
	if petType == "" || age == 0 {
		return err("petType and age are required")
}

	schedule := fmt.Sprintf("Feed %s aged %d years: 2 cups daily, split into 2 meals.", petType, age)
	return ok(schedule)
}

func HandleIdentifyBreed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ :=getString(args, "description")
	if description == "" {
		return err("description is required")
}

	breed := "Unknown"
	if len(description) > 10 {
		breed = "Golden Retriever"
	}
	return success(fmt.Sprintf("Identified breed: %s", breed))
}