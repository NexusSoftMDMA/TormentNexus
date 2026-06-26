package tools

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func HandleDecodeJwt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	secret, _ :=getString(args, "secret")
	if token == "" {
		return err("token is required")
}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return err("invalid JWT format")
}

	headerBytes, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return err("invalid header encoding")
}

	payloadBytes, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return err("invalid payload encoding")
}

	var header, payload map[string]interface{}
	if e := json.Unmarshal(headerBytes, &header); e != nil {
		return err("invalid header JSON")
}

	if e := json.Unmarshal(payloadBytes, &payload); e != nil {
		return err("invalid payload JSON")
}

	if secret != "" {
		signingInput := parts[0] + "." + parts[1]
		sigReceived, e := base64.RawURLEncoding.DecodeString(parts[2])
		if e != nil {
			return err("invalid signature encoding")
}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signingInput))
		expected := mac.Sum(nil)
		if !hmac.Equal(sigReceived, expected) {
			return err("invalid signature")

	}
	payload["_header"] = header
	payload["_valid"] = secret != ""
	b, _ := json.Marshal(payload)
	return ok("Decoded: " + string(b))
}
}