package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HandleHashlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	secret, _ :=getString(args, "secret")
	if action == "generate" {
		hash := sha256.Sum256([]byte(secret))
		return ok(fmt.Sprintf("Hashlock: %s", hex.EncodeToString(hash[:])))
	} else if action == "verify" {
		hash, _ :=getString(args, "hash")
		computed := sha256.Sum256([]byte(secret))
		if hex.EncodeToString(computed[:]) == hash {
			return success("Hashlock verified")
}

		return err("Hashlock mismatch")
}

	return err("Invalid action: use 'generate' or 'verify'")
}