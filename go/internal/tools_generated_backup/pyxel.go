package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		message = "no message"
	}
	resp, e := http.DefaultClient.Get("https://echo.example.com/echo?msg=" + message)
	if e != nil {
		return err("failed to call echo service")
}

	defer resp.Body.Close()
	return ok("Echo response received")
}