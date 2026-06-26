package tools

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
)

func HandleSign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message must be provided")
}

	h := sha256.Sum256([]byte(msg))
	sig := base64.StdEncoding.EncodeToString(h[:])
	return success("signature: " + sig)
}

func HandleVerify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	sig, _ :=getString(args, "signature")
	if msg == "" || sig == "" {
		return err("message and signature are required")
}

	h := sha256.Sum256([]byte(msg))
	expected := base64.StdEncoding.EncodeToString(h[:])
	if sig == expected {
		return ok("signature is valid")
}

	return ok("signature is invalid")
}