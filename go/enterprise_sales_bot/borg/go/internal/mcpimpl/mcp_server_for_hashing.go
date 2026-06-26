package mcpimpl

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HandleHash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	algorithm, _ :=getString(args, "algorithm")
	text, _ :=getString(args, "text")
	var hash string
	switch algorithm {
	case "sha256":
		h := sha256.Sum256([]byte(text))
		hash = hex.EncodeToString(h[:])
	case "sha1":
		h := sha1.Sum([]byte(text))
		hash = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum([]byte(text))
		hash = hex.EncodeToString(h[:])
	default:
		return err("unsupported algorithm: " + algorithm)
}

	return ok(fmt.Sprintf("Hash: %s", hash))
}

func HandleVerify_mcp_server_for_hashing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	algorithm, _ :=getString(args, "algorithm")
	text, _ :=getString(args, "text")
	expected, _ :=getString(args, "hash")
	var computed string
	switch algorithm {
	case "sha256":
		h := sha256.Sum256([]byte(text))
		computed = hex.EncodeToString(h[:])
	case "sha1":
		h := sha1.Sum([]byte(text))
		computed = hex.EncodeToString(h[:])
	case "md5":
		h := md5.Sum([]byte(text))
		computed = hex.EncodeToString(h[:])
	default:
		return err("unsupported algorithm: " + algorithm)
}

	if computed == expected {
		return success("Hash matches")
}

	return err("Hash does not match")
}