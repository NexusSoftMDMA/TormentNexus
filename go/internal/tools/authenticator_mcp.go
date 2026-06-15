//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleLogin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	if username == "" || password == "" {
		return err("username and password are required")
}

	// Simulate authentication against a hardcoded user
	if username != "admin" || password != "secret" {
		return err("invalid credentials")
}

	token := fmt.Sprintf("tok_%s_%d", username, 123456)
	return ok(token)
}

func HandleVerify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	// Simulate token verification
	if !strings.HasPrefix(token, "tok_") {
		return err("invalid token format")
}

	parts := strings.SplitN(token, "_", 3)
	if len(parts) < 3 || parts[2] != "123456" {
		return err("invalid token")
}

	userInfo := map[string]string{"username": parts[1]}
	data, e := json.Marshal(userInfo)
	if e != nil {
		return err("failed to marshal user info")
}

	return success(string(data))
}
