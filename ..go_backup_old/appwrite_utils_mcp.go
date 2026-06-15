package tools

import (
	"context"
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

func HandleGenerateId(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prefix, _ :=getString(args, "prefix")
	length, _ :=getInt(args, "length")
	if length <= 0 {
		length = 20
	}
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		n, e := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if e != nil {
			return err("failed to generate random number")
}

		sb.WriteByte(chars[n.Int64()])

	return ok(prefix + sb.String())
}

}

func HandleValidateId(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	matched, e := regexp.MatchString("^[a-zA-Z0-9]{5,36}$", id)
	if e != nil {
		return err("validation regex failed")
}

	if matched {
		return success("id is valid")
}

	return ok("id is invalid")
}