package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func HandleCountWords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	words := strings.Fields(text)
	count := len(words)
	return ok(fmt.Sprintf("Word count: %d", count))
}

func HandleSha256(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	h := sha256.Sum256([]byte(input))
	hashStr := hex.EncodeToString(h[:])
	return ok(fmt.Sprintf("SHA-256: %s", hashStr))
}