package tools

import (
	"context"
	"encoding/base64"
)

func HandleEncodeBase64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	return ok(encoded)
}

func HandleDecodeBase64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	decoded, e := base64.StdEncoding.DecodeString(input)
	if e != nil {
		return err("invalid base64: " + e.Error())
}

	return ok(string(decoded))
}