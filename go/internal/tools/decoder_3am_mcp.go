//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/base64"
	"encoding/hex"
)

func HandleDecode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	decType, _ :=getString(args, "type")
	switch decType {
	case "base64":
		decoded, e := base64.StdEncoding.DecodeString(input)
		if e != nil {
			return err("base64 decode failed: " + e.Error())
		return ok("decoded: " + string(decoded))
	case "hex":
		decoded, e := hex.DecodeString(input)
		if e != nil {
			return err("hex decode failed: " + e.Error())
		return ok("decoded: " + string(decoded))
	default:
		return err("unsupported type: " + decType)

}
}
}
}