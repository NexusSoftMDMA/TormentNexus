package mcpimpl

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"time"
)

func HandleGenerateSecret(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	secret := make([]byte, 20)
	_, e := rand.Read(secret)
	if e != nil {
		return err("failed to generate secret")
}

	return ok(base32.StdEncoding.EncodeToString(secret))
}

func HandleVerifyTOTP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	secretStr, _ :=getString(args, "secret")
	tokenStr, _ :=getString(args, "token")
	secret, e := base32.StdEncoding.DecodeString(secretStr)
	if e != nil {
		return err("invalid secret encoding")
}

	t := time.Now().Unix() / 30
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(t & 0xff)
		t >>= 8
	}
	mac := hmac.New(sha1.New, secret)
	mac.Write(msg)
	hash := mac.Sum(nil)
	offset := hash[19] & 0xf
	bin := int32(hash[offset]&0x7f)<<24 | int32(hash[offset+1]&0xff)<<16 |
		int32(hash[offset+2]&0xff)<<8 | int32(hash[offset+3]&0xff)
	otp := bin % 1000000
	expected := fmt.Sprintf("%06d", otp)
	if expected == tokenStr {
		return success("valid")
}

	return err("invalid token")
}