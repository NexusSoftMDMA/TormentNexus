package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ToolResponse is the response structure for MCP tool handlers.

// ok returns a success response with the given text content.
func ok(textContent string) (ToolResponse, error) {
	return ToolResponse{TextContent: textContent}, nil
}

// e returns an error response with the given error message.
func err(errMsg string) (ToolResponse, error) {
	return ToolResponse{TextContent: errMsg}, fmt.Errorf(errMsg)
}

// getString retrieves the value of the specified key from the args map.
func getString(args map[string]interface{}, key string) string {
	val, found := args[key]
	if !found {
		return ""
	}
	strVal, found := val.(string)
	if !found {
		return ""
	}
	return strVal
}

// getInt retrieves the value of the specified key from the args map as an integer.
func getInt(args map[string]interface{}, key string) (int, error) {
	strVal, _ :=getString(args, key)
	if strVal == "" {
		return 0, fmt.Errorf("missing key: %s", key)
}

	intVal, e := strconv.Atoi(strVal)
	if e != nil {
		return 0, e
	}
	return intVal, nil
}

// getBool retrieves the value of the specified key from the args map as a boolean.
func getBool(args map[string]interface{}, key string) (bool, error) {
	strVal, _ :=getString(args, key)
	if strVal == "" {
		return false, fmt.Errorf("missing key: %s", key)
}

	boolVal, e := strconv.ParseBool(strVal)
	if e != nil {
		return false, e
	}
	return boolVal, nil
}

// HandleExample is an example handler function.
func HandleExample(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Example implementation
	textContent := "Handled example request"
	return ok(textContent)
}

// HandleAnother is another example handler function.
func HandleAnother(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Another example implementation
	textContent := "Handled another request"
	return ok(textContent)
}

// HandleThird is a third example handler function.
func HandleThird(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Third example implementation
	textContent := "Handled third request"
	return ok(textContent)
}

// HandleFourth is a fourth example handler function.
func HandleFourth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Fourth example implementation
	textContent := "Handled fourth request"
	return ok(textContent)
}

// HandleFifth is a fifth example handler function.
func HandleFifth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Fifth example implementation
	textContent := "Handled fifth request"
	return ok(textContent)
}