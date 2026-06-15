package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func HandleRandomNumber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	min, _ :=getInt(args, "min")
	max, _ :=getInt(args, "max")
	if max <= min {
		max = min + 100
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(max-min) + min
	return ok(fmt.Sprintf("Random number: %d", n))
}

func HandleRandomString(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	length, _ :=getInt(args, "length")
	if length <= 0 {
		length = 10
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return ok(fmt.Sprintf("Random string: %s", string(b)))
}