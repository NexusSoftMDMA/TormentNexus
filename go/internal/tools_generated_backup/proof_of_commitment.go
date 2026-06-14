package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HandleCommit generates a SHA-256 commitment from a message.
func HandleCommit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	hash := sha256.Sum256([]byte(message))
	commitment := hex.EncodeToString(hash[:])
	return ok(fmt.Sprintf("Commitment: %s", commitment))
}

// HandleVerify checks if a message matches a given commitment.
func HandleVerify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	commitment, _ :=getString(args, "commitment")
	if message == "" || commitment == "" {
		return err("message and commitment are required")
}

	hash := sha256.Sum256([]byte(message))
	expected := hex.EncodeToString(hash[:])
	if expected == commitment {
		return success("Commitment verified")
}

	return err("Commitment does not match")
}