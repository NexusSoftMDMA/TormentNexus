//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleListExamples(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(fmt.Sprintf("Available examples: echo, calculator, weather"))
}

func HandleGetExample(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	switch name {
	case "echo":
		return success("Echo example: returns the input back")
	case "calculator":
		return success("Calculator example: performs arithmetic")
	case "weather":
		return success("Weather example: gets current weather")
	default:
		return err("unknown example: " + name)

}
}