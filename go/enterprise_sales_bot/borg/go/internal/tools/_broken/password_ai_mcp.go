package mcpimpl

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func HandleGeneratePassword(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	length, _ :=getInt(args, "length")
	if length < 8 {
		length = 16
	}
	special, _ :=getBool(args, "use_special")
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if special {
		chars += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		n, e := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if e != nil {
			return err("failed to generate random number")
		}
		password[i] = chars[n.Int64()]
	}
	return ok(fmt.Sprintf("Password: %s", string(password)))
}

func HandleCheckStrength(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pwd, _ :=getString(args, "password")
	if len(pwd) == 0 {
		return err("password is required")
	}
	strength := "weak"
	if len(pwd) >= 12 {
		strength = "strong"
	} else if len(pwd) >= 8 {
		strength = "medium"
	}
	hash := sha256.Sum256([]byte(pwd))
	hashStr := hex.EncodeToString(hash[:])
	return ok(fmt.Sprintf("Strength: %s, Hash: %s", strength, hashStr))
}