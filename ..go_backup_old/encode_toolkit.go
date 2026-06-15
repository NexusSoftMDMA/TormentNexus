package tools

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func HandleEncode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	method, _ :=getString(args, "encoding")
	if method == "" {
		method = "base64"
	}
	var out string
	switch method {
	case "base64":
		out = base64.StdEncoding.EncodeToString([]byte(text))
	case "hex":
		out = hex.EncodeToString([]byte(text))
	default:
		return err("unsupported encoding: " + method)
}

	return ok(fmt.Sprintf("Encoded: %s", out))
}

func HandleDecode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	method, _ :=getString(args, "encoding")
	if method == "" {
		method = "base64"
	}
	var bytes []byte
	var e error
	switch method {
	case "base64":
		bytes, e = base64.StdEncoding.DecodeString(text)
	case "hex":
		bytes, e = hex.DecodeString(text)
	default:
		return err("unsupported encoding: " + method)
}

	if e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Decoded: %s", string(bytes)))
}