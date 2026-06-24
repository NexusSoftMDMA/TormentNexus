package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	upper := strings.ToUpper(message)
	body, e := json.Marshal(map[string]string{"echo": upper})
	if e != nil {
		return err("failed to marshal response")
}

	resp, e := http.DefaultClient.Post("https://echo.example.com", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to call echo service")
}

	defer resp.Body.Close()
	return ok("echo sent successfully")
}