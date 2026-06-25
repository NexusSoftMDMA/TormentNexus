package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ToolResponse is the response type for tool handlers.

// ok returns a success response with the given text content.
func ok(textContent string) (ToolResponse, error) {
	return ToolResponse{TextContent: textContent}, nil
}

// e returns an error response with the given error message.
func err(e error) (ToolResponse, error) {
	return ToolResponse{}, e
}

// getString retrieves the value of the specified key from the arguments map.
func getString(args map[string]interface{}, key string) (string, error) {
	val, found := args[key]
	if !found {
		return "", fmt.Errorf("key %s not found in arguments", key)
}

	strVal, found := val.(string)
	if !found {
		return "", fmt.Errorf("value for key %s is not a string", key)
}

	return strVal, nil
}

// HandleExample is an example handler function.
func HandleExample(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Example implementation
	textContent, _ :=getString(args, "exampleKey")
	if e != nil {
		return err(e)
}

	return ok(textContent)
}

// HandleAnother is another example handler function.
func HandleAnother(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Another example implementation
	return ok("This is another example response")
}

// HandleThird is yet another example handler function.
func HandleThird(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Yet another example implementation
	return ok("This is the third example response")
}

// HandleFourth is the fourth example handler function.
func HandleFourth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Fourth example implementation
	return ok("This is the fourth example response")
}

// HandleFifth is the fifth example handler function.
func HandleFifth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Fifth example implementation
	return ok("This is the fifth example response")
}