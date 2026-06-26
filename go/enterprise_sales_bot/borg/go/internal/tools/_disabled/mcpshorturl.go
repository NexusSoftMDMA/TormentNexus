package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HandleCreateShortUrl generates a short hash from a given URL and returns it.
func HandleCreateShortUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	hash := sha256.Sum256([]byte(url))
	short := hex.EncodeToString(hash[:])[:8]
	return ok(fmt.Sprintf("Short URL: %s", short))
}